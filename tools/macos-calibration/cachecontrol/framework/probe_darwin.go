//go:build darwin && !offscreen

package main

/*
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>
extern bool vellumCacheConfigure(bool default_cache);
extern size_t vellumCacheSnapshot(uint64_t *out, size_t capacity);
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

func cacheConfigure(defaultCache bool) error {
	if !bool(C.vellumCacheConfigure(C.bool(defaultCache))) {
		return errors.New("cache policy already configured or resource created")
	}
	return nil
}
func cacheSnapshot() ([]uint64, error) {
	var values [6]C.uint64_t
	if C.vellumCacheSnapshot(&values[0], 6) != 6 {
		return nil, errors.New("cache policy snapshot size")
	}
	out := make([]uint64, 6)
	for i, value := range values {
		out[i] = uint64(value)
	}
	return out, nil
}
