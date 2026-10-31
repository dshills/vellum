//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

// This separate module exercises the actual managed widgets through public APIs.
// It creates software windows only and is not physical pointer certification.
// The qualified descriptor runner excludes the app-local native drag extension;
// drag math/history are tested separately and the hook is checked in the Mac app.
func TestPublicPaletteCanvasConsumer(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			ctx := context.Background()
			clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
			if err != nil {
				t.Fatal(err)
			}
			defer clock.Close()
			executor, err := execution.NewManual(8)
			if err != nil {
				t.Fatal(err)
			}
			defer executor.Close()
			m, err := newModel()
			if err != nil {
				t.Fatal(err)
			}
			defer m.palette.Close()
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			var windowRef app.WindowRef
			if windowRef, err = a.OpenWindow(app.Window{ID: "placement", Title: "Placement", Width: 1100, Height: 760, View: func() ui.Node { return managedTestView(view(a, windowRef, m)) }}); err != nil {
				t.Fatal(err)
			}
			config := cogent.ManualConfig{Clock: clock, Executor: executor, Commands: m.commands, LogicalDPI: 160, OwnerBatch: batch, MaxWindows: 2, EventCapacity: 128, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
			if err := fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || (!strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf")) {
					return nil
				}
				data, err := fs.ReadFile(fonts.Default, path)
				if err != nil {
					return err
				}
				digest := sha256.Sum256(data)
				config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(digest[:])})
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			driver, err := cogent.NewManual(ctx, a, config)
			if err != nil {
				t.Fatal(err)
			}
			defer driver.Close(ctx)
			settle := func() {
				for range 4096 {
					s, err := driver.Snapshot(ctx)
					if err != nil {
						t.Fatal(err)
					}
					ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
					for _, w := range s.Windows {
						ready = ready || w.Events > 0 || w.ReadyPaint
						if !ready && w.LayoutPending {
							if err := driver.RequestFrame(ctx, w.ID); err != nil {
								t.Fatal(err)
							}
							ready = true
						}
					}
					if !ready {
						return
					}
					if _, err := driver.Step(ctx, batch); err != nil {
						t.Fatal(err)
					}
				}
				t.Fatal("dispatch bound exhausted")
			}
			element := func(id string) (cogent.ManualWindowSnapshot, cogent.ManualElement) {
				s, err := driver.Snapshot(ctx)
				if err != nil {
					t.Fatal(err)
				}
				for _, w := range s.Windows {
					for _, e := range w.Elements {
						if e.ID == id {
							return w, e
						}
					}
				}
				t.Fatal("missing control", id)
				return cogent.ManualWindowSnapshot{}, cogent.ManualElement{}
			}
			input := func(i cogent.ManualInput) {
				w, _ := element("root")
				i.Window = w.ID
				i.WindowIncarnation = w.Incarnation
				if err := driver.Input(ctx, i); err != nil {
					t.Fatalf("input %+v: %v", i, err)
				}
				settle()
			}
			click := func(id string) {
				_, e := element(id)
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					input(cogent.ManualInput{Type: typ, Button: events.Left, Position: e.Bounds.Min.Add(e.Bounds.Size().Div(2))})
				}
			}
			chooseMenu := func(menu, action string) {
				click(menu)
				click("menu:" + base64.RawURLEncoding.EncodeToString([]byte(menu)) + ":" + base64.RawURLEncoding.EncodeToString([]byte(action)))
			}
			settle()
			bounds := func() map[string]image.Rectangle {
				snapshot, err := driver.Snapshot(ctx)
				if err != nil {
					t.Fatal(err)
				}
				result := map[string]image.Rectangle{}
				for _, e := range snapshot.Windows[0].Elements {
					result[e.ID] = e.Bounds
				}
				return result
			}
			unchanged := func(before map[string]image.Rectangle) {
				t.Helper()
				after := bounds()
				for id, old := range before {
					if got, ok := after[id]; !ok || got != old {
						t.Fatalf("mouse-down resized or removed %s: before=%v after=%v present=%t", id, old, got, ok)
					}
				}
			}
			for _, id := range []string{"fit", "reset"} {
				before := bounds()
				_, control := element(id)
				at := control.Bounds.Min.Add(control.Bounds.Size().Div(2))
				input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: at})
				unchanged(before)
				input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: at})
			}
			// A held palette press preserves every measured control while live
			// command admission stays blocked. A short release restores admission
			// without placing an object or changing the layout.
			idle := bounds()
			_, first := element("tool-block")
			at := first.Bounds.Min.Add(first.Bounds.Size().Div(2))
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: at})
			if !m.palette.State().Active {
				t.Fatal("palette did not capture press")
			}
			unchanged(idle)
			if m.commands.Invoke(ctx, "fit") == nil {
				t.Fatal("frozen appearance admitted command during capture")
			}
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: at})
			unchanged(idle)
			if m.palette.State().Active || m.commands.Enabled("fit") != nil || len(m.scene.Objects) != 0 {
				t.Fatal("short release did not restore idle admission without placement")
			}
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: at})
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeEscape})
			unchanged(idle)
			if m.palette.State().Active || m.commands.Enabled("fit") != nil {
				t.Fatal("Escape retained capture presentation or blocked admission")
			}
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: at})
			unchanged(idle)
			for _, id := range []string{"block", "connector", "note"} {
				beforePress := bounds()
				_, source := element("tool-" + id)
				_, target := element("drawing")
				input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: source.Bounds.Min.Add(source.Bounds.Size().Div(2))})
				unchanged(beforePress)
				pos := target.Bounds.Min.Add(target.Bounds.Size().Div(2)).Add(image.Pt(37, -19))
				input(cogent.ManualInput{Type: events.MouseMove, Position: pos})
				if !m.palette.State().Accepted || len(m.scene.Objects) != m.journal.Snapshot().UndoCount {
					t.Fatal("preview changed history or rejected valid drop")
				}
				input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: pos})
			}
			if len(m.scene.Objects) != 3 || m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("routed placements did not commit once")
			}
			original := slices.Clone(m.scene.Objects)
			click("undo")
			if len(m.scene.Objects) != 2 {
				t.Fatal("routed Undo failed")
			}
			click("redo")
			if !slices.Equal(original, m.scene.Objects) {
				t.Fatal("routed Redo changed identity")
			}
			// Source pointer clicks never place. They focus the native button; Return
			// follows the same validation and command path at the current world center.
			click("tool-block")
			if len(m.scene.Objects) != 3 {
				t.Fatal("palette click placed")
			}
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter, Rune: '\n'})
			if len(m.scene.Objects) != 4 || m.journal.Snapshot().UndoCount != 4 {
				t.Fatal("focused keyboard placement failed")
			}
			fourth := m.scene.Objects[3]
			input(cogent.ManualInput{Type: events.KeyChord, Rune: 'z', Code: key.CodeZ, Modifiers: key.Modifiers(1 << key.Meta)})
			if len(m.scene.Objects) != 3 {
				t.Fatal("Cmd+Z did not undo one placement")
			}
			input(cogent.ManualInput{Type: events.KeyChord, Rune: 'Z', Code: key.CodeZ, Modifiers: key.Modifiers(1<<key.Meta | 1<<key.Shift)})
			if len(m.scene.Objects) != 4 || m.scene.Objects[3] != fourth {
				t.Fatal("Cmd+Shift+Z changed placement identity")
			}
			click("reject-notes")
			click("tool-note")
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter, Rune: '\n'})
			if len(m.scene.Objects) != 4 {
				t.Fatal("keyboard bypassed validation")
			}
			_, target := element("drawing")
			center := target.Bounds.Min.Add(target.Bounds.Size().Div(2)).Add(image.Pt(20, 20))
			// Fresh canvas input still selects, pans and zooms after captured releases.
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: center})
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: center})
			if m.scene.Selected == "" {
				t.Fatal("canvas selection failed")
			}
			inspection, err := driver.Snapshot(ctx)
			if err != nil || len(inspection.Windows[0].Canvases) != 1 || len(inspection.Windows[0].Canvases[0].Objects) == 0 {
				t.Fatal("public canvas semantics missing", err)
			}
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeHome, Modifiers: key.Modifiers(1 << key.Alt)})
			if m.scene.Selected != m.scene.Objects[0].ID {
				t.Fatal("public keyboard object selection failed")
			}

			// The same managed callback must collapse even when Home reselects the primary.
			if err := a.Do(func() {
				m.setSelection([]string{m.scene.Objects[0].ID, m.scene.Objects[1].ID}, m.scene.Objects[0].ID)
				_ = a.Invalidate("placement")
			}); err != nil {
				t.Fatal(err)
			}
			settle()
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeHome, Modifiers: key.Modifiers(1 << key.Alt)})
			if len(m.selectedIDs()) != 1 {
				t.Fatal("same-primary OptionHome did not collapse group")
			}
			before := m.scene.View
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeRightArrow})
			if m.scene.View.Center == before.Center {
				t.Fatal("canvas keyboard pan failed")
			}
			before = m.scene.View
			input(cogent.ManualInput{Type: events.Scroll, Position: center, Delta: math32.Vec2(0, -40)})
			if m.scene.View.Zoom <= before.Zoom {
				t.Fatal("canvas wheel zoom failed")
			}
			moved := slices.Clone(m.scene.Objects)
			dragView := m.scene.View
			index := m.selectedIndex()
			x := m.scene.Objects[index].X
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeRightArrow, Modifiers: key.Modifiers(1<<key.Meta | 1<<key.Alt)})
			if m.scene.Objects[index].X != x+20 || m.scene.View != dragView {
				t.Fatal("move shortcut was intercepted by canvas navigation")
			}
			click("undo")
			chooseMenu("edit-menu", "delete")
			if len(m.scene.Objects) != 3 {
				t.Fatal("routed delete")
			}
			click("undo")
			if !slices.Equal(m.scene.Objects, moved) {
				t.Fatal("delete undo order/identity")
			}
			input(cogent.ManualInput{Type: events.KeyChord, Code: key.CodeBackspace})
			if len(m.scene.Objects) != 3 {
				t.Fatal("Backspace command did not delete")
			}
			click("undo")

			index = m.selectedIndex()
			originalSize := m.scene.Objects[index]
			chooseMenu("resize-menu", "width-more")
			if m.scene.Objects[index].Width != originalSize.Width+20 || m.scene.Objects[index].Height != originalSize.Height {
				t.Fatal("routed resize width")
			}
			click("undo")
			if m.scene.Objects[index] != originalSize {
				t.Fatal("routed resize undo")
			}
			click("redo")
			if m.scene.Objects[index].Width != originalSize.Width+20 {
				t.Fatal("routed resize redo")
			}
			click("undo")
			before = m.scene.View
			center = target.Bounds.Min.Add(image.Pt(20, 20)) // Empty space pans; dragging an object now moves it.
			input(cogent.ManualInput{Type: events.MouseDown, Button: events.Left, Position: center})
			if _, err := clock.Advance(core.TimingSettings.SlideStartTime + time.Millisecond); err != nil {
				t.Fatal(err)
			}
			input(cogent.ManualInput{Type: events.MouseMove, Button: events.Left, Position: center.Add(image.Pt(20, 10))})
			input(cogent.ManualInput{Type: events.MouseMove, Button: events.Left, Position: center.Add(image.Pt(40, 20))})
			input(cogent.ManualInput{Type: events.MouseUp, Button: events.Left, Position: center.Add(image.Pt(40, 20))})
			if m.scene.View.Center == before.Center {
				t.Fatal("canvas pointer pan failed")
			}
			if _, err := driver.Capture(ctx, "placement"); err != nil {
				t.Fatal(err)
			}
			if m.palette.State().Active || a.Pending().Reserved != 0 {
				t.Fatal("consumer retained drag state", a.Pending())
			}
		})
	}
}

// The descriptor manual runner qualifies managed nodes, not arbitrary app code.
func managedTestView(n ui.Node) ui.Node {
	children := n.Children[:0]
	for _, child := range n.Children {
		if child.ID != "canvas-edit-events" && child.ID != "layers-drag-events" && child.ID != "command-palette-keys" {
			children = append(children, managedTestView(child))
		}
	}
	n.Children = children
	return n
}
