package cogent

import (
	"errors"
	"fmt"
	"html"
	"image"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/system"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/workbench"
)

func (w *Workbench) layout() (workbench.Layout, error) {
	if w.Group == nil {
		return w.Workspace.Layout(), nil
	}
	win, ok := w.Group.Session.Window(w.SessionWindow)
	if !ok {
		return workbench.Layout{}, fmt.Errorf("workbench: session window absent")
	}
	w.bounds = win.Bounds
	w.title = win.Title
	w.menuWindows = w.Group.Session.Layout().Windows
	return workbench.Layout{Version: workbench.Version, Root: win.Root}, nil
}

// reportError runs only on the owner. Unhandled errors are returned by Sync
// to the driver's central failure path, rather than disappearing in a callback.
func (w *Workbench) reportError(err error) bool {
	if w.OnError != nil {
		w.OnError(err)
		return true
	}
	if w.Group != nil && w.Group.OnError != nil {
		w.Group.OnError(err)
		return true
	}
	return false
}

func (w *Workbench) queue(change func() error) {
	if w.retired.Load() {
		return
	}
	generation := w.generation.Load()
	err := w.Owner.PostWindow(w.windowRef, func() {
		if w.retired.Load() || generation != w.generation.Load() {
			return
		}
		if err := change(); err != nil && !w.reportError(err) {
			w.pendingError.CompareAndSwap(nil, &err)
		}
		w.Owner.Invalidate(w.Window)
	})
	if err != nil && !errors.Is(err, app.ErrClosed) && !errors.Is(err, app.ErrWindowClosed) {
		w.pendingError.CompareAndSwap(nil, &err)
		w.Owner.Invalidate(w.Window)
	}
}
func (w *Workbench) syncFocus() {
	if w.Group == nil {
		return
	}
	f := w.Group.Session.Layout().Focus
	if f.Window != w.SessionWindow || f.Serial == w.focusSerial && f.Panel == w.focusPanel {
		return
	}
	if button := w.buttons[f.Panel]; button != nil {
		button.SetFocus()
		if native := w.host.Scene.RenderWindow(); !w.isManual() && native != nil {
			native.SystemWindow.Raise()
		}
		w.focusSerial = f.Serial
		w.focusPanel = f.Panel
	}
}
func (w *Workbench) bindWindowGeometry() {
	if w.Group == nil {
		return
	}
	if w.isManual() {
		w.host.OnShow(func(events.Event) {
			if w.positioned {
				return
			}
			w.positioned = true
			w.queue(func() error {
				if w.Group.OnWindowReady != nil {
					w.Group.OnWindowReady(w.SessionWindow)
				}
				return nil
			})
		})
		return
	}
	w.host.OnShow(func(events.Event) {
		if w.positioned {
			return
		}
		native := w.host.Scene.RenderWindow()
		if native == nil {
			return
		}
		w.positioned = true
		native.SystemWindow.SetName(w.Window)
		w.syncWindowGeometry()
		capture := func(e events.Event) {
			if e.Type() == events.Window {
				we, ok := e.(*events.WindowEvent)
				if !ok || we.Action != events.WinMove {
					return
				}
			}
			size := native.SystemWindow.WinSize()
			pos := native.SystemWindow.Position(nil)
			b := workbench.Bounds{X: pos.X, Y: pos.Y, Width: size.X, Height: size.Y, Positioned: true}
			w.queueWindowBounds(b)
		}
		w.host.Scene.On(events.WindowResize, capture)
		w.host.Scene.On(events.Window, capture)
		w.queue(func() error {
			if w.Group.OnWindowReady != nil {
				w.Group.OnWindowReady(w.SessionWindow)
			}
			return nil
		})
	})
}

// One pending snapshot and one scheduled flush bound resize persistence to
// ten writes per second. A final native close captures current geometry directly.
const nativeBoundsInterval = 100 * time.Millisecond

type nativeBoundsCapture struct {
	bounds   workbench.Bounds
	epoch    uint64
	sequence uint64
}

