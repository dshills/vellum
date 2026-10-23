//go:build darwin && cgo && !ios

// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
package core

/*
#include <pthread.h>
#include <pthread/qos.h>
static pthread_override_t vellum_frame_priority_start(void) {
    return pthread_override_qos_class_start_np(pthread_self(), QOS_CLASS_USER_INTERACTIVE, 0);
}
*/
import "C"

import "runtime"

// Private evaluation candidate. Keep the Go worker on this thread until its
// temporary override is ended; never return an elevated thread to the pool.
type renderPriority struct {
	token C.pthread_override_t
	owner C.pthread_t
}

func beginRenderPriority() *renderPriority {
	runtime.LockOSThread()
	token := C.vellum_frame_priority_start()
	if token == nil {
		runtime.UnlockOSThread()
		panic("frame priority override unavailable")
	}
	return &renderPriority{token: token, owner: C.pthread_self()}
}

func (p *renderPriority) end() {
	if p == nil || p.token == nil {
		return
	}
	if C.pthread_self() != p.owner {
		panic("frame priority released on another thread")
	}
	token := p.token
	p.token = nil
	if C.pthread_override_qos_class_end_np(token) != 0 {
		// Fail this private candidate and keep the thread out of the Go pool.
		panic("frame priority override retirement failed")
	}
	runtime.UnlockOSThread()
}
