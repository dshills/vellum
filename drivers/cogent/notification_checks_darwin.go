//go:build darwin && cgo && !offscreen && vellum_native_checks

package cogent

/*
#include <stdint.h>
#include <stdlib.h>
void vellumNotificationCheck(uint64_t token, const char *json);
char *vellumNotificationCheckBundle(void);
int vellumNotificationCheckValidate(const char *json);
char *vellumNotificationCheckFixture(int mode);
*/
import "C"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/dshills/vellum/app"
)

const nativeNotificationCheckReplyMax = 64 << 10
const nativeNotificationCheckBundlePrefix = "com.vellum.native-checks.notifications."

// NativeNotificationItemCheck describes only a requested framework-owned ID.
// These diagnostics do not exist in ordinary or offscreen SDK builds.
type NativeNotificationItemCheck struct {
	ID, Title, Subtitle, Body string
	Sound                     bool
}

type NativeNotificationBundleCheck struct {
	AppBundle        bool
	BundleIdentifier string
}

// Pending and Delivered are separate asynchronous OS observations. Their union
// is not an atomic snapshot. NativeOutstanding counts retained service operations
// after this query's terminal ACK, including canceled submissions not yet ACKed.
type NativeNotificationInventoryCheck struct {
	BundleIdentifier   string
	MainThreadDispatch bool
	Pending, Delivered []NativeNotificationItemCheck
	NativeOutstanding  int
}

func NativeNotificationCheckBundle() (NativeNotificationBundleCheck, error) {
	var result NativeNotificationBundleCheck
	text := C.vellumNotificationCheckBundle()
	if text == nil {
		return result, app.ErrUnavailable
	}
	defer C.free(unsafe.Pointer(text))
	err := json.Unmarshal([]byte(C.GoString(text)), &result)
	return result, err
}

// InspectNativeNotifications never prompts, removes items, installs a delegate
// or changes settings. Native queries require an isolated private-check bundle.
func InspectNativeNotifications(ctx context.Context, ids []string) (NativeNotificationInventoryCheck, error) {
	var result NativeNotificationInventoryCheck
	if ctx == nil {
		return result, fmt.Errorf("cogent: notification inventory needs a context")
	}
	if err := app.ValidateNotificationIDs(ids); err != nil {
		return result, err
	}
	wire, err := nativeOperationRun(ctx, func(token uint64) {
		data, err := json.Marshal(struct{ IDs []string }{ids})
		if err != nil {
			nativeServiceFinish(token, nativeResult{err: err})
			return
		}
		text := C.CString(string(data))
		defer C.free(unsafe.Pointer(text))
		C.vellumNotificationCheck(C.uint64_t(token), text)
	}, nil)
	if err != nil {
		return result, err
	}
	result, err = decodeNativeNotificationInventory(wire, ids)
	if err == nil {
		nativeCallbacks.Lock()
		result.NativeOutstanding = len(nativeCallbacks.operations)
		nativeCallbacks.Unlock()
	}
	return result, err
}

func decodeNativeNotificationInventory(wire string, ids []string) (NativeNotificationInventoryCheck, error) {
	var result NativeNotificationInventoryCheck
	if len(wire) > nativeNotificationCheckReplyMax {
		return result, app.ErrServiceTooLarge
	}
	if err := app.ValidateNotificationIDs(ids); err != nil {
		return result, err
	}
	var value struct {
		BundleIdentifier   *string
		MainThreadDispatch *bool
		Pending, Delivered *[]json.RawMessage
	}
	if err := notificationCheckDecode(wire, &value); err != nil {
		return result, err
	}
	if value.BundleIdentifier == nil || len(*value.BundleIdentifier) > 256 || !strings.HasPrefix(*value.BundleIdentifier, nativeNotificationCheckBundlePrefix) || value.MainThreadDispatch == nil || !*value.MainThreadDispatch || value.Pending == nil || value.Delivered == nil {
		return result, app.ErrUnavailable
	}
	result.BundleIdentifier, result.MainThreadDispatch = *value.BundleIdentifier, true
	allowed := make(map[string]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}
	for i, raw := range [2][]json.RawMessage{*value.Pending, *value.Delivered} {
		if len(raw) > len(ids) {
			return NativeNotificationInventoryCheck{}, app.ErrUnavailable
		}
		items := make([]NativeNotificationItemCheck, 0, len(raw))
		seen := make(map[string]bool, len(raw))
		for _, row := range raw {
			var item struct {
				ID, Title, Subtitle, Body *string
				Sound                     *bool
			}
			if err := notificationCheckDecode(string(row), &item); err != nil {
				return NativeNotificationInventoryCheck{}, err
			}
			if item.ID == nil || !allowed[*item.ID] || seen[*item.ID] || item.Title == nil || item.Subtitle == nil || item.Body == nil || item.Sound == nil {
				return NativeNotificationInventoryCheck{}, app.ErrUnavailable
			}
			n := app.Notification{ID: *item.ID, Title: *item.Title, Subtitle: *item.Subtitle, Body: *item.Body, Sound: *item.Sound}
			if err := n.Validate(); err != nil {
				return NativeNotificationInventoryCheck{}, err
			}
			seen[n.ID] = true
			items = append(items, NativeNotificationItemCheck{n.ID, n.Title, n.Subtitle, n.Body, n.Sound})
		}
		if i == 0 {
			result.Pending = items
		} else {
			result.Delivered = items
		}
	}
	return result, nil
}

func notificationCheckDecode(wire string, value any) error {
	d := json.NewDecoder(bytes.NewBufferString(wire))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return fmt.Errorf("cogent: invalid notification inventory: %w", app.ErrUnavailable)
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return app.ErrUnavailable
	}
	return nil
}

func nativeNotificationCheckValidate(wire string) bool {
	text := C.CString(wire)
	defer C.free(unsafe.Pointer(text))
	return C.vellumNotificationCheckValidate(text) != 0
}

func nativeNotificationCheckFixture(mode int) string {
	text := C.vellumNotificationCheckFixture(C.int(mode))
	if text == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text)
}
