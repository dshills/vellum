package ui

import "testing"

func TestInspectorRetainedSectionsMixedValuesAndRemoval(t *testing.T) {
	calls := 0
	p := &Inspector{ID: "p", Title: "Inspector", Sections: []InspectorSection{{ID: "geometry", Title: "Geometry", Expanded: true, Properties: []InspectorProperty{{ID: "name", Label: "Name", Mixed: true, Editor: Node{ID: "name", Kind: TextField, Text: "Actual value"}}}}}}
	p.Disclosure = func(e InspectorDisclosureEvent) { calls++; p.Sections[0].Expanded = e.Expanded }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	section := n.Children[1]
	editor := section.Children[0].Children[1]
	if editor.Text != "Actual value" || len(section.Children[0].Children) != 3 {
		t.Fatal("mixed presentation replaced accepted value")
	}
	section.Change(ChangeEvent{Checked: false})
	n, err = p.View()
	if err != nil || n.Children[1].Checked || n.Children[1].Children[0].Children[1].ID != editor.ID {
		t.Fatal(n, err)
	}
	p.Sections = nil
	section.Change(ChangeEvent{Checked: true})
	if calls != 1 {
		t.Fatal("removed section acted")
	}
	p.LabelWidth = 500
	if _, err = p.View(); err == nil {
		t.Fatal("unbounded width admitted")
	}
}

func TestInspectorRejectsEditorNamespace(t *testing.T) {
	p := &Inspector{ID: "p", Title: "P", Sections: []InspectorSection{{ID: "s", Title: "Section", Properties: []InspectorProperty{{ID: "f", Label: "Field", Editor: Node{ID: "p/inspector/stolen", Kind: TextField}}}}}}
	if _, err := p.View(); err == nil {
		t.Fatal("reserved editor namespace admitted")
	}
}

func TestInspectorRejectsCyclicEditorDescription(t *testing.T) {
	children := make([]Node, 1)
	editor := Node{ID: "loop", Kind: Column, Children: children}
	children[0] = editor
	p := &Inspector{ID: "inspector", Title: "Properties", Sections: []InspectorSection{{ID: "section", Title: "Section", Properties: []InspectorProperty{{ID: "p", Label: "Property", Editor: editor}}}}}
	if _, err := p.View(); err == nil {
		t.Fatal("cyclic editor admitted")
	}
}
