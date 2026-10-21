//go:build darwin && !offscreen

package main

/*
#cgo CFLAGS: -I../..
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore -framework Metal
#include <stdlib.h>
#include "observe.h"
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func observe(phase int) map[string]any {
	s := C.cal_observe(C.int(phase))
	defer C.free(unsafe.Pointer(s))
	var d map[string]any
	if e := json.Unmarshal([]byte(C.GoString(s)), &d); e != nil {
		panic(e)
	}
	return d
}
