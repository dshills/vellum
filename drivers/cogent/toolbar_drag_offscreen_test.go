//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image"
	"math"
	"slices"
	"testing"
	"time"
)

func TestToolbarCustomizationDragRouting(t *testing.T) {
	for _, dpi := range []float32{160, 320} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("dpi%g-dark%v", dpi, dark), func(t *testing.T) {
				cfg := manualConfig(t, 7)
				cfg.LogicalDPI = dpi
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				source := fullColorButtonSource(t, false)
				changes, calls := 0, 0
				c := ui.ToolbarCustomizer{ID: "customize", Open: true, Style: cfg.Theme.Kinds[ui.FormGroup], Value: ui.ToolbarLayout{Hidden: []string{"item-2"}}}
				c.Change = func(v ui.ToolbarLayout) { changes++; c.Value = v }
				c.Dismiss = func() { c.Open = false }
				c.Toolbar = ui.Node{ID: "tools", Kind: ui.Toolbar}
				for i := 0; i < 16; i++ {
					c.Toolbar.Children = append(c.Toolbar.Children, ui.Node{ID: fmt.Sprintf("item-%d", i), Kind: ui.Button, Text: fmt.Sprintf("Action %d", i), IconImage: source, Activate: func() { calls++ }})
				}
				view := func() ui.Node {
					panel, e := c.View()
					if e != nil {
						panic(e)
					}
					return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{panel}}
				}
				a, e := app.New(app.Config{})
				if e != nil {
					t.Fatal(e)
				}
				t.Cleanup(a.Close)
				if e = a.Open(app.Window{ID: "one", Title: "Customization drag", Width: 1000, Height: 1400, View: view}); e != nil {
					t.Fatal(e)
				}
				m, e := NewManual(testContext, a, cfg)
				if e != nil {
					t.Fatal(e)
				}
				t.Cleanup(func() { m.Close(testContext) })
				unitSettle(t, m)
				w := m.implementation.find("one")
				root := w.state.root
				bounds := func(id string) image.Rectangle { _, v := imageManualElement(t, m, id); return v.Bounds }
				input := func(typ events.Types, p image.Point) {
					t.Helper()
					if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: p}); e != nil {
						t.Fatal(e)
					}
					unitSettle(t, m)
				}
				grip := func(id string) *toolbarDragBinding {
					return findRecord(root, ui.ToolbarCustomizationDragID(c.ID, id)).control.toolbarDrag
				}
				start := func(id string) image.Point {
					p := bounds(ui.ToolbarCustomizationDragID(c.ID, id))
					at := p.Min.Add(p.Size().Div(2))
					input(events.MouseDown, at)
					if grip(id).active == nil {
						t.Fatal("capture did not start", id)
					}
					return at
				}
				rowBefore := func(id string) image.Point {
					show := findRecord(root, ui.ToolbarCustomizationVisibilityID(c.ID, id)).widget.AsWidget()
					row := core.AsWidget(show.Parent)
					return image.Pt(row.Geom.TotalBBox.Min.X+5, row.Geom.TotalBBox.Min.Y+1).Add(row.Scene.SceneGeom.Pos)
				}
				if grip("item-2").button.Scene.SceneGeom.Pos == (image.Point{}) {
					t.Fatal("fixture must exercise a nonzero modal scene origin")
				}
				end := rowBefore("item-0")
				// A hidden action still has full artwork and can be moved without being shown.
				artRow := core.AsWidget(findRecord(root, ui.ToolbarCustomizationVisibilityID(c.ID, "item-2")).widget.AsTree().Parent)
				foundArt := false
				for _, child := range artRow.Children {
					if im, ok := child.(*managedImage); ok {
						foundArt = true
						if im.source != source {
							t.Fatal("drag row lost full-color artwork")
						}
					}
				}
				if !foundArt {
					t.Fatal("drag row missing artwork")
				}
				start("item-2")
				input(events.MouseMove, end)
				if g := grip("item-2").active; g == nil || !g.dragging || g.stage == nil {
					t.Fatal("missing threshold/preview")
				}
				animated := grip("item-2").active
				for _, id := range []string{ui.ToolbarCustomizationDragID(c.ID, "item-0"), ui.ToolbarCustomizationVisibilityID(c.ID, "item-0"), ui.ToolbarCustomizationLaterID(c.ID, "item-0")} {
					target := findRecord(root, id).widget.AsWidget()
					hover := events.NewMouse(events.LongHoverStart, events.Left, end, 0)
					target.HandleEvent(hover)
					unitSettle(t, m)
					if !hover.IsHandled() || grip("item-2").active != animated {
						t.Fatal("target help canceled held drag", id)
					}
				}
				rowHover := events.NewMouse(events.LongHoverStart, events.Left, end, 0)
				core.AsWidget(grip("item-0").button.Parent).HandleEvent(rowHover)
				unitSettle(t, m)
				if !rowHover.IsHandled() || grip("item-2").active != animated {
					t.Fatal("non-control row help escaped shared drag guard")
				}
				if animated.animation == nil {
					t.Fatal("drag lacks bounded scene animation")
				}
				if _, e := cfg.Clock.Advance(50 * time.Millisecond); e != nil {
					t.Fatal(e)
				}
				if _, e := m.Frame(testContext, "one"); e != nil {
					t.Fatal(e)
				}
				if animated.alpha <= .4 || animated.alpha >= 1 {
					t.Fatal("insertion fade did not interpolate", animated.alpha)
				}
				beforeAlpha := animated.alpha
				input(events.MouseMove, rowBefore("item-1"))
				if _, e := cfg.Clock.Advance(30 * time.Millisecond); e != nil {
					t.Fatal(e)
				}
				if _, e := m.Frame(testContext, "one"); e != nil {
					t.Fatal(e)
				}
				if animated.alpha < beforeAlpha || animated.lineY == animated.fromY || animated.lineY == animated.targetY || math.IsNaN(float64(animated.lineY)) {
					t.Fatal("insertion motion did not interpolate continuously", animated.lineY, animated.fromY, animated.targetY, animated.alpha)
				}
				input(events.MouseMove, end)
				animation, animationScene := animated.animation, animated.animationScene
				lateFrame := animation.Func
				input(events.MouseUp, end)
				if animated.animation != nil || (animation != nil && (!animation.Done || animation.Func != nil || animation.Widget != nil)) || (animationScene != nil && len(animationScene.Animations) != 0) {
					t.Fatal("release retained animation references")
				}
				lateFrame(animation)
				if animated.scroll.Property(toolbarDragActiveProperty) != nil {
					t.Fatal("release retained shared list drag capability")
				}
				if animated.hasLine || animated.stage != nil || (animationScene != nil && len(animationScene.Animations) != 0) {
					t.Fatal("late frame recreated a completed preview")
				}
				input(events.MouseUp, end)
				if changes != 1 || c.Value.Order[0] != "item-2" || !slices.Equal(c.Value.Hidden, []string{"item-2"}) || calls != 0 {
					t.Fatal("drag commit incorrect", c.Value, changes, calls)
				}
				if grip("item-2").active != nil {
					t.Fatal("release retained drag")
				}
				// Grip click and below-threshold release are no-ops.
				at := start("item-2")
				input(events.MouseMove, at.Add(image.Pt(1, 1)))
				input(events.MouseUp, at)
				if changes != 1 {
					t.Fatal("click reordered")
				}
				// Outside release is consumed, including when it lands on Done.
				start("item-2")
				done := bounds(ui.ToolbarCustomizationDoneID(c.ID))
				p := done.Min.Add(done.Size().Div(2))
				input(events.MouseMove, p)
				input(events.MouseUp, p)
				if changes != 1 || !c.Open {
					t.Fatal("outside release clicked Done")
				}
				// Escape cancels only this gesture; its pending release cannot toggle a row.
				start("item-2")
				input(events.MouseMove, rowBefore("item-1"))
				if e = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeEscape, Modifiers: 1 << key.NumLock}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
				input(events.MouseUp, rowBefore("item-0"))
				if changes != 1 || !c.Open || grip("item-2").active != nil {
					t.Fatal("Escape leaked/clicked")
				}
				// Wheel scrolling stays in the list even while the source becomes clipped.
				at = start("item-2")
				input(events.MouseMove, at.Add(image.Pt(20, 0)))
				scroll := findRecord(root, c.ID+"/items").widget.(*core.Frame)
				initial := scroll.Scrolls[math32.Y].Value
				if e = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.Scroll, Position: at, Delta: math32.Vec2(0, 1000)}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
				if scroll.Scrolls[math32.Y].Value <= initial || grip("item-2").active == nil {
					t.Fatal("captured scroll failed")
				}
				p = scroll.Geom.ContentBBox.Min.Add(image.Pt(10, scroll.Geom.ContentBBox.Dy()-5)).Add(scroll.Scene.SceneGeom.Pos)
				input(events.MouseMove, p)
				input(events.MouseUp, p)
				if changes != 2 || !slices.Contains(c.Value.Hidden, "item-2") {
					t.Fatal("scrolled drop lost hidden action", changes, c.Value)
				}
				// Application rejection leaves the accepted layout unchanged.
				accepted := ui.ToolbarLayout{Order: slices.Clone(c.Value.Order), Hidden: slices.Clone(c.Value.Hidden)}
				c.Change = func(ui.ToolbarLayout) { changes++ }
				scroll.Scrolls[math32.Y].SetValue(0)
				scroll.ScrollChanged(math32.Y, scroll.Scrolls[math32.Y])
				unitSettle(t, m)
				first := c.Value.Order[0]
				start(first)
				p = rowBefore(c.Value.Order[2])
				input(events.MouseMove, p)
				input(events.MouseUp, p)
				if changes != 3 || !slices.Equal(accepted.Order, c.Value.Order) || !slices.Equal(accepted.Hidden, c.Value.Hidden) {
					t.Fatal("rejected drop mutated accepted layout")
				}
				c.Change = func(v ui.ToolbarLayout) { changes++; c.Value = v }
				// Capture loss clears preview/references and consumes the pending release.
				at = start(first)
				input(events.MouseMove, at.Add(image.Pt(20, 0)))
				bLost := grip(first)
				lost := bLost.active
				lostStage := lost.stage
				lost.scroll.Scene.Events.CancelPointerCapture()
				if bLost.active != nil {
					t.Fatal("capture loss retained gesture")
				}
				if _, ok := lostStage.Sprites.SpriteByName(lost.sprite); ok {
					t.Fatal("capture loss retained preview")
				}
				input(events.MouseUp, rowBefore(c.Value.Order[2]))
				if changes != 3 {
					t.Fatal("capture loss committed")
				}
				// A changed order or removed source invalidates the held geometry.
				start(first)
				oldBinding := grip(first)
				c.Value.Order[0], c.Value.Order[1] = c.Value.Order[1], c.Value.Order[0]
				a.Invalidate("one")
				unitSettle(t, m)
				if oldBinding.active != nil {
					t.Fatal("order change retained capture")
				}
				input(events.MouseUp, rowBefore(c.Value.Order[2]))
				first = c.Value.Order[0]
				start(first)
				oldBinding = grip(first)
				catalogue := c.Toolbar
				c.Toolbar.Children = slices.DeleteFunc(slices.Clone(c.Toolbar.Children), func(n ui.Node) bool { return n.ID == first })
				c.Value = ui.ToolbarLayout{}
				a.Invalidate("one")
				unitSettle(t, m)
				if oldBinding.active != nil || oldBinding.scope.Load() != nil {
					t.Fatal("removed source retained capture")
				}
				input(events.MouseUp, rowBefore("item-3"))
				c.Toolbar = catalogue
				c.Value = accepted
				a.Invalidate("one")
				unitSettle(t, m)
				// Closing/reopening the panel also cancels held gestures.
				first = c.Value.Order[0]
				start(first)
				oldBinding = grip(first)
				c.Open = false
				a.Invalidate("one")
				unitSettle(t, m)
				if oldBinding.active != nil {
					t.Fatal("closed panel retained capture")
				}
				c.Open = true
				a.Invalidate("one")
				unitSettle(t, m)
				// Teardown while a gesture is active retires every captured native reference.
				scroll.Scrolls[math32.Y].SetValue(0)
				scroll.ScrollChanged(math32.Y, scroll.Scrolls[math32.Y])
				unitSettle(t, m)
				id := c.Value.Order[0]
				start(id)
				b := grip(id)
				g := b.active
				c.Disabled = true
				a.Invalidate("one")
				unitSettle(t, m)
				if b.active != nil || g.scroll.Scene.Events.PointerCaptured(g.token) {
					t.Fatal("disabled panel retained capture")
				}
				c.Disabled = false
				c.Open = true
				a.Invalidate("one")
				unitSettle(t, m)
				start(id)
				b = grip(id)
				if e = m.Close(testContext); e != nil {
					t.Fatal(e)
				}
				if b.active != nil || b.scope.Load() != nil {
					t.Fatal("window close retained drag")
				}
			})
		}
	}
}
