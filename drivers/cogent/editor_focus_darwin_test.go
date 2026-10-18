//go:build darwin && cgo

package cogent

import (
	"cogentcore.org/core/text/textcore"
	"fmt"
	"github.com/dshills/vellum/editor"
	"reflect"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
)

// This backend seam records the actual main-dispatched retirement payload. It
// creates no native view/window; Cocoa protocol acceptance is qualified separately.
type fieldRetirementBackend struct {
	sequence   uint64
	calls      int
	retired    bool
	payloadErr error
}

func (b *fieldRetirementBackend) Publish(client uint64, s core.TextFieldEditingState, start, end int, active bool, ack uint64, rect [4]float64, geometry textinput.Geometry, maximum int) (bool, error) {
	b.calls++
	if client != 0 || s.Text != "" || active || start != -1 || end != -1 || len(geometry.Cells) != 0 || maximum != textinput.MaxDocumentBytes {
		b.payloadErr = fmt.Errorf("retirement depended on refused outgoing text/geometry: client=%d textBytes=%d active=%t", client, len(s.Text), active)
		return false, b.payloadErr
	}
	if ack < b.sequence {
		return false, nil
	}
	b.retired = true
	return true, nil
}
func (*fieldRetirementBackend) Settle(uint64) bool { return false }
func (*fieldRetirementBackend) Close()             {}

func TestManagedTextFieldNativeRetirementIndependentOfRefusedLegacyValue(t *testing.T) {
	for _, text := range []string{strings.Repeat("x", textinput.MaxDocumentBytes+1), "invalid\x00legacy"} {
		t.Run(text[:1], func(t *testing.T) {
			a, w, tw, _, model := editorFixture(t, "peer")
			var changes []string
			field := focusLossField(t, a, w, tw, "retire-invalid", &changes)
			backend := &fieldRetirementBackend{sequence: 3}
			tw.native = backend
			tw.nativeReady.Store(true)
			tw.setActiveClient(field.control.client, true)
			if strings.ContainsRune(text, 0) {
				// Native edit admission already refuses NUL. A legacy setter can
				// still create it, so exercise that exact publication refusal.
				field.widget.(*core.TextField).SetText(text)
			} else if !field.widget.(*core.TextField).ApplyNativeEdit(text, 0, 0, -1, -1, "insert") {
				t.Fatal("legacy refusal fixture")
			}
			field.widget.(*core.TextField).Send(events.FocusLost)
			tw.ack = 2
			var accepted bool
			if err := a.Do(func() {
				accepted = tw.finishFieldFocusLosses(func(*record) (bool, error) { return tw.deactivateNativeClient() })
			}); err != nil {
				t.Fatal("owner retirement admission", err)
			}
			if backend.payloadErr != nil {
				t.Fatal(backend.payloadErr)
			}
			if accepted || backend.retired || !field.control.fieldFocusLost.Load() {
				t.Fatal("unacknowledged old native context was bypassed")
			}
			if err := a.Drain(); err != nil {
				t.Fatal("owner drain before quiescence", err)
			}
			if len(changes) != 0 {
				t.Fatal("partial/invalid Change before quiescence")
			}
			tw.ack = 3
			if err := a.Do(func() {
				accepted = tw.finishFieldFocusLosses(func(*record) (bool, error) { return tw.deactivateNativeClient() })
			}); err != nil {
				t.Fatal("owner retirement acknowledgment", err)
			}
			if backend.payloadErr != nil {
				t.Fatal(backend.payloadErr)
			}
			if !accepted {
				t.Fatal("valid native retirement blocked on refused Go value")
			}
			if err := a.Drain(); err != nil {
				t.Fatal("owner drain after quiescence", err)
			}
			if !backend.retired || backend.calls != 2 || tw.activeClient != 0 || field.control.fieldNativeActive.Load() || field.control.fieldFocusLost.Load() || len(changes) != 0 || field.widget.(*core.TextField).EditingState().Text != "A" || model.Snapshot().History.UndoCount != 0 {
				t.Fatal("invalid legacy refusal wedged activation or mutated control/peer", backend.calls, changes)
			}
		})
	}
}

// The fake backend exercises the real publication/owner path without creating
// an NSApplication or native window. It is not Cocoa protocol/IME evidence.
type disabledEditorPublicationBackend struct {
	client uint64
	calls  int
	active bool
	text   string
}

