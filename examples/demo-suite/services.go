package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sampleText = "Vellum native services demo\nUnicode: café · שלום · 日本語 · 👋🏽\n"
const notificationID = "showcase-service-demo"

// Native service calls run outside the UI owner. Only one operation is admitted.
func (d *demo) service(name string, work func(context.Context) (string, error)) {
	if d.busy {
		return
	}
	ctx, cancel := context.WithTimeout(d.owner.Context(), 2*time.Minute)
	d.busy = true
	d.cancelService = cancel
	d.serviceResult = name + "…"
	d.refresh()
	go func() {
		text, err := work(ctx)
		cancel()
		_ = d.owner.PostWait(d.owner.Context(), func() {
			d.busy = false
			d.cancelService = nil
			if err != nil {
				if errors.Is(err, app.ErrCanceled) || errors.Is(err, context.Canceled) {
					d.serviceResult = name + ": canceled"
				} else {
					d.serviceResult = name + ": " + err.Error()
				}
			} else {
				d.serviceResult = text
			}
			d.refresh()
		})
	}()
}
func (d *demo) nativeAction(action string) {
	d.service(action, func(ctx context.Context) (string, error) {
		s, err := d.owner.Services()
		if err != nil {
			return "", err
		}
		switch action {
		case "Install app menu":
			err = s.SetMenu(ctx, []app.MenuGroup{{Label: "Showcase", Commands: []string{"show-services", "show-editor", "show-modal", "increment-demo"}}})
			return "Showcase menu installed — its commands use the shared Go registry", err
		case "Show tray":
			t, e := d.owner.Tray()
			if e != nil {
				return "", e
			}
			e = t.SetTray(ctx, app.TrayOptions{Title: "Vellum", Tooltip: "Vellum native services demo", Menu: []app.TrayMenuItem{{Command: "show-services"}, {Command: "increment-demo"}, {Separator: true}, {Command: "show-modal"}}})
			return "Vellum status item installed in the macOS menu bar", e
		case "Hide tray":
			t, e := d.owner.Tray()
			if e != nil {
				return "", e
			}
			return "Status item removed", t.ClearTray(ctx)
		case "Copy rich sample":
			c, e := d.owner.Clipboard()
			if e != nil {
				return "", e
			}
			e = c.WriteClipboardData(ctx, []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte(sampleText)}, {MIME: "text/html", Data: []byte("<strong>Vellum native services demo</strong><p>Unicode: café · שלום · 日本語 · 👋🏽</p>")}, {MIME: "application/x-vellum-showcase", Data: []byte(`{"demo":"native-services","version":1}`)}})
			return "Copied plain text, HTML and custom MIME sample to the clipboard", e
		case "Read clipboard":
			c, e := d.owner.Clipboard()
			if e != nil {
				return "", e
			}
			items, e := c.ReadClipboardData(ctx, []string{"text/plain", "text/html", "application/x-vellum-showcase"})
			if e != nil {
				return "", e
			}
			var lines []string
			for _, r := range items {
				preview := string(r.Data)
				if len([]rune(preview)) > 300 {
					preview = string([]rune(preview)[:300]) + "…"
				}
				lines = append(lines, fmt.Sprintf("%s (%d bytes): %s", r.MIME, len(r.Data), preview))
			}
			if len(lines) == 0 {
				return "No requested clipboard formats available", nil
			}
			return strings.Join(lines, "\n"), nil
		case "Open file":
			path, e := s.ChooseFile(ctx, app.FileOptions{Mode: app.OpenFile, Title: "Choose a file — preview its path", Extensions: []string{"txt", "md", "go"}})
			if e != nil {
				return "", e
			}
			return "Selected: " + path + "\nThis chooser demonstration does not read or change the file.", nil
		case "Save sample":
			path, e := s.ChooseFile(ctx, app.FileOptions{Mode: app.SaveFile, Title: "Save the Vellum sample", Filename: "Vellum-demo.txt", Extensions: []string{"txt"}})
			if e != nil {
				return "", e
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			e = os.WriteFile(path, []byte(sampleText), 0600)
			return "Sample saved to " + path, e
		case "Notification settings", "Request permission", "Send notification", "Withdraw notification":
			n, e := d.owner.Notifications()
			if e != nil {
				return "", e
			}
			switch action {
			case "Notification settings":
				v, e := n.NotificationSettings(ctx)
				return permissionSummary(v), e
			case "Request permission":
				v, e := n.RequestNotificationPermission(ctx)
				return permissionSummary(v), e
			case "Send notification":
				e = n.Notify(ctx, app.Notification{ID: notificationID, Title: "Vellum Showcase", Body: "Go called the macOS local notification service."})
				return "Notification accepted by macOS. Banner display depends on macOS settings and foreground policy.", e
			default:
				return "Withdrawal dispatched for this demo's notification", n.WithdrawNotifications(ctx, []string{notificationID})
			}
		}
		return "", fmt.Errorf("unknown service action %q", action)
	})
}
func permissionSummary(s app.NotificationSettings) string {
	names := map[app.NotificationPermission]string{app.NotificationPermissionUnknown: "Unknown", app.NotificationNotDetermined: "Not determined", app.NotificationDenied: "Denied", app.NotificationAuthorized: "Authorized", app.NotificationProvisional: "Provisional"}
	return fmt.Sprintf("Notification permission: %s · alerts: %t · sound: %t", names[s.Permission], s.AlertEnabled, s.SoundEnabled)
}
func (d *demo) servicesView() ui.Node {
	button := func(action string) ui.Node {
		return ui.Node{ID: "service-" + strings.ReplaceAll(action, " ", "-"), Kind: ui.Button, Text: action, Disabled: d.busy, Activate: func() { d.nativeAction(action) }}
	}
	row := func(id string, actions ...string) ui.Node {
		n := ui.Node{ID: id, Kind: ui.Row}
		for _, a := range actions {
			n.Children = append(n.Children, button(a))
		}
		return n
	}
	return ui.Node{ID: "services", Kind: ui.Column, Grow: true, Children: []ui.Node{
		label("service-intro", "Try the real macOS services. Each button reports its result below."),
		row("menu-actions", "Install app menu", "Show tray", "Hide tray"),
		label("clipboard-note", "Copy replaces your clipboard with a demo sample. Read displays requested formats here."),
		row("clipboard-actions", "Copy rich sample", "Read clipboard"),
		row("file-actions", "Open file", "Save sample"),
		row("notification-actions", "Notification settings", "Request permission", "Send notification", "Withdraw notification"),
		label("notification-note", "Notifications require the packaged app. Request permission may show the macOS authorization prompt."),
		{ID: "modal-button", Kind: ui.Button, Text: "Show modal dialog", Activate: func() { d.modal = true; d.refresh() }},
		{ID: "drop-target", Kind: ui.Column, MinHeight: 65, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))}, FileDrop: func(e ui.FileDropEvent) bool {
			var names []string
			for _, p := range e.Paths {
				names = append(names, filepath.Base(p))
			}
			d.serviceResult = "Dropped files: " + strings.Join(names, ", ") + " (files remain unchanged)"
			d.refresh()
			return true
		}, Children: []ui.Node{label("drop-label", "Drop files from Finder here to test native file-drop routing.")}},
		{ID: "cancel-service", Kind: ui.Button, Text: "Cancel current operation", Disabled: !d.busy, Activate: func() {
			if d.cancelService != nil {
				d.cancelService()
			}
		}},
		label("service-output", d.serviceResult),
	}}
}
func (d *demo) commands() (*command.Registry, error) {
	r := command.New()
	for _, c := range []command.Command{
		{ID: "show-services", Label: "macOS Services", Execute: func(context.Context) error { d.tab = "services"; d.refresh(); return nil }},
		{ID: "show-editor", Label: "Document Editor", Execute: func(context.Context) error { d.tab = "editing"; d.refresh(); return nil }},
		{ID: "show-modal", Label: "Show Modal Dialog", Execute: func(context.Context) error { d.modal = true; d.refresh(); return nil }},
		{ID: "increment-demo", Label: "Increment Counter", Available: func() error {
			if !d.enabled {
				return fmt.Errorf("counter disabled")
			}
			return nil
		}, Execute: func(context.Context) error { d.count++; d.tab = "overview"; d.refresh(); return nil }},
	} {
		if err := r.Register(c); err != nil {
			return nil, err
		}
	}
	return r, nil
}
