package cogent

import (
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/editor"
)

func requireFieldSemanticState(t *testing.T, r *record, expected core.TextFieldEditingState) {
	t.Helper()
	nodes, _ := accessibilitySnapshot(r, nil)
	if len(nodes) != 1 {
		t.Fatalf("field semantic count: %d", len(nodes))
	}
	anchor, _ := textinput.UTF16Offset(expected.Text, expected.Anchor)
	focus, _ := textinput.UTF16Offset(expected.Text, expected.Focus)
	selection := [2]int{min(anchor, focus), max(anchor, focus) - min(anchor, focus)}
	if nodes[0].Value != expected.Text || nodes[0].TextSelection == nil || *nodes[0].TextSelection != selection || nodes[0].TextFocus != focus {
		t.Fatalf("stale field semantics: value=%q selection=%v focus=%d, accepted=%+v", nodes[0].Value, nodes[0].TextSelection, nodes[0].TextFocus, expected)
	}
}

func TestFieldAccessibilityAfterCoalescedFocusHandoff(t *testing.T) {
	for _, refocus := range []bool{false, true} {
		name := "outgoing"
		if refocus {
			name = "refocused"
		}
		t.Run(name, func(t *testing.T) {
			a, w, tw, _, model := editorFixture(t, "peer")
			var changes []string
			field := focusLossField(t, a, w, tw, name, &changes)
			native := field.widget.(*core.TextField)
			tw.nativeReady.Store(true)
			tw.setActiveClient(field.control.client, true)
			if err := a.Do(func() {
				tw.apply(fixtureInput(field, 1, "mark", "A", "Aに", 2, 2, 1, 2))
				// A prior geometry/native publication captured this preview. Later
				// requests coalesce, so no publication selects this client again.
				prior := native.EditingState()
				field.control.fieldState = &prior
				requireFieldSemanticState(t, field, prior)
			}); err != nil {
				t.Fatal(err)
			}
			native.Send(events.FocusLost)
			if err := a.Do(func() {
				tw.apply(fixtureInput(field, 2, "mark", "Aに", "A日本", 3, 3, 1, 3))
				tw.apply(fixtureInput(field, 3, "unmark", "A日本", "A日本", 3, 3, -1, -1))
				if refocus {
					native.SetState(true, states.Focused)
					if !native.SetEditingSelection(3, 1) {
						t.Fatal("refocus selection refused")
					}
				}
				if !tw.finishFieldFocusLosses(func(r *record) (bool, error) {
					if r != field || tw.ack != 3 {
						t.Fatal("handoff crossed native acknowledgment")
					}
					tw.setActiveClient(0, false)
					return true, nil
				}) {
					t.Fatal("acknowledged field loss did not finish")
				}
				requireFieldSemanticState(t, field, native.EditingState())
			}); err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if len(changes) != 1 || changes[0] != "A日本" || field.control.fieldNativeActive.Load() || field.control.fieldFocusLost.Load() || model.Snapshot().History.UndoCount != 0 {
				t.Fatalf("handoff changed delivery/lifetime/peer: changes=%v", changes)
			}
		})
	}
}

func TestFieldAccessibilityAfterInactiveControlledRefresh(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "peer")
	var changes []string
	field := focusLossField(t, a, w, tw, "controlled-refresh", &changes)
	native := field.widget.(*core.TextField)
	if err := a.Do(func() {
		prior := native.EditingState()
		field.control.fieldState = &prior
		next := field.node
		next.Text = "B😀z"
		field.control.fieldSelection = &editor.Selection{Anchor: 3, Focus: 1}
		var err error
		field, err = reconcile(w.body, field, next, a, w.window.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireFieldSemanticState(t, field, native.EditingState())
		if native.EditingState().Text != next.Text {
			t.Fatal("controlled text not installed")
		}
	}); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(changes) != 0 || field.control.fieldNativeActive.Load() {
		t.Fatal("controlled refresh emitted Change or activated field")
	}
}

func TestFieldAccessibilityRetainedRefreshPreservesPreview(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "peer")
	var changes []string
	field := focusLossField(t, a, w, tw, "preview-refresh", &changes)
	native := field.widget.(*core.TextField)
	if err := a.Do(func() {
		tw.apply(fixtureInput(field, 1, "mark", "A", "Aに", 2, 2, 1, 2))
		prior := native.EditingState()
		field.control.fieldState = &prior
		var err error
		field, err = reconcile(w.body, field, field.node, a, w.window.ID)
		if err != nil {
			t.Fatal(err)
		}
		requireFieldSemanticState(t, field, prior)
		if _, _, marked := native.NativeCompositionRange(); !marked || native.EditingState() != prior {
			t.Fatal("descriptor echo erased accepted preedit")
		}
	}); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(changes) != 0 {
		t.Fatal("preview refresh emitted Change")
	}
}
