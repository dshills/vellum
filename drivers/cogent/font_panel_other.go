//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeChooseFontPanel(context.Context, *app.Application, app.WindowRef, app.FontPanelOptions) (string, error) {
	return "", app.ErrUnavailable
}
