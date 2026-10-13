package sdkcontract_test

import (
	"reflect"
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestPublicDisclosureDescriptorAndFocusContract(t *testing.T) {
	expanded := false
	view := func() ui.Node {
		return ui.Node{ID: "inspector", Kind: ui.Disclosure, Text: "Inspector", Checked: expanded,
			Change: func(e ui.ChangeEvent) { expanded = e.Checked }, Children: []ui.Node{{ID: "name", Kind: ui.TextField, Text: "Part"}}}
	}
	if err := ui.Validate(view()); err != nil || !reflect.DeepEqual(ui.FocusOrder(view()), []string{"inspector"}) {
		t.Fatal("public collapsed descriptor", err)
	}
	view().Change(ui.ChangeEvent{ID: "inspector", Kind: ui.Disclosure, Checked: true})
	if !reflect.DeepEqual(ui.FocusOrder(view()), []string{"inspector", "name"}) {
		t.Fatal("public expansion/focus contract")
	}
}
