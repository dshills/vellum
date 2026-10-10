// Package workbench manages dock groups independently of any native toolkit.
package workbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

const Version = 1

type Axis string

const (
	Horizontal Axis = "horizontal"
	Vertical   Axis = "vertical"
)

type Group struct {
	ID     string   `json:"id"`
	Axis   Axis     `json:"axis,omitempty"`
	Ratio  float64  `json:"ratio,omitempty"`
	First  *Group   `json:"first,omitempty"`
	Second *Group   `json:"second,omitempty"`
	Tabs   []string `json:"tabs,omitempty"`
	Active string   `json:"active,omitempty"`
}
type Layout struct {
	Version int    `json:"version"`
	Root    *Group `json:"root"`
}
type Workspace struct {
	layout Layout
	panels map[string]bool
}

func New(l Layout, panels []string) (*Workspace, error) {
	w := &Workspace{panels: map[string]bool{}}
	for _, id := range panels {
		if id == "" || w.panels[id] {
			return nil, fmt.Errorf("workbench: duplicate or empty panel %q", id)
		}
		w.panels[id] = true
	}
	if err := w.Restore(l); err != nil {
		return nil, err
	}
	return w, nil
}
func clone(g *Group) *Group {
	if g == nil {
		return nil
	}
	out := *g
	out.Tabs = append([]string(nil), g.Tabs...)
	out.First = clone(g.First)
	out.Second = clone(g.Second)
	return &out
}
func (w *Workspace) Layout() Layout { return Layout{Version: Version, Root: clone(w.layout.Root)} }
func (w *Workspace) Restore(l Layout) error {
	if err := w.validate(l); err != nil {
		return err
	}
	w.layout = Layout{Version: Version, Root: clone(l.Root)}
	return nil
}
func (w *Workspace) validate(l Layout) error {
	if l.Version != Version || l.Root == nil {
		return fmt.Errorf("workbench: unsupported version or missing root")
	}
	groups := map[string]bool{}
	panels := map[string]bool{}
	var walk func(*Group, int) error
	walk = func(g *Group, depth int) error {
		if g == nil || depth > 32 || g.ID == "" || groups[g.ID] {
			return fmt.Errorf("workbench: invalid group, cycle, or depth")
		}
		groups[g.ID] = true
		if len(groups) > 256 {
			return fmt.Errorf("workbench: group limit exceeded")
		}
		if g.Axis != "" {
			if g.Axis != Horizontal && g.Axis != Vertical || g.First == nil || g.Second == nil || len(g.Tabs) != 0 || g.Active != "" || math.IsNaN(g.Ratio) || math.IsInf(g.Ratio, 0) || g.Ratio < 0.05 || g.Ratio > 0.95 {
				return fmt.Errorf("workbench: invalid split %q", g.ID)
			}
			if err := walk(g.First, depth+1); err != nil {
				return err
			}
			return walk(g.Second, depth+1)
		}
		if g.First != nil || g.Second != nil || len(g.Tabs) == 0 {
			return fmt.Errorf("workbench: invalid tab group %q", g.ID)
		}
		active := false
		for _, id := range g.Tabs {
			if !w.panels[id] || panels[id] {
				return fmt.Errorf("workbench: unknown or duplicate panel %q", id)
			}
			panels[id] = true
			active = active || id == g.Active
		}
		if !active {
			return fmt.Errorf("workbench: active tab absent")
		}
		return nil
	}
	if err := walk(l.Root, 0); err != nil {
		return err
	}
	if len(panels) != len(w.panels) {
		return fmt.Errorf("workbench: layout must contain every registered panel")
	}
	return nil
}
func find(g *Group, id string) *Group {
	if g == nil {
		return nil
	}
	if g.ID == id {
		return g
	}
	if n := find(g.First, id); n != nil {
		return n
	}
	return find(g.Second, id)
}
func remove(g *Group, panel string) *Group {
	if g.Axis != "" {
		g.First = remove(g.First, panel)
		g.Second = remove(g.Second, panel)
		if g.First == nil {
			return g.Second
		}
		if g.Second == nil {
			return g.First
		}
		return g
	}
	for i, id := range g.Tabs {
		if id == panel {
			g.Tabs = append(g.Tabs[:i], g.Tabs[i+1:]...)
			if len(g.Tabs) == 0 {
				return nil
			}
			if g.Active == panel {
				g.Active = g.Tabs[0]
			}
			break
		}
	}
	return g
}

// Move commits a validated copy, preserving every panel exactly once.
func (w *Workspace) Move(panel, target string, index int) error {
	if !w.panels[panel] {
		return fmt.Errorf("workbench: unknown panel")
	}
	l := w.Layout()
	g := find(l.Root, target)
	if g == nil || g.Axis != "" {
		return fmt.Errorf("workbench: target is not a tab group")
	}
	for _, id := range g.Tabs {
		if id == panel && len(g.Tabs) == 1 {
			return nil
		}
	}
	l.Root = remove(l.Root, panel)
	g = find(l.Root, target)
	if g == nil {
		return fmt.Errorf("workbench: move removed target")
	}
	if index < 0 || index > len(g.Tabs) {
		return fmt.Errorf("workbench: invalid tab index")
	}
	g.Tabs = append(g.Tabs, "")
	copy(g.Tabs[index+1:], g.Tabs[index:])
	g.Tabs[index] = panel
	g.Active = panel
	return w.Restore(l)
}
func (w *Workspace) Split(panel, target, newID string, axis Axis, before bool) error {
	if !w.panels[panel] || newID == "" || find(w.layout.Root, newID) != nil || find(w.layout.Root, newID+".tab") != nil {
		return fmt.Errorf("workbench: invalid panel or new group identity")
	}
	l := w.Layout()
	l.Root = remove(l.Root, panel)
	g := find(l.Root, target)
	if g == nil {
		return fmt.Errorf("workbench: cannot split own sole panel")
	}
	old := clone(g)
	added := &Group{ID: newID + ".tab", Tabs: []string{panel}, Active: panel}
	*g = Group{ID: newID, Axis: axis, Ratio: 0.5, First: old, Second: added}
	if before {
		g.First, g.Second = g.Second, g.First
	}
	return w.Restore(l)
}
func (w *Workspace) Activate(group, panel string) error {
	g := find(w.layout.Root, group)
	if g == nil || g.Axis != "" {
		return fmt.Errorf("workbench: tab group required")
	}
	for _, id := range g.Tabs {
		if id == panel {
			g.Active = panel
			return nil
		}
	}
	return fmt.Errorf("workbench: panel absent")
}
func (w *Workspace) Resize(group string, ratio float64) error {
	l := w.Layout()
	g := find(l.Root, group)
	if g == nil || g.Axis == "" {
		return fmt.Errorf("workbench: split required")
	}
	g.Ratio = ratio
	return w.Restore(l)
}
func (w *Workspace) Encode() ([]byte, error) { return json.MarshalIndent(w.Layout(), "", "  ") }
func (w *Workspace) Decode(data []byte) error {
	if len(data) > 65536 {
		return fmt.Errorf("workbench: layout exceeds 64 KiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var l Layout
	if err := d.Decode(&l); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("workbench: trailing content")
	}
	return w.Restore(l)
}

// Save atomically replaces a layout file. Failed writes preserve the previous file.
func (w *Workspace) Save(path string) error {
	data, err := w.Encode()
	if err != nil {
		return err
	}
	return saveBytes(path, data)
}

func saveBytes(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".vellum-layout-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func (w *Workspace) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		return err
	}
	return w.Decode(data)
}
