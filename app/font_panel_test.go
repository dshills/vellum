package app

import (
	"context"
	"math"
	"testing"
)

type fontPanelTestServices struct{ legacyClipboardServices }

func (fontPanelTestServices) ChooseFont(context.Context, WindowRef, FontPanelOptions) (FontPanelSelection, error) {
	return FontPanelSelection{Name: "HelveticaNeue", Family: "Helvetica Neue", Size: 13}, nil
}
func TestFontPanelBoundsAndOptionalService(t *testing.T) {
	o := FontPanelOptions{Title: "Fonts", Name: "HelveticaNeue", Size: 13}
	if o.Validate() != nil {
		t.Fatal("valid font refused")
	}
	for _, size := range []float32{0, 145, float32(math.NaN()), float32(math.Inf(1))} {
		o.Size = size
		if o.Validate() == nil {
			t.Fatal("invalid size admitted")
		}
	}
	a, _ := New(Config{})
	defer a.Close()
	a.SetServices(legacyClipboardServices{})
	if _, err := a.FontPanels(); err != ErrUnavailable {
		t.Fatal(err)
	}
	a.SetServices(fontPanelTestServices{})
	if _, err := a.FontPanels(); err != nil {
		t.Fatal(err)
	}
}
