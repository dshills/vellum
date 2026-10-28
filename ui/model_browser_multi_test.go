package ui

import "testing"

func TestBrowserModifiedActivation(t *testing.T) {
	var id string
	var extend bool
	b := ModelBrowser{ID: "b", Label: "Bodies", Items: []ModelBrowserItem{{ID: "a", Label: "A"}, {ID: "c", Label: "C"}}, SelectedIDs: []string{"a", "c"}, SelectModified: func(s string, v bool) { id = s; extend = v }}
	n, e := b.View()
	if e != nil {
		t.Fatal(e)
	}
	var visit func(Node)
	visit = func(n Node) {
		if n.ID == ModelBrowserRowID("b", "c")+"/select" {
			n.ActivateModified(ActivationEvent{Shift: true})
		}
		for _, c := range n.Children {
			visit(c)
		}
	}
	visit(n)
	if id != "c" || !extend {
		t.Fatal("modifier lost")
	}
	b.SelectedIDs = []string{"missing"}
	if _, e = b.View(); e == nil {
		t.Fatal("missing selection accepted")
	}
}
