//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func configureAccessibility(*app.Application) error                { return nil }
func windowAccessibilityBridge(*nativeWindow) *accessibilityBridge { return nil }
func nativeAccessibilityUpdate(uint64, string)                     {}
func nativeAccessibilityGeometry(uint64, string)                   {}
func InspectAccessibility(context.Context, string) (AccessibilityInspection, error) {
	return AccessibilityInspection{}, app.ErrUnavailable
}
func PerformAccessibilityAction(context.Context, string, string, string, string) (bool, error) {
	return false, app.ErrUnavailable
}

func releaseAccessibilityConfiguration() {}
