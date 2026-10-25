//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeQuickLook(context.Context, *app.Application, app.WindowRef, app.QuickLookOptions) error {
	return app.ErrUnavailable
}
