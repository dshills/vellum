//go:build darwin && cgo && !offscreen && vellum_native_checks

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dshills/vellum/app"
)

func TestNativeNotificationCheckPureBoundsAndDetachedContent(t *testing.T) {
	for _, wire := range []string{`{"IDs":["one","two"]}`, `{"IDs":["ASCII-_ ."]}`} {
		want := !strings.Contains(wire, " ")
		if nativeNotificationCheckValidate(wire) != want {
			t.Fatalf("ID validation: %q", wire)
		}
	}
	invalid := []string{`{}`, `null`, `[]`, `{"IDs":[]}`, `{"IDs":null}`, `{"IDs":[true]}`, `{"IDs":["one","one"]}`, `{"IDs":["x/y"]}`, `{"IDs":["🙂"]}`, `{"IDs":["x\u0000y"]}`, `{"IDs":["one"],"extra":0}`, `{"IDs":["` + strings.Repeat("x", 129) + `"]}`}
	ids := make([]string, 33)
	for i := range ids {
		ids[i] = strings.Repeat("x", i+1)
	}
	tooMany, _ := json.Marshal(struct{ IDs []string }{ids})
	invalid = append(invalid, string(tooMany), strings.Repeat(" ", 8193))
	for _, wire := range invalid {
		if nativeNotificationCheckValidate(wire) {
			t.Errorf("accepted malformed/bounded request %q", wire[:min(len(wire), 160)])
		}
	}
	wire := nativeNotificationCheckFixture(0)
	state, err := decodeNativeNotificationInventory(wire, []string{"fixture-0", "fixture-1"})
	if err != nil || len(state.Pending) != 0 || len(state.Delivered) != 2 || state.Delivered[0].ID != "fixture-0" || state.Delivered[0].Title != "Vellum 🙂" || state.Delivered[0].Subtitle != `"quoted"` || state.Delivered[0].Body != "detached\nbody" || state.Delivered[0].Sound {
		t.Fatalf("actual native content/wire fixture: %+v %v", state, err)
	}
	// The native autorelease pool and request/content objects have already ended.
	// Further fixture calls and returned-value mutation must leave the wire intact.
	state.Delivered[0].Body = "changed"
	if nativeNotificationCheckFixture(1) != "" || nativeNotificationCheckFixture(2) != "" {
		t.Fatal("aggregate overflow or duplicate result was truncated/accepted")
	}
	again, err := decodeNativeNotificationInventory(wire, []string{"fixture-0", "fixture-1"})
	if err != nil || again.Delivered[0].Body != "detached\nbody" {
		t.Fatal("detached native reply was changed")
	}
}

func TestNativeNotificationCheckReplyRefusalsAndCanceledAdmission(t *testing.T) {
	withMenuCommitState(t)
	wire := nativeNotificationCheckFixture(0)
	for _, broken := range []string{
		strings.Replace(wire, `"MainThreadDispatch":true`, `"MainThreadDispatch":false`, 1),
		strings.Replace(wire, `"Body":"detached\nbody"`, `"Body":null`, 1),
		strings.ReplaceAll(wire, "fixture-1", "fixture-0"),
		strings.Replace(wire, "com.vellum.native-checks.notifications.fixture", "com.other.app", 1),
		wire + " {}", `{}`, `{"BundleIdentifier":null}`, strings.Repeat(" ", nativeNotificationCheckReplyMax+1),
	} {
		if broken == wire {
			t.Fatal("malformed-result mutation missed native fixture")
		}
		if _, err := decodeNativeNotificationInventory(broken, []string{"fixture-0", "fixture-1"}); err == nil {
			t.Fatal("accepted malformed/unbounded native result")
		}
	}
	if _, err := decodeNativeNotificationInventory(wire, []string{"other"}); err == nil {
		t.Fatal("accepted an unrequested ID")
	}
	nativeCallbacks.Lock()
	before := len(nativeCallbacks.operations)
	nativeCallbacks.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := InspectNativeNotifications(ctx, []string{"one"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled preflight: %v", err)
	}
	if _, err := InspectNativeNotifications(context.Background(), []string{"one", "one"}); err == nil {
		t.Fatal("duplicate IDs admitted")
	}
	nativeCallbacks.Lock()
	after := len(nativeCallbacks.operations)
	nativeCallbacks.Unlock()
	if before != after {
		t.Fatal("failed preflight leaked service admission")
	}
	if _, err := InspectNativeNotifications(nil, []string{"one"}); err == nil {
		t.Fatal("nil context admitted")
	}
	nativeCallbacks.Lock()
	nativeCallbacks.stopped = true
	nativeCallbacks.Unlock()
	if _, err := InspectNativeNotifications(context.Background(), []string{"one"}); !errors.Is(err, app.ErrClosed) {
		t.Fatalf("shutdown admission: %v", err)
	}
}
