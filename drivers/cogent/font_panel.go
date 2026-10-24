package cogent

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"io"
)

func (s *platformServices) ChooseFont(ctx context.Context, parent app.WindowRef, o app.FontPanelOptions) (app.FontPanelSelection, error) {
	if err := o.Validate(); err != nil {
		return app.FontPanelSelection{}, err
	}
	scope, end, err := s.panelContext(ctx, parent)
	if err != nil {
		return app.FontPanelSelection{}, err
	}
	defer end()
	wire, err := nativeChooseFontPanel(scope, s.owner, parent, o)
	if err != nil {
		return app.FontPanelSelection{}, err
	}
	return decodeFontPanel(wire)
}
func decodeFontPanel(wire string) (app.FontPanelSelection, error) {
	var v struct {
		Name, Family *string
		Size         *float32
		Bold, Italic *bool
	}
	d := json.NewDecoder(bytes.NewBufferString(wire))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || v.Name == nil || v.Family == nil || v.Size == nil || v.Bold == nil || v.Italic == nil {
		return app.FontPanelSelection{}, app.ErrUnavailable
	}
	var tail any
	if d.Decode(&tail) != io.EOF {
		return app.FontPanelSelection{}, app.ErrUnavailable
	}
	value := app.FontPanelSelection{Name: *v.Name, Family: *v.Family, Size: *v.Size, Bold: *v.Bold, Italic: *v.Italic}
	if value.Validate() != nil {
		return app.FontPanelSelection{}, app.ErrUnavailable
	}
	return value, nil
}
