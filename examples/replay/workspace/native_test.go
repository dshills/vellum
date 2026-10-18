package workspace

import (
	"cogentcore.org/core/events/key"
	"context"
	"encoding/json"
	"image"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/replay"
	"github.com/dshills/vellum/ui"
)

func nativeGolden(t *testing.T, rows int, index int) (Capture, NativeCapture) {
	t.Helper()
	all, e := ExpectedCaptures(rows)
	if e != nil {
		t.Fatal(e)
	}
	c := all[index]
	var want Capture
	if e = json.Unmarshal(c.Data, &want); e != nil {
		t.Fatal(e)
	}
	got := NativeCapture{Sequence: c.Sequence, Label: c.Label, Model: want.Model}
	for _, w := range want.Windows {
		snap := cogent.DesktopSnapshot{Frame: 1, Size: w.Bounds.Size(), LogicalDPI: 160, Grids: w.Grids, Trees: w.Trees}
		// Immutable software goldens deliberately normalize frame/generation tokens.
		// These are synthetic oracle controls, never native composition evidence.
		for i := range snap.Grids {
			snap.Grids[i].Frames = 1
			snap.Grids[i].Generation = 1
		}
		for i := range snap.Trees {
			snap.Trees[i].Frames = 1
			snap.Trees[i].Generation = 1
		}
		for _, el := range w.Elements {
			if el.Kind == ui.Canvas2D && el.Visible && !strings.Contains(el.ID, "canvas-object:") {
				snap.Canvases = append(snap.Canvases, cogent.DesktopCanvasInspection{CanvasInspection: cogent.CanvasInspection{ID: el.ID, Generation: 1, View: want.Model.Canvas.View, Bounds: [4]float32{float32(el.Bounds.Min.X), float32(el.Bounds.Min.Y), float32(el.Bounds.Dx()), float32(el.Bounds.Dy())}}, Objects: want.Model.Canvas.Objects, Selected: want.Model.Canvas.Selected, Preview: func() placement.State {
					if c.Label == "placement-preview" {
						return placement.State{Active: true, Dragging: true, Accepted: true, ItemID: "block", TargetID: "drawing", Position: canvas.Point{X: 40, Y: -20}, Preview: canvas.Rect{Width: 80, Height: 50}}
					}
					return placement.State{}
				}()})
			}
			if el.Visible {
				v, _ := json.Marshal(el.Value)
				snap.Elements = append(snap.Elements, cogent.DesktopElement{ID: el.ID, Value: v, Bounds: el.Bounds, Focused: el.Focused, Selected: el.Selected, Enabled: el.Enabled})
			}
		}
		got.Windows = append(got.Windows, NativeWindowCapture{ID: w.ID, Snapshot: snap})
	}
	return want, got
}

