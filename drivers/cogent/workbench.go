package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"encoding/json"
	"example.com/vellum/app"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
	"fmt"
	"html"
	"math"
)

type Panel struct {
	ID, Title string
	View      func() ui.Node
}

// Workbench renders the validated dock model. Dock containers may rebuild while
// compatible panel content survives moves and keyed UI changes.
type Workbench struct {
	OnError   func(error)
	parents   map[string]*core.Frame
	tabs      map[string]*core.Tabs
	splits    map[string]*core.Splits
	Owner     *app.Application
	Window    string
	Workspace *workbench.Workspace
	Panels    []Panel
	signature string
	records   map[string]*record
	host      *core.Frame
	syncing   bool
	theme     ui.Theme
}

func (w *Workbench) SetTheme(theme ui.Theme) { w.theme = theme.Clone() }

func (w *Workbench) Mount(parent tree.Node) core.Widget {
	if w == nil || w.Owner == nil || w.Workspace == nil {
		return nil
	}
	w.host = core.NewFrame(parent)
	w.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
	return w.host
}
func (w *Workbench) Sync(core.Widget) (err error) {
	if w == nil || w.Owner == nil || w.Workspace == nil {
		return fmt.Errorf("workbench: owner and workspace required")
	}
	if w.host == nil {
		return fmt.Errorf("workbench: not mounted")
	}
	// A failed native extension update is a lifetime boundary, not a rollback
	// of arbitrary extension side effects. Clear every partial dock cache so a
	// later direct Sync can rebuild safely instead of retaining destroyed nodes.
	defer func() {
		if err == nil {
			return
		}
		w.host.DeleteChildren()
		w.records = nil
		w.parents = nil
		w.tabs = nil
		w.splits = nil
		w.signature = ""
	}()
	l := w.Workspace.Layout()
	signature := shape(l.Root)
	w.syncing = true
	defer func() { w.syncing = false }()
	if signature != w.signature {
		retained := w.records
		parking := core.NewFrame(w.host.AsTree().Parent)
		for _, r := range retained {
			if r != nil && !tree.IsNil(r.widget) {
				tree.MoveToParent(r.widget, parking)
			}
		}
		defer parking.Delete()
		w.host.DeleteChildren()
		w.records = map[string]*record{}
		w.parents = map[string]*core.Frame{}
		w.tabs = map[string]*core.Tabs{}
		w.splits = map[string]*core.Splits{}
		if err := w.build(w.host, l.Root, retained); err != nil {
			return err
		}
		for id, r := range retained {
			if w.records[id] == nil && r != nil && !tree.IsNil(r.widget) {
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
			return fmt.Errorf("workbench: panel %q absent", p.ID)
		}
		n := p.View()
		if err := ui.Validate(n); err != nil {
			return err
		}
		var err error
		r, err = reconcile(w.parents[p.ID], r, n, w.Owner, w.Window, w.theme)
		if err != nil {
			return err
		}
		w.records[p.ID] = r
	}
	return nil
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
			w.Owner.Do(func() {
				s := split.Splits()
				if len(s) == 2 {
					if err := w.Workspace.Resize(g.ID, float64(s[0])); err != nil && w.OnError != nil {
						w.OnError(err)
					}
					w.Owner.Invalidate(w.Window)
				}
			})
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
		frame, _ := tabs.NewTab(html.EscapeString(panel.Title))
		n := panel.View()
		if err := ui.Validate(n); err != nil {
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
	tabs.OnChange(func(events.Event) {
		if w.syncing {
			return
		}
		w.Owner.Do(func() {
			_, i := tabs.CurrentTab()
			if i >= 0 && i < len(g.Tabs) {
				if err := w.Workspace.Activate(g.ID, g.Tabs[i]); err != nil && w.OnError != nil {
					w.OnError(err)
				}
				w.Owner.Invalidate(w.Window)
			}
		})
	})
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
