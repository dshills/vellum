package ui

import (
	"reflect"
	"testing"
)

func TestIconsPortableContract(t *testing.T) {
	list := Symbols()
	if len(list) != 23 {
		t.Fatal("public symbol catalog", list)
	}
	for _, symbol := range list {
		for _, kind := range []Kind{Icon, Button, ToggleButton} {
			n := Node{ID: "symbol", Kind: kind, Icon: symbol, Accessibility: Accessibility{Name: "Action"}}
			if err := Validate(n); err != nil {
				t.Fatal(kind, symbol, err)
			}
		}
	}
	list[0] = "changed"
	if Symbols()[0] == "changed" {
		t.Fatal("caller mutated the built-in catalog")
	}
	valid := Node{ID: "root", Kind: Column, Children: []Node{
		{ID: "decorative", Kind: Icon, Icon: SymbolInfo},
		{ID: "named", Kind: Icon, Icon: SymbolWarning, Text: "Warning"},
		{ID: "save", Kind: Button, Icon: SymbolSave, Text: "Save", Appearance: ButtonPrimary},
		{ID: "toggle", Kind: ToggleButton, Icon: SymbolCheck, Accessibility: Accessibility{Name: "Show guides"}},
	}}
	if err := Validate(valid); err != nil {
		t.Fatal(err)
	}
	if got := FocusOrder(valid); !reflect.DeepEqual(got, []string{"save", "toggle"}) {
		t.Fatal("symbols added focus stops", got)
	}
	for _, n := range []Node{
		{ID: "bad", Kind: Icon},
		{ID: "bad", Kind: Icon, Icon: "<svg/>"},
		{ID: "bad", Kind: Icon, Icon: SymbolInfo, Activate: func() {}},
		{ID: "bad", Kind: Icon, Icon: SymbolInfo, Change: func(ChangeEvent) {}},
		{ID: "bad", Kind: Icon, Icon: SymbolInfo, Children: []Node{{ID: "child", Kind: Label}}},
		{ID: "bad", Kind: Button, Icon: SymbolSearch},
		{ID: "bad", Kind: ToggleButton, Icon: SymbolCheck, Text: " \t", Accessibility: Accessibility{Name: " "}},
		{ID: "bad", Kind: Label, Icon: SymbolInfo},
		{ID: "bad", Kind: Checkbox, Icon: SymbolCheck},
		{ID: "bad", Kind: Button, Text: "Save", Appearance: "unknown"},
		{ID: "bad", Kind: ToggleButton, Text: "Pin", Appearance: ButtonTool},
	} {
		if err := Validate(n); err == nil {
			t.Fatal("invalid symbol/presentation accepted", n)
		}
	}
}
