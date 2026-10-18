package workspace

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dshills/vellum/virtual"
)

// nativeGridReference is an independent fixed-row reference for this authored
// record. It never reads candidate scroll offsets or a candidate projection.
// Source replacement preserves the old stable row and intra-row offset, then
// clamps to the separately allocated viewport, as required by virtual.Grid.
// The only authored grid wheel is sequence 63: logical (76, -288).
type nativeGridReference struct {
	sequence                 uint64
	loaded, sorted, filtered bool
	order                    []int
	softwareY, nativeY       int
	x                        int
}

func nativeReferenceOrder(m Snapshot) []int {
	if !m.Loaded {
		return nil
	}
	out := make([]int, 0, m.StoredRows)
	for row := 0; row < m.StoredRows; row++ {
		if !m.Filtered || strings.Contains(strconv.Itoa(int(generatedValue(row, 1))), "7") {
			out = append(out, row)
		}
	}
	if m.Sorted {
		sort.Slice(out, func(i, j int) bool {
			a, b := generatedValue(out[i], 0), generatedValue(out[j], 0)
			if a == b {
				return out[i] < out[j]
			}
			return a > b
		})
	}
	return out
}
func nativeReferenceReplace(old, next []int, y int) int {
	if len(old) == 0 {
		return y
	}
	index := min(y/24, len(old)-1)
	key, offset := old[index], y%24
	for rank, row := range next {
		if row == key {
			return rank*24 + offset
		}
	}
	return y // Removed anchor retains the offset before viewport clamping.
}
func nativeReferenceClamp(y, count, height int) int {
	return max(0, min(y, max(0, count*24-max(0, height-28))))
}
func (r *nativeGridReference) offsets(sequence uint64, m Snapshot, viewport virtual.GridViewport) (int, int, error) {
	if sequence <= r.sequence || m.Grid.Config.RowHeight != 24 || m.Grid.Config.HeaderHeight != 28 || viewport.Width <= 0 || viewport.Height <= 0 {
		return 0, 0, fmt.Errorf("unsupported native reference sequence/configuration")
	}
	changed := r.loaded != m.Loaded || r.sorted != m.Sorted || r.filtered != m.Filtered
	order := r.order
	software, native := r.softwareY, r.nativeY
	if changed {
		order = nativeReferenceOrder(m)
		software = nativeReferenceReplace(r.order, order, software)
		native = nativeReferenceReplace(r.order, order, native)
	}
	if r.sequence < 63 && sequence >= 63 {
		r.x += 76
		software -= 288
		native -= 288
	}
	software = nativeReferenceClamp(software, len(order), m.Grid.Viewport.Height)
	native = nativeReferenceClamp(native, len(order), viewport.Height)
	if len(order) != m.Grid.Stamp.Rows || m.Grid.TotalHeight != len(order)*24 || software != m.Grid.Viewport.ScrollY || r.x != m.Grid.Viewport.ScrollX {
		return 0, 0, fmt.Errorf("immutable grid record disagrees with independent source/anchor/wheel reference at %d", sequence)
	}
	r.sequence, r.loaded, r.sorted, r.filtered = sequence, m.Loaded, m.Sorted, m.Filtered
	r.order, r.softwareY, r.nativeY = order, software, native
	return r.x, native, nil
}
func (r *nativeGridReference) compare(want Capture, got NativeCapture) error {
	x, y, err := r.offsets(got.Sequence, want.Model, got.Model.Grid.Viewport)
	if err != nil {
		return err
	}
	// These expectations depend on immutable data and native viewport dimensions,
	// never the candidate offset under test. Other domain fields stay exact.
	want.Model.Grid.Viewport.ScrollX, want.Model.Grid.Viewport.ScrollY = x, y
	return compareNativeCapture(want, got)
}
