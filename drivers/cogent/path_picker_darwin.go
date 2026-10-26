//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumChoosePath(uint64_t,uint64_t,const char *);
void vellumCancelPath(uint64_t);
void vellumRevealPath(uint64_t,const char *);
int vellumPathRecipeFixture(const char *);
int vellumPathCleanupFixture(uint64_t);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeChoosePath(ctx context.Context, a *app.Application, w app.WindowRef, o app.PathOptions) (string, error) {
	parent, e := alertParentToken(a, w)
	if e != nil {
		return "", e
	}
	wire, e := json.Marshal(o)
	if e != nil {
		return "", e
	}
	return nativeOperationRun(ctx, func(token uint64) {
		s := C.CString(string(wire))
		defer C.free(unsafe.Pointer(s))
		C.vellumChoosePath(C.uint64_t(token), C.uint64_t(parent), s)
	}, func(token uint64) { C.vellumCancelPath(C.uint64_t(token)) })
}
func nativeRevealPath(ctx context.Context, path string) error {
	_, e := nativeOperationRun(ctx, func(token uint64) {
		s := C.CString(path)
		defer C.free(unsafe.Pointer(s))
		C.vellumRevealPath(C.uint64_t(token), s)
	}, func(token uint64) { nativeServiceFinish(token, nativeResult{err: app.ErrCanceled}) })
	return e
}
func nativePathRecipeFixture(wire string) bool {
	s := C.CString(wire)
	defer C.free(unsafe.Pointer(s))
	return C.vellumPathRecipeFixture(s) == 1
}

func nativePathCleanupFixture(token uint64) bool {
	return C.vellumPathCleanupFixture(C.uint64_t(token)) == 1
}
