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
	"os"
	"slices"
	"sync"
	"time"
)

const windowID = "tab-close-check"

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
	if err = a.Open(app.Window{ID: windowID, Title: "Vellum document-tab check", Width: 720, Height: 480, View: m.view}); err != nil {
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
		err = errors.New("tab-close worker did not join")
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 {
		err = errors.Join(err, errors.New("tab-close native cleanup incomplete"))
	}
	outcome := "FAILED_NATIVE_TAB_CLOSE"
	if errors.Join(runErr, err) == nil {
		outcome = "PASS_NATIVE_TAB_CLOSE"
	}
	encoded := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.tab-close.result", "result": outcome, "native_cleanup": cleanup, "preferences_changed": false, "clipboard_access": false, "physical_input_qualified": false, "human_voiceover_qualified": false})
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
func element(s cogent.AccessibilityInspection, domain, role string) (cogent.AccessibilityElement, bool) {
	for _, e := range s.Elements {
		if e.DomainID == domain && e.Role == role {
			return e, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func closeElement(s cogent.AccessibilityInspection, page string) (cogent.AccessibilityElement, bool) {
	for _, e := range s.Elements {
		if e.DomainID == page && e.Role == "AXButton" && e.Name == map[string]string{"a": "Close First", "b": "Close Second"}[page] {
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
	var last cogent.AccessibilityInspection
	expectedDocuments := []string{"a", "b"}
	selected := "a"
	draft := "Original draft"
	requests, edits, decisions := 0, 0, 0
	reject, disabled := true, false
	await := func(inner context.Context) (bool, error) {
		accepted := false
		if err := owner(inner, a, func() error {
			if m.failure != nil {
				return m.failure
			}
			if m.requests > requests || m.selections != 0 || m.edits > edits || m.decisions > decisions {
				return fmt.Errorf("unexpected callback counts: close=%d select=%d edit=%d decision=%d", m.requests, m.selections, m.edits, m.decisions)
			}
			accepted = m.requests == requests && m.edits == edits && m.decisions == decisions && slices.Equal(m.documents, expectedDocuments) && m.selected == selected && m.draft == draft && m.reject == reject && m.disabled == disabled
			return nil
		}); err != nil {
			return false, err
		}
		s, err := cogent.InspectAccessibility(inner, windowID)
		if err != nil {
			return false, err
		}
		last = s
		decision, decisionExists := element(s, "reject", "AXCheckBox")
		ready := accepted && s.Reachable && s.NativeWindow && decisionExists && decision.Value == reject && decision.Enabled && decision.ObjectIdentity > 0 && decision.Frame[2] > 0 && decision.Frame[3] > 0
		for _, page := range []string{"a", "b"} {
			c, present := closeElement(s, page)
			exists := slices.Contains(expectedDocuments, page)
			ready = ready && present == exists
			if exists {
				tab, ok := element(s, page, "AXRadioButton")
				ready = ready && ok && tab.Selected == (page == selected) && tab.Frame[2] > 0 && tab.Frame[3] > 0 && c.Parent == tab.ID && c.ObjectIdentity > 0 && c.Enabled == !disabled && c.Frame[2] > 0 && c.Frame[3] > 0 && c.HitTestID == c.ID
			}
		}
		emptyElement, empty := element(s, "empty", "AXStaticText")
		ready = ready && (empty == (len(expectedDocuments) == 0))
		if empty {
			ready = ready && emptyElement.ObjectIdentity > 0 && emptyElement.Frame[2] > 0 && emptyElement.Frame[3] > 0
		}
		if slices.Contains(expectedDocuments, "a") && selected == "a" {
			field, ok := element(s, "draft", "AXTextField")
			ready = ready && ok && field.Value == draft && field.ObjectIdentity > 0 && field.Frame[2] > 0 && field.Frame[3] > 0
		}
		return ready, nil
	}
	if err := wait(ctx, await); err != nil {
		return fmt.Errorf("initial native tabs: %w", err)
	}
	oldA, _ := closeElement(last, "a")
	oldB, _ := closeElement(last, "b")
	originalField, _ := element(last, "draft", "AXTextField")
	stage := func(name string, action func(context.Context, cogent.AccessibilityInspection) (bool, error)) error {
		// Reissue only after an admitted proposal drains without owner acceptance.
		// Once the exact counters/state match, wait for native publication alone.
		admitted := 0
		err := wait(ctx, func(inner context.Context) (bool, error) {
			ready, err := await(inner)
			if err != nil {
				return false, err
			}
			if ready {
				if admitted == 0 {
					return false, errors.New("stage completed without native admission")
				}
				return true, nil
			}
			accepted := false
			if err = owner(inner, a, func() error {
				accepted = m.requests == requests && m.edits == edits && m.decisions == decisions && slices.Equal(m.documents, expectedDocuments) && m.selected == selected && m.draft == draft && m.reject == reject && m.disabled == disabled
				return nil
			}); err != nil {
				return false, err
			}
			if accepted {
				return false, nil
			}
			ok, err := action(inner, last)
			if err != nil {
				return false, err
			}
			if ok {
				admitted++
				if err = owner(inner, a, func() error { return m.failure }); err != nil {
					return false, err
				}
			}
			return false, nil
		})
		recordErr := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.tab-close.step", "stage": name, "admissions": admitted, "expected_close_requests": requests, "snapshot": last, "passed": err == nil})
		return errors.Join(err, recordErr)
	}
	closeAction := func(page string) func(context.Context, cogent.AccessibilityInspection) (bool, error) {
		return func(inner context.Context, s cogent.AccessibilityInspection) (bool, error) {
			e, ok := closeElement(s, page)
			if !ok {
				return false, errors.New("close target missing")
			}
			return cogent.PerformAccessibilityAction(inner, windowID, e.ID, "press", "")
		}
	}
	draft = "Unsaved draft"
	edits = 1
	if err := stage("native-draft", func(inner context.Context, s cogent.AccessibilityInspection) (bool, error) {
		e, ok := element(s, "draft", "AXTextField")
		if !ok {
			return false, errors.New("draft missing")
		}
		return cogent.PerformAccessibilityAction(inner, windowID, e.ID, "set", draft)
	}); err != nil {
		return err
	}
	for _, page := range []string{"b", "a"} {
		requests++
		if err := stage("reject-"+page, closeAction(page)); err != nil {
			return err
		}
		field, _ := element(last, "draft", "AXTextField")
		if field.ObjectIdentity != originalField.ObjectIdentity {
			return errors.New("rejected close replaced draft")
		}
	}
	reject = false
	decisions = 1
	if err := stage("acceptance-enabled", func(inner context.Context, s cogent.AccessibilityInspection) (bool, error) {
		e, ok := element(s, "reject", "AXCheckBox")
		if !ok {
			return false, errors.New("decision checkbox missing")
		}
		return cogent.PerformAccessibilityAction(inner, windowID, e.ID, "set", "false")
	}); err != nil {
		return err
	}
	requests++
	expectedDocuments = []string{"a"}
	if err := stage("accept-nonselected", closeAction("b")); err != nil {
		return err
	}
	refused := func(id string) error {
		inner, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		ok, err := cogent.PerformAccessibilityAction(inner, windowID, id, "press", "")
		if err != nil {
			return err
		}
		if ok {
			return errors.New("retired or disabled action admitted")
		}
		return owner(inner, a, func() error {
			if m.requests != requests {
				return errors.New("refused close reached owner")
			}
			return m.failure
		})
	}
	if err := refused(oldB.ID); err != nil {
		return err
	}
	transition := func(fn func()) error {
		inner, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		if err := owner(inner, a, func() error { fn(); return a.Invalidate(windowID) }); err != nil {
			return err
		}
		return wait(inner, await)
	}
	disabled = true
	if err := transition(func() { m.disabled = true }); err != nil {
		return err
	}
	if err := refused(oldA.ID); err != nil {
		return err
	}
	disabled = false
	if err := transition(func() { m.disabled = false }); err != nil {
		return err
	}
	requests++
	expectedDocuments = nil
	if err := stage("accept-last", closeAction("a")); err != nil {
		return err
	}
	if err := refused(oldA.ID); err != nil {
		return err
	}
	expectedDocuments = []string{"a", "b"}
	selected = "a"
	if err := transition(func() { m.documents = []string{"a", "b"}; m.selected = "a" }); err != nil {
		return err
	}
	nextA, _ := closeElement(last, "a")
	nextB, _ := closeElement(last, "b")
	if nextA.ObjectIdentity == oldA.ObjectIdentity || nextB.ObjectIdentity == oldB.ObjectIdentity {
		return errors.New("removed native close object identity reused")
	}
	requests++
	expectedDocuments = []string{"b"}
	selected = "b"
	if err := stage("accept-selected", closeAction("a")); err != nil {
		return err
	}
	requests++
	expectedDocuments = nil
	if err := stage("accept-final-reopened", closeAction("b")); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.tab-close.completed", "close_requests": requests, "native_text_edits": edits, "native_decision_changes": decisions, "selection_callbacks": 0, "rejects_preserve_native_draft_identity": true, "retired_absent_actions_refused": true, "disabled_action_refused": true, "reopened_native_identities_replaced": true, "physical_input_qualified": false, "human_voiceover_qualified": false})
}
