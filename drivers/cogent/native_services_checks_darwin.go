//go:build darwin && cgo && !offscreen && vellum_native_checks

package cogent

/*
#cgo CFLAGS: -DVELLUM_NATIVE_CHECKS
#include <stdint.h>
#include <stdlib.h>
void vellumNativeServicesCheck(uint64_t token, const char *json);
char *vellumNativeServicesCleanupCheck(void);
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"unsafe"

	"cogentcore.org/core/system/driver/desktop"
)

// These diagnostics are deliberately absent from ordinary SDK builds. They
// inspect only this adapter's owned resources; they are not a UI automation API.
type NativeServiceItemCheck struct {
	Label     string `json:"label"`
	Index     int    `json:"index"`
	Enabled   bool   `json:"enabled"`
	Separator bool   `json:"separator"`
}
type NativeServiceMenuCheck struct {
	Owned, OriginalRestored, TargetRetired bool
	Items                                  []NativeServiceItemCheck
}
type NativeServiceTrayCheck struct {
	Present, Visible, Template, TargetRetired bool
	Identity                                  uint64
	Title, Tooltip, AccessibilityLabel        string
	ImageWidth, ImageHeight                   float64
	PixelWidth, PixelHeight                   int
	Items                                     []NativeServiceItemCheck
}
type NativeServiceFilePanelCheck struct {
	Tracked, Visible, Open, Save int
}

type NativeServiceCheckState struct {
	MainThread      bool
	CleanupObserved bool
	WindowTitles    []string // Visible NSWindows owned by this process.
	FilePanels      NativeServiceFilePanelCheck
	// NativeOutstanding is sampled after this query ACK, separately from the
	// main-thread Cocoa snapshot; it is not an atomic cross-thread inventory.
	NativeOutstanding            int
	Menu                         NativeServiceMenuCheck
	Tray                         NativeServiceTrayCheck
	CallbacksStopped             bool
	PendingCallbacks             int
	MenuCallbacks, TrayCallbacks int
}

const (
	NativeCheckSnapshot = iota
	NativeCheckCaptureOriginalMenu
	NativeCheckMenuAction
	NativeCheckCaptureStaleMenu
	NativeCheckStaleMenuAction
	NativeCheckTrayAction
	NativeCheckCaptureStaleTray
	NativeCheckStaleTrayAction
)

func CheckNativeServices(ctx context.Context, operation, index int) (NativeServiceCheckState, error) {
	var state NativeServiceCheckState
	if ctx == nil || operation < NativeCheckSnapshot || operation > NativeCheckStaleTrayAction || index < 0 || index >= 128 {
		return state, fmt.Errorf("cogent: invalid native check operation")
	}
	value, err := nativeOperationRun(ctx, func(token uint64) {
		data, _ := json.Marshal(struct{ Operation, Index int }{operation, index})
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumNativeServicesCheck(C.uint64_t(token), text)
	}, nil)
	if err == nil {
		err = json.Unmarshal([]byte(value), &state)
	}
	if err == nil && !state.MainThread {
		err = fmt.Errorf("cogent: native check did not run on main thread")
	}
	if err == nil {
		nativeCallbacks.Lock()
		state.NativeOutstanding = len(nativeCallbacks.operations)
		nativeCallbacks.Unlock()
	}
	return state, err
}

var nativeServiceCleanupCheck atomic.Pointer[NativeServiceCheckState]

// Register before Driver.Run. Reverse hook order makes this observer run after
// production service retirement, while AppKit is still valid. It never queues
// work or enters the UI owner. The caller can remove it if Run never starts.
func ObserveNativeServiceCleanup() (func(), error) {
	return desktop.TheApp.AddMainCleanup(func() {
		text := C.vellumNativeServicesCleanupCheck()
		defer C.free(unsafe.Pointer(text))
		state := NativeServiceCheckState{}
		if text != nil {
			_ = json.Unmarshal([]byte(C.GoString(text)), &state)
		}
		nativeCallbacks.Lock()
		state.CallbacksStopped = nativeCallbacks.stopped
		state.PendingCallbacks = len(nativeCallbacks.operations)
		state.NativeOutstanding = state.PendingCallbacks
		state.MenuCallbacks = len(nativeCallbacks.menus)
		state.TrayCallbacks = len(nativeCallbacks.trays)
		nativeCallbacks.Unlock()
		nativeServiceCleanupCheck.Store(&state)
	})
}

func NativeServiceCleanupObservation() (NativeServiceCheckState, bool) {
	state := nativeServiceCleanupCheck.Load()
	if state == nil {
		return NativeServiceCheckState{}, false
	}
	return *state, true
}
