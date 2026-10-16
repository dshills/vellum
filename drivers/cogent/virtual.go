package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/virtual"
	"html"
	"strings"
)

type VirtualTable struct {
	input   *virtualMountScope
	OnError func(error)
	wheel   float32
	Owner   *app.Application
	Window  string
	Model   *virtual.Table
	Columns []string
	Select  func(virtual.Row)
	host    *core.Frame
	rows    []*core.Button
	status  *core.Text
}

func (v *VirtualTable) Mount(parent tree.Node) core.Widget {
	if v == nil || v.Owner == nil || v.Model == nil || v.Model.Source() == nil || v.Model.VisibleCount() < 1 || v.Model.VisibleCount() > virtual.MaxVisible {
		return nil
	}
	ref, err := captureMountScope(parent, v.Owner, v.Window)
	if err != nil {
		if v.OnError != nil {
			v.OnError(err)
		}
		return nil
	}
	lease := &virtualMountScope{owner: v.Owner, ref: ref, window: v.Window}
	v.input = lease
	v.rows = nil
	v.wheel = 0
	v.host = core.NewFrame(parent)
	v.host.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
		s.Direction = styles.Column
		s.SetAbilities(true, abilities.Focusable)
	})
	core.NewText(v.host).SetText(html.EscapeString(strings.Join(v.Columns, "      ")))
	bar := core.NewFrame(v.host)
	bar.Styler(func(s *styles.Style) { s.Direction = styles.Row })
	for _, delta := range []int{-v.Model.VisibleCount(), v.Model.VisibleCount()} {
		d := delta
		text := "Next page"
		if d < 0 {
			text = "Previous page"
		}
		core.NewButton(bar).SetText(text).OnClick(func(events.Event) { v.scrollFrom(lease, d) })
	}
	v.status = core.NewText(bar)
	for i := 0; i < v.Model.VisibleCount(); i++ {
		slot := i
		b := core.NewButton(v.host)
		b.SetName(fmt.Sprintf("table-slot-%d", i))
		b.Styler(func(s *styles.Style) {
			s.Grow.X = 1
			s.Min.Y = units.Dp(20)
			s.Padding.Set(units.Dp(1))
			s.Font.Size = units.Dp(12)
		})
		b.OnClick(func(events.Event) {
			v.ownerEvent(lease, func() {
				r, err := v.Model.SelectVisible(slot)
				if err != nil {
					if v.OnError != nil {
						v.OnError(err)
					}
				} else if v.Select != nil {
					v.Select(r)
				}
				v.Owner.Invalidate(v.Window)
			})
		})
		v.rows = append(v.rows, b)
	}
	v.host.On(events.Scroll, func(e events.Event) {
		if s, ok := e.(*events.MouseScroll); ok {
			v.ownerEvent(lease, func() { d := wheelRows(&v.wheel, s.Delta.Y); v.scrollOwned(d) })
			e.SetHandled()
		}
	})
	return v.host
}
func (v *VirtualTable) scroll(delta int) { v.scrollFrom(v.input, delta) }
func (v *VirtualTable) scrollFrom(lease *virtualMountScope, delta int) {
	v.ownerEvent(lease, func() { v.scrollOwned(delta) })
}
func (v *VirtualTable) scrollOwned(delta int) {
	if delta == 0 {
		return
	}
	old := v.Model.StartRow()
	start := v.Model.StartRow() + delta
	if start < 0 {
		start = 0
	}
	maximum := v.Model.Len() - v.Model.VisibleCount()
	if maximum < 0 {
		maximum = 0
	}
	if start > maximum {
		start = maximum
	}
	if err := v.Model.SetViewport(start, v.Model.VisibleCount()); err != nil {
		if v.OnError != nil {
			v.OnError(err)
		}
		return
	}
	if old != v.Model.StartRow() {
		v.Owner.Invalidate(v.Window)
	}
}
func (v *VirtualTable) Sync(core.Widget) error {
	if v == nil || v.Model == nil || v.Model.Source() == nil || v.host == nil || v.Model.VisibleCount() != len(v.rows) {
		return fmt.Errorf("virtual: mounted table and fixed row-pool size required")
	}
	rows := v.Model.View()
	for i, b := range v.rows {
		text := ""
		invisible := i >= len(rows)
		selected := false
		if !invisible {
			text = html.EscapeString(strings.Join(rows[i].Cells, "      "))
			selected = rows[i].ID == v.Model.Selected
		}
		if b.Text != text || b.StateIs(states.Invisible) != invisible || b.StateIs(states.Selected) != selected {
			b.SetText(text)
			b.SetState(invisible, states.Invisible)
			b.SetState(selected, states.Selected)
			b.Update()
		}
	}
	first := v.Model.StartRow() + 1
	if len(rows) == 0 {
		first = 0
	}
	text := fmt.Sprintf("%d–%d of %d · %d pooled rows", first, v.Model.StartRow()+len(rows), v.Model.Len(), len(v.rows))
	if v.status.Text != text {
		v.status.SetText(text).Update()
	}
	return nil
}

