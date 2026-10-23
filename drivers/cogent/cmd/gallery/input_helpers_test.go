package main

import (
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestInputHelpersGalleryPublicModels(t *testing.T) {
	g := newGallery()
	if err := ui.Validate(g.view()); err != nil {
		t.Fatal(err)
	}
	m := g.inputHelpers
	node := func(id string) ui.Node { return findGalleryDocumentNode(m.view(), id) }
	node(ui.SearchInputID("object-search")).Change(ui.ChangeEvent{Text: "note"})
	if m.search.Text != "note" || node("match-2").Text != "Note" {
		t.Fatal("query did not filter")
	}
	m.reject = true
	node(ui.SearchClearID("object-search")).Activate()
	if m.search.Text != "note" || m.requests != 2 {
		t.Fatal("rejected clear changed state")
	}
	m.reject = false
	node(ui.SearchClearID("object-search")).Activate()
	if m.search.Text != "" {
		t.Fatal("clear did not restore results")
	}
	node("drawing-name").Change(ui.ChangeEvent{Text: ""})
	if node("drawing-name-field/field/error").Text != "Error: Enter a drawing name." {
		t.Fatal("missing validation text")
	}
	node("input-notice/notice/action").Activate()
	if m.name != "Untitled" {
		t.Fatal("notice action did not reset")
	}
	m.setDisabled(true)
	if err := ui.Validate(m.view()); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		m.setAppearance(dark)
		if err := ui.Validate(m.view()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInputHelpersGalleryReportsDescriptorError(t *testing.T) {
	m := newInputHelpersGallery()
	m.search.ID = ""
	n := m.view()
	if n.ID != "inputs" || n.Kind != ui.Label || !strings.Contains(n.Text, "Cannot render widgets:") {
		t.Fatal("missing controlled render error", n)
	}
}

func TestInputHelpersMacSpacingAndReset(t *testing.T) {
	m := newInputHelpersGallery()
	theme := ui.MacTheme(false)
	if theme.Kinds[ui.FormGroup].Padding.Value.Left < 12 {
		t.Fatal("groups require visible insets")
	}
	m.quantity = 23
	m.name = "Draft"
	m.search.Text = "note"
	findGalleryDocumentNode(m.view(), "input-notice/notice/action").Activate()
	if m.quantity != 1 || m.name != "Untitled" || m.search.Text != "" {
		t.Fatal("Reset fields did not reset each field")
	}
}

func TestMacWidgetExecutableDefaults(t *testing.T) {
	for _, tc := range []struct {
		name          string
		widgets, dark bool
	}{{"/tmp/gallery", false, false}, {"/tmp/Vellum Mac Widgets", true, false}, {"/tmp/Vellum Mac Widgets Dark", true, true}, {"/tmp/Vellum Mac Widgets Other", false, false}} {
		widgets, dark := macWidgetExecutableDefaults(tc.name)
		if widgets != tc.widgets || dark != tc.dark {
			t.Fatal(tc, widgets, dark)
		}
	}
}

func TestPersonalWidgetGalleryModelsAndDisabled(t *testing.T) {
	m := newInputHelpersGallery()
	node := func(id string) ui.Node { return findGalleryDocumentNode(m.view(), id) }
	node(ui.TokenInputID("drawing-tags")).Change(ui.ChangeEvent{Text: "Review"})
	if len(m.tokens.Tokens) != 2 || m.tokens.Tokens[1].Text != "Review" {
		t.Fatal("token addition")
	}
	node(ui.TokenRemoveID("drawing-tags", "tag-1")).Activate()
	if len(m.tokens.Tokens) != 1 {
		t.Fatal("remove")
	}
	node(ui.FontSizeID("drawing-font")).Change(ui.ChangeEvent{Number: 21})
	if m.fontPicker.Value.Size != 21 {
		t.Fatal("size")
	}
	node(ui.PathSegmentID("drawing-path", "root")).Activate()
	if len(m.pathControl.Items) != 1 {
		t.Fatal("path navigation")
	}
	node(ui.SecureInputID("demo-password")).Change(ui.ChangeEvent{Text: "test-only"})
	if m.secureField.Text != "test-only" {
		t.Fatal("secure commit")
	}
	for _, dark := range []bool{false, true} {
		m.setAppearance(dark)
		if err := ui.Validate(m.view()); err != nil {
			t.Fatal(err)
		}
	}
	m.setDisabled(true)
	if !m.tokens.Disabled || !m.fontPicker.Disabled || !m.pathControl.Disabled || !m.secureField.Disabled || !m.colorWell.Disabled {
		t.Fatal("missing disable propagation")
	}
	if err := ui.Validate(m.view()); err != nil {
		t.Fatal(err)
	}
}
