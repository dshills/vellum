//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"strings"
	"testing"
)

func notificationWire(t *testing.T, r nativeNotificationRequest) []byte {
	t.Helper()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func TestNotificationNativeParserAndContentObjectsWithoutCenter(t *testing.T) {
	for _, sound := range []bool{false, true} {
		r := nativeNotificationRequest{Action: 3, Notification: app.Notification{ID: "job_1.2", Title: "Complete 🦋", Subtitle: "subtitle", Body: "line one\nline two", Sound: sound}}
		b := notificationWire(t, r)
		if !nativeNotificationValidate(b) || !nativeNotificationContentFixture(b) {
			t.Fatal("native content mismatch")
		}
		if nativeServicePayloadFixture(b) != 7 {
			t.Fatal("detached producer lifetime")
		}
	}
	for _, r := range []nativeNotificationRequest{{Action: 1}, {Action: 2}, {Action: 4, IDs: []string{"a", "b"}}} {
		if !nativeNotificationValidate(notificationWire(t, r)) {
			t.Fatal(r)
		}
	}
	invalid := []nativeNotificationRequest{{Action: 0}, {Action: 5}, {Action: 1, Notification: app.Notification{ID: "extra"}}, {Action: 2, IDs: []string{"x"}}, {Action: 3, Notification: app.Notification{ID: "x"}}, {Action: 3, Notification: app.Notification{ID: "a/b", Title: "title"}}, {Action: 3, Notification: app.Notification{ID: "x", Title: strings.Repeat("a", 257)}}, {Action: 3, Notification: app.Notification{ID: "x", Title: "a\x00b"}}, {Action: 4}, {Action: 4, IDs: []string{"x", "x"}}, {Action: 4, IDs: make([]string, 33)}}
	for _, r := range invalid {
		if nativeNotificationValidate(notificationWire(t, r)) {
			t.Fatal("invalid native payload", r.Action)
		}
	}
	for _, b := range [][]byte{[]byte(`{"Action":true,"Notification":{},"IDs":null}`), []byte(`{"Action":1.5,"Notification":{},"IDs":null}`), []byte(`null`), []byte(strings.Repeat("x", notificationWireMax+1))} {
		if nativeNotificationValidate(b) {
			t.Fatal("malformed wire admitted")
		}
	}
	base := notificationWire(t, nativeNotificationRequest{Action: 3, Notification: app.Notification{ID: "x", Title: "Valid"}})
	for _, mutate := range []func(map[string]any){func(r map[string]any) { r["Action"] = true }, func(r map[string]any) { r["Action"] = 1.5 }, func(r map[string]any) { r["extra"] = 0 }, func(r map[string]any) { r["Notification"].(map[string]any)["Sound"] = 1 }, func(r map[string]any) { r["Notification"].(map[string]any)["Body"] = []string{"wrong type"} }} {
		var r map[string]any
		json.Unmarshal(base, &r)
		mutate(r)
		wire, _ := json.Marshal(r)
		if nativeNotificationValidate(wire) {
			t.Fatal("typed payload admitted", string(wire))
		}
	}
}
func TestNotificationNativeSettingsNormalizationAndPermissionResult(t *testing.T) {
	withMenuCommitState(t)
	for _, tc := range []struct {
		raw  int
		want app.NotificationPermission
	}{{0, app.NotificationNotDetermined}, {1, app.NotificationDenied}, {2, app.NotificationAuthorized}, {3, app.NotificationProvisional}, {4, app.NotificationPermissionUnknown}, {-1, app.NotificationPermissionUnknown}} {
		s, err := decodeNotificationSettings(nativeNotificationSettingsFixture(tc.raw, 2, 1))
		if err != nil || s.Permission != tc.want || !s.AlertEnabled || s.SoundEnabled {
			t.Fatal(s, err)
		}
	}
	_, err := nativeOperationRun(context.Background(), func(token uint64) { nativeServiceResultFixture(token, nil, 4) }, nil)
	if !errors.Is(err, app.ErrNotificationPermission) {
		t.Fatal(err)
	}
}