type VirtualTree struct {
	input        *virtualMountScope
	OnError      func(error)
	wheel        float32
	Owner        *app.Application
	Window       string
	Model        *virtual.Tree
	Start, Count int
	Select       func(string)
	host         *core.Frame
	rows         []*core.Button
	visible      []virtual.TreeRow
}

func (v *VirtualTree) Mount(parent tree.Node) core.Widget {
	if v == nil || v.Owner == nil || v.Model == nil || v.Count < 1 || v.Count > virtual.MaxVisible {
		return nil
	}
	ref, err := captureMountScope(parent, v.Owner, v.Window)
	if err != nil {
		if v.OnError != nil {
			v.OnError(err)
		}
		return nil
	}
	lease := &virtualMountScope{owner: v.Owner, ref: ref, window: v.Window}
	v.input = lease
	v.rows = nil
	v.wheel = 0
	v.host = core.NewFrame(parent)
	v.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Direction = styles.Column })
	mountAccessibilityTreePager(v)
	for i := 0; i < v.Count; i++ {
		slot := i
		b := core.NewButton(v.host)
		b.Styler(func(s *styles.Style) {
			s.Grow.X = 1
			s.Min.Y = units.Dp(20)
			s.Padding.Set(units.Dp(1))
			s.Font.Size = units.Dp(12)
		})
		b.OnClick(func(events.Event) {
			v.ownerEvent(lease, func() {
				if slot >= len(v.visible) {
					return
				}
				r := v.visible[slot]
				v.Model.Select(r.ID)
				if r.HasChildren {
					v.Model.Toggle(r.ID)
				}
				if v.Select != nil {
					v.Select(r.ID)
				}
				v.Owner.Invalidate(v.Window)
			})
		})
		v.rows = append(v.rows, b)
	}
	v.host.On(events.Scroll, func(e events.Event) {
		if s, ok := e.(*events.MouseScroll); ok {
			v.ownerEvent(lease, func() {
				d := wheelRows(&v.wheel, s.Delta.Y)
				if d == 0 {
					return
				}
				start := v.Start + d
				if start < 0 {
					start = 0
				}
				rows, err := v.Model.View(start, v.Count)
				if err == nil && len(rows) > 0 {
					v.Start = start
				}
				v.Owner.Invalidate(v.Window)
			})
			e.SetHandled()
		}
	})
	return v.host
}
func (v *VirtualTree) Sync(core.Widget) error {
	if v == nil || v.Model == nil || v.host == nil || v.Count != len(v.rows) {
		return fmt.Errorf("virtual: mounted tree and fixed row-pool size required")
	}
	maximum := v.Model.VisibleLen() - v.Count
	if maximum < 0 {
		maximum = 0
	}
	if v.Start > maximum {
		v.Start = maximum
	}
	rows, err := v.Model.View(v.Start, v.Count)
	if err != nil {
		return err
	}
	v.visible = rows
	for i, b := range v.rows {
		text := ""
		invisible := i >= len(rows)
		selected := false
		if !invisible {
			r := rows[i]
			prefix := ""
			if r.HasChildren {
				prefix = "▸ "
				if r.Expanded {
					prefix = "▾ "
				}
			}
			text = html.EscapeString(strings.Repeat("  ", r.Depth) + prefix + r.Label)
			selected = r.ID == v.Model.Selected
		}
		if b.Text != text || b.StateIs(states.Invisible) != invisible || b.StateIs(states.Selected) != selected {
			b.SetText(text)
			b.SetState(invisible, states.Invisible)
			b.SetState(selected, states.Selected)
			b.Update()
		}
	}
	return nil
}

func wheelRows(remainder *float32, delta float32) int {
	*remainder += delta
	rows := int(*remainder / 24)
	*remainder -= float32(rows) * 24
	return rows
}

// One immutable mount lease per host. Reusing an extension object cannot make
// listeners from a former native pool adopt the newly mounted opening.
type virtualMountScope struct {
	owner  *app.Application
	ref    app.WindowRef
	window string
}

func (s *virtualMountScope) live(owner *app.Application, window string) bool {
	return s != nil && s.owner == owner && s.window == window && s.ref.Context().Err() == nil
}
func (v *VirtualTable) ownerEvent(lease *virtualMountScope, fn func()) error {
	if lease == nil {
		return app.ErrInvalidWindowRef
	}
	return lease.owner.Do(func() {
		if v.input == lease && lease.live(v.Owner, v.Window) {
			fn()
		}
	})
}
func (v *VirtualTree) ownerEvent(lease *virtualMountScope, fn func()) error {
	if lease == nil {
		return app.ErrInvalidWindowRef
	}
	return lease.owner.Do(func() {
		if v.input == lease && lease.live(v.Owner, v.Window) {
			fn()
		}
	})
}
