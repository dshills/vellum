//go:build darwin && cgo

package inputbridge

/*
#cgo LDFLAGS: -framework AppKit -framework Carbon
#include "input.h"
*/
import "C"
import (
	"errors"
	"math"
	"runtime/cgo"
	"strings"
	"unicode/utf8"
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
func goVellumPreflight(h C.uintptr_t, sequence, clientID C.uint64_t, kind, text, beforeText *C.char, anchor, focus, markedStart, markedEnd C.int) C.bool {
	i := cgo.Handle(h).Value().(*Input)
	if i.pending {
		panic("nested native input admission")
	}
	i.pending = i.admit(Update{Sequence: uint64(sequence), ClientID: uint64(clientID), BeforeText: C.GoString(beforeText), Kind: C.GoString(kind), Text: C.GoString(text), Anchor: int(anchor), Focus: int(focus), MarkedStart: int(markedStart), MarkedEnd: int(markedEnd)})
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

// UpdateGeometry validates an owned snapshot before making one native transaction.
// A rejected/stale publication leaves previously accepted state intact.
func (i *Input) UpdateGeometry(text string, anchor, focus, markedStart, markedEnd int, active bool, ack uint64, rect [4]float64, g Geometry) error {
	_, err := i.UpdateGeometryForClient(0, text, anchor, focus, markedStart, markedEnd, active, ack, rect, g)
	return err
}

// UpdateGeometryForClient preserves the old client until its preedit is committed
// and acknowledged. Updates carry their originating client and prior native text.
func (i *Input) UpdateGeometryForClient(clientID uint64, text string, anchor, focus, markedStart, markedEnd int, active bool, ack uint64, rect [4]float64, g Geometry) (bool, error) {
	const maxCInt = 1<<31 - 1
	for _, v := range []int{anchor, focus, markedStart, markedEnd} {
		if v < -1 || v > maxCInt {
			return false, errors.New("native offset out of range")
		}
	}
	for _, v := range rect {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false, errors.New("nonfinite insertion anchor")
		}
	}
	if rect[2] < 0 || rect[3] < 0 {
		return false, errors.New("invalid insertion anchor")
	}
	if err := g.Validate(text); err != nil {
		return false, err
	}
	cells := make([]C.VellumTextCell, len(g.Cells))
	for j, c := range g.Cells {
		cells[j] = C.VellumTextCell{start: C.int(c.Start), end: C.int(c.End), line: C.int(c.Line), x: C.double(c.Rect[0]), y: C.double(c.Rect[1]), width: C.double(c.Rect[2]), height: C.double(c.Rect[3])}
	}
	carets := make([]C.VellumTextCaret, len(g.Carets))
	for j, c := range g.Carets {
		carets[j] = C.VellumTextCaret{offset: C.int(c.Offset), line: C.int(c.Line), x: C.double(c.Rect[0]), y: C.double(c.Rect[1]), height: C.double(c.Rect[3])}
	}
	var cp *C.VellumTextCell
	var kp *C.VellumTextCaret
	if len(cells) > 0 {
		cp = &cells[0]
	}
	if len(carets) > 0 {
		kp = &carets[0]
	}
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	accepted := C.vellum_input_client_geometry_update(i.native, C.uint64_t(clientID), p, C.int(anchor), C.int(focus), C.int(markedStart), C.int(markedEnd), C.bool(active), C.uint64_t(ack), C.double(rect[0]), C.double(rect[1]), C.double(rect[2]), C.double(rect[3]), cp, C.int(len(cells)), kp, C.int(len(carets)), C.double(g.Viewport[0]), C.double(g.Viewport[1]), C.double(g.Viewport[2]), C.double(g.Viewport[3]), C.int(g.FocusLine))
	return bool(accepted), nil
}

// RenderedRectToScreen is a read-only test oracle. It derives scale from native
// view bounds and framebuffer dimensions, independently of the published DPR.
func (i *Input) RenderedRectToScreen(rect [4]float64, pixels [2]int) [4]float64 {
	r := C.vellum_input_rendered_rect(i.native, C.double(rect[0]), C.double(rect[1]), C.double(rect[2]), C.double(rect[3]), C.int(pixels[0]), C.int(pixels[1]))
	return [4]float64{float64(r.x), float64(r.y), float64(r.width), float64(r.height)}
}

func (i *Input) UpdateClient(clientID uint64, text string, anchor, focus, markedStart, markedEnd int, active bool, ack uint64, rect [4]float64) bool {
	if !utf8.ValidString(text) || strings.ContainsRune(text, 0) || len(text) > 1<<20 {
		return false
	}
	for _, v := range []int{anchor, focus, markedStart, markedEnd} {
		if v < -1 || v > 1<<31-1 {
			return false
		}
	}
	for _, v := range rect {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	if rect[2] < 0 || rect[3] < 0 || math.IsInf(rect[0]+rect[2], 0) || math.IsInf(rect[1]+rect[3], 0) {
		return false
	}
	p := C.CString(text)
	defer C.free(unsafe.Pointer(p))
	return bool(C.vellum_input_client_update(i.native, C.uint64_t(clientID), p, C.int(anchor), C.int(focus), C.int(markedStart), C.int(markedEnd), C.bool(active), C.uint64_t(ack), C.double(rect[0]), C.double(rect[1]), C.double(rect[2]), C.double(rect[3])))
}
