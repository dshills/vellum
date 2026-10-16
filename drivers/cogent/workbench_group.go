package cogent

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

var workbenchGroupIDs atomic.Uint64

// WorkbenchGroup binds a shared domain session to independent native adapters.
// Construct and mutate it only on the application's owner (Do, Drain or View).
// Panel.View closures own documents; native records never cross window contexts.
type WorkbenchGroup struct {
	Owner         *app.Application
	Session       *workbench.Session
	Panels        []Panel
	OnError       func(error)
	OnWindowReady func(logicalWindowID string)
	// PersistencePath enables atomic autosave. A failed save rejects the model
	// change and compensates native registrations before returning the error.
	PersistencePath string
	windows         map[string]*Workbench
	next            uint64
	identity        uint64
	open            func(app.Window) (app.WindowRef, error)
	close           func(app.WindowRef) error
	manualMode      *manualWorkbenchMode
}

func NewWorkbenchGroup(owner *app.Application, session *workbench.Session, panels []Panel) (*WorkbenchGroup, error) {
	if owner == nil || session == nil {
		return nil, fmt.Errorf("workbench: owner and session required")
	}
	ids := map[string]bool{}
	for _, id := range session.Panels() {
		ids[id] = true
	}
	for _, p := range panels {
		if !ids[p.ID] || p.Title == "" || p.View == nil && p.NewView == nil {
			return nil, fmt.Errorf("workbench: invalid or duplicate panel %q", p.ID)
		}
		delete(ids, p.ID)
	}
	if len(ids) > 0 {
		return nil, fmt.Errorf("workbench: missing registered panels")
	}
	return &WorkbenchGroup{Owner: owner, Session: session, Panels: append([]Panel(nil), panels...), windows: map[string]*Workbench{}, identity: workbenchGroupIDs.Add(1), open: owner.OpenWindow, close: owner.CloseWindow}, nil
}
func (g *WorkbenchGroup) newWindow(spec workbench.WindowLayout) (*Workbench, app.Window) {
	g.next++
	nativeID := "vellum.workbench." + strconv.FormatUint(g.identity, 10) + "." + spec.ID + "." + strconv.FormatUint(g.next, 10)
	panels := append([]Panel(nil), g.Panels...)
	for i := range panels {
		if factory := panels[i].NewView; factory != nil {
			var view func() ui.Node
			panels[i].View = func() ui.Node {
				if view == nil {
					view = factory(nativeID)
				}
				if view == nil {
					return ui.Node{}
				}
				return view()
			}
		}
	}
	w := &Workbench{Owner: g.Owner, Window: nativeID, Group: g, SessionWindow: spec.ID, Panels: panels}
	return w, app.Window{ID: nativeID, Title: spec.Title, Width: float32(spec.Bounds.Width), Height: float32(spec.Bounds.Height), View: func() ui.Node { return ui.Node{ID: "workbench", Kind: ui.Extension, Custom: w, Grow: true} }}
}

// WindowID resolves a logical session identity to its current lifetime ID.
// Resolve on the owner each time rather than retaining IDs across rollback.
func (g *WorkbenchGroup) WindowID(logical string) (string, bool) {
	w, ok := g.windows[logical]
	if !ok {
		return "", false
	}
	return w.Window, true
}

