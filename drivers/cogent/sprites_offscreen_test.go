//go:build offscreen

package cogent

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system/composer"
	"cogentcore.org/core/system/driver/offscreen"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// Observe real software composition without retaining source images or changing
// copy/cache policy. Full-scene RGBA Over identifies the sprite layer in these
// fixtures: popup/dialog scenes are explicitly required to be smaller, and the
// scrim uses Uniform. These are copy requests, never measured GPU uploads.
type spriteCopyObservation struct {
	composer.Drawer
	compositions, spriteCopies, scrims, overlays int
	mainSize                                     image.Point
}

func (d *spriteCopyObservation) Start() { d.compositions++; d.Drawer.Start() }
func (d *spriteCopyObservation) Copy(dp image.Point, src image.Image, sr image.Rectangle, op draw.Op, unchanged bool) {
	if op == draw.Over {
		switch src.(type) {
		case *image.Uniform:
			d.scrims++
		case *image.RGBA:
			if src.Bounds().Size() == d.mainSize && dp == (image.Point{}) {
				d.spriteCopies++
			} else {
				d.overlays++
			}
		}
	}
	d.Drawer.Copy(dp, src, sr, op, unchanged)
}
func (d *spriteCopyObservation) GetImage() *image.RGBA {
	return d.Drawer.(interface{ GetImage() *image.RGBA }).GetImage()
}
func (d *spriteCopyObservation) CanRedraw() bool {
	if original, ok := d.Drawer.(interface{ CanRedraw() bool }); ok {
		return original.CanRedraw()
	}
	return true
}
func (d *spriteCopyObservation) reset(size image.Point) {
	d.mainSize = size
	d.compositions, d.spriteCopies, d.scrims, d.overlays = 0, 0, 0, 0
}
func watchSpriteCopies(t *testing.T, m *Manual) *spriteCopyObservation {
	t.Helper()
	w := m.implementation.find("one").state
	native := w.body.Scene.RenderWindow().SystemWindow
	if !native.Lock() {
		t.Fatal("fixture native window closed")
	}
	defer native.Unlock()
	cp, ok := native.Composer().(*composer.ComposerDrawer)
	if !ok {
		t.Fatal("fixture has no actual ComposerDrawer")
	}
	if _, ok := cp.Drawer.(interface{ GetImage() *image.RGBA }); !ok {
		t.Fatal("fixture drawer has no software image")
	}
	p := &spriteCopyObservation{Drawer: cp.Drawer}
	cp.Drawer = p
	return p
}
func spriteFrame(t *testing.T, m *Manual, p *spriteCopyObservation, label string) *image.RGBA {
	t.Helper()
	w := m.implementation.find("one").state
	p.reset(w.body.Scene.SceneGeom.Size)
	// A real scene invalidation ensures the test observes composition rather
	// than passing vacuously on an unchanged Frame that draws nothing.
	w.body.Scene.NeedsRender()
	if _, err := m.Frame(testContext, "one"); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	im, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if p.compositions == 0 {
		t.Fatal(label, "did not execute actual composition")
	}
	t.Logf("SPRITE_SOFTWARE label=%s size=%dx%d compositions=%d sprite_copy_requests=%d scrim_copies=%d overlay_copies=%d rgba_sha256=%x", label, im.Rect.Dx(), im.Rect.Dy(), p.compositions, p.spriteCopies, p.scrims, p.overlays, sha256.Sum256(im.Pix))
	return im
}
func requireNoSpriteCopies(t *testing.T, p *spriteCopyObservation, label string) {
	t.Helper()
	if p.spriteCopies != 0 {
		t.Errorf("%s emitted %d unnecessary full-scene RGBA sprite copy requests", label, p.spriteCopies)
	}
}
func requireSameSpritePixels(t *testing.T, label string, got, want *image.RGBA) {
	t.Helper()
	if got.Rect != want.Rect || got.Stride != want.Stride || !bytes.Equal(got.Pix, want.Pix) {
		t.Error(label, "failed exact detached software pixel restoration")
	}
}
func spriteImage(name string, c color.RGBA, trace *[]string) *core.Sprite {
	im := image.NewRGBA(image.Rect(0, 0, 12, 12))
	draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
	sp := core.NewImageSprite(name, image.Pt(250, 180), im)
	original := sp.Draw
	sp.Draw = func(p *paint.Painter) {
		if trace != nil {
			*trace = append(*trace, name)
		}
		original(p)
	}
	sp.Active = true
	return sp
}
func spriteClick(t *testing.T, m *Manual, id string) {
	t.Helper()
	snapshot, err := m.Snapshot(testContext)
	if err != nil || len(snapshot.Windows) != 1 {
		t.Fatal("sprite click window", err)
	}
	w := snapshot.Windows[0]
	var found *ManualElement
	for i := range w.Elements {
		if w.Elements[i].ID == id {
			found = &w.Elements[i]
			break
		}
	}
	if found == nil || !found.Enabled || !found.Visible || found.Bounds.Empty() {
		t.Fatal("sprite click target unavailable", id)
	}
	pos := found.Bounds.Min.Add(found.Bounds.Size().Div(2))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: pos}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
}

