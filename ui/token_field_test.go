package ui

import (
	"strings"
	"testing"
)

func TestTokenFieldControlledEditsAndStaleRemoval(t *testing.T) {
	f := &TokenField{ID: "tags", Tokens: []Token{{ID: "a", Text: "Design"}}, Suggestions: []Option{{ID: "mac", Label: "macOS"}}}
	var edits []TokenEvent
	f.Change = func(e TokenEvent) { edits = append(edits, e) }
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	helperNode(t, n, TokenInputID("tags")).Change(ChangeEvent{Text: "  Drawing  "})
	helperNode(t, n, TokenRemoveID("tags", "a")).Activate()
	if len(edits) != 2 || edits[0].Text != "Drawing" || !edits[1].Remove || edits[1].TokenID != "a" || len(f.Tokens) != 1 {
		t.Fatal(edits, f.Tokens)
	}
	f.Tokens[0].Text = "Renamed"
	helperNode(t, n, TokenRemoveID("tags", "a")).Activate()
	if len(edits) != 2 {
		t.Fatal("stale chip action")
	}
	f.Disabled = true
	helperNode(t, n, TokenInputID("tags")).Change(ChangeEvent{Text: "hidden"})
	if len(edits) != 2 {
		t.Fatal("disabled")
	}
}
func TestTokenFieldValidationAndCompletionRevocation(t *testing.T) {
	f := &TokenField{ID: "tags", Suggestions: []Option{{ID: "mac", Label: "macOS"}}}
	calls, invalid := 0, 0
	f.Change = func(e TokenEvent) {
		calls++
		if e.SuggestionID != "mac" || e.Text != "macOS" {
			t.Fatal(e)
		}
	}
	f.Invalid = func(error) { invalid++ }
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	menu := helperNode(t, n, ComboMenuID("tags/tokens/entry"))
	menu.Change(ChangeEvent{SelectedID: "mac"})
	if calls != 1 {
		t.Fatal(calls)
	}
	// Refreshing a retained input revokes the prior catalog callback.
	f.Suggestions = nil
	if _, err = f.View(); err != nil {
		t.Fatal(err)
	}
	menu.Change(ChangeEvent{SelectedID: "mac"})
	if calls != 1 {
		t.Fatal("stale suggestion")
	}
	helperNode(t, n, TokenInputID("tags")).Change(ChangeEvent{Text: strings.Repeat("a", 257)})
	if invalid != 1 {
		t.Fatal(invalid)
	}
	f.Tokens = []Token{{ID: "x", Text: "same"}, {ID: "x", Text: "other"}}
	if _, err = f.View(); err == nil {
		t.Fatal("duplicate IDs")
	}
	f.Tokens = make([]Token, MaxTokens+1)
	if _, err = f.View(); err == nil {
		t.Fatal("unbounded")
	}
}

func TestTokenFieldRevokesUnrenderedSuggestionChanges(t *testing.T) {
	for _, change := range []string{"removed", "disabled", "renamed", "metadata", "over-limit"} {
		t.Run(change, func(t *testing.T) {
			calls := 0
			f := &TokenField{ID: "tags", Suggestions: []Option{{ID: "work", Label: "Work"}}, Change: func(TokenEvent) { calls++ }}
			n, err := f.View()
			if err != nil {
				t.Fatal(err)
			}
			menu := helperNode(t, n, ComboMenuID("tags/tokens/entry"))
			switch change {
			case "removed":
				f.Suggestions = nil
			case "disabled":
				f.Suggestions[0].Disabled = true
			case "renamed":
				f.Suggestions[0].Label = "Different"
			case "metadata":
				f.Suggestions[0].Checkable = true
			case "over-limit":
				f.Suggestions = append(f.Suggestions, make([]Option, MaxMenuOptions)...)
			}
			// Owner data changed, but the old rendered action is still queued.
			menu.Change(ChangeEvent{SelectedID: "work"})
			if calls != 0 {
				t.Fatal("obsolete suggestion accepted before next View", change)
			}
		})
	}
}
