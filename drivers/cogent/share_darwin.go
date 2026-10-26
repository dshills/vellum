//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumShare(uint64_t token,uint64_t parent,const char *json);
void vellumCancelShare(uint64_t token);
int vellumShareRecipeFixture(const char *json);
int vellumShareCleanupFixture(uint64_t token,int selected);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeShare(ctx context.Context, a *app.Application, w app.WindowRef, o app.ShareOptions) (string, error) {
	parent, err := alertParentToken(a, w)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	if len(data) > nativeServiceRequestMax {
		return "", app.ErrServiceTooLarge
	}
	return nativeOperationRun(ctx, func(token uint64) {
		s := C.CString(string(data))
		defer C.free(unsafe.Pointer(s))
		C.vellumShare(C.uint64_t(token), C.uint64_t(parent), s)
	}, func(token uint64) { C.vellumCancelShare(C.uint64_t(token)) })
}
func nativeShareRecipeFixture(wire string) bool {
	s := C.CString(wire)
	defer C.free(unsafe.Pointer(s))
	return C.vellumShareRecipeFixture(s) == 1
}
func nativeShareCleanupFixture(token uint64, selected bool) bool {
	var flag C.int
	if selected {
		flag = 1
	}
	return C.vellumShareCleanupFixture(C.uint64_t(token), flag) == 1
}
