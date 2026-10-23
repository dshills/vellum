//go:build darwin && cgo && !ios && vellum_priority_test

package core

/*
#include <pthread.h>
#include <pthread/qos.h>
*/
import "C"

import "unsafe"

type priorityState struct {
	thread uintptr
	class  uint32
	rel    int
}

func currentPriorityState() priorityState {
	var class C.qos_class_t
	var rel C.int
	if C.pthread_get_qos_class_np(C.pthread_self(), &class, &rel) != 0 {
		panic("thread QoS query failed")
	}
	return priorityState{uintptr(unsafe.Pointer(C.pthread_self())), uint32(class), int(rel)}
}
