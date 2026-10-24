package sdkcontract_test

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"testing"
	"time"
)

func TestPublicMacEssentialsWithoutNativeDependencies(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 220, G: 40, B: 180, A: 255})
	art, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	well := &ui.ImageWell{ID: "sdk-well", Label: "Artwork", Image: art, ChooseIcon: art, Request: func(ui.ImageWellEvent) bool { return true }}
	picker := &ui.IconPicker{ID: "sdk-picker", Label: "Choose artwork", SelectedID: "drawing", Items: []ui.IconPickerItem{{ID: "drawing", Label: "Drawing", Category: "Documents", Image: art}}, Change: func(ui.IconPickerEvent) {}}
	settings := &ui.SettingsPane{ID: "sdk-settings", Title: "Settings", SelectedID: "general", Pages: []ui.SettingsPage{{ID: "general", Label: "General", IconImage: art, Children: []ui.Node{{ID: "sdk-name", Kind: ui.TextField, Text: "Drawing"}}}}, Navigate: func(ui.SettingsEvent) {}}
	calendar := &ui.Calendar{ID: "sdk-calendar", Value: time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC), FirstWeekday: time.Monday, HeaderIconImage: art, Change: func(ui.CalendarEvent) {}}
	for _, view := range []func() (ui.Node, error){well.View, picker.View, settings.View, calendar.View} {
		n, err := view()
		if err != nil {
			t.Fatal(err)
		}
		if err = ui.Validate(n); err != nil {
			t.Fatal(err)
		}
	}
	options := app.AlertOptions{Title: "Keep drawing?", Style: app.AlertWarning, Buttons: []app.AlertButton{{ID: "keep", Label: "Keep", Role: app.AlertDefault}, {ID: "cancel", Label: "Cancel", Role: app.AlertCancel}}}
	if err := options.Validate(); err != nil {
		t.Fatal(err)
	}
	// Compile the optional service contract without invoking any platform UI.
	var service app.AlertServices
	_ = service
}
