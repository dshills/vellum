package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/virtual"
)

const columnCount = 20
const maxSourceRows = 1_000_000
const maxCellOverrides = 4096
const maxCellTextBytes = 256

var errSourceConfig = errors.New("grid example: invalid source config")
var errCellValidation = errors.New("grid example: cell text must be nonblank UTF-8, at most 256 bytes, without line breaks or NUL")
var errCellConflict = errors.New("grid example: cell changed since editing began")

type cellAddress struct{ key, column string }
type cellOverride struct {
	text    string
	version uint64
}

type sourceCalls struct{ RowKey, IndexOf, Cell uint64 }
type gridSource struct {
	identity          string
	revision          uint64
	total             int
	reverse, filtered bool
	// Only diagnostic counters mutate. Domain data and mapping are immutable for
	// each source revision; the application owner serializes all reads.
	calls *sourceCalls
	// A source revision owns an immutable, bounded sparse map. Reprojection
	// shares it; an edit copies it before publishing a new source revision.
	overrides map[cellAddress]cellOverride
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
	value, err := s.cellValue(key, column)
	return value.text, err
}

// cellValue is counter-free and resolves domain identity even while a row is
// hidden. Grid's edit transaction separately requires its current visibility.
func (s *gridSource) cellValue(key, column string) (cellOverride, error) {
	if s == nil || len(key) != 13 || !strings.HasPrefix(key, "row:") || len(column) != 9 || !strings.HasPrefix(column, "column:") {
		return cellOverride{}, virtual.ErrGridRange
	}
	row, rowErr := strconv.Atoi(key[4:])
	col, colErr := strconv.Atoi(column[7:])
	if rowErr != nil || colErr != nil || row < 0 || row >= s.total || col < 0 || col >= columnCount || rowKey(row) != key || columnID(col) != column {
		return cellOverride{}, virtual.ErrGridRange
	}
	if value, found := s.overrides[cellAddress{key, column}]; found {
		return value, nil
	}
	// The fixed seed and integer-only value function are independent of projection.
	value := (uint64(row)*1103515245 + uint64(col)*12345 + 7) & 0xffffffff
	return cellOverride{fmt.Sprintf("%06d · %02d · %08x", row, col, value), 1}, nil
}

func validateCellText(text string) error {
	if !utf8.ValidString(text) || len(text) > maxCellTextBytes || strings.TrimSpace(text) == "" || strings.ContainsAny(text, "\x00\r\n") {
		return errCellValidation
	}
	return nil
}

func (s *gridSource) withCell(key, column, text string, baseVersion uint64) (*gridSource, error) {
	if err := validateCellText(text); err != nil {
		return nil, err
	}
	value, err := s.cellValue(key, column)
	if err != nil {
		return nil, err
	}
	if value.version != baseVersion {
		return nil, errCellConflict
	}
	if s.revision == math.MaxUint64 || value.version == math.MaxUint64 {
		return nil, virtual.ErrGridExhausted
	}
	address := cellAddress{key, column}
	if _, found := s.overrides[address]; !found && len(s.overrides) >= maxCellOverrides {
		return nil, virtual.ErrGridCapacity
	}
	// Clone admitted strings as well as the map so a tiny supplied substring
	// cannot retain a much larger caller backing allocation.
	next := *s
	next.overrides = make(map[cellAddress]cellOverride, len(s.overrides)+1)
	for address, prior := range s.overrides {
		next.overrides[address] = prior
	}
	next.overrides[cellAddress{strings.Clone(key), strings.Clone(column)}] = cellOverride{strings.Clone(text), value.version + 1}
	next.revision++
	return &next, nil
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

// selectionRowIndex owns only immutable metadata. It does not borrow the
// owner's source pointer or its diagnostic counters, so detached selections may
// visit their captured identities concurrently after projection or unmount.
type selectionRowIndex struct {
	identity          string
	revision          uint64
	total             int
	reverse, filtered bool
}

func (s *gridSource) SelectionSnapshot() (virtual.GridRowIndex, error) {
	if s == nil || s.identity == "" || s.revision == 0 || s.total < 1 || s.total > maxSourceRows {
		return nil, errSourceConfig
	}
	return selectionRowIndex{identity: strings.Clone(s.identity), revision: s.revision, total: s.total, reverse: s.reverse, filtered: s.filtered}, nil
}

func (s selectionRowIndex) Identity() string { return s.identity }
func (s selectionRowIndex) Revision() uint64 { return s.revision }
func (s selectionRowIndex) RowCount() int {
	if s.filtered {
		return (s.total + 1) / 2
	}
	return s.total
}
func (s selectionRowIndex) RowKey(index int) (string, error) {
	if index < 0 || index >= s.RowCount() {
		return "", virtual.ErrGridRange
	}
	if s.reverse {
		index = s.RowCount() - 1 - index
	}
	if s.filtered {
		index *= 2
	}
	return rowKey(index), nil
}
func (s selectionRowIndex) IndexOf(key string) (int, bool, error) {
	// The existing inverse is pure; this value contains no owner/counter pointer.
	index, found := (&gridSource{total: s.total, reverse: s.reverse, filtered: s.filtered}).rowIndex(key)
	return index, found, nil
}

var _ virtual.GridSelectionSource = (*gridSource)(nil)
var _ virtual.GridRowIndex = selectionRowIndex{}

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

// Only these opt-in wrappers implement the public edit source contract. The
// default gridSource and variableGridSource keep their legacy behavior.
type editableGridSource struct{ *gridSource }
type editableVariableGridSource struct{ *variableGridSource }

func editableValue(s *gridSource, key, column string) (virtual.GridEditValue, error) {
	value, err := s.cellValue(key, column)
	if err != nil {
		return virtual.GridEditValue{}, err
	}
	// Column zero demonstrates a source-controlled read-only cell. An external
	// domain update may still replace it; user edit admission may not.
	return virtual.GridEditValue{Text: value.text, Version: value.version, Editable: column != columnID(0)}, nil
}
func (s *editableGridSource) EditCell(key, column string) (virtual.GridEditValue, error) {
	return editableValue(s.gridSource, key, column)
}
func (s *editableVariableGridSource) EditCell(key, column string) (virtual.GridEditValue, error) {
	return editableValue(s.gridSource, key, column)
}

var _ virtual.GridEditSource = (*editableGridSource)(nil)
var _ virtual.GridEditSource = (*editableVariableGridSource)(nil)
