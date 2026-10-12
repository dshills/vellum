package app

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestClipboardValidation(t *testing.T) {
	for _, bad := range []string{"", "text", "text/", "/plain", "Text/plain", "text/plain;charset=utf-8", "text/*", " text/plain", "text/plain\x00", "tèxt/plain", "text/plain/extra", "application/" + strings.Repeat("a", 128)} {
		if ValidateClipboardFormats([]string{bad}) == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	if ValidateClipboardFormats([]string{"text/plain", "text/plain"}) == nil {
		t.Fatal("duplicate accepted")
	}
	if ValidateClipboardFormats(nil) == nil {
		t.Fatal("empty request accepted")
	}
	if err := ValidateClipboardData([]ClipboardRepresentation{{MIME: "application/vnd.vellum+json", Data: []byte{0, 255}}, {MIME: "text/plain", Data: []byte("Résumé 😀")}}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{{255}, {'a', 0, 'b'}} {
		if ValidateClipboardData([]ClipboardRepresentation{{MIME: "text/plain", Data: bad}}) == nil {
			t.Fatal("invalid text accepted")
		}
	}
	if err := ValidateClipboardData([]ClipboardRepresentation{{MIME: "image/png", Data: make([]byte, ClipboardMaxRepresentationBytes+1)}}); !errors.Is(err, ErrClipboardTooLarge) {
		t.Fatal(err)
	}
	large := make([]byte, ClipboardMaxRepresentationBytes)
	if err := ValidateClipboardData([]ClipboardRepresentation{{MIME: "a/one", Data: large}, {MIME: "a/two", Data: large}}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateClipboardData([]ClipboardRepresentation{{MIME: "a/one", Data: large}, {MIME: "a/two", Data: large}, {MIME: "a/three", Data: []byte{1}}}); !errors.Is(err, ErrClipboardTooLarge) {
		t.Fatal(err)
	}
}

type legacyClipboardServices struct{}

func (legacyClipboardServices) ReadClipboard(context.Context) (string, error) { return "", nil }
func (legacyClipboardServices) WriteClipboard(context.Context, string) error  { return nil }
func (legacyClipboardServices) ChooseFile(context.Context, FileOptions) (string, error) {
	return "", nil
}
func (legacyClipboardServices) SetMenu(context.Context, []MenuGroup) error { return nil }

type extendedClipboardServices struct{ legacyClipboardServices }

func (extendedClipboardServices) ClipboardCapabilities() ClipboardCapabilities {
	return ClipboardCapabilities{Text: true}
}
func (extendedClipboardServices) ReadClipboardData(context.Context, []string) ([]ClipboardRepresentation, error) {
	return nil, nil
}
func (extendedClipboardServices) WriteClipboardData(context.Context, []ClipboardRepresentation) error {
	return nil
}
func TestOptionalClipboardCompatibility(t *testing.T) {
	a, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if _, err = a.Clipboard(); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	a.SetServices(legacyClipboardServices{})
	if _, err = a.Services(); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Clipboard(); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	a.SetServices(extendedClipboardServices{})
	c, err := a.Clipboard()
	if err != nil || !c.ClipboardCapabilities().Text {
		t.Fatal(err)
	}
	a.Close()
	if _, err = a.Clipboard(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
