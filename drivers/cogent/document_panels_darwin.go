//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
void vellumChooseDocumentPanel(uint64_t token,uint64_t parent,const char *json);
void vellumCancelDocumentPanel(uint64_t token);
char *vellumDocumentPanelRecipeFixture(const char *json);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"sync"
	"unicode/utf8"
	"unsafe"
)

var documentPanelPreviews = struct {
	sync.Mutex
	events map[uint64]chan string
}{events: map[uint64]chan string{}}

//export vellumDocumentPanelPreview
func vellumDocumentPanelPreview(token C.uint64_t, text *C.char) {
	if text == nil || !nativeServiceActive(uint64(token)) {
		return
	}
	n := C.strnlen(text, 4097)
	if n > 4096 {
		return
	}
	wire := C.GoStringN(text, C.int(n))
	if !utf8.ValidString(wire) {
		return
	}
	deliverDocumentPanelPreview(uint64(token), wire)
}

func deliverDocumentPanelPreview(token uint64, wire string) {
	if len(wire) > 4096 || !utf8.ValidString(wire) || !nativeServiceActive(token) {
		return
	}
	documentPanelPreviews.Lock()
	defer documentPanelPreviews.Unlock()
	queue := documentPanelPreviews.events[token]
	if queue == nil {
		return
	}
	select {
	case queue <- wire:
	default:
		select {
		case <-queue:
		default:
		}
		select {
		case queue <- wire:
		default:
		}
	}
}

// One admitted terminal operation and one latest-preview slot per session.
// No native callback invokes application code; previews run on the worker.
func nativeDocumentPanel(ctx context.Context, owner *app.Application, parent app.WindowRef, payload any, preview func(string)) (string, error) {
	parentToken, err := alertParentToken(owner, parent)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if len(data) > nativeServiceRequestMax {
		return "", app.ErrServiceTooLarge
	}
	return runDocumentPanel(ctx, data, preview, func(token uint64, data []byte) {
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumChooseDocumentPanel(C.uint64_t(token), C.uint64_t(parentToken), text)
	}, func(token uint64) { C.vellumCancelDocumentPanel(C.uint64_t(token)) })
}
func runDocumentPanel(ctx context.Context, data []byte, preview func(string), start func(uint64, []byte), cancel func(uint64)) (string, error) {
	child, end := context.WithCancel(ctx)
	defer end()
	events := make(chan string, 1)
	done := make(chan nativeResult, 1)
	go func() {
		var token uint64
		value, err := nativeOperationRun(alertWindowContext{Context: child, window: ctx}, func(id uint64) {
			token = id
			documentPanelPreviews.Lock()
			documentPanelPreviews.events[id] = events
			documentPanelPreviews.Unlock()
			start(id, data)
		}, cancel)
		documentPanelPreviews.Lock()
		delete(documentPanelPreviews.events, token)
		documentPanelPreviews.Unlock()
		done <- nativeResult{value: value, err: err}
	}()
	for {
		select {
		case result := <-done:
			return result.value, result.err
		case value := <-events:
			if ctx.Err() == nil && child.Err() == nil && preview != nil {
				preview(value)
			}
		}
	}
}
func nativeChooseColorPanel(ctx context.Context, owner *app.Application, parent app.WindowRef, o app.ColorPanelOptions) (string, error) {
	return nativeDocumentPanel(ctx, owner, parent, struct {
		Kind    string
		Options app.ColorPanelOptions
	}{"color", o}, func(wire string) {
		if value, err := decodeColorPanel(wire, o.ShowsAlpha); err == nil && o.Preview != nil {
			o.Preview(value)
		}
	})
}
func nativeDocumentPanelRecipeFixture(wire string) string {
	text := C.CString(wire)
	defer C.free(unsafe.Pointer(text))
	result := C.vellumDocumentPanelRecipeFixture(text)
	if result == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result)
}
