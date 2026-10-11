package virtual

// GridSortKind selects exact UTF-8 byte order or strict signed decimal int64
// order. Equal tuples preserve the input snapshot's original row order.
type GridSortKind uint8

const (
	GridSortText GridSortKind = iota
	GridSortInteger
)

type GridSortColumn struct {
	ColumnID   string
	Kind       GridSortKind
	Descending bool
}

type GridFilterKind uint8

const (
	GridFilterContains GridFilterKind = iota
	GridFilterEquals
)

// GridFilter performs case-sensitive exact UTF-8 equality or substring matching.
// A query ANDs its filters. No application callback executes inside sorting.
type GridFilter struct {
	ColumnID string
	Kind     GridFilterKind
	Text     string
}

type GridQuery struct {
	Sort    []GridSortColumn
	Filters []GridFilter
}

// GridProjectionLimits bound rows, retained index arrays and transient working
// arrays/string payload. They exclude borrowed immutable domain memory, Go
// allocator overhead, native/GPU memory and separate caller-held results.
type GridProjectionLimits struct {
	MaxRows, MaxIndexBytes, MaxWorkingBytes, MaxCellBytes, MaxQueryBytes int
}

type GridProjectionStats struct {
	InputRows, Rows, IndexBytes, WorkingBytes int
	CellReads, SortComparisons                uint64
}

// GridProjection is an immutable worker-prepared view of an immutable indexed
// snapshot. BuildGridProjection is its constructor. Optional row metrics and
// editing capabilities are preserved only when the input supplies them.
// SelectionSnapshot is constant-time and safe for concurrent worker reads.
type GridProjection interface {
	GridSource
	GridSelectionSource
	InputStamp() GridStamp
	Stats() GridProjectionStats
	gridProjectionMarker()
}
