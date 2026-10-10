//go:build (!darwin && !windows && (!linux || android)) || !cgo

package cogent

import (
	"context"
	"example.com/vellum/app"
)

func nativeChooseFile(context.Context, app.FileOptions) (string, error) {
	return "", app.ErrUnavailable
}
func nativeSetMenu(context.Context, []nativeMenuGroup, *platformServices) error {
	return app.ErrUnavailable
}
func nativeShutdown() {}

func nativeReadClipboard(context.Context) (string, error) { return "", app.ErrUnavailable }
func nativeWriteClipboard(context.Context, string) error  { return app.ErrUnavailable }
