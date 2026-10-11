package virtual

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

type variableRowSource struct {
	*indexedGridSource
	index      *RowHeights
	offsetHook func(int) (int, error)
	atHook     func(int) (int, error)
}

func variableSource(t *testing.T, heights []int) *variableRowSource {
	t.Helper()
	index, err := NewRowHeights(heights)
	if err != nil {
		t.Fatal(err)
	}
	return &variableRowSource{indexedGridSource: gridSource(len(heights)), index: index}
}
func (s *variableRowSource) RowOffset(i int) (int, error) {
	if s.offsetHook != nil {
		return s.offsetHook(i)
	}
	return s.index.RowOffset(i)
}
func (s *variableRowSource) RowAt(pixel int) (int, error) {
	if s.atHook != nil {
		return s.atHook(pixel)
	}
	return s.index.RowAt(pixel)
}

func TestRowHeightsImmutableBoundariesAndInverse(t *testing.T) {
	rng := rand.New(rand.NewSource(81277))
	for trial := 0; trial < 100; trial++ {
		heights := make([]int, rng.Intn(100))
		for i := range heights {
			heights[i] = 1 + rng.Intn(100)
		}
		original := append([]int(nil), heights...)
		index, err := NewRowHeights(heights)
		if err != nil {
			t.Fatal(err)
		}
		clear(heights)
		copyIndex := *index // Public copies share only immutable private storage.
		if copyIndex.RowCount() != len(original) {
			t.Fatal("row count")
		}
		prefix := 0
		for row, height := range original {
			if offset, err := index.RowOffset(row); err != nil || offset != prefix {
				t.Fatal(row, offset, prefix, err)
			}
			for pixel := prefix; pixel < prefix+height; pixel++ {
				if got, err := copyIndex.RowAt(pixel); err != nil || got != row {
					t.Fatal("half-open inverse", row, pixel, got, err)
				}
			}
			prefix += height
		}
		if total, err := index.RowOffset(len(original)); err != nil || total != prefix {
			t.Fatal(total, prefix, err)
		}
		for _, pixel := range []int{-1, prefix, math.MaxInt} {
			if _, err := index.RowAt(pixel); !errors.Is(err, ErrGridRange) {
				t.Fatal("out of range pixel", pixel, err)
			}
		}
		for _, row := range []int{-1, len(original) + 1} {
			if _, err := index.RowOffset(row); !errors.Is(err, ErrGridRange) {
				t.Fatal("out of range boundary", row, err)
			}
		}
	}
}

type cancellingIndexContext struct {
	context.Context
	cancel func()
	calls  int
}

