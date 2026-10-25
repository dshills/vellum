//go:build darwin && cgo

package cogent

/*
#cgo darwin LDFLAGS: -framework QuickLookUI
#include <stdlib.h>
#include <stdint.h>
void vellumQuickLook(uint64_t token,uint64_t parent,const char *json);
void vellumCancelQuickLook(uint64_t token);
char *vellumQuickLookRecipeFixture(const char *json);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeQuickLook(ctx context.Context, owner *app.Application, parent app.WindowRef, o app.QuickLookOptions) error {
	parentToken, err := alertParentToken(owner, parent)
	if err != nil {
		return err
	}
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	if len(data) > nativeServiceRequestMax {
		return app.ErrServiceTooLarge
	}
	_, err = nativeOperationRun(ctx, func(token uint64) {
		s := C.CString(string(data))
		defer C.free(unsafe.Pointer(s))
		C.vellumQuickLook(C.uint64_t(token), C.uint64_t(parentToken), s)
	}, func(token uint64) { C.vellumCancelQuickLook(C.uint64_t(token)) })
	return err
}
func nativeQuickLookRecipeFixture(wire string) string {
	s := C.CString(wire)
	defer C.free(unsafe.Pointer(s))
	p := C.vellumQuickLookRecipeFixture(s)
	if p == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p)
}
