package sdkcontract_test

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

// Independent module compilation, with no driver internals or native side effects.
func TestPublicPostV1MacWidgets(t *testing.T) {
	help, e := ui.WithTooltip(ui.Node{ID: "help", Kind: ui.Button, Text: "Help"}, "Plain help")
	if e != nil || help.Tooltip == "" {
		t.Fatal(e)
	}
	link := &ui.Link{ID: "link", Label: "Docs", URL: "https://example.com", Open: func(ui.LinkEvent) {}}
	if _, e = link.View(); e != nil {
		t.Fatal(e)
	}
	share := &ui.ShareButton{ID: "share", Share: func(ui.ShareEvent) {}}
	if _, e = share.View(); e != nil {
		t.Fatal(e)
	}
	if e = (app.ShareOptions{AnchorID: "share", Items: []app.ShareItem{{Kind: app.ShareText, Value: "Synthetic"}}}).Validate(); e != nil {
		t.Fatal(e)
	}
	zone := &ui.FileDropZone{ID: "files", Extensions: []string{"txt"}, Change: func(ui.FileDropZoneEvent) error { return nil }}
	if _, e = zone.View(); e != nil {
		t.Fatal(e)
	}
	palette := &ui.CommandPalette{ID: "commands", Items: []ui.CommandPaletteItem{{ID: "save", Label: "Save"}}, Invoke: func(ui.CommandPaletteEvent) error { return nil }}
	if _, e = palette.View(); e != nil {
		t.Fatal(e)
	}
}
