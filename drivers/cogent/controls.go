package cogent

import (
	"errors"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"html"
	"reflect"
	"strings"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

// SemanticsProperty exposes stable, typed control metadata on native widgets.
// It is an adapter seam, not a claim that a platform screen reader consumes it.
const SemanticsProperty = "vellum.semantics"
const disabledProperty = "vellum.disabled"

type Semantics struct {
	ID               string
	Role             ui.Kind
	Name, Help       string
	Disabled         bool
	Checked          bool
	Number           float32
	Text, SelectedID string
}
type controlState struct {
	a                 *app.Application
	window            string
	tabs              *tabState
	dialog            *dialogState
	wake              func(events.Event)
	onError           func(error)
	committed         bool
	disabled          bool // UI-owner snapshot; queued callbacks never inspect native ancestors.
	client            uint64
	textEpoch         uint64
	editor            *editorBinding
	fieldState        *core.TextFieldEditingState
	fieldSelection    *editor.Selection
	fieldNativeActive atomic.Bool           // native acceptance snapshot; focus callbacks may run outside the application owner
	fieldFocusLost    atomic.Bool           // retain a loss until native acknowledged deactivation, including refocus
	retired           bool                  // UI-owner lifetime marker; native tree destruction is window-owned.
	restore           bool                  // Restore accepted controlled value during the next native refresh.
	refused           atomic.Pointer[error] // Native-event refusal handed to the owner refresh.
}
type tabState struct {
	pages   map[string]*core.Frame
	buttons map[string]*core.Tab
	order   []string
}
type dialogState struct {
	host                       *accessibilityDialogHost
	body                       *core.Body
	active, closing, dismissed bool
}

func mountControl(parent tree.Node, r *record, a *app.Application, window string) (core.Widget, bool, error) {
	if err := validateNativeEditor(r.node); err != nil {
		return nil, true, err
	}
	if err := ui.ValidateControl(r.node); err != nil {
		return nil, true, err
	}
	c := &controlState{a: a, window: window}
	r.control = c
	var widget core.Widget
	switch r.node.Kind {
	case ui.Checkbox:
		widget = core.NewSwitch(parent).SetType(core.SwitchCheckbox)
	case ui.Selector:
		widget = core.NewChooser(parent)
	case ui.Slider:
		widget = core.NewSlider(parent)
	case ui.NumberField:
		widget = core.NewSpinner(parent)
	case ui.TextField:
		c.client = textClientID.Add(1)
		c.textEpoch = 1
		field := core.NewTextField(parent)
		field.SetNativeEditingListener(func(core.TextFieldEditingState) { requestTextPublication(window) })
		field.SetNativeFocusLostHandler(func() bool {
			// Focus can change from either the scene loop or an owner refresh.
			// Do not inspect widgets, records or owner-only native shadows here.
			t := lookupTextWindow(window)
			if t == nil || !t.nativeReady.Load() || !t.acceptsNativeInput() || (!c.fieldNativeActive.Load() && !c.fieldFocusLost.Load()) {
				return false
			}
			c.fieldFocusLost.Store(true)
			requestTextPublication(window)
			return true
		})
		widget = field
	case ui.MultilineText:
		c.client = textClientID.Add(1)
		widget = mountEditor(parent, r)
	case ui.Scroll:
		w := core.NewFrame(parent)
		w.Styler(func(s *styles.Style) { s.Direction = styles.Column; s.Overflow.Set(styles.OverflowAuto) })
		widget = w
	case ui.Tabs:
		widget = core.NewTabs(parent)
		c.tabs = &tabState{pages: map[string]*core.Frame{}, buttons: map[string]*core.Tab{}}
	case ui.Dialog:
		w := core.NewFrame(parent)
		w.Styler(func(s *styles.Style) { s.Display = styles.DisplayNone })
		widget = w
		c.dialog = &dialogState{host: newAccessibilityDialogHost(w, r)}
		c.dialog.host.Styler(func(s *styles.Style) { s.Direction = styles.Column; s.Grow.Set(1, 1) })
	default:
		r.control = nil
		return nil, false, nil
	}
	r.widget = widget
	if r.node.Kind == ui.Checkbox || r.node.Kind == ui.Selector || r.node.Kind == ui.Slider || r.node.Kind == ui.NumberField || r.node.Kind == ui.TextField {
		widget.AsWidget().OnChange(func(e events.Event) { dispatchControl(r, e) })
	}
	if err := prepareControl(r, r.node); err != nil {
		return nil, true, err
	}
	if err := syncControl(r, ui.Node{}); err != nil {
		return nil, true, err
	}
	return widget, true, nil
}

func controlDisabled(r *record) bool {
	if r.node.Disabled {
		return true
	}
	for p := r.widget.AsTree().Parent; p != nil; p = p.AsTree().Parent {
		if disabled, ok := p.AsTree().Property(disabledProperty).(bool); ok && disabled {
			return true
		}
	}
	return false
}
func postControl(r *record, fn func()) {
	if tree.IsNil(r.widget) {
		return
	}
	if err := r.control.a.PostWindow(r.control.window, func() {
		if !r.control.retired {
			fn()
		}
	}); err != nil {
		if errors.Is(err, app.ErrClosed) || errors.Is(err, app.ErrWindowClosed) {
			return
		}
		// Refusal can occur outside the owner. Coalesce the error atomically;
		// accepted values and user error callbacks are handled by native refresh.
		r.control.refused.Store(&err)
		r.control.a.Invalidate(r.control.window)
	}
}
func dispatchControl(r *record, e events.Event) {
	if r.control == nil {
		return
	}
	change := ui.ChangeEvent{ID: r.node.ID, Kind: r.node.Kind}
	switch w := r.widget.(type) {
	case *core.Switch:
		change.Checked = w.IsChecked()
	case *core.Chooser:
		id, ok := w.CurrentItem.Value.(string)
		if !ok || w.CurrentIndex < 0 {
			return
		}
		change.SelectedID = id
	case *core.Slider:
		change.Number = w.Value
	case *core.Spinner:
		change.Number = w.Value
	case *core.TextField:
		change.Text = w.EditingState().Text
	default:
		return
	}
	postControl(r, func() {
		if r.control.disabled || r.node.Change == nil {
			requestControlRestore(r)
			return
		}
		switch change.Kind {
		case ui.TextField:
			if len(change.Text) > textinput.MaxDocumentBytes || strings.ContainsRune(change.Text, 0) {
				editorError(r, fmt.Errorf("cogent: text field native document budget/encoding refused"))
				return
			}
		case ui.Selector:
			enabled := false
			for _, o := range r.node.Options {
				if o.ID == change.SelectedID && !o.Disabled {
					enabled = true
					break
				}
			}
			if !enabled {
				requestControlRestore(r)
				return
			}
		case ui.Slider, ui.NumberField:
			value, ok := ui.ClampNumeric(change.Number, r.node.Numeric)
			if !ok {
				requestControlRestore(r)
				return
			}
			change.Number = value
		}
		r.node.Change(change)
		r.control.committed = true
		r.control.a.Invalidate(r.control.window)
	})
}

// This function only updates owner fields/model invalidation. A Drain may run
// from another window's event loop, where this widget's native lock is absent.
func requestControlRestore(r *record) {
	r.control.restore = true
	r.control.a.Invalidate(r.control.window)
}

func syncSemantics(r *record) {
	name := r.node.Accessibility.Name
	if name == "" {
		name = r.node.Text
	}
	disabled := controlDisabled(r)
	sem := Semantics{ID: r.node.ID, Role: r.node.Kind, Name: name, Help: r.node.Accessibility.Help, Disabled: disabled, Checked: r.node.Checked, Number: r.node.Number, Text: r.node.Text, SelectedID: r.node.SelectedID}
	r.widget.AsTree().SetProperty(SemanticsProperty, sem)
	r.widget.AsTree().SetProperty(disabledProperty, r.node.Disabled)
	r.widget.AsWidget().Tooltip = html.EscapeString(r.node.Accessibility.Help)
	r.widget.AsWidget().SetState(disabled, states.Disabled)
	if r.control != nil && r.control.dialog != nil {
		r.control.dialog.host.SetProperty(disabledProperty, disabled)
	}
}

func syncControl(r *record, previous ui.Node) error {
	if err := ui.ValidateControl(r.node); err != nil {
		return err
	}
	if r.control == nil {
		return nil
	}
	// reconcile attaches the full topology before syncRecord visits parents then
	// children; parent semantic properties are current before this snapshot.
	r.control.disabled = controlDisabled(r)
	refused := r.control.refused.Swap(nil)
	restoring := r.control.restore || refused != nil
	committed := r.control.committed || restoring
	r.control.committed = false
	r.control.restore = false
	switch w := r.widget.(type) {
	case *core.Switch:
		w.SetText(html.EscapeString(r.node.Text)).SetChecked(r.node.Checked)
	case *core.Chooser:
		w.Items = nil
		for _, o := range r.node.Options {
			if !o.Disabled {
				w.Items = append(w.Items, core.ChooserItem{Value: o.ID, Text: html.EscapeString(o.Label)})
			}
		}
		chosen := -1
		for i, o := range w.Items {
			if o.Value == r.node.SelectedID {
				chosen = i
				break
			}
		}
		if chosen >= 0 {
			w.SetCurrentIndex(chosen)
		} else {
			label := r.node.Text
			for _, o := range r.node.Options {
				if o.ID == r.node.SelectedID {
					label = o.Label
				}
			}
			w.CurrentItem = core.ChooserItem{}
			w.SetPlaceholder(html.EscapeString(label))
		}
	case *core.Slider:
		n := r.node.Numeric
		w.Min = n.Min
		w.Max = n.Max
		w.Step = n.Step
		w.PageStep = n.Step
		w.EnforceStep = false
		w.SetValue(r.node.Number)
	case *core.Spinner:
		w.Precision = 9 // round-trip every float32 without crossing tight bounds
		n := r.node.Numeric
		w.SetMin(n.Min).SetMax(n.Max)
		w.Step = n.Step
		w.EnforceStep = false
		if previous.Number != r.node.Number || previous.Numeric != r.node.Numeric || committed {
			w.SetValue(r.node.Number)
		}
	case *textcore.Editor:
		syncEditor(r)
	case *core.TextField:
		if previous.Text != r.node.Text {
			r.control.textEpoch++
		}
		if previous.Text != r.node.Text || committed {
			w.SetText(r.node.Text)
		}
		if r.control.fieldSelection != nil {
			selection := r.control.fieldSelection
			r.control.fieldSelection = nil
			w.SetEditingSelection(selection.Anchor, selection.Focus)
		}
	case *core.Tabs:
		if err := prepareControl(r, r.node); err != nil {
			return err
		}
		for _, o := range r.node.Options {
			b := r.control.tabs.buttons[o.ID]
			b.SetText(html.EscapeString(o.Label)).SetState(controlDisabled(r) || o.Disabled, states.Disabled)
		}
		w.SelectTabByName(r.node.SelectedID)
		selectTabButtons(r, r.node.SelectedID)
	}
	syncSemantics(r)
	if r.control.dialog != nil {
		syncDialog(r)
	}
	// Value setters preserve identity; a redraw is necessary when externally
	// controlled values/option labels change, including same-length option edits.
	if restoring || previous.SoftWrap != r.node.SoftWrap || previous.Text != r.node.Text || previous.Checked != r.node.Checked || previous.Number != r.node.Number || previous.Numeric != r.node.Numeric || previous.SelectedID != r.node.SelectedID || previous.Disabled != r.node.Disabled || previous.Accessibility != r.node.Accessibility || !reflect.DeepEqual(previous.Options, r.node.Options) {
		r.widget.AsWidget().Update()
	}
	if refused != nil {
		r.widget.AsTree().SetProperty("vellum.control-error", *refused)
		if r.control.onError != nil {
			r.control.onError(*refused)
		}
	}
	return nil
}

// prepareControl allocates child hosts before the retained tree is reparented.
// Only newly needed tabs are allocated; removed pages are disposed after their
// managed descendants have moved, in finishControl.
func prepareControl(r *record, n ui.Node) error {
	if err := validateNativeEditor(n); err != nil {
		return err
	}
	if r.control == nil || r.control.tabs == nil {
		return nil
	}
	if err := ui.ValidateControl(n); err != nil {
		return err
	}
	c := r.control.tabs
	w := r.widget.(*core.Tabs)
	for _, o := range n.Options {
		if c.pages[o.ID] == nil {
			page, button := w.NewTab(o.ID)
			button.SetText(html.EscapeString(o.Label))
			c.pages[o.ID] = page
			c.buttons[o.ID] = button
			c.order = append(c.order, o.ID)
			id := o.ID
			button.OnClick(func(e events.Event) {
				postControl(r, func() {
					enabled := false
					for _, current := range r.node.Options {
						if current.ID == id {
							enabled = !r.control.disabled && !current.Disabled
							break
						}
					}
					if !enabled || r.node.Change == nil {
						requestControlRestore(r)
						return
					}
					r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.Tabs, SelectedID: id})
					r.control.a.Invalidate(r.control.window)
				})
			})
		}
	}
	return nil
}
func childHost(r *record, id string) tree.Node {
	if r.control != nil {
		if r.control.tabs != nil {
			return r.control.tabs.pages[id]
		}
		if r.control.dialog != nil {
			return r.control.dialog.host
		}
	}
	return r.widget
}
func finishControl(r *record, n ui.Node) {
	if r.control == nil || r.control.tabs == nil {
		return
	}
	c := r.control.tabs
	w := r.widget.(*core.Tabs)
	desired := map[string]bool{}
	for _, o := range n.Options {
		desired[o.ID] = true
	}
	for i := len(c.order) - 1; i >= 0; i-- {
		id := c.order[i]
		if !desired[id] {
			w.DeleteTabIndex(i)
			delete(c.pages, id)
			delete(c.buttons, id)
			c.order = append(c.order[:i], c.order[i+1:]...)
		}
	}
	pages := make([]tree.Node, len(n.Options))
	buttons := make([]tree.Node, len(n.Options))
	c.order = make([]string, len(n.Options))
	for i, o := range n.Options {
		pages[i] = c.pages[o.ID]
		buttons[i] = c.buttons[o.ID]
		c.order[i] = o.ID
	}
	// Cogent Tabs owns two internal frames; never replace the Tabs child array.
	w.Child(0).AsTree().Children = buttons
	w.Child(1).AsTree().Children = pages
	w.SelectTabByName(n.SelectedID)
	selectTabButtons(r, n.SelectedID)
}

