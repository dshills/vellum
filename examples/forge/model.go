package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/viewport"
	"image/color"
	"slices"
)

type feature struct{ ID, Label, Object string }
type document struct {
	Objects   []viewport.Object
	Features  []feature
	Marker    int
	Selected  string
	Selection []string
}
type workbench struct {
	clipboardBusy                                     bool
	clipboardCancel                                   context.CancelFunc
	pasteKey                                          [32]byte
	pasteCount                                        int
	hideBrowser, hideInspector                        bool
	workspaceLoaded                                   bool
	preferencesBlocked                                bool
	drag                                              *handleDrag
	handleMode                                        string
	snap                                              bool
	handleValues                                      [3]string
	application                                       *app.Application
	ref                                               app.WindowRef
	commands                                          *command.Registry
	journal                                           *history.Journal
	state                                             document
	scene                                             viewport.Scene
	surface                                           *cogent.SceneView3D
	ribbon                                            ui.Ribbon
	browser                                           ui.ModelBrowser
	timeline                                          ui.FeatureTimeline
	marking                                           ui.MarkingMenu
	dimension                                         ui.QuantityField
	expanded                                          []string
	active, tab, featureID, status, draft, validation string
	files                                             *vdoc.Session[studyFile]
	prompt                                            string
	recovery                                          *recoveryStore
	recovered                                         *studyFile
	cleanClose                                        bool
	next                                              uint64
	left, right                                       float32
}

