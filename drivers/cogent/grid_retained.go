package cogent

import (
	"html"
	"math"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type gridCellKey struct{ row, column string }

// The pool contains exactly the current admitted projection. Matching by domain
// identity keeps overlapping shaped text; departing slots are bounded donors for
// entering identities, rather than a cache of rows that have left the viewport.
func (g *gridBinding) retainCells(next virtual.GridSnapshot) ([]*core.Text, []*core.Text, error) {
	old := make(map[gridCellKey]*core.Text, len(g.cells))
	if len(g.cells) == len(g.projection.Cells) {
		for i, c := range g.projection.Cells {
			old[gridCellKey{c.RowKey, c.ColumnID}] = g.cells[i]
		}
	}
	keys := make(map[gridCellKey]bool, len(next.Cells))
	used := make(map[*core.Text]bool, len(g.cells))
	cells := make([]*core.Text, len(next.Cells))
	for i, c := range next.Cells {
		key := gridCellKey{c.RowKey, c.ColumnID}
		if keys[key] {
			return nil, nil, ErrGridBinding
		}
		keys[key] = true
		if w := old[key]; w != nil {
			cells[i], used[w] = w, true
		}
	}
	donors := make([]*core.Text, 0, len(g.cells))
	for _, w := range g.cells {
		if !used[w] {
			donors = append(donors, w)
		}
	}
	n := 0
	for i, w := range cells {
		if w == nil && n < len(donors) {
			cells[i] = donors[n]
			n++
		}
	}
	return cells, donors[n:], nil
}

func (g *gridBinding) retainHeaders(next virtual.GridSnapshot) ([]*core.Text, []*core.Text, error) {
	old := make(map[string]*core.Text, len(g.headers))
	if len(g.headers) == len(g.projection.Headers) {
		for i, h := range g.projection.Headers {
			old[h.ColumnID] = g.headers[i]
		}
	}
	keys := make(map[string]bool, len(next.Headers))
	used := make(map[*core.Text]bool, len(g.headers))
	headers := make([]*core.Text, len(next.Headers))
	for i, h := range next.Headers {
		if keys[h.ColumnID] {
			return nil, nil, ErrGridBinding
		}
		keys[h.ColumnID] = true
		if w := old[h.ColumnID]; w != nil {
			headers[i], used[w] = w, true
		}
	}
	donors := make([]*core.Text, 0, len(g.headers))
	for _, w := range g.headers {
		if !used[w] {
			donors = append(donors, w)
		}
	}
	n := 0
	for i, w := range headers {
		if w == nil && n < len(donors) {
			headers[i] = donors[n]
			n++
		}
	}
	return headers, donors[n:], nil
}

// Private non-cell children (currently the inline editor) remain last, above
// every cell. Assignment does not rerun native OnAdd or reset retained state.
func (g *gridBinding) orderPanes() {
	managed := make(map[tree.Node]bool, len(g.cells)+len(g.headers))
	for _, w := range g.cells {
		managed[w] = true
	}
	for _, w := range g.headers {
		managed[w] = true
	}
	for _, pane := range g.panes {
		children := make([]tree.Node, 0, len(pane.Children))
		for _, w := range g.headers {
			if w.Parent == pane {
				children = append(children, w)
			}
		}
		for _, w := range g.cells {
			if w.Parent == pane {
				children = append(children, w)
			}
		}
		for _, child := range pane.Children {
			if !managed[child] {
				children = append(children, child)
			}
		}
		pane.Children = children
	}
}

type gridLayoutStamp struct {
	valid    bool
	scene    *core.Scene
	parent   tree.Node
	geom     core.GeomState
	style    ui.Style
	scale    float32
	width    int
	height   int
	frozen   int
	header   int
	config   virtual.GridConfig
	disabled bool
	help     string
}

func (g *gridBinding) layoutStamp() gridLayoutStamp {
	style, err := ui.ResolveStyle(g.r.node, g.r.theme)
	if err != nil || g.host.Scene == nil || g.host.Parent == nil {
		return gridLayoutStamp{}
	}
	s := g.projection.State
	return gridLayoutStamp{true, g.host.Scene, g.host.Parent, g.host.Geom, style,
		g.host.Styles.UnitContext.Px(1), s.Viewport.Width, s.Viewport.Height,
		s.FrozenWidth, s.Config.HeaderHeight, s.Config, g.r.control.disabled,
		g.r.node.Accessibility.Help}
}

func (g *gridBinding) canLayoutLocally(next virtual.GridSnapshot) bool {
	if !g.layout.valid || !g.projectionReady || next.State.Edit.Active ||
		g.projection.State.Edit.Active || g.editing != nil {
		return false
	}
	current := g.layoutStamp()
	current.width, current.height = next.State.Viewport.Width, next.State.Viewport.Height
	current.frozen, current.header, current.config = next.State.FrozenWidth, next.State.Config.HeaderHeight, next.State.Config
	return current == g.layout && validGridScale(current.scale)
}

func gridPhysicalSize(r virtual.GridRect, scale float32) (int, int) {
	edge := func(v int) int { return int(math.Round(float64(v) * float64(scale))) }
	return edge(r.X+r.Width) - edge(r.X), edge(r.Y+r.Height) - edge(r.Y)
}

// Placement itself does not request layout. The accepted fixed-allocation root
// sizes only dirty text on its originating scene; ordinary fallback still runs
// the full recursive host updater/style and scene layout.
func (g *gridBinding) placeRetained(w *core.Text, pane *gridPane, rect, clip virtual.GridRect, value string, selected, cursor bool) (style, size, moved bool) {
	oldRect, hasRect := w.Property("vellum.grid.rect").(virtual.GridRect)
	oldCursor, _ := w.Property("vellum.grid.cursor").(bool)
	value = html.EscapeString(value)
	moved = w.Parent != pane
	style = moved || w.Text != value || w.StateIs(states.Selected) != selected ||
		w.StateIs(states.Invisible) != clip.Empty() || oldCursor != cursor
	oldWidth, oldHeight := gridPhysicalSize(oldRect, g.host.Styles.UnitContext.Px(1))
	width, height := gridPhysicalSize(rect, g.host.Styles.UnitContext.Px(1))
	size = style || !hasRect || oldWidth != width || oldHeight != height ||
		!clip.Empty() && w.PaintText() == nil
	g.place(w, pane, rect, clip, value, selected, cursor)
	return
}

// LayoutLocal implements Core's explicitly fixed-allocation boundary. It never
// calls the domain source or changes its parent's sizing demands. Complete
// scene painting follows, preserving sibling order and clearing old positions.
func (w *gridWidget) LayoutLocal() bool {
	g := w.binding
	if g == nil || !g.live() || !g.layout.valid || g.layoutStamp() != g.layout ||
		g.projection.State.Edit.Active || g.editing != nil {
		if g != nil && g.live() {
			g.layout.valid = false
			g.host.Update()
		}
		return false
	}
	for text := range g.layoutDirty {
		if text.This == nil || text.Scene != w.Scene {
			return false
		}
		text.SizeUp()
		text.SizeDown(0)
		text.SizeFinal()
	}
	w.Position()
	w.ApplyScenePos()
	return true
}
