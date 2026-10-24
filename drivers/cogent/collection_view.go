package cogent

import (
	"html"
	"slices"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/text"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

// Presentation reads retained, detached Node data only. Model callbacks remain
// on the existing bounded originating-window owner queue and ListBox leases.
func styleCollectionHost(r *record, s *styles.Style) {
	if r.node.Collection == nil {
		return
	}
	s.Direction = styles.Row
	s.Wrap = true
	s.Grow.Set(1, 0)
	s.Min.X = units.Dp(80)
	s.Min.Y = units.Dp(160)
	s.Max.Y = units.Dp(360)
	s.Gap.Set(units.Dp(12))
	s.Padding.Set(units.Dp(12))
	s.Overflow.Set(styles.OverflowHidden, styles.OverflowAuto)
}
func collectionItem(r *record, id string) ui.CollectionItem {
	if c := r.node.Collection; c != nil {
		for _, item := range c.Items {
			if item.ID == id {
				return item
			}
		}
	}
	return ui.CollectionItem{}
}
func collectionWidth(r *record) float32 {
	if c := r.node.Collection; c != nil && c.ItemWidth != 0 {
		return c.ItemWidth
	}
	return 112
}
func styleCollectionTile(r *record, id string, s *styles.Style) {
	if r.node.Collection == nil {
		return
	}
	s.Direction = styles.Column
	s.Grow.Set(0, 0)
	s.Min.Set(units.Dp(collectionWidth(r)-16), units.Dp(96))
	s.Max.X = units.Dp(collectionWidth(r) - 16)
	s.Padding.Set(units.Dp(8))
	s.Gap.Set(units.Dp(6))
	s.Border.Radius.Set(units.Dp(7))
	s.IconSize.Zero()
	s.Justify.Content = styles.Start
	s.SetAbilities(true, abilities.DoubleClickable)
}

func bindCollectionChoice(r *record, id string, b *core.Switch) {
	// FinalMaker replaces Switch's private checkbox stack for collections only.
	// Existing ListBox semantic options, eligibility and membership remain intact.
	b.FinalMaker(func(p *tree.Plan) {
		if makeDocumentRow(r, id, p) {
			return
		}
		if r.node.Collection == nil {
			option := choiceOption(r, id)
			if r.node.Kind != ui.ListBox || !r.node.Multiple || option.IconImage == nil {
				return
			}
			p.Children = nil
			tree.AddAt(p, "list-thumbnail", func(w *managedImage) {
				w.Styler(func(s *styles.Style) {
					s.Min.Set(units.Dp(18), units.Dp(18))
					s.Max.Set(units.Dp(18), units.Dp(18))
					s.Grow.Set(0, 0)
					s.Padding.Zero()
					s.Margin.Zero()
				})
				w.Updater(func() { w.source = choiceOption(r, id).IconImage; w.fit = ui.ImageContain })
			})
			tree.AddAt(p, "list-label", func(w *core.Text) {
				w.Styler(func(s *styles.Style) { s.RenderBox = false; s.SetNonSelectable() })
				w.Updater(func() { w.SetText(html.EscapeString(choiceOption(r, id).Label)) })
			})
			if option.Badge != "" {
				tree.AddAt(p, "list-badge", func(w *core.Text) { w.Updater(func() { w.SetText(html.EscapeString(choiceOption(r, id).Badge)) }) })
			}
			return
		}
		p.Children = nil
		item := collectionItem(r, id)
		if item.Thumbnail != nil {
			tree.AddAt(p, "collection-thumbnail", func(w *managedImage) {
				w.Styler(func(s *styles.Style) {
					s.Min.Set(units.Dp(48), units.Dp(48))
					s.Max.Set(units.Dp(48), units.Dp(48))
					s.Grow.Set(0, 0)
				})
				w.Updater(func() { w.source = collectionItem(r, id).Thumbnail; w.fit = ui.ImageContain })
			})
		} else {
			tree.AddAt(p, "collection-icon", func(w *core.Icon) {
				w.Styler(func(s *styles.Style) { s.IconSize.Set(units.Dp(48)); s.RenderBox = false })
				w.Updater(func() {
					icon := collectionItem(r, id).Icon
					if icon == "" {
						icon = ui.SymbolDocument
					}
					w.SetIcon(nativeSymbol(icon))
				})
			})
		}
		tree.AddAt(p, "collection-label", func(w *core.Text) {
			// Keep the intrinsic paragraph centered with the fixed artwork. The
			// native allocation policy also centers shorter wrapped lines.
			w.AllocationLayout = true
			w.Styler(func(s *styles.Style) {
				s.RenderBox = false
				s.SetNonSelectable()
				s.SetTextWrap(true)
				s.Text.Align = text.Center
				s.Max.X = units.Dp(collectionWidth(r) - 16)
				s.Grow.X = 0
				s.Font.Size.Dp(12)
			})
			w.Updater(func() {
				label := collectionItem(r, id).Label
				shown := []rune(label)
				if len(shown) > 28 {
					label = string(shown[:25]) + "…"
				}
				w.SetText(html.EscapeString(label))
			})
		})
	})
	b.FinalStyler(func(s *styles.Style) { styleCollectionTile(r, id, s) })
	b.OnFirst(events.DoubleClick, func(e events.Event) {
		if r.node.Collection == nil || e.Modifiers() != 0 {
			return
		}
		e.SetHandled()
		queueCollectionActivation(r, id, b)
	})
}
func queueCollectionActivation(r *record, id string, b *core.Switch) {
	guard := listBoxActionGuard(r, id, b)
	catalog := r.control.choice.listCatalog
	if r.node.Collection == nil || r.node.Collection.Activate == nil || !guard() {
		return
	}
	postControl(r, func() {
		if !guard() || r.control.choice.listCatalog != catalog || r.node.Collection == nil || r.node.Collection.Activate == nil {
			return
		}
		r.node.Collection.Activate(ui.CollectionEvent{ID: r.node.ID, ItemID: id})
		requestControlRestore(r)
	})
}

// Wrapped layout is measured in scene dots. Counting equal row origins avoids
// guessing columns from DPI, viewport width, padding or scrollbar allocations.
func collectionColumns(r *record) int {
	c := r.control.choice
	if len(c.order) == 0 {
		return 1
	}
	first := c.buttons[c.order[0]]
	if first == nil {
		return 1
	}
	y := first.Geom.Pos.Total.Y
	columns := 0
	for _, id := range c.order {
		b := c.buttons[id]
		if b == nil || b.Geom.Pos.Total.Y != y {
			break
		}
		columns++
	}
	return max(1, columns)
}
func collectionKey(r *record, id string, b *core.Switch, e events.Event) {
	if e.IsHandled() || r.node.Collection == nil || !listBoxActionGuard(r, id, b)() {
		return
	}
	mods := e.Modifiers() & ^((1 << key.CapsLock) | (1 << key.NumLock))
	if mods & ^((1<<key.Shift)|(1<<key.Meta)) != 0 {
		return
	}
	if e.KeyCode() == key.CodeA && mods == 1<<key.Meta {
		e.SetHandled()
		queueListSelectAll(r, id, b)
		return
	}
	if e.KeyCode() == key.CodeReturnEnter && mods == 0 {
		e.SetHandled()
		queueCollectionActivation(r, id, b)
		return
	}
	dx, dy := 0, 0
	switch e.KeyCode() {
	case key.CodeLeftArrow:
		dx = -1
	case key.CodeRightArrow:
		dx = 1
	case key.CodeUpArrow:
		dy = -1
	case key.CodeDownArrow:
		dy = 1
	case key.CodeSpacebar:
		e.SetHandled()
		queueListSelection(r, id, b, mods, id)
		return
	default:
		return
	}
	target, ok := ui.MoveCollectionItem(r.node.Collection.Items, id, collectionColumns(r), dx, dy)
	if !ok {
		return
	}
	e.SetHandled()
	next := r.control.choice.buttons[target]
	if next == nil {
		return
	}
	next.SetFocus()
	next.ScrollToThis()
	// Command+arrow moves focus without destroying accepted membership.
	if !key.HasAnyModifier(e.Modifiers(), key.Meta) || key.HasAnyModifier(e.Modifiers(), key.Shift) {
		queueListSelection(r, target, next, mods, id)
	}
}

// Called before native choices update so a thumbnail/icon replacement refreshes
// the private plan even when labels and eligibility remain unchanged.
func collectionPresentationChanged(a, b *ui.CollectionLayout) bool {
	if a == nil || b == nil {
		return a != b
	}
	return a.ItemWidth != b.ItemWidth || !slices.Equal(a.Items, b.Items)
}

// Snapshot source comparison is identity-only; no mutable owner state is read.
func collectionSourceLive(previous, current *ui.CollectionLayout) bool {
	return previous.SameSource(current)
}
