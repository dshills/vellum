//go:build darwin && cgo

package main

/*
#cgo LDFLAGS: -framework Foundation
void *vellum_check_pool_begin(void);
void vellum_check_pool_end(void *pool);
void *vellum_bare_view_new(void);
void vellum_bare_view_free(void *view);
*/
import "C"
import "unsafe"

func poolBegin() unsafe.Pointer { return C.vellum_check_pool_begin() }
func poolEnd(p unsafe.Pointer)  { C.vellum_check_pool_end(p) }

func bareViewNew() unsafe.Pointer   { return C.vellum_bare_view_new() }
func bareViewFree(p unsafe.Pointer) { C.vellum_bare_view_free(p) }
