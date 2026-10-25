package sdkcontract_test

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPublicQuickLookContract(t *testing.T) {
	o := app.QuickLookOptions{Items: []app.QuickLookItem{{ID: "a", Path: "/tmp/a.txt", Title: "A"}}, SelectedID: "a"}
	if o.Validate() != nil {
		t.Fatal("options")
	}
	d := &ui.DocumentList{ID: "d", Label: "Files", Items: []ui.DocumentItem{{ID: "a", Name: "A"}}, Preview: func(ui.DocumentEvent) {}}
	if _, e := d.View(); e != nil {
		t.Fatal(e)
	}
}
