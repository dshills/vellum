//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdint.h>
int vellumBindOpenFiles(uint64_t token);
void vellumUnbindOpenFiles(uint64_t token);
*/
import "C"
import (
	"encoding/json"
	"github.com/dshills/vellum/app"
	"sync/atomic"
	"unsafe"
)

var openFilesToken atomic.Uint64
var currentOpenFiles atomic.Pointer[openFilesReceiver]

//export vellumReceiveOpenFiles
func vellumReceiveOpenFiles(token C.uint64_t, data *C.char, length C.int, initial C.int) C.int {
	if data == nil || length <= 0 || length > 64<<10 || uint64(token) != openFilesToken.Load() {
		return 0
	}
	receiver := currentOpenFiles.Load()
	if receiver == nil {
		return 0
	}
	var paths []string
	if json.Unmarshal(C.GoBytes(unsafe.Pointer(data), length), &paths) != nil {
		return 0
	}
	if !receiver.enqueue(OpenFilesEvent{paths, initial != 0}) {
		return 0
	}
	return 1
}
func configureOpenFiles(owner *app.Application, callback func(OpenFilesEvent)) (*openFilesReceiver, func(), error) {
	if callback == nil {
		return nil, func() {}, nil
	}
	receiver := &openFilesReceiver{owner: owner, callback: callback}
	token := openFilesToken.Add(1)
	currentOpenFiles.Store(receiver)
	if C.vellumBindOpenFiles(C.uint64_t(token)) == 0 {
		receiver.close()
		C.vellumUnbindOpenFiles(C.uint64_t(token))
		currentOpenFiles.Store(nil)
		return nil, func() {}, app.ErrUnavailable
	}
	return receiver, func() {
		receiver.close()
		currentOpenFiles.CompareAndSwap(receiver, nil)
		C.vellumUnbindOpenFiles(C.uint64_t(token))
	}, nil
}
