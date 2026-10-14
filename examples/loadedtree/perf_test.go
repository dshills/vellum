package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
	"time"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func TestTreePerfConfigAndTrajectory(t *testing.T) {
	for _, nodes := range []int{10_000, 100_000} {
		for _, deep := range []bool{false, true} {
			if err := (treePerfConfig{nodes, 240, 128, deep}).validate(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, config := range []treePerfConfig{{1000, 240, 128, false}, {10000, 119, 128, false}, {10000, 2049, 128, false}, {10000, 240, 31, false}, {10000, 240, 33, false}, {10000, 240, 2049, false}} {
		if config.validate() == nil {
			t.Fatal("invalid workload admitted", config)
		}
	}
	offset := 0
	for i := 0; i < 128; i++ {
		offset += treePerfDelta(i)
		if offset < 0 || offset > 192 {
			t.Fatal("trajectory escaped warm working set", i, offset)
		}
		if (i+1)%treePerfPeriod == 0 && offset != 0 {
			t.Fatal("trajectory did not close")
		}
	}
}
func perfTreeGeometry(rows int) cogent.TreeInspection {
	g := cogent.TreeInspection{Generation: 3, Frames: 1, TotalNodes: 100_000, VisibleRows: 100_000, RetainedRows: rows, Scale: 1, Bounds: image.Rect(10, 20, 510, 920), Viewport: virtual.OutlineViewport{Width: 500, Height: 900}}
	for i := 0; i < rows; i++ {
		g.Rows = append(g.Rows, cogent.TreeRowGeometry{Key: fmt.Sprint(i), Index: i, Bounds: image.Rect(10, 20+i*9, 510, 29+i*9)})
	}
	return g
}
func TestTreePerfGeometryRequiresActualClippedWorkload(t *testing.T) {
	g := perfTreeGeometry(64)
	original, err := treePerfGeometryCheck(g)
	if err != nil {
		t.Fatal(err)
	}
	// Retained pool count, source cardinality and empty overscan do not count as
	// visible rows. Changing source ordinals cannot change normalized geometry.
	g.Rows = append(g.Rows, cogent.TreeRowGeometry{Key: "empty"})
	g.RetainedRows++
	for i := range g.Rows {
		g.Rows[i].Index += 1000
	}
	detached, err := treePerfGeometryCheck(g)
	if err != nil || detached.ClippedRows != 64 || detached.LayoutSHA256 != original.LayoutSHA256 {
		t.Fatal(detached, err)
	}
	cases := map[string]func(*cogent.TreeInspection){
		"small":     func(g *cogent.TreeInspection) { g.Rows[0].Bounds = image.Rectangle{} },
		"duplicate": func(g *cogent.TreeInspection) { g.Rows[1].Key = g.Rows[0].Key },
		"escaped":   func(g *cogent.TreeInspection) { g.Rows[0].Bounds.Min.X-- },
		"pending":   func(g *cogent.TreeInspection) { g.Pending = true },
		"pool":      func(g *cogent.TreeInspection) { g.RetainedRows = 501 },
		"unloaded":  func(g *cogent.TreeInspection) { g.TotalNodes = 0 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			g := perfTreeGeometry(64)
			mutate(&g)
			if _, err := treePerfGeometryCheck(g); err == nil {
				t.Fatal("invalid measured workload qualified")
			}
		})
	}
}

// ErrTreeClosed also represents an unmounted startup binding. Once the
// consumer has observed the binding, that same error must terminate promptly.
func TestTreePerfInspectionDistinguishesStartupAndRetirement(t *testing.T) {
	for _, ready := range []bool{false, true} {
		t.Run(fmt.Sprint(ready), func(t *testing.T) {
			owner, err := app.New(app.Config{QueueCapacity: 8})
			if err != nil {
				t.Fatal(err)
			}
			defer owner.Close()
			if err = owner.Open(app.Window{ID: "perf", Title: "inert perf inspector", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label} }}); err != nil {
				t.Fatal(err)
			}
			p := &treePerfProgram{owner: owner, ready: ready} // no native binding exists
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			result := make(chan error, 1)
			go func() { _, e := p.inspect(ctx); result <- e }()
			watchdog := time.NewTimer(time.Second)
			defer watchdog.Stop()
			tick := time.NewTicker(time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case e := <-result:
					if ready && !errors.Is(e, cogent.ErrTreeClosed) {
						t.Fatal("retired tree retried", e)
					}
					if !ready && !errors.Is(e, context.DeadlineExceeded) {
						t.Fatal("startup bypassed bounded wait", e)
					}
					return
				case <-tick.C:
					if _, e := owner.DrainN(8); e != nil {
						t.Fatal(e)
					}
				case <-watchdog.C:
					t.Fatal("inspection failed to terminate")
				}
			}
		})
	}
	owner, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if err = owner.Open(app.Window{ID: "perf", Title: "closed perf origin", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	if err = owner.CloseWindow("perf"); err != nil {
		t.Fatal(err)
	}
	p := &treePerfProgram{owner: owner}
	if _, err = p.inspect(context.Background()); err == nil {
		t.Fatal("closed logical origin retried as startup")
	}
}
