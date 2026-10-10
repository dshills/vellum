//go:build darwin && cgo

package main

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdint.h>
#include <stdlib.h>
void vellumDropFixture(uint64_t token, const char *payload);
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

var fixtureSerial atomic.Uint64
var fixtureCalls = struct {
	sync.Mutex
	results map[uint64]chan int
}{results: make(map[uint64]chan int)}

//export vellumDropFixtureActive
func vellumDropFixtureActive(token C.uint64_t) C.int {
	fixtureCalls.Lock()
	defer fixtureCalls.Unlock()
	if _, ok := fixtureCalls.results[uint64(token)]; ok {
		return 1
	}
	return 0
}

//export vellumDropFixtureResult
func vellumDropFixtureResult(token C.uint64_t, status C.int) {
	fixtureCalls.Lock()
	defer fixtureCalls.Unlock()
	if result, ok := fixtureCalls.results[uint64(token)]; ok {
		select {
		case result <- int(status):
		default:
		}
	}
}
func receiveFileURLs(ctx context.Context, title string, paths []string, x, y float64) error {
	payload, err := json.Marshal(struct {
		Title string
		Paths []string
		X, Y  float64
	}{title, paths, x, y})
	if err != nil {
		return err
	}
	token := fixtureSerial.Add(1)
	result := make(chan int, 1)
	fixtureCalls.Lock()
	fixtureCalls.results[token] = result
	fixtureCalls.Unlock()
	defer func() { fixtureCalls.Lock(); delete(fixtureCalls.results, token); fixtureCalls.Unlock() }()
	value := C.CString(string(payload))
	defer C.free(unsafe.Pointer(value))
	C.vellumDropFixture(C.uint64_t(token), value)
	select {
	case status := <-result:
		if status != 1 {
			return fmt.Errorf("Cocoa file-URL receive fixture status %d", status)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