func (c *cancellingIndexContext) Err() error {
	c.calls++
	if c.calls == 4 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestRowHeightsPreparationRefusalAndCancellation(t *testing.T) {
	for _, heights := range [][]int{{0}, {-1}, {math.MaxInt, 1}} {
		if index, err := NewRowHeights(heights); index != nil || !errors.Is(err, ErrGridInvalid) {
			t.Fatal(index, err)
		}
	}
	for _, index := range []*RowHeights{nil, {}} {
		if _, err := index.RowOffset(0); !errors.Is(err, ErrGridInvalid) {
			t.Fatal(err)
		}
		if _, err := index.RowAt(0); !errors.Is(err, ErrGridInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := NewRowHeightsContext(nil, nil); !errors.Is(err, ErrGridInvalid) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if index, err := NewRowHeightsContext(ctx, nil); index != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(index, err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	heights := make([]int, 1000)
	for i := range heights {
		heights[i] = 7
	}
	during := &cancellingIndexContext{Context: ctx, cancel: cancel}
	if index, err := NewRowHeightsContext(during, heights); index != nil || !errors.Is(err, context.Canceled) || during.calls != 4 {
		t.Fatal("mid-preparation cancellation", index, err, during.calls)
	}
	// This is caller-owned input storage. The rejected constructor must not
	// allocate or touch a second over-budget prefix.
	overBudget := make([]int, rowHeightsMaxBytes/(strconv.IntSize/8))
	if index, err := NewRowHeights(overBudget); index != nil || !errors.Is(err, ErrGridCapacity) {
		t.Fatal(index, err)
	}
}

func TestGridVariableRowsBruteGeometry(t *testing.T) {
	rng := rand.New(rand.NewSource(719002))
	for trial := 0; trial < 500; trial++ {
		heights := make([]int, rng.Intn(40))
		total := 0
		for i := range heights {
			heights[i] = 1 + rng.Intn(70)
			total += heights[i]
		}
		source := variableSource(t, heights)
		columns := gridColumns(1+rng.Intn(20), 1)
		for i := range columns {
			columns[i].Width = 1 + rng.Intn(80)
		}
		config := GridConfig{RowHeight: 7, HeaderHeight: rng.Intn(30), FrozenColumns: rng.Intn(len(columns) + 1), OverscanRows: rng.Intn(3), OverscanColumns: rng.Intn(3)}
		g := mustGrid(t, source, columns, config, GridViewport{Width: 1 + rng.Intn(500), Height: 1 + rng.Intn(300), ScrollX: rng.Intn(1000), ScrollY: rng.Intn(3000)})
		snapshot := gridSnapshot(t, g)
		if !snapshot.State.VariableHeights || snapshot.State.TotalHeight != total || snapshot.State.Config.RowHeight != 7 {
			t.Fatal(snapshot.State)
		}
		v := snapshot.State.Viewport
		want := map[[2]int]GridRect{}
		frozen := 0
		for _, c := range columns[:config.FrozenColumns] {
			frozen += c.Width
		}
		x := 0
		for col, c := range columns {
			left, right, cellX := 0, min(v.Width, frozen), x
			if col >= config.FrozenColumns {
				left, right, cellX = min(v.Width, frozen), v.Width, x-v.ScrollX
			}
			y := config.HeaderHeight - v.ScrollY
			for row, height := range heights {
				x0, x1 := max(cellX, left), min(cellX+c.Width, right)
				y0, y1 := max(y, min(v.Height, config.HeaderHeight)), min(y+height, v.Height)
				if x1 > x0 && y1 > y0 {
					want[[2]int{row, col}] = GridRect{x0, y0, x1 - x0, y1 - y0}
				}
				y += height
			}
			x += c.Width
		}
		got := map[[2]int]GridRect{}
		for _, c := range snapshot.Cells {
			start := 0
			for _, h := range heights[:c.RowIndex] {
				start += h
			}
			if c.Rect.Y != config.HeaderHeight+start-v.ScrollY || c.Rect.Height != heights[c.RowIndex] {
				t.Fatal("natural indexed geometry", c)
			}
			if !c.Clip.Empty() {
				got[[2]int{c.RowIndex, c.ColumnIndex}] = c.Clip
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("trial%d viewport%+v got%v want%v", trial, v, got, want)
		}
	}
}

func TestGridVariableAnchorShrinkGrowAndReorder(t *testing.T) {
	s := variableSource(t, []int{100, 100, 100, 100})
	g := mustGrid(t, s, gridColumns(1, 80), GridConfig{RowHeight: 24, HeaderHeight: 10}, GridViewport{Width: 80, Height: 30, ScrollY: 150})
	old := gridSnapshot(t, g).Cells[0].Ref
	if err := g.Activate(old); err != nil {
		t.Fatal(err)
	}
	next := variableSource(t, []int{100, 20, 100, 100})
	next.revision = 2
	if err := g.ReplaceSource(next); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.Viewport.ScrollY != 119 || state.Active.RowKey != "r:1" || g.anchorKey != "r:1" || g.anchorOffset != 19 {
		t.Fatal("shrunken anchor escaped its row", state, g.anchorKey, g.anchorOffset)
	}
	if err := g.Activate(old); !errors.Is(err, ErrGridStale) {
		t.Fatal("old geometry ref survived", err)
	}
	next = variableSource(t, []int{100, 100, 20, 100})
	next.revision, next.reverse = 3, true
	if err := g.ReplaceSource(next); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.Viewport.ScrollY != 219 || state.Active.RowKey != "r:1" || g.anchorKey != "r:1" || g.anchorOffset != 19 {
		t.Fatal("reordered anchor", state, g.anchorKey, g.anchorOffset)
	}
	next = variableSource(t, []int{100, 100, 200, 100})
	next.revision, next.reverse = 4, true
	if err := g.ReplaceSource(next); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.Viewport.ScrollY != 219 || state.Active.RowKey != "r:1" || state.TotalHeight != 500 {
		t.Fatal("grown anchor", state)
	}
}

func TestGridVariablePixelPagingTallRowsAndIntBoundary(t *testing.T) {
	g := mustGrid(t, variableSource(t, []int{10, 120, 5, 40, 8}), gridColumns(1, 80), GridConfig{RowHeight: 24, HeaderHeight: 7}, GridViewport{Width: 80, Height: 57})
	for _, step := range []struct {
		direction int
		key       string
	}{{1, "r:1"}, {1, "r:2"}, {-1, "r:1"}, {-1, "r:0"}} {
		if err := g.Page(step.direction); err != nil || gridState(t, g).Active.RowKey != step.key {
			t.Fatal(step, g.state(), err)
		}
	}
	before := g.state()
	if err := g.Page(-1); err != nil || g.state() != before {
		t.Fatal("endpoint no-op", err, g.state())
	}
	g = mustGrid(t, variableSource(t, []int{math.MaxInt - 1, 1}), gridColumns(1, 1), GridConfig{RowHeight: 1}, GridViewport{Width: 1, Height: 1})
	if err := g.JumpToKey("r:1", "c:0"); err != nil {
		t.Fatal(err)
	}
	if state := gridState(t, g); state.TotalHeight != math.MaxInt || state.Viewport.ScrollY != math.MaxInt-1 {
		t.Fatal(state)
	}
	if err := g.Page(1); err != nil || gridState(t, g).Active.RowKey != "r:1" {
		t.Fatal("saturated pixel page", err)
	}
	if err := g.Page(-1); err != nil || gridState(t, g).Active.RowKey != "r:0" {
		t.Fatal("giant previous row", err)
	}
	if _, err := NewGrid(variableSource(t, []int{math.MaxInt - 1, 1}), gridColumns(1, 1), GridConfig{RowHeight: 1, HeaderHeight: 1}, GridViewport{Width: 1, Height: 1}); !errors.Is(err, ErrGridSource) {
		t.Fatal("total plus header overflow admitted", err)
	}
}
