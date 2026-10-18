package cogent

import (
	"html"
	"image"
	"math"
	"reflect"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/htmltext"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/text"
	"github.com/dshills/vellum/virtual"
)

// A scene may select the ordinary widget route as an independent reference.
const gridCellSurfaceProperty = "vellum.driver.reference_grid_cell_widgets"

// Each entry belongs to the current admitted projection only. There is no
// detached widget, offscreen-row cache, font key or cross-frame paint fragment.
type gridCellLayout struct {
	key         gridCellKey
	text        string
	shape       *shaped.Lines
	shapeEpoch  uint64
	pane, class int
	box         image.Rectangle
	position    math32.Vector2
	content     math32.Vector2
	size        math32.Vector2
}

type gridCellSurface struct {
	cells      []gridCellLayout
	styles     [2][4]styles.Style
	styleReady [2]bool
	epoch      [2]uint64
	dirty      [2]bool
	prepared   [2]bool
	completed  bool
	failed     bool
}

func (g *gridBinding) wantsCellSurface(next virtual.GridSnapshot) bool {
	if g.host.Scene == nil || !g.host.Scene.PassiveTextSurfaceAllowed() || !g.layout.valid ||
		!g.projectionReady || g.cellSurface != nil && g.cellSurface.failed || next.State.Edit.Active || g.editing != nil ||
		g.host.Scene.GlobalLayoutPending() || !g.canLayoutLocally(next) {
		return false
	}
	if reference, _ := g.host.Scene.Property(gridCellSurfaceProperty).(bool); reference {
		return false
	}
	if experimental, _ := g.host.Scene.Property(gridScrollReuseProperty).(bool); experimental {
		return false
	}
	if core.DebugSettings.RenderTrace {
		return false
	}
	for _, p := range g.panes[2:] {
		if p.Parts != nil {
			return false
		}
		for _, child := range p.Children {
			found := false
			for _, cell := range g.cells {
				if cell == child {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		for class := 0; class < 4; class++ {
			if _, ok := gridSurfaceCellStyle(p, class&1 != 0, class&2 != 0); !ok {
				return false
			}
		}
	}
	return true
}

func gridSurfaceAbsolute(v units.Value) bool {
	if v.Custom != nil || math.IsNaN(float64(v.Value)) || math.IsInf(float64(v.Value), 0) {
		return false
	}
	switch v.Unit {
	case units.UnitDp, units.UnitPx, units.UnitCm, units.UnitMm, units.UnitQ, units.UnitIn, units.UnitPc, units.UnitPt, units.UnitDot:
		return true
	}
	return false
}

func (g *gridBinding) failCellSurface() {
	if g.cellSurface == nil || g.cellSurface.failed {
		return
	}
	g.cellSurface.failed = true
	g.cellSurface.completed = false
	g.layoutPending.Store(true)
	g.layout.valid = false
	g.withdrawMeasurement()
	_ = g.owner.Invalidate(g.window)
}

// Mirror the pinned passive Text stylers, without constructing a Widget or
// running arbitrary updaters. Only managed absolute-unit inherited styles admit.
// The independent original Text route is the style/pixel oracle.
func gridSurfaceCellStyle(p *gridPane, selected, cursor bool) (styles.Style, bool) {
	var s styles.Style
	parent := &p.Styles
	if p.Scene == nil || parent.Font.Size.Dots <= 0 ||
		!gridSurfaceAbsolute(parent.Font.Size) || !gridSurfaceAbsolute(parent.Text.Indent) || !gridSurfaceAbsolute(parent.Text.Hanging) ||
		math.IsNaN(float64(parent.Opacity)) || math.IsInf(float64(parent.Opacity), 0) {
		return s, false
	}
	if _, ok := opaqueGridUniform(parent.Color); !ok {
		return s, false
	}
	s.Defaults()
	s.SetState(selected, states.Selected)
	s.StateLayer = s.State.StateLayer()
	s.Font.Family = rich.SansSerif
	s.InheritFields(parent)
	// WidgetBase's normal selected styling precedes Text's normal styling.
	if selected {
		s.Background = colors.Scheme.Select.Container
		s.Color = colors.Scheme.Select.OnContainer
	}
	s.GrowWrap = true
	s.Text.LineHeight = 24.0 / 16
	s.Font.Weight = rich.Normal
	s.Padding.Zero()
	s.Margin.Zero()
	s.Border.Width.Zero()
	if cursor {
		s.Border.Style.Set(styles.BorderSolid)
		s.Border.Width.Set(units.Px(1))
		s.Border.Color.Set(colors.Scheme.Primary.Base)
	}
	s.Grow.SetZero()
	s.Text.WhiteSpace = text.WhiteSpacePre
	s.Abilities = 0
	// WidgetBase's final styler removes the unused focus reservation.
	s.MaxBorder = s.Border
	s.UnitContext = parent.UnitContext
	s.UnitContext.SetSizes(float32(p.Scene.SceneGeom.Size.X), float32(p.Scene.SceneGeom.Size.Y), 0, 0, p.Geom.Size.Alloc.Content.X, p.Geom.Size.Alloc.Content.Y)
	s.Font.ToDots(&s.UnitContext)
	s.Font.SetUnitContext(&s.UnitContext)
	s.ToDots()
	if !plainGridBox(&s) && (!cursor || s.Opacity != 1 || len(s.BoxShadow) != 0 || len(s.MaxBoxShadow) != 0 || !sides.AreZero(s.Border.Radius.Dots().Sides)) {
		return styles.Style{}, false
	}
	return s, true
}

func (g *gridBinding) retireCellSurface() {
	if g.cellSurface != nil {
		clear(g.cellSurface.cells)
		*g.cellSurface = gridCellSurface{}
		g.cellSurface = nil
	}
}

func (g *gridBinding) syncCellSurface(next virtual.GridSnapshot, local bool) error {
	previous := g.cellSurface
	surface := &gridCellSurface{}
	if previous != nil {
		surface.styles, surface.styleReady, surface.epoch, surface.dirty = previous.styles, previous.styleReady, previous.epoch, previous.dirty
	}
	old := make(map[gridCellKey]gridCellLayout)
	if previous != nil {
		for _, entry := range previous.cells {
			old[entry.key] = entry
		}
	}
	seen := make(map[gridCellKey]bool, len(next.Cells))
	surface.cells = make([]gridCellLayout, len(next.Cells))
	for i, cell := range next.Cells {
		key := gridCellKey{cell.RowKey, cell.ColumnID}
		if seen[key] {
			return ErrGridBinding
		}
		seen[key] = true
		entry := old[key]
		if entry.text != cell.Text {
			entry.shape, entry.shapeEpoch = nil, 0
		}
		entry.key, entry.text = key, cell.Text
		oldPane := entry.pane
		entry.pane = 3
		if cell.Frozen {
			entry.pane = 2
		}
		if oldPane != entry.pane {
			entry.shape, entry.shapeEpoch = nil, 0
		}
		selected := cell.Active
		if next.State.Config.RowSelection {
			selected = cell.Selected
		}
		class := 0
		if selected {
			class = 1
		}
		if cell.Active && next.State.Config.RowSelection {
			class |= 2
		}
		if entry.class != class {
			entry.shape, entry.shapeEpoch = nil, 0
		}
		entry.class = class
		surface.cells[i] = entry
	}
	headers, retired, err := g.retainHeaders(next)
	if err != nil {
		return err
	}
	// All eligibility/allocation checks finish before replacing live children.
	g.layoutPending.Store(true)
	for _, cell := range g.cells {
		delete(g.layoutDirty, cell)
		cell.Delete()
	}
	g.cells = nil
	for _, header := range retired {
		delete(g.layoutDirty, header)
		header.Delete()
	}
	g.projection = next
	g.headers = headers
	g.cellSurface = surface
	if previous != nil {
		clear(previous.cells)
		*previous = gridCellSurface{}
	}
	for i, header := range next.Headers {
		w := g.headers[i]
		created := w == nil
		if created {
			w = core.NewText(g.panes[0])
			g.installTextStyle(w)
			g.headers[i] = w
		}
		pane := 1
		if header.Frozen {
			pane = 0
		}
		sty, size, moved := g.placeRetained(w, g.panes[pane], header.Rect, header.Clip, header.Title, false, false)
		if moved && !created {
			local = false
		}
		if size {
			if g.layoutDirty == nil {
				g.layoutDirty = make(map[*core.Text]bool)
			}
			g.layoutDirty[w] = true
		}
		if sty || created {
			w.UpdateWidget()
			w.Style()
		}
	}
	g.syncScrollbars(next.State)
	g.orderPanes()
	if !local || !g.host.Scene.NeedsLocalLayout(g.host) {
		g.layout.valid = false
		g.host.Update()
	}
	g.projectedDisabled, g.projectedHelp = g.r.control.disabled, g.r.node.Accessibility.Help
	g.projectionReady = true
	return nil
}

func (p *gridPane) Style() {
	p.Frame.Style()
	if p.part >= 2 && p.binding != nil && p.binding.cellSurface != nil {
		p.binding.cellSurface.dirty[p.part-2] = true
	}
}

func (p *gridPane) ApplyScenePos() {
	p.Frame.ApplyScenePos()
	g := p.binding
	if p.part < 2 || g == nil || g.cellSurface == nil {
		return
	}
	surface := g.cellSurface
	if surface.failed || len(p.Children) != 0 || p.Parts != nil || !p.Scene.PassiveTextSurfaceAllowed() {
		g.failCellSurface()
		return
	}
	pi := int(p.part) - 2
	var current [4]styles.Style
	for class := range current {
		style, ok := gridSurfaceCellStyle(p, class&1 != 0, class&2 != 0)
		if !ok {
			g.failCellSurface()
			return
		}
		current[class] = style
	}
	if surface.dirty[pi] || !surface.styleReady[pi] || !reflect.DeepEqual(surface.styles[pi], current) {
		if surface.epoch[pi] == math.MaxUint64 {
			g.failCellSurface()
			return
		}
		surface.epoch[pi]++
		surface.styles[pi], surface.styleReady[pi], surface.dirty[pi] = current, true, false
	}
	scale := p.Styles.UnitContext.Px(1)
	edge := func(v int) float32 { return float32(math.Round(float64(v) * float64(scale))) }
	for i := range surface.cells {
		entry := &surface.cells[i]
		if entry.pane != int(p.part) {
			continue
		}
		cell := g.projection.Cells[i]
		s := &surface.styles[pi][entry.class]
		if entry.shape == nil || entry.shapeEpoch != surface.epoch[pi] {
			sty, tsty := s.NewRichText()
			source, _ := htmltext.HTMLPreToRich([]byte(html.EscapeString(entry.text)), sty, nil)
			tsty.Align, tsty.AlignV = text.Start, text.Start
			entry.shape = p.Scene.TextShaper().WrapLines(source, sty, tsty, math32.Vector2{})
			entry.shapeEpoch = surface.epoch[pi]
		}
		r := cell.Rect
		entry.position = p.Geom.Pos.Content.Add(p.Geom.Scroll).Add(math32.Vec2(edge(r.X)-edge(p.rect.X), edge(r.Y)-edge(p.rect.Y)))
		entry.size = math32.Vec2(edge(r.X+r.Width)-edge(r.X), edge(r.Y+r.Height)-edge(r.Y))
		entry.content = entry.position.Add(s.BoxSpace().Pos().Floor())
		entry.box = math32.RectFromPosSizeMax(entry.position, entry.size).Intersect(p.Geom.ContentBBox)
		if cell.Clip.Empty() {
			entry.box = image.Rectangle{}
		}
	}
	surface.prepared[pi] = true
	surface.completed = surface.prepared[0] && surface.prepared[1]
}

func (p *gridPane) renderCellSurface() bool {
	g := p.binding
	if p.part < 2 || g == nil || g.cellSurface == nil {
		return false
	}
	if g.cellSurface.failed || !g.cellSurface.completed || len(p.Children) != 0 || p.Parts != nil || g.cellSurface.dirty[0] || g.cellSurface.dirty[1] || !p.Scene.PassiveTextSurfaceAllowed() {
		g.failCellSurface()
		return true
	}
	if !p.StartRender() {
		return true
	}
	defer p.EndRender()
	p.Frame.Render()
	pc := &p.Scene.Painter
	var batch *render.TextBatch
	flush := func() {
		if batch != nil && batch.Len() != 0 {
			pc.Render.Add(batch)
		}
		batch = nil
	}
	for i := range g.cellSurface.cells {
		entry := &g.cellSurface.cells[i]
		if entry.pane != int(p.part) || entry.box.Empty() || entry.shape == nil {
			continue
		}
		style := g.cellSurface.styles[int(p.part)-2][entry.class]
		s := &style
		s.ComputeActualBackground(p.Styles.ActualBackground)
		background := s.Background
		if background == nil {
			background = p.Styles.ActualBackground
		}
		color, opaque := opaqueGridUniform(background)
		parent, parentOpaque := opaqueGridUniform(p.Styles.ActualBackground)
		if plainGridBox(s) && opaque && parentOpaque && color == parent {
			if batch == nil {
				batch = render.NewTextBatch(pc.Context())
			}
			batch.Add(entry.shape, entry.content, entry.box)
			continue
		}
		flush()
		pc.PushContext(nil, render.NewBoundsRect(entry.box, s.Border.Radius.Dots()))
		pc.Paint.Defaults()
		pc.StandardBox(s, entry.position, entry.size, p.Styles.ActualBackground)
		pc.DrawText(entry.shape, entry.content)
		pc.PopContext()
	}
	flush()
	return true
}

func (g *gridBinding) retainedCellLayouts() int {
	if g.cellSurface != nil {
		return len(g.cellSurface.cells)
	}
	return len(g.cells)
}

func (w *gridWidget) ApplyScenePos() {
	if w.binding != nil && w.binding.cellSurface != nil {
		w.binding.cellSurface.prepared = [2]bool{}
		w.binding.cellSurface.completed = false
	}
	w.Frame.ApplyScenePos()
}
