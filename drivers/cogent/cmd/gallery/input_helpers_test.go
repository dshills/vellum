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

func TestInputHelpersResetRestoresWholeDemoAndRevokesOldEditors(t *testing.T) {
	for _, dark := range []bool{false, true} {
		m := newInputHelpersGallery()
		m.setAppearance(dark)
		old := m.view()
		secure := findGalleryDocumentNode(old, ui.SecureInputID("demo-password"))
		token := findGalleryDocumentNode(old, ui.TokenInputID("drawing-tags"))
		font := findGalleryDocumentNode(old, ui.FontSizeID("drawing-font"))
		path := findGalleryDocumentNode(old, ui.PathSegmentID("drawing-path", "root"))
		combo := findGalleryDocumentNode(old, ui.ComboInputID("unit-combo"))
		date := findGalleryDocumentNode(old, ui.DateTimeInputID("reminder-date"))
		reveal := findGalleryDocumentNode(old, "demo-password/secure/reveal")
		oldReset := findGalleryDocumentNode(old, "input-notice/notice/action")
		findGalleryDocumentNode(old, "demo-password/secure/reveal").Activate()
		m.secureField.Text = "changed-demo-text"
		m.pathControl.Items = m.pathControl.Items[:1]
		m.fontPicker.Value.Size = 30
		m.tokens.Tokens = nil
		m.tokens.Draft = "draft"
		m.colorWell.Color.R = 99
		m.combo.Text = "Inches"
		m.datePicker.Value = m.datePicker.Value.AddDate(0, 1, 0)
		m.selectedObjects = []string{"block"}
		m.rejectSelection = true
		m.reject = true
		m.requests = 7
		m.contextAction, m.contextChecked, m.popoverOpen = "grid", true, true
		m.fontError, m.tokenError, m.colorError, m.dateError = "font", "tags", "color", "date"
		findGalleryDocumentNode(m.view(), "input-notice/notice/action").Activate()
		fresh := newInputHelpersGallery()
		if m.dark != dark || m.disabled || m.name != fresh.name || m.quantity != fresh.quantity || m.search.Text != "" || m.requests != 0 || m.reject || m.rejectSelection || len(m.selectedObjects) != 0 || m.contextAction != "" || m.contextChecked || m.popoverOpen {
			t.Fatal("reset left edited gallery state", dark)
		}
		if m.secureField.Text != fresh.secureField.Text || len(m.pathControl.Items) != 5 || m.fontPicker.Value != fresh.fontPicker.Value || len(m.tokens.Tokens) != 1 || m.tokens.Tokens[0].Text != "Design" || m.tokens.Draft != "" || m.colorWell.Color != fresh.colorWell.Color || m.combo.Text != "Millimeters" || !m.datePicker.Value.Equal(fresh.datePicker.Value) || m.fontError+m.tokenError+m.colorError+m.dateError != "" {
			t.Fatal("reset left edited widget state", dark)
		}
		if findGalleryDocumentNode(m.view(), ui.SecureInputID("demo-password")).Reveal {
			t.Fatal("reset kept password revealed")
		}
		// Old controlled editor proposals must not update the fresh demo.
		secure.Change(ui.ChangeEvent{Text: "stale"})
		token.Change(ui.ChangeEvent{Text: "stale"})
		font.Change(ui.ChangeEvent{Number: 99})
		path.Activate()
		combo.Change(ui.ChangeEvent{Text: "stale"})
		date.Change(ui.ChangeEvent{Text: "2030-01-01 00:00"})
		reveal.Activate()
		if m.secureField.Text != fresh.secureField.Text || findGalleryDocumentNode(m.view(), ui.SecureInputID("demo-password")).Reveal || !m.datePicker.Value.Equal(fresh.datePicker.Value) || len(m.pathControl.Items) != 5 || m.fontPicker.Value.Size != 16 || len(m.tokens.Tokens) != 1 || m.combo.Text != "Millimeters" {
			t.Fatal("old editor modified reset demo")
		}
		m.name = "Keep current edits"
		oldReset.Activate()
		if m.name != "Keep current edits" {
			t.Fatal("old reset action remained live")
		}
		// New callbacks must be bound to the same owner, rather than a copied model.
		now := m.view()
		findGalleryDocumentNode(now, ui.SecureInputID("demo-password")).Change(ui.ChangeEvent{Text: "fresh-demo-text"})
		findGalleryDocumentNode(now, ui.TokenInputID("drawing-tags")).Change(ui.ChangeEvent{Text: "Review"})
		findGalleryDocumentNode(now, ui.FontSizeID("drawing-font")).Change(ui.ChangeEvent{Number: 21})
		if m.secureField.Text != "fresh-demo-text" || len(m.tokens.Tokens) != 2 || m.tokens.Tokens[1].Text != "Review" || m.fontPicker.Value.Size != 21 {
			t.Fatal("fresh callback lost original gallery owner")
		}
		if err := ui.Validate(m.view()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInputHelpersProgrammaticResetPreservesDisabledAppearance(t *testing.T) {
	m := newInputHelpersGallery()
	m.setAppearance(true)
	m.setDisabled(true)
	m.name = "Changed"
	m.resetFields()
	if !m.dark || !m.disabled || !m.secureField.Disabled || !m.tokens.Disabled || !m.fontPicker.Disabled || !m.pathControl.Disabled || !m.colorWell.Disabled || !m.notice.Disabled || m.name != "Untitled" {
		t.Fatal("programmatic reset changed appearance/availability")
	}
	if err := ui.Validate(m.view()); err != nil {
		t.Fatal(err)
	}
}
