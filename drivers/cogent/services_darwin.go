//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
void vellumChooseFile(uint64_t token, const char *json);
void vellumCancelFile(uint64_t token);
void vellumSetMenu(uint64_t token, const char *json);
void vellumShutdown(void);
int vellumServiceDismissFixture(void);
*/
import "C"

import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"sync"
	"sync/atomic"
	"unicode/utf8"
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

func nativeServiceDismissFixture() bool { return C.vellumServiceDismissFixture() == 1 }

const nativeServicePendingMax = 8
const nativeServiceRequestMax = 1 << 20
const nativeServiceResultMax = 64 << 10

var nativeCallbacks = struct {
	sync.Mutex
	operations map[uint64]nativeOperation
	menuToken  uint64
	stopped    bool
	menus      map[uint64]nativeMenuCallback
	trayToken  uint64
	trays      map[uint64]nativeMenuCallback
}{operations: make(map[uint64]nativeOperation), menus: make(map[uint64]nativeMenuCallback), trays: make(map[uint64]nativeMenuCallback)}

// Main-thread dispatch uses Cocoa's asynchronous queue instead of Cogent's
// synchronous RunOnMain, so shutdown cannot strand a sender on a stopped loop.
func nativeOperationRun(ctx context.Context, start func(uint64), cancel func(uint64)) (string, error) {
	if err := nativeServiceContextErr(ctx); err != nil {
		return "", err
	}
	result := make(chan nativeResult, 1)
	nativeCallbacks.Lock()
	if nativeCallbacks.stopped {
		nativeCallbacks.Unlock()
		return "", app.ErrClosed
	}
	if len(nativeCallbacks.operations) >= nativeServicePendingMax {
		nativeCallbacks.Unlock()
		return "", app.ErrServiceBusy
	}
	var token uint64
	for {
		previous := serviceToken.Load()
		if previous == ^uint64(0) {
			nativeCallbacks.Unlock()
			return "", app.ErrUnavailable
		}
		if serviceToken.CompareAndSwap(previous, previous+1) {
			token = previous + 1
			break
		}
	}
	nativeCallbacks.operations[token] = nativeOperation{ctx: ctx, result: result}
	nativeCallbacks.Unlock()
	// Caller cancellation retains admission until the native terminal callback.
	start(token)
	select {
	case value := <-result:
		if err := nativeServiceContextErr(ctx); err != nil {
			return "", err
		}
		return value.value, value.err
	case <-ctx.Done():
		if cancel != nil {
			cancel(token)
		}
		return "", ctx.Err()
	}
}

//export vellumServiceActive
func vellumServiceActive(token C.uint64_t) C.int {
	if nativeServiceActive(uint64(token)) {
		return 1
	}
	return 0
}
func nativeServiceActive(token uint64) bool {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[token]
	return exists && !nativeCallbacks.stopped && nativeServiceContextErr(op.ctx) == nil
}

//export vellumServiceResult
func vellumServiceResult(token C.uint64_t, text *C.char, status C.int) {
	nativeCallbacks.Lock()
	op, exists := nativeCallbacks.operations[uint64(token)]
	nativeCallbacks.Unlock()
	if !exists {
		return
	}
	value := nativeResult{}
	if err := nativeServiceContextErr(op.ctx); err != nil {
		nativeServiceFinish(uint64(token), nativeResult{err: err})
		return
	}
	if status == 1 {
		value.err = app.ErrCanceled
	} else if status == 2 {
		value.err = app.ErrUnavailable
	} else if status == 3 {
		value.err = app.ErrServiceTooLarge
	} else if status == 4 {
		value.err = app.ErrNotificationPermission
	} else if status != 0 {
		value.err = app.ErrUnavailable
	} else if text != nil {
		n := C.strnlen(text, C.size_t(nativeServiceResultMax+1))
		if n > nativeServiceResultMax {
			value.err = app.ErrServiceTooLarge
		} else {
			value.value = C.GoStringN(text, C.int(n))
		}
	}
	nativeServiceFinish(uint64(token), value)
}
func nativeServiceFinish(token uint64, value nativeResult) {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	op, exists := nativeCallbacks.operations[token]
	if !exists {
		return
	}
	delete(nativeCallbacks.operations, token)
	if err := nativeServiceContextErr(op.ctx); err != nil {
		value = nativeResult{err: err}
	} else if len(value.value) > nativeServiceResultMax {
		value = nativeResult{err: app.ErrServiceTooLarge}
	} else if !utf8.ValidString(value.value) {
		value = nativeResult{err: app.ErrUnavailable}
	}
	select {
	case op.result <- value:
	default:
	}
}

