//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdint.h>
int vellumBindQuit(uint64_t token);
void vellumUnbindQuit(uint64_t token);
int vellumTestQuitDelegate(void);
*/
import "C"
import (
	"github.com/dshills/vellum/app"
	"sync/atomic"
)

var nativeQuitToken atomic.Uint64
var currentNativeQuit atomic.Pointer[quitCallback]

//export vellumNativeQuitCallback
func vellumNativeQuitCallback(token C.uint64_t) {
	if uint64(token) != nativeQuitToken.Load() {
		return
	}
	if callback := currentNativeQuit.Load(); callback != nil {
		callback.request()
	}
}
func bindNativeQuit(callback *quitCallback) (func(), error) {
	token := nativeQuitToken.Add(1)
	if token == 0 {
		return nil, app.ErrUnavailable
	}
	currentNativeQuit.Store(callback)
	if C.vellumBindQuit(C.uint64_t(token)) == 0 {
		currentNativeQuit.CompareAndSwap(callback, nil)
		return nil, app.ErrUnavailable
	}
	return func() {
		currentNativeQuit.CompareAndSwap(callback, nil)
		C.vellumUnbindQuit(C.uint64_t(token))
	}, nil
}
func testNativeQuitDelegate() bool { return C.vellumTestQuitDelegate() != 0 }
