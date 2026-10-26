//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func InspectNativeContent(context.Context, *app.Application, app.WindowRef) (string, error) {
	return "", app.ErrUnavailable
}
