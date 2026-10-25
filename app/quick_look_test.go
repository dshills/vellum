package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type quickLookTestServices struct{ legacyClipboardServices }

func (quickLookTestServices) PreviewFiles(context.Context, WindowRef, QuickLookOptions) error {
	return nil
}
func TestQuickLookOptionsAndOptionalService(t *testing.T) {
	good := QuickLookOptions{Items: []QuickLookItem{{ID: "a", Path: filepath.Join(os.TempDir(), "a.txt"), Title: "A"}}, SelectedID: "a"}
	if good.Validate() != nil {
		t.Fatal("valid")
	}
	for _, o := range []QuickLookOptions{{}, {Items: good.Items, SelectedID: "missing"}, {Items: append(good.Items, good.Items...), SelectedID: "a"}, {Items: []QuickLookItem{{ID: "a", Path: "https://example.org/a", Title: "A"}}, SelectedID: "a"}, {Items: []QuickLookItem{{ID: "a", Path: "/tmp/a", Title: strings.Repeat("x", 257)}}, SelectedID: "a"}} {
		if o.Validate() == nil {
			t.Fatal("invalid", o)
		}
	}
	a, _ := New(Config{})
	defer a.Close()
	a.SetServices(legacyClipboardServices{})
	if _, e := a.QuickLook(); e != ErrUnavailable {
		t.Fatal(e)
	}
	a.SetServices(quickLookTestServices{})
	if _, e := a.QuickLook(); e != nil {
		t.Fatal(e)
	}
}
