//go:build darwin && !offscreen

package main

/*
#include <stdlib.h>
#include "display_clock.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"unsafe"
)

func displayConfigure(enabled bool) error {
	flag := C.int(0)
	if enabled {
		flag = 1
	}
	if C.dc_configure(flag) != 1 {
		return errors.New("display clock configuration")
	}
	return nil
}
func displayStart() error {
	if C.dc_start() != 1 {
		return errors.New("display clock start")
	}
	return nil
}
func displayFinish() (map[string]any, error) {
	if C.dc_stop() != 1 {
		return nil, errors.New("display clock stop")
	}
	raw := C.dc_snapshot()
	if raw == nil {
		return nil, errors.New("display clock snapshot")
	}
	defer C.free(unsafe.Pointer(raw))
	var out map[string]any
	if err := json.Unmarshal([]byte(C.GoString(raw)), &out); err != nil {
		return nil, err
	}
	if out["invalid"] != false || out["overflow"] != false {
		return out, errors.New("display clock invalid or overflow")
	}
	return out, nil
}
