//go:build offscreen

package cogent

import (
	"strings"
	"testing"

	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestManagedTreeStableSyncQuiescesAndPreservesPendingState(t *testing.T) {
	model := nativeOutline(t, 200, true)
	disabled, help := false, "initial help"
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: model, Grow: true, Disabled: disabled, Accessibility: ui.Accessibility{Help: help}}
	})
	unitSettle(t, m)
	binding := m.implementation.find("one").state.root.control.outline
	updates := 0
	if err := a.Do(func() {
		binding.host.Updater(func() { updates++ })
		for _, row := range binding.rows {
			row.Updater(func() { updates++ })
		}
	}); err != nil {
		t.Fatal(err)
	}
	var request virtual.OutlineRequest
	initial := treeState(t, a, model)
	for _, phase := range []string{"stable", "scroll", "pending", "cancel", "disabled", "help"} {
		if err := a.Do(func() {
			switch phase {
			case "scroll":
				v := treeStateUnsafe(model)
				v.ScrollY = 300
				if err := model.SetViewport(v); err != nil {
					t.Fatal(err)
				}
			case "pending":
				var err error
				request, err = model.Prepare(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "Node"})
				if err != nil {
					t.Fatal(err)
				}
			case "cancel":
				if err := model.Cancel(request); err != nil {
					t.Fatal(err)
				}
			case "disabled":
				disabled = true
			case "help":
				help = "new help"
			}
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		if err := a.Do(func() {
			state, err := model.State()
			if err != nil {
				t.Fatal(err)
			}
			measured, err := InspectTree(a, model)
			if err != nil {
				t.Fatal(err)
			}
			if measured.Pending != state.Pending || strings.Contains(binding.status.Text, "building") != state.Pending {
				t.Errorf("%s lost same-generation pending state", phase)
			}
			if (phase == "pending" || phase == "cancel") && state.Generation != initial.Generation+1 {
				t.Errorf("%s unexpectedly changed model generation", phase)
			}
			if binding.host.StateIs(states.Disabled) != disabled || binding.host.Tooltip != help {
				t.Errorf("%s lost effective disabled/help synchronization", phase)
			}
			updates = 0
			for range 3 {
				if err := syncTree(binding.r); err != nil {
					t.Fatal(err)
				}
			}
			if updates != 0 || binding.host.Scene.GlobalLayoutPending() {
				t.Errorf("%s unchanged projection requested full update/layout: %d", phase, updates)
			}
			// Restore a noncanonical native drag even when its accepted model pixel
			// is unchanged. Repeating the exact canonical value must then quiesce.
			binding.vertical.Value = .123
			if err := syncTree(binding.r); err != nil {
				t.Fatal(err)
			}
			want := float32(float64(state.Viewport.ScrollY) / float64(treeMaxY(state)))
			if binding.vertical.Value != want {
				t.Errorf("%s did not restore canonical slider", phase)
			}
			if updates != 0 || binding.host.Scene.GlobalLayoutPending() {
				t.Errorf("%s slider restoration forced full layout", phase)
			}
		}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
}