func validNativeBounds(b workbench.Bounds) bool {
	return b.Width >= 160 && b.Height >= 96 && b.Width <= 16384 && b.Height <= 16384 &&
		b.X >= -1000000 && b.X <= 1000000 && b.Y >= -1000000 && b.Y <= 1000000 &&
		!(b.X == -32000 && b.Y == -32000)
}
func (w *Workbench) queueWindowBounds(b workbench.Bounds) {
	if w.isManual() || !validNativeBounds(b) {
		return
	}
	capture := &nativeBoundsCapture{bounds: b, epoch: w.geometryEpoch.Load(), sequence: w.boundsSequence.Add(1)}
	// Concurrent producers cannot replace a newer pending observation. The
	// sequence also rejects timer callbacks that reach the owner out of order.
	for {
		if capture.sequence != w.boundsSequence.Load() {
			return
		}
		pending := w.pendingBounds.Load()
		if pending != nil && pending.sequence >= capture.sequence {
			return
		}
		if w.pendingBounds.CompareAndSwap(pending, capture) {
			break
		}
	}
	if w.boundsQueued.CompareAndSwap(false, true) {
		time.AfterFunc(nativeBoundsInterval, w.flushWindowBounds)
	}
}
func (w *Workbench) flushWindowBounds() {
	// Clear the scheduling flag before taking the snapshot: a concurrent new
	// capture either joins this flush or schedules its own, never gets stranded.
	w.boundsQueued.Store(false)
	w.postWindowBounds(w.pendingBounds.Swap(nil))
}
func (w *Workbench) postWindowBounds(capture *nativeBoundsCapture) {
	if capture == nil {
		return
	}
	lifetime, err := scopedWindowContext(w.Owner, w.windowRef)
	if err != nil || lifetime.Err() != nil {
		return
	}
	err = w.Owner.PostWindow(w.windowRef, func() {
		if capture.sequence != w.boundsSequence.Load() || capture.epoch != w.geometryEpoch.Load() || w.Group == nil || w.Group.windows[w.SessionWindow] != w {
			return
		}
		if err := w.captureWindowBounds(capture.bounds); err != nil && !w.reportError(err) {
			w.pendingError.CompareAndSwap(nil, &err)
		}
		w.Owner.Invalidate(w.Window)
	})
	if err != nil && !errors.Is(err, app.ErrClosed) && !errors.Is(err, app.ErrWindowClosed) {
		w.pendingError.CompareAndSwap(nil, &err)
		w.Owner.Invalidate(w.Window)
	}
}
func (w *Workbench) captureWindowBounds(b workbench.Bounds) error {
	if !validNativeBounds(b) {
		return nil
	}
	current, ok := w.Group.Session.Window(w.SessionWindow)
	if !ok {
		return nil
	}
	if current.Bounds != b {
		if err := w.Group.captureNativeBounds(w.SessionWindow, b); err != nil {
			return err
		}
	}
	// Native-originated values are observations, not requests to reapply.
	// A later native event may already be visible while its capture is pending.
	w.appliedBounds, w.boundsApplied = b, true
	return nil
}
func (w *Workbench) syncWindowGeometry() {
	if w.isManual() || w.Group == nil || !w.positioned {
		return
	}
	native := w.host.Scene.RenderWindow()
	if native == nil {
		return
	}
	win := native.SystemWindow
	if win.Title() != w.title {
		w.host.Scene.Body.Title = w.title
		if w.host.Scene.Stage != nil {
			w.host.Scene.Stage.Title = w.title
		}
		win.SetTitle(w.title)
	}
	w.syncWindowBounds(win)
}

// Apply a model request once. Native move/resize may already be visible while
// its owner capture is queued; an unrelated refresh must not push old bounds.
func (w *Workbench) syncWindowBounds(win system.Window) {
	b := w.bounds
	if w.boundsApplied && w.appliedBounds == b {
		return
	}
	w.appliedBounds, w.boundsApplied = b, true
	if size := win.WinSize(); size != image.Pt(b.Width, b.Height) {
		win.SetWinSize(image.Pt(b.Width, b.Height))
	}
	if b.Positioned {
		if pos := win.Position(nil); pos != image.Pt(b.X, b.Y) {
			win.SetPos(image.Pt(b.X, b.Y), nil)
			win.ConstrainFrame(true)
		}
	}
}

type dockHit struct {
	group  string
	index  int
	axis   workbench.Axis
	before bool
}

