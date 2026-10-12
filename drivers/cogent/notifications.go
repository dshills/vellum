package cogent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"io"
)

const (
	notificationSettingsAction   = 1
	notificationPermissionAction = 2
	notificationDeliveryAction   = 3
	notificationWithdrawAction   = 4
	notificationWireMax          = 32 << 10
)

type nativeNotificationRequest struct {
	Action       int
	Notification app.Notification
	IDs          []string
}

func (s *platformServices) NotificationCapabilities() app.NotificationCapabilities {
	return nativeNotificationCapabilities()
}
func (s *platformServices) NotificationSettings(ctx context.Context) (app.NotificationSettings, error) {
	return s.notificationSettings(ctx, notificationSettingsAction)
}
func (s *platformServices) RequestNotificationPermission(ctx context.Context) (app.NotificationSettings, error) {
	return s.notificationSettings(ctx, notificationPermissionAction)
}
func (s *platformServices) notificationSettings(ctx context.Context, action int) (app.NotificationSettings, error) {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return app.NotificationSettings{}, err
	}
	defer cancel()
	wire, err := nativeNotifications(scope, nativeNotificationRequest{Action: action})
	if err != nil {
		return app.NotificationSettings{}, err
	}
	return decodeNotificationSettings(wire)
}
func (s *platformServices) Notify(ctx context.Context, n app.Notification) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err = n.Validate(); err != nil {
		return err
	}
	_, err = nativeNotifications(scope, nativeNotificationRequest{Action: notificationDeliveryAction, Notification: n})
	return err
}
func (s *platformServices) WithdrawNotifications(ctx context.Context, ids []string) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err = app.ValidateNotificationIDs(ids); err != nil {
		return err
	}
	_, err = nativeNotifications(scope, nativeNotificationRequest{Action: notificationWithdrawAction, IDs: ids})
	return err
}
func decodeNotificationSettings(wire string) (app.NotificationSettings, error) {
	var result app.NotificationSettings
	if len(wire) > 256 {
		return result, app.ErrServiceTooLarge
	}
	// Pointer fields distinguish omitted required fields from legitimate zeroes.
	var value struct {
		Permission                 *app.NotificationPermission
		AlertEnabled, SoundEnabled *bool
	}
	d := json.NewDecoder(bytes.NewBufferString(wire))
	d.DisallowUnknownFields()
	if err := d.Decode(&value); err != nil {
		return result, fmt.Errorf("cogent: invalid notification settings: %w", app.ErrUnavailable)
	}
	var trailing any
	if d.Decode(&trailing) != io.EOF || value.Permission == nil || value.AlertEnabled == nil || value.SoundEnabled == nil || *value.Permission > app.NotificationProvisional {
		return result, app.ErrUnavailable
	}
	result = app.NotificationSettings{Permission: *value.Permission, AlertEnabled: *value.AlertEnabled, SoundEnabled: *value.SoundEnabled}
	return result, nil
}
