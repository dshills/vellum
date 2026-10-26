//go:build darwin && cgo

package cogent

/*
#cgo darwin LDFLAGS: -framework PDFKit -framework AVKit -framework AVFoundation -framework CoreMedia
#include <stdint.h>
#include <stdlib.h>
void vellumNativeContentProbe(uint64_t,uint64_t);
int vellumPDFDecodeFixture(const char *,int);
int vellumNativeFrameFixture(void);
int vellumMediaDecodeFixture(const char *,int);
*/
import "C"
import (
	"context"
	"github.com/dshills/vellum/app"
	"unsafe"
)

// InspectNativeContent is a diagnostic of actual embedded AppKit hosts in the
// exact live window. Call from a worker; it does not certify physical input.
func InspectNativeContent(ctx context.Context, a *app.Application, w app.WindowRef) (string, error) {
	parent, e := alertParentToken(a, w)
	if e != nil {
		return "", e
	}
	return nativeOperationRun(ctx, func(token uint64) { C.vellumNativeContentProbe(C.uint64_t(parent), C.uint64_t(token)) }, func(token uint64) { nativeServiceFinish(token, nativeResult{err: app.ErrCanceled}) })
}
func nativePDFDecodeFixture(path string, want int) bool {
	s := C.CString(path)
	defer C.free(unsafe.Pointer(s))
	return C.vellumPDFDecodeFixture(s, C.int(want)) == 1
}
func nativeFrameFixture() bool { return C.vellumNativeFrameFixture() == 1 }

func nativeMediaDecodeFixture(path string, want bool) bool {
	s := C.CString(path)
	defer C.free(unsafe.Pointer(s))
	w := 0
	if want {
		w = 1
	}
	return C.vellumMediaDecodeFixture(s, C.int(w)) == 1
}
