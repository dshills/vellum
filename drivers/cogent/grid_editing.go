package cogent

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// GridEditorID identifies the private inline field in measured/manual/AX
// observations. It is not a permitted descriptor Node.ID.
func GridEditorID(grid string) string {
	return "grid:" + base64.RawURLEncoding.EncodeToString([]byte(grid)) + ":editor"
}

type gridEditBinding struct {
	grid              *gridBinding
	record            *record
	session, revision uint64
	accepted          string
	visible           bool
	focusedOnce       bool
}

func gridRetainedEditingRecord(r *record) *record {
	if r != nil && r.control != nil && r.control.grid != nil {
		if b := r.control.grid.editing; b != nil && !b.record.control.retired {
			return b.record
		}
	}
	return nil
}
func gridEditingRecord(r *record) *record {
	if r != nil && r.control != nil && r.control.grid != nil {
		if b := r.control.grid.editing; b != nil && b.visible && !b.record.control.retired {
			return b.record
		}
	}
	return nil
}
func (b *gridEditBinding) current() (virtual.GridEditState, error) {
	if !b.grid.live() || b.record.control.retired {
		return virtual.GridEditState{}, ErrGridClosed
	}
	s, err := b.grid.model.Edit()
	if err != nil {
		return s, err
	}
	if !s.Active || s.Session != b.session || s.Revision != b.revision || s.Text != b.accepted {
		return s, ErrTextInputStale
	}
	return s, nil
}
func (b *gridEditBinding) writable() bool {
	s, err := b.current()
	return err == nil && s.Status == virtual.GridEditReady && b.visible && b.grid.eligible()
}
func validGridDraft(text string, maximum int) bool {
	return len(text) <= maximum && utf8.ValidString(text) && !strings.ContainsAny(text, "\x00\r\n")
}
func (b *gridEditBinding) maximum() int { return b.grid.projection.State.Config.MaxEditBytes }

