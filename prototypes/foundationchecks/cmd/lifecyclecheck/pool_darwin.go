//go:build darwin && cgo

package main

/*
#cgo LDFLAGS: -framework Foundation
void *vellum_check_pool_begin(void);
void vellum_check_pool_end(void *pool);
*/
import "C"
import "unsafe"

func poolBegin() unsafe.Pointer { return C.vellum_check_pool_begin() }
func poolEnd(p unsafe.Pointer)  { C.vellum_check_pool_end(p) }
