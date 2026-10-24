package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestMacToolbarPreferencesRoundTripAndRestore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "toolbars.json")
	m := newMacToolbarGallery(false)
	m.loadPreferences(path)
	findGalleryDocumentNode(m.view(), "toolbar-Regular/customize").Activate()
	n := m.view()
	findGalleryDocumentNode(n, ui.ToolbarCustomizationEarlierID(m.customizer.ID, "toolbar-Regular/Save")).Activate()
	n = m.view()
	findGalleryDocumentNode(n, ui.ToolbarCustomizationVisibilityID(m.customizer.ID, "toolbar-Regular/New")).Change(ui.ChangeEvent{Checked: false})
	if m.preferencesError != nil {
		t.Fatal(m.preferencesError)
	}
	restored := newMacToolbarGallery(false)
	restored.loadPreferences(path)
	if !reflect.DeepEqual(restored.layouts, m.layouts) {
		t.Fatal("layout did not round trip", restored.layouts, m.layouts)
	}
	if len(restored.layouts["toolbar-Small"].Order) != 0 {
		t.Fatal("peer toolbar changed")
	}
	applied := findGalleryDocumentNode(restored.view(), "toolbar-Regular")
	if applied.Children[0].ID != "toolbar-Regular/Save" || applied.Children[0].IconImage == nil {
		t.Fatal("loaded toolbar lost order/color")
	}
	findGalleryDocumentNode(m.view(), ui.ToolbarCustomizationRestoreID(m.customizer.ID)).Activate()
	restored.loadPreferences(path)
	if len(restored.layouts["toolbar-Regular"].Hidden) != 0 {
		t.Fatal("defaults not saved")
	}
	info, e := os.Stat(path)
	if e != nil || info.Mode().Perm() != 0600 {
		t.Fatal("unexpected preference file mode", info, e)
	}
	temps, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".toolbars-*.tmp"))
	if len(temps) != 0 {
		t.Fatal("temporary file leak")
	}
}
func TestMacToolbarPreferencesMigrationAndInvalidFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolbars.json")
	data, _ := json.Marshal(macToolbarPreferences{Version: 1, Layouts: map[string]ui.ToolbarLayout{"toolbar-Regular": {Order: []string{"retired", "toolbar-Regular/Save", "toolbar-Regular/Save"}, Hidden: []string{"retired", "toolbar-Regular/Open"}}}})
	os.WriteFile(path, data, 0600)
	m := newMacToolbarGallery(false)
	m.loadPreferences(path)
	n := findGalleryDocumentNode(m.view(), "toolbar-Regular")
	if n.Children[0].ID != "toolbar-Regular/Save" || m.preferencesError != nil {
		t.Fatal("migration failed", m.preferencesError)
	}
	for _, bad := range []string{"{", `{"version":99,"layouts":{}}`, string(make([]byte, toolbarPreferenceLimit+1))} {
		os.WriteFile(path, []byte(bad), 0600)
		m = newMacToolbarGallery(false)
		m.loadPreferences(path)
		if m.preferencesError == nil || len(m.layouts) != 0 {
			t.Fatal("invalid preferences accepted")
		}
		m.savePreferences()
		after, _ := os.ReadFile(path)
		if string(after) != bad {
			t.Fatal("invalid/newer preferences overwritten")
		}

	}
}
func TestMacToolbarPreferencesSaveFailureRetainsMemory(t *testing.T) {
	m := newMacToolbarGallery(false)
	blocker := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocker, []byte("untouched"), 0600)
	m.preferencesPath = filepath.Join(blocker, "toolbars.json")
	m.layouts["toolbar-Regular"] = ui.ToolbarLayout{Hidden: []string{"toolbar-Regular/New"}}
	m.savePreferences()
	if m.preferencesError == nil || len(m.layouts["toolbar-Regular"].Hidden) != 1 {
		t.Fatal("save failure lost in-memory setting")
	}
	after, _ := os.ReadFile(blocker)
	if string(after) != "untouched" {
		t.Fatal("save damaged existing file")
	}
}

func TestMacToolbarDisplayPreferencesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolbars.json")
	m := newMacToolbarGallery(false)
	m.loadPreferences(path)
	n := m.view()
	findGalleryDocumentNode(n, "toolbar-Regular").ContextMenu.Change(ui.ChangeEvent{SelectedID: "above"})
	restored := newMacToolbarGallery(false)
	restored.loadPreferences(path)
	if restored.displays["toolbar-Regular"] != "above" || findGalleryDocumentNode(restored.view(), "toolbar-Regular/Save").IconPlacement != ui.IconAboveText {
		t.Fatal("display mode did not round trip")
	}
	findGalleryDocumentNode(m.view(), "toolbar-display/text").Activate()
	restored.loadPreferences(path)
	if restored.display != "text" || len(restored.displays) != 0 {
		t.Fatal("global display choice did not round trip")
	}
}

func TestMacToolbarDisplayLegacyReloadResetsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolbars.json")
	os.WriteFile(path, []byte(`{"version":1,"layouts":{}}`), 0600)
	m := newMacToolbarGallery(false)
	m.display = "above"
	m.displays["toolbar-Regular"] = "text"
	m.loadPreferences(path)
	if m.display != "both" || len(m.displays) != 0 || m.preferencesError != nil {
		t.Fatal("legacy file retained stale display")
	}
}
