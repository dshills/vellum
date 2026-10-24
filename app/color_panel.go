package app

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"
)

// ColorPanelOptions describes an explicit native color session. Preview is
// optional and runs serially on the calling worker, never on the UI owner.
// Preview deliveries are coalesced; only the returned Apply value is committed.
type ColorPanelOptions struct {
	Title      string
	Color      color.NRGBA
	ShowsAlpha bool
	Preview    func(color.NRGBA) `json:"-"`
}

func (o ColorPanelOptions) Validate() error {
	if strings.TrimSpace(o.Title) == "" || len(o.Title) > 256 || !utf8.ValidString(o.Title) || strings.ContainsRune(o.Title, 0) {
		return fmt.Errorf("vellum: native color panel needs bounded title")
	}
	return nil
}

// ColorPanelServices is optional and worker-only. The exact parent lifetime
// owns the session. Apply returns an sRGB straight-alpha color; Cancel, Escape or
// closing the panel returns ErrCanceled. Closing the parent revokes the request.
// macOS owns one shared panel; a panel created outside this service is refused.
type ColorPanelServices interface {
	ChooseColor(context.Context, WindowRef, ColorPanelOptions) (color.NRGBA, error)
}

func (a *Application) ColorPanels() (ColorPanelServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	p, ok := s.(ColorPanelServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return p, nil
}
