package ui

import (
	"reflect"
	"testing"
)

func TestChromeContainersAndFocus(t *testing.T) {
	for _, kind := range []Kind{Toolbar, StatusBar, FormGroup} {
		n := Node{ID: "container", Kind: kind, Text: "Title", Children: []Node{{ID: "first", Kind: Button}, {ID: "disabled", Kind: Button, Disabled: true}, {ID: "last", Kind: TextField}}}
		if err := Validate(n); err != nil {
			t.Fatal(kind, err)
		}
		if !IsContainer(kind) || IsFocusable(kind) || !reflect.DeepEqual(FocusOrder(n), []string{"first", "last"}) {
			t.Fatal("container changed child focus order", kind, FocusOrder(n))
		}
		n.Disabled = true
		if got := FocusOrder(n); len(got) != 0 {
			t.Fatal("disabled container exposed focus", kind, got)
		}
		n.Disabled = false
		n.Activate = func() {}
		if Validate(n) == nil {
			t.Fatal("container activation accepted", kind)
		}
		n.Activate = nil
		n.Change = func(ChangeEvent) {}
		if Validate(n) == nil {
			t.Fatal("container change accepted", kind)
		}
	}
	n := Node{ID: "tools", Kind: Toolbar, Orientation: Vertical}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	n.Kind = StatusBar
	if Validate(n) == nil {
		t.Fatal("status orientation accepted")
	}
	n.Kind = FormGroup
	if Validate(n) == nil {
		t.Fatal("form orientation accepted")
	}
}
