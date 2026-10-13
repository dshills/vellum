//go:build darwin && cgo && !offscreen && vellum_native_checks

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumNativeClipboardCheck(uint64_t token,const char *json);
int vellumClipboardGuardPrivateFixture(const char *path);
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"unsafe"
)

// Acceptance-only configuration. A separate guardian holds the original raw
// clipboard and durable backup; this adapter never receives that content.
type NativeClipboardCheckConfig struct {
	Revision  int64
	Nonce     string
	StatePath string
}
type NativeClipboardCheckState struct {
	MainThread bool
	Configured bool
	Revision   int64
	Phase      string
	Kind       string
	ItemCount  int
}

func nativeClipboardCheck(ctx context.Context, request any) (NativeClipboardCheckState, error) {
	var state NativeClipboardCheckState
	if ctx == nil {
		return state, fmt.Errorf("cogent: native clipboard check needs context")
	}
	data, err := json.Marshal(request)
	if err != nil || len(data) > 32768 {
		return state, fmt.Errorf("cogent: invalid clipboard check request")
	}
	value, err := nativeOperationRun(ctx, func(token uint64) {
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumNativeClipboardCheck(C.uint64_t(token), text)
	}, nil)
	if err == nil {
		err = json.Unmarshal([]byte(value), &state)
	}
	if err == nil && (!state.MainThread || !state.Configured) {
		err = fmt.Errorf("cogent: clipboard check not configured on main thread")
	}
	return state, err
}
func ConfigureNativeClipboardCheck(ctx context.Context, config NativeClipboardCheckConfig) (NativeClipboardCheckState, error) {
	return nativeClipboardCheck(ctx, struct {
		Operation string
		Config    NativeClipboardCheckConfig
	}{"configure", config})
}
func InspectNativeClipboardCheck(ctx context.Context) (NativeClipboardCheckState, error) {
	return nativeClipboardCheck(ctx, struct{ Operation string }{"inspect"})
}

// Seed two owned test items. The first has text/marker; the second alone has PNG.
// Reads still run through the unmodified public clipboard service implementation.
func SeedNativeClipboardCheckItems(ctx context.Context, png []byte) (NativeClipboardCheckState, error) {
	if len(png) == 0 || len(png) > 8192 {
		return NativeClipboardCheckState{}, fmt.Errorf("cogent: clipboard fixture PNG size")
	}
	return nativeClipboardCheck(ctx, struct {
		Operation string
		PNG       []byte
	}{"seed", png})
}

// Main-thread startup fixture; uses only a new private named pasteboard.
func CheckNativeClipboardGuardPrivate(statePath string) error {
	path := C.CString(statePath)
	defer C.free(unsafe.Pointer(path))
	if C.vellumClipboardGuardPrivateFixture(path) != 31 {
		return fmt.Errorf("cogent: native clipboard guard interposition fixture failed")
	}
	return nil
}