func (b *disabledEditorPublicationBackend) Publish(client uint64, state core.TextFieldEditingState, start, end int, active bool, ack uint64, rect [4]float64, geometry textinput.Geometry, maximum int) (bool, error) {
	if client != b.client {
		return false, fmt.Errorf("wrong outgoing editor client %d", client)
	}
	b.calls++
	b.active = active
	b.text = state.Text
	return true, nil
}
func (*disabledEditorPublicationBackend) Settle(uint64) bool { return false }
func (*disabledEditorPublicationBackend) Close()             {}

func TestDisabledEditorPublicationQuiescesAndRestoresDivergence(t *testing.T) {
	for _, mutation := range []string{"none", "text", "selection", "model-epoch", "composition-selection", "composition-unmark"} {
		t.Run(mutation, func(t *testing.T) {
			a, w, tw, r, model := editorFixture(t, "A😀B")
			if mutation == "composition-selection" || mutation == "composition-unmark" {
				if err := a.Do(func() {
					s := model.Snapshot()
					next, markedEnd := "A😀xB", 3
					if mutation == "composition-selection" {
						next, markedEnd = "A😀e\u0301B", 4
					}
					if err := model.Preview(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: editor.Range{Start: 0, End: 3}, Before: s.VisibleText, Insert: next, Selection: editor.Selection{Anchor: 3, Focus: 3}, Marked: editor.Range{Start: 2, End: markedEnd}}); err != nil {
						t.Fatal(err)
					}
					syncEditor(r)
				}); err != nil {
					t.Fatal(err)
				}
			}
			backend := &disabledEditorPublicationBackend{client: r.control.client}
			tw.native = backend
			tw.nativeReady.Store(true)
			tw.setActiveClient(r.control.client, true) // previously accepted native editor
			node := r.node
			node.Disabled = true
			reconcileDisabled := func() {
				t.Helper()
				var err error
				w.root, err = reconcile(w.body, w.root, node, a, w.window.ID)
				if err != nil {
					t.Fatal(err)
				}
				if w.root != r {
					t.Fatal("disabled editor was replaced")
				}
			}
			if err := a.Do(reconcileDisabled); err != nil {
				t.Fatal(err)
			}
			if !r.control.disabled || tw.focusedClient() != nil {
				t.Fatal("disabled outgoing-client fixture")
			}
			if err := a.Do(func() {
				switch mutation {
				case "text":
					r.control.editor.buffer.SetText([]byte("unaccepted"))
					r.widget.(*textcore.Editor).Update()
				case "selection", "composition-selection":
					if !r.widget.(*textcore.Editor).SetEditingSelection(1, 1) {
						t.Fatal("selection fixture")
					}
				case "composition-unmark":
					if err := model.CommitComposition(); err != nil {
						t.Fatal(err)
					}
				case "model-epoch":
					s := model.Snapshot()
					if err := model.Journal().Replace(true, func() error { return model.Install(s.Text, s.Selection) }); err != nil {
						t.Fatal(err)
					}
				}
				a.TakeDirty(w.window.Ref())
				r.control.restore = false
			}); err != nil {
				t.Fatal(err)
			}
			accepted := model.Snapshot()
			if err := a.Do(func() {
				tw.publish()
				dirty := a.TakeDirty(w.window.Ref())
				if dirty != (mutation != "none") || r.control.restore != (mutation != "none") {
					t.Fatal("outgoing publication quiescence/restoration", mutation, dirty, r.control.restore)
				}
				if !reflect.DeepEqual(model.Snapshot(), accepted) {
					t.Fatal("publication changed accepted model/history")
				}
				if mutation != "none" {
					reconcileDisabled()
					a.TakeDirty(w.window.Ref())
				}
				for range 8 {
					tw.publish()
					if a.TakeDirty(w.window.Ref()) || r.control.restore {
						t.Fatal("synchronized disabled publication redirties owner", mutation)
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
			if !r.widget.(*textcore.Editor).IsDisabled() || backend.calls < 8 || backend.active || backend.text != accepted.VisibleText || !reflect.DeepEqual(model.Snapshot(), accepted) {
				t.Fatal("inactive publication/model lifetime", backend.calls, backend.active, backend.text)
			}
		})
	}
}
