package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
	"os"
	"slices"
)

func (s *platformServices) Share(ctx context.Context, parent app.WindowRef, o app.ShareOptions) (app.ShareResult, error) {
	if err := o.Validate(); err != nil {
		return app.ShareResult{}, err
	}
	scope, end, err := s.panelContext(ctx, parent)
	if err != nil {
		return app.ShareResult{}, err
	}
	defer end()
	o.Items = slices.Clone(o.Items)
	for _, v := range o.Items {
		if err := nativeServiceContextErr(scope); err != nil {
			return app.ShareResult{}, err
		}
		if v.Kind == app.ShareFile {
			info, e := os.Stat(v.Value)
			if e != nil {
				return app.ShareResult{}, e
			}
			if !info.Mode().IsRegular() {
				return app.ShareResult{}, app.ErrUnavailable
			}
		}
	}
	service, err := nativeShare(scope, s.owner, parent, o)
	return app.ShareResult{Service: service}, err
}
