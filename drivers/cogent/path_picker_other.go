//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeChoosePath(context.Context, *app.Application, app.WindowRef, app.PathOptions) (string, error) {
	return "", app.ErrUnavailable
}
func nativeRevealPath(context.Context, string) error { return app.ErrUnavailable }
