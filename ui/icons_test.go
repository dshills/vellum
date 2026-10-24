package ui

import (
	"github.com/dshills/vellum/canvas"
	"reflect"
	"testing"
)

func TestIconsPortableContract(t *testing.T) {
	list := Symbols()
	if len(list) != 36 {
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

func TestFullColorButtonIconContract(t *testing.T) {
	source := imageSource(t)
	for _, kind := range []Kind{Button, MenuButton} {
		n := Node{ID: "color-action", Kind: kind, IconImage: source, Accessibility: Accessibility{Name: "Open"}}
		if kind == MenuButton {
			n.Options = []Option{{ID: "open", Label: "Open"}}
			n.Change = func(ChangeEvent) {}
		}
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
		if got := FocusOrder(n); len(got) != 1 || got[0] != n.ID {
			t.Fatal("artwork added focus", got)
		}
	}
	for _, n := range []Node{
		{ID: "bad", Kind: Button, IconImage: source},
		{ID: "bad", Kind: Button, Text: "Open", IconImage: source, Icon: SymbolOpen},
		{ID: "bad", Kind: Label, Text: "Open", IconImage: source},
		{ID: "bad", Kind: ToggleButton, Text: "Open", IconImage: source},
		{ID: "bad", Kind: Button, Text: "Open", IconImage: &canvas.TiledImage{}},
	} {
		if err := Validate(n); err == nil {
			t.Fatal("invalid full-color icon accepted", n.Kind)
		}
	}
}

func TestButtonIconPlacementPortableContract(t *testing.T) {
	source := imageSource(t)
	for _, n := range []Node{
		{ID: "inline", Kind: Button, Icon: SymbolOpen, IconPlacement: IconLeading},
		{ID: "stacked", Kind: Button, Text: "Open", Icon: SymbolOpen, IconPlacement: IconAboveText},
		{ID: "color", Kind: Button, Text: "Open", IconImage: source, IconPlacement: IconAboveText},
	} {
		if n.Text == "" {
			n.Accessibility.Name = "Open"
		}
		if err := Validate(n); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []Node{
		{ID: "bad", Kind: Button, Text: "Open", Icon: SymbolOpen, IconPlacement: "unknown"},
		{ID: "bad", Kind: Button, Text: "Open", IconPlacement: IconAboveText},
		{ID: "bad", Kind: Button, Icon: SymbolOpen, Accessibility: Accessibility{Name: "Open"}, IconPlacement: IconAboveText},
		{ID: "bad", Kind: Label, Text: "Open", IconPlacement: IconLeading},
		{ID: "bad", Kind: MenuButton, Text: "Open", Icon: SymbolOpen, IconPlacement: IconAboveText, Options: []Option{{ID: "open", Label: "Open"}}},
	} {
		if err := Validate(n); err == nil {
			t.Fatal("invalid placement accepted", n.Kind, n.IconPlacement)
		}
	}
}
