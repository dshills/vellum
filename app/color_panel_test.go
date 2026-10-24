package app

import (
	"context"
	"image/color"
	"strings"
	"testing"
)

type colorPanelTestServices struct{ legacyClipboardServices }

func (colorPanelTestServices) ChooseColor(context.Context, WindowRef, ColorPanelOptions) (color.NRGBA, error) {
	return color.NRGBA{R: 12, A: 255}, nil
}
func TestColorPanelOptionsAndOptionalService(t *testing.T) {
	for _, title := range []string{"", " ", "a\x00b", strings.Repeat("x", 257)} {
		if (ColorPanelOptions{Title: title}).Validate() == nil {
			t.Fatal("bad title")
		}
	}
	if err := (ColorPanelOptions{Title: "Colors"}).Validate(); err != nil {
		t.Fatal(err)
	}
	a, _ := New(Config{})
	defer a.Close()
	a.SetServices(legacyClipboardServices{})
	if _, err := a.ColorPanels(); err != ErrUnavailable {
		t.Fatal(err)
	}
	a.SetServices(colorPanelTestServices{})
	if _, err := a.ColorPanels(); err != nil {
		t.Fatal(err)
	}
}
