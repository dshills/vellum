//go:build darwin && cgo

package cogent

import (
	"fmt"
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
