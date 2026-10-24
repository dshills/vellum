package cogent

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"image/color"
	"io"
)

func (s *platformServices) panelContext(ctx context.Context, parent app.WindowRef) (context.Context, func(), error) {
	scope, end, err := s.serviceContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	owned := false
	for _, w := range s.owner.Windows() {
		owned = owned || w.Ref() == parent
	}
	if !owned || parent == (app.WindowRef{}) {
		end()
		return nil, nil, app.ErrInvalidWindowRef
	}
	if parent.Context().Err() != nil {
		end()
		return nil, nil, app.ErrWindowClosed
	}
	joined, cancel := context.WithCancel(scope)
	stop := context.AfterFunc(parent.Context(), cancel)
	return alertWindowContext{Context: joined, window: parent.Context()}, func() { stop(); cancel(); end() }, nil
}
func (s *platformServices) ChooseColor(ctx context.Context, parent app.WindowRef, o app.ColorPanelOptions) (color.NRGBA, error) {
	if err := o.Validate(); err != nil {
		return color.NRGBA{}, err
	}
	scope, end, err := s.panelContext(ctx, parent)
	if err != nil {
		return color.NRGBA{}, err
	}
	defer end()
	if !o.ShowsAlpha {
		o.Color.A = 255
	}
	wire, err := nativeChooseColorPanel(scope, s.owner, parent, o)
	if err != nil {
		return color.NRGBA{}, err
	}
	return decodeColorPanel(wire, o.ShowsAlpha)
}
func decodeColorPanel(wire string, alpha bool) (color.NRGBA, error) {
	var v struct{ R, G, B, A *uint8 }
	d := json.NewDecoder(bytes.NewBufferString(wire))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || v.R == nil || v.G == nil || v.B == nil || v.A == nil {
		return color.NRGBA{}, app.ErrUnavailable
	}
	var tail any
	if d.Decode(&tail) != io.EOF || !alpha && *v.A != 255 {
		return color.NRGBA{}, app.ErrUnavailable
	}
	return color.NRGBA{R: *v.R, G: *v.G, B: *v.B, A: *v.A}, nil
}