// syncGridDraft runs only on the owner at an originating event boundary. Marked
// text remains private until an explicit native seal; it never mutates source.
func syncGridDraft(r *record) error {
	b := r.control.gridEdit
	if b == nil {
		return nil
	}
	if _, err := b.current(); err != nil {
		return err
	}
	field := r.widget.(*core.TextField)
	if _, _, marked := field.NativeCompositionRange(); marked {
		if cacheFieldEditingState(r, field.EditingState()) {
			_ = b.grid.owner.Invalidate(b.grid.window)
		}
		return nil
	}
	state := field.EditingState()
	if !validGridDraft(state.Text, b.maximum()) {
		field.SetText(b.accepted)
		cacheFieldEditingState(r, field.EditingState())
		r.control.textEpoch++
		return virtual.ErrGridCapacity
	}
	if state.Text == b.accepted {
		if cacheFieldEditingState(r, state) {
			_ = b.grid.owner.Invalidate(b.grid.window)
		}
		return nil
	}
	if !b.grid.eligible() {
		return ErrTextInputStale
	} // admitted old input may finish after view-only hiding/conflict
	if err := b.grid.model.SetEditText(state.Text); err != nil {
		field.SetText(b.accepted)
		r.control.textEpoch++
		cacheFieldEditingState(r, field.EditingState())
		return err
	}
	edit, err := b.grid.model.Edit()
	if err != nil {
		return err
	}
	b.accepted, b.revision = edit.Text, edit.Revision
	r.node.Text = edit.Text // accepted local value; retained refresh must not reset preedit/undo
	cacheFieldEditingState(r, state)
	if callback := b.grid.r.node.GridChange; callback != nil {
		s, err := b.grid.model.State()
		if err != nil {
			return err
		}
		callback(ui.GridEvent{ID: b.grid.r.node.ID, Reason: ui.GridEditing, State: s})
	}
	return b.grid.owner.Invalidate(b.grid.window)
}
func (g *gridBinding) disposeEditing() {
	if g.editing == nil {
		return
	}
	r := g.editing.record
	g.editing = nil
	disposeControl(r)
	if !tree.IsNil(r.widget) {
		r.widget.AsTree().Delete()
	}
}
func syncGridEditing(g *gridBinding) error {
	edit := g.projection.State.Edit
	if !edit.Active {
		g.disposeEditing()
		g.host.Styles.SetAbilities(true, abilities.Focusable)
		return nil
	}
	if g.editing != nil && g.editing.session != edit.Session {
		g.disposeEditing()
	}
	var cell *virtual.GridCell
	for i := range g.projection.Cells {
		c := &g.projection.Cells[i]
		if c.RowKey == edit.RowKey && c.ColumnID == edit.ColumnID && !c.Clip.Empty() {
			cell = c
			break
		}
	}
	created := g.editing == nil
	if created {
		r := &record{node: ui.Node{ID: GridEditorID(g.r.node.ID), Kind: ui.TextField, Text: edit.Text, Accessibility: ui.Accessibility{Name: "Edit " + edit.ColumnID}}, theme: g.r.theme, inheritedTheme: g.r.theme, scope: g.windowRef}
		// Change is a controlled compatibility seam. Origin boundaries/settlement
		// publish the draft first; a later focus-loss callback is therefore a no-op.
		r.node.Change = func(ui.ChangeEvent) {} // immutable compatibility notification; origin owns widget reads
		widget, _, err := mountControl(g.panes[2], r, g.owner, g.window)
		if err != nil {
			return err
		}
		r.widget = widget
		b := &gridEditBinding{grid: g, record: r, session: edit.Session, revision: edit.Revision, accepted: edit.Text}
		r.control.gridEdit = b
		g.editing = b
		bindControlWake(r, g.r.control.wake, g.r.control.onError)
		field := widget.(*core.TextField)
		field.SetName(r.node.ID)
		field.Styler(func(s *styles.Style) {
			applyStyle(s, r.node, r.theme)
			s.Margin.Zero()
			s.Padding.Set(units.Px(1))
			s.Grow.SetZero()
			s.Border.Width.Set(units.Px(1))
			s.SetAbilities(true, abilities.Focusable)
		})
		field.OnFirst(events.KeyChord, func(e events.Event) { g.editingKey(b, e) })
		field.SetText(edit.Text)
		field.SetEditingSelection(0, len([]rune(edit.Text)))
		cacheFieldEditingState(r, field.EditingState())
	}
	b := g.editing
	r := b.record
	field := r.widget.(*core.TextField)
	dirty := created || !reflect.DeepEqual(r.theme, g.r.theme)
	oldInvisible, oldDisabled, oldFocused := field.StateIs(states.Invisible), field.StateIs(states.Disabled), field.StateIs(states.Focused)
	oldParent, oldScene, oldTooltip := field.Parent, field.Scene, field.Tooltip
	oldRect, hasRect := field.Property("vellum.grid.rect").(virtual.GridRect)
	r.theme = g.r.theme
	b.visible = cell != nil
	if b.revision != edit.Revision || b.accepted != edit.Text || r.control.restore {
		dirty = true
		// An external owner draft replacement invalidates every old shadow. Never
		// import a later native preedit into that replacement.
		r.control.textEpoch++
		r.control.restore = false
		field.SetText(edit.Text)
		b.revision, b.accepted = edit.Revision, edit.Text
		r.node.Text = edit.Text
		cacheFieldEditingState(r, field.EditingState())
	}
	title := edit.ColumnID
	for _, header := range g.projection.Headers {
		if header.ColumnID == edit.ColumnID && header.Title != "" {
			title = header.Title
			break
		}
	}
	r.node.Accessibility.Name = "Edit " + title
	if cell != nil {
		r.node.Accessibility.Name += ", row " + strconv.Itoa(cell.RowIndex+1)
	}
	r.node.Disabled = g.r.control.disabled || edit.Status != virtual.GridEditReady
	r.control.disabled = g.r.control.disabled // Ready/visibility are admission rules, not a queued-input lifetime
	syncSemantics(r)

	if r.control.fieldSelection != nil {
		dirty = true
		sel := r.control.fieldSelection
		r.control.fieldSelection = nil
		field.SetEditingSelection(sel.Anchor, sel.Focus)
	}
	field.SetState(!b.visible, states.Invisible)
	field.SetState(r.node.Disabled, states.Disabled)
	if cell != nil {
		pane := g.panes[3]
		if cell.Frozen {
			pane = g.panes[2]
		}
		if field.Parent != pane {
			tree.MoveToParent(field, pane)
			field.SetScene(pane.Scene)
		}
		field.SetProperty("vellum.grid.rect", cell.Rect)
		// The private field owns the host Tab stop only while visible and writable.
		g.host.Styles.SetAbilities(!b.visible || r.node.Disabled, abilities.Focusable)
		if !r.node.Disabled && !field.StateIs(states.Focused) && !b.focusedOnce {
			b.focusedOnce = true
			if field.Scene != nil {
				field.SetFocus()
			} else {
				field.StartFocus()
			}
		}
	}
	g.host.Styles.SetAbilities(!b.visible || r.node.Disabled, abilities.Focusable)
	if (!b.visible || r.node.Disabled) && field.StateIs(states.Focused) {
		g.host.SetFocus()
	}
	dirty = dirty || oldInvisible != field.StateIs(states.Invisible) || oldDisabled != field.StateIs(states.Disabled) || oldFocused != field.StateIs(states.Focused) || oldParent != field.Parent || oldScene != field.Scene || oldTooltip != field.Tooltip
	if cell != nil && (!hasRect || oldRect != cell.Rect) {
		dirty = true
	}
	// A stable owner publication must not schedule another full frame. Native
	// text/preedit/selection APIs already request their needed rendering. Keep
	// their synchronization above; full field layout is for an actual change.
	if dirty {
		field.Update()
	}
	// A private field can repaint its preedit without a layout or ancestor
	// render. Repaint the bounded host once when completed editor data differs,
	// so its capture runs after the field and publishes the actual preview.
	// An unchanged owner publication must not request another frame.
	measured := g.measurement.Load()
	var editor *GridEditorGeometry
	if b.visible {
		state := field.EditingState()
		bounds := field.Geom.TotalBBox
		if field.Scene != nil {
			bounds = bounds.Add(field.Scene.SceneGeom.Pos)
		}
		editor = &GridEditorGeometry{ID: r.node.ID, Session: edit.Session, Text: state.Text, Focused: field.StateIs(states.Focused), Bounds: bounds}
	}
	if measured == nil || (measured.Editor == nil) != (editor == nil) || editor != nil && *measured.Editor != *editor {
		g.host.NeedsRender()
	}
	return nil
}