func TestNativeMacSelectAllIntentIsNarrowAndExplicit(t *testing.T) {
	for _, sequence := range []uint64{17, 122} {
		o := replay.Operation{Sequence: sequence, Kind: replay.KeyChord, Key: "a", Modifiers: 2}
		mods, note := nativeMacModifiers(o)
		if note != "mac-select-all" || !key.HasAnyModifier(mods, key.Meta) || key.HasAnyModifier(mods, key.Control) || o.Modifiers != 2 {
			t.Fatalf("Select All intent not translated without changing the record: %v %q", mods, note)
		}
	}
	for _, o := range []replay.Operation{{Sequence: 16, Kind: replay.KeyChord, Key: "a", Modifiers: 2}, {Sequence: 17, Kind: replay.KeyDown, Key: "a", Modifiers: 2}, {Sequence: 17, Kind: replay.KeyChord, Key: "a"}} {
		mods, note := nativeMacModifiers(o)
		if mods != transportModifiers(o.Modifiers) || note != "" {
			t.Fatal("unrelated physical chord was changed")
		}
	}
}
func TestNativeRecordOracleRetainsAllAuthoredCaptures(t *testing.T) {
	for _, rows := range []int{10000, MaxRows} {
		record, e := Record(rows)
		if e != nil {
			t.Fatal(e)
		}
		captures, e := ExpectedCaptures(rows)
		if e != nil {
			t.Fatal(e)
		}
		if len(record.Operations) != 216 || len(captures) != 27 {
			t.Fatal("immutable record changed")
		}
		for index := range captures {
			want, got := nativeGolden(t, rows, index)
			if e = compareNativeCapture(want, got); e != nil {
				t.Fatalf("%d/%s: %v", rows, got.Label, e)
			}
		}
	}
}
func TestNativeRecordDomainFocusAndGeometryFailures(t *testing.T) {
	want, got := nativeGolden(t, 10000, 6)
	got.Model.Name = "unexpected"
	if e := compareNativeCapture(want, got); e == nil || !strings.Contains(e.Error(), "Name") {
		t.Fatalf("changed model accepted: %v", e)
	}
	want, got = nativeGolden(t, 10000, 6)
	for i := range got.Windows {
		for j := range got.Windows[i].Snapshot.Elements {
			got.Windows[i].Snapshot.Elements[j].Focused = false
		}
	}
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("missing native focus accepted")
	}
	want, got = nativeGolden(t, 10000, 6)
	got.Windows[0].Snapshot.Grids[0].Bounds.Max.X += 40
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("inconsistent viewport geometry accepted")
	}
	want, got = nativeGolden(t, 10000, 6)
	got.Windows = got.Windows[:1]
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("missing window accepted")
	}
}
func TestNativeTargetNamespaceAndWorkflowRefusals(t *testing.T) {
	snap := cogent.DesktopSnapshot{Elements: []cogent.DesktopElement{{ID: "workbench:ZG9jaw:bm90ZXM:node:ZWRpdG9y"}}}
	if _, e := nativeTarget(snap, "workbench:ZG9jaw:bm90ZXM:editor"); e != nil {
		t.Fatal(e)
	}
	if _, e := nativeTarget(snap, "workbench:broken"); e == nil {
		t.Fatal("malformed target accepted")
	}
	if _, e := NativeWorkflow(context.Background(), nil, nil, nil, 10000); e != replay.ErrInvalid {
		t.Fatal(e)
	}
}

func TestNativeAnchorsUseCurrentCompositeAllocation(t *testing.T) {
	// The native client and independently allocated panes deliberately have
	// unrelated ratios to the authored 1600x1000 fixture.
	s := cogent.DesktopSnapshot{Size: image.Pt(3200, 1800), LogicalDPI: 240}
	s.Grids = []cogent.GridInspection{{Bounds: image.Rect(900, 200, 1400, 900), Cells: []cogent.GridCellGeometry{{Column: 0, Bounds: image.Rect(900, 228, 964, 252)}}}}
	s.Trees = []cogent.TreeInspection{{Rows: []cogent.TreeRowGeometry{{Key: "node:000000", DisclosureBounds: image.Rect(74, 220, 92, 244)}}}}
	s.Elements = []cogent.DesktopElement{{ID: cogent.ManualWorkbenchTabID("dock", "notes"), Enabled: true, Bounds: image.Rect(1100, 80, 1200, 130)}, {ID: cogent.ManualWorkbenchTabID("dock", "canvas"), Enabled: true, Bounds: image.Rect(800, 80, 920, 130)}}
	view := canvas.Viewport{Width: 600, Height: 300, Zoom: 2, Center: canvas.Point{X: 10, Y: 15}}
	s.Canvases = []cogent.DesktopCanvasInspection{{CanvasInspection: cogent.CanvasInspection{ID: "workbench:ZG9jaw:Y2FudmFz:drawing", View: view, Bounds: [4]float32{1000, 500, 900, 450}}}}
	a := nativeAnchors{}
	for _, tc := range []struct {
		seq   uint64
		point image.Point
	}{{56, image.Pt(932, 240)}, {63, image.Pt(1150, 550)}, {89, image.Pt(83, 232)}, {98, image.Pt(83, 232)}, {150, image.Pt(1150, 105)}, {152, image.Pt(1005, 105)}, {153, image.Pt(860, 105)}, {154, image.Pt(860, 105)}, {163, image.Pt(1516, 608)}, {167, image.Pt(1516, 608)}, {175, image.Pt(1516, 608)}, {206, image.Pt(1516, 608)}} {
		got, e := a.point(replay.Operation{Sequence: tc.seq}, s)
		if e != nil || got != tc.point {
			t.Fatalf("anchor %d = %v %v want %v", tc.seq, got, e, tc.point)
		}
	}
	if _, e := a.point(replay.Operation{Sequence: 999}, s); e == nil {
		t.Fatal("unknown absolute point silently scaled")
	}
	s.Canvases[0].Bounds[2] -= 20
	if _, e := a.point(replay.Operation{Sequence: 163}, s); e == nil {
		t.Fatal("clipped canvas anchor accepted")
	}
}
func TestNativeCanvasPublicationAndPreviewAreRequired(t *testing.T) {
	want, got := nativeGolden(t, 10000, 0)
	got.Windows[0].Snapshot.Canvases = nil
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("absent canvas publication accepted")
	}
	want, got = nativeGolden(t, 10000, 18)
	got.Windows[0].Snapshot.Canvases[0].Preview = placement.State{}
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("missing completed placement preview accepted")
	}
	want, got = nativeGolden(t, 10000, 19)
	got.Windows[0].Snapshot.Canvases[0].Objects = nil
	if e := compareNativeCapture(want, got); e == nil {
		t.Fatal("unpainted model object accepted")
	}
}

