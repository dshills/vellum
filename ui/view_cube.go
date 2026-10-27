package ui

import (
	"fmt"
	"github.com/dshills/vellum/viewport"
)

// ViewCube exposes keyboard/accessible orientation alternatives for an in-canvas
// cube (such as cogent.SceneView3D). Rendering and pointer picking use the shared
// viewport helpers; this control routes the same face identifiers.
type ViewCube struct {
	ID, Label string
	Disabled  bool
	Change    func(viewport.View)
	Style     Style
}

func (p *ViewCube) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) {
		return Node{}, fmt.Errorf("ui: invalid view control")
	}
	id := p.ID
	root := Node{ID: id, Kind: Row, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, Style: Style{Gap: Some(float32(4)), GrowY: Some(float32(0))}.Overlay(p.Style)}
	for _, face := range []struct {
		v     viewport.View
		label string
	}{{viewport.Isometric, "Home"}, {viewport.Front, "Front"}, {viewport.Top, "Top"}, {viewport.Right, "Right"}, {viewport.Back, "Back"}, {viewport.Left, "Left"}, {viewport.Bottom, "Bottom"}} {
		v := face.v
		root.Children = append(root.Children, Node{ID: id + "/view/" + string(v), Kind: Button, Text: face.label, Disabled: p.Change == nil, Appearance: ButtonQuiet, Style: Style{Padding: Some(Insets{Left: 7, Right: 7, Top: 3, Bottom: 3}), Typography: Some(MacTypography(11, 400)), GrowX: Some(float32(0))}, Activate: func() {
			if p.ID == id && !p.Disabled && p.Change != nil {
				p.Change(v)
			}
		}})
	}
	return root, Validate(root)
}
