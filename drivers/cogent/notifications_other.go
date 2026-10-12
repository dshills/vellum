//go:build !darwin || !cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
)

func nativeNotificationCapabilities() app.NotificationCapabilities {
	return app.NotificationCapabilities{}
}
func nativeNotifications(context.Context, nativeNotificationRequest) (string, error) {
	return "", app.ErrUnavailable
}
