//go:build (windows || (linux && !android)) && cgo

package cogent

import (
	"errors"
	"github.com/go-gl/glfw/v3.4/glfw"
	"testing"
)

func TestClipboardOperationPreservesNativeFailures(t *testing.T) {
	// The platform-error code is intentionally not exported by the pinned Go
	// wrapper; 0x10008 is GLFW_PLATFORM_ERROR in its GLFW public C header.
	platformError := glfw.ErrorCode(0x10008)
	for _, reading := range []bool{true, false} {
		calls := 0
		value, err := nativeClipboardOperation(reading, func() string { return "discard on error" }, func() (glfw.ErrorCode, string) {
			calls++
			if calls == 1 {
				return glfw.NoError, ""
			}
			return platformError, "clipboard backend refused the operation"
		})
		var native *glfw.Error
		if value != "" || !errors.As(err, &native) || native.Code != platformError || native.Desc != "clipboard backend refused the operation" || calls != 2 {
			t.Fatalf("read=%v value=%q error=%v calls=%d", reading, value, err, calls)
		}
	}
}
func TestClipboardOperationClearsStaleErrorBeforeAction(t *testing.T) {
	calls := 0
	value, err := nativeClipboardOperation(true, func() string {
		if calls != 1 {
			t.Fatalf("operation ran before stale error was cleared")
		}
		return "owned text"
	}, func() (glfw.ErrorCode, string) {
		calls++
		if calls == 1 {
			return glfw.PlatformUnavailable, "prior unrelated failure"
		}
		return glfw.NoError, ""
	})
	if err != nil || value != "owned text" || calls != 2 {
		t.Fatalf("%q %v calls=%d", value, err, calls)
	}
}
func TestClipboardMissingTextFlavorIsOnlyAcceptedForRead(t *testing.T) {
	for _, reading := range []bool{true, false} {
		calls := 0
		value, err := nativeClipboardOperation(reading, func() string { return "" }, func() (glfw.ErrorCode, string) {
			calls++
			if calls == 1 {
				return glfw.NoError, ""
			}
			return glfw.FormatUnavailable, "no text flavor"
		})
		if value != "" || (reading && err != nil) || (!reading && err == nil) {
			t.Fatalf("read=%v value=%q error=%v", reading, value, err)
		}
	}
}