func TestNativeSequentialGridAnchorReference(t *testing.T) {
	for _, rows := range []int{10000, MaxRows} {
		captures, err := ExpectedCaptures(rows)
		if err != nil {
			t.Fatal(err)
		}
		var same, native nativeGridReference
		for index, c := range captures {
			want, _ := nativeGolden(t, rows, index)
			view := want.Model.Grid.Viewport
			if _, y, err := same.offsets(c.Sequence, want.Model, view); err != nil || y != view.ScrollY {
				t.Fatalf("software %d/%s: %d %v", rows, c.Label, y, err)
			}
			if view.Height == 480 {
				view.Height = 482
			}
			_, y, err := native.offsets(c.Sequence, want.Model, view)
			if err != nil {
				t.Fatal(err)
			}
			switch c.Label {
			case "sorted":
				expected := want.Model.Grid.Viewport.ScrollY
				if rows == 10000 {
					expected -= 2
				}
				if y != expected {
					t.Fatalf("sorted anchor %d: got %d want %d", rows, y, expected)
				}
			case "grid-scrolled-selected":
				expected := want.Model.Grid.Viewport.ScrollY
				if rows == 10000 {
					expected -= 2
				}
				if y != expected {
					t.Fatalf("wheel anchor %d: got %d want %d", rows, y, expected)
				}
			case "filtered":
				if y != want.Model.Grid.Viewport.ScrollY-2 {
					t.Fatal("filtered anchor/clamp incorrect")
				}
			}
		}
	}
	// Stable key survives reorder; missing key preserves its old offset and
	// independently clamps. This is not equal-distance-from-bottom normalization.
	if y := nativeReferenceReplace([]int{1, 2, 3}, []int{3, 1, 2}, 29); y != 53 {
		t.Fatal(y)
	}
	if y := nativeReferenceReplace([]int{1, 2, 3}, []int{1, 3}, 29); y != 29 {
		t.Fatal(y)
	}
	want, got := nativeGolden(t, 10000, 0)
	got.Model.Grid.Viewport.ScrollY = 1
	var bad nativeGridReference
	if err := bad.compare(want, got); err == nil {
		t.Fatal("wrong native scroll accepted")
	}
	var inconsistent nativeGridReference
	want.Model.Grid.Viewport.ScrollY = 1
	if _, _, err := inconsistent.offsets(got.Sequence, want.Model, got.Model.Grid.Viewport); err == nil {
		t.Fatal("wrong immutable source offset accepted")
	}
}
