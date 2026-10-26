package demo

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"slices"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type Shape struct {
	ID, Name, Kind string
	X, Y, Size     float64
	Color          color.NRGBA
}
type formaState struct {
	Shapes   []Shape
	Template int
	Selected string
}
type Forma struct {
	owner
	state   formaState
	next    uint64
	journal *history.Journal
	surface *cogent.MeshCanvas
	view    canvas.Viewport
	draft   float32
}

func NewForma() (*Forma, error) {
	j, e := history.New(64)
	if e != nil {
		return nil, e
	}
	m := &Forma{journal: j, next: 4, view: canvas.Viewport{Width: 800, Height: 560, Center: canvas.Point{X: 400, Y: 280}, Zoom: 1}, draft: 164}
	m.state = formaState{Shapes: []Shape{{"shape-1", "Cobalt orbit", "circle", 570, 215, 246, rgb(41, 95, 235)}, {"shape-2", "Midnight block", "square", 667, 337, 112, rgb(34, 48, 78)}, {"shape-3", "Coral accent", "circle", 598, 350, 164, rgb(246, 115, 83)}}, Selected: "shape-3"}
	m.surface = &cogent.MeshCanvas{View: &m.view, Background: rgb(232, 235, 241), Prepare: m.prepare, Pick: m.pick}
	m.refresh()
	m.register("forma.add", "Add circle", "Meta+N", func() error {
		if len(m.state.Shapes) >= 48 {
			return errors.New("shape limit reached")
		}
		return nil
	}, func() error { return m.Add("circle") })
	m.register("forma.undo", "Undo", "Meta+Z", func() error {
		if j.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, m.Undo)
	m.register("forma.redo", "Redo", "Meta+Shift+Z", func() error {
		if j.Snapshot().RedoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, m.Redo)
	return m, nil
}
func (m *Forma) Bind(a *app.Application, r app.WindowRef) {
	m.bind(a, r)
	m.surface.Owner = a
	m.surface.Window = r.ID()
}
func (m *Forma) refresh() {
	m.surface.Image = &canvas.ImageLayer{Image: posterAssets()[m.state.Template], Bounds: canvas.Bounds{Min: canvas.Point{}, Max: canvas.Point{X: 800, Y: 560}}, Filter: canvas.ImageLinear}
	m.surface.Revision++
	for _, s := range m.state.Shapes {
		if s.ID == m.state.Selected {
			m.draft = float32(s.Size)
		}
	}
	m.changed()
}
func (m *Forma) apply(title string, after formaState) error {
	before := m.state
	before.Shapes = slices.Clone(before.Shapes)
	after.Shapes = slices.Clone(after.Shapes)
	return m.journal.Apply(history.Change{Label: title, Apply: func() error { m.state = after; m.refresh(); return nil }, Revert: func() error { m.state = before; m.refresh(); return nil }})
}
func (m *Forma) Add(kind string) error {
	if kind != "circle" && kind != "square" {
		return errors.New("unknown shape")
	}
	if len(m.state.Shapes) >= 48 {
		return errors.New("shape limit reached")
	}
	id := fmt.Sprintf("shape-%d", m.next)
	after := m.state
	after.Shapes = append(slices.Clone(after.Shapes), Shape{id, "New " + kind, kind, 440 + float64(m.next%5)*28, 250 + float64(m.next%4)*24, 100, rgb(242, 176, 72)})
	after.Selected = id
	if e := m.apply("Add "+kind, after); e != nil {
		return e
	}
	m.next++
	return nil
}
func (m *Forma) Select(id string) {
	for _, s := range m.state.Shapes {
		if s.ID == id {
			m.state.Selected = id
			m.refresh()
			return
		}
	}
}
func (m *Forma) ChangeSelected(c color.NRGBA, size float64, dx float64) error {
	after := m.state
	after.Shapes = slices.Clone(after.Shapes)
	for i, s := range after.Shapes {
		if s.ID == after.Selected {
			if size < 40 || size > 300 || math.IsNaN(size) {
				return errors.New("size outside40..300")
			}
			if c == s.Color && size == s.Size && dx == 0 {
				return nil
			}
			after.Shapes[i].Color = c
			after.Shapes[i].Size = size
			after.Shapes[i].X = math.Max(40, math.Min(760, s.X+dx))
			return m.apply("Edit "+s.Name, after)
		}
	}
	return errors.New("select a shape")
}
func (m *Forma) selected() Shape {
	for _, s := range m.state.Shapes {
		if s.ID == m.state.Selected {
			return s
		}
	}
	return Shape{}
}
func (m *Forma) Undo() error { return m.journal.Undo() }
func (m *Forma) Redo() error { return m.journal.Redo() }
func (m *Forma) Fit() {
	m.view.Center = canvas.Point{X: 400, Y: 280}
	m.view.Zoom = .88 * math.Min(m.view.Width/800, m.view.Height/560)
	m.changed()
}
func (m *Forma) pick(p canvas.Point) {
	for i := len(m.state.Shapes) - 1; i >= 0; i-- {
		s := m.state.Shapes[i]
		hit := math.Abs(p.X-s.X) <= s.Size/2 && math.Abs(p.Y-s.Y) <= s.Size/2
		if s.Kind == "circle" {
			hit = math.Hypot(p.X-s.X, p.Y-s.Y) <= s.Size/2
		}
		if hit {
			m.Select(s.ID)
			return
		}
	}
}
func (m *Forma) prepare(canvas.Viewport) (*canvas.Mesh, error) {
	b := &meshBuilder{}
	for _, s := range m.state.Shapes {
		if s.Kind == "circle" {
			b.circle(s.X, s.Y, s.Size/2, s.Color)
		} else {
			b.quad(s.X-s.Size/2, s.Y-s.Size/2, s.Size, s.Size, s.Color)
		}
		if s.ID == m.state.Selected {
			r := s.Size/2 + 6
			p := []canvas.Point{{X: s.X - r, Y: s.Y - r}, {X: s.X + r, Y: s.Y - r}, {X: s.X + r, Y: s.Y + r}, {X: s.X - r, Y: s.Y + r}, {X: s.X - r, Y: s.Y - r}}
			b.stroke(p, 1.5, blue)
			for _, pt := range p[:4] {
				b.quad(pt.X-3, pt.Y-3, 6, 6, white)
				b.stroke([]canvas.Point{{X: pt.X - 3, Y: pt.Y - 3}, {X: pt.X + 3, Y: pt.Y - 3}, {X: pt.X + 3, Y: pt.Y + 3}, {X: pt.X - 3, Y: pt.Y + 3}, {X: pt.X - 3, Y: pt.Y - 3}}, 1, blue)
			}
		}
	}
	return b.mesh()
}
func (m *Forma) Toolbar() []cogent.TitlebarToolbarItem {
	return []cogent.TitlebarToolbarItem{{ID: "add", Label: "Add circle", Kind: ui.Button, PNG: toolbarIcon(blue, "add"), Command: "forma.add"}, {ID: "space", Kind: ui.ToolbarFlexibleSpace}, {ID: "undo", Label: "Undo", Kind: ui.Button, PNG: toolbarIcon(rgb(139, 100, 207), "reset"), Command: "forma.undo"}, {ID: "redo", Label: "Redo", Kind: ui.Button, PNG: toolbarIcon(rgb(52, 167, 165), "reset"), Command: "forma.redo"}, {ID: "fit", Label: "Fit artboard", Kind: ui.Button, PNG: toolbarIcon(rgb(245, 171, 54), "folder"), Activate: m.Fit}}
}
func (m *Forma) View() ui.Node {
	left := col("forma-sidebar", pad(col("forma-brand", label("forma-brand-title", "Forma", 24, 700, ink), label("forma-brand-sub", "A space for ideas", 12, 400, muted)), 16), rule("forma-rule1"))
	projects := section("forma-projects", "WORKSPACE", nav("forma-poster", "Orbital studies", m.state.Template == 0, blue, func() { a := m.state; a.Template = 0; _ = m.apply("Change paper", a) }), nav("forma-blueprint", "Blue hour", m.state.Template == 1, rgb(138, 112, 221), func() { a := m.state; a.Template = 1; _ = m.apply("Change paper", a) }))
	left.Children = append(left.Children, pad(projects, 12), rule("forma-rule2"))
	layers := section("forma-layers", "LAYERS")
	for i := len(m.state.Shapes) - 1; i >= 0; i-- {
		s := m.state.Shapes[i]
		n := nav("layer-"+s.ID, s.Name, s.ID == m.state.Selected, s.Color, func() { m.Select(s.ID) })
		n.IconImage = icon(s.Color, s.Kind)
		layers.Children = append(layers.Children, n)
	}
	scroll := ui.Node{ID: "forma-layer-scroll", Kind: ui.Scroll, Grow: true, Children: []ui.Node{pad(layers, 12)}}
	left.Children = append(left.Children, scroll, pad(col("forma-sidebar-foot", label("forma-local", "Local design study", 11, 500, muted), label("forma-shape-count", fmt.Sprintf("%d vector layers", len(m.state.Shapes)), 11, 400, muted)), 14))
	left = width(left, 202)
	left.Style.GrowY = ui.Some(float32(1))
	left.Style.Background = ui.Some(rgb(247, 248, 250))
	head := row("forma-heading", col("forma-heading-text", label("forma-kicker", "POSTER EXPLORATIONS / 001", 10, 600, muted), label("forma-document-title", "Make some space", 23, 600, ink)), spacer("forma-heading-space"), pill("forma-status", "Design study", rgb(78, 100, 165), rgb(233, 238, 249)), button("forma-fit", "Fit artboard", m.Fit))
	head = pad(head, 22)
	surface := ui.Node{ID: "forma-canvas", Kind: ui.Extension, Grow: true, Custom: m.surface, Accessibility: ui.Accessibility{Name: "Poster canvas", Help: "Drag to pan, scroll to zoom. Select a layer here or in the Layers list."}}
	footer := pad(row("forma-footer", label("forma-coordinates", "800 × 560  /  RGB", 11, 500, muted), spacer("forma-footer-space"), label("forma-navigation", "Drag to pan  ·  Scroll to zoom", 11, 400, muted)), 14)
	center := col("forma-center", head, surface, footer)
	center.Style.GrowY = ui.Some(float32(1))
	center.Style.Background = ui.Some(rgb(252, 252, 253))
	selected := m.selected()
	inspector := pad(col("forma-inspector", label("forma-inspector-title", "Inspector", 15, 600, ink), rule("forma-inspector-rule"), section("forma-selection", "SELECTED LAYER", label("forma-selection-name", selected.Name, 18, 600, ink), label("forma-selection-kind", selected.Kind+" / vector geometry", 12, 400, muted))), 20)
	colors := row("forma-swatches")
	for _, sw := range []struct {
		name string
		c    color.NRGBA
	}{{"Coral", rgb(246, 115, 83)}, {"Cobalt", rgb(41, 95, 235)}, {"Ink", rgb(34, 48, 78)}, {"Honey", rgb(242, 176, 72)}} {
		n := button("swatch-"+sw.name, "", func() { s := m.selected(); _ = m.ChangeSelected(sw.c, s.Size, 0) })
		n.Accessibility.Name = "Fill " + sw.name
		n.Style.Background = ui.Some(sw.c)
		n.Style.CornerRadius = ui.Some(float32(14))
		n.Style.MinSize = ui.Some(ui.Size{Width: 26, Height: 26})
		n.Style.MaxSize = ui.Some(ui.Size{Width: 26, Height: 26})
		n.Style.Padding = ui.Some(ui.Insets{})
		n.Style.Border = ui.Some(ui.Border{})
		colors.Children = append(colors.Children, n)
	}
	inspector.Children = append(inspector.Children, section("forma-fill", "FILL COLOR", colors), rule("forma-rule3"), section("forma-size", "TRANSFORM", label("forma-size-value", fmt.Sprintf("%.0f × %.0f pt", m.draft, m.draft), 26, 400, ink), ui.Node{ID: "forma-size-slider", Kind: ui.Slider, Number: m.draft, Numeric: ui.Numeric{Min: 40, Max: 300, Step: 4}, Accessibility: ui.Accessibility{Name: "Selected layer size"}, Change: func(e ui.ChangeEvent) { m.draft = e.Number; m.changed() }, Style: ui.Style{GrowX: ui.Some(float32(1))}}, button("forma-apply-size", "Apply size", func() { s := m.selected(); _ = m.ChangeSelected(s.Color, float64(m.draft), 0) }), button("forma-nudge", "Nudge right", func() { s := m.selected(); _ = m.ChangeSelected(s.Color, s.Size, 16) })), rule("forma-rule4"), section("forma-insert", "INSERT", button("forma-add-circle", "Add circle", func() { _ = m.Add("circle") }), button("forma-add-square", "Add square", func() { _ = m.Add("square") })), spacer("forma-inspector-space"), label("forma-tip", "Every edit is undoable.", 12, 500, muted), label("forma-shortcuts", "⌘Z Undo   ⇧⌘Z Redo", 11, 400, muted))
	inspector = width(inspector, 218)
	inspector.Style.GrowY = ui.Some(float32(1))
	inspector.Style.Background = ui.Some(white)
	inspector.Style.Gap = ui.Some(float32(22))
	return ui.Node{ID: "forma-root", Kind: ui.Row, Grow: true, Children: []ui.Node{left, center, inspector}, Style: ui.Style{Gap: ui.Some(float32(1)), Background: ui.Some(line)}}

}