func TestManagedSpriteEmptyAndRetiredLayersUseNoCopy(t *testing.T) {
	for _, dpi := range []float32{96, 144} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, _ := simpleManual(t, config, func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Sprite composition fixture"}}}
			})
			unitSettle(t, m)
			probe := watchSpriteCopies(t, m)
			stage := m.implementation.find("one").state.body.Scene.Stage
			baseline := spriteFrame(t, m, probe, "empty")
			requireNoSpriteCopies(t, probe, "empty")

			inactive := spriteImage("fixture-inactive", color.RGBA{R: 255, A: 255}, nil)
			inactive.Active = false
			noDraw := &core.Sprite{Name: "fixture-no-draw", Active: true}
			stage.Sprites.Add(inactive)
			stage.Sprites.Add(noDraw)
			got := spriteFrame(t, m, probe, "inactive-and-no-draw")
			requireNoSpriteCopies(t, probe, "inactive and nil Draw")
			requireSameSpritePixels(t, "inactive and nil Draw", got, baseline)
			if !stage.Sprites.HasActive() {
				t.Error("nil-Draw active sprite lost its existing deadline eligibility")
			}

			var trace []string
			first := spriteImage("first", color.RGBA{R: 255, A: 255}, &trace)
			normal := spriteImage("normal", color.RGBA{G: 255, A: 255}, &trace)
			final := spriteImage("final", color.RGBA{B: 255, A: 255}, &trace)
			stage.Sprites.Lock()
			stage.Sprites.First.Set(first.Name, first)
			stage.Sprites.AddNoLock(normal)
			stage.Sprites.Final.Set(final.Name, final)
			stage.Sprites.SetModifiedNoLock()
			stage.Sprites.Unlock()
			got = spriteFrame(t, m, probe, "active-tiers")
			if probe.spriteCopies != probe.compositions || len(trace) != 3*probe.compositions {
				t.Error("active sprites did not draw once per real composition", probe.spriteCopies, probe.compositions, trace)
			}
			for i := 0; i+3 <= len(trace); i += 3 {
				if !slices.Equal(trace[i:i+3], []string{"first", "normal", "final"}) {
					t.Error("active tier callback order changed", trace)
				}
			}
			if got.RGBAAt(255, 185) != (color.RGBA{B: 255, A: 255}) {
				t.Error("final tier did not overlay earlier colors", got.RGBAAt(255, 185))
			}

			stage.Sprites.InactivateSprite("first", "normal", "final")
			got = spriteFrame(t, m, probe, "deactivated")
			requireNoSpriteCopies(t, probe, "deactivated")
			requireSameSpritePixels(t, "deactivated", got, baseline)
			for _, name := range []string{"first", "normal", "final", "fixture-inactive", "fixture-no-draw"} {
				if !stage.Sprites.Delete(name) {
					t.Error("sprite deletion missing", name)
				}
			}
			got = spriteFrame(t, m, probe, "deleted")
			requireNoSpriteCopies(t, probe, "deleted")
			requireSameSpritePixels(t, "deleted", got, baseline)

			// Resize the owned software native backend, then route its ordinary
			// WindowResize event. No window-manager or cached GPU claim is made.
			native := m.implementation.find("one").state.body.Scene.RenderWindow().SystemWindow.(*offscreen.Window)
			if !native.Lock() {
				t.Fatal("resize native window closed")
			}
			native.PixelSize, native.WnSize = image.Pt(480, 300), image.Pt(480, 300)
			native.Unlock()
			native.Events().WindowResize()
			unitSettle(t, m)
			resized := spriteFrame(t, m, probe, "resized-empty")
			if resized.Rect.Size() != image.Pt(480, 300) {
				t.Fatal("resize event did not change actual software target", resized.Rect)
			}
			requireNoSpriteCopies(t, probe, "resized empty")
			stage.Sprites.Add(normal)
			stage.Sprites.ActivateSprite("normal")
			got = spriteFrame(t, m, probe, "resized-reactivated")
			if probe.spriteCopies != probe.compositions || got.RGBAAt(255, 185) != (color.RGBA{G: 255, A: 255}) {
				t.Error("resized retained sprite layer failed reactivation", probe.spriteCopies, got.RGBAAt(255, 185))
			}
			stage.Sprites.Delete("normal")
			got = spriteFrame(t, m, probe, "resized-deleted")
			requireNoSpriteCopies(t, probe, "resized deleted")
			requireSameSpritePixels(t, "resized deleted", got, resized)
		})
	}
}

