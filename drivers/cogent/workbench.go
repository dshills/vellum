package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"html"
	"math"
	"sync/atomic"
)

// workbenchActiveProperty is the owner snapshot used by nested modal controls.
// Set it before reconciliation, including the initial mount of inactive panels.
const workbenchActiveProperty = "vellum.workbench.active"

type Panel struct {
	ID, Title string
	View      func() ui.Node
	// NewView creates a private per-window view for panels containing native
	// extensions. Its returned closure should share domain state, not widgets.
	NewView func(nativeWindowID string) func() ui.Node
}

// Workbench renders the validated dock model. Dock containers may rebuild while
// compatible panel content survives moves and keyed UI changes.
type Workbench struct {
	OnError        func(error)
	parents        map[string]*core.Frame
	tabs           map[string]*core.Tabs
	splits         map[string]*core.Splits
	Owner          *app.Application
	Window         string
	Workspace      *workbench.Workspace
	Group          *WorkbenchGroup
	SessionWindow  string
	buttons        map[string]*core.Tab
	pendingError   atomic.Pointer[error]
	retired        atomic.Bool
	geometryEpoch  atomic.Uint64
	boundsSequence atomic.Uint64
	pendingBounds  atomic.Pointer[nativeBoundsCapture]
	boundsQueued   atomic.Bool
	generation     atomic.Uint64
	generationID   uint64
	menuWindows    []workbench.WindowLayout
	layoutSnapshot *workbench.Group
	focusSerial    uint64
	focusPanel     string
	bounds         workbench.Bounds
	appliedBounds  workbench.Bounds
	boundsApplied  bool
	title          string
	positioned     bool
	Panels         []Panel
	signature      string
	records        map[string]*record
	host           *core.Frame
	syncing        bool
	theme          ui.Theme
	// Manual replay freezes validated panel descriptors for one reconciliation.
	// Nil retains ordinary desktop View evaluation. No captured views survive it.
	manualViews map[string]ui.Node
	manualMode  *manualWorkbenchMode
	inputScope  atomic.Pointer[workbenchInputScope]
}

// One manual adapter's lease contains no model, owner or widget references.
// Rollback/closure deactivate it only after every native scene has retired.
type manualWorkbenchMode struct{ active atomic.Bool }

func (w *Workbench) isManual() bool { return w.manualMode != nil && w.manualMode.active.Load() }

func (w *Workbench) SetTheme(theme ui.Theme) { w.theme = theme.Clone() }

func (w *Workbench) Mount(parent tree.Node) core.Widget {
	if w == nil || w.Owner == nil || w.Workspace == nil && w.Group == nil {
		return nil
	}
	w.retired.Store(false)
	w.setInputScope(true, false, nil)
	w.host = core.NewFrame(parent)
	w.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
	w.bindWindowGeometry()
	return w.host
}
func (w *Workbench) Sync(core.Widget) (err error) {
	if w == nil || w.Owner == nil || w.Workspace == nil && w.Group == nil {
		return fmt.Errorf("workbench: owner and workspace required")
	}
	if w.host == nil || w.host.This == nil {
		return fmt.Errorf("workbench: not mounted")
	}
	if pending := w.pendingError.Swap(nil); pending != nil && !w.reportError(*pending) {
		return fmt.Errorf("workbench: queued native change: %w", *pending)
	}
	// A failed native extension update is a lifetime boundary, not a rollback
	// of arbitrary extension side effects. Clear every partial dock cache so a
	// later direct Sync can rebuild safely instead of retaining destroyed nodes.
	defer func() {
		if err == nil {
			return
		}
		w.disposeControlRecords()
		w.host.DeleteChildren()
		w.records = nil
		w.parents = nil
		w.tabs = nil
		w.splits = nil
		w.buttons = nil
		w.layoutSnapshot = nil
		w.signature = ""
	}()
	l, err := w.layout()
	if err != nil {
		return err
	}
	w.retired.Store(false) // explicit retry rebuilds a cleared but still attached host
	w.layoutSnapshot = l.Root
	signature := shape(l.Root)
	w.syncing = true
	defer func() { w.syncing = false }()
	if signature != w.signature {
		retained := w.records
		parking := core.NewFrame(w.host.AsTree().Parent)
		for _, r := range retained {
			if r != nil && !tree.IsNil(r.widget) {
				withdrawDataGeometry(r)
				tree.MoveToParent(r.widget, parking)
			}
		}
		defer func() {
			// A failed build can leave old panels unadopted in parking.
			// Retire their private bindings before deleting their widgets.
			for id, r := range retained {
				if w.records[id] != r {
					disposeControlTree(r)
				}
			}
			parking.Delete()
		}()
		w.host.DeleteChildren()
		w.generation.Add(1)
		w.records = map[string]*record{}
		w.parents = map[string]*core.Frame{}
		w.tabs = map[string]*core.Tabs{}
		w.buttons = map[string]*core.Tab{}
		w.splits = map[string]*core.Splits{}
		if err := w.build(w.host, l.Root, retained); err != nil {
			return err
		}
		for id, r := range retained {
			if w.records[id] == nil && r != nil && !tree.IsNil(r.widget) {
				disposeControlTree(r)
				r.widget.AsTree().Delete()
			}
		}
		w.signature = signature
		w.host.Update()
	}
	w.syncLayout(l.Root)
	for _, p := range w.Panels {
		r := w.records[p.ID]
		if r == nil {
			continue
		}
		n, err := w.panelView(p)
		if err != nil {
			return err
		}
		r, err = reconcile(w.parents[p.ID], r, n, w.Owner, w.Window, w.theme)
		if err != nil {
			return err
		}
		w.records[p.ID] = r
	}
	w.syncWindowGeometry()
	w.syncFocus()
	return nil
}

