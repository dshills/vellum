package virtual

import (
	"context"
	"fmt"
	"math"
	"strconv"
)

// GridRowMetrics optionally extends GridSource with indexed variable heights.
// RowOffset returns boundaries 0 through RowCount inclusive; boundary 0 is zero
// and every next boundary is strictly greater. RowAt maps a logical pixel in
// [0, RowOffset(RowCount)) to its containing row. Reads must be bounded, indexed,
// nonblocking and immutable within the source revision, like RowKey/IndexOf.
// Any change to heights, order or membership must advance that revision.
type GridRowMetrics interface {
	RowOffset(boundary int) (int, error)
	RowAt(pixel int) (int, error)
}

// RowHeights is an immutable prefix index implementing GridRowMetrics. Its
// private prefix does not borrow the input slice. Preparation is O(N) time and
// (N+1)*sizeof(int) retained bytes; build it before installing a source, usually
// on a worker. RowOffset is O(1), RowAt O(log N). Copies share immutable storage.
type RowHeights struct{ prefix []int }

const rowHeightsMaxBytes = 64 << 20

func NewRowHeights(heights []int) (*RowHeights, error) {
	return NewRowHeightsContext(context.Background(), heights)
}

// NewRowHeightsContext prepares an index with at most 64 MiB of prefix storage.
// Positive heights and their sum are checked before publication. Cancellation
// is checked before allocation and at least every 256 rows. The caller must not
// mutate heights while construction runs. Larger/specialized datasets may
// supply their own GridRowMetrics instead of this materialized helper.
func NewRowHeightsContext(ctx context.Context, heights []int) (*RowHeights, error) {
	if ctx == nil {
		return nil, ErrGridInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(heights) >= rowHeightsMaxBytes/(strconv.IntSize/8) {
		return nil, ErrGridCapacity
	}
	prefix := make([]int, len(heights)+1)
	for i, height := range heights {
		if i%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		if height <= 0 || height > math.MaxInt-prefix[i] {
			return nil, ErrGridInvalid
		}
		prefix[i+1] = prefix[i] + height
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &RowHeights{prefix: prefix}, nil
}

func (h *RowHeights) RowCount() int {
	if h == nil || len(h.prefix) == 0 {
		return 0
	}
	return len(h.prefix) - 1
}
func (h *RowHeights) RowOffset(boundary int) (int, error) {
	if h == nil || len(h.prefix) == 0 {
		return 0, ErrGridInvalid
	}
	if boundary < 0 || boundary >= len(h.prefix) {
		return 0, ErrGridRange
	}
	return h.prefix[boundary], nil
}
func (h *RowHeights) RowAt(pixel int) (int, error) {
	if h == nil || len(h.prefix) == 0 {
		return 0, ErrGridInvalid
	}
	if pixel < 0 || pixel >= h.prefix[len(h.prefix)-1] {
		return 0, ErrGridRange
	}
	// Upper bound: a pixel exactly at a boundary belongs to the following row.
	lo, hi := 0, len(h.prefix)-1
	for lo < hi {
		mid := lo + (hi-lo)/2
		if h.prefix[mid+1] <= pixel {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo, nil
}

// gridRows is a prospective, validated geometry profile. Transitions never
// assign it to Grid until every layout/identity/stamp check has succeeded.
type gridRows struct {
	count, height, total int
	metrics              GridRowMetrics
}

func newGridRows(source GridSource, stamp GridStamp, config GridConfig) (gridRows, error) {
	r := gridRows{count: stamp.Rows, height: config.RowHeight}
	if metrics, ok := source.(GridRowMetrics); ok {
		r.metrics = metrics
		start, err := metrics.RowOffset(0)
		if err != nil {
			return gridRows{}, fmt.Errorf("%w: row offset: %w", ErrGridSource, err)
		}
		total, err := metrics.RowOffset(stamp.Rows)
		if err != nil {
			return gridRows{}, fmt.Errorf("%w: row extent: %w", ErrGridSource, err)
		}
		if start != 0 || total < stamp.Rows || (stamp.Rows == 0 && total != 0) || total > math.MaxInt-config.HeaderHeight {
			return gridRows{}, ErrGridSource
		}
		r.total = total
	} else {
		if stamp.Rows > (math.MaxInt-config.HeaderHeight)/config.RowHeight {
			return gridRows{}, ErrGridInvalid
		}
		r.total = stamp.Rows * config.RowHeight
	}
	return r, nil
}
func (r gridRows) current() error {
	if r.metrics == nil {
		return nil
	}
	start, err := r.metrics.RowOffset(0)
	if err != nil {
		return fmt.Errorf("%w: row offset: %w", ErrGridSource, err)
	}
	total, err := r.metrics.RowOffset(r.count)
	if err != nil {
		return fmt.Errorf("%w: row extent: %w", ErrGridSource, err)
	}
	if start != 0 || total != r.total {
		return ErrGridSource
	}
	return nil
}
func checkedGridRows(source GridSource, stamp GridStamp, rows gridRows) error {
	if err := sameStamp(source, stamp); err != nil {
		return err
	}
	if rows.metrics != nil {
		if err := rows.current(); err != nil {
			return err
		}
		return sameStamp(source, stamp)
	}
	return nil
}
func (r gridRows) bounds(row int) (start, height int, err error) {
	if row < 0 || row >= r.count {
		return 0, 0, ErrGridRange
	}
	if r.metrics == nil {
		return row * r.height, r.height, nil
	}
	start, err = r.metrics.RowOffset(row)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: row offset: %w", ErrGridSource, err)
	}
	end, err := r.metrics.RowOffset(row + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: row end: %w", ErrGridSource, err)
	}
	if start < 0 || end <= start || end > r.total || (row == 0 && start != 0) || (row+1 == r.count && end != r.total) {
		return 0, 0, ErrGridSource
	}
	if row > 0 {
		previous, err := r.metrics.RowOffset(row - 1)
		if err != nil {
			return 0, 0, fmt.Errorf("%w: previous row offset: %w", ErrGridSource, err)
		}
		if previous < 0 || previous >= start {
			return 0, 0, ErrGridSource
		}
	}
	for _, pixel := range []int{start, end - 1} {
		inverse, err := r.metrics.RowAt(pixel)
		if err != nil {
			return 0, 0, fmt.Errorf("%w: row inverse: %w", ErrGridSource, err)
		}
		if inverse != row {
			return 0, 0, ErrGridSource
		}
	}
	return start, end - start, nil
}
func (r gridRows) at(pixel int) (int, error) {
	if pixel < 0 || pixel >= r.total {
		return 0, ErrGridRange
	}
	if r.metrics == nil {
		return pixel / r.height, nil
	}
	row, err := r.metrics.RowAt(pixel)
	if err != nil {
		return 0, fmt.Errorf("%w: row at: %w", ErrGridSource, err)
	}
	if row < 0 || row >= r.count {
		return 0, ErrGridSource
	}
	start, height, err := r.bounds(row)
	if err != nil {
		return 0, err
	}
	if pixel < start || pixel-start >= height {
		return 0, ErrGridSource
	}
	return row, nil
}
