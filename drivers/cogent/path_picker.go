package cogent

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"os"
)

func (s *platformServices) ChoosePath(ctx context.Context, w app.WindowRef, o app.PathOptions) (string, error) {
	if e := o.Validate(); e != nil {
		return "", e
	}
	scope, end, e := s.panelContext(ctx, w)
	if e != nil {
		return "", e
	}
	defer end()
	path, e := nativeChoosePath(scope, s.owner, w, o)
	if e != nil {
		return "", e
	}
	if e = ui.ValidatePickerPath(path, o.Kind, o.Extensions); e != nil {
		return "", e
	}
	info, e := os.Stat(path)
	if e != nil {
		return "", e
	}
	if o.Kind == ui.FilePickerFolder && !info.IsDir() || o.Kind == ui.FilePickerFile && !info.Mode().IsRegular() {
		return "", fmt.Errorf("cogent: chosen path kind differs")
	}
	if e = nativeServiceContextErr(scope); e != nil {
		return "", e
	}
	return path, nil
}
func (s *platformServices) RevealPath(ctx context.Context, path string) error {
	if _, e := ui.CopyFileDropPaths([]string{path}); e != nil {
		return e
	}
	if len(path) > 4096 {
		return app.ErrServiceTooLarge
	}
	scope, end, e := s.serviceContext(ctx)
	if e != nil {
		return e
	}
	defer end()
	if _, e = os.Stat(path); e != nil {
		return e
	}
	return nativeRevealPath(scope, path)
}
