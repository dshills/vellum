//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include <stdint.h>
char *vellumAXScopedFixture(const char *json);
char *vellumAXPayloadFixture(const char *json,int consume);
char *vellumAXGridFixture(const char *json,const char *tableID,int column,int row);
char *vellumAXTextFixture(const char *node,const char *geometry,int start,int length,int line);
int vellumAXSetValueFixture(uint64_t token,uint64_t generation,const char *node,const char *value);
int vellumAXAttach(uint64_t token, void *view, void *window);
int vellumAXSnapshotValid(const char *json);
char *vellumAXHitFixture(const char *json, const char *bounds, double x, double y);
void vellumAXDetach(uint64_t token);
void vellumAXUpdate(uint64_t token, const char *json);
void vellumAXGeometry(uint64_t token, const char *json);
void vellumAXTextGeometry(uint64_t token,const char *json);
void vellumAXTextProbe(uint64_t token,uint64_t request,const char *id,int start,int length,double x,double y);
void vellumAXProbe(uint64_t token, uint64_t request);
void vellumAXPerform(uint64_t token, uint64_t request, const char *id, const char *action, const char *value);
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
)

type accessibilityAttachment struct {
	bridge *accessibilityBridge
	text   *textWindow
}

func (n *accessibilityAttachment) Focus(bool) {}
func (n *accessibilityAttachment) Close() {
	if n.text != nil && n.text.native != nil {
		n.text.native.Close()
	}
	n.bridge.closeFrom("native-attachment-close")
	C.vellumAXDetach(C.uint64_t(n.bridge.token))
}
func configureAccessibility(a *app.Application) error {
	if desktop.TheApp.NativeAttach != nil {
		return fmt.Errorf("cogent: accessibility requires exclusive native attachment hook")
	}
	desktop.TheApp.NativeAttach = func(w *desktop.Window) (desktop.NativeAttachment, error) {
		b := newAccessibilityBridge(a)
		if C.vellumAXAttach(C.uint64_t(b.token), w.Glw.GetCocoaView(), w.Glw.GetCocoaWindow()) == 0 {
			b.closeFrom("native-attach-failed")
			return nil, fmt.Errorf("cogent: native accessibility attachment failed")
		}
		text, err := attachTextInput(a, w)
		if err != nil {
			b.closeFrom("text-attach-failed")
			C.vellumAXDetach(C.uint64_t(b.token))
			return nil, err
		}
		return &accessibilityAttachment{bridge: b, text: text}, nil
	}
	return nil
}
func windowAccessibilityBridge(w *nativeWindow) *accessibilityBridge {
	if w == nil || w.body == nil || w.body.Scene == nil {
		return nil
	}
	render := w.body.Scene.RenderWindow()
	if render == nil {
		return nil
	}
	native, ok := render.SystemWindow.(*desktop.Window)
	if !ok {
		return nil
	}
	attachment, ok := native.Native.(*accessibilityAttachment)
	if !ok {
		return nil
	}
	return attachment.bridge
}
func nativeAccessibilitySnapshotValid(payload string) bool {
	s := C.CString(payload)
	defer C.free(unsafe.Pointer(s))
	return C.vellumAXSnapshotValid(s) != 0
}
func nativeAccessibilityHitFixture(payload, geometry string, x, y float64) string {
	snapshot, bounds := C.CString(payload), C.CString(geometry)
	defer C.free(unsafe.Pointer(snapshot))
	defer C.free(unsafe.Pointer(bounds))
	result := C.vellumAXHitFixture(snapshot, bounds, C.double(x), C.double(y))
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}

func nativeAccessibilityUpdate(token uint64, payload string) {
	s := C.CString(payload)
	defer C.free(unsafe.Pointer(s))
	C.vellumAXUpdate(C.uint64_t(token), s)
}
func nativeAccessibilityGeometry(token uint64, payload string) {
	s := C.CString(payload)
	defer C.free(unsafe.Pointer(s))
	C.vellumAXGeometry(C.uint64_t(token), s)
}

//export vellumAXAction
func vellumAXAction(token, generation C.uint64_t, id, kind, value *C.char) C.int {
	if accessibilityDeliver(uint64(token), uint64(generation), C.GoString(id), accessibilityAction{Kind: C.GoString(kind), Value: C.GoString(value)}) {
		return 1
	}
	return 0
}

var accessibilityRequest atomic.Uint64
var accessibilityResults = struct {
	sync.Mutex
	requests map[uint64]chan string
}{requests: map[uint64]chan string{}}

//export vellumAXResult
func vellumAXResult(request C.uint64_t, value *C.char) {
	accessibilityResults.Lock()
	result := accessibilityResults.requests[uint64(request)]
	accessibilityResults.Unlock()
	if result != nil {
		select {
		case result <- C.GoString(value):
		default:
		}
	}
}

const accessibilityRequestTimeout = 5 * time.Second

