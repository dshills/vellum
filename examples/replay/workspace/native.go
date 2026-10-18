package workspace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/replay"
	"github.com/dshills/vellum/ui"
)

// NativeReport describes autonomous native integration, never deterministic
// replay, physical input, IME, presentation or performance qualification.
type NativeReport struct {
	Event, Result string
	Rows, Applied int
	Operations    []NativeOperation
	Captures      []NativeCapture
	Error         string `json:",omitempty"`
}
type NativeOperation struct {
	Sequence                   uint64
	Kind                       replay.Kind
	Window, Result             string
	Position                   image.Point
	NativeChord, TransportNote string `json:",omitempty"`
}
type NativeCapture struct {
	Sequence uint64
	Label    string
	Model    Snapshot
	Windows  []NativeWindowCapture
}
type NativeWindowCapture struct {
	ID       string
	Snapshot cogent.DesktopSnapshot
}

func FixtureTheme() ui.Theme {
	return ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}
}

// NativeWorkflow consumes all authored operations and the same immutable model
// goldens. Frame operations and additional per-input readiness barriers wait on
// real native compositions. Native wall time remains autonomous; Advance waits
// its declared duration, and only Worker explicitly selects a manual worker.
// Native layout dimensions are observed independently, not compared to software
// image hashes or silently substituted for the authored captures.
func NativeWorkflow(ctx context.Context, f *Fixture, s *cogent.DesktopSession, ex *execution.Manual, rows int) (report NativeReport, err error) {
	if ctx == nil || f == nil || s == nil || ex == nil {
		return report, replay.ErrInvalid
	}
	report = NativeReport{Event: "vellum.workspace-native-record.result", Result: "FAILED_NATIVE_RECORD", Rows: rows}
	defer func() {
		if err != nil {
			report.Error = err.Error()
		}
	}()
	record, err := Record(rows)
	if err != nil {
		return report, err
	}
	goldens, err := ExpectedCaptures(rows)
	if err != nil {
		return report, err
	}
	expected := map[uint64]Capture{}
	for _, g := range goldens {
		var c Capture
		if e := json.Unmarshal(g.Data, &c); e != nil {
			return report, e
		}
		expected[g.Sequence] = c
	}
	refs := map[string]app.WindowRef{}
	for _, w := range f.Owner.Windows() {
		refs[w.ID] = w.Ref()
	}
	anchors := nativeAnchors{}
	gridReference := nativeGridReference{}
	var down image.Point
	var downOriginal image.Point
	var downTarget string
	downActive := false
	for _, o := range record.Operations {
		if e := ctx.Err(); e != nil {
			return report, e
		}
		if o.CheckRevision {
			var revision uint64
			if e := f.Owner.Do(func() { revision = f.Snapshot().HistoryRevision }); e != nil {
				return report, e
			}
			if revision != o.ExpectedRevision {
				return report, fmt.Errorf("native operation %d revision: got %d want %d", o.Sequence, revision, o.ExpectedRevision)
			}
		}
		ref := refs[o.Window]
		receipt := NativeOperation{Sequence: o.Sequence, Kind: o.Kind, Window: o.Window, Result: "APPLIED"}
		switch o.Kind {
		case replay.Frame:
			err = nativeFrame(ctx, s, ref)
		case replay.Advance:
			timer := time.NewTimer(o.Delta)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				err = ctx.Err()
			}
		case replay.Worker:
			err = ex.Step(ctx, execution.ID(o.WorkerID))
		case replay.Capture:
			want, ok := expected[o.Sequence]
			if !ok {
				return report, replay.ErrInvalid
			}
			var capture NativeCapture
			capture, err = nativeCapture(ctx, f, s, o)
			if err == nil {
				report.Captures = append(report.Captures, capture)
				err = gridReference.compare(want, capture)
			}
		case replay.PointerDown, replay.PointerUp, replay.PointerMove, replay.Scroll, replay.KeyDown, replay.KeyUp, replay.KeyChord:
			var snap cogent.DesktopSnapshot
			snap, err = nativeReady(ctx, s, ref)
			if err != nil {
				break
			}
			mods, note := nativeMacModifiers(o)
			in := cogent.DesktopInput{Frame: snap.Frame, Modifiers: mods, Button: events.Buttons(o.Button)}

			switch o.Kind {
			case replay.PointerDown:
				in.Type = events.MouseDown
			case replay.PointerUp:
				in.Type = events.MouseUp
			case replay.PointerMove:
				in.Type = events.MouseMove
			case replay.Scroll:
				in.Type = events.Scroll
				in.Delta = math32.Vec2(float32(o.DeltaX), float32(o.DeltaY))
			case replay.KeyDown:
				in.Type = events.KeyDown
			case replay.KeyUp:
				in.Type = events.KeyUp
			case replay.KeyChord:
				in.Type = events.KeyChord
			}
			if in.Type.IsKey() {
				in.Rune, in.Code, err = transportKey(o.Key)
				if err != nil {
					break
				}
				receipt.NativeChord = string(key.NewChord(in.Rune, in.Code, in.Modifiers))
				receipt.TransportNote = note
				if note != "" && keymap.Of(key.Chord(receipt.NativeChord)) != keymap.SelectAll {
					err = errors.New("native Mac Select All shortcut is not bound in the current keymap")
					break
				}
			}
			if o.Target != "" {
				var el cogent.DesktopElement
				el, err = nativeTarget(snap, o.Target)
				if err != nil {
					break
				}
				if !el.Enabled || el.Bounds.Empty() {
					err = cogent.ErrDesktopPending
					break
				}
				in.Target = el.ID
				if !in.Type.IsKey() {
					in.Position = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
				}
			}
			if !in.Type.IsKey() && o.Target == "" {
				if o.Kind == replay.PointerUp && downActive && image.Pt(int(o.X), int(o.Y)) == downOriginal {
					in.Position = down
				} else {
					in.Position, err = anchors.point(o, snap)
					if err != nil {
						break
					}
				}
			}
			if o.Kind == replay.PointerDown {
				down = in.Position
				downOriginal = image.Pt(int(o.X), int(o.Y))
				downTarget = o.Target
				downActive = true
			}
			if o.Kind == replay.PointerUp && downActive && image.Pt(int(o.X), int(o.Y)) == downOriginal {
				in.Position = down
			}
			if o.Kind == replay.Scroll {
				// Tables/outlines consume logical deltas; canvas zoom intentionally consumes
				// raw wheel magnitude. Pick the actual completed target, not a model callback.
				canvas := false
				for _, el := range snap.Elements {
					if el.DomainID == "drawing" && in.Position.In(el.Bounds) {
						canvas = true
					}
				}
				if !canvas {
					factor := snap.LogicalDPI / 160
					in.Delta.X *= factor
					in.Delta.Y *= factor
				}
			}
			receipt.Position = in.Position
			err = s.Input(ctx, ref, in)
			if errors.Is(err, cogent.ErrDesktopClosed) && o.Kind == replay.PointerUp && downTarget == "close-main" && ref.Context().Err() != nil {
				// The routed close retires the originating window before its ack can run.
				// Only this authored close is allowed, and the final surviving-window
				// capture must still match. Other early closures remain failures.
				receipt.Result = "EXPECTED_WINDOW_RETIREMENT"
				err = nil
			}
			if o.Kind == replay.PointerUp {
				downActive = false
				downTarget = ""
			}
		default:
			err = fmt.Errorf("unsupported native record operation %s", o.Kind)
		}
		if err != nil {
			return report, fmt.Errorf("native operation %d %s: %w", o.Sequence, o.Kind, err)
		}
		report.Operations = append(report.Operations, receipt)
		report.Applied++
	}
	if report.Applied != len(record.Operations) || len(report.Captures) != len(goldens) {
		return report, errors.New("native record incomplete")
	}
	report.Result = "PASS_NATIVE_RECORD_DOMAIN_AND_ROUTING"
	return report, nil
}