// hitDock uses only the current native window's geometry under its render lock.
// Tab insertion positions are final indices after removal of the source panel.
func (w *Workbench) hitDock(panel string, pos image.Point) (dockHit, bool) {
	if w.layoutSnapshot == nil {
		return dockHit{}, false
	}
	var hit func(*workbench.Group) (dockHit, bool)
	hit = func(g *workbench.Group) (dockHit, bool) {
		if g.Axis != "" {
			if h, ok := hit(g.First); ok {
				return h, true
			}
			return hit(g.Second)
		}
		tabs := w.tabs[g.ID]
		if tabs == nil || !pos.In(tabs.Geom.TotalBBox) {
			return dockHit{}, false
		}
		oldIndex := -1
		for i, id := range g.Tabs {
			if id == panel {
				oldIndex = i
			}
		}
		for i, id := range g.Tabs {
			button := w.buttons[id]
			if button == nil {
				continue
			}
			box := button.Geom.TotalBBox
			if pos.In(box) {
				index := i
				if pos.X >= (box.Min.X+box.Max.X)/2 {
					index++
				}
				if oldIndex >= 0 && oldIndex < index {
					index--
				}
				if index < 0 {
					index = 0
				}
				return dockHit{group: g.ID, index: index}, true
			}
		}
		box := tabs.Geom.TotalBBox
		band := min(64, box.Dx()/4, box.Dy()/4)
		h := dockHit{group: g.ID, index: len(g.Tabs)}
		if oldIndex >= 0 {
			h.index--
		}
		switch {
		case pos.X < box.Min.X+band:
			h.axis = workbench.Horizontal
			h.before = true
		case pos.X >= box.Max.X-band:
			h.axis = workbench.Horizontal
		case pos.Y < box.Min.Y+band:
			h.axis = workbench.Vertical
			h.before = true
		case pos.Y >= box.Max.Y-band:
			h.axis = workbench.Vertical
		}
		return h, true
	}
	return hit(w.layoutSnapshot)
}
func hasGroup(root *workbench.Group, id string) bool {
	if root == nil {
		return false
	}
	return root.ID == id || hasGroup(root.First, id) || hasGroup(root.Second, id)
}
func (w *Workbench) drop(panel string, hit dockHit) error {
	if hit.axis != "" {
		var id string
		for {
			w.generationID++
			id = fmt.Sprintf("dock-%s-%d", w.Window, w.generationID)
			exists := false
			roots := []*workbench.Group{w.layoutSnapshot}
			if w.Group != nil {
				roots = nil
				for _, win := range w.Group.Session.Layout().Windows {
					roots = append(roots, win.Root)
				}
			}
			for _, root := range roots {
				exists = exists || hasGroup(root, id) || hasGroup(root, id+".tab")
			}
			if !exists {
				break
			}
		}
		if w.Group != nil {
			return w.Group.Apply(func(s *workbench.Session) error {
				return s.Split(panel, w.SessionWindow, hit.group, id, hit.axis, hit.before)
			})
		}
		return w.Workspace.Split(panel, hit.group, id, hit.axis, hit.before)
	}
	if w.Group != nil {
		return w.Group.Apply(func(s *workbench.Session) error { return s.Dock(panel, w.SessionWindow, hit.group, hit.index) })
	}
	return w.Workspace.Move(panel, hit.group, hit.index)
}
func (w *Workbench) bindTabPointer(button *core.Tab, panel, group string) {
	button.Styler(func(s *styles.Style) { s.SetAbilities(true, abilities.Slideable) })
	button.OnFirst(events.SlideStart, func(e events.Event) { e.SetHandled() })
	button.OnFirst(events.SlideMove, func(e events.Event) { e.SetHandled() })
	button.OnFirst(events.SlideStop, func(e events.Event) {
		e.SetHandled()
		if hit, ok := w.hitDock(panel, e.WindowPos()); ok {
			w.queueInput(func() error { return w.drop(panel, hit) })
			return
		}
		if w.Group != nil {
			// A missing dock target inside the source scene is a cancelled
			// drag (for example over a splitter or padding). SceneGeom uses
			// render-window coordinates, just like WindowPos. An unrendered
			// scene cannot establish that the release was outside its bounds.
			sceneBounds := w.host.Scene.SceneGeom.Bounds()
			if sceneBounds.Empty() || e.WindowPos().In(sceneBounds) {
				return
			}
			bounds := w.floatingBounds(e.WindowPos())
			w.queueInput(func() error { return w.Group.Float(panel, bounds) })
		}
	})
	if w.Group == nil {
		return
	}
	// Menu targets are captured from a model snapshot refreshed by Sync on this
	// native owner. Building a menu must not read another window's native tree.
	button.AddContextMenu(func(menu *core.Scene, _ image.Point) {
		core.NewButton(menu).SetText("Float panel").OnClick(func(events.Event) {
			w.queueInput(func() error { return w.Group.Float(panel, workbench.Bounds{Width: 640, Height: 480}) })
		})
		for _, window := range w.menuWindows {
			var visit func(*workbench.Group)
			visit = func(target *workbench.Group) {
				if target.Axis != "" {
					visit(target.First)
					visit(target.Second)
					return
				}
				if window.ID == w.SessionWindow && target.ID == group {
					return
				}
				windowID, groupID, index := window.ID, target.ID, len(target.Tabs)
				label := "Dock in " + window.Title + " / " + groupID
				core.NewButton(menu).SetText(html.EscapeString(label)).OnClick(func(events.Event) {
					w.queueInput(func() error {
						return w.Group.Apply(func(s *workbench.Session) error { return s.Dock(panel, windowID, groupID, index) })
					})
				})
			}
			visit(window.Root)
		}
	})
}
func (w *Workbench) floatingBounds(pos image.Point) workbench.Bounds {
	b := workbench.Bounds{Width: 640, Height: 480}
	if w.isManual() {
		return b
	}
	if native := w.host.Scene.RenderWindow(); native != nil {
		win := native.SystemWindow
		origin, size, pixels := win.Position(nil), win.WinSize(), win.Size()
		if pixels.X > 0 && pixels.Y > 0 {
			b.X = origin.X + pos.X*size.X/pixels.X
			b.Y = origin.Y + pos.Y*size.Y/pixels.Y
			b.Positioned = true
		}
	}
	return b
}
