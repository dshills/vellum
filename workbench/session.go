package workbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// SessionVersion identifies a complete multiwindow layout; Layout remains v1.
const SessionVersion = 2
const maxSessionBytes = 262144

// Bounds are in OS window-manager units, independent of render pixel density.
// Positioned distinguishes a requested origin from an automatic OS placement.
type Bounds struct {
	X          int  `json:"x,omitempty"`
	Y          int  `json:"y,omitempty"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
	Positioned bool `json:"positioned,omitempty"`
}

type WindowLayout struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Bounds Bounds `json:"bounds"`
	Root   *Group `json:"root"`
}

type Focus struct {
	Window string `json:"window"`
	Panel  string `json:"panel"`
	Serial uint64 `json:"serial"`
}

type SessionLayout struct {
	Version int            `json:"version"`
	Windows []WindowLayout `json:"windows"`
	Focus   Focus          `json:"focus"`
}

// Session owns layout only. Documents and undo histories remain application
// domain data shared by every per-window panel View. Access is owner-confined.
type Session struct {
	layout SessionLayout
	panels []string
}

func NewSession(l SessionLayout, panels []string) (*Session, error) {
	for _, id := range panels {
		if !validName(id) {
			return nil, fmt.Errorf("workbench: invalid panel identity")
		}
	}
	s := &Session{panels: append([]string(nil), panels...)}
	if _, err := New(Layout{Version: Version, Root: &Group{ID: "registry", Tabs: append([]string(nil), panels...), Active: firstPanel(panels)}}, panels); err != nil {
		return nil, err
	}
	if err := s.Restore(l); err != nil {
		return nil, err
	}
	return s, nil
}
func validName(s string) bool { return s != "" && utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func firstPanel(ids []string) string {
	if len(ids) > 0 {
		return ids[0]
	}
	return ""
}
func cloneSession(l SessionLayout) SessionLayout {
	out := l
	out.Windows = append([]WindowLayout(nil), l.Windows...)
	for i := range out.Windows {
		out.Windows[i].Root = clone(out.Windows[i].Root)
	}
	return out
}
func (s *Session) Layout() SessionLayout { return cloneSession(s.layout) }
func (s *Session) Panels() []string      { return append([]string(nil), s.panels...) }
func (s *Session) Window(id string) (WindowLayout, bool) {
	for _, w := range s.layout.Windows {
		if w.ID == id {
			w.Root = clone(w.Root)
			return w, true
		}
	}
	return WindowLayout{}, false
}
func windowIndex(l *SessionLayout, id string) int {
	for i := range l.Windows {
		if l.Windows[i].ID == id {
			return i
		}
	}
	return -1
}
func groupPanels(g *Group) []string {
	if g == nil {
		return nil
	}
	if g.Axis == "" {
		return append([]string(nil), g.Tabs...)
	}
	return append(groupPanels(g.First), groupPanels(g.Second)...)
}
func (s *Session) Restore(l SessionLayout) error {
	if l.Version != SessionVersion || len(l.Windows) == 0 || len(l.Windows) > 32 {
		return fmt.Errorf("workbench: invalid session version or window count")
	}
	known := map[string]bool{}
	for _, id := range s.panels {
		known[id] = true
	}
	windows, groups, panels := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, win := range l.Windows {
		if !validName(win.ID) || windows[win.ID] || !validName(win.Title) || win.Bounds.Width < 160 || win.Bounds.Height < 96 || win.Bounds.Width > 16384 || win.Bounds.Height > 16384 || win.Bounds.X < -1000000 || win.Bounds.X > 1000000 || win.Bounds.Y < -1000000 || win.Bounds.Y > 1000000 {
			return fmt.Errorf("workbench: invalid window %q", win.ID)
		}
		windows[win.ID] = true
		// Validate structure before traversing it: reject cycles and depth limits.
		ids, err := collectPanelIDs(win.Root)
		if err != nil {
			return err
		}
		if _, err := New(Layout{Version: Version, Root: win.Root}, ids); err != nil {
			return err
		}
		var visit func(*Group) error
		visit = func(g *Group) error {
			if !validName(g.ID) {
				return fmt.Errorf("workbench: invalid group identity")
			}
			if groups[g.ID] {
				return fmt.Errorf("workbench: group identity duplicated across windows")
			}
			groups[g.ID] = true
			if len(groups) > 256 {
				return fmt.Errorf("workbench: group limit exceeded")
			}
			if g.Axis != "" {
				if err := visit(g.First); err != nil {
					return err
				}
				return visit(g.Second)
			}
			for _, id := range g.Tabs {
				if !known[id] || panels[id] {
					return fmt.Errorf("workbench: unknown or duplicate session panel %q", id)
				}
				panels[id] = true
			}
			return nil
		}
		if err := visit(win.Root); err != nil {
			return err
		}
	}
	if len(panels) != len(known) {
		return fmt.Errorf("workbench: session must contain every registered panel")
	}
	i := windowIndex(&l, l.Focus.Window)
	if i < 0 || !containsPanel(l.Windows[i].Root, l.Focus.Panel) {
		return fmt.Errorf("workbench: focused panel absent from focused window")
	}
	if !panelActive(l.Windows[i].Root, l.Focus.Panel) {
		return fmt.Errorf("workbench: focused panel must be active")
	}
	encoded, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) > maxSessionBytes {
		return fmt.Errorf("workbench: session exceeds 256 KiB")
	}
	s.layout = cloneSession(l)
	return nil
}
func collectPanelIDs(g *Group) ([]string, error) {
	seen := map[*Group]bool{}
	var walk func(*Group, int) ([]string, error)
	walk = func(g *Group, d int) ([]string, error) {
		if g == nil || seen[g] || d > 32 {
			return nil, fmt.Errorf("workbench: invalid group, cycle, or depth")
		}
		seen[g] = true
		if len(seen) > 256 {
			return nil, fmt.Errorf("workbench: group limit exceeded")
		}
		if g.Axis == "" {
			return append([]string(nil), g.Tabs...), nil
		}
		a, e := walk(g.First, d+1)
		if e != nil {
			return nil, e
		}
		b, e := walk(g.Second, d+1)
		return append(a, b...), e
	}
	return walk(g, 0)
}
func containsPanel(g *Group, panel string) bool {
	if g == nil {
		return false
	}
	for _, id := range g.Tabs {
		if id == panel {
			return true
		}
	}
	return containsPanel(g.First, panel) || containsPanel(g.Second, panel)
}
func panelActive(g *Group, panel string) bool {
	if g == nil {
		return false
	}
	return g.Active == panel || panelActive(g.First, panel) || panelActive(g.Second, panel)
}
func focusPanel(l *SessionLayout, window, panel string) {
	l.Focus.Window = window
	l.Focus.Panel = panel
	l.Focus.Serial++
}
func (s *Session) transaction(fn func(*SessionLayout) error) error {
	l := s.Layout()
	if err := fn(&l); err != nil {
		return err
	}
	return s.Restore(l)
}
func (s *Session) Activate(window, group, panel string) error {
	return s.transaction(func(l *SessionLayout) error {
		i := windowIndex(l, window)
		if i < 0 {
			return fmt.Errorf("workbench: unknown window")
		}
		g := find(l.Windows[i].Root, group)
		if g == nil || g.Axis != "" {
			return fmt.Errorf("workbench: tab group required")
		}
		for _, id := range g.Tabs {
			if id == panel {
				g.Active = panel
				focusPanel(l, window, panel)
				return nil
			}
		}
		return fmt.Errorf("workbench: panel absent")
	})
}
func (s *Session) Resize(window, group string, ratio float64) error {
	return s.transaction(func(l *SessionLayout) error {
		i := windowIndex(l, window)
		if i < 0 {
			return fmt.Errorf("workbench: unknown window")
		}
		g := find(l.Windows[i].Root, group)
		if g == nil || g.Axis == "" {
			return fmt.Errorf("workbench: split required")
		}
		g.Ratio = ratio
		return nil
	})
}
func (s *Session) SetBounds(window string, b Bounds) error {
	return s.transaction(func(l *SessionLayout) error {
		i := windowIndex(l, window)
		if i < 0 {
			return fmt.Errorf("workbench: unknown window")
		}
		l.Windows[i].Bounds = b
		return nil
	})
}

// Dock moves a panel to a final insertion index after removing it from its
// source, collapsing empty groups/windows. Empty windows never persist.
func (s *Session) Dock(panel, targetWindow, targetGroup string, index int) error {
	return s.transaction(func(l *SessionLayout) error {
		ti := windowIndex(l, targetWindow)
		if ti < 0 {
			return fmt.Errorf("workbench: unknown target window")
		}
		g := find(l.Windows[ti].Root, targetGroup)
		if g == nil || g.Axis != "" {
			return fmt.Errorf("workbench: tab group required")
		}
		if len(g.Tabs) == 1 && g.Tabs[0] == panel {
			if index != 0 {
				return fmt.Errorf("workbench: invalid tab index")
			}
			focusPanel(l, targetWindow, panel)
			return nil
		}
		found := false
		for i := range l.Windows {
			found = found || containsPanel(l.Windows[i].Root, panel)
			l.Windows[i].Root = remove(l.Windows[i].Root, panel)
		}
		if !found {
			return fmt.Errorf("workbench: unknown panel")
		}
		g = find(l.Windows[ti].Root, targetGroup)
		if g == nil || index < 0 || index > len(g.Tabs) {
			return fmt.Errorf("workbench: invalid dock target or tab index")
		}
		g.Tabs = append(g.Tabs, "")
		copy(g.Tabs[index+1:], g.Tabs[index:])
		g.Tabs[index] = panel
		g.Active = panel
		pruneWindows(l)
		focusPanel(l, targetWindow, panel)
		return nil
	})
}
func pruneWindows(l *SessionLayout) {
	kept := l.Windows[:0]
	for _, w := range l.Windows {
		if w.Root != nil {
			kept = append(kept, w)
		}
	}
	l.Windows = kept
}
func (s *Session) Split(panel, targetWindow, targetGroup, newID string, axis Axis, before bool) error {
	return s.transaction(func(l *SessionLayout) error {
		for _, w := range l.Windows {
			if find(w.Root, newID) != nil || find(w.Root, newID+".tab") != nil {
				return fmt.Errorf("workbench: duplicate group identity")
			}
		}
		if newID == "" {
			return fmt.Errorf("workbench: empty split identity")
		}
		ti := windowIndex(l, targetWindow)
		if ti < 0 {
			return fmt.Errorf("workbench: unknown target window")
		}
		found := false
		for i := range l.Windows {
			found = found || containsPanel(l.Windows[i].Root, panel)
			l.Windows[i].Root = remove(l.Windows[i].Root, panel)
		}
		if !found {
			return fmt.Errorf("workbench: unknown panel")
		}
		g := find(l.Windows[ti].Root, targetGroup)
		if g == nil || g.Axis != "" {
			return fmt.Errorf("workbench: cannot split own sole panel or non-tab target")
		}
		old := clone(g)
		added := &Group{ID: newID + ".tab", Tabs: []string{panel}, Active: panel}
		*g = Group{ID: newID, Axis: axis, Ratio: .5, First: old, Second: added}
		if before {
			g.First, g.Second = g.Second, g.First
		}
		pruneWindows(l)
		focusPanel(l, targetWindow, panel)
		return nil
	})
}
func (s *Session) Float(panel string, window WindowLayout) error {
	return s.transaction(func(l *SessionLayout) error {
		if windowIndex(l, window.ID) >= 0 {
			return fmt.Errorf("workbench: duplicate floating window")
		}
		found := false
		for i := range l.Windows {
			found = found || containsPanel(l.Windows[i].Root, panel)
			l.Windows[i].Root = remove(l.Windows[i].Root, panel)
		}
		if !found {
			return fmt.Errorf("workbench: unknown panel")
		}
		window.Root = &Group{ID: window.ID + ".tabs", Tabs: []string{panel}, Active: panel}
		l.Windows = append(l.Windows, window)
		pruneWindows(l)
		focusPanel(l, window.ID, panel)
		return nil
	})
}

// CloseWindow returns its panels to a surviving tab group, preserving their
// ordering and the previous focused panel if it belonged to the closing window.
func (s *Session) CloseWindow(window, targetWindow, targetGroup string) error {
	return s.transaction(func(l *SessionLayout) error {
		i, j := windowIndex(l, window), windowIndex(l, targetWindow)
		if i < 0 || j < 0 || i == j {
			return fmt.Errorf("workbench: closing requires distinct surviving window")
		}
		g := find(l.Windows[j].Root, targetGroup)
		if g == nil || g.Axis != "" {
			return fmt.Errorf("workbench: tab group required")
		}
		ids := groupPanels(l.Windows[i].Root)
		g.Tabs = append(g.Tabs, ids...)
		if l.Focus.Window == window {
			g.Active = l.Focus.Panel
			focusPanel(l, targetWindow, l.Focus.Panel)
		}
		l.Windows = append(l.Windows[:i], l.Windows[i+1:]...)
		return nil
	})
}
func (s *Session) Encode() ([]byte, error) { return json.MarshalIndent(s.Layout(), "", "  ") }
func (s *Session) Decode(data []byte) error {
	if len(data) > maxSessionBytes {
		return fmt.Errorf("workbench: session exceeds 256 KiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var l SessionLayout
	if err := d.Decode(&l); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("workbench: trailing content")
	}
	return s.Restore(l)
}
func (s *Session) Save(path string) error {
	data, err := s.Encode()
	if err != nil {
		return err
	}
	return saveBytes(path, data)
}
func (s *Session) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxSessionBytes+1))
	if err != nil {
		return err
	}
	return s.Decode(data)
}
