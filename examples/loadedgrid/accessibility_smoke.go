package main

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"math"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// This separate acceptance process queries actual Cocoa objects. Its queries,
// identity records and waits are excluded from the ordinary performance mode.
func gridAccessibilityReady(ctx context.Context, p *perfProgram) (cogent.GridInspection, map[string]cogent.AccessibilityElement, error) {
	for {
		g, err := p.inspect(ctx)
		if err != nil {
			return g, nil, err
		}
		s, err := cogent.InspectAccessibility(ctx, "perf")
		if err == nil && s.Reachable && s.NativeWindow {
			byDomain := map[string]cogent.AccessibilityElement{}
			for _, x := range s.Elements {
				byDomain[x.DomainID] = x
			}
			host := byDomain["data"]
			if host.GridIncarnation == g.Incarnation && host.GridGeneration == g.Generation && !host.GridPending && host.Frame[2] > 0 {
				if _, err := checkPerfGeometry(g); err != nil {
					return g, nil, err
				}
				scale := float64(g.Bounds.Dx()) / host.Frame[2]
				visible, actual := 0, 0
				for _, x := range s.Elements {
					if x.Role == "AXCell" && x.GridScope == host.ID {
						actual++
					}
				}
				for _, cell := range g.Cells {
					if cell.Bounds.Empty() {
						continue
					}
					visible++
					x, found := byDomain[cell.ID]
					if !found || x.GridPending || x.GridIncarnation != g.Incarnation || x.GridGeneration != g.Generation || x.GridTargetGeneration != g.Generation || x.GridScope != host.ID || x.ObjectIdentity == 0 || x.Value != cell.Text || x.Name != cell.Text || x.RowIndex != [2]int{cell.Row, 1} || x.ColumnIndex != [2]int{cell.Column, 1} || x.HitTestID != x.ID {
						return g, nil, fmt.Errorf("grid AX cell semantics/identity mismatch: %s %+v", cell.ID, x)
					}
					want := [4]float64{host.Frame[0] + float64(cell.Bounds.Min.X-g.Bounds.Min.X)/scale, host.Frame[1] + float64(g.Bounds.Max.Y-cell.Bounds.Max.Y)/scale, float64(cell.Bounds.Dx()) / scale, float64(cell.Bounds.Dy()) / scale}
					for i := range want {
						if math.Abs(want[i]-x.Frame[i]) > .001 {
							return g, nil, fmt.Errorf("grid AX atomic frame mismatch: %s got%v want%v", cell.ID, x.Frame, want)
						}
					}
				}
				if actual != visible || visible < perfMinCells || host.RowCount != g.Stamp.Rows || host.ColumnCount != g.Columns {
					return g, nil, fmt.Errorf("grid AX bounded visible cardinality: native%d measured%d", actual, visible)
				}
				return g, byDomain, nil
			}
		}
		select {
		case <-ctx.Done():
			return g, nil, fmt.Errorf("grid AX readiness: %w; query%v", ctx.Err(), err)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (p *perfProgram) checkGridAccessibility(ctx context.Context, metrics *cogent.FrameMetrics) error {
	g, before, err := gridAccessibilityReady(ctx, p)
	if err != nil {
		return err
	}
	if err := verifyPerfCells(p.source, g); err != nil {
		return err
	}
	reused := 0
	for i, delta := range [][2]int{{0, 12}, {0, 12}, {96, 0}, {0, -12}, {-96, 0}, {0, -12}} {
		_, _, err = p.scroll(ctx, metrics, delta[0], delta[1], uint64(i+1))
		if err != nil {
			return err
		}
		g, after, err := gridAccessibilityReady(ctx, p)
		if err != nil {
			return err
		}
		if err := verifyPerfCells(p.source, g); err != nil {
			return err
		}
		for id, previous := range before {
			if previous.Role != "AXCell" {
				continue
			}
			if current, found := after[id]; found {
				if current.ObjectIdentity != previous.ObjectIdentity {
					return fmt.Errorf("grid AX recreated overlapping cell %s", id)
				}
				reused++
			}
		}
		before = after
	}
	if reused < perfMinCells {
		return fmt.Errorf("grid AX reuse fixture too small: %d", reused)
	}
	// Replacing the binding under the same textual Node.ID must create fresh
	// objects, even for the same domain cell IDs and unchanged source values.
	oldIncarnation := before["data"].GridIncarnation
	err = p.owned(func() error {
		state, err := p.grid.State()
		if err != nil {
			return err
		}
		p.grid, err = virtual.NewGrid(p.projection, perfColumns(), state.Config, state.Viewport)
		if err != nil {
			return err
		}
		return p.owner.Invalidate("perf")
	})
	if err != nil {
		return err
	}
	g, after, err := gridAccessibilityReady(ctx, p)
	if err != nil {
		return err
	}
	if g.Incarnation == oldIncarnation {
		return errors.New("grid AX replacement reused binding incarnation")
	}
	replaced := 0
	for id, old := range before {
		if old.GridScope == "" {
			continue
		}
		if current, found := after[id]; found {
			if current.ObjectIdentity == old.ObjectIdentity {
				return fmt.Errorf("grid AX replacement resurrected native object %s", id)
			}
			replaced++
		}
	}
	if replaced < perfMinCells {
		return fmt.Errorf("grid AX replacement fixture too small: %d", replaced)
	}
	var target cogent.AccessibilityElement
	var targetRow, targetColumn string
	for _, cell := range g.Cells {
		if !cell.Bounds.Empty() {
			target = after[cell.ID]
			targetRow, targetColumn = cell.RowKey, cell.ColumnID
			break
		}
	}
	accepted, err := cogent.PerformAccessibilityAction(ctx, "perf", target.ID, "press", "")
	if err != nil || !accepted {
		return fmt.Errorf("grid AX current cell press refused: %t %v", accepted, err)
	}
	var selected bool
	for !selected {
		err = p.owned(func() error {
			state, e := p.grid.State()
			selected = state.Active.RowKey == targetRow && state.Active.ColumnID == targetColumn && state.Generation > g.Generation
			return e
		})
		if err != nil {
			return err
		}
		if selected {
			break
		}
		select {
		case <-ctx.Done():
			return errors.New("grid AX press did not reach model before deadline")
		case <-time.After(5 * time.Millisecond):
		}
	}
	_, _, err = gridAccessibilityReady(ctx, p)
	if err != nil {
		return err
	}
	return perfEmit("grid.accessibility.pass", map[string]any{"rows": g.Stamp.Rows, "visible_minimum": perfMinCells, "overlap_objects_reused": reused, "replacement_objects_retired": replaced, "native_press_admitted": accepted, "scope": "actual Cocoa stored-grid semantics/geometry/identity; not screen-reader or performance qualification"})
}

func runGridAccessibilitySmoke(rows int, variable bool) error {
	// Keep this separate acceptance path as isolated as ordinary perf mode.
	restoreGeometry, err := isolatePerfGeometry()
	if err != nil {
		return err
	}
	defer restoreGeometry()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	p, _, err := newPerfProgram(ctx, perfConfig{Rows: rows, Variable: variable, Samples: 120, Warmup: 128})
	if err != nil {
		return err
	}
	defer p.owner.Close()
	metrics, err := cogent.NewFrameMetrics(16)
	if err != nil {
		return err
	}
	failure := make(chan error, 1)
	driver := cogent.Driver{FrameMetrics: metrics, Theme: ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15}), Foreground: ui.Some(color.NRGBA{R: 20, G: 20, B: 20, A: 255}), Background: ui.Some(color.NRGBA{R: 250, G: 250, B: 250, A: 255})}}, OnError: func(e error) {
		select {
		case failure <- e:
		default:
		}
		cancel()
	}}
	result := make(chan error, 1)
	go func() { result <- p.checkGridAccessibility(ctx, metrics); p.owner.Close() }()
	native := p.owner.Run(ctx, driver)
	var checked error
	select {
	case checked = <-result:
	case <-time.After(2 * time.Second):
		checked = errors.New("grid AX worker did not complete before native closure")
	}
	select {
	case err := <-failure:
		native = errors.Join(native, err)
	default:
	}
	if result := errors.Join(native, checked); result != nil {
		return result
	}
	if _, err := cogent.InspectAccessibility(context.Background(), "perf"); err == nil {
		return errors.New("grid AX closed attachment remained queryable")
	}
	return perfEmit("grid.accessibility.closed", map[string]any{"attachment_query_refused": true, "geometry_persistence_disabled_in_fixture": core.DebugSettings.DisableWindowGeometrySaver})
}
