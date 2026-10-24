package cogent

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"html"
	"image/color"
	"reflect"
	"slices"
)

func documentRow(r *record, id string) ui.DocumentRow {
	if l := r.node.DocumentList; l != nil {
		for _, row := range l.Rows {
			if row.ID == id {
				return row
			}
		}
	}
	return ui.DocumentRow{}
}
func documentChanged(a, b *ui.DocumentListLayout) bool {
	if a == nil || b == nil {
		return a != b
	}
	return !a.SameSource(b) || !slices.Equal(a.Columns, b.Columns) || !reflect.DeepEqual(a.Rows, b.Rows)
}
func makeDocumentRow(r *record, id string, p *tree.Plan) bool {
	l := r.node.DocumentList
	if l == nil {
		return false
	}
	p.Children = nil
	for col := range l.Columns {
		index := col
		tree.AddAt(p, "document-cell-"+l.Columns[col].ID, func(cell *core.Frame) {
			cell.Styler(func(s *styles.Style) {
				if r.node.DocumentList == nil || index >= len(r.node.DocumentList.Columns) {
					return
				}
				width := r.node.DocumentList.Columns[index].Width
				s.Min.X = units.Dp(width)
				s.Max.X = units.Dp(width)
				s.Grow.Set(0, 0)
				s.Padding.Zero()
				s.Margin.Zero()
				s.Gap.Set(units.Dp(5))
				s.Border.Width.Zero()
				s.MaxBorder.Width.Zero()
				s.RenderBox = false
				s.Align.Items = styles.Center
				s.Overflow.Set(styles.OverflowHidden)
			})
			cell.Maker(func(plan *tree.Plan) {
				row := documentRow(r, id)
				if r.node.DocumentList == nil || index >= len(r.node.DocumentList.Columns) {
					return
				}
				if r.node.DocumentList.Columns[index].ID == "name" {
					if row.Image != nil {
						tree.AddAt(plan, "art", func(w *managedImage) {
							w.Styler(func(s *styles.Style) {
								s.Min.Set(units.Dp(18), units.Dp(18))
								s.Max.Set(units.Dp(18), units.Dp(18))
								s.Grow.Set(0, 0)
								s.Padding.Zero()
								s.Margin.Zero()
							})
							w.Updater(func() { w.source = documentRow(r, id).Image; w.fit = ui.ImageContain })
						})
					} else if row.Icon != "" {
						tree.AddAt(plan, "symbol", func(w *core.Icon) {
							w.Styler(func(s *styles.Style) { s.IconSize.Set(units.Dp(18)); s.RenderBox = false })
							w.Updater(func() { w.SetIcon(nativeSymbol(documentRow(r, id).Icon)) })
						})
					}
				}
				tree.AddAt(plan, "text", func(w *core.Text) {
					w.Styler(func(s *styles.Style) {
						s.RenderBox = false
						s.SetNonSelectable()
						s.SetTextWrap(false)
						s.Padding.Zero()
						s.Margin.Zero()
					})
					w.Updater(func() {
						row := documentRow(r, id)
						if index < len(row.Cells) {
							w.SetText(html.EscapeString(row.Cells[index]))
						}
					})
				})
			})
		})
	}
	return true
}
func styleDocumentRow(r *record, id string, s *styles.Style) {
	l := r.node.DocumentList
	if l == nil {
		return
	}
	s.Padding.Zero()
	s.Gap.Zero()
	s.Min.Y = units.Dp(28)
	s.Grow.Set(1, 0)
	s.Border.Width.Zero()
	s.MaxBorder.Width.Zero()
	s.Border.Radius.Zero()
	s.SetAbilities(true, abilities.DoubleClickable)
	if !ui.ListOptionSelected(r.node, id) {
		for i, row := range l.Rows {
			if row.ID == id && i%2 == 1 {
				c := color.NRGBA{R: 0, G: 0, B: 0, A: 9}
				if r.theme.Palette.Set && r.theme.Palette.Value.Window.R < 100 {
					c = color.NRGBA{R: 255, G: 255, B: 255, A: 9}
				}
				s.Background = colors.Uniform(c)
				break
			}
		}
	}
}
func queueDocumentActivation(r *record, id string, b *core.Switch) {
	guard := listBoxActionGuard(r, id, b)
	source := r.node.DocumentList
	postControl(r, func() {
		if guard() && source != nil && source.SameSource(r.node.DocumentList) && r.node.DocumentList.Activate != nil {
			r.node.DocumentList.Activate(ui.DocumentEvent{ID: r.node.ID, ItemID: id})
			requestControlRestore(r)
		}
	})
}
func bindDocumentChoice(r *record, id string, b *core.Switch) {
	b.FinalStyler(func(s *styles.Style) { styleDocumentRow(r, id, s) })
	b.OnDoubleClick(func(e events.Event) {
		if r.node.DocumentList != nil {
			e.SetHandled()
			queueDocumentActivation(r, id, b)
		}
	})
}
func bindColumnResize(r *record, b *core.Button) {
	b.FinalStyler(func(s *styles.Style) {
		if r.node.ColumnResize != nil {
			s.SetAbilities(true, abilities.Slideable)
			s.Padding.Zero()
			s.Margin.Zero()
			s.Border.Width.Zero()
			s.MaxBorder.Width.Zero()
			s.Min.X = units.Dp(8)
			s.Max.X = units.Dp(8)
			s.Grow.Set(0, 0)
		}
	})
	queue := func(h *ui.ColumnResizeHandle, value float32) {
		postControl(r, func() {
			if !controlDisabled(r) && h.SameSource(r.node.ColumnResize) && r.node.ColumnResize.Change != nil {
				r.node.ColumnResize.Change(value)
				requestControlRestore(r)
			}
		})
	}
	b.OnFirst(events.KeyChord, func(e events.Event) {
		h := r.node.ColumnResize
		if h == nil || controlDisabled(r) || e.IsHandled() || key.HasAnyModifier(e.Modifiers(), key.Meta, key.Control, key.Alt) {
			return
		}
		delta := float32(1)
		if key.HasAnyModifier(e.Modifiers(), key.Shift) {
			delta = 10
		}
		switch e.KeyCode() {
		case key.CodeLeftArrow:
			delta = -delta
		case key.CodeRightArrow:
		default:
			return
		}
		e.SetHandled()
		queue(h, h.Width+delta)
	})
	b.OnFirst(events.MouseDown, func(e events.Event) {
		h := r.node.ColumnResize
		if h == nil || controlDisabled(r) || b.Scene == nil || e.MouseButton() != events.Left {
			return
		}
		e.SetHandled()
		scale := float32(placementScale(&b.WidgetBase))
		if scale <= 0 {
			return
		}
		start := e.Pos()
		scene := b.Scene
		var token uint64
		next, err := scene.Events.CapturePointer(b, e, func(event events.Event) {
			if !h.SameSource(r.node.ColumnResize) || controlDisabled(r) || r.control.retired || r.focusVisibility.Load() == 2 {
				scene.Events.ReleasePointer(token)
				return
			}
			if event.Type() == events.MouseUp {
				scene.Events.ReleasePointer(token)
				return
			}
			if event.Type() == events.MouseDrag || event.Type() == events.MouseMove {
				queue(h, h.Width+float32(event.Pos().X-start.X)/scale)
			}
		}, func() {})
		if err != nil {
			scene.Events.ConsumePointerRelease()
			return
		}
		token = next
		b.SetFocusQuiet()
	})
}
