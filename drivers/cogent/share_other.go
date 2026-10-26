//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeShare(context.Context, *app.Application, app.WindowRef, app.ShareOptions) (string, error) {
	return "", app.ErrUnavailable
}
