//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumShowAlert(uint64_t token,uint64_t parent,const char *json);
void vellumCancelAlert(uint64_t token);
char *vellumAlertRecipeFixture(const char *json);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeShowAlert(ctx context.Context, owner *app.Application, parent app.WindowRef, o app.AlertOptions) (app.AlertResult, error) {
	parentToken, err := alertParentToken(owner, parent)
	if err != nil {
		return app.AlertResult{}, err
	}
	payload, err := json.Marshal(o)
	if err != nil {
		return app.AlertResult{}, err
	}
	if len(payload) > nativeServiceRequestMax {
		return app.AlertResult{}, app.ErrServiceTooLarge
	}
	wire, err := nativeOperationRun(ctx, func(token uint64) {
		text := C.CString(string(payload))
		defer C.free(unsafe.Pointer(text))
		C.vellumShowAlert(C.uint64_t(token), C.uint64_t(parentToken), text)
	}, func(token uint64) { C.vellumCancelAlert(C.uint64_t(token)) })
	if err != nil {
		return app.AlertResult{}, err
	}
	return decodeAlertResult(wire, o)
}
func nativeAlertRecipeFixture(wire string) string {
	input := C.CString(wire)
	defer C.free(unsafe.Pointer(input))
	text := C.vellumAlertRecipeFixture(input)
	if text == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text)
}