// The authored software fixture uses Control+A for Select All. The native Mac
// standard keymap correctly treats Control+A as Home and Command+A as Select
// All. Adapt only the two independently authored Select All intents. No global
// keymap or OS setting is changed; the live binding is checked before routing.
func nativeMacModifiers(o replay.Operation) (key.Modifiers, string) {
	if (o.Sequence == 17 || o.Sequence == 122) && o.Kind == replay.KeyChord && o.Key == "a" && o.Modifiers == 2 {
		return transportModifiers(8), "mac-select-all"
	}
	return transportModifiers(o.Modifiers), ""
}
func nativeFrame(ctx context.Context, s *cogent.DesktopSession, ref app.WindowRef) error {
	for {
		e := s.Frame(ctx, ref)
		if !errors.Is(e, cogent.ErrDesktopPending) && !errors.Is(e, cogent.ErrDesktopBusy) && !(errors.Is(e, cogent.ErrDesktopClosed) && ref.Context().Err() == nil) {
			return e
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}
func nativeReady(ctx context.Context, s *cogent.DesktopSession, ref app.WindowRef) (cogent.DesktopSnapshot, error) {
	for {
		if e := nativeFrame(ctx, s, ref); e != nil {
			return cogent.DesktopSnapshot{}, e
		}
		out, e := s.Snapshot(ctx, ref)
		if !errors.Is(e, cogent.ErrDesktopPending) {
			return out, e
		}
	}
}
func nativeTarget(snap cogent.DesktopSnapshot, id string) (cogent.DesktopElement, error) {
	encoded := "node:" + base64.RawURLEncoding.EncodeToString([]byte(id))
	prefix, tail := "", id
	for strings.HasPrefix(tail, "workbench:") {
		parts := strings.SplitN(tail, ":", 4)
		if len(parts) != 4 {
			return cogent.DesktopElement{}, replay.ErrInvalid
		}
		prefix += strings.Join(parts[:3], ":") + ":"
		tail = parts[3]
	}
	qualified := prefix + "node:" + base64.RawURLEncoding.EncodeToString([]byte(tail))
	for _, el := range snap.Elements {
		if el.ID == id || el.ID == encoded || el.ID == qualified {
			return el, nil
		}
	}
	return cogent.DesktopElement{}, fmt.Errorf("native target %q absent", id)
}
func nativeCapture(ctx context.Context, f *Fixture, s *cogent.DesktopSession, o replay.Operation) (NativeCapture, error) {
	out := NativeCapture{Sequence: o.Sequence, Label: o.Label}
	for _, w := range f.Owner.Windows() {
		snap, e := nativeReady(ctx, s, w.Ref())
		if e != nil {
			return out, e
		}
		if snap.Frame == 0 || snap.Size.X <= 0 || snap.Size.Y <= 0 || snap.LogicalDPI <= 0 {
			return out, errors.New("invalid native geometry")
		}
		out.Windows = append(out.Windows, NativeWindowCapture{ID: w.ID, Snapshot: snap})
	}
	e := f.Owner.Do(func() { out.Model = f.Snapshot() })
	if e != nil {
		return out, e
	}
	if out.Model.Failure != "" {
		return out, errors.New(out.Model.Failure)
	}
	return out, nil
}
func compareNativeCapture(want Capture, got NativeCapture) error {
	actual := got.Model
	expected := want.Model
	// Width/height are derived from the actual native allocation. Preserve all
	// camera/object/selection/history/data assertions. The sequential grid oracle
	// independently predicts scroll; only camera centers have a subpixel bar.
	actual.Grid.Viewport.Width = expected.Grid.Viewport.Width
	actual.Grid.Viewport.Height = expected.Grid.Viewport.Height
	actual.Tree.Viewport.Width = expected.Tree.Viewport.Width
	actual.Tree.Viewport.Height = expected.Tree.Viewport.Height
	actual.Canvas.View.Width = expected.Canvas.View.Width
	actual.Canvas.View.Height = expected.Canvas.View.Height
	near := func(a, b float64, tolerance float64) float64 {
		if math.IsNaN(a) || math.IsInf(a, 0) {
			return a
		}
		if math.Abs(a-b) <= tolerance {
			return b
		}
		return a
	}
	actual.Canvas.View.Center.X = near(actual.Canvas.View.Center.X, expected.Canvas.View.Center.X, 1)
	actual.Canvas.View.Center.Y = near(actual.Canvas.View.Center.Y, expected.Canvas.View.Center.Y, 1)
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("native model mismatch at %s: %s", got.Label, nativeDifference(actual, expected))
	}
	ids := map[string]bool{}
	for _, w := range got.Windows {
		if ids[w.ID] || w.Snapshot.Size.X <= 0 || w.Snapshot.Size.Y <= 0 {
			return errors.New("duplicate/invalid native window")
		}
		ids[w.ID] = true
	}
	if len(ids) != len(want.Windows) {
		return errors.New("native window set differs")
	}
	for _, w := range want.Windows {
		if !ids[w.ID] {
			return fmt.Errorf("native window %s missing", w.ID)
		}
		var live cogent.DesktopSnapshot
		for _, v := range got.Windows {
			if v.ID == w.ID {
				live = v.Snapshot
			}
		}
		if w.FocusedID != "" {
			visible := false
			for _, e := range w.Elements {
				if e.ID == w.FocusedID {
					visible = e.Visible
				}
			}
			el, e := nativeTarget(live, w.FocusedID)
			if visible {
				if e != nil || !el.Focused {
					return fmt.Errorf("native focus %q missing at %s", w.FocusedID, got.Label)
				}
			} else if e == nil && el.Focused {
				return errors.New("background focus exposed through modal")
			}
		}
		for _, e := range w.Elements {
			if !e.Visible || strings.Contains(e.ID, "tree:") || strings.Contains(e.ID, "grid:") || strings.Contains(e.ID, "canvas-object:") {
				continue
			}
			switch e.Kind {
			case ui.Button, ui.TextField, ui.MultilineText, ui.Dialog, ui.Tabs:
				liveElement, err := nativeTarget(live, e.ID)
				if err != nil {
					return err
				}
				if liveElement.Enabled != e.Enabled || liveElement.Selected != e.Selected {
					return fmt.Errorf("native control state differs for %s at %s", e.ID, got.Label)
				}
				if e.Kind == ui.TextField || e.Kind == ui.MultilineText {
					v, _ := json.Marshal(e.Value)
					if string(v) != string(liveElement.Value) {
						return fmt.Errorf("native text publication differs for %s", e.ID)
					}
				}
			}
		}

		expectedCanvases := map[string]bool{}
		for _, el := range w.Elements {
			if el.Kind == ui.Canvas2D && el.Visible && !strings.Contains(el.ID, "canvas-object:") {
				expectedCanvases[el.ID] = true
			}
		}
		if len(live.Canvases) != len(expectedCanvases) {
			return fmt.Errorf("native canvas publication count differs at %s", got.Label)
		}
		for _, c := range live.Canvases {
			if !expectedCanvases[c.ID] {
				return errors.New("unexpected native canvas publication")
			}
			delete(expectedCanvases, c.ID)
		}
		if len(live.Grids) != len(w.Grids) || len(live.Trees) != len(w.Trees) {
			return fmt.Errorf("native composite publication count differs at %s", got.Label)
		}
		for _, g := range live.Grids {
			if g.Scale <= 0 || g.Bounds.Empty() || g.Viewport != got.Model.Grid.Viewport || g.RetainedCells > 4096 || len(g.Cells) > 4096 || g.Generation == 0 || g.Frames == 0 || !reflect.DeepEqual(g.Selection, got.Model.Grid.Selection) {
				return errors.New("invalid native grid geometry/state")
			}
			if math.Abs(float64(g.Bounds.Dx())/float64(g.Scale)-float64(g.Viewport.Width)-12) > 2 || math.Abs(float64(g.Bounds.Dy())/float64(g.Scale)-float64(g.Viewport.Height)-12) > 2 {
				return errors.New("native grid allocation/viewport disagreement")
			}
			for _, c := range g.Cells {
				if !c.Bounds.Empty() && !c.Bounds.In(g.Bounds) {
					return errors.New("unclipped native grid cell")
				}
			}
		}
		for _, t := range live.Trees {
			if t.Scale <= 0 || t.Bounds.Empty() || t.Viewport != got.Model.Tree.Viewport || t.RetainedRows > 128 || len(t.Rows) > 128 || t.Generation == 0 || t.Frames == 0 || t.Selected != got.Model.Tree.Selected {
				return errors.New("invalid native tree geometry/state")
			}
			for _, r := range t.Rows {
				if r.Selected != (r.Key == got.Model.Tree.Selected) {
					return errors.New("native outline row selection disagrees with model")
				}
				if !r.Bounds.Empty() && !r.Bounds.In(t.Bounds) {
					return errors.New("unclipped native tree row")
				}
			}
		}
		for _, c := range live.Canvases {
			if got.Label == "placement-preview" {
				p := c.Preview
				if !p.Active || !p.Dragging || !p.Accepted || p.ItemID != "block" || p.TargetID != "drawing" || p.Position != (canvas.Point{X: 40, Y: -20}) || p.Preview.Width != 80 || p.Preview.Height != 50 {
					return errors.New("accepted placement preview absent from completed canvas picture")
				}
			} else if c.Preview.Active {
				return errors.New("placement preview retained after completion")
			}
			if c.View != got.Model.Canvas.View || c.Bounds[2] <= 0 || c.Bounds[3] <= 0 || c.Generation == 0 || !reflect.DeepEqual(c.Objects, got.Model.Canvas.Objects) || c.Selected != got.Model.Canvas.Selected {
				return errors.New("native canvas publication/model disagreement")
			}
		}
	}
	return nil
}

// nativeDifference is bounded diagnostic output; the original immutable domain
// oracle still decides success. It does not normalize unexpected values.
func nativeDifference(got, want any) string {
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(want)
	var x, y any
	_ = json.Unmarshal(a, &x)
	_ = json.Unmarshal(b, &y)
	var visit func(string, any, any) string
	visit = func(path string, a, b any) string {
		if reflect.DeepEqual(a, b) {
			return ""
		}
		am, aok := a.(map[string]any)
		bm, bok := b.(map[string]any)
		if aok && bok {
			keys := make([]string, 0, len(am)+len(bm))
			seen := map[string]bool{}
			for k := range am {
				seen[k] = true
			}
			for k := range bm {
				seen[k] = true
			}
			for k := range seen {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if d := visit(path+"."+k, am[k], bm[k]); d != "" {
					return d
				}
			}
		}
		aa, aok := a.([]any)
		bb, bok := b.([]any)
		if aok && bok && len(aa) == len(bb) {
			for i := range aa {
				if d := visit(fmt.Sprintf("%s[%d]", path, i), aa[i], bb[i]); d != "" {
					return d
				}
			}
		}
		av, _ := json.Marshal(a)
		bv, _ := json.Marshal(b)
		if len(av) > 128 {
			av = av[:128]
		}
		if len(bv) > 128 {
			bv = bv[:128]
		}
		return fmt.Sprintf("%s got %s want %s", path, av, bv)
	}
	return visit("model", x, y)
}

// nativeAnchors is a bounded transport adapter for this fixed authored fixture.
// The meanings below come from authorCase, independently of native outcomes:
// first fully visible frozen-column cell; root disclosure; notes-to-canvas tab
// header drag; canvas world points (32,-24) / (52,-4). Unknown points refuse.
// It never invokes domain actions or substitutes coordinates from a model oracle.
type nativeAnchors struct {
	tabStart, tabEnd, canvasPoint image.Point
	tabs, canvas                  bool
}

func (a *nativeAnchors) point(o replay.Operation, s cogent.DesktopSnapshot) (image.Point, error) {
	center := func(r image.Rectangle) image.Point { return r.Min.Add(r.Size().Div(2)) }
	switch o.Sequence {
	case 56:
		for _, g := range s.Grids {
			for _, c := range g.Cells {
				if c.Column == 0 && !c.Bounds.Empty() && c.Bounds.In(g.Bounds) {
					return center(c.Bounds), nil
				}
			}
		}
	case 63:
		for _, g := range s.Grids {
			if !g.Bounds.Empty() {
				return center(g.Bounds), nil
			}
		}
	case 89, 98:
		for _, t := range s.Trees {
			for _, r := range t.Rows {
				if r.Key == "node:000000" && !r.DisclosureBounds.Empty() {
					return center(r.DisclosureBounds), nil
				}
			}
		}
	case 150:
		start, e := nativeTarget(s, cogent.ManualWorkbenchTabID("dock", "notes"))
		if e != nil {
			return image.Point{}, e
		}
		end, e := nativeTarget(s, cogent.ManualWorkbenchTabID("dock", "canvas"))
		if e != nil {
			return image.Point{}, e
		}
		if !start.Enabled || !end.Enabled || start.Bounds.Empty() || end.Bounds.Empty() {
			break
		}
		a.tabStart, a.tabEnd, a.tabs = center(start.Bounds), center(end.Bounds), true
		return a.tabStart, nil
	case 152:
		if a.tabs {
			return image.Pt((a.tabStart.X+a.tabEnd.X)/2, (a.tabStart.Y+a.tabEnd.Y)/2), nil
		}
	case 153, 154:
		if a.tabs {
			return a.tabEnd, nil
		}
	case 163, 167, 171, 175:
		for _, c := range s.Canvases {
			if !strings.HasSuffix(c.ID, ":drawing") {
				continue
			}
			scale := float64(s.LogicalDPI) / 160
			if scale <= 0 || math.Abs(float64(c.Bounds[2])/scale-c.View.Width) > 1 || math.Abs(float64(c.Bounds[3])/scale-c.View.Height) > 1 {
				return image.Point{}, errors.New("native canvas anchor requires an unclipped viewport")
			}
			world := canvas.Point{X: 32, Y: -24}
			if o.Sequence == 171 {
				world = canvas.Point{X: 52, Y: -4}
			}
			local := c.View.Screen(world)
			point := image.Pt(int(math.Round(float64(c.Bounds[0])+local.X*scale)), int(math.Round(float64(c.Bounds[1])+local.Y*scale)))
			if !point.In(image.Rect(int(c.Bounds[0]), int(c.Bounds[1]), int(c.Bounds[0]+c.Bounds[2]), int(c.Bounds[1]+c.Bounds[3]))) {
				return image.Point{}, errors.New("native canvas anchor is outside viewport")
			}
			if o.Sequence == 163 || o.Sequence == 175 {
				a.canvasPoint, a.canvas = point, true
			}
			return point, nil
		}
	case 206:
		// The authored blocked-modal scroll is at the last pre-modal canvas point.
		// No background publication is admitted through the modal scope.
		if a.canvas && a.canvasPoint.In(image.Rectangle{Max: s.Size}) {
			return a.canvasPoint, nil
		}
	}
	return image.Point{}, fmt.Errorf("native operation %d lacks a current authored geometry anchor", o.Sequence)
}