// Open registers every session window. Call before Run or on the owner during
// Run. Window IDs are never reused, including compensating rollback windows.
func (g *WorkbenchGroup) Open() error { return g.apply(g.Session.Layout(), false, false) }
func (g *WorkbenchGroup) Apply(change func(*workbench.Session) error) error {
	if change == nil {
		return fmt.Errorf("workbench: nil transaction")
	}
	draft, err := workbench.NewSession(g.Session.Layout(), g.Session.Panels())
	if err != nil {
		return err
	}
	if err = change(draft); err != nil {
		return err
	}
	return g.apply(draft.Layout(), true, false)
}
func (g *WorkbenchGroup) apply(next workbench.SessionLayout, persist, nativeGeometry bool) error {
	if g.Owner.Context().Err() != nil {
		return app.ErrClosed
	}
	if g.manualMode != nil && g.manualMode.active.Load() {
		if g.PersistencePath != "" {
			return ErrManualUnsupported
		}
		for _, after := range next.Windows {
			if after.Bounds.Positioned {
				return fmt.Errorf("%w: manual window positioning", ErrManualUnsupported)
			}
		}
		for _, before := range g.Session.Layout().Windows {
			for _, after := range next.Windows {
				if after.ID == before.ID && after.Bounds != before.Bounds {
					return fmt.Errorf("%w: manual window geometry is fixed", ErrManualUnsupported)
				}
			}
		}
	}
	previous := g.Session.Layout()
	for id, w := range g.windows {
		ctx, err := scopedWindowContext(g.Owner, w.windowRef)
		if err != nil || ctx.Err() != nil {
			return fmt.Errorf("workbench: native registration %q is no longer live", id)
		}
	}
	desired := map[string]workbench.WindowLayout{}
	for _, w := range next.Windows {
		desired[w.ID] = w
	}
	added, removed := map[string]*Workbench{}, map[string]workbench.WindowLayout{}
	rollback := func(cause error) error {
		var cleanup []error
		// Restore removed registrations first so rollback never closes the last
		// window of a still-live application while another must be reopened.
		for _, spec := range previous.Windows {
			if _, ok := removed[spec.ID]; ok {
				w, win := g.newWindow(spec)
				if ref, err := g.open(win); err != nil {
					cleanup = append(cleanup, err)
				} else {
					w.windowRef = ref
					g.windows[spec.ID] = w
				}
			}
		}
		for id, w := range added {
			delete(g.windows, id)
			if err := g.close(w.windowRef); err != nil {
				cleanup = append(cleanup, err)
			}
		}
		if len(cleanup) > 0 {
			g.Owner.Close()
			return errors.Join(append([]error{cause, fmt.Errorf("workbench: native rollback failed; application closed")}, cleanup...)...)
		}
		g.Owner.Invalidate("")
		return cause
	}
	for _, spec := range next.Windows {
		if g.windows[spec.ID] == nil {
			w, win := g.newWindow(spec)
			ref, err := g.open(win)
			if err != nil {
				return rollback(err)
			}
			w.windowRef = ref
			g.windows[spec.ID] = w
			added[spec.ID] = w
		}
	}
	for _, spec := range previous.Windows {
		if _, ok := desired[spec.ID]; !ok {
			if w := g.windows[spec.ID]; w != nil {
				// Retire before invoking close: a synchronous native close hook
				// must see this adapter as retired rather than reenter Apply.
				delete(g.windows, spec.ID)
				if err := g.close(w.windowRef); err != nil {
					g.windows[spec.ID] = w
					return rollback(err)
				}
				removed[spec.ID] = spec
			}
		}
	}
	draft, err := workbench.NewSession(next, g.Session.Panels())
	if err != nil {
		return rollback(err)
	}
	if persist && g.PersistencePath != "" {
		if err := draft.Save(g.PersistencePath); err != nil {
			return rollback(err)
		}
	}
	if err := g.Session.Restore(next); err != nil {
		return rollback(err)
	}
	for _, before := range previous.Windows {
		if after, ok := desired[before.ID]; ok && before.Bounds != after.Bounds && !nativeGeometry {
			if w := g.windows[before.ID]; w != nil {
				w.geometryEpoch.Add(1)
			}
		}
	}
	// Publication is the commit point. Cancellation after it can refuse the
	// wake, but must not turn an accepted transaction into a reported failure.
	g.Owner.Invalidate("")
	return nil
}
func (g *WorkbenchGroup) captureNativeBounds(window string, bounds workbench.Bounds) error {
	draft, err := workbench.NewSession(g.Session.Layout(), g.Session.Panels())
	if err != nil {
		return err
	}
	if err := draft.SetBounds(window, bounds); err != nil {
		return err
	}
	return g.apply(draft.Layout(), true, true)
}
func (g *WorkbenchGroup) Save(path string) error {
	if g.manualMode != nil && g.manualMode.active.Load() {
		return ErrManualUnsupported
	}
	return g.Session.Save(path)
}
func (g *WorkbenchGroup) Load(path string) error {
	if g.manualMode != nil && g.manualMode.active.Load() {
		return ErrManualUnsupported
	}
	serial := g.Session.Layout().Focus.Serial
	if err := g.Apply(func(s *workbench.Session) error {
		if err := s.Load(path); err != nil {
			return err
		}
		l := s.Layout()
		l.Focus.Serial = serial + 1
		return s.Restore(l)
	}); err != nil {
		return err
	}
	// Load is an explicit geometry request even when saved values equal the
	// last model values and a newer native capture is still pending.
	for _, w := range g.windows {
		w.geometryEpoch.Add(1)
		w.boundsApplied = false
	}
	g.Owner.Invalidate("")
	return nil
}
func (g *WorkbenchGroup) Float(panel string, bounds workbench.Bounds) error {
	// Logical identities also remain unique over this controller's lifetime.
	var id string
	for {
		g.next++
		id = "floating-" + strconv.FormatUint(g.next, 10)
		if _, exists := g.Session.Window(id); !exists {
			break
		}
	}
	title := panel
	for _, p := range g.Panels {
		if p.ID == panel {
			title = p.Title
			break
		}
	}
	return g.Apply(func(s *workbench.Session) error {
		return s.Float(panel, workbench.WindowLayout{ID: id, Title: title, Bounds: bounds})
	})
}
func firstTabGroup(root *workbench.Group) *workbench.Group {
	for root != nil && root.Axis != "" {
		root = root.First
	}
	return root
}
func (g *WorkbenchGroup) Close(window string) error {
	l := g.Session.Layout()
	if len(l.Windows) == 1 {
		return fmt.Errorf("workbench: cannot close final session window")
	}
	for _, target := range l.Windows {
		if target.ID != window {
			group := firstTabGroup(target.Root)
			return g.Apply(func(s *workbench.Session) error { return s.CloseWindow(window, target.ID, group.ID) })
		}
	}
	return fmt.Errorf("workbench: no surviving window")
}

