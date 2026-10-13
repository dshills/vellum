//go:build darwin && cgo && !offscreen && vellum_native_checks

// notifications-smoke checks a private bundle without requesting permission,
// installing a notification delegate, playing sound or changing OS settings.
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type config struct {
	nonce, previous, bundleID string
	unbundled                 bool
}
type outcome struct {
	report  map[string]any
	service app.NotificationServices
	err     error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func validNonce(value string) bool {
	data, err := hex.DecodeString(value)
	return err == nil && len(data) == 32 && hex.EncodeToString(data) == value
}
func (c config) validate() error {
	if !validNonce(c.nonce) || c.previous != "" && (!validNonce(c.previous) || c.previous == c.nonce) {
		return errors.New("notification fixture requires distinct lowercase 32-byte nonces")
	}
	if c.unbundled {
		if c.bundleID != "" || c.previous != "" {
			return errors.New("unbundled refusal cannot name a bundle or previous run")
		}
	} else if !strings.HasPrefix(c.bundleID, "com.vellum.native-checks.notifications.") || len(c.bundleID) > 128 || strings.Trim(c.bundleID, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-") != "" {
		return errors.New("notification fixture requires an isolated check bundle ID")
	}
	return nil
}
func (c config) ids() []string {
	ids := []string{c.nonce + ".one", c.nonce + ".two"}
	if c.previous != "" {
		ids = append(ids, c.previous+".one", c.previous+".two")
	}
	return ids
}
func run() error {
	var cfg config
	flag.StringVar(&cfg.nonce, "nonce", "", "new lowercase 32-byte fixture nonce")
	flag.StringVar(&cfg.previous, "previous-nonce", "", "previous completed run in this same private bundle")
	flag.StringVar(&cfg.bundleID, "bundle-id", "", "exact private bundle identifier")
	flag.BoolVar(&cfg.unbundled, "expect-unbundled", false, "require a bare executable and unavailable native operations")
	flag.Parse()
	if err := cfg.validate(); err != nil {
		return err
	}
	metadata, err := cogent.NativeNotificationCheckBundle()
	if err != nil {
		return err
	}
	if cfg.unbundled && metadata.AppBundle || !cfg.unbundled && (!metadata.AppBundle || metadata.BundleIdentifier != cfg.bundleID) {
		return fmt.Errorf("unexpected fixture bundle: %+v", metadata)
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	remove, err := cogent.ObserveNativeServiceCleanup()
	if err != nil {
		return err
	}
	defer remove()
	if err = a.Open(app.Window{ID: "notifications", Title: "Vellum notification acceptance", Width: 420, Height: 180, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Private notification check; no permission request"}}}
	}}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	workerCtx, workerCancel := context.WithTimeout(ctx, 45*time.Second)
	defer workerCancel()
	finished := make(chan outcome, 1)
	failure := make(chan error, 1)
	go func() { finished <- check(workerCtx, a, cfg, metadata); a.Close() }()
	runErr := a.Run(ctx, cogent.Driver{OnError: func(err error) {
		select {
		case failure <- err:
		default:
		}
	}})
	// A normal close follows the buffered worker result. An unexpected loop
	// exit cannot guarantee native cleanup, but must still preserve its report.
	// Wait longer than the worker's fresh 15-second cleanup deadline.
	wait := time.NewTimer(16 * time.Second)
	defer wait.Stop()
	result := awaitOutcome(finished, workerCancel, wait.C, fixtureReport(cfg, metadata))
	cancel()
	select {
	case err = <-failure:
		result.err = errors.Join(result.err, err)
	default:
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 {
		result.err = errors.Join(result.err, errors.New("notification native terminal callbacks did not retire"))
	}
	if result.service != nil {
		if _, err = result.service.NotificationSettings(context.Background()); !errors.Is(err, app.ErrClosed) {
			result.err = errors.Join(result.err, fmt.Errorf("closed notification service: %v", err))
		}
	}
	result.report["native_cleanup"] = cleanup
	result.report["raw_result"] = "FAILED"
	if errors.Join(runErr, result.err) == nil {
		result.report["raw_result"] = "PASS"
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		result.err = errors.Join(result.err, err)
	}
	return errors.Join(runErr, result.err)
}

func fixtureReport(cfg config, metadata cogent.NativeNotificationBundleCheck) map[string]any {
	return map[string]any{"bundle": metadata, "nonce": cfg.nonce, "previous_nonce": cfg.previous, "permission_requested": false, "delegate_changed": false, "sound_requested": false, "os_preferences_changed": false, "human_presentation_qualified": false, "delivery_qualified": false, "withdrawal_qualified": false}
}

func awaitOutcome(finished <-chan outcome, stopWorker context.CancelFunc, deadline <-chan time.Time, fallback map[string]any) outcome {
	select {
	case result := <-finished:
		return result
	default:
	}
	stopWorker()
	var result outcome
	select {
	case result = <-finished:
	case <-deadline:
		result = outcome{report: fallback, err: errors.New("notification fixture worker did not finish after cleanup deadline")}
	}
	if result.report == nil {
		result.report = fallback
	}
	result.report["native_loop_ended_early"] = true
	result.report["owned_cleanup_verified"] = false
	result.report["delivery_qualified"], result.report["withdrawal_qualified"] = false, false
	result.report["result"] = "FAILED_NATIVE_NOTIFICATION_EARLY_LOOP_EXIT"
	result.err = errors.Join(result.err, errors.New("native loop ended before worker completion; owned cleanup is unverified"))
	return result
}

func check(ctx context.Context, a *app.Application, cfg config, metadata cogent.NativeNotificationBundleCheck) (result outcome) {
	result.report = fixtureReport(cfg, metadata)
	for {
		s, err := cogent.InspectAccessibility(ctx, "notifications")
		if err == nil && s.Reachable && len(s.Elements) > 0 && s.Elements[0].Frame[2] > 0 {
			break
		}
		select {
		case <-ctx.Done():
			result.err = ctx.Err()
			return
		case <-time.After(15 * time.Millisecond):
		}
	}
	service, err := a.Notifications()
	if err != nil {
		result.err = err
		return
	}
	result.service = service
	capabilities := service.NotificationCapabilities()
	if !capabilities.Permissions || !capabilities.LocalDelivery || !capabilities.Withdrawal {
		result.err = errors.New("native notification transport capability missing")
		return
	}
	ids := cfg.ids()
	notes := []app.Notification{{ID: ids[0], Title: "Vellum check 🙂", Subtitle: "private fixture", Body: "line one\n\"line two\" " + cfg.nonce}, {ID: ids[1], Title: "Vellum preserved second", Body: cfg.nonce}}
	if cfg.unbundled {
		_, settingsErr := service.NotificationSettings(ctx)
		// Metadata has already proved this is not an .app. Production admission
		// refuses before accessing the center or requesting permission.
		_, permissionErr := service.RequestNotificationPermission(ctx)
		notifyErr := service.Notify(ctx, notes[0])
		withdrawErr := service.WithdrawNotifications(ctx, ids)
		_, inventoryErr := cogent.InspectNativeNotifications(ctx, ids)
		for _, refusal := range []error{settingsErr, permissionErr, notifyErr, withdrawErr, inventoryErr} {
			if !errors.Is(refusal, app.ErrUnavailable) {
				result.err = fmt.Errorf("bare executable service refusal: %v", refusal)
				return
			}
		}
		result.report["result"] = "PASS_NATIVE_NOTIFICATION_UNBUNDLED_REFUSAL"
		return
	}
	// Cleanup stays inside the live loop with its own deadline. Empty inventory
	// is accepted only after all earlier native submissions have terminal ACKs.
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		state, cleanupErr := clean(cleanupCtx, service, ids)
		result.report["owned_cleanup"] = state
		result.report["owned_cleanup_verified"] = cleanupErr == nil
		result.err = errors.Join(result.err, cleanupErr)
	}()
	settings, err := service.NotificationSettings(ctx)
	if err != nil {
		result.err = fmt.Errorf("packaged settings: %w", err)
		return
	}
	result.report["settings_before"] = settings
	initial, err := cogent.InspectNativeNotifications(ctx, ids)
	if err != nil || len(initial.Pending)+len(initial.Delivered) != 0 || initial.NativeOutstanding != 0 {
		result.err = errors.Join(err, errors.New("private current/previous notification IDs were not initially empty"))
		return
	}
	result.report["initial_inventory"] = initial
	if settings.Permission == app.NotificationNotDetermined || settings.Permission == app.NotificationDenied {
		for _, n := range notes {
			if err = service.Notify(ctx, n); !errors.Is(err, app.ErrNotificationPermission) {
				result.err = fmt.Errorf("nonprompt permission refusal: %v", err)
				return
			}
		}
		after, e := service.NotificationSettings(ctx)
		state, inventoryErr := cogent.InspectNativeNotifications(ctx, ids)
		if e != nil || inventoryErr != nil || after != settings || len(state.Pending)+len(state.Delivered) != 0 || state.NativeOutstanding != 0 {
			result.err = errors.Join(e, inventoryErr, errors.New("Notify changed permissions or retained an unauthorized item"))
			return
		}
		result.report["settings_after"], result.report["after_refusal_inventory"] = after, state
		result.report["result"] = "PASS_NATIVE_NOTIFICATION_PERMISSION_REFUSAL"
		return
	}
	if settings.Permission != app.NotificationAuthorized && settings.Permission != app.NotificationProvisional {
		result.err = errors.New("unknown native notification authorization")
		return
	}
	for _, n := range notes {
		if err = service.Notify(ctx, n); err != nil {
			result.err = err
			return
		}
	}
	state, err := waitInventory(ctx, ids, func(s cogent.NativeNotificationInventoryCheck) bool { return matches(s, notes) })
	if err != nil {
		result.err = fmt.Errorf("accepted delivery was not observed: %w", err)
		return
	}
	result.report["delivered_inventory"] = state
	notes[0].Title, notes[0].Body = "Vellum replacement", "updated "+cfg.nonce
	if err = service.Notify(ctx, notes[0]); err != nil {
		result.err = err
		return
	}
	state, err = waitInventory(ctx, ids, func(s cogent.NativeNotificationInventoryCheck) bool { return matches(s, notes) })
	if err != nil {
		result.err = fmt.Errorf("replacement inventory: %w", err)
		return
	}
	result.report["replacement_inventory"] = state
	if err = service.WithdrawNotifications(ctx, ids[:1]); err != nil {
		result.err = err
		return
	}
	state, err = waitInventory(ctx, ids, func(s cogent.NativeNotificationInventoryCheck) bool { return matches(s, notes[1:]) })
	if err != nil {
		result.err = fmt.Errorf("targeted withdrawal did not preserve second item: %w", err)
		return
	}
	after, err := service.NotificationSettings(ctx)
	if err != nil || after != settings {
		result.err = errors.Join(err, errors.New("notification settings changed during fixture"))
		return
	}
	result.report["settings_after"], result.report["targeted_withdrawal_inventory"] = after, state
	result.report["delivery_qualified"], result.report["withdrawal_qualified"] = true, true
	result.report["result"] = "PASS_NATIVE_NOTIFICATION_DELIVERY_REPLACEMENT_WITHDRAWAL"
	return
}
func matches(s cogent.NativeNotificationInventoryCheck, notes []app.Notification) bool {
	if s.NativeOutstanding != 0 || len(s.Pending) != 0 || len(s.Delivered) != len(notes) {
		return false
	}
	for _, n := range notes {
		found := false
		for _, item := range s.Delivered {
			if item.ID == n.ID && item.Title == n.Title && item.Subtitle == n.Subtitle && item.Body == n.Body && item.Sound == n.Sound {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func waitInventory(ctx context.Context, ids []string, accept func(cogent.NativeNotificationInventoryCheck) bool) (cogent.NativeNotificationInventoryCheck, error) {
	var state cogent.NativeNotificationInventoryCheck
	for {
		current, err := cogent.InspectNativeNotifications(ctx, ids)
		if err != nil {
			return state, err
		}
		state = current
		if accept(state) {
			return state, nil
		}
		select {
		case <-ctx.Done():
			return state, ctx.Err()
		case <-time.After(30 * time.Millisecond):
		}
	}
}
func clean(ctx context.Context, service app.NotificationServices, ids []string) (cogent.NativeNotificationInventoryCheck, error) {
	state, err := waitInventory(ctx, ids, func(s cogent.NativeNotificationInventoryCheck) bool { return s.NativeOutstanding == 0 })
	if err != nil {
		return state, fmt.Errorf("submission ACKs unresolved before cleanup: %w", err)
	}
	if err = service.WithdrawNotifications(ctx, ids); err != nil {
		return state, fmt.Errorf("owned notification withdrawal: %w", err)
	}
	state, err = waitInventory(ctx, ids, func(s cogent.NativeNotificationInventoryCheck) bool {
		return s.NativeOutstanding == 0 && len(s.Pending)+len(s.Delivered) == 0
	})
	if err != nil {
		return state, fmt.Errorf("owned notification removal unproved: %w", err)
	}
	return state, nil
}
