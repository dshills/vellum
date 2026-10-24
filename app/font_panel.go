package app

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

type FontPanelSelection struct {
	Name, Family string
	Size         float32
	Bold, Italic bool
}

func (v FontPanelSelection) Validate() error {
	for _, text := range []string{v.Name, v.Family} {
		if strings.TrimSpace(text) == "" || len(text) > 256 || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
			return fmt.Errorf("vellum: invalid native font identity")
		}
	}
	if math.IsNaN(float64(v.Size)) || math.IsInf(float64(v.Size), 0) || v.Size < 6 || v.Size > 144 {
		return fmt.Errorf("vellum: native font size must be 6..144 points")
	}
	return nil
}

// FontPanelOptions names an installed PostScript font. Preview runs serially on
// the calling worker and coalesces. Only Use Font returns an accepted selection.
type FontPanelOptions struct {
	Title, Name string
	Size        float32
	Preview     func(FontPanelSelection) `json:"-"`
}

func (o FontPanelOptions) Validate() error {
	if len(o.Title) > 256 || strings.TrimSpace(o.Title) == "" || !utf8.ValidString(o.Title) || strings.ContainsRune(o.Title, 0) {
		return fmt.Errorf("vellum: native font panel needs bounded title")
	}
	return (FontPanelSelection{Name: o.Name, Family: o.Name, Size: o.Size}).Validate()
}

type FontPanelServices interface {
	ChooseFont(context.Context, WindowRef, FontPanelOptions) (FontPanelSelection, error)
}

func (a *Application) FontPanels() (FontPanelServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	p, ok := s.(FontPanelServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return p, nil
}
