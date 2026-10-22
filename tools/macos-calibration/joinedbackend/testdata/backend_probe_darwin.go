//go:build darwin && !offscreen

package main

/*
#include <stdlib.h>
#include <stdbool.h>
extern bool vellumProbeConfigure(bool enabled);
extern size_t vellumProbeSnapshot(unsigned char *out, size_t capacity);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"time"
	"unsafe"
)

func backendConfigure(enabled bool) error {
	if !bool(C.vellumProbeConfigure(C.bool(enabled))) {
		return errors.New("backend trace already configured")
	}
	return nil
}
func backendSnapshot() (map[string]any, error) {
	const capacity = 2 << 20
	buffer := C.malloc(capacity)
	if buffer == nil {
		return nil, errors.New("backend trace buffer allocation")
	}
	defer C.free(buffer)
	deadline := time.Now().Add(2 * time.Second)
	for {
		n := C.vellumProbeSnapshot((*C.uchar)(buffer), capacity)
		if n >= capacity {
			return nil, errors.New("backend trace byte bound")
		}
		var doc map[string]any
		if err := json.Unmarshal(C.GoBytes(unsafe.Pointer(buffer), C.int(n)), &doc); err != nil {
			return nil, err
		}
		if doc["pending"] == float64(0) {
			return doc, nil
		}
		if time.Now().After(deadline) {
			return doc, errors.New("backend trace callbacks did not drain")
		}
		time.Sleep(time.Millisecond)
	}
}
