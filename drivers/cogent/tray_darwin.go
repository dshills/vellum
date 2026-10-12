//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include <stdint.h>
void vellumSetTray(uint64_t token, const char *json);
int vellumTrayValidate(const char *json);
int vellumServicePayloadFixture(const char *json);
int vellumTrayImageFixture(const char *json);
*/
import "C"

import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeTrayCapabilities() app.TrayCapabilities {
	return app.TrayCapabilities{StatusItem: true, PNGIcon: true, CommandMenu: true}
}
func nativeSetTray(ctx context.Context, desc nativeTray, png []byte, ids []string, s *platformServices) error {
	var installedToken uint64
	defer func() { nativeTrayCleanup(installedToken) }()
	_, err := nativeOperationRun(ctx, func(token uint64) {
		if err := trayPixels(&desc, png); err != nil {
			nativeServiceFinish(token, nativeResult{err: err})
			return
		}
		data, _ := json.Marshal(desc)
		if len(data) > nativeServiceRequestMax {
			nativeServiceFinish(token, nativeResult{err: app.ErrServiceTooLarge})
			return
		}
		installedToken = token
		nativeCallbacks.Lock()
		nativeCallbacks.trays[token] = nativeMenuCallback{ids: ids, service: s}
		nativeCallbacks.Unlock()
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumSetTray(C.uint64_t(token), text)
	}, nil)
	return err
}
func nativeTrayCommit(token uint64, remove bool) bool {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[token]
	_, callbackExists := nativeCallbacks.trays[token]
	if nativeCallbacks.stopped || !exists || !callbackExists || nativeServiceContextErr(op.ctx) != nil {
		return false
	}
	delete(nativeCallbacks.trays, nativeCallbacks.trayToken)
	if remove {
		delete(nativeCallbacks.trays, token)
		nativeCallbacks.trayToken = 0
	} else {
		nativeCallbacks.trayToken = token
	}
	return true
}
func nativeTrayCleanup(token uint64) {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	if nativeCallbacks.trayToken != token {
		delete(nativeCallbacks.trays, token)
	}
}
func nativeTrayCurrent(token uint64, s *platformServices) bool {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	entry, ok := nativeCallbacks.trays[token]
	return ok && !nativeCallbacks.stopped && nativeCallbacks.trayToken == token && entry.service == s && !s.stopped.Load()
}
func nativeTrayAction(token uint64, index int) {
	nativeCallbacks.Lock()
	entry, ok := nativeCallbacks.trays[token]
	if !ok || nativeCallbacks.trayToken != token || index < 0 || index >= len(entry.ids) {
		nativeCallbacks.Unlock()
		return
	}
	id, s := entry.ids[index], entry.service
	nativeCallbacks.Unlock()
	if s != nil {
		s.invokeMenuIf(id, func() bool { return nativeTrayCurrent(token, s) })
	}
}

//export vellumTrayCommit
func vellumTrayCommit(token C.uint64_t, remove C.int) C.int {
	if nativeTrayCommit(uint64(token), remove != 0) {
		return 1
	}
	return 0
}

//export vellumTrayAction
func vellumTrayAction(token C.uint64_t, index C.int) { nativeTrayAction(uint64(token), int(index)) }

// Pure decoder/ownership fixtures. Neither helper accesses NSStatusBar or NSMenu.
func nativeTrayValidate(data []byte) bool {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	return C.vellumTrayValidate(text) != 0
}
func nativeServicePayloadFixture(data []byte) int {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	return int(C.vellumServicePayloadFixture(text))
}
func nativeTrayImageFixture(data []byte) int {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	return int(C.vellumTrayImageFixture(text))
}