// Native selection may return early when a reordered page retains the same
// numeric stack index. Set button states from stable model IDs independently.
func selectTabButtons(r *record, selected string) {
	for id, button := range r.control.tabs.buttons {
		button.SetSelected(id == selected)
	}
}

func syncDialog(r *record) {
	d := r.control.dialog
	if !r.node.Open || controlDisabled(r) || !accessibilityDialogVisible(r) {
		d.dismissed = false
		closeDialog(r)
		return
	}
	if d.active || d.dismissed {
		return
	}
	scene := r.widget.AsWidget().Scene
	if scene == nil || scene.RenderWindow() == nil {
		return
	}
	body := prepareDialogBody(r)
	d.active = true
	body.NewDialog(r.widget).Run()
}

// prepareDialogBody keeps one stable child host across native modal scenes.
func prepareDialogBody(r *record) *core.Body {
	d := r.control.dialog
	d.body = core.NewBody(r.node.Text)
	body := d.body
	bindAccessibilityDialogKeyboard(r, body)
	body.Scene.OnFirst(events.Custom, func(e events.Event) {
		if r.control.wake != nil {
			r.control.wake(e)
		}
	})

	tree.MoveToParent(d.host, body)
	d.host.SetScene(body.Scene)
	body.OnClose(func(e events.Event) {
		// Preserve retained children before Cogent destroys the closing Scene.
		tree.MoveToParent(d.host, r.widget)
		d.host.SetScene(r.widget.AsWidget().Scene)
		d.active = false
		d.body = nil
		if d.closing {
			return
		}
		d.dismissed = true
		postControl(r, func() {
			if r.node.Dismiss != nil {
				r.node.Dismiss()
			}
			r.control.a.Invalidate(r.control.window)
		})
	})
	body.AddBottomBar(func(bar *core.Frame) { body.AddCancel(bar).SetText("Close") })
	return body
}

