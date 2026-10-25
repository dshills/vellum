package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
	"os"
	"slices"
)

func (s *platformServices) PreviewFiles(ctx context.Context, parent app.WindowRef, o app.QuickLookOptions) error {
	if err := o.Validate(); err != nil {
		return err
	}
	scope, end, err := s.panelContext(ctx, parent)
	if err != nil {
		return err
	}
	defer end()
	o.Items = slices.Clone(o.Items)
	for _, item := range o.Items {
		if err := nativeServiceContextErr(scope); err != nil {
			return err
		}
		info, err := os.Stat(item.Path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return app.ErrUnavailable
		}
	}
	return nativeQuickLook(scope, s.owner, parent, o)
}
