//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumOpenURL(uint64_t token,const char *url);
void vellumCancelOpenURL(uint64_t token);
int vellumURLRecipeFixture(const char *url);
*/
import "C"
import (
	"context"
	"unsafe"
)

func nativeOpenURL(ctx context.Context, value string) error {
	_, err := nativeOperationRun(ctx, func(token uint64) {
		s := C.CString(value)
		defer C.free(unsafe.Pointer(s))
		C.vellumOpenURL(C.uint64_t(token), s)
	}, func(token uint64) { C.vellumCancelOpenURL(C.uint64_t(token)) })
	return err
}
func nativeURLRecipeFixture(value string) bool {
	s := C.CString(value)
	defer C.free(unsafe.Pointer(s))
	return C.vellumURLRecipeFixture(s) == 1
}
