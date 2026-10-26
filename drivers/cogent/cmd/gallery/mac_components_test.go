package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacComponentThemePreviewRetainsArtworkAndValue(t *testing.T) {
	m := newMacComponentsGallery(false)
	n := m.view()
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	original := findGalleryDocumentNode(n, "components/color-button").IconImage
	findGalleryDocumentNode(n, "components/name").Change(ui.ChangeEvent{Text: "Drawing"})
	findGalleryDocumentNode(n, "components/appearance/Dark").Activate()
	findGalleryDocumentNode(n, "components/accent").Change(ui.ChangeEvent{SelectedID: "purple"})
	next := m.view()
	if err := ui.Validate(next); err != nil {
		t.Fatal(err)
	}
	if !m.dark || m.accent != "purple" || m.name != "Drawing" || findGalleryDocumentNode(next, "components/color-button").IconImage != original {
		t.Fatal("theme transition lost state or artwork")
	}
}

func TestMacComponentPagesAreConsumable(t *testing.T) {
	for _, dark := range []bool{false, true} {
		m := newMacComponentsGallery(dark)
		for _, page := range []string{"media-player", "pdf", "date-range", "path-picker", "checklist", "theme", "menus", "split", "inspector", "browser", "rules", "image-well", "icon-picker", "settings", "alerts", "calendar", "documents", "outline", "native-colors", "native-fonts", "angle-dial", "quick-look", "find-replace", "search-scopes", "gradient", "range-slider"} {
			m.page = page
			n := m.view()
			if err := ui.Validate(n); err != nil {
				t.Fatal(page, err)
			}
			var checkErrors func(ui.Node)
			checkErrors = func(node ui.Node) {
				if strings.HasSuffix(node.ID, "/error") || strings.HasSuffix(node.ID, "-error") {
					t.Fatal(page, node.ID, node.Text)
				}
				for _, child := range node.Children {
					checkErrors(child)
				}
			}
			checkErrors(n)
			if page == "browser" && findGalleryDocumentNode(n, "components/browser").ID == "" {
				t.Fatal("browser failed to construct")
			}
		}
	}
}

func TestMacComponentRuleEditsRetainAcrossTheme(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "rules"
	n := m.view()
	add := findGalleryDocumentNode(n, ui.RuleControlID("components/rules", "root", "add-rule"))
	if add.IconImage == nil {
		t.Fatal("missing color artwork")
	}
	add.Activate()
	m.dark = true
	m.accent = "green"
	n = m.view()
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	if len(m.rules.Root.Group.Rules) != 4 || findGalleryDocumentNode(n, ui.RuleControlID("components/rules", "root", "add-rule")).IconImage != add.IconImage {
		t.Fatal("theme lost rules or art")
	}
}

func TestMacComponentMenuChecksAreIndependentOfActionCount(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "menus"
	menuView := func() ui.Node {
		t.Helper()
		n := findGalleryDocumentNode(m.view(), "components/actions")
		if n.ID != "components/actions" || n.Change == nil || len(n.Options) != 3 || n.Options[2].ID != "grid" {
			t.Fatal("missing menu demonstration/check option")
		}
		return n
	}
	menu := menuView()
	menu.Change(ui.ChangeEvent{SelectedID: "export-png"})
	menu = menuView()
	if menu.Options[2].Checked {
		t.Fatal("export toggled grid")
	}
	menu.Change(ui.ChangeEvent{SelectedID: "grid", Checked: true})
	if !menuView().Options[2].Checked {
		t.Fatal("grid proposal not accepted")
	}
}

func TestMacImageWellSampleAndOffscreenChoose(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "image-well"
	n := m.view()
	if err := ui.Validate(n); err != nil {
		t.Fatal(err)
	}
	findGalleryDocumentNode(n, "essentials/sample-image").Activate()
	if m.essentials.well.Image == nil || m.count != 1 {
		t.Fatal("sample not accepted")
	}
	n = m.view()
	findGalleryDocumentNode(n, "essentials/well/image-well/choose").Activate()
	if m.essentials.well.Busy {
		t.Fatal("offscreen chooser started worker")
	}
}

func TestImageWellLoaderRejectsDirectoriesAndOversize(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadWellImage(dir); err == nil {
		t.Fatal("directory admitted")
	}
	path := filepath.Join(dir, "huge.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(ui.MaxEncodedImageBytes + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := loadWellImage(path); err == nil {
		t.Fatal("oversize file admitted")
	}
}

func TestMacSettingsPreferencesRemainModelOwned(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "settings"
	n := m.view()
	findGalleryDocumentNode(n, "essentials/settings/name").Change(ui.ChangeEvent{Text: "Renamed"})
	m.essentials.settings.SelectedID = "appearance"
	m.view()
	m.essentials.settings.SelectedID = "general"
	n = m.view()
	if findGalleryDocumentNode(n, "essentials/settings/name").Text != "Renamed" {
		t.Fatal("preference lost")
	}
}

func TestMacFontLatePreviewCannotOverwriteFinalOrNewSession(t *testing.T) {
	m := newMacComponentsGallery(false)
	m.page = "native-fonts"
	m.view()
	e := m.documents
	e.nativeBusy = true
	e.nativeEpoch = 1
	preview := app.FontPanelSelection{Name: "Menlo-Regular", Family: "Menlo", Size: 24}
	if !e.acceptFontPreview(1, preview) {
		t.Fatal("current preview refused")
	}
	accepted := e.fontValue
	e.nativeBusy = false
	e.fontPreview = accepted
	if e.acceptFontPreview(1, preview) || e.fontPreview != accepted {
		t.Fatal("late preview overwrote completion")
	}
	e.nativeBusy = true
	e.nativeEpoch = 2
	if e.acceptFontPreview(1, preview) {
		t.Fatal("old session preview admitted")
	}
}
