//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeTrayCapabilities() app.TrayCapabilities { return app.TrayCapabilities{} }
func nativeSetTray(context.Context, nativeTray, []byte, []string, *platformServices) error {
	return app.ErrUnavailable
}
