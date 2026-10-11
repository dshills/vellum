//go:build !darwin || !cgo

package cogent

import (
	"cogentcore.org/core/core"
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
)

func nativeTextAttachment(*nativeWindow) *textWindow { return nil }
func nativeTextScale(*nativeWindow) float64          { return 1 }
func publishTextOnMain(*textWindow, uint64, core.TextFieldEditingState, int, int, bool, uint64, [4]float64, textinput.Geometry, int) (bool, error) {
	return false, nil
}
func settleTextOnMain(ctx context.Context, t *textWindow, client, id uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return t.pump.Send(textMessage{Kind: "settle", Request: id})
}
func InspectTextInput(context.Context, string) (string, error)    { return "", app.ErrUnavailable }
func ProbeTextInput(context.Context, string, int) (string, error) { return "", app.ErrUnavailable }
func QueryTextInput(context.Context, string, int, int, [2]float64) (string, error) {
	return "", app.ErrUnavailable
}

func startNativeTextPublication(t *textWindow) { t.nativeReady.Store(true) }
