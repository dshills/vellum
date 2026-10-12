package app

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"unicode/utf8"
)

const (
	TrayMaxPNGBytes  = 1 << 20
	TrayMaxIconSize  = 256
	TrayMaxMenuItems = 128
)

// TrayMenuItem references a registered command or an explicit separator.
type TrayMenuItem struct {
	Command   string
	Separator bool
}

// TrayOptions updates the application's single menu-bar status item. At least
// Title or IconPNG is required. PNG is decoded before native installation; its
// dimensions are bounded independently of encoded size. Template asks macOS to
// tint a monochrome icon for its current appearance. The native icon fits 18×18
// Cocoa points while preserving aspect ratio.
// Keep all input slices immutable until SetTray returns.
type TrayOptions struct {
	Title, Tooltip string
	// AccessibilityLabel defaults to Tooltip, then Title. An icon-only item must
	// supply one of these names; a decorative unnamed button is refused.
	AccessibilityLabel string
	IconPNG            []byte
	Template           bool
	Menu               []TrayMenuItem
}

// Validate checks metadata and PNG header bounds. The adapter must additionally
// decode the complete PNG before installing it; header validity is not file validity.
func (o TrayOptions) Validate() error {
	if len(o.Title) > 256 || len(o.Tooltip) > 1024 || len(o.AccessibilityLabel) > 1024 || len(o.IconPNG) > TrayMaxPNGBytes || len(o.Menu) > TrayMaxMenuItems {
		return ErrServiceTooLarge
	}
	for _, s := range []string{o.Title, o.Tooltip, o.AccessibilityLabel} {
		if !utf8.ValidString(s) || strings.ContainsRune(s, 0) {
			return fmt.Errorf("vellum: invalid tray text")
		}
	}
	if strings.TrimSpace(o.Title) == "" && len(o.IconPNG) == 0 {
		return fmt.Errorf("vellum: tray needs title or icon")
	}
	if strings.TrimSpace(o.AccessibilityLabel) == "" && strings.TrimSpace(o.Tooltip) == "" && strings.TrimSpace(o.Title) == "" {
		return fmt.Errorf("vellum: tray needs accessible name")
	}
	if o.Template && len(o.IconPNG) == 0 {
		return fmt.Errorf("vellum: tray template needs icon")
	}
	for _, item := range o.Menu {
		if len(item.Command) > 128 {
			return ErrServiceTooLarge
		}
		if item.Separator {
			if item.Command != "" {
				return fmt.Errorf("vellum: tray separator has command")
			}
		} else if item.Command == "" || !utf8.ValidString(item.Command) || strings.ContainsRune(item.Command, 0) {
			return fmt.Errorf("vellum: invalid tray command")
		}
	}
	if len(o.IconPNG) > 0 {
		cfg, err := png.DecodeConfig(bytes.NewReader(o.IconPNG))
		if err != nil {
			return fmt.Errorf("vellum: invalid tray PNG: %w", err)
		}
		if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > TrayMaxIconSize || cfg.Height > TrayMaxIconSize {
			return ErrServiceTooLarge
		}
	}
	return nil
}

type TrayCapabilities struct{ StatusItem, PNGIcon, CommandMenu bool }

// TrayServices is an optional Services extension. Invoke methods from workers,
// outside UI-owner callbacks. Labels/enabled state are captured at SetTray;
// activation posts to the owner and rechecks command availability. Cancellation
// can race an accepted native replacement and cannot promise rollback. Tray
// lifetime is application-wide and does not depend on one particular window.
type TrayServices interface {
	TrayCapabilities() TrayCapabilities
	SetTray(context.Context, TrayOptions) error
	ClearTray(context.Context) error
}

func (a *Application) Tray() (TrayServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	t, ok := s.(TrayServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return t, nil
}
