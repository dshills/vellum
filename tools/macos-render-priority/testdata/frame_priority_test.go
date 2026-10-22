//go:build darwin && cgo && !ios

package core

import (
	"runtime"
	"sync"
	"testing"
)

func TestScopedRenderPriorityNativeRetirementAndPanic(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	before := currentPriorityState()
	t.Logf("requested base QoS before override: class=%d relative=%d", before.class, before.rel)
	for i := 0; i < 1000; i++ {
		p := beginRenderPriority()
		if p.token == nil || currentPriorityState() != before {
			t.Fatal("override missing or changed requested base QoS/thread")
		}
		p.end()
		p.end() // Repeated cleanup cannot end a native token twice.
		if currentPriorityState() != before {
			t.Fatal("native cleanup changed base QoS or owner")
		}
	}
	func() {
		defer func() {
			if recover() != "test render failure" {
				t.Fatal("render panic not retained")
			}
		}()
		p := beginRenderPriority()
		defer p.end()
		panic("test render failure")
	}()
	if currentPriorityState() != before {
		t.Fatal("panic cleanup changed base QoS/thread")
	}
}

func TestConcurrentRenderPriorityKeepsOwnersSeparate(t *testing.T) {
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			before := currentPriorityState()
			for j := 0; j < 25; j++ {
				p := beginRenderPriority()
				runtime.Gosched()
				p.end()
				if currentPriorityState() != before {
					t.Error("worker moved or changed its requested base QoS")
					return
				}
			}
		}()
	}
	workers.Wait()
}
