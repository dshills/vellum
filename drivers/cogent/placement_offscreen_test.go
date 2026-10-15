//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

type placementFixture struct {
	alternate                                                    *placement.Source
	tabbed, hidden, modal                                        bool
	manual                                                       *Manual
	app                                                          *app.Application
	controller                                                   *placement.Controller[string]
	target                                                       *placement.Target
	model                                                        *canvas.Scene
	metadata                                                     []ui.CanvasObject
	calls, keyboard, unrelated                                   int
	reject, hideSource, hideTarget, disabled, cancelSnap, scroll bool
}

func newPlacementFixture(t *testing.T, batch int, dpi float32) *placementFixture {
	t.Helper()
	f := &placementFixture{}
	a, err := app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	f.app = a
	t.Cleanup(a.Close)
	f.model, err = canvas.New([]canvas.Rect{{ID: "existing", X: -100, Y: -100, Width: 200, Height: 200, Color: color.RGBA{R: 80, G: 80, B: 80, A: 255}}})
	if err != nil {
		t.Fatal(err)
	}
	f.model.Selected = "existing"
	f.model.View.Zoom = 1.5
	f.model.View.Center = canvas.Point{X: 40, Y: 70}
	f.controller, err = placement.New(placement.Config[string]{Items: []placement.Item[string]{{ID: "box", Label: "Box", Payload: "box-payload", Preview: canvas.Rect{Width: 40, Height: 24, Color: color.RGBA{R: 40, G: 120, B: 220, A: 255}}}}, Command: func(_ context.Context, r placement.Request[string]) error {
		if r.Payload != "box-payload" || r.ItemID != "box" || r.TargetID != "canvas" {
			return errors.New("typed placement corrupted")
		}
		f.calls++
		f.model.Objects = append(f.model.Objects, canvas.Rect{ID: fmt.Sprintf("box-%d", f.calls), X: r.Position.X, Y: r.Position.Y, Width: 40, Height: 24, Color: color.RGBA{R: 40, G: 120, B: 220, A: 255}})
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(f.controller.Close)
	f.target, err = f.controller.NewTarget(placement.TargetConfig[string]{ID: "canvas", Snap: func(r placement.Request[string]) (canvas.Point, error) {
		if f.cancelSnap {
			f.controller.Source("box").Cancel()
		}
		return r.Position, nil
	}, Accept: func(placement.Request[string]) error {
		if f.reject {
			return placement.ErrRejected
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = a.Open(app.Window{ID: "main", Title: "Palette placement", Width: 1000, Height: 700, View: func() ui.Node {
		sources := []ui.Node{{ID: "heading", Kind: ui.Label, Text: "Tools"}}
		if !f.hideSource {
			source := f.controller.Source("box")
			if f.alternate != nil {
				source = f.alternate
			}
			sources = append(sources, ui.Node{ID: "box", Kind: ui.Button, Text: "Box", Disabled: f.disabled, PlacementSource: source, Activate: func() {
				f.keyboard++
				if err := f.controller.Source("box").PlaceKeyboard(a.Context(), f.target, f.model.View.Center); err != nil {
					t.Error(err)
				}
				a.Invalidate("main")
			}})
		}
		target := ui.Node{ID: "drawing", Kind: ui.Canvas2D, Canvas: f.model, CanvasObjects: f.metadata, PlacementTarget: f.target, CanvasGrid: 20, Grow: true, Disabled: f.disabled, Accessibility: ui.Accessibility{Name: "Drawing canvas"}}
		if f.hideTarget {
			target = ui.Node{ID: "removed", Kind: ui.Label, Text: "Target removed"}
		}
		if f.scroll {
			target.MinHeight = 1000
			target = ui.Node{ID: "scroller", Kind: ui.Scroll, Grow: true, Children: []ui.Node{target}}
		}
		palette := ui.Node{ID: "palette", Kind: ui.Column, MinWidth: 120, Children: sources}
		if f.tabbed {
			selected := "shown"
			if f.hidden {
				selected = "hidden"
			}
			palette = ui.Node{ID: "palette-tabs", Kind: ui.Tabs, SelectedID: selected, Options: []ui.Option{{ID: "shown", Label: "Tools"}, {ID: "hidden", Label: "Other"}}, Children: []ui.Node{{ID: "shown", Kind: ui.Column, Children: []ui.Node{palette}}, {ID: "hidden", Kind: ui.Label, Text: "Hidden tools"}}}
		}
		return ui.Node{ID: "root", Kind: ui.Row, Grow: true, Children: []ui.Node{palette, target, {ID: "modal", Kind: ui.Dialog, Text: "Blocked", Open: f.modal, Children: []ui.Node{{ID: "modal-label", Kind: ui.Label, Text: "Blocked"}}}}}
	}})
	if err != nil {
		t.Fatal(err)
	}
	config := manualConfig(t, batch)
	config.LogicalDPI = dpi
	f.manual, err = NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.manual.Close(testContext); err != nil {
			t.Error(err)
		}
		placementWindows.Lock()
		defer placementWindows.Unlock()
		if len(placementWindows.windows) != 0 || len(placementWindows.models) != 0 {
			t.Error("native placement registry retained bindings", len(placementWindows.windows), len(placementWindows.models))
		}
	})
	manualDockSettle(t, f.manual, batch)
	for _, binding := range placementWindows.windows[placementKey{a, "main"}].bindings {
		if binding.model != nil {
			binding.r.widget.AsWidget().On(events.MouseUp, func(events.Event) { f.unrelated++ })
			binding.r.widget.AsWidget().OnClick(func(events.Event) { f.unrelated++ })
		}
	}
	return f
}
func (f *placementFixture) frame(t *testing.T) *placementFrame {
	t.Helper()
	for _, b := range placementWindows.windows[placementKey{f.app, "main"}].bindings {
		if b.model != nil {
			frame := b.frame.Load()
			if frame != nil {
				return frame
			}
		}
	}
	t.Fatal("canvas has no completed frame geometry")
	return nil
}
func (f *placementFixture) input(t *testing.T, batch int, typ events.Types, p image.Point, code key.Codes) {
	t.Helper()
	w, _ := manualDockElement(t, f.manual, "main", "root")
	if err := f.manual.Input(testContext, ManualInput{Window: "main", WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: p, Code: code}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, f.manual, batch)
}
func (f *placementFixture) start(t *testing.T, batch int) (image.Point, canvas.Point) {
	t.Helper()
	_, source := manualDockElement(t, f.manual, "main", "box")
	f.input(t, batch, events.MouseDown, source.Bounds.Min.Add(source.Bounds.Size().Div(2)), 0)
	frame := f.frame(t)
	g := frame.geometry
	// A fixed location inside the actual content, beyond its center in X and
	// above it in Y. Expected world geometry is asserted independently below.
	p := image.Pt(int(math.Round(g.Origin.X+.6*g.View.Width*g.Scale)), int(math.Round(g.Origin.Y+.4*g.View.Height*g.Scale)))
	want := canvas.Point{X: g.View.Center.X + (float64(p.X)-g.Origin.X-g.View.Width*g.Scale/2)/(g.View.Zoom*g.Scale), Y: g.View.Center.Y + (float64(p.Y)-g.Origin.Y-g.View.Height*g.Scale/2)/(g.View.Zoom*g.Scale)}
	f.input(t, batch, events.MouseMove, p, 0)
	if !f.controller.State().Dragging || !f.controller.State().Accepted {
		t.Fatal("actual pointer failed to show accepted preview", f.controller.State())
	}
	if _, err := f.manual.Capture(testContext, "main"); err != nil {
		t.Fatal("preview has no completed software image", err)
	}
	return p, want
}
func TestPlacementActualRoutingScaledAcrossBatches(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		for _, dpi := range []float32{96, 160, 240} {
			t.Run(fmt.Sprintf("%d-%g", batch, dpi), func(t *testing.T) {
				f := newPlacementFixture(t, batch, dpi)
				p, want := f.start(t, batch)
				f.input(t, batch, events.MouseUp, p, 0)
				if f.calls != 1 || f.keyboard != 0 || f.unrelated != 0 || f.model.Selected != "existing" || len(f.model.Objects) != 2 || f.controller.State().Active {
					t.Fatal("release committed twice, activated a click or retained a gesture", f.calls, f.keyboard, f.unrelated, f.model.Selected, f.controller.State())
				}
				actual := f.model.Objects[1]
				if math.Abs(actual.X-want.X) > 1e-8 || math.Abs(actual.Y-want.Y) > 1e-8 {
					t.Fatal("native pan/zoom/display coordinates changed", actual, want)
				}
				f.input(t, batch, events.MouseUp, p, 0)
				if f.calls != 1 {
					t.Fatal("duplicate release placed twice")
				}
			})
		}
	}
}
func TestPlacementRoutedCancellationAndRetirement(t *testing.T) {
	for _, variant := range []string{"escape", "outside", "rejected", "source-removed", "target-removed", "disabled-aba", "viewport-stale", "close"} {
		t.Run(variant, func(t *testing.T) {
			f := newPlacementFixture(t, 7, 160)
			p, _ := f.start(t, 7)
			switch variant {
			case "escape":
				f.input(t, 7, events.KeyChord, image.Point{}, key.CodeEscape)
			case "outside":
				p = image.Pt(2, 2)
			case "rejected":
				f.app.Do(func() { f.reject = true })
			case "source-removed":
				f.app.Do(func() { f.hideSource = true; f.app.Invalidate("main") })
				manualDockSettle(t, f.manual, 7)
			case "target-removed":
				f.app.Do(func() { f.hideTarget = true; f.app.Invalidate("main") })
				manualDockSettle(t, f.manual, 7)
			case "disabled-aba":
				for _, disabled := range []bool{true, false} {
					f.app.Do(func() { f.disabled = disabled; f.app.Invalidate("main") })
					manualDockSettle(t, f.manual, 7)
				}
			case "viewport-stale":
				f.app.Do(func() { f.model.View.Center.X += 500 }) // deliberately unpainted domain change
			case "close":
				f.app.CloseWindow("main")
				if err := f.manual.Close(testContext); err != nil {
					t.Fatal(err)
				}
			}
			if variant != "close" {
				f.input(t, 7, events.MouseUp, p, 0)
			}
			if f.calls != 0 || f.keyboard != 0 || f.unrelated != 0 || len(f.model.Objects) != 1 || f.controller.State().Active {
				t.Fatal("canceled native drop changed model or retained state", f.calls, f.keyboard, f.unrelated, f.controller.State())
			}
		})
	}
}

func TestPlacementCancellationReservedAgainstFullQueue(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	p, _ := f.start(t, 7)
	if f.app.Pending().Reserved != 1 {
		t.Fatal("gesture missing cancellation reservation", f.app.Pending())
	}
	for {
		err := f.app.Post(func() {})
		if errors.Is(err, app.ErrQueueFull) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	// Invoke the actual native loss seam before draining the saturated owner.
	w := placementWindows.windows[placementKey{f.app, "main"}]
	w.active.scene.Events.CancelPointerCapture()
	if f.app.Pending().ReadyCompletions != 1 {
		t.Fatal("loss cancellation was not admitted", f.app.Pending())
	}
	manualDockSettle(t, f.manual, 7)
	f.input(t, 7, events.MouseUp, p, 0)
	if f.controller.State().Active || f.calls != 0 || f.unrelated != 0 || f.app.Pending().Reserved != 0 {
		t.Fatal("loss retained state", f.controller.State(), f.app.Pending())
	}
}
func TestPlacementCallbackCancellationRefundsReservation(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	f.cancelSnap = true
	_, source := manualDockElement(t, f.manual, "main", "box")
	f.input(t, 7, events.MouseDown, source.Bounds.Min.Add(source.Bounds.Size().Div(2)), 0)
	g := f.frame(t).geometry
	p := image.Pt(int(g.Origin.X+g.View.Width*g.Scale/2), int(g.Origin.Y+g.View.Height*g.Scale/2))
	f.input(t, 7, events.MouseMove, p, 0)
	if f.controller.State().Active || f.app.Pending().Reserved != 0 {
		t.Fatal("callback cancellation leaked credit", f.app.Pending())
	}
	f.input(t, 7, events.MouseUp, p, 0)
	if f.calls != 0 || f.unrelated != 0 {
		t.Fatal("canceled callback placed or propagated")
	}
}
func TestPlacementReservationRefusalConsumesRelease(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	var tickets []*app.Completion
	for {
		c, err := f.app.ReserveCompletion("main", nil)
		if errors.Is(err, app.ErrCompletionCapacity) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		tickets = append(tickets, c)
	}
	defer func() {
		for _, c := range tickets {
			c.Abort()
		}
	}()
	_, source := manualDockElement(t, f.manual, "main", "box")
	f.input(t, 7, events.MouseDown, source.Bounds.Min.Add(source.Bounds.Size().Div(2)), 0)
	g := f.frame(t).geometry
	f.input(t, 7, events.MouseUp, image.Pt(int(g.Origin.X+20), int(g.Origin.Y+20)), 0)
	if f.calls != 0 || f.keyboard != 0 || f.unrelated != 0 || f.controller.State().Active || f.app.Pending().Reserved != len(tickets) {
		t.Fatal("refused capture leaked or propagated", f.app.Pending())
	}
}
func TestPlacementViewportFeedbackFullQueueAndNewerFrame(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	var b *placementBinding
	for _, v := range placementWindows.windows[placementKey{f.app, "main"}].bindings {
		if v.model != nil {
			b = v
		}
	}
	for {
		err := f.app.Post(func() {})
		if errors.Is(err, app.ErrQueueFull) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	frame := *b.frame.Load()
	frame.geometry.View.Width += 17
	b.frame.Store(&frame)
	b.requestViewport(&frame)
	if b.viewportPending.Load() != &frame || b.viewportQueued.Load() {
		t.Fatal("full queue lost pending feedback")
	}
	f.app.Do(func() {
		if !b.applyPendingViewport() {
			t.Fatal("owner fallback lost dimensions")
		}
	})
	if f.model.View.Width != frame.geometry.View.Width {
		t.Fatal("wrong dimensions")
	}
	f.app.DrainN(32)
	next := frame
	next.geometry.View.Width += 31
	b.frame.Store(&next)
	b.requestViewport(&next)
	// A second render replaces the queued frame; the callback must apply latest.
	latest := next
	latest.geometry.View.Width += 9
	b.frame.Store(&latest)
	b.requestViewport(&latest)
	f.app.DrainN(32)
	if b.viewportPending.Load() != nil || b.viewportQueued.Load() || f.model.View.Width != latest.geometry.View.Width {
		t.Fatal("newer viewport stranded")
	}
	manualDockSettle(t, f.manual, 7)
}
func TestPlacementClosedWindowCanvasAndFeedbackCannotMutate(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	if err := f.app.Open(app.Window{ID: "sibling", Title: "Sibling", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "sibling-root", Kind: ui.Label, Text: "Alive"} }}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, f.manual, 7)
	var b *placementBinding
	for _, v := range placementWindows.windows[placementKey{f.app, "main"}].bindings {
		if v.model != nil {
			b = v
		}
	}
	frame := *b.frame.Load()
	frame.geometry.View.Width += 23
	b.frame.Store(&frame)
	b.requestViewport(&frame)
	before := f.model.View
	selected := f.model.Selected
	if err := f.app.CloseWindow("main"); err != nil {
		t.Fatal(err)
	}
	// Direct native listeners simulate input already in native delivery, after
	// domain closure and before native topology reconciliation.
	event := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
	b.r.widget.AsWidget().Send(events.KeyChord, event)
	f.app.DrainN(32)
	if f.model.View != before || f.model.Selected != selected || b.live() {
		t.Fatal("closed domain accepted canvas input or feedback")
	}
	manualDockSettle(t, f.manual, 7)
}

func TestPlacementNativeScrolledContentAndPreviewPixels(t *testing.T) {
	f := newPlacementFixture(t, 7, 160)
	f.app.Do(func() { f.scroll = true; f.app.Invalidate("main") })
	manualDockSettle(t, f.manual, 7)
	// Set the actual native ancestor scrollbar, then route the pointer through
	// completed clipped geometry. Canvas wheel input intentionally zooms itself.
	var scroller *core.Frame
	var find func(*record)
	find = func(r *record) {
		if r.node.ID == "scroller" {
			scroller = r.widget.(*core.Frame)
		}
		for _, child := range r.children {
			find(child)
		}
	}
	find(f.manual.implementation.find("main").state.root)
	if scroller == nil || !scroller.ScrollDimToContentEnd(math32.Y) {
		t.Fatal("native scroller missing")
	}
	manualDockSettle(t, f.manual, 7)
	g := f.frame(t).geometry
	if g.Origin.Y >= g.Clip.Min.Y {
		t.Fatal("fixture did not scroll unclipped origin", g)
	}
	_, source := manualDockElement(t, f.manual, "main", "box")
	f.input(t, 7, events.MouseDown, source.Bounds.Min.Add(source.Bounds.Size().Div(2)), 0)
	before, err := f.manual.Capture(testContext, "main")
	if err != nil {
		t.Fatal(err)
	}
	p := image.Pt(int((g.Clip.Min.X+g.Clip.Max.X)/2), int((g.Clip.Min.Y+g.Clip.Max.Y)/2))
	want := canvas.Point{X: g.View.Center.X + (float64(p.X)-g.Origin.X-g.View.Width*g.Scale/2)/(g.View.Zoom*g.Scale), Y: g.View.Center.Y + (float64(p.Y)-g.Origin.Y-g.View.Height*g.Scale/2)/(g.View.Zoom*g.Scale)}
	f.input(t, 7, events.MouseMove, p, 0)
	active := placementWindows.windows[placementKey{f.app, "main"}].active
	if active == nil {
		t.Fatal("no native capture")
	}
	if sprite, ok := active.scene.Stage.Main.Sprites.SpriteByName(active.sprite); !ok || !sprite.Active || !sprite.EventBBox.Empty() {
		t.Fatal("drag preview absent or interactive")
	}
	after, err := f.manual.Capture(testContext, "main")
	if err != nil {
		t.Fatal(err)
	}
	changes := 0
	for y := int(g.Clip.Min.Y) + 4; y < int(g.Clip.Max.Y)-4; y++ {
		for x := int(g.Clip.Min.X) + 4; x < int(g.Clip.Max.X)-4; x++ {
			if before.At(x, y) != after.At(x, y) {
				changes++
			}
		}
	}
	if changes < 50 {
		t.Fatal("placement preview did not affect canvas pixels", changes)
	}
	f.input(t, 7, events.MouseUp, p, 0)
	if f.calls != 1 || math.Abs(f.model.Objects[1].X-want.X) > 1e-8 || math.Abs(f.model.Objects[1].Y-want.Y) > 1e-8 {
		t.Fatal("native scrolling changed placement", f.model.Objects, want)
	}
	if _, ok := active.scene.Stage.Main.Sprites.SpriteByName(active.sprite); ok {
		t.Fatal("terminal drag sprite retained")
	}
}

func TestPlacementQueuedActivationKeepsOriginLease(t *testing.T) {
	for _, route := range []string{"keyboard", "semantic"} {
		for _, variant := range []string{"removed", "hidden-aba", "disabled-aba", "modal-aba", "replaced"} {
			t.Run(route+"-"+variant, func(t *testing.T) {
				f := newPlacementFixture(t, 7, 160)
				f.app.Do(func() { f.tabbed = true; f.app.Invalidate("main") })
				manualDockSettle(t, f.manual, 7)
				native := f.manual.implementation.find("main")
				var source *record
				var find func(*record)
				find = func(r *record) {
					if r.node.ID == "box" {
						source = r
					}
					for _, c := range r.children {
						find(c)
					}
				}
				find(native.state.root)
				bridge := newAccessibilityBridge(f.app)
				bridge.bind("main")
				defer bridge.close()
				publish := func() {
					nodes, handlers := accessibilitySnapshot(native.state.root, bridge)
					bridge.updateOwned(nodes, handlers)
				}
				publish()
				if route == "keyboard" {
					source.widget.AsWidget().Send(events.Click)
				} else if !accessibilityDeliver(bridge.token, bridge.generation.Load(), accessibilityManagedID("box"), accessibilityAction{Kind: "press"}) {
					t.Fatal("semantic activation refused admission")
				}
				if f.app.Pending().Ordinary == 0 {
					t.Fatal("activation was not queued")
				}
				refresh := func() {
					f.app.Invalidate("main")
					if err := f.manual.implementation.refresh(native); err != nil {
						t.Fatal(err)
					}
					publish()
				}
				switch variant {
				case "removed":
					f.hideSource = true
					refresh()
				case "hidden-aba":
					f.hidden = true
					refresh()
					f.hidden = false
					refresh()
				case "disabled-aba":
					f.disabled = true
					refresh()
					f.disabled = false
					refresh()
				case "modal-aba":
					f.modal = true
					refresh()
					f.modal = false
					refresh()
				case "replaced":
					other, err := placement.New(placement.Config[string]{Items: []placement.Item[string]{{ID: "other", Label: "Other", Preview: canvas.Rect{Width: 20, Height: 20}}}, Command: func(context.Context, placement.Request[string]) error { return nil }})
					if err != nil {
						t.Fatal(err)
					}
					defer other.Close()
					f.alternate = other.Source("other")
					refresh()
				}
				if err := f.app.Drain(); err != nil {
					t.Fatal(err)
				}
				if f.calls != 0 || f.keyboard != 0 || f.controller.State().Active {
					t.Fatal("queued placement survived origin retirement", f.calls, f.keyboard)
				}
				manualDockSettle(t, f.manual, 7)
			})
		}
	}
}
