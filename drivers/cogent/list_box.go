package cogent

import (
	"slices"
	"strings"
	"unicode"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
)

// Row and catalog epochs are distinct from accepted selection. A removed or
// disabled row must never inherit a queued action after the same ID returns.
type listBoxRowEpoch struct{ disabled bool }
type listBoxEligibility struct {
	id       string
	disabled bool
}
type listBoxCatalogEpoch struct {
	options    []listBoxEligibility
	collection *ui.CollectionLayout
	documents  *ui.DocumentListLayout
}

func styleListBox(s *styles.Style) {
	s.Direction = styles.Column
	s.Min.X = units.Dp(160)
	s.Max.Y = units.Dp(240)
	s.Gap.Zero()
	s.Padding.Set(units.Dp(2))
	s.Overflow.Set(styles.OverflowHidden, styles.OverflowAuto)
	s.Border.Width.Set(units.Dp(1))
	s.Border.Color.Set(colors.Scheme.OutlineVariant)
	s.Border.Radius.Set(units.Dp(4))
}
func styleListBoxRow(s *styles.Style) {
	s.Grow.Set(1, 0)
	s.Justify.Content = styles.Start
	s.Padding.Set(units.Dp(6), units.Dp(8))
	s.Border.Radius.Set(units.Dp(2))
}

func setListBoxScope(r *record, visible, disabled bool, scopes []*tabCloseScope) {
	if r.control == nil || r.control.choice == nil || r.node.Kind != ui.ListBox {
		return
	}
	c := r.control.choice
	old := c.listScope.Load()
	if old != nil && !old.revoked.Load() && old.visible == visible && old.disabled == disabled && slices.Equal(old.scopes, scopes) {
		return
	}
	if old != nil {
		old.revoked.Store(true)
	}
	c.listScope.Store(&workbenchInputScope{visible: visible, disabled: disabled, scopes: slices.Clone(scopes)})
}

func syncListBoxVisibility(r *record, visible bool, scopes []*tabCloseScope) {
	if r.control == nil || r.control.choice == nil || r.node.Kind != ui.ListBox {
		return
	}
	c := r.control.choice
	c.listContext.visible = visible
	if !slices.Equal(c.listContext.scopes, scopes) {
		c.listContext.scopes = slices.Clone(scopes)
	}
	setListBoxScope(r, visible, controlDisabled(r), scopes)
}

// Returned predicates run only on the owner, including AX delivery. Merely
// changing SelectedID or a label leaves current row input admissible.
func listBoxActionGuard(r *record, id string, button *core.Switch) func() bool {
	if r.node.Kind != ui.ListBox {
		return func() bool { return true }
	}
	c := r.control.choice
	scope, row := c.listScope.Load(), c.listRows[id]
	collection := r.node.Collection
	documents := r.node.DocumentList
	return func() bool {
		return row != nil && scope.allowed() && collectionSourceLive(collection, r.node.Collection) && documents.SameSource(r.node.DocumentList) && choiceLive(r, id, button) && r.control.choice == c && c.listRows[id] == row
	}
}
func listBoxHostGuard(r *record) func() bool {
	c := r.control.choice
	scope, catalog := c.listScope.Load(), c.listCatalog
	return func() bool {
		return !r.control.retired && r.control.choice == c && scope.allowed() && catalog != nil && c.listCatalog == catalog
	}
}
func syncListBoxEpochs(r *record) bool {
	c := r.control.choice
	sourceChanged := c.listCatalog == nil || (!collectionSourceLive(c.listCatalog.collection, r.node.Collection) || !c.listCatalog.documents.SameSource(r.node.DocumentList))
	changed := sourceChanged || len(c.listCatalog.options) != len(r.node.Options)
	for i, o := range r.node.Options {
		if c.listCatalog == nil || i >= len(c.listCatalog.options) || c.listCatalog.options[i] != (listBoxEligibility{o.ID, o.Disabled}) {
			changed = true
		}
		if epoch := c.listRows[o.ID]; epoch == nil || epoch.disabled != o.Disabled || sourceChanged {
			c.listRows[o.ID] = &listBoxRowEpoch{disabled: o.Disabled}
		}
	}
	if changed {
		options := make([]listBoxEligibility, len(r.node.Options))
		for i, o := range r.node.Options {
			options[i] = listBoxEligibility{o.ID, o.Disabled}
		}
		c.listCatalog = &listBoxCatalogEpoch{options: options, collection: r.node.Collection, documents: r.node.DocumentList}
	}
	// Keep only the current detached presentation; identity-compatible image
	// updates need not retain an old thumbnail catalog.
	c.listCatalog.collection = r.node.Collection
	return changed
}

