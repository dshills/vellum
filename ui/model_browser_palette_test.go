package ui

import (
	"image/color"
	"testing"
)

func TestSelectedBrowserActionsAndBadgeUseAccentInk(t *testing.T) {
	p := ModelBrowser{ID: "b", Label: "Browser", SelectedID: "one", ActiveID: "one", Items: []ModelBrowserItem{{ID: "one", Label: "One", Badge: "1", Visible: true, Activatable: true}}, Change: func(ModelBrowserEvent) {}}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	for _, dark := range []bool{false, true} {
		theme := MacThemeWithAccent(dark, color.NRGBA{220, 165, 30, 255})
		var walk func(Node)
		walk = func(n Node) {
			for _, id := range []string{"/select", "/visibility", "/active", "/badge"} {
				if n.ID == ModelBrowserRowID("b", "one")+id {
					s, e := ResolveStyle(n, theme)
					if e != nil || s.Foreground.Value != theme.Palette.Value.AccentInk {
						t.Fatal("selection contrast", n.ID, e)
					}
				}
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(n)
	}
}