func (w *Workbench) panelView(p Panel) (ui.Node, error) {
	var n ui.Node
	if w.manualViews != nil {
		var ok bool
		n, ok = w.manualViews[p.ID]
		if !ok {
			return n, fmt.Errorf("workbench: panel %q missing checked manual view", p.ID)
		}
	} else {
		if p.View == nil {
			return n, fmt.Errorf("workbench: panel %q has no view", p.ID)
		}
		n = p.View()
	}
	return n, ui.Validate(n)
}
func shape(g *workbench.Group) string {
	c := *g
	c.Ratio = 0
	c.Active = ""
	if c.First != nil {
		a := shape(c.First)
		b := shape(c.Second)
		c.First = nil
		c.Second = nil
		data, _ := json.Marshal(c)
		return string(data) + a + b
	}
	data, _ := json.Marshal(c)
	return string(data)
}
func (w *Workbench) build(parent tree.Node, g *workbench.Group, retained map[string]*record) error {
	if g.Axis != "" {
		split := core.NewSplits(parent).SetSplits(float32(g.Ratio), float32(1-g.Ratio))
		split.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
			if g.Axis == workbench.Vertical {
				s.Direction = styles.Column
			}
		})
		for _, child := range []*workbench.Group{g.First, g.Second} {
			frame := core.NewFrame(split)
			frame.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
			if err := w.build(frame, child, retained); err != nil {
				return err
			}
		}
		w.splits[g.ID] = split
		split.OnChange(func(events.Event) {
			if w.syncing {
				return
			}
			s := split.Splits()
			if len(s) == 2 {
				ratio := math.Max(.05, math.Min(.95, float64(s[0])))
				w.queueInput(func() error {
					if w.Group != nil {
						return w.Group.Apply(func(model *workbench.Session) error { return model.Resize(w.SessionWindow, g.ID, ratio) })
					}
					return w.Workspace.Resize(g.ID, ratio)
				})
			}
		})
		return nil
	}
	tabs := core.NewTabs(parent).SetType(core.FunctionalTabs).SetCloseIcon("")
	tabs.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
	active := 0
	for i, id := range g.Tabs {
		var panel *Panel
		for j := range w.Panels {
			if w.Panels[j].ID == id {
				panel = &w.Panels[j]
				break
			}
		}
		if panel == nil || panel.View == nil {
			return fmt.Errorf("workbench: unregistered panel %q", id)
		}
		frame, button := tabs.NewTab(html.EscapeString(panel.Title))
		frame.SetProperty(workbenchActiveProperty, g.Active == id)
		w.buttons[id] = button
		w.bindTabPointer(button, id, g.ID)
		button.OnClick(func(events.Event) {
			if w.syncing || w.buttons[id] != button {
				return
			}
			w.queueInput(func() error {
				if w.buttons[id] != button {
					return nil
				}
				if w.Group != nil {
					return w.Group.Apply(func(model *workbench.Session) error { return model.Activate(w.SessionWindow, g.ID, id) })
				}
				return w.Workspace.Activate(g.ID, id)
			})
		})
		n, err := w.panelView(*panel)
		if err != nil {
			return err
		}
		r, err := reconcile(frame, retained[id], n, w.Owner, w.Window, w.theme)
		if err != nil {
			return err
		}
		w.records[id] = r
		w.parents[id] = frame
		if g.Active == id {
			active = i
		}
	}
	w.tabs[g.ID] = tabs
	tabs.SelectTabIndex(active)
	return nil
}

func (w *Workbench) syncLayout(g *workbench.Group) {
	if g.Axis != "" {
		if s := w.splits[g.ID]; s != nil {
			current := s.Splits()
			if len(current) == 2 && math.Abs(float64(current[0])-g.Ratio) > .0001 {
				s.SetSplits(float32(g.Ratio), float32(1-g.Ratio)).Update()
			}
		}
		w.syncLayout(g.First)
		w.syncLayout(g.Second)
		return
	}
	for _, id := range g.Tabs {
		if frame := w.parents[id]; frame != nil {
			frame.SetProperty(workbenchActiveProperty, g.Active == id)
		}
	}
	if tabs := w.tabs[g.ID]; tabs != nil {
		for i, id := range g.Tabs {
			if id == g.Active {
				_, old := tabs.CurrentTab()
				if old != i {
					tabs.SelectTabIndex(i)
				}
				break
			}
		}
	}
}

// disposeControlRecords retires this extension's private record graph. Widget
// destruction still belongs to its originating scene; shared domain models and
// panel factories remain application-owned. It is safe to call again.
func (w *Workbench) disposeControlRecords() {
	w.retired.Store(true)
	if scope := w.inputScope.Swap(nil); scope != nil {
		scope.revoked.Store(true)
	}
	w.generation.Add(1)
	w.pendingError.Store(nil)
	for _, r := range w.records {
		disposeControlTree(r)
	}
	w.records = nil
	w.parents = nil
	w.tabs = nil
	w.splits = nil
	w.buttons = nil
	w.layoutSnapshot = nil
	w.signature = ""
}
