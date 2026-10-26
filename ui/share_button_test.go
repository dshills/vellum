package ui

import "testing"

func TestShareButtonControlledActivation(t *testing.T) {
	calls := 0
	s := &ShareButton{ID: "share", Share: func(e ShareEvent) {
		if e.ID != "share" {
			t.Fatal(e)
		}
		calls++
	}}
	n, e := s.View()
	if e != nil {
		t.Fatal(e)
	}
	if n.Icon != SymbolShare {
		t.Fatal("missing share artwork")
	}
	n.Activate()
	s.Disabled = true
	n.Activate()
	s.Disabled = false
	s.ID = "new"
	n.Activate()
	if calls != 1 {
		t.Fatal(calls)
	}
}
