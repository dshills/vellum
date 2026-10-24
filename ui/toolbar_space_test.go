package ui

import (
	"reflect"
	"testing"
)

func TestToolbarSpacesCustomizationAndOverflow(t *testing.T) {
	c := customizationCatalogue()
	c.Children = append(c.Children, Node{ID: "fixed", Kind: ToolbarSpace}, Node{ID: "flex", Kind: ToolbarFlexibleSpace})
	if e := Validate(WithToolbarOverflow(c)); e != nil {
		t.Fatal(e)
	}
	layout := ToolbarLayout{Order: []string{"c", "flex", "a", "fixed", "b"}, Hidden: []string{"c", "fixed"}}
	n, e := layout.Apply(c)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(customizationIDs(n), []string{"flex", "a", "b"}) {
		t.Fatal(customizationIDs(n))
	}
	if got := FocusOrder(n); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatal("spaces entered focus order", got)
	}
	c1 := ToolbarCustomizer{ID: "customize", Toolbar: c, Open: true, Value: layout, Dismiss: func() {}, Change: func(ToolbarLayout) {}}
	panel, e := c1.View()
	if e != nil {
		t.Fatal(e)
	}
	if customizationNode(t, panel, ToolbarCustomizationVisibilityID(c1.ID, "flex")).Text != "Flexible Space" {
		t.Fatal("missing spacer label")
	}
	hidden, e := (ToolbarLayout{Hidden: []string{"a", "b", "c"}}).Apply(c)
	if e != nil || len(hidden.Children) != 0 {
		t.Fatal("all-hidden bar retained spaces")
	}
	bad := Node{ID: "fixed", Kind: ToolbarSpace, Activate: func() {}}
	if Validate(bad) == nil {
		t.Fatal("interactive spacer accepted")
	}
}
