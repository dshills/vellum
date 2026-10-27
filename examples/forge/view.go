package main

import (
	"bytes"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/color"
	"image/png"
)

func pngIcon(name string, c color.NRGBA) []byte {
	var b bytes.Buffer
	if e := png.Encode(&b, artwork(name, c)); e != nil {
		panic(e)
	}
	return b.Bytes()
}
func text(id, s string, size float32, weight int, tint color.NRGBA) ui.Node {
	return ui.Node{ID: id, Kind: ui.Label, Text: s, Style: ui.Style{Typography: ui.Some(ui.MacTypography(size, weight)), Foreground: ui.Some(tint), Padding: ui.Some(ui.Insets{}), GrowY: ui.Some(float32(0))}}
}
func column(id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Children: children, Style: ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(10)), GrowX: ui.Some(float32(1))}}
}
func row(id string, children ...ui.Node) ui.Node {
	n := column(id, children...)
	n.Kind = ui.Row
	n.Style.GrowY = ui.Some(float32(0))
	return n
}
func button(id, label string, run func()) ui.Node {
	return ui.Node{ID: id, Kind: ui.Button, Text: label, Activate: run, Style: ui.Style{GrowY: ui.Some(float32(0))}}
}
func must(n ui.Node, e error) ui.Node {
	if e != nil {
		panic(e)
	}
	return n
}
func (m *workbench) toolbar() []cogent.TitlebarToolbarItem {
	return []cogent.TitlebarToolbarItem{{ID: "add", Label: "New body", Kind: ui.Button, PNG: pngIcon("add", rgb(63, 128, 229)), Command: "forge.add"}, {ID: "flex", Kind: ui.ToolbarFlexibleSpace}, {ID: "undo", Label: "Undo", Kind: ui.Button, PNG: pngIcon("undo", rgb(143, 101, 207)), Command: "forge.undo"}, {ID: "redo", Label: "Redo", Kind: ui.Button, PNG: pngIcon("redo", rgb(45, 158, 155)), Command: "forge.redo"}, {ID: "fit", Label: "Fit assembly", Kind: ui.Button, PNG: pngIcon("fit", rgb(236, 158, 51)), Activate: m.fit}}
}
func (m *workbench) ribbonView() ui.Node {
	a := func(id, label, art string, c color.NRGBA) ui.RibbonAction {
		return ui.RibbonAction{ID: id, Label: label, IconImage: artwork(art, c)}
	}
	blue := rgb(58, 122, 225)
	teal := rgb(48, 166, 158)
	gold := rgb(236, 168, 52)
	purple := rgb(148, 105, 210)
	m.ribbon.ID = "ribbon"
	m.ribbon.Label = "Design tools"
	m.ribbon.SelectedTab = m.tab
	m.ribbon.Tabs = []ui.RibbonTab{{ID: "solid", Label: "Solid", Groups: []ui.RibbonGroup{{ID: "create", Label: "CREATE", Actions: []ui.RibbonAction{a("add", "New body", "add", blue)}}, {ID: "modify", Label: "MODIFY", Actions: []ui.RibbonAction{a("move", "Move +4 mm", "move", teal), a("delete", "Delete", "delete", rgb(224, 101, 97))}}, {ID: "view", Label: "DISPLAY", Actions: []ui.RibbonAction{a("grid", "Grid", "grid", gold), a("projection", "Projection", "projection", purple)}}, {ID: "navigate", Label: "NAVIGATE", Actions: []ui.RibbonAction{a("orbit", "Orbit", "orbit", blue), a("pan", "Pan", "pan", teal), a("select", "Select", "select", gold)}}}}, {ID: "inspect", Label: "Inspect", Groups: []ui.RibbonGroup{{ID: "orientation", Label: "ORIENTATION", Actions: []ui.RibbonAction{a("front", "Front", "front", blue), a("top", "Top", "top", teal), a("right", "Right", "right", purple), a("fit", "Fit assembly", "fit", gold)}}, {ID: "context", Label: "ACTIONS", Actions: []ui.RibbonAction{a("marking", "Action wheel", "wheel", purple)}}}}, {ID: "body", Label: "Body tools", Contextual: true, Disabled: m.selectedIndex() < 0, Groups: []ui.RibbonGroup{{ID: "transform", Label: "SELECTED BODY", Actions: []ui.RibbonAction{a("move", "Move +4 mm", "move", teal), a("delete", "Delete body", "delete", rgb(224, 101, 97))}}}}}
	m.ribbon.SelectTab = func(id string) { m.tab = id; m.changed() }
	m.ribbon.Action = func(e ui.RibbonEvent) { m.action(e.ActionID) }
	m.ribbon.Style = ui.Style{GrowY: ui.Some(float32(0)), Background: ui.Some(rgb(249, 250, 252))}
	return must(m.ribbon.View())
}
func (m *workbench) browserView() ui.Node {
	items := []ui.ModelBrowserItem{{ID: "assembly", Label: "Linear actuator", Visible: true, Activatable: true, IconImage: artwork("cube", rgb(117, 131, 153)), Badge: fmt.Sprint(len(m.state.Objects))}}
	for _, o := range m.state.Objects {
		items = append(items, ui.ModelBrowserItem{ID: o.ID, Parent: "assembly", Label: o.Name, Visible: o.Visible, Activatable: true, IconImage: artwork("cube", o.Color)})
	}
	m.browser = ui.ModelBrowser{ID: "browser", Label: "Assembly browser", Items: items, ExpandedIDs: m.expanded, SelectedID: m.state.Selected, ActiveID: m.active, Change: func(e ui.ModelBrowserEvent) {
		switch e.Kind {
		case ui.ModelBrowserSelect:
			if e.ItemID != "assembly" {
				m.selectObject(e.ItemID)
			}
		case ui.ModelBrowserVisibility:
			if e.ItemID == "assembly" {
				s := clone(m.state)
				for j := range s.Objects {
					s.Objects[j].Visible = e.Value
				}
				m.result(m.apply("Change assembly visibility", s))
			} else {
				m.result(m.visibility(e.ItemID, e.Value))
			}
		case ui.ModelBrowserActivate:
			m.active = e.ItemID
			m.changed()
		case ui.ModelBrowserExpand:
			if e.Value {
				m.expanded = []string{"assembly"}
			} else {
				m.expanded = nil
			}
			m.changed()
		}
	}}
	// An active body may have been deleted by a command since the previous view.
	exists := m.active == "assembly"
	for _, o := range m.state.Objects {
		exists = exists || o.ID == m.active
	}
	if !exists {
		m.active = ""
		m.browser.ActiveID = ""
	}
	return must(m.browser.View())
}
func (m *workbench) inspectorView() ui.Node {
	ink := rgb(40, 46, 58)
	muted := rgb(117, 126, 141)
	o := m.selected()
	name := "No body selected"
	if o.ID != "" {
		name = o.Name
	}
	m.dimension = ui.QuantityField{ID: "width", Label: "Width", Text: m.draft, Unit: "mm", Min: 2, Max: 240, Disabled: o.ID == "", ValidationMessage: m.validation, Variables: map[string]float64{}, Change: func(s string) { m.draft = s; m.validation = ""; m.changed() }, Invalid: func(e error) { m.validation = e.Error(); m.changed() }, Commit: func(v float64) { m.result(m.resize(v)) }}
	if o.ID != "" {
		m.dimension.Variables["width"] = (o.Bounds.Max.X - o.Bounds.Min.X) * 20
	}
	n := column("inspector", text("inspect-heading", "INSPECTOR", 10, 600, muted), text("selected-name", name, 17, 600, ink), text("inspect-type", "Solid body · Box proxy", 11, 400, muted), must(m.dimension.View()), text("unit-help", "Try 2 in, 50 mm, or width / 2", 11, 400, muted), button("nudge", "Move right · 4 mm", func() { m.result(m.nudge()) }), button("remove", "Delete body", func() { m.result(m.deleteSelected()) }), text("demo-heading", "ABOUT THIS WORKBENCH", 10, 600, muted), text("demo-copy", "Original Fusion-inspired UI", 11, 400, muted), text("demo-copy2", "Live GPU viewport · local history", 11, 400, muted), text("demo-copy3", "No CAD kernel or cloud service", 11, 400, muted))
	n.Children[5].Disabled = o.ID == ""
	n.Children[6].Disabled = o.ID == ""
	n.Style.Padding = ui.Some(ui.UniformInsets(18))
	n.Style.Background = ui.Some(rgb(250, 251, 253))
	n.Style.GrowY = ui.Some(float32(1))
	return n
}
func (m *workbench) View() ui.Node {
	ink := rgb(42, 49, 61)
	muted := rgb(116, 125, 140)
	left := column("project", text("project-label", "PROJECT BROWSER", 10, 600, muted), text("project-title", "Actuator study", 20, 600, ink), text("project-subtitle", "Design workspace  /  v1", 11, 400, muted), must((ui.Node{ID: "browser-scroll", Kind: ui.Scroll, Grow: true, Children: []ui.Node{m.browserView()}}), nil), text("browser-tip", "Select a body, toggle its eye,\nor activate a component.", 11, 400, muted))
	left.Style.Padding = ui.Some(ui.Insets{Top: 20, Left: 14, Right: 10, Bottom: 14})
	left.Style.Background = ui.Some(rgb(244, 246, 250))
	left.Style.GrowY = ui.Some(float32(1))
	viewControl := &ui.ViewCube{ID: "orientation", Label: "Camera orientation", Change: m.orient}
	title := row("canvas-title", text("canvas-name", "Linear actuator", 14, 600, ink), text("canvas-mode", "Solid • Perspective / Orthographic", 11, 400, muted))
	title.Style.Padding = ui.Some(ui.Insets{Left: 16, Right: 12, Top: 10, Bottom: 8})
	center := column("canvas-column", title, ui.Node{ID: "scene", Kind: ui.Extension, Custom: m.surface, Grow: true, Accessibility: ui.Accessibility{Name: "3D assembly viewport", Help: "Drag orbit. Shift drag pan. Scroll zoom. Home resets orientation."}}, must(viewControl.View()), text("canvas-hints", "Drag to orbit  ·  ⇧ drag to pan  ·  Scroll to zoom  ·  Click to select  ·  Right-click for actions", 11, 400, muted))
	center.Style.GrowY = ui.Some(float32(1))
	center.Style.Gap = ui.Some(float32(6))
	center.Style.Background = ui.Some(rgb(241, 243, 247))
	center.Style.Padding = ui.Some(ui.Insets{Bottom: 10})
	rightSplit := ui.Node{ID: "inspector-split", Kind: ui.SplitPane, Orientation: ui.Horizontal, Number: m.right, Grow: true, Change: func(e ui.ChangeEvent) { m.right = e.Number; m.changed() }, Children: []ui.Node{center, m.inspectorView()}}
	panes := ui.Node{ID: "browser-split", Kind: ui.SplitPane, Orientation: ui.Horizontal, Number: m.left, Grow: true, Change: func(e ui.ChangeEvent) { m.left = e.Number; m.changed() }, Children: []ui.Node{left, rightSplit}}
	features := []ui.FeatureItem{}
	validSelection := ""
	for _, f := range m.state.Features {
		features = append(features, ui.FeatureItem{ID: f.ID, Label: f.Label, IconImage: artwork("cube", rgb(104, 142, 191))})
		if f.ID == m.featureID {
			validSelection = f.ID
		}
	}
	m.timeline = ui.FeatureTimeline{ID: "timeline", Label: "FEATURE HISTORY", Items: features, Marker: m.state.Marker, SelectedID: validSelection, Change: m.featureChange, Style: ui.Style{Padding: ui.Some(ui.Insets{Top: 10, Left: 18, Right: 18, Bottom: 8}), Background: ui.Some(rgb(250, 251, 253))}}
	m.marking.ID = "action-wheel"
	m.marking.Label = "Workbench actions"

	m.marking.Choose = m.action
	m.marking.Dismiss = func() { m.marking.Open = false; m.changed() }
	m.marking.Actions = []ui.MarkingAction{{ID: "fit", Label: "Fit", IconImage: artwork("fit", rgb(236, 168, 52))}, {ID: "orbit", Label: "Orbit", IconImage: artwork("orbit", rgb(58, 122, 225))}, {ID: "pan", Label: "Pan", IconImage: artwork("pan", rgb(48, 166, 158))}, {ID: "select", Label: "Select", IconImage: artwork("select", rgb(148, 105, 210))}, {ID: "grid", Label: "Grid", IconImage: artwork("grid", rgb(236, 168, 52))}, {ID: "undo", Label: "Undo", IconImage: artwork("undo", rgb(148, 105, 210)), Disabled: m.journal.Snapshot().UndoCount == 0}, {ID: "move", Label: "Move +4 mm", IconImage: artwork("move", rgb(48, 166, 158)), Disabled: m.selectedIndex() < 0}, {ID: "add", Label: "New body", IconImage: artwork("add", rgb(58, 122, 225))}}
	footer := text("status", fmt.Sprintf("%s    ·    %d bodies    ·    Undo %d / Redo %d", m.status, len(m.state.Objects), m.journal.Snapshot().UndoCount, m.journal.Snapshot().RedoCount)+fmt.Sprintf("  ·  Camera %.2f / %.2f / %.2f  ·  Target %.2f, %.2f, %.2f", m.scene.Camera.Yaw, m.scene.Camera.Pitch, m.scene.Camera.Distance, m.scene.Camera.Target.X, m.scene.Camera.Target.Y, m.scene.Camera.Target.Z), 11, 400, muted)
	footer.Style.Padding = ui.Some(ui.Insets{Left: 18, Bottom: 8, Top: 3})
	return ui.Node{ID: "forge-root", Kind: ui.Column, Grow: true, Style: ui.Style{Gap: ui.Some(float32(0)), Padding: ui.Some(ui.Insets{}), Background: ui.Some(rgb(248, 249, 252))}, Children: []ui.Node{m.ribbonView(), panes, must(m.timeline.View()), footer, must(m.marking.View())}}
}