func queueListBox(r *record, id string, button *core.Switch) {
	guard := listBoxActionGuard(r, id, button)
	if !guard() || r.node.Change == nil {
		return
	}
	postControl(r, func() {
		if guard() && r.node.Change != nil {
			// Deliver absolute input proposals in FIFO order. Comparing to the
			// last rendered descriptor drops a held B→A return to accepted A.
			r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.ListBox, SelectedID: id})
			r.control.committed = true
		}
		requestControlRestore(r)
	})
}

func listBoxKey(r *record, id string, button *core.Switch, e events.Event) {
	if !listBoxActionGuard(r, id, button)() {
		return
	}
	shortcutMods := e.Modifiers() & (1<<key.Meta | 1<<key.Control | 1<<key.Alt | 1<<key.Shift)
	if e.KeyCode() == key.CodeA && shortcutMods == 1<<key.Meta {
		if r.node.Multiple && r.node.SelectionChange != nil {
			e.SetHandled()
			queueListSelectAll(r, id, button)
		}
		return
	}
	target := ""
	activate := false
	switch e.KeyCode() {
	case key.CodeUpArrow:
		target, _ = ui.MoveListBoxOption(r.node, id, -1)
	case key.CodeDownArrow:
		target, _ = ui.MoveListBoxOption(r.node, id, 1)
	case key.CodeHome:
		target, _ = ui.MoveListBoxOption(r.node, "", 1)
	case key.CodeEnd:
		target, _ = ui.MoveListBoxOption(r.node, "", -1)
	case key.CodePageUp, key.CodePageDown:
		// Full row allocation remains available even for clipped/offscreen rows.
		row := int(button.Geom.Size.Actual.Total.Y)
		page := 1
		if row > 0 {
			page = min(ui.MaxListBoxOptions, max(1, r.widget.AsWidget().Geom.ContentBBox.Dy()/row))
		}
		if e.KeyCode() == key.CodePageUp {
			page = -page
		}
		target, _ = ui.MoveListBoxOption(r.node, id, page)
	case key.CodeSpacebar, key.CodeReturnEnter:
		target = id
		activate = true
	default:
		// Application shortcuts and VoiceOver modifier chords are not typeahead.
		// Keep Shift available for ordinary uppercase letter cycling.
		if key.HasAnyModifier(e.Modifiers(), key.Meta, key.Control, key.Alt) {
			return
		}
		// Single initial-letter cycling is deterministic: no timer, hidden
		// accumulated search buffer or background state is retained.
		runeValue := unicode.ToLower(e.KeyRune())
		if !unicode.IsLetter(runeValue) && !unicode.IsDigit(runeValue) {
			return
		}
		start := slices.Index(r.control.choice.order, id)
		for offset := 1; offset <= len(r.node.Options); offset++ {
			o := r.node.Options[(start+offset)%len(r.node.Options)]
			label := []rune(strings.TrimSpace(o.Label))
			if !o.Disabled && len(label) > 0 && unicode.ToLower(label[0]) == runeValue {
				target = o.ID
				break
			}
		}
	}
	if target == "" {
		return
	}
	e.SetHandled()
	next := r.control.choice.buttons[target]
	next.SetFocus()
	next.ScrollToThis()
	if target != id || activate {
		if r.node.Multiple {
			queueListSelection(r, target, next, e.Modifiers(), id)
		} else {
			queueListBox(r, target, next)
		}
	}
}
