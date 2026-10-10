//go:build darwin && cgo

package inputbridge

/*
#cgo LDFLAGS: -framework AppKit -framework Carbon
#include "input.h"
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

// Input is attached to one native view. All methods run on Cocoa's main thread.
// Callbacks carry owned values and must enqueue; they cannot touch Go widgets.
type Input struct {
	native   unsafe.Pointer
	handle   cgo.Handle
	callback func(Update)
}

func Attach(view unsafe.Pointer, callback func(Update)) *Input {
	i := &Input{callback: callback}
	i.handle = cgo.NewHandle(i)
	i.native = C.vellum_input_attach(view, C.uintptr_t(i.handle))
	if i.native == nil {
		i.handle.Delete()
		return nil
	}
	return i
}
func (i *Input) Update(text string, anchor, focus, markedStart, markedEnd int, active bool, ack uint64, rect [4]float64) {
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	C.vellum_input_update(i.native, p, C.int(anchor), C.int(focus), C.int(markedStart), C.int(markedEnd), C.bool(active), C.uint64_t(ack), C.double(rect[0]), C.double(rect[1]), C.double(rect[2]), C.double(rect[3]))
}
func (i *Input) Close() {
	if i.native != nil {
		C.vellum_input_close(i.native)
		i.native = nil
		i.handle.Delete()
	}
}
func (i *Input) Probe(step int) string {
	p := C.vellum_input_probe(i.native, C.int(step))
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}
func Source() string {
	p := C.vellum_input_source()
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}

//export goVellumInput
func goVellumInput(h C.uintptr_t, sequence C.uint64_t, kind, text *C.char, anchor, focus, markedStart, markedEnd C.int) {
	i := cgo.Handle(h).Value().(*Input)
	i.callback(Update{Sequence: uint64(sequence), Kind: C.GoString(kind), Text: C.GoString(text), Anchor: int(anchor), Focus: int(focus), MarkedStart: int(markedStart), MarkedEnd: int(markedEnd)})
}

func (i *Input) Caret(rect [4]float64) {
	C.vellum_input_caret(i.native, C.double(rect[0]), C.double(rect[1]), C.double(rect[2]), C.double(rect[3]))
}
func Environment() string {
	p := C.vellum_input_environment()
	if p == nil {
		return "{}"
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}
