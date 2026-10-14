//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"cogentcore.org/core/core"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"os"
	"strings"
	"sync"
	"time"
)

const windowID = "breadcrumb-check"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	core.DebugSettings.DisableWindowGeometrySaver = true
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
	m := newModel()
	if err = a.Open(app.Window{ID: windowID, Title: "Vellum breadcrumb check", Width: 720, Height: 480, View: m.view}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- check(ctx, a, m); a.Close() }()
	runErr := a.Run(ctx, cogent.Driver{})
	cancel()
	select {
	case err = <-done:
	case <-time.After(6 * time.Second):
		err = errors.New("breadcrumb worker did not join")
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 {
		err = errors.Join(err, errors.New("breadcrumb native cleanup incomplete"))
	}
	outcome := "FAILED_NATIVE_BREADCRUMB"
	if errors.Join(runErr, err) == nil {
		outcome = "PASS_NATIVE_BREADCRUMB"
	}
	encoded := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.breadcrumb.result", "result": outcome, "native_cleanup": cleanup, "preferences_changed": false, "clipboard_access": false, "physical_input_qualified": false, "human_voiceover_qualified": false})
	return errors.Join(runErr, err, encoded)
}

// A FIFO acknowledgement observes all previously admitted callbacks. Cancellation
// revokes work that has not started and joins a callback that already claimed it.
func owner(ctx context.Context, a *app.Application, fn func() error) error {
	reply := make(chan error, 1)
	var mu sync.Mutex
	phase := 0
	if err := a.PostWait(ctx, func() {
		mu.Lock()
		if phase == 2 || ctx.Err() != nil {
			phase = 2
			mu.Unlock()
			reply <- ctx.Err()
			return
		}
		phase = 1
		mu.Unlock()
		err := fn()
		mu.Lock()
		phase = 3
		mu.Unlock()
		reply <- err
	}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		mu.Lock()
		if phase == 0 {
			phase = 2
			mu.Unlock()
			return ctx.Err()
		}
		running := phase == 1 || phase == 3
		mu.Unlock()
		if running {
			return errors.Join(ctx.Err(), <-reply)
		}
		return ctx.Err()
	}
}

