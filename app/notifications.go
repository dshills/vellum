package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var ErrNotificationPermission = errors.New("vellum: notification permission required")

const (
	NotificationMaxIDBytes    = 128
	NotificationMaxWithdraw   = 32
	NotificationMaxTitleBytes = 256
	NotificationMaxBodyBytes  = 4096
)

type NotificationPermission uint8

const (
	NotificationPermissionUnknown NotificationPermission = iota
	NotificationNotDetermined
	NotificationDenied
	NotificationAuthorized
	NotificationProvisional
)

// NotificationSettings is a current OS snapshot, not a delivery guarantee.
type NotificationSettings struct {
	Permission                 NotificationPermission
	AlertEnabled, SoundEnabled bool
}

// Notification identifies one immediate local notification. Reusing ID replaces
// that notification through the OS. Sound is opt-in and respects OS settings.
type Notification struct {
	ID, Title, Subtitle, Body string
	Sound                     bool
}

func (n Notification) Validate() error {
	if err := ValidateNotificationIDs([]string{n.ID}); err != nil {
		return err
	}
	if len(n.Title) > NotificationMaxTitleBytes || len(n.Subtitle) > NotificationMaxTitleBytes || len(n.Body) > NotificationMaxBodyBytes {
		return ErrServiceTooLarge
	}
	for _, s := range []string{n.Title, n.Subtitle, n.Body} {
		if !utf8.ValidString(s) || strings.ContainsRune(s, 0) {
			return fmt.Errorf("vellum: invalid notification text")
		}
	}
	if strings.TrimSpace(n.Title) == "" && strings.TrimSpace(n.Body) == "" {
		return fmt.Errorf("vellum: notification needs title or body")
	}
	return nil
}

// ValidateNotificationIDs bounds a withdrawal batch and the framework-owned
// identifier namespace. IDs use ASCII letters, digits, '.', '_' and '-'.
func ValidateNotificationIDs(ids []string) error {
	if len(ids) > NotificationMaxWithdraw {
		return ErrServiceTooLarge
	}
	if len(ids) == 0 {
		return fmt.Errorf("vellum: notification IDs required")
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if len(id) > NotificationMaxIDBytes {
			return ErrServiceTooLarge
		}
		if id == "" || seen[id] {
			return fmt.Errorf("vellum: empty or duplicate notification ID")
		}
		for _, b := range []byte(id) {
			if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '.' || b == '_' || b == '-') {
				return fmt.Errorf("vellum: invalid notification ID")
			}
		}
		seen[id] = true
	}
	return nil
}

// Capabilities describe the implemented transport, independently of bundle
// identity, current authorization or actual OS presentation.
type NotificationCapabilities struct{ Permissions, LocalDelivery, Withdrawal bool }

// NotificationServices is optional; legacy Services remain compatible. Call
// from workers after services become available. Only RequestNotificationPermission
// may prompt. Notify never requests permission. Acceptance does not prove display.
// Withdrawal acknowledges dispatch, not OS completion. Cancellation cannot undo
// an accepted request or dismiss a permission dialog. Delivered items may outlive
// the application; WithdrawNotifications removes only the supplied owned IDs.
type NotificationServices interface {
	NotificationCapabilities() NotificationCapabilities
	NotificationSettings(context.Context) (NotificationSettings, error)
	RequestNotificationPermission(context.Context) (NotificationSettings, error)
	Notify(context.Context, Notification) error
	WithdrawNotifications(context.Context, []string) error
}

func (a *Application) Notifications() (NotificationServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	n, ok := s.(NotificationServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return n, nil
}
