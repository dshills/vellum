//go:build darwin && !offscreen

package main

/*
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore -framework Metal
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>
extern bool vellumCacheConfigure(bool default_cache);
extern size_t vellumCacheSnapshot(uint64_t *out, size_t capacity);
extern bool vellumReuseConfigure(bool enabled);
extern size_t vellumReuseSnapshot(unsigned char *out, size_t capacity);
extern bool vellumUploadConfigure(bool enabled);
extern size_t vellumUploadSnapshot(unsigned char *out, size_t capacity);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"os"
	"time"
	"unsafe"
)

func probeConfigure(enabled bool) error {
	if len(os.Args) != 4 || (os.Args[3] != "writecombined" && os.Args[3] != "default") {
		return errors.New("cache mode required")
	}
	if !bool(C.vellumCacheConfigure(C.bool(os.Args[3] == "default"))) {
		return errors.New("cache control already configured")
	}
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
			var values [6]C.uint64_t
			if C.vellumCacheSnapshot(&values[0], 6) != 6 {
				return nil, errors.New("cache snapshot bound")
			}
			policy := make([]uint64, 6)
			for i, value := range values {
				policy[i] = uint64(value)
			}
			doc["cache_policy"] = policy
			return doc, nil
		}
		if time.Now().After(deadline) {
			return doc, errors.New("upload trace callbacks did not drain")
		}
		time.Sleep(time.Millisecond)
	}
}

func reuseConfigure(enabled bool) error {
	if !bool(C.vellumReuseConfigure(C.bool(enabled))) {
		return errors.New("reuse trace already configured")
	}
	return nil
}
func reuseSnapshot() (map[string]any, error) {
	const capacity = 2 << 20
	buffer := C.malloc(capacity)
	if buffer == nil {
		return nil, errors.New("reuse trace buffer allocation")
	}
	defer C.free(buffer)
	deadline := time.Now().Add(2 * time.Second)
	for {
		n := C.vellumReuseSnapshot((*C.uchar)(buffer), capacity)
		if n >= capacity {
			return nil, errors.New("reuse trace byte bound")
		}
		var doc map[string]any
		if err := json.Unmarshal(C.GoBytes(unsafe.Pointer(buffer), C.int(n)), &doc); err != nil {
			return nil, err
		}
		if doc["pending"] == float64(0) {
			return doc, nil
		}
		if time.Now().After(deadline) {
			return doc, errors.New("reuse trace callbacks did not drain")
		}
		time.Sleep(time.Millisecond)
	}
}
