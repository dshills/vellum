//go:build darwin && cgo

package inputbridge

/*
#cgo LDFLAGS: -framework AppKit -framework Carbon
#include "input.h"
*/
import "C"
import (
	"math"
	"runtime/cgo"
	"unsafe"
)

// Input is attached to one native view. All methods run on Cocoa's main thread.
// Callbacks carry owned values and must enqueue; they cannot touch Go widgets.
type Input struct {
	native  unsafe.Pointer
	handle  cgo.Handle
	admit   func(Update) bool
	commit  func()
	pending bool
}

func Attach(view unsafe.Pointer, admit func(Update) bool, commit func()) *Input {
	i := &Input{admit: admit, commit: commit}
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

//export goVellumPreflight
func goVellumPreflight(h C.uintptr_t, sequence C.uint64_t, kind, text *C.char, anchor, focus, markedStart, markedEnd C.int) C.bool {
	i := cgo.Handle(h).Value().(*Input)
	if i.pending {
		panic("nested native input admission")
	}
	i.pending = i.admit(Update{Sequence: uint64(sequence), Kind: C.GoString(kind), Text: C.GoString(text), Anchor: int(anchor), Focus: int(focus), MarkedStart: int(markedStart), MarkedEnd: int(markedEnd)})
	return C.bool(i.pending)
}

//export goVellumCommit
func goVellumCommit(h C.uintptr_t) {
	i := cgo.Handle(h).Value().(*Input)
	if !i.pending {
		panic("native commit without admission")
	}
	i.pending = false
	i.commit()
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

// ProbeInactive exercises optional superclass methods on an inactive attached view.
func (i *Input) ProbeInactive() string {
	p := C.vellum_input_inactive_probe(i.native)
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}

// QueryProbe observes read-only Cocoa text queries on its owner thread.
// Coordinates are screen points, ranges are UTF-16 units. Empty means unavailable.
func (i *Input) QueryProbe(start, length int, point [2]float64) string {
	const maxCInt = 1<<31 - 1
	if start < 0 || length < 0 || start > maxCInt || length > maxCInt || math.IsNaN(point[0]) || math.IsNaN(point[1]) || math.IsInf(point[0], 0) || math.IsInf(point[1], 0) {
		return ""
	}
	p := C.vellum_input_query_probe(i.native, C.int(start), C.int(length), C.double(point[0]), C.double(point[1]))
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}
