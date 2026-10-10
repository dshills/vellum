//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include <stdint.h>
void vellumChooseFile(uint64_t token, const char *json);
void vellumCancelFile(uint64_t token);
void vellumSetMenu(uint64_t token, const char *json);
void vellumReadClipboard(uint64_t token);
void vellumWriteClipboard(uint64_t token, const char *text);
void vellumShutdown(void);
*/
import "C"

import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"sync"
	"sync/atomic"
	"unsafe"
)

type nativeResult struct {
	value string
	err   error
}
type nativeMenuCallback struct {
	ids     []string
	service *platformServices
}
type nativeOperation struct {
	ctx    context.Context
	result chan nativeResult
}

var serviceToken atomic.Uint64
var nativeCallbacks = struct {
	sync.Mutex
	operations map[uint64]nativeOperation
	menuToken  uint64
	stopped    bool
	menus      map[uint64]nativeMenuCallback
}{operations: make(map[uint64]nativeOperation), menus: make(map[uint64]nativeMenuCallback)}

// Main-thread dispatch uses Cocoa's asynchronous queue instead of Cogent's
// synchronous RunOnMain, so shutdown cannot strand a sender on a stopped loop.
func nativeOperationRun(ctx context.Context, start func(uint64)) (string, error) {
	if err := nativeServiceContextErr(ctx); err != nil {
		return "", err
	}
	token := serviceToken.Add(1)
	result := make(chan nativeResult, 1)
	nativeCallbacks.Lock()
	if nativeCallbacks.stopped {
		nativeCallbacks.Unlock()
		return "", app.ErrClosed
	}
	nativeCallbacks.operations[token] = nativeOperation{ctx: ctx, result: result}
	nativeCallbacks.Unlock()
	defer func() { nativeCallbacks.Lock(); delete(nativeCallbacks.operations, token); nativeCallbacks.Unlock() }()
	start(token)
	select {
	case value := <-result:
		return value.value, value.err
	case <-ctx.Done():
		C.vellumCancelFile(C.uint64_t(token))
		return "", ctx.Err()
	}
}

//export vellumServiceActive
func vellumServiceActive(token C.uint64_t) C.int {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[uint64(token)]
	if exists && nativeServiceContextErr(op.ctx) == nil {
		return 1
	}
	return 0
}

//export vellumServiceResult
func vellumServiceResult(token C.uint64_t, text *C.char, status C.int) {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[uint64(token)]
	if !exists {
		return
	}
	value := nativeResult{value: C.GoString(text)}
	if status == 1 {
		value.err = app.ErrCanceled
	} else if status == 2 {
		value.err = app.ErrUnavailable
	}
	select {
	case op.result <- value:
	default:
	}
}
func nativeChooseFile(ctx context.Context, o app.FileOptions) (string, error) {
	payload, _ := json.Marshal(o)
	return nativeOperationRun(ctx, func(token uint64) {
		text := C.CString(string(payload))
		defer C.free(unsafe.Pointer(text))
		C.vellumChooseFile(C.uint64_t(token), text)
	})
}
func nativeReadClipboard(ctx context.Context) (string, error) {
	return nativeOperationRun(ctx, func(token uint64) { C.vellumReadClipboard(C.uint64_t(token)) })
}
func nativeWriteClipboard(ctx context.Context, value string) error {
	_, err := nativeOperationRun(ctx, func(token uint64) {
		text := C.CString(value)
		defer C.free(unsafe.Pointer(text))
		C.vellumWriteClipboard(C.uint64_t(token), text)
	})
	return err
}
func nativeSetMenu(ctx context.Context, groups []nativeMenuGroup, service *platformServices) error {
	ids := []string{}
	type item struct {
		Label    string
		Index    int
		Disabled bool
	}
	type group struct {
		Label string
		Items []item
	}
	payload := []group{}
	for _, g := range groups {
		v := group{Label: g.Label, Items: []item{}}
		for _, i := range g.Items {
			v.Items = append(v.Items, item{Label: i.Label, Index: len(ids), Disabled: i.Disabled})
			ids = append(ids, i.Command)
		}
		payload = append(payload, v)
	}
	data, _ := json.Marshal(payload)
	var installedToken uint64
	defer func() { nativeMenuCleanup(installedToken) }()
	_, err := nativeOperationRun(ctx, func(token uint64) {
		installedToken = token
		nativeCallbacks.Lock()
		nativeCallbacks.menus[token] = nativeMenuCallback{ids: ids, service: service}
		nativeCallbacks.Unlock()
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumSetMenu(C.uint64_t(token), text)
	})
	return err
}

// nativeMenuCommit transfers pending callback ownership to the installed menu.
// The operation and its callback must both still exist and be uncanceled. The
// native main thread calls it immediately before replacing NSApp.mainMenu.
// Cancellation cleanup holds the same lock: once accepted, the installed callback
// remains alive even when its caller returns a cancellation error afterward.
func nativeMenuCommit(token uint64) bool {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[token]
	_, callbackExists := nativeCallbacks.menus[token]
	if !exists || !callbackExists || nativeServiceContextErr(op.ctx) != nil {
		return false
	}
	if nativeCallbacks.menuToken != token {
		delete(nativeCallbacks.menus, nativeCallbacks.menuToken)
	}
	nativeCallbacks.menuToken = token
	return true
}
func nativeMenuCleanup(token uint64) {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	if nativeCallbacks.menuToken != token {
		delete(nativeCallbacks.menus, token)
	}
}

//export vellumMenuCommit
func vellumMenuCommit(token C.uint64_t) C.int {
	if nativeMenuCommit(uint64(token)) {
		return 1
	}
	return 0
}

//export vellumMenuAction
func vellumMenuAction(token C.uint64_t, index C.int) { nativeMenuAction(uint64(token), int(index)) }
func nativeMenuAction(token uint64, index int) {
	nativeCallbacks.Lock()
	entry, ok := nativeCallbacks.menus[token]
	if !ok || index < 0 || index >= len(entry.ids) {
		nativeCallbacks.Unlock()
		return
	}
	id, service := entry.ids[index], entry.service
	nativeCallbacks.Unlock()
	// invokeMenu only Posts; commands execute later in the UI-owner Drain.
	if service != nil {
		service.invokeMenu(id)
	}
}

// Pending callers are released even if Cocoa no longer services its main queue.
func stopNativeCallbacks() {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	// The native main loop is single-run per process (Driver.processRun).
	nativeCallbacks.stopped = true
	for _, op := range nativeCallbacks.operations {
		err := nativeServiceContextErr(op.ctx)
		if err == nil {
			err = app.ErrClosed
		}
		select {
		case op.result <- nativeResult{err: err}:
		default:
		}
	}
	nativeCallbacks.operations = make(map[uint64]nativeOperation)
	nativeCallbacks.menus = make(map[uint64]nativeMenuCallback)
	nativeCallbacks.menuToken = 0
}
func nativeShutdown() { stopNativeCallbacks(); C.vellumShutdown() }