func TestManagedSpriteTextCursorExplicitBlinkAndFocusLoss(t *testing.T) {
	config := manualConfig(t, 7)
	config.LogicalDPI = 96
	m, _ := simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "away", Sequence: 1}, Children: []ui.Node{{ID: "away", Kind: ui.Button, Text: "Away"}, {ID: "field", Kind: ui.TextField, Text: "Caret 😀", Change: func(ui.ChangeEvent) {}}}}
	})
	unitSettle(t, m)
	probe := watchSpriteCopies(t, m)
	stage := m.implementation.find("one").state.body.Scene.Stage
	spriteFrame(t, m, probe, "cursor-unfocused")
	requireNoSpriteCopies(t, probe, "unfocused cursor")
	spriteClick(t, m, "field")
	on := spriteFrame(t, m, probe, "cursor-on")
	cursor, ok := stage.Sprites.SpriteByName("TextField.Cursor")
	if !ok || cursor == nil || !cursor.Active || probe.spriteCopies == 0 {
		t.Fatal("actual field focus did not activate cursor composition")
	}
	blink := core.TimingSettings.CursorBlinkTime
	if blink <= 0 {
		t.Fatal("fixture requires the existing positive blink interval")
	}
	before, err := m.Snapshot(testContext)
	if err != nil || len(before.Windows) != 1 {
		t.Fatal("cursor deadline snapshot", err)
	}
	if !before.Windows[0].FuturePaint || !before.Windows[0].NextPaint.After(config.Clock.Now()) {
		t.Fatal("active cursor lost future paint deadline")
	}
	if _, err := config.Clock.Advance(blink); err != nil {
		t.Fatal(err)
	}
	off := spriteFrame(t, m, probe, "cursor-off")
	stage.Sprites.Lock()
	isOn, _ := cursor.Properties["on"].(bool)
	active := cursor.Active
	stage.Sprites.Unlock()
	if isOn || !active || probe.spriteCopies == 0 || bytes.Equal(on.Pix, off.Pix) {
		t.Error("explicit clock did not blink actual active cursor pixels off", isOn, active, probe.spriteCopies)
	}
	if _, err := config.Clock.Advance(blink); err != nil {
		t.Fatal(err)
	}
	onAgain := spriteFrame(t, m, probe, "cursor-on-again")
	requireSameSpritePixels(t, "cursor on again", onAgain, on)
	spriteClick(t, m, "away")
	spriteFrame(t, m, probe, "cursor-focus-lost")
	requireNoSpriteCopies(t, probe, "cursor focus lost")
	stage.Sprites.Lock()
	active = cursor.Active
	stage.Sprites.Unlock()
	if active {
		t.Error("actual focus handoff retained active outgoing cursor")
	}
}

func TestManagedSpriteDialogScrimAndPopupWithoutSpriteLayer(t *testing.T) {
	config := manualConfig(t, 7)
	config.LogicalDPI = 96
	var a *app.Application
	open := false
	var popup *core.Stage
	var m *Manual
	m, a = simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
			{ID: "open", Kind: ui.Button, Text: "Open dialog", Activate: func() {
				open = true
				if err := a.Invalidate("one"); err != nil {
					t.Error(err)
				}
			}},
			{ID: "popup", Kind: ui.Button, Text: "Open popup", Activate: func() {
				ctx := findRecord(m.implementation.find("one").state.root, "popup").widget
				scene := core.NewScene("sprite-popup")
				core.NewText(scene).SetText("Contextual popup")
				popup = core.NewPopupStage(core.TooltipStage, scene, ctx).SetPos(image.Pt(180, 100)).Run()
			}},
			{ID: "dialog", Kind: ui.Dialog, Text: "Dialog", Open: open, Children: []ui.Node{{ID: "close", Kind: ui.Button, Text: "Close", Activate: func() {
				open = false
				if err := a.Invalidate("one"); err != nil {
					t.Error(err)
				}
			}}}},
		}}
	})
	unitSettle(t, m)
	probe := watchSpriteCopies(t, m)
	base := spriteFrame(t, m, probe, "dialog-baseline")
	requireNoSpriteCopies(t, probe, "dialog baseline")
	spriteClick(t, m, "open")
	dialog := findRecord(m.implementation.find("one").state.root, "dialog").control.dialog.body
	if dialog == nil || dialog.Scene.SceneGeom.Size == probe.mainSize {
		t.Fatal("fixture requires a smaller real dialog scene")
	}
	shown := spriteFrame(t, m, probe, "dialog-open")
	requireNoSpriteCopies(t, probe, "dialog open")
	if probe.scrims == 0 || probe.overlays == 0 || shown.RGBAAt(2, 2) == base.RGBAAt(2, 2) {
		t.Error("actual dialog/scrim composition missing", probe.scrims, probe.overlays)
	}
	spriteClick(t, m, "close")
	closed := spriteFrame(t, m, probe, "dialog-closed")
	requireNoSpriteCopies(t, probe, "dialog closed")
	if probe.scrims != 0 || probe.overlays != 0 || closed.RGBAAt(2, 2) != base.RGBAAt(2, 2) {
		t.Error("dialog/scrim pixels were retained after real close", probe.scrims, probe.overlays)
	}
	spriteClick(t, m, "popup")
	if popup == nil || popup.Scene.SceneGeom.Size == probe.mainSize {
		t.Fatal("fixture requires a smaller real popup scene")
	}
	spriteFrame(t, m, probe, "popup-open")
	requireNoSpriteCopies(t, probe, "popup open")
	if probe.overlays == 0 {
		t.Errorf("actual popup did not compose: scene=%v renderer=%T painter=%v main=%p originating=%p", popup.Scene.SceneGeom, popup.Scene.Renderer, popup.Scene.Painter.State != nil, popup.Main, m.implementation.find("one").state.body.Scene.Stage)
	}
	if !popup.ClosePopup() {
		t.Fatal("actual popup close refused")
	}
	spriteFrame(t, m, probe, "popup-closed")
	requireNoSpriteCopies(t, probe, "popup closed")
	if probe.overlays != 0 || probe.scrims != 0 {
		t.Error("closed popup still composed", probe.overlays, probe.scrims)
	}
}

