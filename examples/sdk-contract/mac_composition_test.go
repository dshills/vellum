package sdkcontract_test

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"testing"
)

func TestPublicMacCompositionWidgetsAndTheme(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 220, G: 60, B: 30, A: 255})
	art, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	branch, err := ui.NewMenuBranch([]ui.Option{{ID: "export", Label: "Export PNG", IconImage: art, Shortcut: "⇧⌘E"}})
	if err != nil {
		t.Fatal(err)
	}
	split := &ui.SplitButton{ID: "sdk-split", Label: "Save", IconImage: art, Primary: func() {}, Options: []ui.Option{{ID: "formats", Label: "Formats", Branch: branch}}, Change: func(ui.ChangeEvent) {}}
	inspector := &ui.Inspector{ID: "sdk-inspector", Title: "Properties", IconImage: art, Sections: []ui.InspectorSection{{ID: "geometry", Title: "Geometry", Expanded: true, Properties: []ui.InspectorProperty{{ID: "name", Label: "Name", Mixed: true, Editor: ui.Node{ID: "sdk-name", Kind: ui.TextField, Text: "Drawing"}}}}}}
	browser := &ui.ColumnBrowser{ID: "sdk-browser", Label: "Library", Items: []ui.BrowserItem{{ID: "folder", Label: "Folder", Thumbnail: art, Children: []ui.BrowserItem{{ID: "drawing", Label: "Drawing", Thumbnail: art}}}}, Path: []string{"folder"}, Change: func(ui.BrowserEvent) {}}
	rules := &ui.RuleEditor{ID: "sdk-rules", Fields: []ui.RuleField{{ID: "name", Label: "Name", Kind: ui.RuleText}}, Root: ui.Rule{ID: "all", Group: &ui.RuleGroup{Match: ui.RuleAll, Rules: []ui.Rule{{ID: "text", FieldID: "name", Operator: ui.RuleContains, Value: ui.RuleValue{Text: "Plan"}}}}}, AddIconImage: art, GroupIconImage: art, Change: func(ui.RuleEditEvent) {}}
	constructors := []func() (ui.Node, error){split.View, inspector.View, browser.View, rules.View}
	children := []ui.Node{}
	for _, view := range constructors {
		n, err := view()
		if err != nil {
			t.Fatal(err)
		}
		children = append(children, n)
	}
	for _, dark := range []bool{false, true} {
		theme := ui.MacThemeWithAccent(dark, color.NRGBA{R: 126, G: 65, B: 194, A: 255})
		root := ui.Node{ID: "sdk-root", Kind: ui.Column, Theme: &theme, Children: children}
		if err := ui.Validate(root); err != nil {
			t.Fatal(err)
		}
	}
	if children[0].Children[0].IconImage != art {
		t.Fatal("full-color image replaced")
	}
}
