package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dshills/vellum/virtual"
)

const columnCount = 20
const maxSourceRows = 1_000_000

var errSourceConfig = errors.New("grid example: invalid source config")

type sourceCalls struct{ RowKey, IndexOf, Cell uint64 }
type gridSource struct {
	identity          string
	revision          uint64
	total             int
	reverse, filtered bool
	// Only diagnostic counters mutate. Domain data and mapping are immutable for
	// each source revision; the application owner serializes all reads.
	calls *sourceCalls
}

func newSource(identity string, rows int) (*gridSource, error) {
	if identity == "" || rows < 1 || rows > maxSourceRows {
		return nil, errSourceConfig
	}
	return &gridSource{identity: identity, revision: 1, total: rows, calls: &sourceCalls{}}, nil
}
func (s *gridSource) Identity() string { return s.identity }
func (s *gridSource) Revision() uint64 { return s.revision }
func (s *gridSource) RowCount() int {
	if s.filtered {
		return (s.total + 1) / 2
	}
	return s.total
}
func rowKey(row int) string      { return fmt.Sprintf("row:%09d", row) }
func columnID(column int) string { return fmt.Sprintf("column:%02d", column) }
func (s *gridSource) domainRow(index int) (int, error) {
	if index < 0 || index >= s.RowCount() {
		return 0, virtual.ErrGridRange
	}
	if s.reverse {
		index = s.RowCount() - 1 - index
	}
	if s.filtered {
		index *= 2
	}
	return index, nil
}
func (s *gridSource) rowIndex(key string) (int, bool) {
	if len(key) != 13 || !strings.HasPrefix(key, "row:") {
		return 0, false
	}
	row, err := strconv.Atoi(key[4:])
	if err != nil || row < 0 || row >= s.total || rowKey(row) != key {
		return 0, false
	}
	if s.filtered {
		if row%2 != 0 {
			return 0, false
		}
		row /= 2
	}
	if s.reverse {
		row = s.RowCount() - 1 - row
	}
	return row, true
}
func (s *gridSource) RowKey(index int) (string, error) {
	s.calls.RowKey++
	row, err := s.domainRow(index)
	if err != nil {
		return "", err
	}
	return rowKey(row), nil
}
func (s *gridSource) IndexOf(key string) (int, bool, error) {
	s.calls.IndexOf++
	index, found := s.rowIndex(key)
	return index, found, nil
}
func (s *gridSource) Cell(key, column string) (string, error) {
	s.calls.Cell++
	if _, found := s.rowIndex(key); !found {
		return "", virtual.ErrGridRange
	}
	if len(column) != 9 || !strings.HasPrefix(column, "column:") {
		return "", virtual.ErrGridRange
	}
	col, err := strconv.Atoi(column[7:])
	if err != nil || col < 0 || col >= columnCount || columnID(col) != column {
		return "", virtual.ErrGridRange
	}
	row, _ := strconv.Atoi(key[4:])
	// The fixed seed and integer-only value function are independent of projection.
	value := (uint64(row)*1103515245 + uint64(col)*12345 + 7) & 0xffffffff
	return fmt.Sprintf("%06d · %02d · %08x", row, col, value), nil
}
func (s *gridSource) projected(reverse, filtered bool) (*gridSource, error) {
	if s.revision == math.MaxUint64 {
		return nil, virtual.ErrGridExhausted
	}
	next := *s
	next.reverse, next.filtered = reverse, filtered
	next.revision++
	return &next, nil
}
func sourceColumns() []virtual.GridColumn {
	columns := make([]virtual.GridColumn, columnCount)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: columnID(i), Title: fmt.Sprintf("Column %02d", i), Width: 120}
	}
	return columns
}

// rowMetricCalls counts public indexed metric callbacks, separately from the
// fixed source's unchanged row/cell callback-count fixtures.
type rowMetricCalls struct{ Offset, At uint64 }

// variableGridSource adds optional metrics without making the default fixed
// gridSource implement them. Heights are associated with stable domain keys,
// so reordering/filtering changes offsets while each retained row keeps height.
type variableGridSource struct {
	*gridSource
	metrics *rowMetricCalls
}

func domainRowHeight(row int) int {
	return [...]int{23, 35, 47, 29}[row%4]
}

// cycle describes the projected height period. Its maximum length is four,
// independent of row count; the even-only projection has length two.
func (s *variableGridSource) cycle() ([4]int, int, int, error) {
	var heights [4]int
	if s == nil || s.gridSource == nil || s.total < 1 || s.total > maxSourceRows {
		return heights, 0, 0, errSourceConfig
	}
	period, step, first := 4, 1, 0
	if s.filtered {
		period, step = 2, 2
	}
	if s.reverse {
		first, step = (s.RowCount()-1)*step, -step
	}
	sum := 0
	for i := 0; i < period; i++ {
		residue := (first + i*step) % 4
		if residue < 0 {
			residue += 4
		}
		heights[i] = domainRowHeight(residue)
		sum += heights[i]
	}
	return heights, period, sum, nil
}

func periodicOffset(boundary int, heights [4]int, period, sum int) int {
	offset := boundary / period * sum
	for i := 0; i < boundary%period; i++ {
		offset += heights[i]
	}
	return offset
}

func (s *variableGridSource) RowOffset(boundary int) (int, error) {
	if s != nil && s.metrics != nil {
		s.metrics.Offset++
	}
	heights, period, sum, err := s.cycle()
	if err != nil {
		return 0, err
	}
	if boundary < 0 || boundary > s.RowCount() {
		return 0, virtual.ErrGridRange
	}
	return periodicOffset(boundary, heights, period, sum), nil
}

func (s *variableGridSource) RowAt(pixel int) (int, error) {
	if s != nil && s.metrics != nil {
		s.metrics.At++
	}
	heights, period, sum, err := s.cycle()
	if err != nil {
		return 0, err
	}
	extent := periodicOffset(s.RowCount(), heights, period, sum)
	if pixel < 0 || pixel >= extent {
		return 0, virtual.ErrGridRange
	}
	row, remainder := pixel/sum*period, pixel%sum
	for i := 0; i < period; i++ {
		if remainder < heights[i] {
			return row + i, nil
		}
		remainder -= heights[i]
	}
	return 0, errSourceConfig // positive finite period makes this unreachable
}

var _ virtual.GridSource = (*variableGridSource)(nil)
