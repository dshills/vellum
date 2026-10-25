//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
int vellumTitlebarAttach(uint64_t token,void *window,const char *json,const char *identifier);
void vellumTitlebarDetach(uint64_t token);
void vellumTitlebarAllowed(uint64_t token,int allowed);
void vellumTitlebarCustomize(uint64_t token);
int vellumTitlebarCustomizationSupported(void);
void vellumTitlebarRequestAttach(uint64_t,const char *,const char *);
void vellumTitlebarStateChanged(uint64_t);
*/
import "C"
import (
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/desktop"
	"fmt"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func validateTitlebarToolbarSupport(p *preparedTitlebarToolbar) error {
	if p != nil && (system.TheApp == nil || system.TheApp.Platform() != system.MacOS) {
		return fmt.Errorf("cogent: titlebar prototype requires the macOS desktop driver")
	}
	if p != nil && p.customizationID != "" && C.vellumTitlebarCustomizationSupported() == 0 {
		return fmt.Errorf("cogent: native titlebar customization requires macOS 15 or newer")
	}
	return nil
}
func attachTitlebarToolbar(a *app.Application, p *preparedTitlebarToolbar, b *accessibilityBridge, w *desktop.Window) error {
	if p == nil {
		return nil
	}
	binding := newTitlebarToolbarBinding(a, p)
	binding.token = b.token
	b.titlebar.Store(binding)
	data := C.CString(p.wire)
	defer C.free(unsafe.Pointer(data))
	identifier := C.CString(p.customizationID)
	defer C.free(unsafe.Pointer(identifier))
	if C.vellumTitlebarAttach(C.uint64_t(b.token), w.Glw.GetCocoaWindow(), data, identifier) == 0 {
		binding.close()
		C.vellumTitlebarDetach(C.uint64_t(b.token))
		b.titlebar.Store(nil)
		return fmt.Errorf("cogent: native titlebar toolbar attachment failed")
	}
	return nil
}
func closeTitlebarToolbar(b *accessibilityBridge) {
	if b == nil {
		return
	}
	if binding := b.titlebar.Swap(nil); binding != nil {
		binding.close()
		C.vellumTitlebarDetach(C.uint64_t(b.token))
	}
}
func titlebarToolbarAllowed(b *accessibilityBridge, allowed bool) {
	if binding := b.titlebar.Load(); binding != nil && binding.setAllowed(allowed) {
		value := C.int(0)
		if allowed {
			value = 1
		}
		C.vellumTitlebarAllowed(C.uint64_t(b.token), value)
	}
}

//export vellumTitlebarAction
func vellumTitlebarAction(token C.uint64_t, id *C.char) C.int {
	if deliverTitlebarToolbarAction(uint64(token), C.GoString(id)) {
		return 1
	}
	return 0
}

func nativeTitlebarCustomize(token uint64) { C.vellumTitlebarCustomize(C.uint64_t(token)) }

//export vellumTitlebarCustomizationScope
func vellumTitlebarCustomizationScope(token C.uint64_t, customizing C.int) C.int {
	if titlebarBindingForToken(uint64(token)).setCustomizing(customizing != 0) {
		return 1
	}
	return 0
}

//export vellumTitlebarItemsChanged
func vellumTitlebarItemsChanged(token C.uint64_t) {
	titlebarBindingForToken(uint64(token)).itemsChanged()
}

func nativeTitlebarRequestAttach(token uint64, wire, identifier string) {
	data, id := C.CString(wire), C.CString(identifier)
	defer C.free(unsafe.Pointer(data))
	defer C.free(unsafe.Pointer(id))
	C.vellumTitlebarRequestAttach(C.uint64_t(token), data, id)
}
func nativeTitlebarStateChanged(token uint64) { C.vellumTitlebarStateChanged(C.uint64_t(token)) }

//export vellumTitlebarIsLive
func vellumTitlebarIsLive(token C.uint64_t) C.int {
	b := titlebarBindingForToken(uint64(token))
	if b != nil && !b.closed.Load() && b.config.ref.Context().Err() == nil {
		return 1
	}
	return 0
}

//export vellumTitlebarAttached
func vellumTitlebarAttached(token C.uint64_t, success C.int) {
	b := titlebarBindingForToken(uint64(token))
	if b == nil || b.closed.Load() {
		return
	}
	if success == 0 {
		b.close()
		b.attachFailed.Store(true)
	}
	// Invalidation is thread-safe and coalesced. A full owner action queue must
	// not lose the attachment acknowledgement or leave the toolbar disabled.
	if b.config.ref.Context().Err() == nil {
		_ = b.owner.Invalidate(b.config.ref.ID())
	}
}

//export vellumTitlebarStateJSON
func vellumTitlebarStateJSON(token C.uint64_t) *C.char {
	b := titlebarBindingForToken(uint64(token))
	if b == nil || b.closed.Load() {
		return nil
	}
	state := b.takeAvailability()
	if state == nil {
		return nil
	}
	return C.CString(state.json)
}
