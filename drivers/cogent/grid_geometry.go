package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/states"
	"encoding/base64"
	"image"
	"math"
	"slices"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/virtual"
)

// GridInspection is a detached publication from a completed native frame.
// Coordinates are clipped scene pixels; the model viewport remains logical px.
type GridInspection struct {
	Window, ID                      string
	Incarnation, Generation, Frames uint64
	CellIncarnation                 uint64
	Stamp                           virtual.GridStamp
	Viewport                        virtual.GridViewport
	Columns                         int
	RetainedCells, RetainedHeaders  int
	RowSelection                    bool
	Selection                       virtual.GridSelectionState
	Scale                           float32
	Bounds                          image.Rectangle
	Cells                           []GridCellGeometry
	Headers                         []GridHeaderGeometry
	Edit                            virtual.GridEditState
	Editor                          *GridEditorGeometry
}
type GridEditorGeometry struct {
	ID      string
	Session uint64
	Text    string
	Focused bool
	Bounds  image.Rectangle
}
type GridCellGeometry struct {
	ID, RowKey, ColumnID, Text string
	Row, Column                int
	Frozen, Active, Selected   bool
	Bounds                     image.Rectangle
}
type GridHeaderGeometry struct {
	ID, ColumnID, Title string
	Column              int
	Frozen              bool
	Bounds              image.Rectangle
}

// GridCellID encodes a generated target identity. The grid: prefix is reserved
// by ui.Validate; callers may use this ID for captured input, not Node.ID.
func GridCellID(grid, row, column string) string {
	return "grid:" + base64.RawURLEncoding.EncodeToString([]byte(grid)) + ":cell:" + base64.RawURLEncoding.EncodeToString([]byte(row)) + ":" + base64.RawURLEncoding.EncodeToString([]byte(column))
}
func gridRowID(grid, row string) string {
	return "grid:" + base64.RawURLEncoding.EncodeToString([]byte(grid)) + ":row:" + base64.RawURLEncoding.EncodeToString([]byte(row))
}
func gridHeaderID(grid, column string) string {
	return "grid:" + base64.RawURLEncoding.EncodeToString([]byte(grid)) + ":header:" + base64.RawURLEncoding.EncodeToString([]byte(column))
}
func gridViewportRequest(bounds image.Rectangle, scale float32) (gridRequest, error) {
	if !validGridScale(scale) || bounds.Dx() < 0 || bounds.Dy() < 0 {
		return gridRequest{}, ErrGridBudget
	}
	width := math.Floor(float64(bounds.Dx())/float64(scale)) - gridScrollbar
	height := math.Floor(float64(bounds.Dy())/float64(scale)) - gridScrollbar
	if math.IsNaN(width) || math.IsInf(width, 0) || math.IsNaN(height) || math.IsInf(height, 0) || width > gridNativeMaxDimension || height > gridNativeMaxDimension {
		return gridRequest{}, ErrGridBudget
	}
	return gridRequest{max(1, int(width)), max(1, int(height)), scale}, nil
}
func (g *gridBinding) capture() {
	if !g.live() || g.projection.State.Generation == 0 {
		return
	}
	if g.frames == math.MaxUint64 {
		g.withdrawMeasurement()
		return
	}
	g.frames++
	scale := g.host.Styles.UnitContext.Px(1)
	bounds := g.host.Geom.ContentBBox
	requested, err := gridViewportRequest(bounds, scale)
	if err != nil {
		g.refuseGeometry()
		return
	}
	g.geometryRefused = false
	g.requestViewport(requested)
	g.requestMu.Lock()
	request := g.requested
	g.requestMu.Unlock()
	s := g.projection.State
	if request.width != s.Viewport.Width || request.height != s.Viewport.Height {
		g.withdrawMeasurement()
		return
	}
	previous := g.measurement.Load()
	wasPending := g.layoutPending.Load() || g.host.Scene != nil && g.host.Scene.GlobalLayoutPending()
	m := &GridInspection{Window: g.window, ID: g.r.node.ID, Incarnation: g.incarnation, Generation: s.Generation, CellIncarnation: g.epoch, Frames: g.frames, Stamp: s.Stamp, Viewport: s.Viewport, Columns: s.Columns, RetainedCells: g.retainedCellLayouts(), RetainedHeaders: len(g.headers), RowSelection: s.Config.RowSelection, Selection: s.Selection, Scale: g.host.Styles.UnitContext.Px(1), Bounds: g.host.Geom.ContentBBox, Cells: make([]GridCellGeometry, len(g.projection.Cells)), Headers: make([]GridHeaderGeometry, len(g.projection.Headers))}
	offset := image.Point{}
	if g.host.Scene != nil {
		offset = g.host.Scene.SceneGeom.Pos
	}
	m.Bounds = m.Bounds.Add(offset)
	if g.cellSurface != nil && !g.cellSurface.completed {
		g.withdrawMeasurement()
		return
	}
	for i, c := range g.projection.Cells {
		var bounds image.Rectangle
		if g.cellSurface != nil {
			bounds = g.cellSurface.cells[i].box.Add(offset)
		} else {
			bounds = g.cells[i].Geom.TotalBBox.Add(offset)
		}
		if c.Clip.Empty() {
			bounds = image.Rectangle{}
		}
		m.Cells[i] = GridCellGeometry{ID: GridCellID(m.ID, c.RowKey, c.ColumnID), RowKey: c.RowKey, ColumnID: c.ColumnID, Text: c.Text, Row: c.RowIndex, Column: c.ColumnIndex, Frozen: c.Frozen, Active: c.Active, Selected: c.Selected, Bounds: bounds}
	}
	for i, h := range g.projection.Headers {
		bounds := g.headers[i].Geom.TotalBBox.Add(offset)
		if h.Clip.Empty() {
			bounds = image.Rectangle{}
		}
		m.Headers[i] = GridHeaderGeometry{gridHeaderID(m.ID, h.ColumnID), h.ColumnID, h.Title, h.ColumnIndex, h.Frozen, bounds}
	}
	m.Edit = s.Edit
	if edit := gridEditingRecord(g.r); edit != nil {
		field := edit.widget.(*core.TextField)
		state := field.EditingState()
		m.Editor = &GridEditorGeometry{ID: edit.node.ID, Session: s.Edit.Session, Text: state.Text, Focused: field.StateIs(states.Focused), Bounds: field.Geom.TotalBBox.Add(offset)}
	}
	g.layout = g.layoutStamp()
	g.measurementScene.Store(g.host.Scene)
	g.measurement.Store(m)
	g.layoutPending.Store(false)
	if wasPending || !sameGridAccessibilityMeasurement(previous, m) {
		_ = g.owner.Invalidate(g.window)
	}
}

