package app

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNotificationValidationBoundsAndIdentity(t *testing.T) {
	for _, n := range []Notification{{ID: "job-1", Title: "Complete 🦋"}, {ID: "abc_1.2", Body: "line one\nline two", Sound: true}, {ID: strings.Repeat("a", 128), Title: strings.Repeat("a", 256), Subtitle: strings.Repeat("b", 256), Body: strings.Repeat("c", 4096)}} {
		if err := n.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []Notification{{}, {ID: "x"}, {ID: "a/b", Title: "a"}, {ID: "x", Title: "\xff"}, {ID: "x", Body: "bad\x00body"}, {ID: "x", Subtitle: "only subtitle"}, {ID: "x", Title: " \n", Body: "\t"}} {
		if n.Validate() == nil {
			t.Fatal("invalid notification admitted", n)
		}
	}
	for _, n := range []Notification{{ID: strings.Repeat("x", 129), Title: "a"}, {ID: "x", Title: strings.Repeat("a", 257)}, {ID: "x", Subtitle: strings.Repeat("a", 257)}, {ID: "x", Body: strings.Repeat("a", 4097)}} {
		if !errors.Is(n.Validate(), ErrServiceTooLarge) {
			t.Fatal("unbounded notification")
		}
	}
	for _, ids := range [][]string{nil, {"a", "a"}, {""}, {"a\x00b"}, {"🦋"}} {
		if ValidateNotificationIDs(ids) == nil {
			t.Fatal("invalid IDs admitted", ids)
		}
	}
	if !errors.Is(ValidateNotificationIDs(make([]string, 33)), ErrServiceTooLarge) {
		t.Fatal("unbounded withdrawal")
	}
}

type notificationFake struct{ legacyClipboardServices }

func (notificationFake) NotificationCapabilities() NotificationCapabilities {
	return NotificationCapabilities{Permissions: true}
}
func (notificationFake) NotificationSettings(context.Context) (NotificationSettings, error) {
	return NotificationSettings{Permission: NotificationDenied}, nil
}
func (notificationFake) RequestNotificationPermission(context.Context) (NotificationSettings, error) {
	return NotificationSettings{}, nil
}
func (notificationFake) Notify(context.Context, Notification) error            { return nil }
func (notificationFake) WithdrawNotifications(context.Context, []string) error { return nil }
func TestNotificationsOptionalLegacyCompatibility(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	for _, s := range []Services{nil, legacyClipboardServices{}} {
		a.SetServices(s)
		if _, err := a.Notifications(); !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
	}
	a.SetServices(notificationFake{})
	n, err := a.Notifications()
	if err != nil || !n.NotificationCapabilities().Permissions {
		t.Fatal(err)
	}
	settings, err := n.NotificationSettings(context.Background())
	if err != nil || settings.Permission != NotificationDenied {
		t.Fatal(settings, err)
	}
	a.Close()
	if _, err = a.Notifications(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