func closeDialog(r *record) {
	d := r.control.dialog
	if d == nil || d.body == nil {
		return
	}
	body := d.body
	d.closing = true
	tree.MoveToParent(d.host, r.widget)
	d.host.SetScene(r.widget.AsWidget().Scene)
	body.Close()
	d.body = nil
	d.active = false
	d.closing = false
}
func flushDialogs(r *record) {
	if r == nil {
		return
	}
	if r.control != nil && r.control.dialog != nil {
		syncDialog(r)
	}
	for _, c := range r.children {
		flushDialogs(c)
	}
	for _, c := range activeWorkbenchRecords(r) {
		flushDialogs(c)
	}
}

// bindControlWake keeps queued owner work flowing while a modal owns events.
func bindControlWake(r *record, wake func(events.Event), onError ...func(error)) {
	if r == nil {
		return
	}
	if r.control != nil {
		r.control.wake = wake
		if len(onError) > 0 {
			r.control.onError = onError[0]
		}
	}
	for _, c := range r.children {
		bindControlWake(c, wake, onError...)
	}
	for _, c := range activeWorkbenchRecords(r) {
		bindControlWake(c, wake, onError...)
	}
}

func disposeControl(r *record) {
	if r == nil {
		return
	}
	if r.control != nil {
		r.control.retired = true
		r.control.fieldFocusLost.Store(false)
		r.control.fieldNativeActive.Store(false)
		if field, ok := r.widget.(*core.TextField); ok {
			field.SetNativeFocusLostHandler(nil)
			field.SetNativeEditingListener(nil)
		}
		disposeEditor(r)
	}
	if r.control != nil && r.control.dialog != nil {
		closeDialog(r)
	}
}
func disposeControlTree(r *record) {
	if r == nil {
		return
	}
	disposeControl(r)
	for _, c := range r.children {
		disposeControlTree(c)
	}
}

func applyFocus(root *record, request ui.FocusRequest, last *ui.FocusRequest) error {
	return applyAccessibleFocus(root, request, last)
}

// A Tabs descriptor focuses an enabled header button, never its nonfocusable
// native container. Prefer the selected option, then the first enabled option.
func focusWidget(r *record) (core.Widget, error) {
	if r.node.Kind != ui.Tabs {
		return r.widget, nil
	}
	var first *core.Tab
	for _, option := range r.node.Options {
		if option.Disabled {
			continue
		}
		button := r.control.tabs.buttons[option.ID]
		if tree.IsNil(button) {
			continue
		}
		if option.ID == r.node.SelectedID {
			return button, nil
		}
		if first == nil {
			first = button
		}
	}
	if first != nil {
		return first, nil
	}
	return nil, fmt.Errorf("cogent: focus target %q has no enabled tab", r.node.ID)
}
