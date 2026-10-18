//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
char *vellumWindowVisibilityObserverFixture(void);
*/
import "C"
import "unsafe"

//export vellumWindowVisibility
func vellumWindowVisibility(token C.uint64_t, visible C.int) {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.bridges[uint64(token)]
	accessibilityRegistry.Unlock()
	if b != nil {
		b.visibility.observe(visible != 0)
	}
}

func nativeVisibilityObserverFixture() string {
	p := C.vellumWindowVisibilityObserverFixture()
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}
