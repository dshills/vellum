package main

import (
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"testing"
)

func readyActionFixture(t *testing.T) (actionState, cogent.AccessibilityInspection) {
	t.Helper()
	m, err := editor.New("A😀e\u0301\r\n", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Replace("A😀e\u0301\r\nlast\r\n\r\n", editor.Selection{Anchor: 1, Focus: 2}); err != nil {
		t.Fatal(err)
	}
	state := actionState{Editor: m.Snapshot(), Status: "Document edited", Filename: "Untitled • Unsaved", EditorID: "document"}
	tree := cogent.AccessibilityInspection{Generation: 7, Reachable: true, NativeWindow: true}
	for id, value := range map[string]string{"document": state.Editor.VisibleText, "status": state.Status, "filename": state.Filename, "find": state.Search} {
		e := cogent.AccessibilityElement{DomainID: id, Value: value, Enabled: true}
		if id == "document" {
			e.TextSelection = [2]int{1, 2}
		}
		tree.Elements = append(tree.Elements, e)
	}
	for id, enabled := range map[string]bool{"undo": true, "redo": false, "open": true, "save": true, "save-as": true, "cancel-operation": false} {
		tree.Elements = append(tree.Elements, cogent.AccessibilityElement{DomainID: id, Role: "AXButton", Enabled: enabled})
	}
	return state, tree
}
func TestActionReadinessRejectsPendingSemanticTransitions(t *testing.T) {
	state, tree := readyActionFixture(t)
	if !actionReady(tree, state, "undo") {
		t.Fatal("current Unicode selection/action not ready")
	}
	cases := []struct {
		name   string
		change func(*cogent.AccessibilityInspection)
	}{
		{"unreachable", func(s *cogent.AccessibilityInspection) { s.Reachable = false }},
		{"unpublished", func(s *cogent.AccessibilityInspection) { s.Generation = 0 }},
		{"nonstring value", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "document" {
					s.Elements[i].Value = []string{"malformed"}
				}
			}
		}},
		{"wrong value", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "document" {
					s.Elements[i].Value = "stale"
				}
			}
		}},
		{"same value marked selection", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "document" {
					s.Elements[i].TextSelection = [2]int{1, 0}
				}
			}
		}},
		{"stale status", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "status" {
					s.Elements[i].Value = "Finishing text input"
				}
			}
		}},
		{"disabled action", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "undo" {
					s.Elements[i].Enabled = false
				}
			}
		}},
		{"stale redo availability", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "redo" {
					s.Elements[i].Enabled = true
				}
			}
		}},
		{"stale filename", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "filename" {
					s.Elements[i].Value = "Untitled"
				}
			}
		}},
		{"wrong action role", func(s *cogent.AccessibilityInspection) {
			for i := range s.Elements {
				if s.Elements[i].DomainID == "undo" {
					s.Elements[i].Role = "AXStaticText"
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := tree
			changed.Elements = append([]cogent.AccessibilityElement(nil), tree.Elements...)
			tc.change(&changed)
			if actionReady(changed, state, "undo") {
				t.Fatal("stale native publication accepted")
			}
		})
	}
	if actionReady(tree, state, "absent") {
		t.Fatal("missing action accepted")
	}
	remounted := state
	remounted.EditorID = "document-remounted"
	if actionReady(tree, remounted, "undo") {
		t.Fatal("retired editor ID accepted for remounted binding")
	}
	remountedTree := tree
	remountedTree.Elements = append([]cogent.AccessibilityElement(nil), tree.Elements...)
	for i := range remountedTree.Elements {
		if remountedTree.Elements[i].DomainID == "document" {
			remountedTree.Elements[i].DomainID = remounted.EditorID
		}
	}
	if !actionReady(remountedTree, remounted, "undo") {
		t.Fatal("current remounted binding refused")
	}
	state.Editor.Selection.Focus = 1000
	if actionReady(tree, state, "undo") {
		t.Fatal("invalid scalar selection accepted")
	}
}
func TestActionReadinessSupportsModalScope(t *testing.T) {
	state, tree := readyActionFixture(t)
	state.Prompt = "open"
	tree.Elements = []cogent.AccessibilityElement{{DomainID: "decision"}, {DomainID: "discard-open", Role: "AXButton", Enabled: true}, {DomainID: "keep", Role: "AXButton", Enabled: true}}
	if !actionReady(tree, state, "discard-open") || !actionReady(tree, state, "keep") {
		t.Fatal("current modal actions refused")
	}
	state.Prompt = "close"
	if actionReady(tree, state, "discard-open") {
		t.Fatal("wrong modal action accepted")
	}
	tree.Elements = append(tree.Elements, cogent.AccessibilityElement{DomainID: "document"})
	if actionReady(tree, state, "keep") {
		t.Fatal("outside document leaked into modal scope")
	}
}
func TestActionDeliveryIsDistinctFromAdmission(t *testing.T) {
	before, _ := readyActionFixture(t)
	before.Operation = 8
	if actionDelivered("undo", before, before) {
		t.Fatal("admission with no owner delivery accepted")
	}
	after := before
	after.Operation++
	for _, id := range []string{"undo", "redo", "discard-open"} {
		if !actionDelivered(id, before, after) {
			t.Fatal("single owner delivery refused", id)
		}
	}
	after.Operation++
	if actionDelivered("undo", before, after) {
		t.Fatal("duplicate delivery accepted")
	}
	before.Prompt = "close"
	after = before
	after.Prompt = ""
	if !actionDelivered("keep", before, after) {
		t.Fatal("modal cancellation not observed")
	}
	if actionDelivered("keep", before, before) {
		t.Fatal("undelivered modal action accepted")
	}
}
