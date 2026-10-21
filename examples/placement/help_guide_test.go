package main

import (
	"context"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestHelpGuideTopicsNoHistoryAndStaleCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		state := f.m.journal.Snapshot()
		handled, err := f.m.commands.Shortcut(context.Background(), "F1")
		if err != nil || !handled || f.m.prompt != "help-guide" {
			t.Fatal("help shortcut", err)
		}
		old := f.m.helpGuideDialog(f.a, f.ref)
		draft := f.m.helpDraft
		for _, topic := range helpTopics {
			geometryDialogControl(t, f.m.helpGuideDialog(f.a, f.ref), "help-topic").Change(ui.ChangeEvent{SelectedID: topic.id})
			root := f.m.helpGuideDialog(f.a, f.ref)
			if err := ui.Validate(root); err != nil {
				t.Fatal(err)
			}
			content := geometryDialogControl(t, root, "help-content")
			if len(content.Children) < 4 {
				t.Fatal("missing guide topic", topic.id)
			}
			for _, p := range content.Children {
				if strings.TrimSpace(p.Text) == "" {
					t.Fatal("empty guide paragraph")
				}
			}
		}
		if state != f.m.journal.Snapshot() {
			t.Fatal("help changed history")
		}
		geometryDialogControl(t, old, "help-close").Activate()
		f.m.prompt = "close"
		old.Dismiss()
		geometryDialogControl(t, old, "help-topic").Change(ui.ChangeEvent{SelectedID: "gestures"})
		if f.m.prompt != "close" || draft.topic != "limits" {
			t.Fatal("stale help callback changed state")
		}
		if err := f.m.commands.Invoke(f.ref.Context(), "help-guide"); err == nil || f.m.prompt != "close" {
			t.Fatal("Help replaced active modal")
		}
		f.m.prompt = ""
		if err := f.m.commands.Invoke(f.ref.Context(), "help-guide"); err != nil {
			t.Fatal(err)
		}
		newer := f.m.helpDraft
		old.Dismiss()
		if f.m.helpDraft != newer || f.m.prompt != "help-guide" {
			t.Fatal("old dismiss closed new guide")
		}
	})
}