// WindowCloseExtension is notified synchronously on the owner before a native
// registration closes. Driver recursively invokes it for nested extensions.
type WindowCloseExtension interface{ WindowClosed() error }

func (w *Workbench) WindowClosed() error {
	if w.Group == nil {
		return nil
	}
	if current := w.Group.windows[w.SessionWindow]; current != w {
		return nil
	}
	w.geometryEpoch.Add(1)
	w.pendingBounds.Store(nil)
	if native := w.host; !w.isManual() && native != nil && native.Scene.RenderWindow() != nil {
		win := native.Scene.RenderWindow().SystemWindow
		size, pos := win.WinSize(), win.Position(nil)
		if err := w.captureWindowBounds(workbench.Bounds{X: pos.X, Y: pos.Y, Width: size.X, Height: size.Y, Positioned: true}); err != nil {
			return err
		}
	}
	if len(w.Group.Session.Layout().Windows) == 1 {
		if path := w.Group.PersistencePath; path != "" {
			return w.Group.Save(path)
		}
		return nil
	}
	return w.Group.Close(w.SessionWindow)
}
func closeWorkbenchExtensions(r *record) error {
	if r == nil {
		return nil
	}
	var failures []error
	if extension, ok := r.node.Custom.(WindowCloseExtension); r.node.Kind == ui.Extension && ok {
		if err := extension.WindowClosed(); err != nil {
			failures = append(failures, err)
		}
	}
	for _, child := range r.children {
		if err := closeWorkbenchExtensions(child); err != nil {
			failures = append(failures, err)
		}
	}
	if w, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok && w != nil {
		for _, panel := range w.Panels {
			if err := closeWorkbenchExtensions(w.records[panel.ID]); err != nil {
				failures = append(failures, err)
			}
		}
	}
	return errors.Join(failures...)
}