// Notify once when a completed frame becomes unavailable, even when a viewport
// request is refused or is a no-op. This revokes its owner/native AX actions.
func (g *gridBinding) withdrawMeasurement() {
	if g.measurement.Swap(nil) != nil {
		_ = g.owner.Invalidate(g.window)
	}
}

// Frame ordinals advance on every render. Only changed completed data/geometry
// need an owner publication; resize, scale and editor movement also matter when
// the portable grid generation is unchanged.
func sameGridAccessibilityMeasurement(previous, current *GridInspection) bool {
	if previous == nil || current == nil {
		return previous == current
	}
	// Check scalar provenance first: an ordinary scroll changes generation and
	// can return before visiting any cell. Keep every publication field except
	// Frames here when extending GridInspection.
	if previous.Window != current.Window || previous.ID != current.ID ||
		previous.Incarnation != current.Incarnation || previous.Generation != current.Generation ||
		previous.CellIncarnation != current.CellIncarnation || previous.Stamp != current.Stamp ||
		previous.Viewport != current.Viewport || previous.Columns != current.Columns ||
		previous.RetainedCells != current.RetainedCells || previous.RetainedHeaders != current.RetainedHeaders ||
		previous.RowSelection != current.RowSelection || previous.Selection != current.Selection ||
		previous.Scale != current.Scale || previous.Bounds != current.Bounds || previous.Edit != current.Edit {
		return false
	}
	if (previous.Editor == nil) != (current.Editor == nil) ||
		previous.Editor != nil && *previous.Editor != *current.Editor {
		return false
	}
	// Preserve nil/empty distinctions from the previous comparison, without a
	// reflective walk over completed cell/header geometry on the render thread.
	return (previous.Cells == nil) == (current.Cells == nil) &&
		(previous.Headers == nil) == (current.Headers == nil) &&
		slices.Equal(previous.Cells, current.Cells) && slices.Equal(previous.Headers, current.Headers)
}

// Current retained geometry is also checked by native hit/AX projection. A
// pending layout request cannot reuse a last-frame cell merely because its
// portable generation has not changed yet.
func (g *gridBinding) currentMeasurement() *GridInspection {
	m := g.measurement.Load()
	if g.layoutPending.Load() || m == nil || m.Generation != g.projection.State.Generation || m.Stamp != g.projection.State.Stamp || m.Viewport != g.projection.State.Viewport {
		return nil
	}
	if scene := g.measurementScene.Load(); scene != nil && scene.GlobalLayoutPending() {
		return nil
	}
	// Publication precedes clearing local pending. Refuse an older pointer if a
	// concurrent completed frame changed it while the fences were inspected.
	if g.measurement.Load() != m {
		return nil
	}
	g.requestMu.Lock()
	request := g.requested
	g.requestMu.Unlock()
	if request.width != m.Viewport.Width || request.height != m.Viewport.Height || request.scale != m.Scale {
		return nil
	}
	return m
}

// InspectGrid must be called on the application owner (Do/PostWait). It does not
// render, dispatch, mutate the model, or fetch Cell strings from its source.
func InspectGrid(a *app.Application, model *virtual.Grid) (GridInspection, error) {
	gridRegistry.Lock()
	g := gridRegistry.models[model]
	gridRegistry.Unlock()
	if g == nil {
		return GridInspection{}, ErrGridClosed
	}
	if g.owner != a {
		return GridInspection{}, ErrGridBinding
	}
	if !g.live() || g.r.control.retired {
		return GridInspection{}, ErrGridClosed
	}
	s, err := model.State()
	if err != nil {
		return GridInspection{}, err
	}
	m := g.currentMeasurement()
	if m == nil || m.Generation != s.Generation || m.Stamp != s.Stamp || m.Viewport != s.Viewport {
		return GridInspection{}, ErrGridLayout
	}
	g.requestMu.Lock()
	request := g.requested
	g.requestMu.Unlock()
	if request.width != m.Viewport.Width || request.height != m.Viewport.Height || request.scale != m.Scale {
		return GridInspection{}, ErrGridLayout
	}
	out := *m
	if m.Editor != nil {
		e := *m.Editor
		out.Editor = &e
	}
	out.Cells = append([]GridCellGeometry(nil), m.Cells...)
	out.Headers = append([]GridHeaderGeometry(nil), m.Headers...)
	return out, nil
}
