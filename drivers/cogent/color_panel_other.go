//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeChooseColorPanel(context.Context, *app.Application, app.WindowRef, app.ColorPanelOptions) (string, error) {
	return "", app.ErrUnavailable
}