func itemElement(s cogent.AccessibilityInspection, item string) (cogent.AccessibilityElement, bool) {
	id := ui.BreadcrumbItemID("path", item)
	for _, e := range s.Elements {
		if e.DomainID == id {
			return e, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func wait(ctx context.Context, fn func(context.Context) (bool, error)) error {
	inner, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		ok, err := fn(inner)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-inner.Done():
			return inner.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func check(ctx context.Context, a *app.Application, m *model) error {
	expected := fullPath()
	requests := 0
	disabled, readOnly, reject := false, false, true
	var last cogent.AccessibilityInspection
	await := func(inner context.Context) (bool, error) {
		accepted := false
		if err := owner(inner, a, func() error {
			if m.failure != nil {
				return m.failure
			}
			if m.requests > requests {
				return fmt.Errorf("unexpected ancestor callback count %d > %d", m.requests, requests)
			}
			accepted = m.requests == requests && m.path.Disabled == disabled && (m.path.Navigate == nil) == readOnly && m.reject == reject && len(m.path.Items) == len(expected)
			if accepted {
				for i, item := range expected {
					if m.path.Items[i] != item {
						accepted = false
						break
					}
				}
			}
			return nil
		}); err != nil {
			return false, err
		}
		s, err := cogent.InspectAccessibility(inner, windowID)
		if err != nil {
			return false, err
		}
		last = s
		ready := accepted && s.Reachable && s.NativeWindow
		for _, e := range s.Elements {
			if strings.Contains(e.DomainID, "/breadcrumb/separator/") {
				return false, errors.New("decorative breadcrumb separator announced")
			}
		}
		for i, item := range expected {
			e, ok := itemElement(s, item.ID)
			role := "AXButton"
			if readOnly || i == len(expected)-1 {
				role = "AXStaticText"
			}
			ready = ready && ok && e.Role == role && e.Name == item.Label && e.ObjectIdentity > 0 && e.Frame[2] > 0 && e.Frame[3] > 0
			if role == "AXButton" {
				ready = ready && e.Enabled == (!disabled && !item.Disabled) && e.HitTestID == e.ID
			}
		}
		for _, item := range fullPath() {
			present := false
			for _, keep := range expected {
				present = present || keep.ID == item.ID
			}
			_, actual := itemElement(s, item.ID)
			ready = ready && actual == present
		}
		return ready, nil
	}
	record := func(name string, admissions int, err error) error {
		return errors.Join(err, json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.breadcrumb.step", "stage": name, "admissions": admissions, "expected_requests": requests, "snapshot": last, "passed": err == nil}))
	}
	transition := func(name string, fn func()) error {
		err := owner(ctx, a, func() error { fn(); return a.Invalidate(windowID) })
		if err == nil {
			err = wait(ctx, await)
		}
		return record(name, 0, err)
	}
	if err := wait(ctx, await); err != nil {
		return record("initial", 0, err)
	}
	if err := record("initial", 0, nil); err != nil {
		return err
	}
	oldFolder, _ := itemElement(last, "folder")
	refuse := func(id string) error {
		ok, err := cogent.PerformAccessibilityAction(ctx, windowID, id, "press", "")
		if err != nil {
			return err
		}
		if ok {
			return errors.New("current, disabled, read-only or absent breadcrumb action admitted")
		}
		return owner(ctx, a, func() error {
			if m.requests != requests {
				return errors.New("refused breadcrumb action reached owner")
			}
			return m.failure
		})
	}
	current, _ := itemElement(last, "document")
	if err := refuse(current.ID); err != nil {
		return err
	}
	// A proposal may be admitted yet become stale before its owner turn. Retry
	// only after acknowledgement proves that no callback was accepted.
	proposal := func(name, item string) error {
		admissions := 0
		err := wait(ctx, func(inner context.Context) (bool, error) {
			ready, err := await(inner)
			if err != nil {
				return false, err
			}
			if ready {
				if admissions == 0 {
					return false, errors.New("proposal completed without native admission")
				}
				return true, nil
			}
			matched := false
			if err = owner(inner, a, func() error { matched = m.requests == requests; return m.failure }); err != nil {
				return false, err
			}
			if matched {
				return false, nil
			}
			e, ok := itemElement(last, item)
			if !ok {
				return false, errors.New("ancestor target missing")
			}
			admitted, err := cogent.PerformAccessibilityAction(inner, windowID, e.ID, "press", "")
			if err != nil {
				return false, err
			}
			if admitted {
				admissions++
				if err = owner(inner, a, func() error { return m.failure }); err != nil {
					return false, err
				}
			}
			return false, nil
		})
		return record(name, admissions, err)
	}
	requests = 1
	if err := proposal("reject-folder", "folder"); err != nil {
		return err
	}
	kept, _ := itemElement(last, "folder")
	if kept.ObjectIdentity != oldFolder.ObjectIdentity {
		return errors.New("rejected navigation replaced ancestor native identity")
	}
	disabled = true
	if err := transition("disable-path", func() { m.path.Disabled = true }); err != nil {
		return err
	}
	if err := refuse(oldFolder.ID); err != nil {
		return err
	}
	disabled = false
	if err := transition("enable-path", func() { m.path.Disabled = false }); err != nil {
		return err
	}
	// Native selector dispatch runs while this owner callback is claimed. The
	// selector posts its proposal; ordered IDs change before that proposal drains.
	// The existing native callback generation and component guard remain active.
	target, _ := itemElement(last, "home")
	expected[0], expected[1] = expected[1], expected[0]
	admitted := false
	err := owner(ctx, a, func() error {
		var err error
		admitted, err = cogent.PerformAccessibilityAction(ctx, windowID, target.ID, "press", "")
		if err != nil {
			return err
		}
		if !admitted {
			return errors.New("stale path fixture had no native admission")
		}
		m.path.Items[0], m.path.Items[1] = m.path.Items[1], m.path.Items[0]
		return a.Invalidate(windowID)
	})
	if err == nil {
		err = wait(ctx, await)
	}
	admissions := 0
	if admitted {
		admissions = 1
	}
	if err = record("stale-path", admissions, err); err != nil {
		return err
	}
	expected = fullPath()
	if err := transition("restore-path", func() { m.path.Items = fullPath() }); err != nil {
		return err
	}
	reject = false
	requests = 2
	expected = fullPath()[:2]
	if err := owner(ctx, a, func() error { m.reject = false; return nil }); err != nil {
		return err
	}
	if err := proposal("accept-folder", "folder"); err != nil {
		return err
	}
	folder, _ := itemElement(last, "folder")
	if err := refuse(folder.ID); err != nil {
		return err
	}
	requests = 3
	expected = fullPath()[:1]
	if err := proposal("accept-home", "home"); err != nil {
		return err
	}
	if err := refuse(oldFolder.ID); err != nil {
		return err
	}
	expected = fullPath()
	if err := transition("reset-path", func() { m.path.Items = fullPath() }); err != nil {
		return err
	}
	nextFolder, _ := itemElement(last, "folder")
	if nextFolder.ObjectIdentity == oldFolder.ObjectIdentity {
		return errors.New("removed ancestor native identity reused")
	}
	expected[1].Disabled = true
	if err := transition("disable-ancestor", func() { m.path.Items[1].Disabled = true }); err != nil {
		return err
	}
	if err := refuse(nextFolder.ID); err != nil {
		return err
	}
	expected[1].Disabled = false
	if err := transition("enable-ancestor", func() { m.path.Items[1].Disabled = false }); err != nil {
		return err
	}
	readOnly = true
	if err := transition("read-only", func() { m.path.Navigate = nil }); err != nil {
		return err
	}
	for _, item := range expected {
		e, _ := itemElement(last, item.ID)
		if err := refuse(e.ID); err != nil {
			return err
		}
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.breadcrumb.completed", "requests": requests, "rejected_requests": 1, "accepted_requests": 2, "stale_native_admission": admitted, "stale_path_rejected": true, "disabled_path_and_item_refused": true, "readonly_and_current_refused": true, "absent_action_refused": true, "reopened_ancestor_identity_replaced": true, "literal_current_name": "<Current>", "physical_input_qualified": false, "human_voiceover_qualified": false})
}
