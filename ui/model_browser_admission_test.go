package ui

import "testing"

func TestModelBrowserModifiedAdmission(t *testing.T) {
	calls := 0
	p := &ModelBrowser{ID: "b", Label: "Browser", Items: []ModelBrowserItem{{ID: "one", Label: "One", Visible: true}}, Change: func(ModelBrowserEvent) {}, SelectModified: func(string, bool) { calls++ }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	var find func(Node) Node
	find = func(n Node) Node {
		if n.ID == ModelBrowserRowID("b", "one")+"/select" {
			return n
		}
		for _, c := range n.Children {
			v := find(c)
			if v.ID != "" {
				return v
			}
		}
		return Node{}
	}
	selectNode := find(n)
	if selectNode.ActivateModified == nil {
		t.Fatal("missing modified select")
	}
	p.Disabled = true
	selectNode.ActivateModified(ActivationEvent{Shift: true})
	if calls != 0 {
		t.Fatalf("disabled browser delivered %d modified selection callbacks", calls)
	}
}
