package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/dshills/vellum/virtual"
)

const columns = 20
const maxRows = 1_000_000
const maxVersions = 4096
const maxEditBytes = 64

var errData = errors.New("loadedgrid: invalid stored data")
var errValue = errors.New("loadedgrid: enter a canonical signed int64")
var errConflict = errors.New("loadedgrid: domain cell changed")

type record struct{ Values [columns]int64 }
type address struct{ Row, Column int }

// data and every published array/map are immutable. Only workers construct them.
type data struct {
	Records  []record
	Versions map[address]uint64
	Revision uint64
}
type readCounts struct{ Keys, Inverse, Cells atomic.Uint64 }
type countSnapshot struct{ Keys, Inverse, Cells uint64 }

func (c *readCounts) snapshot() countSnapshot {
	return countSnapshot{c.Keys.Load(), c.Inverse.Load(), c.Cells.Load()}
}

type fullSource struct {
	identity string
	data     *data
	rows     int
	counts   *readCounts
}
type variableSource struct{ *fullSource }

type loadResult struct {
	base    *data
	sources [2]*fullSource
	views   [2]virtual.GridProjection
}

func key(row int) string    { return fmt.Sprintf("row:%09d", row) }
func column(col int) string { return fmt.Sprintf("column:%02d", col) }
func parseKey(k string, rows int) (int, error) {
	if len(k) != 13 || !strings.HasPrefix(k, "row:") {
		return 0, virtual.ErrGridRange
	}
	n, e := strconv.Atoi(k[4:])
	if e != nil || n < 0 || n >= rows || key(n) != k {
		return 0, virtual.ErrGridRange
	}
	return n, nil
}
func parseColumn(k string) (int, error) {
	if len(k) != 9 || !strings.HasPrefix(k, "column:") {
		return 0, virtual.ErrGridRange
	}
	n, e := strconv.Atoi(k[7:])
	if e != nil || n < 0 || n >= columns || column(n) != k {
		return 0, virtual.ErrGridRange
	}
	return n, nil
}
func generate(ctx context.Context, rows int, progress func(float64, string)) (*data, error) {
	if ctx == nil || rows < 1 || rows > maxRows {
		return nil, errData
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	values := make([]record, rows)
	for r := range values {
		if r%1024 == 0 {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			if progress != nil {
				progress(float64(r)/float64(rows), "Generating stored records")
			}
		}
		for c := range values[r].Values {
			values[r].Values[c] = int64((uint64(r)*1103515245+uint64(c)*12345+7)%2_000_001) - 1_000_000
		}
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return &data{Records: values, Revision: 2}, nil
}
func emptySource(identity string) *fullSource {
	return &fullSource{strings.Clone(identity), &data{Revision: 1}, 0, &readCounts{}}
}
func source(identity string, d *data, rows int) (*fullSource, error) {
	if identity == "" || d == nil || d.Revision == 0 || rows < 0 || rows > len(d.Records) || len(d.Records) > maxRows {
		return nil, errData
	}
	return &fullSource{strings.Clone(identity), d, rows, &readCounts{}}, nil
}
func (s *fullSource) Identity() string { return s.identity }
func (s *fullSource) Revision() uint64 { return s.data.Revision }
func (s *fullSource) RowCount() int    { return s.rows }
func (s *fullSource) RowKey(index int) (string, error) {
	s.counts.Keys.Add(1)
	if index < 0 || index >= s.rows {
		return "", virtual.ErrGridRange
	}
	return key(index), nil
}
func (s *fullSource) IndexOf(k string) (int, bool, error) {
	s.counts.Inverse.Add(1)
	n, e := parseKey(k, s.rows)
	if e != nil {
		return 0, false, nil
	}
	return n, true, nil
}
func (s *fullSource) value(k, c string) (int64, uint64, error) {
	r, e := parseKey(k, s.rows)
	if e != nil {
		return 0, 0, e
	}
	col, e := parseColumn(c)
	if e != nil {
		return 0, 0, e
	}
	version := uint64(1)
	if v, ok := s.data.Versions[address{r, col}]; ok {
		version = v
	}
	return s.data.Records[r].Values[col], version, nil
}
func (s *fullSource) Cell(k, c string) (string, error) {
	s.counts.Cells.Add(1)
	v, _, e := s.value(k, c)
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(v, 10), nil
}
func (s *fullSource) EditCell(k, c string) (virtual.GridEditValue, error) {
	v, version, e := s.value(k, c)
	if e != nil {
		return virtual.GridEditValue{}, e
	}
	col, _ := parseColumn(c)
	return virtual.GridEditValue{Text: strconv.FormatInt(v, 10), Version: version, Editable: col != 0}, nil
}

// Capturing selection borrows only immutable row membership; it never keeps the
// diagnostic counters and does not scan/retain another million-row array.
type rowIndex struct {
	identity string
	revision uint64
	rows     int
}

func (s *fullSource) SelectionSnapshot() (virtual.GridRowIndex, error) {
	return rowIndex{strings.Clone(s.identity), s.Revision(), s.rows}, nil
}
func (s rowIndex) Identity() string { return s.identity }
func (s rowIndex) Revision() uint64 { return s.revision }
func (s rowIndex) RowCount() int    { return s.rows }
func (s rowIndex) RowKey(i int) (string, error) {
	if i < 0 || i >= s.rows {
		return "", virtual.ErrGridRange
	}
	return key(i), nil
}
func (s rowIndex) IndexOf(k string) (int, bool, error) {
	i, e := parseKey(k, s.rows)
	return i, e == nil, nil
}
func input(s *fullSource, variable bool) virtual.GridSource {
	if variable {
		return &variableSource{s}
	}
	return s
}

var heights = [4]int{23, 35, 47, 29}

func offset(boundary int) int {
	cycles, remainder := boundary/4, boundary%4
	total := cycles * 134
	for i := 0; i < remainder; i++ {
		total += heights[i]
	}
	return total
}
func (s *variableSource) RowOffset(boundary int) (int, error) {
	if boundary < 0 || boundary > s.rows {
		return 0, virtual.ErrGridRange
	}
	return offset(boundary), nil
}
func (s *variableSource) RowAt(pixel int) (int, error) {
	if pixel < 0 || pixel >= offset(s.rows) {
		return 0, virtual.ErrGridRange
	}
	row, p := pixel/134*4, pixel%134
	for _, height := range heights {
		if p < height {
			return row, nil
		}
		row++
		p -= height
	}
	return 0, virtual.ErrGridSource
}
func numeric(text string) (int64, error) {
	if len(text) == 0 || len(text) > maxEditBytes {
		return 0, errValue
	}
	n, e := strconv.ParseInt(text, 10, 64)
	if e != nil || strconv.FormatInt(n, 10) != text {
		return 0, errValue
	}
	return n, nil
}

// Copy all stored records off-owner. Sharing old immutable revisions is explicit;
// this copy and the builder's index/work allocations have separate budgets.
func edited(ctx context.Context, s *fullSource, k, c, text string, base uint64) (*fullSource, error) {
	if ctx == nil || s == nil {
		return nil, errData
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	n, e := numeric(text)
	if e != nil {
		return nil, e
	}
	_, version, e := s.value(k, c)
	if e != nil {
		return nil, e
	}
	if version != base {
		return nil, errConflict
	}
	if version == math.MaxUint64 || s.Revision() == math.MaxUint64 {
		return nil, virtual.ErrGridExhausted
	}
	r, _ := parseKey(k, s.rows)
	col, _ := parseColumn(c)
	a := address{r, col}
	if _, ok := s.data.Versions[a]; !ok && len(s.data.Versions) >= maxVersions {
		return nil, virtual.ErrGridCapacity
	}
	values := make([]record, s.rows)
	for first := 0; first < len(values); first += 1024 {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		last := min(first+1024, len(values))
		copy(values[first:last], s.data.Records[first:last])
	}
	versions := make(map[address]uint64, len(s.data.Versions)+1)
	for a, v := range s.data.Versions {
		versions[a] = v
	}
	values[r].Values[col] = n
	versions[a] = version + 1
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	return source(s.identity, &data{Records: values, Versions: versions, Revision: s.Revision() + 1}, s.rows)
}
func (s *fullSource) logicalBytes() uint64 {
	return uint64(len(s.data.Records)) * uint64(unsafe.Sizeof(record{}))
}
func limits() virtual.GridProjectionLimits {
	return virtual.GridProjectionLimits{MaxRows: maxRows, MaxIndexBytes: 64 << 20, MaxWorkingBytes: 256 << 20, MaxCellBytes: maxEditBytes, MaxQueryBytes: 4096}
}
