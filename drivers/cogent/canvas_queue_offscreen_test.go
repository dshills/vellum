//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"testing"
)

// Admit through the production native callback gateway, hold the owner queue,
// mutate/reconcile independently, then deliver. These are software routing tests.
func TestCanvasQueuedActionsKeepOrigin(t *testing.T) {
	for _, variant := range []string{"unchanged", "changed-object", "changed-label", "removed-object", "disabled-aba", "modal-aba", "removed-canvas", "replaced-model", "closed-window", "capture-aba"} {
		t.Run(variant, func(t *testing.T) {
			f := newPlacementFixture(t, 7, 160)
			f.app.Do(func() {
				f.metadata = []ui.CanvasObject{{ID: "existing", Name: "Existing"}}
				f.model.Selected = ""
				f.app.Invalidate("main")
			})
			manualDockSettle(t, f.manual, 7)
			native := f.manual.implementation.find("main")
			if variant == "capture-aba" {
				findRecord(native.state.root, "box").widget.AsWidget().SetFocus()
			}
			bridge := newAccessibilityBridge(f.app)
			bridge.bind("main")
			defer bridge.close()
			publish := func() {
				f.app.Do(func() { n, h := accessibilitySnapshot(native.state.root, bridge); bridge.updateOwned(n, h) })
			}
			publish()
			generation := bridge.generation.Load()
			if !accessibilityDeliver(bridge.token, generation, accessibilityManagedID(CanvasObjectID("drawing", "existing")), accessibilityAction{Kind: "press"}) {
				t.Fatal("not admitted")
			}
			if f.app.Pending().Ordinary == 0 {
				t.Fatal("not queued")
			}
			refresh := func() {
				f.app.Invalidate("main")
				if err := f.manual.implementation.refresh(native); err != nil {
					t.Fatal(err)
				}
				publish()
			}
			original := f.model
			switch variant {
			case "changed-object":
				f.app.Do(func() { f.model.Objects[0].X++ })
			case "changed-label":
				f.app.Do(func() { f.metadata[0].Name = "Changed" })
				refresh()
			case "removed-object":
				f.app.Do(func() { f.model.Objects = nil; f.metadata = nil })
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
			case "removed-canvas":
				f.hideTarget = true
				refresh()
			case "replaced-model":
				f.app.Do(func() {
					replacement := *f.model
					replacement.Objects = append([]canvas.Rect(nil), f.model.Objects...)
					f.model = &replacement
				})
				refresh()
			case "closed-window":
				if err := f.app.CloseWindow("main"); err != nil {
					t.Fatal(err)
				}
			case "capture-aba":
				var source *record
				for _, b := range placementWindows.windows[placementKey{f.app, "main"}].bindings {
					if b.source != nil {
						source = b.r
					}
				}
				if source == nil {
					t.Fatal("no palette source")
				}
				_, el := manualDockElement(t, f.manual, "main", "box")
				point := el.Bounds.Min.Add(el.Bounds.Size().Div(2))
				source.widget.AsWidget().Send(events.MouseDown, events.NewMouse(events.MouseDown, events.Left, point, 0))
				window := placementWindows.windows[placementKey{f.app, "main"}]
				if window.active == nil {
					t.Fatal("production pointer capture not acquired")
				}
				window.active.scene.Events.CancelPointerCapture()
				if window.active != nil {
					t.Fatal("capture loss not synchronous")
				}
				// Same semantic generation: only the origin action epoch rejects this ABA.
				publish()
				if bridge.generation.Load() != generation {
					t.Fatal("capture ABA changed semantics; epoch oracle weakened")
				}
			}
			if err := f.app.Drain(); err != nil && !(variant == "closed-window" && errors.Is(err, app.ErrClosed)) {
				t.Fatal(err)
			}
			if variant == "unchanged" {
				if original.Selected != "existing" {
					t.Fatal("fresh queued action failed")
				}
			} else if original.Selected != "" || f.model.Selected != "" {
				t.Fatal("stale queued action changed selection", original.Selected, f.model.Selected)
			}
			if f.calls != 0 || f.controller.State().Active {
				t.Fatal("canvas semantic action created object or retained gesture")
			}
			if variant != "closed-window" {
				f.app.Invalidate("main")
				manualDockSettle(t, f.manual, 7)
			}
		})
	}
}
