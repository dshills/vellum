//go:build !offscreen

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
	"image"
)

type manualImplementation struct{}

func NewManual(context.Context, *app.Application, ManualConfig) (*Manual, error) {
	return nil, ErrManualUnsupported
}
func (*Manual) Input(context.Context, ManualInput) error { return ErrManualUnsupported }
func (*Manual) Step(context.Context, int) (ManualStepResult, error) {
	return ManualStepResult{}, ErrManualUnsupported
}
func (*Manual) Frame(context.Context, string) (ManualStepResult, error) {
	return ManualStepResult{}, ErrManualUnsupported
}
func (*Manual) Snapshot(context.Context) (ManualSnapshot, error) {
	return ManualSnapshot{}, ErrManualUnsupported
}
func (*Manual) Capture(context.Context, string) (*image.RGBA, error) {
	return nil, ErrManualUnsupported
}
func (*Manual) Close(context.Context) error { return ErrManualUnsupported }

func (*Manual) RequestFrame(context.Context, string) error { return ErrManualUnsupported }
