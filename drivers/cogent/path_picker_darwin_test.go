//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPathPickerRecipeWithoutDialog(t *testing.T) {
	for _, kind := range []ui.FilePickerKind{ui.FilePickerFile, ui.FilePickerFolder} {
		b, _ := json.Marshal(app.PathOptions{Kind: kind, Title: "Synthetic"})
		if !nativePathRecipeFixture(string(b)) {
			t.Fatal(kind)
		}
	}
	for _, wire := range []string{`{}`, `{"Kind":"other","Title":"x","InitialPath":"","Extensions":[]}`, `{"Kind":"folder","Title":"x","InitialPath":"","Extensions":["txt"]}`} {
		if nativePathRecipeFixture(wire) {
			t.Fatal(wire)
		}
	}
}

func TestPathPickerCancellationCleanupWithoutDialog(t *testing.T) {
	_, err := nativeOperationRun(context.Background(), func(token uint64) {
		if !nativePathCleanupFixture(token) {
			t.Fatal("path session cleanup or terminal idempotence")
		}
	}, nil)
	if !errors.Is(err, app.ErrCanceled) {
		t.Fatal(err)
	}
}
