//go:build darwin && cgo

package cogent

/*
#include <stdlib.h>
char *vellumNativeWindowSnapshotFixture(void);
*/
import "C"

import "unsafe"

// nativeServiceWindowSnapshot is shared with the opt-in native-check DTO. It
// contains detached sampled values, never native pointers or retained windows.
type nativeServiceWindowSnapshot struct {
	Title                                        string
	Visible, Miniaturized, Key, OcclusionVisible bool
	Frame                                        [4]float64
	// OrderedIndex is the position in NSApp.orderedWindows, or -1 if absent.
	OrderedIndex int
	Opaque       bool
	Alpha        float64
	Level        int
	// IdentityAvailable distinguishes current observations from older records.
	// Cocoa may reuse WindowNumber; class names describe types, not ownership.
	// Neither a zero/negative number nor an empty ContentViewClass is inferred
	// to mean a live, dead, managed, helper or leaked native object.
	IdentityAvailable             bool
	WindowClass, ContentViewClass string
	WindowNumber                  int64
}

func nativeWindowSnapshotFixture() string {
	text := C.vellumNativeWindowSnapshotFixture()
	if text == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text)
}
