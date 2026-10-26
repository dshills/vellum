package ui

import (
	"strings"
	"testing"
)

func TestLinkValidationAndLiveActions(t *testing.T) {
	for _, s := range []string{"file:///tmp/a", "javascript:alert(1)", "https://user:pass@example.com", "https://", "//example.com", "https://example.com/a b", "https://example.com/\n", strings.Repeat("x", 4097)} {
		if ValidateLinkURL(s) == nil {
			t.Fatal("bad URL", s)
		}
	}
	calls := []string{}
	l := &Link{ID: "docs", Label: "Docs <plain>", URL: "https://example.com/docs?q=a%20b#here", Open: func(e LinkEvent) { calls = append(calls, "open:"+e.URL) }, Copy: func(e LinkEvent) { calls = append(calls, "copy:"+e.URL) }}
	n, e := l.View()
	if e != nil {
		t.Fatal(e)
	}
	n.Activate()
	n.ContextMenu.Change(ChangeEvent{SelectedID: "copy"})
	if len(calls) != 2 || len(FocusOrder(n)) != 1 {
		t.Fatal(calls)
	}
	l.Disabled = true
	n.Activate()
	n.ContextMenu.Change(ChangeEvent{SelectedID: "copy"})
	if len(calls) != 2 {
		t.Fatal("disabled invoked")
	}
	l.Disabled = false
	l.URL = "https://example.com/changed"
	n.Activate()
	if len(calls) != 2 {
		t.Fatal("stale URL invoked")
	}
	l.ID = "changed"
	n.ContextMenu.Change(ChangeEvent{SelectedID: "copy"})
	if len(calls) != 2 {
		t.Fatal("stale identity invoked")
	}
}
