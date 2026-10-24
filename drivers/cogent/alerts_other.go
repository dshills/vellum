//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeShowAlert(context.Context, *app.Application, app.WindowRef, app.AlertOptions) (app.AlertResult, error) {
	return app.AlertResult{}, app.ErrUnavailable
}
