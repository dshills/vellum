//go:build offscreen

package workspace

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/replay"
)

var authorOutput = flag.String("workspace-record-out", "", "explicitly author new software records into a new directory")
var testContext = context.Background()

type pilot struct {
	t           *testing.T
	b           *Backend
	runner      *replay.Runner
	record      replay.Record
	checkpoints []replay.Checkpoint
}

func newPilot(t *testing.T, rows int) *pilot {
	b, e := NewBackend(testContext, rows, 7)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := b.Close(context.Background()); e != nil {
			t.Error(e)
		}
	})
	runner, e := replay.NewRunner(replay.RunnerConfig{EventBatch: 7, MaxSteps: 100000, MaxCheckpointBytes: 32 << 20}, b)
	if e != nil {
		t.Fatal(e)
	}
	return &pilot{t: t, b: b, runner: runner, record: replay.Record{Manifest: b.Manifest()}}
}
func (p *pilot) window(id string) cogent.ManualWindowSnapshot {
	p.t.Helper()
	s, e := p.b.Driver.Snapshot(testContext)
	if e != nil {
		p.t.Fatal(e)
	}
	for _, w := range s.Windows {
		if w.ID == id {
			return w
		}
	}
	p.t.Fatal("missing window", id)
	return cogent.ManualWindowSnapshot{}
}
func (p *pilot) emit(window string, o replay.Operation) {
	p.t.Helper()
	o.At = p.b.Clock.Now().Sub(p.record.Manifest.Start)
	if o.Kind != replay.Worker {
		if window == "" {
			window = "workspace"
		}
		w := p.window(window)
		o.Window = w.ID
		o.WindowIncarnation = w.Incarnation
	}
	o.Sequence = uint64(len(p.record.Operations)) + 1
	p.record.Operations = append(p.record.Operations, o)
	one := o
	one.Sequence = 1
	result, e := p.runner.Run(testContext, replay.Record{Manifest: p.record.Manifest, Operations: []replay.Operation{one}})
	if e != nil {
		p.t.Fatalf("author operation %d %s window=%s target=%s: %v residual=%+v", o.Sequence, o.Kind, o.Window, o.Target, e, result.Residual)
	}
	for i := range result.Checkpoints {
		result.Checkpoints[i].Sequence = o.Sequence
	}
	p.checkpoints = append(p.checkpoints, result.Checkpoints...)
}
func (p *pilot) frames() {
	p.t.Helper()
	s, e := p.b.Driver.Snapshot(testContext)
	if e != nil {
		p.t.Fatal(e)
	}
	for _, w := range s.Windows {
		p.emit(w.ID, replay.Operation{Kind: replay.Frame})
	}
}
func (p *pilot) capture(label string) {
	p.t.Helper()
	window := "workspace"
	s, e := p.b.Driver.Snapshot(testContext)
	if e != nil {
		p.t.Fatal(e)
	}
	if len(s.Windows) == 1 {
		window = s.Windows[0].ID
	}
	p.emit(window, replay.Operation{Kind: replay.Capture, Label: label})
}
func (p *pilot) element(id string) cogent.ManualElement {
	p.t.Helper()
	w := p.window("workspace")
	for _, e := range w.Elements {
		if e.ID == id {
			if !e.Visible || !e.Enabled || e.Bounds.Empty() {
				p.t.Fatal("target unavailable", id, e)
			}
			return e
		}
	}
	p.t.Fatal("target absent", id)
	return cogent.ManualElement{}
}
func (p *pilot) point(point image.Point) {
	p.t.Helper()
	for _, kind := range []replay.Kind{replay.PointerDown, replay.PointerUp} {
		p.emit("", replay.Operation{Kind: kind, X: float64(point.X), Y: float64(point.Y), Button: 1})
	}
	p.frames()
}
func (p *pilot) click(id string) {
	p.t.Helper()
	e := p.element(id)
	point := e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	p.emit("", replay.Operation{Kind: replay.PointerDown, Target: e.ID, TargetIncarnation: e.Incarnation, X: float64(point.X), Y: float64(point.Y), Button: 1})
	p.emit("", replay.Operation{Kind: replay.PointerUp, X: float64(point.X), Y: float64(point.Y), Button: 1})
	p.frames()
}
func (p *pilot) key(key string, mods uint32) {
	p.t.Helper()
	p.emit("", replay.Operation{Kind: replay.KeyChord, Key: key, Modifiers: mods})
	p.frames()
}
func (p *pilot) worker() {
	p.t.Helper()
	ids := p.b.Executor.PendingIDs()
	if len(ids) != 1 {
		p.t.Fatal("expected one admitted worker", ids)
	}
	p.emit("", replay.Operation{Kind: replay.Worker, WorkerID: uint64(ids[0])})
	p.frames()
}
func (p *pilot) grid() cogent.GridInspection {
	p.t.Helper()
	w := p.window("workspace")
	for _, g := range w.Grids {
		if !g.Bounds.Empty() && len(g.Cells) > 0 {
			return g
		}
	}
	p.t.Fatal("no visible completed grid")
	return cogent.GridInspection{}
}
func (p *pilot) tree() cogent.TreeInspection {
	p.t.Helper()
	w := p.window("workspace")
	for _, tree := range w.Trees {
		if !tree.Bounds.Empty() && len(tree.Rows) > 0 {
			return tree
		}
	}
	p.t.Fatal("no visible completed outline")
	return cogent.TreeInspection{}
}
func (p *pilot) disclose() {
	p.t.Helper()
	tree := p.tree()
	for _, row := range tree.Rows {
		if row.Key == "node:000000" && !row.DisclosureBounds.Empty() {
			p.point(row.DisclosureBounds.Min.Add(row.DisclosureBounds.Size().Div(2)))
			return
		}
	}
	p.t.Fatal("no root disclosure")
}
func (p *pilot) scroll(point image.Point, dx, dy float64) {
	p.t.Helper()
	p.emit("", replay.Operation{Kind: replay.Scroll, X: float64(point.X), Y: float64(point.Y), DeltaX: dx, DeltaY: dy})
	p.frames()
}
func (p *pilot) tab(id string)        { p.t.Helper(); p.click(cogent.ManualWorkbenchTabID("dock", id)) }
func control(panel, id string) string { return cogent.ManualWorkbenchControlID("dock", panel, id) }
func authorCase(t *testing.T, rows int) (replay.Record, []replay.Checkpoint) {
	t.Helper()
	p := newPilot(t, rows)
	p.frames()
	p.capture("initial")
	p.click("open")
	p.capture("open-admitted")
	p.worker()
	p.capture("opened")
	p.click(control("properties", "name"))
	p.key("a", 2)
	for _, r := range "Harbor" {
		p.key(string(r), 0)
	}
	p.key("Enter", 0)
	p.capture("property-edited")
	p.tab("grid")
	p.capture("grid-visible")
	p.click("sort")
	p.capture("sort-admitted")
	p.worker()
	p.capture("sorted")
	g := p.grid()
	var cell cogent.GridCellGeometry
	for _, c := range g.Cells {
		if c.Column == 0 && !c.Bounds.Empty() && c.Bounds.In(g.Bounds) {
			cell = c
			break
		}
	}
	if cell.Bounds.Empty() {
		t.Fatal("no fully visible frozen-column cell")
	}
	p.point(cell.Bounds.Min.Add(cell.Bounds.Size().Div(2)))
	p.key("Down", 0)
	g = p.grid()
	center := g.Bounds.Min.Add(g.Bounds.Size().Div(2))
	p.scroll(center, 128, -480)
	p.capture("grid-scrolled-selected")
	p.click("filter")
	p.capture("filter-admitted")
	p.click("cancel")
	p.capture("cancel-requested")
	p.worker()
	p.capture("cancel-delivered")
	p.click("filter")
	p.worker()
	p.capture("filtered")
	p.disclose()
	p.capture("tree-collapse-admitted")
	p.worker()
	p.capture("tree-collapsed")
	p.disclose()
	p.worker()
	p.capture("tree-expanded")
	p.click("tree-filter")
	p.worker()
	p.capture("tree-filtered")
	p.tab("notes")
	p.click(control("notes", "editor"))
	p.key("a", 2)
	for _, r := range "Notes Ω" {
		p.key(string(r), 0)
	}
	p.key("Enter", 0)
	p.capture("notes-edited")
	from := p.element(cogent.ManualWorkbenchTabID("dock", "notes"))
	to := p.element(cogent.ManualWorkbenchTabID("dock", "canvas"))
	start := from.Bounds.Min.Add(from.Bounds.Size().Div(2))
	end := to.Bounds.Min.Add(to.Bounds.Size().Div(2))
	p.emit("", replay.Operation{Kind: replay.PointerDown, X: float64(start.X), Y: float64(start.Y), Button: 1})
	p.emit("", replay.Operation{Kind: replay.Advance, Delta: core.TimingSettings.SlideStartTime + time.Millisecond})
	p.emit("", replay.Operation{Kind: replay.PointerMove, X: float64((start.X + end.X) / 2), Y: float64((start.Y + end.Y) / 2)})
	p.emit("", replay.Operation{Kind: replay.PointerMove, X: float64(end.X), Y: float64(end.Y)})
	p.emit("", replay.Operation{Kind: replay.PointerUp, X: float64(end.X), Y: float64(end.Y), Button: 1})
	p.frames()
	p.capture("tabs-reordered")
	p.tab("canvas")
	tool := p.element(control("canvas", "tool"))
	drawing := p.element(control("canvas", "drawing"))
	start = tool.Bounds.Min.Add(tool.Bounds.Size().Div(2))
	end = drawing.Bounds.Min.Add(drawing.Bounds.Size().Div(2)).Add(image.Pt(32, -24))
	p.emit("", replay.Operation{Kind: replay.PointerDown, Target: tool.ID, TargetIncarnation: tool.Incarnation, X: float64(start.X), Y: float64(start.Y), Button: 1})
	p.emit("", replay.Operation{Kind: replay.PointerMove, X: float64(end.X), Y: float64(end.Y)})
	p.frames()
	p.capture("placement-preview")
	p.emit("", replay.Operation{Kind: replay.PointerUp, X: float64(end.X), Y: float64(end.Y), Button: 1})
	p.frames()
	p.capture("placed")
	p.point(end.Add(image.Pt(20, 20)))
	p.scroll(end, 0, -100)
	p.key("Right", 0)
	p.capture("canvas-navigated")
	p.click("undo")
	p.capture("placement-undone")
	p.click("redo")
	p.capture("placement-redone")
	p.click("undo")
	p.click("undo")
	p.capture("undone")
	p.click("modal")
	p.capture("modal-open")
	p.scroll(end, 0, -100)
	p.key("Escape", 0)
	p.capture("modal-closed")
	p.click("close-main")
	p.capture("surviving-inspector")
	if e := p.record.Validate(replay.Limits{}); e != nil {
		t.Fatal(e)
	}
	return p.record, p.checkpoints
}
func TestAuthorWorkspaceRecords(t *testing.T) {
	if *authorOutput == "" {
		t.Skip("explicit authoring only; acceptance uses frozen records")
	}
	if e := os.Mkdir(*authorOutput, 0755); e != nil {
		t.Fatal(e)
	}
	for _, rows := range []int{10000, MaxRows} {
		t.Run(fmt.Sprint(rows), func(t *testing.T) {
			record, checkpoints := authorCase(t, rows)
			for name, value := range map[string]any{fmt.Sprintf("%d.json", rows): record, fmt.Sprintf("%d.expected.json", rows): checkpoints} {
				data, e := json.MarshalIndent(value, "", "  ")
				if e != nil {
					t.Fatal(e)
				}
				if e := os.WriteFile(filepath.Join(*authorOutput, name), append(data, '\n'), 0644); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