func accessibilityRequestRun(ctx context.Context, b *accessibilityBridge, start func(uint64)) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("cogent: accessibility query requires context")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if b == nil || b.closed.Load() {
		return "", b.closedError()
	}
	// A stopped main queue cannot deliver Objective-C completion. Bound even
	// Background contexts, and let attachment retirement cancel immediately.
	ctx, cancel := context.WithTimeout(ctx, accessibilityRequestTimeout)
	defer cancel()
	request := accessibilityRequest.Add(1)
	result := make(chan string, 1)
	accessibilityResults.Lock()
	accessibilityResults.requests[request] = result
	accessibilityResults.Unlock()
	defer func() {
		accessibilityResults.Lock()
		delete(accessibilityResults.requests, request)
		accessibilityResults.Unlock()
	}()
	start(request)
	select {
	case value := <-result:
		if b.closed.Load() {
			return "", b.closedError()
		}
		return value, nil
	case <-b.done:
		return "", b.closedError()
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// InspectAccessibility asynchronously queries the actual Cocoa view hierarchy.
// Call from a worker; waiting on the UI/native owner would block publication.
// Queries/actions have a five-second maximum wait and cancel on attachment close.
func InspectAccessibility(ctx context.Context, window string) (AccessibilityInspection, error) {
	b, err := lookupAccessibility(window)
	if err != nil {
		return AccessibilityInspection{}, err
	}
	value, err := accessibilityRequestRun(ctx, b, func(request uint64) { C.vellumAXProbe(C.uint64_t(b.token), C.uint64_t(request)) })
	if err != nil {
		return AccessibilityInspection{}, err
	}
	var out AccessibilityInspection
	err = json.Unmarshal([]byte(value), &out)
	return out, err
}

// PerformAccessibilityAction exercises a native NSAccessibility action method,
// returning whether it was admitted to the owner queue, not whether the model
// accepted its value. Inspect after an owner refresh to observe accepted state.
func PerformAccessibilityAction(ctx context.Context, window, id, kind, value string) (bool, error) {
	b, err := lookupAccessibility(window)
	if err != nil {
		return false, err
	}
	b.mu.Lock()
	if b.handlers[id] == nil {
		id = accessibilityManagedID(id)
	}
	b.mu.Unlock()
	result, err := accessibilityRequestRun(ctx, b, func(request uint64) {
		i, k, v := C.CString(id), C.CString(kind), C.CString(value)
		defer C.free(unsafe.Pointer(i))
		defer C.free(unsafe.Pointer(k))
		defer C.free(unsafe.Pointer(v))
		C.vellumAXPerform(C.uint64_t(b.token), C.uint64_t(request), i, k, v)
	})
	return result == "true", err
}

func releaseAccessibilityConfiguration() { desktop.TheApp.NativeAttach = nil }

func nativeAccessibilityTextGeometry(token uint64, payload string) {
	s := C.CString(payload)
	defer C.free(unsafe.Pointer(s))
	C.vellumAXTextGeometry(C.uint64_t(token), s)
}

// QueryAccessibilityText probes NSAccessibility text selectors in process. It
// does not certify external AX or screen-reader workflows. Call from a worker.
func QueryAccessibilityText(ctx context.Context, window, id string, start, length int, point [2]float64) (string, error) {
	if ctx == nil || start < 0 || length < 0 || start > math.MaxInt32 || length > math.MaxInt32 || math.IsNaN(point[0]) || math.IsNaN(point[1]) || math.IsInf(point[0], 0) || math.IsInf(point[1], 0) {
		return "", fmt.Errorf("cogent: invalid accessibility text query")
	}
	b, err := lookupAccessibility(window)
	if err != nil {
		return "", err
	}
	b.mu.Lock()
	if b.handlers[id] == nil {
		id = accessibilityManagedID(id)
	}
	b.mu.Unlock()
	return accessibilityRequestRun(ctx, b, func(request uint64) {
		i := C.CString(id)
		defer C.free(unsafe.Pointer(i))
		C.vellumAXTextProbe(C.uint64_t(b.token), C.uint64_t(request), i, C.int(start), C.int(length), C.double(point[0]), C.double(point[1]))
	})
}

func nativeAccessibilityTextFixture(node, geometry string, start, length, line int) string {
	n, g := C.CString(node), C.CString(geometry)
	defer C.free(unsafe.Pointer(n))
	defer C.free(unsafe.Pointer(g))
	result := C.vellumAXTextFixture(n, g, C.int(start), C.int(length), C.int(line))
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}

func nativeAccessibilitySetValueFixture(token, generation uint64, node, value string) bool {
	n, v := C.CString(node), C.CString(value)
	defer C.free(unsafe.Pointer(n))
	defer C.free(unsafe.Pointer(v))
	return C.vellumAXSetValueFixture(C.uint64_t(token), C.uint64_t(generation), n, v) != 0
}

func nativeAccessibilityGridFixture(payload, table string, column, row int) string {
	p, t := C.CString(payload), C.CString(table)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(t))
	result := C.vellumAXGridFixture(p, t, C.int(column), C.int(row))
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}

// nativeAccessibilityScopedFixture runs only the inert production transaction
// fixture; it has no native attachment/main-queue or successful action callback.
func nativeAccessibilityScopedFixture(steps string) string {
	s := C.CString(steps)
	defer C.free(unsafe.Pointer(s))
	result := C.vellumAXScopedFixture(s)
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}

// nativeAccessibilityPayloadFixture tests the real JSON producer without an
// attachment, queue scheduling, or OS settings changes.
func nativeAccessibilityPayloadFixture(payload string, consume bool) string {
	s := C.CString(payload)
	defer C.free(unsafe.Pointer(s))
	delivery := C.int(0)
	if consume {
		delivery = 1
	}
	result := C.vellumAXPayloadFixture(s, delivery)
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}
