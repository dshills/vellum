//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeOpenURL(context.Context, string) error { return app.ErrUnavailable }
