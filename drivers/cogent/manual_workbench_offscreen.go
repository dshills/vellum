//go:build offscreen

package cogent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

// prepareManualView checks the whole hosted descriptor graph before mounting.
// Panel factories run on the owner once per refresh; the captured descriptors
// are shared by Workbench's initial build and subsequent reconciliation.
func (p *manualImplementation) prepareManualView(window string, root ui.Node) (func(), error) {
	count := 0
	views := map[*Workbench]map[string]ui.Node{}
	models := map[any]bool{}
	known := map[*Workbench]bool{}
	var existing func(*record)
	existing = func(r *record) {
		if r == nil {
			return
		}
		if w, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok && w != nil {
			known[w] = true
		}
		retainedWorkbenchRecords(r, func(_ string, child *record) { existing(child) })
		for _, child := range r.children {
			existing(child)
		}
	}
	if current := p.find(window); current != nil {
		existing(current.state.root)
	}
	var walk func(ui.Node, int) error
	walk = func(n ui.Node, depth int) error {
		if depth >= 128 {
			return fmt.Errorf("%w: workbench hierarchy exceeds 128 levels", ErrManualConfig)
		}
		count++
		if count > p.config.MaxControls {
			return fmt.Errorf("%w: descriptor control limit", ErrManualConfig)
		}
		if n.FileDrop != nil {
			return fmt.Errorf("%w: native OS drop services", ErrManualUnsupported)
		}
		for _, prefix := range []string{"workbench:", "dock:", "dock-option:"} {
			if strings.HasPrefix(n.ID, prefix) {
				return fmt.Errorf("%w: reserved manual workbench target namespace", ErrManualConfig)
			}
		}
		claim := func(model any) error {
			if models[model] {
				return fmt.Errorf("%w: shared managed model", ErrManualConfig)
			}
			models[model] = true
			return nil
		}
		if n.Grid != nil {
			if err := claim(n.Grid); err != nil {
				return err
			}
		}
		if n.Tree != nil {
			if err := claim(n.Tree); err != nil {
				return err
			}
		}
		if n.Editor != nil {
			if err := claim(n.Editor); err != nil {
				return err
			}
		}

		if n.Kind == ui.Extension {
			w, ok := n.Custom.(*Workbench)
			if !ok || w == nil {
				return fmt.Errorf("%w: custom extension is not a qualified workbench", ErrManualUnsupported)
			}
			if w.Owner != p.owner || w.Window != window || w.Group == nil && w.Workspace == nil {
				return fmt.Errorf("%w: workbench owner/window/model", ErrManualConfig)
			}
			if _, duplicate := views[w]; duplicate || w.manualViews != nil || w.isManual() && w.manualMode != p.workbenchMode {
				return fmt.Errorf("%w: shared or recursive workbench", ErrManualConfig)
			}
			if w.host != nil && w.host.This != nil && !known[w] {
				return fmt.Errorf("%w: already mounted workbench", ErrManualConfig)
			}
			if w.Group != nil && (w.Group.Owner != p.owner || w.Group.PersistencePath != "") {
				return fmt.Errorf("%w: workbench ownership or filesystem autosave", ErrManualUnsupported)
			}
			panels := slices.Clone(w.Panels)
			captured := map[string]ui.Node{}
			views[w] = captured // recursion fails before evaluating another factory
			var layout *workbench.Group
			if w.Group != nil {
				if w.Group.Session == nil {
					return fmt.Errorf("%w: missing workbench session", ErrManualConfig)
				}
				win, ok := w.Group.Session.Window(w.SessionWindow)
				if !ok || w.Group.windows[w.SessionWindow] != w {
					return fmt.Errorf("%w: workbench session window", ErrManualConfig)
				}
				layout = win.Root
				current := p.find(window)
				if current == nil || win.Bounds.Positioned || win.Bounds.Width != int(current.state.window.Width) || win.Bounds.Height != int(current.state.window.Height) {
					return fmt.Errorf("%w: fixed manual window geometry", ErrManualUnsupported)
				}
			} else {
				layout = w.Workspace.Layout().Root
			}
			var visit func(*workbench.Group) error
			visit = func(g *workbench.Group) error {
				if g == nil {
					return ErrManualConfig
				}
				if g.Axis != "" {
					if err := visit(g.First); err != nil {
						return err
					}
					return visit(g.Second)
				}
				for _, id := range g.Tabs {
					count++ // retained native header, independent of panel content
					if count > p.config.MaxControls {
						return ErrManualConfig
					}
					var panel *Panel
					for i := range panels {
						if panels[i].ID == id {
							panel = &panels[i]
							break
						}
					}
					if panel == nil || panel.View == nil {
						return fmt.Errorf("%w: workbench panel %q", ErrManualConfig, id)
					}
					view := panel.View()
					if err := walk(view, depth+1); err != nil {
						return fmt.Errorf("panel %q: %w", id, err)
					}
					if err := ui.Validate(view); err != nil {
						return err
					}
					captured[id] = view
				}
				return nil
			}
			if err := visit(layout); err != nil {
				return err
			}
		}
		for _, child := range n.Children {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, 0); err != nil {
		return nil, err
	}
	if err := ui.Validate(root); err != nil {
		return nil, err
	}
	for w := range views {
		if w.Owner != p.owner || w.Window != window || w.Group != nil && (w.Group.Owner != p.owner || w.Group.PersistencePath != "") {
			return nil, fmt.Errorf("%w: workbench changed during capture", ErrManualConfig)
		}
	}
	for w, captured := range views {
		w.manualViews = captured
		w.manualMode = p.workbenchMode
		if w.Group != nil {
			w.Group.manualMode = p.workbenchMode
		}
	}
	return func() {
		for w := range views {
			w.manualViews = nil
		}
	}, nil
}

// Inspections are already detached from the binding. Qualify their transport
// identities consistently with the panel's manual targets; domain keys stay raw.
func qualifyManualGrid(g *GridInspection, prefix string) {
	if prefix == "" {
		return
	}
	g.ID = prefix + g.ID
	for i := range g.Cells {
		g.Cells[i].ID = prefix + g.Cells[i].ID
	}
	for i := range g.Headers {
		g.Headers[i].ID = prefix + g.Headers[i].ID
	}
	if g.Editor != nil {
		g.Editor.ID = prefix + g.Editor.ID
	}
}
func qualifyManualTree(t *TreeInspection, prefix string) {
	if prefix == "" {
		return
	}
	t.ID = prefix + t.ID
	for i := range t.Rows {
		t.Rows[i].ID = prefix + t.Rows[i].ID
		t.Rows[i].DisclosureID = prefix + t.Rows[i].DisclosureID
	}
	for i := range t.Pages {
		t.Pages[i].ID = prefix + t.Pages[i].ID
	}
}