func TestManagedSpriteActualDragMoveDropRetiresLayer(t *testing.T) {
	config := manualConfig(t, 7)
	config.LogicalDPI = 96
	m, _ := simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "source", Kind: ui.Button, Text: "Drag source"}, {ID: "target", Kind: ui.Button, Text: "Drop target"}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	source := findRecord(w.state.root, "source").widget
	target := findRecord(w.state.root, "target").widget
	starts, moves, drops := 0, 0, 0
	source.AsWidget().FinalStyler(func(s *styles.Style) { s.SetAbilities(true, abilities.Draggable) })
	target.AsWidget().FinalStyler(func(s *styles.Style) { s.SetAbilities(true, abilities.Droppable) })
	source.AsWidget().On(events.DragStart, func(e events.Event) { starts++; source.AsWidget().Scene.Events.DragStart(source, "sprite payload", e) })
	source.AsWidget().On(events.DragMove, func(events.Event) { moves++ })
	target.AsWidget().On(events.Drop, func(e events.Event) {
		de, ok := e.(*events.DragDrop)
		if !ok || de.Data != "sprite payload" || de.Source != source || de.Target != target {
			t.Error("actual drop identity/data mismatch")
			return
		}
		drops++
		de.SetHandled()
	})
	source.AsWidget().Update()
	target.AsWidget().Update()
	unitSettle(t, m)
	probe := watchSpriteCopies(t, m)
	from := source.AsWidget().Geom.TotalBBox
	to := target.AsWidget().Geom.TotalBBox
	start := from.Min.Add(from.Size().Div(2))
	end := to.Min.Add(to.Size().Div(2))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseDown, Button: events.Left, Position: start}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if _, err := config.Clock.Advance(core.TimingSettings.DragStartTime + time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	first := start.Add(image.Pt(core.TimingSettings.DragStartDistance+8, 0))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseMove, Position: first}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	stage := w.state.body.Scene.Stage
	sp, ok := stage.Sprites.SpriteByName("__DragSprite__")
	if starts != 1 || !ok || sp == nil || !sp.Active {
		t.Fatal("actual pointer drag did not create sprite", starts, ok)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseMove, Position: end}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	spriteFrame(t, m, probe, "drag-moving")
	if moves == 0 || probe.spriteCopies == 0 || sp.EventBBox.Min != end {
		t.Error("actual drag move failed sprite composition/location", moves, probe.spriteCopies, sp.EventBBox.Min, end)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseUp, Button: events.Left, Position: end}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	spriteFrame(t, m, probe, "drag-dropped")
	requireNoSpriteCopies(t, probe, "drag dropped")
	if drops != 1 {
		t.Error("actual drag/drop did not deliver once", drops)
	}
	if _, ok = stage.Sprites.SpriteByName("__DragSprite__"); ok {
		t.Error("completed actual drop retained drag sprite")
	}
	if source.AsWidget().StateIs(states.Dragging) {
		t.Error("source drag state retained after drop")
	}
}
