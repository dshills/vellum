//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers
#include <stdint.h>
#include <stdlib.h>
void vellumClipboardData(uint64_t token, const void *wire, size_t length, int write);
int vellumClipboardValidate(const void *wire, size_t length);
int vellumClipboardCustomMIMEAvailable(void);
int vellumClipboardAliasValidate(const void *wire, size_t length);
int vellumClipboardPayloadFixture(const void *wire, size_t length);
*/
import "C"

import (
	"context"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeClipboardCapabilities() app.ClipboardCapabilities {
	return app.ClipboardCapabilities{Text: true, PNG: true, HTML: true, CustomMIME: C.vellumClipboardCustomMIMEAvailable() != 0}
}
func nativeClipboardData(ctx context.Context, data []app.ClipboardRepresentation, write bool) ([]byte, error) {
	return clipboardDispatch.run(ctx, func(token uint64) {
		// Encode only after bounded admission. Cocoa copies before this call returns.
		wire := encodeClipboardWire(data)
		flag := C.int(0)
		if write {
			flag = 1
		}
		C.vellumClipboardData(C.uint64_t(token), unsafe.Pointer(&wire[0]), C.size_t(len(wire)), flag)
	})
}

// This helper parses bytes and resolves MIME types only. It never accesses a
// pasteboard and permits offscreen validation of the actual Objective-C parser.
func nativeClipboardValidate(wire []byte) bool {
	if len(wire) == 0 {
		return false
	}
	return C.vellumClipboardValidate(unsafe.Pointer(&wire[0]), C.size_t(len(wire))) != 0
}

func nativeClipboardAliasValidate(wire []byte) bool {
	if len(wire) == 0 {
		return false
	}
	return C.vellumClipboardAliasValidate(unsafe.Pointer(&wire[0]), C.size_t(len(wire))) != 0
}
func nativeClipboardPayloadFixture(wire []byte) int {
	if len(wire) == 0 {
		return 0
	}
	return int(C.vellumClipboardPayloadFixture(unsafe.Pointer(&wire[0]), C.size_t(len(wire))))
}

//export vellumClipboardActive
func vellumClipboardActive(token C.uint64_t) C.int {
	if clipboardDispatch.active(uint64(token)) {
		return 1
	}
	return 0
}

//export vellumClipboardResult
func vellumClipboardResult(token C.uint64_t, wire unsafe.Pointer, length C.size_t, status C.int) {
	var err error
	switch status {
	case 0:
	case 3:
		err = app.ErrClipboardChanged
	case 4:
		err = app.ErrClipboardTooLarge
	default:
		err = app.ErrUnavailable
	}
	var data []byte
	if err == nil && clipboardDispatch.active(uint64(token)) {
		if uint64(length) > clipboardWireMax || (length > 0 && wire == nil) {
			err = app.ErrClipboardTooLarge
		} else if length > 0 {
			data = C.GoBytes(wire, C.int(length))
		}
	}
	clipboardDispatch.finish(uint64(token), data, err)
}
