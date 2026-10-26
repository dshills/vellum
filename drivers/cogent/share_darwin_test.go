//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestNativeShareRecipeWithoutTransfer(t *testing.T) {
	o := app.ShareOptions{AnchorID: "share", Items: []app.ShareItem{{Kind: app.ShareText, Value: "Synthetic <plain> text"}, {Kind: app.ShareURL, Value: "https://example.com"}}}
	b, _ := json.Marshal(o)
	if !nativeShareRecipeFixture(string(b)) {
		t.Fatal("valid recipe")
	}
	for _, wire := range []string{`{}`, `[]`, `{"AnchorID":"share","Items":[{"Kind":"file","Value":"relative"}]}`, `{"AnchorID":"share","Items":[{"Kind":"url","Value":"file:///tmp/a"}]}`, `{"AnchorID":"share","Items":[{"Kind":"unknown","Value":"text"}]}`} {
		if nativeShareRecipeFixture(wire) {
			t.Fatal("invalid recipe", wire)
		}
	}
}
func TestNativeShareCancellationCleanupWithoutTransfer(t *testing.T) {
	for _, selected := range []bool{false, true} {
		_, err := nativeOperationRun(context.Background(), func(token uint64) {
			if !nativeShareCleanupFixture(token, selected) {
				t.Fatal("cleanup or deferred-after-shutdown idempotence failed")
			}
		}, nil)
		if !errors.Is(err, app.ErrCanceled) {
			t.Fatal(err)
		}
	}
}