func (g *gridBinding) editingKey(b *gridEditBinding, e events.Event) {
	k, ok := e.(*events.Key)
	if !ok {
		return
	}
	if key.HasAnyModifier(e.Modifiers(), key.Meta, key.Control, key.Alt) {
		return
	}
	if k.Code == key.CodeReturnEnter {
		e.SetHandled()
		_ = g.owner.Do(func() {
			if g.editing != b || !b.writable() {
				return
			}
			err := SettleGridEdit(g.owner.Context(), g.owner, g.window, g.model, func(err error) {
				if err != nil {
					g.fail(err)
					return
				}
				if g.editing != b || !b.writable() {
					return
				}
				edit, err := g.model.Edit()
				if err != nil {
					g.fail(err)
					return
				}
				if callback := g.r.node.GridCommit; callback != nil {
					callback(ui.GridEditEvent{ID: g.r.node.ID, State: edit})
				}
				_ = g.owner.Invalidate(g.window)
			})
			if err != nil {
				g.fail(err)
			}
		})
	} else if k.Code == key.CodeEscape {
		e.SetHandled()
		field := b.record.widget.(*core.TextField)
		if _, _, marked := field.NativeCompositionRange(); marked {
			state := field.EditingState()
			field.ApplyNativeEdit(state.Text, state.Anchor, state.Focus, -1, -1, "cancel")
			requestTextPublicationScoped(g.windowRef)
			return
		}
		g.post(ui.GridEditing, func() error {
			if g.editing != b {
				return ErrTextInputStale
			}
			return g.model.CancelEdit()
		})
	}
}

// SettleGridEdit admits one originating-window text barrier. Call on the
// Application owner. The callback runs once while the application remains open;
// immediate refusal does not invoke it. Manual mode needs explicit Step; it
// never starts a worker, dispatches a frame, or advances virtual time.
func SettleGridEdit(ctx context.Context, a *app.Application, window string, model *virtual.Grid, callback func(error)) error {
	if ctx == nil || a == nil || model == nil || callback == nil {
		return errors.New("cogent: grid settlement requires context, owner, grid and callback")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	gridRegistry.Lock()
	g := gridRegistry.models[model]
	gridRegistry.Unlock()
	if g == nil || g.owner != a || g.window != window || !g.live() {
		return ErrGridClosed
	}
	b := g.editing
	if b == nil {
		return ErrTextInputStale
	}
	if _, err := b.current(); err != nil {
		return err
	}
	t := lookupTextWindow(window)
	if t == nil || t.owner != a || !t.acceptsNativeInput() {
		return app.ErrWindowClosed
	}
	if t.clients[b.record.control.client] != b.record {
		return ErrTextInputStale
	}
	return t.admitSettlement(ctx, &textSettlement{binding: b.record, gridEdit: b, ctx: ctx, callback: callback})
}

func (g *gridBinding) gridEditBeginKey(e events.Event) bool {
	k, ok := e.(*events.Key)
	if !ok || e.IsHandled() || !g.projection.State.Config.CellEditing {
		return false
	}
	if gridSemanticModifiers(e.Modifiers()) != 0 || k.Code != key.CodeReturnEnter && k.Code != key.CodeF2 {
		return false
	}
	e.SetHandled()
	g.post(ui.GridEditing, func() error {
		edit, err := g.model.Edit()
		if err != nil {
			return err
		}
		if edit.Active {
			return nil
		}
		return g.model.BeginActiveEdit()
	})
	return true
}

// A view-only conflict/hide disables fresh editing, but a previously admitted
// native record still owns its exact old-client FIFO position. Only that record
// temporarily bypasses the private field's status-derived disabled flag; parent
// disabled/modal/lifetime checks have already refused it on the owner.
func applyGridFieldNative(r *record, text string, anchor, focus, start, end int, phase string) bool {
	field := r.widget.(*core.TextField)
	disabled := field.StateIs(states.Disabled)
	if r.control.gridEdit != nil {
		field.SetState(false, states.Disabled)
		defer field.SetState(disabled, states.Disabled)
	}
	return field.ApplyNativeEdit(text, anchor, focus, start, end, phase)
}
