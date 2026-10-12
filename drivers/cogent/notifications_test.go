package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"strings"
	"testing"
)

func TestNotificationSettingsWireRequiresCompleteTypedSnapshot(t *testing.T) {
	s, err := decodeNotificationSettings(`{"Permission":0,"AlertEnabled":false,"SoundEnabled":true}`)
	if err != nil || s.Permission != app.NotificationPermissionUnknown || s.AlertEnabled || !s.SoundEnabled {
		t.Fatal(s, err)
	}
	for _, wire := range []string{`{}`, `{"Permission":1,"AlertEnabled":true}`, `{"Permission":5,"AlertEnabled":false,"SoundEnabled":false}`, `{"Permission":3,"AlertEnabled":1,"SoundEnabled":true}`, `{"Permission":3,"AlertEnabled":true,"SoundEnabled":false,"extra":0}`, `{"Permission":3,"AlertEnabled":true,"SoundEnabled":false} {}`, `null`} {
		if _, err = decodeNotificationSettings(wire); !errors.Is(err, app.ErrUnavailable) {
			t.Fatal(wire, err)
		}
	}
	if _, err = decodeNotificationSettings(strings.Repeat("x", 257)); !errors.Is(err, app.ErrServiceTooLarge) {
		t.Fatal(err)
	}
}
func TestNotificationPreflightNeverDispatchesInvalidCanceledOrClosed(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	s := &platformServices{owner: a}
	if err := s.Notify(context.Background(), app.Notification{ID: "x"}); err == nil {
		t.Fatal("empty content")
	}
	if err := s.WithdrawNotifications(context.Background(), []string{"x", "x"}); err == nil {
		t.Fatal("duplicate IDs")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Notify(ctx, app.Notification{ID: "x", Title: "task"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.WithdrawNotifications(ctx, []string{"x"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, f := range []func(context.Context) (app.NotificationSettings, error){s.NotificationSettings, s.RequestNotificationPermission} {
		if _, err := f(ctx); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err := f(nil); err == nil {
			t.Fatal("nil context")
		}
	}
	a.Close()
	if err := s.Notify(context.Background(), app.Notification{ID: "x", Title: "task"}); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
	if err := s.WithdrawNotifications(context.Background(), []string{"x"}); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}