func vec(x, y, z float64) viewport.Vec3 { return viewport.Vec3{X: x, Y: y, Z: z} }
func rgb(r, g, b uint8) color.NRGBA     { return color.NRGBA{R: r, G: g, B: b, A: 255} }
func clone(s document) document {
	s.Selection = slices.Clone(s.Selection)
	s.Objects = slices.Clone(s.Objects)
	s.Features = slices.Clone(s.Features)
	return s
}
func newWorkbench() *workbench {
	j, _ := history.New(96)
	m := &workbench{journal: j, commands: command.New(), next: 8, expanded: []string{"assembly"}, active: "housing", tab: "solid", left: .19, right: .72, status: "Ready · Local study"}
	c := viewport.DefaultCamera()
	c.Target = vec(0, .45, 0)
	m.scene = viewport.Scene{Camera: c, Grid: true, ShowCube: true}
	makeObject := func(id, name string, min, max viewport.Vec3, c color.NRGBA) viewport.Object {
		return viewport.Object{ID: id, Name: name, Bounds: viewport.Box{Min: min, Max: max}, Color: c, Visible: true}
	}
	m.state.Objects = []viewport.Object{
		makeObject("base", "Machined base", vec(-2, -.18, -1.3), vec(2, .18, 1.3), rgb(154, 168, 188)),
		makeObject("rail-front", "Front guide rail", vec(-1.8, .18, .72), vec(1.8, .35, .9), rgb(103, 119, 143)),
		makeObject("rail-back", "Rear guide rail", vec(-1.8, .18, -.9), vec(1.8, .35, -.72), rgb(103, 119, 143)),
		makeObject("housing", "Motor housing", vec(-1.75, .18, -.65), vec(-.55, 1.3, .65), rgb(53, 111, 220)),
		makeObject("carriage", "Sliding carriage", vec(-.2, .35, -.62), vec(1.05, .9, .62), rgb(243, 142, 68)),
		makeObject("cap", "End support", vec(1.45, .18, -.65), vec(1.8, .8, .65), rgb(76, 172, 164)),
		makeObject("shaft", "Drive shaft", vec(-.55, .45, -.1), vec(1.5, .65, .1), rgb(194, 202, 217)),
	}
	m.state.Objects[6].Shape = viewport.CylinderXShape
	for _, o := range m.state.Objects {
		m.state.Features = append(m.state.Features, feature{ID: "create-" + o.ID, Label: o.Name, Object: o.ID})
	}
	m.state.Marker = len(m.state.Features)
	m.state.Selected = "carriage"
	m.featureID = "create-carriage"
	m.surface = &cogent.SceneView3D{Interaction: m.handleInput, Model: &m.scene, Tool: cogent.SceneOrbit, Select: m.selectObject, SelectModified: m.selectModified, CameraChanged: func(viewport.Camera) { m.changed() }, Context: func(cogent.GPUEvent) { m.marking.Open = true; m.changed() }, OnError: func(e error) { m.status = e.Error(); fmt.Println("Forge GPU:", e) }}
	m.register("add", "New body", "Meta+N", func() error {
		if len(m.state.Objects) >= 48 {
			return fmt.Errorf("demo body limit reached")
		}
		return nil
	}, m.add)
	m.register("delete", "Delete selected body", "Backspace", func() error {
		if m.selectedIndex() < 0 {
			return fmt.Errorf("select a body")
		}
		return nil
	}, m.deleteSelected)
	m.register("undo", "Undo", "Meta+Z", func() error {
		if j.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, j.Undo)
	m.register("redo", "Redo", "Meta+Shift+Z", func() error {
		if j.Snapshot().RedoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, j.Redo)
	m.bindClipboardCommands()
	m.refresh()
	return m
}
func (m *workbench) register(id, label, key string, available func() error, run func() error) {
	if e := m.commands.Register(command.Command{ID: "forge." + id, Label: label, Shortcut: key, Available: available, Execute: func(context.Context) error { e := run(); m.result(e); return e }}); e != nil {
		panic(e)
	}
}
func (m *workbench) result(e error) {
	m.syncRecovery()
	if e != nil {
		m.status = e.Error()
	}
	m.changed()
}
func (m *workbench) changed() {
	if m.application != nil {
		_ = m.application.Invalidate(m.ref.ID())
	}
}
func (m *workbench) selectedIndex() int {
	for j, o := range m.state.Objects {
		if o.ID == m.state.Selected {
			return j
		}
	}
	return -1
}
func (m *workbench) selected() viewport.Object {
	if j := m.selectedIndex(); j >= 0 {
		return m.state.Objects[j]
	}
	return viewport.Object{}
}
func (m *workbench) refresh() {
	m.scene.Objects = slices.Clone(m.state.Objects)
	active := map[string]bool{}
	for j, f := range m.state.Features {
		if f.Object != "" {
			active[f.Object] = j < m.state.Marker
		}
	}
	for j := range m.scene.Objects {
		if enabled, hasFeature := active[m.scene.Objects[j].ID]; hasFeature {
			m.scene.Objects[j].Visible = m.scene.Objects[j].Visible && enabled
		}
	}
	if m.state.Selected == "" && m.tab == "body" {
		m.tab = "solid"
	}
	m.scene.SelectedIDs = slices.Clone(m.state.Selection)
	m.scene.SelectedID = m.state.Selected
	m.scene.HoveredID = ""
	m.addHandles()
	_ = m.scene.Touch()
	o := m.selected()
	if o.ID != "" {
		m.draft = fmt.Sprintf("%.2f mm", (o.Bounds.Max.X-o.Bounds.Min.X)*20)
	} else {
		m.draft = ""
	}
	m.validation = ""
	m.changed()
}
func (m *workbench) selectObject(id string) {
	m.drag = nil
	if id != "" {
		found := false
		for _, o := range m.state.Objects {
			found = found || o.ID == id
		}
		if !found {
			return
		}
	}
	m.state.Selection = nil
	m.state.Selected = id
	m.refresh()
}
func (m *workbench) apply(label string, after document) error {
	if m.clipboardBusy || m.drag != nil {
		return fmt.Errorf("finish or cancel active operation first")
	}
	before := clone(m.state)
	after = clone(after)
	if len(after.Objects) > 48 || len(after.Features) > 96 || after.Marker < 0 || after.Marker > len(after.Features) {
		return fmt.Errorf("demo document limit")
	}
	probe := m.scene
	probe.Objects = after.Objects
	probe.SelectedIDs = slices.Clone(after.Selection)
	probe.SelectedID = after.Selected
	probe.HoveredID = ""
	if e := probe.Validate(); e != nil {
		return e
	}
	err := m.journal.Apply(history.Change{Label: label, Apply: func() error { m.state = clone(after); m.status = label; m.refresh(); return nil }, Revert: func() error { m.state = clone(before); m.status = "Undid " + label; m.refresh(); return nil }})
	m.syncRecovery()
	return err
}
func (m *workbench) add() error {
	if len(m.state.Objects) >= 48 || len(m.state.Features) >= 96 {
		return fmt.Errorf("demo limit reached")
	}
	id := fmt.Sprintf("body-%d", m.next)
	s := clone(m.state)
	x := float64(m.next%4)*.55 - .5
	s.Objects = append(s.Objects, viewport.Object{ID: id, Name: fmt.Sprintf("Body %d", m.next), Bounds: viewport.Box{Min: vec(x, 1.35, -.35), Max: vec(x+.6, 1.95, .35)}, Color: rgb(158, 112, 221), Visible: true})
	s.Features = append(s.Features, feature{ID: "create-" + id, Label: "New body", Object: id})
	s.Marker = len(s.Features)
	s.Selected = id
	s.Selection = nil
	m.next++
	if e := m.apply("Create body", s); e != nil {
		m.next--
		return e
	}
	return nil
}
func (m *workbench) deleteSelected() error {
	j := m.selectedIndex()
	if j < 0 {
		return fmt.Errorf("select a body")
	}
	s := clone(m.state)
	selected := m.selectionSet()
	s.Objects = slices.DeleteFunc(s.Objects, func(o viewport.Object) bool { return selected[o.ID] })
	features := s.Features[:0]
	marker := 0
	for index, f := range s.Features {
		if !selected[f.Object] {
			features = append(features, f)
			if index < s.Marker {
				marker++
			}
		}
	}
	s.Features = features
	s.Marker = marker
	s.Selected = ""
	s.Selection = nil
	return m.apply("Delete body", s)
}
func (m *workbench) resize(value float64) error {
	j := m.selectedIndex()
	if j < 0 {
		return fmt.Errorf("select a body")
	}
	if value < 2 || value > 240 {
		return fmt.Errorf("width outside 2–240 mm")
	}
	s := clone(m.state)
	o := &s.Objects[j]
	center := (o.Bounds.Min.X + o.Bounds.Max.X) / 2
	o.Bounds.Min.X = center - value/40
	o.Bounds.Max.X = center + value/40
	return m.apply("Set body width", s)
}
func (m *workbench) nudge() error {
	j := m.selectedIndex()
	if j < 0 {
		return fmt.Errorf("select a body")
	}
	s := clone(m.state)
	for i := range s.Objects {
		if m.selectionSet()[s.Objects[i].ID] {
			s.Objects[i].Bounds.Min.X += .2
			s.Objects[i].Bounds.Max.X += .2
		}
	}
	return m.apply("Move body +4 mm", s)
}
func (m *workbench) visibility(id string, value bool) error {
	s := clone(m.state)
	ids := map[string]bool{id: true}
	if m.selectionSet()[id] {
		ids = m.selectionSet()
	}
	found := false
	for i := range s.Objects {
		if ids[s.Objects[i].ID] {
			s.Objects[i].Visible = value
			found = true
		}
	}
	if !found {
		return fmt.Errorf("missing body")
	}
	return m.apply("Change body visibility", s)
}

func (m *workbench) featureChange(e ui.FeatureTimelineEvent) {
	if e.Action == "select" {
		m.featureID = e.ItemID
		for _, f := range m.state.Features {
			if f.ID == e.ItemID {
				m.selectObject(f.Object)
				return
			}
		}
		m.changed()
		return
	}
	s := clone(m.state)
	if e.Action == "marker" {
		s.Marker = e.Index
		m.result(m.apply("Move feature marker", s))
		return
	}
	if e.Action == "move" {
		from := -1
		for j, f := range s.Features {
			if f.ID == e.ItemID {
				from = j
			}
		}
		if from < 0 || e.Index < 0 || e.Index >= len(s.Features) {
			return
		}
		item := s.Features[from]
		s.Features = append(s.Features[:from], s.Features[from+1:]...)
		s.Features = append(s.Features, feature{})
		copy(s.Features[e.Index+1:], s.Features[e.Index:])
		s.Features[e.Index] = item
		m.result(m.apply("Reorder feature", s))
	}
}
func (m *workbench) orient(face viewport.View) {
	if e := m.scene.Camera.SetView(face); e != nil {
		m.result(e)
		return
	}
	_ = m.scene.Touch()
	m.changed()
}
func (m *workbench) fit() {
	m.scene.Camera = viewport.DefaultCamera()
	m.scene.Camera.Target = vec(0, .45, 0)
	_ = m.scene.Touch()
	m.changed()
}
func (m *workbench) action(id string) {
	m.marking.Open = false
	switch id {
	case "add":
		m.result(m.add())
	case "delete":
		m.result(m.deleteSelected())
	case "move":
		m.result(m.nudge())
	case "undo":
		m.result(m.journal.Undo())
	case "redo":
		m.result(m.journal.Redo())
	case "fit":
		m.fit()
	case "grid":
		m.scene.Grid = !m.scene.Grid
		_ = m.scene.Touch()
	case "projection":
		m.scene.Camera.Orthographic = !m.scene.Camera.Orthographic
		_ = m.scene.Touch()
	case "orbit":
		m.surface.Tool = cogent.SceneOrbit
	case "pan":
		m.surface.Tool = cogent.ScenePan
	case "select":
		m.surface.Tool = cogent.SceneSelect
	case "front":
		m.orient(viewport.Front)
	case "top":
		m.orient(viewport.Top)
	case "right":
		m.orient(viewport.Right)
	case "marking":
		m.marking.Open = true
	}
	m.changed()
}

func (m *workbench) selectionSet() map[string]bool {
	ids := map[string]bool{}
	if m.state.Selected != "" {
		ids[m.state.Selected] = true
	}
	for _, id := range m.state.Selection {
		ids[id] = true
	}
	return ids
}
func (m *workbench) selectModified(id string, extend bool) {
	if !extend {
		m.selectObject(id)
		return
	}
	if id == "" {
		return
	}
	ids := m.selectionSet()
	if ids[id] {
		delete(ids, id)
	} else {
		ids[id] = true
	}
	m.state.Selection = nil
	m.state.Selected = ""
	for _, o := range m.state.Objects {
		if ids[o.ID] {
			m.state.Selection = append(m.state.Selection, o.ID)
			m.state.Selected = o.ID
		}
	}
	m.refresh()
}
