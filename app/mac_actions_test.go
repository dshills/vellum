package app

import (
	"context"
	"errors"
	"testing"
)

type macActionTestServices struct{ legacyClipboardServices }

func (macActionTestServices) OpenURL(context.Context, string) error { return nil }
func (macActionTestServices) Share(context.Context, WindowRef, ShareOptions) (ShareResult, error) {
	return ShareResult{Service: "Synthetic"}, nil
}
func TestOptionalMacActionServicesPreserveLegacyAndLifetime(t *testing.T) {
	a, e := New(Config{})
	if e != nil {
		t.Fatal(e)
	}
	a.SetServices(legacyClipboardServices{})
	if _, e = a.URLs(); !errors.Is(e, ErrUnavailable) {
		t.Fatal(e)
	}
	if _, e = a.Sharing(); !errors.Is(e, ErrUnavailable) {
		t.Fatal(e)
	}
	a.SetServices(macActionTestServices{})
	if _, e = a.URLs(); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Sharing(); e != nil {
		t.Fatal(e)
	}
	a.Close()
	if _, e = a.URLs(); !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
	if _, e = a.Sharing(); !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
}
