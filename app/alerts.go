package app

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"unicode/utf8"
)

type AlertStyle string

const (
	AlertInformation AlertStyle = "information"
	AlertWarning     AlertStyle = "warning"
	AlertCritical    AlertStyle = "critical"
)

type AlertRole string

const (
	AlertDefault     AlertRole = "default"
	AlertCancel      AlertRole = "cancel"
	AlertDestructive AlertRole = "destructive"
)

type AlertButton struct {
	ID, Label string
	Role      AlertRole
}

// AlertOptions is a detached native alert request. Do not mutate slices until
// ShowAlert returns. A 1..512px PNG <=256KiB preserves original colors; nil uses
// the application's native icon. Exactly one default and <=one cancel allowed.
type AlertOptions struct {
	Title, Message      string
	Style               AlertStyle
	Buttons             []AlertButton
	IconPNG             []byte
	SuppressionLabel    string
	InitiallySuppressed bool
}
type AlertResult struct {
	ButtonID   string
	Suppressed bool
}

func (o AlertOptions) Validate() error {
	text := func(s string, max int) bool {
		return len(s) <= max && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
	}
	if strings.TrimSpace(o.Title) == "" || !text(o.Title, 1024) || !text(o.Message, 4096) || !text(o.SuppressionLabel, 256) || len(o.Buttons) < 1 || len(o.Buttons) > 3 || o.InitiallySuppressed && o.SuppressionLabel == "" {
		return fmt.Errorf("vellum: invalid bounded alert options")
	}
	if o.Style != "" && o.Style != AlertInformation && o.Style != AlertWarning && o.Style != AlertCritical {
		return fmt.Errorf("vellum: invalid alert style")
	}
	defaults, cancels := 0, 0
	seen := map[string]bool{}
	for _, b := range o.Buttons {
		if !text(b.ID, 256) || strings.TrimSpace(b.ID) == "" || !text(b.Label, 128) || strings.TrimSpace(b.Label) == "" || seen[b.ID] {
			return fmt.Errorf("vellum: invalid alert button")
		}
		seen[b.ID] = true
		switch b.Role {
		case AlertDefault:
			defaults++
		case AlertCancel:
			cancels++
		case AlertDestructive:
		case "":
		default:
			return fmt.Errorf("vellum: invalid alert role")
		}
	}
	if defaults != 1 || cancels > 1 {
		return fmt.Errorf("vellum: alert requires exactly one default and <=one cancel")
	}
	if len(o.IconPNG) > 256<<10 {
		return ErrServiceTooLarge
	}
	if len(o.IconPNG) > 0 {
		c, err := png.DecodeConfig(bytes.NewReader(o.IconPNG))
		if err != nil || c.Width < 1 || c.Height < 1 || c.Width > 512 || c.Height > 512 {
			return fmt.Errorf("vellum: invalid alert PNG icon")
		}
	}
	return nil
}

// AlertServices is optional. Call from a worker, never from View/owner callbacks.
// Parent is an exact live WindowRef. Closing it or canceling context revokes the
// operation. Selecting a cancel-role button returns its ID; closing/canceling
// the native sheet without a selected button returns ErrCanceled.
type AlertServices interface {
	ShowAlert(context.Context, WindowRef, AlertOptions) (AlertResult, error)
}

func (a *Application) Alerts() (AlertServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	alerts, ok := s.(AlertServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return alerts, nil
}
