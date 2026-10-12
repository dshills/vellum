package sdkcontract_test

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

// Public API consumer only; no permissions or OS notification delivery.
type consumerNotifications struct {
	memoryClipboard
	permission app.NotificationPermission
	items      map[string]app.Notification
}

func (*consumerNotifications) NotificationCapabilities() app.NotificationCapabilities {
	return app.NotificationCapabilities{Permissions: true, LocalDelivery: true, Withdrawal: true}
}
func (m *consumerNotifications) NotificationSettings(ctx context.Context) (app.NotificationSettings, error) {
	if err := ctx.Err(); err != nil {
		return app.NotificationSettings{}, err
	}
	return app.NotificationSettings{Permission: m.permission}, nil
}
func (m *consumerNotifications) RequestNotificationPermission(ctx context.Context) (app.NotificationSettings, error) {
	if err := ctx.Err(); err != nil {
		return app.NotificationSettings{}, err
	}
	m.permission = app.NotificationAuthorized
	return m.NotificationSettings(ctx)
}
func (m *consumerNotifications) Notify(ctx context.Context, n app.Notification) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := n.Validate(); err != nil {
		return err
	}
	if m.permission != app.NotificationAuthorized {
		return app.ErrNotificationPermission
	}
	if m.items == nil {
		m.items = make(map[string]app.Notification)
	}
	m.items[n.ID] = n
	return nil
}
func (m *consumerNotifications) WithdrawNotifications(ctx context.Context, ids []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := app.ValidateNotificationIDs(ids); err != nil {
		return err
	}
	for _, id := range ids {
		delete(m.items, id)
	}
	return nil
}
func TestPublicNotificationConsumer(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	m := &consumerNotifications{permission: app.NotificationNotDetermined}
	a.SetServices(m)
	n, err := a.Notifications()
	if err != nil || !n.NotificationCapabilities().LocalDelivery {
		t.Fatal(err)
	}
	ctx := context.Background()
	item := app.Notification{ID: "export-1", Title: "Export complete", Body: "Ready"}
	if err = n.Notify(ctx, item); !errors.Is(err, app.ErrNotificationPermission) {
		t.Fatal(err)
	}
	s, err := n.RequestNotificationPermission(ctx)
	if err != nil || s.Permission != app.NotificationAuthorized {
		t.Fatal(s, err)
	}
	if err = n.Notify(ctx, item); err != nil {
		t.Fatal(err)
	}
	item.Body = "Replaced"
	if err = n.Notify(ctx, item); err != nil || len(m.items) != 1 || m.items[item.ID].Body != "Replaced" {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err = n.WithdrawNotifications(canceled, []string{item.ID}); !errors.Is(err, context.Canceled) || len(m.items) != 1 {
		t.Fatal(err)
	}
	if err = n.WithdrawNotifications(ctx, []string{item.ID}); err != nil || len(m.items) != 0 {
		t.Fatal(err)
	}
}
