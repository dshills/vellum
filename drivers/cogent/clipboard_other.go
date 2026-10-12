//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeClipboardCapabilities() app.ClipboardCapabilities { return app.ClipboardCapabilities{} }
func nativeClipboardData(context.Context, []app.ClipboardRepresentation, bool) ([]byte, error) {
	return nil, app.ErrUnavailable
}
