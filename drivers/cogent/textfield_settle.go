package cogent

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

// SettleTextField admits an asynchronous originating-window FIFO barrier for
// one ordinary managed TextField Node.ID. Call on the Application owner. It
// publishes the committed controlled Change before invoking callback once while
// the application remains open. Immediate refusal never invokes callback.
// Inline grid and multiline editors use SettleGridEdit/SettleEditor instead.
// Manual mode requires explicit Step and never starts an ambient worker/frame.
func SettleTextField(ctx context.Context, a *app.Application, window, id string, callback func(error)) error {
	if ctx == nil || a == nil || id == "" || callback == nil {
		return fmt.Errorf("cogent: text field settlement requires context, owner, ID and callback")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	t := lookupTextWindow(window)
	if t == nil || t.owner != a || !t.acceptsNativeInput() {
		return app.ErrWindowClosed
	}
	var binding *record
	for _, r := range t.clients {
		if r.node.ID == id && r.node.Kind == ui.TextField && r.control.gridEdit == nil && !r.control.retired {
			if binding != nil {
				return ErrTextInputStale
			}
			binding = r
		}
	}
	if binding == nil || binding.control.disabled || binding.node.Change == nil || binding.control.fieldChanges.Load() == nil {
		return ErrTextInputStale
	}
	return t.admitSettlement(ctx, &textSettlement{binding: binding, field: binding.control.fieldChanges.Load(), ctx: ctx, callback: callback})
}

// Called only at the originating native FIFO with its render lock and owner.
func (t *textWindow) settledField(id uint64, q *textSettlement) {
	r := q.binding
	if r.control.fieldChanges.Load() != q.field || r.control.disabled || r.node.Change == nil {
		t.complete(id, ErrTextInputStale)
		return
	}
	field := r.widget.(*core.TextField)
	if _, _, marked := field.NativeCompositionRange(); marked {
		if !t.manual {
			t.complete(id, editor.ErrComposition)
			return
		}
		state := field.EditingState()
		if !field.ApplyNativeEdit(state.Text, state.Anchor, state.Focus, -1, -1, "unmark") {
			t.complete(id, ErrTextInputStale)
			return
		}
	}
	state := field.EditingState()
	if len(state.Text) > textinput.MaxDocumentBytes || !utf8.ValidString(state.Text) || strings.ContainsRune(state.Text, 0) {
		requestControlRestore(r)
		t.complete(id, ErrTextInputStale)
		return
	}
	// Finish may enqueue an ordinary Change. Retire that snapshot along with
	// any older queued Changes, then deliver this exact final value on the owner.
	field.FinishNativeFocusLost()
	r.control.fieldFocusLost.Store(false)
	r.control.fieldChanges.Store(&fieldChangeEpoch{})
	cacheFieldEditingState(r, field.EditingState())
	if state.Text != r.control.fieldDelivered {
		r.control.fieldDelivered = state.Text
		r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.TextField, Text: state.Text})
		// Change may close the window or dispose its binding. Do not access the
		// native widget after invoking application code.
		if !t.acceptsNativeInput() {
			t.complete(id, app.ErrWindowClosed)
			return
		}
		if r.control.retired || t.clients[r.control.client] != r {
			t.complete(id, ErrTextInputStale)
			return
		}
		r.control.committed = true
	}
	_ = t.owner.Invalidate(r.control.window)
	t.complete(id, nil)
}
