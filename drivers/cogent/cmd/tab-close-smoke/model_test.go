package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestNativeFixtureRetainsRejectedDraftAndValidFinalEmptyState(t *testing.T) {
	m := newModel()
	m.draft = "uncommitted draft"
	for _, page := range []string{"b", "a"} {
		n := m.view()
		if err := ui.Validate(n); err != nil {
			t.Fatal(err)
		}
		n.Children[1].TabClose(ui.TabCloseEvent{ID: "tabs", PageID: page})
		if len(m.documents) != 2 || m.selected != "a" || m.draft != "uncommitted draft" || m.failure != nil {
			t.Fatal("fixture rejection altered accepted state")
		}
	}
	m.reject = false
	for _, page := range []string{"b", "a"} {
		n := m.view()
		n.Children[1].TabClose(ui.TabCloseEvent{ID: "tabs", PageID: page})
		if err := ui.Validate(m.view()); err != nil {
			t.Fatal("fixture final removal emitted invalid Tabs:", err)
		}
	}
	n := m.view()
	if len(n.Children) != 2 || n.Children[1].ID != "empty" || n.Children[1].Kind != ui.Label {
		t.Fatal("fixture lost final empty state")
	}
	// A stale proposal delivered past the driver guard must make this acceptance
	// fixture fail, rather than becoming an unnoticed no-op close.
	n = m.view()
	m.documents = []string{"a", "b"}
	m.selected = "a"
	n = m.view()
	n.Children[1].TabClose(ui.TabCloseEvent{ID: "tabs", PageID: "absent"})
	if m.failure == nil {
		t.Fatal("fixture silently tolerated stale request")
	}
}
