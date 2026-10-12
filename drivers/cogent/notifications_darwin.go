//go:build darwin && cgo

package cogent

/*
#cgo darwin CFLAGS: -fobjc-arc
#cgo darwin LDFLAGS: -framework UserNotifications
#include <stdlib.h>
#include <stdint.h>
void vellumNotifications(uint64_t token, const char *json);
int vellumNotificationValidate(const char *json);
int vellumNotificationContentFixture(const char *json);
const char *vellumNotificationSettingsFixture(int auth, int alert, int sound);
*/
import "C"
import (
	"context"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"unsafe"
)

func nativeNotificationCapabilities() app.NotificationCapabilities {
	return app.NotificationCapabilities{Permissions: true, LocalDelivery: true, Withdrawal: true}
}
func nativeNotifications(ctx context.Context, r nativeNotificationRequest) (string, error) {
	return nativeOperationRun(ctx, func(token uint64) {
		data, err := json.Marshal(r)
		if err != nil {
			nativeServiceFinish(token, nativeResult{err: err})
			return
		}
		if len(data) > notificationWireMax {
			nativeServiceFinish(token, nativeResult{err: app.ErrServiceTooLarge})
			return
		}
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumNotifications(C.uint64_t(token), text)
	}, nil)
}
func nativeNotificationValidate(data []byte) bool {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	return C.vellumNotificationValidate(text) != 0
}
func nativeNotificationContentFixture(data []byte) bool {
	text := C.CString(string(data))
	defer C.free(unsafe.Pointer(text))
	return C.vellumNotificationContentFixture(text) != 0
}
func nativeNotificationSettingsFixture(auth, alert, sound int) string {
	text := C.vellumNotificationSettingsFixture(C.int(auth), C.int(alert), C.int(sound))
	if text == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text)
}