// Exercise the real bounded C-string reply path without native UI dispatch.
func nativeServiceResultFixture(token uint64, data []byte, status int) {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	vellumServiceResult(C.uint64_t(token), text, C.int(status))
}
func nativeChooseFile(ctx context.Context, o app.FileOptions) (string, error) {
	if len(o.Title) > 1024 || len(o.Directory) > 64<<10 || len(o.Filename) > 64<<10 || len(o.Extensions) > 64 {
		return "", app.ErrServiceTooLarge
	}
	for _, ext := range o.Extensions {
		if len(ext) > 64 {
			return "", app.ErrServiceTooLarge
		}
	}
	return nativeOperationRun(ctx, func(token uint64) {
		payload, _ := json.Marshal(o)
		if len(payload) > nativeServiceRequestMax {
			nativeServiceFinish(token, nativeResult{err: app.ErrServiceTooLarge})
			return
		}
		text := C.CString(string(payload))
		defer C.free(unsafe.Pointer(text))
		C.vellumChooseFile(C.uint64_t(token), text)
	}, func(token uint64) { C.vellumCancelFile(C.uint64_t(token)) })
}
func nativeReadClipboard(ctx context.Context) (string, error) {
	wire, err := nativeClipboardData(ctx, []app.ClipboardRepresentation{{MIME: "text/plain"}}, false)
	if err != nil {
		return "", err
	}
	data, err := decodeClipboardResponse(wire, []string{"text/plain"})
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", nil
	}
	return string(data[0].Data), nil
}
func nativeWriteClipboard(ctx context.Context, value string) error {
	if len(value) > app.ClipboardMaxRepresentationBytes {
		return app.ErrClipboardTooLarge
	}
	_, err := nativeClipboardData(ctx, []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte(value)}}, true)
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
	var installedToken uint64
	defer func() { nativeMenuCleanup(installedToken) }()
	_, err := nativeOperationRun(ctx, func(token uint64) {
		data, _ := json.Marshal(payload)
		if len(data) > nativeServiceRequestMax {
			nativeServiceFinish(token, nativeResult{err: app.ErrServiceTooLarge})
			return
		}
		installedToken = token
		nativeCallbacks.Lock()
		nativeCallbacks.menus[token] = nativeMenuCallback{ids: ids, service: service}
		nativeCallbacks.Unlock()
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumSetMenu(C.uint64_t(token), text)
	}, nil)
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
func nativeMenuCurrent(token uint64, service *platformServices) bool {
	nativeCallbacks.Lock()
	defer nativeCallbacks.Unlock()
	entry, ok := nativeCallbacks.menus[token]
	return ok && !nativeCallbacks.stopped && nativeCallbacks.menuToken == token && entry.service == service
}
func nativeMenuAction(token uint64, index int) {
	nativeCallbacks.Lock()
	entry, ok := nativeCallbacks.menus[token]
	if !ok || nativeCallbacks.stopped || nativeCallbacks.menuToken != token || index < 0 || index >= len(entry.ids) {
		nativeCallbacks.Unlock()
		return
	}
	id, service := entry.ids[index], entry.service
	nativeCallbacks.Unlock()
	// Commands execute later in the UI-owner Drain. Recheck ownership there,
	// since replacement or shutdown can retire the menu after this callback.
	if service != nil {
		service.invokeMenuIf(id, func() bool { return nativeMenuCurrent(token, service) })
	}
}

// Pending callers are released even if Cocoa no longer services its main queue.
func stopNativeCallbacks() {
	clipboardDispatch.stop()
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
	nativeCallbacks.trays = make(map[uint64]nativeMenuCallback)
	nativeCallbacks.trayToken = 0
}
func nativeShutdown() { stopNativeCallbacks(); C.vellumShutdown() }
