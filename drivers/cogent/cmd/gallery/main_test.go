package main

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func node(n ui.Node, id string) ui.Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if result := node(c, id); result.ID != "" {
			return result
		}
	}
	return ui.Node{}
}
func TestGalleryControlledValuesDialogAndFocus(t *testing.T) {
	g := newGallery()
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	node(g.view(), "slider").Change(ui.ChangeEvent{Number: 60})
	if node(g.view(), "number").Number != 60 {
		t.Fatal("numeric controls do not share accepted model value")
	}
	node(g.view(), "focus-name").Activate()
	if g.view().Focus.ID != "project-name" || g.view().Focus.Sequence != 1 {
		t.Fatal("focus request not issued")
	}
	node(g.view(), "enable").Change(ui.ChangeEvent{Checked: false})
	if !node(g.view(), "fields").Disabled {
		t.Fatal("checkbox doesn't disable fields")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal("consumed focus request should survive later disabled state", err)
	}
	node(g.view(), "show-dialog").Activate()
	if !node(g.view(), "preferences").Open {
		t.Fatal("dialog didn't open")
	}
	node(g.view(), "preferences").Dismiss()
	if g.dialog {
		t.Fatal("dismissal didn't close model state")
	}
}

func TestGalleryExpandedControls(t *testing.T) {
	g := newGallery()
	node(g.view(), "snapping").Change(ui.ChangeEvent{Checked: true})
	node(g.view(), "pinned").Change(ui.ChangeEvent{Checked: true})
	node(g.view(), "slider").Change(ui.ChangeEvent{Number: 75})
	if !node(g.view(), "snapping").Checked || !node(g.view(), "pinned").Checked || node(g.view(), "progress").Number != 0.75 {
		t.Fatal("expanded gallery controls do not reflect accepted state")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}

func TestGalleryDisclosureControlledExpansion(t *testing.T) {
	g := newGallery()
	node(g.view(), "details").Change(ui.ChangeEvent{Checked: false})
	if node(g.view(), "details").Checked {
		t.Fatal("collapse not accepted by gallery model")
	}
	if _, ok := ui.FocusTarget(g.view(), "project-name"); ok {
		t.Fatal("collapsed gallery field remains a focus target")
	}
	node(g.view(), "details").Change(ui.ChangeEvent{Checked: true})
	if _, ok := ui.FocusTarget(g.view(), "project-name"); !ok || ui.Validate(g.view()) != nil {
		t.Fatal("expanded gallery field unavailable")
	}
}

func TestGallerySharedChoiceGroups(t *testing.T) {
	g := newGallery()
	node(g.view(), "radio-preset").Change(ui.ChangeEvent{SelectedID: "advanced"})
	if node(g.view(), "segmented-preset").SelectedID != "advanced" || node(g.view(), "selection").SelectedID != "advanced" {
		t.Fatal("choice controls do not share accepted preset")
	}
	node(g.view(), "segmented-preset").Change(ui.ChangeEvent{SelectedID: "basic"})
	if node(g.view(), "radio-preset").SelectedID != "basic" {
		t.Fatal("segmented choice did not update radio value")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}

func TestGalleryChromeContainers(t *testing.T) {
	g := newGallery()
	if node(g.view(), "fields").Kind != ui.FormGroup || node(g.view(), "more-controls").Kind != ui.Toolbar || node(g.view(), "state").Kind != ui.StatusBar {
		t.Fatal("gallery missing reusable containers")
	}
	node(g.view(), "segmented-preset").Change(ui.ChangeEvent{SelectedID: "advanced"})
	if node(g.view(), "state").Text != "Preset advanced · project Vellum" {
		t.Fatal("status does not reflect accepted state")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}

func TestGallerySplitPane(t *testing.T) {
	g := newGallery()
	node(g.view(), "pages").Change(ui.ChangeEvent{SelectedID: "split-demo"})
	node(g.view(), "panes").Change(ui.ChangeEvent{Number: 0.7})
	if g.tab != "split-demo" || node(g.view(), "panes").Number != 0.7 {
		t.Fatal("split pane does not use accepted gallery state")
	}
	node(g.view(), "split-reset").Activate()
	if node(g.view(), "panes").Number != 0.35 {
		t.Fatal("reset did not restore accepted divider value")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}

func TestGalleryIconsAndActionAppearances(t *testing.T) {
	g := newGallery()
	node(g.view(), "pages").Change(ui.ChangeEvent{SelectedID: "symbols"})
	for _, symbol := range ui.Symbols() {
		if art := node(g.view(), "art-"+string(symbol)); art.Kind != ui.Icon || art.Icon != symbol {
			t.Fatal("gallery omitted bundled artwork", symbol)
		}
	}
	node(g.view(), "tool-action").Activate()
	if g.tab != "controls" || g.focus.ID != "project-name" || g.focus.Sequence != 1 {
		t.Fatal("tool action did not select/focus its application target")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}

func TestGalleryManagedImageFitAndReplacement(t *testing.T) {
	g := newGallery()
	old := g.picture
	node(g.view(), "image-fit").Change(ui.ChangeEvent{SelectedID: string(ui.ImageCover)})
	node(g.view(), "image-replace").Activate()
	preview := node(g.view(), "image-preview")
	if preview.Kind != ui.Image || preview.Fit != ui.ImageCover || preview.Image == old || preview.Image.NRGBAAt(4, 4).G != 192 || old.NRGBAAt(4, 4).G != 64 {
		t.Fatal("image replacement did not retain immutable prior source")
	}
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
}
