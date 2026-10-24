package cogent

import (
	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestScopedThemeSwitchRemovalAndSiblingIsolation(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("theme scopes"))
	base := ui.MacTheme(false)
	dark := ui.MacThemeWithAccent(true, color.NRGBA{R: 130, G: 65, B: 190, A: 255})
	root := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "scoped", Kind: ui.Column, Theme: &dark, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "keep"}}}, {ID: "sibling", Kind: ui.Button, Text: "Unchanged"}}}
	rec, err := reconcile(host, nil, root, a, "main", base)
	if err != nil {
		t.Fatal(err)
	}
	field := rec.children[0].children[0]
	original := field.widget
	if field.theme.Palette.Value != dark.Palette.Value || rec.children[1].theme.Palette.Value != base.Palette.Value {
		t.Fatal("scope leaked")
	}
	snapshot := field.theme.Kinds[ui.TextField]
	dark.Kinds[ui.TextField] = ui.Style{}
	if field.theme.Kinds[ui.TextField] != snapshot {
		t.Fatal("published theme alias")
	}
	root.Children[0].Theme = nil
	next, err := reconcile(host, rec, root, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if next.children[0].children[0].widget != original || next.children[0].children[0].theme.Palette.Value != base.Palette.Value {
		t.Fatal("scope removal lost retained control or base theme")
	}
	invalid := base.Clone()
	invalid.Default.MinSize = ui.Some(ui.Size{Width: 500})
	invalid.Default.MaxSize = ui.Some(ui.Size{Width: 1})
	root.Children[0].Theme = &invalid
	if _, err = reconcile(host, next, root, a, "main", base); err == nil {
		t.Fatal("invalid theme update admitted")
	}
	if next.children[0].children[0].widget != original || next.children[0].children[0].theme.Palette.Value != base.Palette.Value {
		t.Fatal("rejection partially published theme")
	}
}

func TestScopedThemeMountThenReconcileWithoutExplicitTheme(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("mount scope"))
	base := ui.MacTheme(false)
	scoped := ui.MacTheme(true)
	root := ui.Node{ID: "root", Kind: ui.Column, Theme: &scoped, Children: []ui.Node{{ID: "leaf", Kind: ui.Button, Text: "Keep"}}}
	rec, err := mount(host, root, a, "main", base)
	if err != nil {
		t.Fatal(err)
	}
	leaf := rec.children[0].widget
	root.Theme = nil
	next, err := reconcile(host, rec, root, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if next.children[0].widget != leaf || next.children[0].theme.Palette.Value != base.Palette.Value {
		t.Fatal("mounted theme scope did not restore base")
	}
}
