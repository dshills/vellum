//go:build darwin && !offscreen

package main

/*
#include <stdlib.h>
#include <stdbool.h>
extern bool vellumUploadConfigure(bool enabled);
extern size_t vellumUploadSnapshot(unsigned char *out, size_t capacity);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"time"
	"unsafe"
)

func probeConfigure(enabled bool) error {
	if !bool(C.vellumUploadConfigure(C.bool(enabled))) {
		return errors.New("upload trace already configured")
	}
	return nil
}
func probeSnapshot() (map[string]any, error) {
	const capacity = 2 << 20
	buffer := C.malloc(capacity)
	if buffer == nil {
		return nil, errors.New("upload trace buffer allocation")
	}
	defer C.free(buffer)
	deadline := time.Now().Add(2 * time.Second)
	for {
		n := C.vellumUploadSnapshot((*C.uchar)(buffer), capacity)
		if n >= capacity {
			return nil, errors.New("upload trace byte bound")
		}
		var doc map[string]any
		if err := json.Unmarshal(C.GoBytes(unsafe.Pointer(buffer), C.int(n)), &doc); err != nil {
			return nil, err
		}
		if doc["pending"] == float64(0) {
			return doc, nil
		}
		if time.Now().After(deadline) {
			return doc, errors.New("upload trace callbacks did not drain")
		}
		time.Sleep(time.Millisecond)
	}
}
