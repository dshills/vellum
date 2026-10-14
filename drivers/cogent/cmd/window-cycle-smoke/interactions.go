//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/ui"
	"image/color"
	"os"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/acceptance"
)

type nativeActionRecord struct {
	Target, Kind, Value     string
	NativeRole              string
	Attempts, Admissions    int
	OwnerAndNativeCompleted bool
	ElapsedSeconds          float64
	AttachmentRetirement    *cogent.AccessibilityRetirement `json:",omitempty"`
}
type nativeActionHooks struct {
	targetRole string
	accepted   func(context.Context) (bool, error)
	inspect    func(context.Context) (cogent.AccessibilityInspection, error)
	perform    func(context.Context, string) (bool, error)
	completed  func(cogent.AccessibilityInspection) bool
}

// Every admitted proposal is followed by an owner acknowledgement before a
// retry. Accepted domain state waits only for publication, never repeating a
// non-idempotent press while native semantics catch up.
func runNativeAction(ctx context.Context, duration time.Duration, target, kind, value string, h nativeActionHooks) (nativeActionRecord, error) {
	started := time.Now()
	record := nativeActionRecord{Target: target, Kind: kind, Value: value}
	err := waitWithin(ctx, duration, func(inner context.Context) (bool, error) {
		accepted, err := h.accepted(inner)
		if err != nil {
			return false, err
		}
		native, err := h.inspect(inner)
		if err != nil {
			return false, err
		}
		if accepted {
			if record.Admissions == 0 {
				return false, errors.New("domain already accepted without a native admission")
			}
			record.OwnerAndNativeCompleted = native.Reachable && native.NativeWindow && h.completed(native)
			return record.OwnerAndNativeCompleted, nil
		}
		if !native.Reachable || !native.NativeWindow {
			return false, nil
		}
		for _, e := range native.Elements {
			if e.DomainID != target || h.targetRole != "" && e.Role != h.targetRole || !e.Enabled || e.Frame[2] <= 0 || e.Frame[3] <= 0 {
				continue
			}
			record.Attempts++
			record.NativeRole = e.Role
			admitted, err := h.perform(inner, e.ID)
			if err != nil {
				return false, err
			}
			if admitted {
				record.Admissions++
				// This posts behind the admitted action in the same application queue.
				// A refusal caused by intervening publication can retry after it drains.
				_, err = h.accepted(inner)
				if err != nil {
					return false, err
				}
			}
			return false, nil
		}
		return false, nil
	})
	record.ElapsedSeconds = time.Since(started).Seconds()
	if err != nil {
		var closed *cogent.AccessibilityClosedError
		if errors.As(err, &closed) {
			retirement := closed.Retirement
			record.AttachmentRetirement = &retirement
		}
		return record, fmt.Errorf("native %s %s: %w (attempts=%d admissions=%d)", target, kind, err, record.Attempts, record.Admissions)
	}
	return record, nil
}

func nativeElement(s cogent.AccessibilityInspection, domain string) (cogent.AccessibilityElement, bool) {
	return nativeElementRole(s, domain, "")
}
func nativeElementRole(s cogent.AccessibilityInspection, domain, role string) (cogent.AccessibilityElement, bool) {
	for _, e := range s.Elements {
		if e.DomainID == domain && (role == "" || e.Role == role) {
			return e, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func exerciseNativeActions(ctx context.Context, a *app.Application, id string, w *acceptance.Workspace) ([]nativeActionRecord, error) {
	records := []nativeActionRecord{}
	var geometry cogent.TreeInspection
	run := func(target, kind, value string, accepted func() (bool, error), completed func(cogent.AccessibilityInspection) bool) error {
		role := ""
		if target == "disclosure" {
			role = "AXDisclosureTriangle"
		}
		record, err := runNativeAction(ctx, 5*time.Second, target, kind, value, nativeActionHooks{
			targetRole: role,
			accepted: func(inner context.Context) (bool, error) {
				ready := false
				reply, err := ownerProbe(inner, a, func() ownerReply {
					if state := w.State(); state.Failure != nil {
						return ownerReply{err: state.Failure}
					}
					var err error
					ready, err = accepted()
					return ownerReply{err: err}
				})
				return ready, errors.Join(err, reply.err)
			},
			inspect: func(inner context.Context) (cogent.AccessibilityInspection, error) {
				return cogent.InspectAccessibility(inner, id)
			},
			perform: func(inner context.Context, nativeID string) (bool, error) {
				return cogent.PerformAccessibilityAction(inner, id, nativeID, kind, value)
			},
			completed: completed,
		})
		records = append(records, record)
		return errors.Join(err, json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.native-action", "window": id, "action": record}))
	}
	valueMatches := func(target string, value any) func(cogent.AccessibilityInspection) bool {
		return func(s cogent.AccessibilityInspection) bool {
			e, ok := nativeElement(s, target)
			return ok && e.Enabled && e.Frame[2] > 0 && e.Frame[3] > 0 && e.Value == value
		}
	}
	changed := func(target string, count int, predicate func(acceptance.ControlState) bool) func() (bool, error) {
		return func() (bool, error) {
			s := w.State()
			if s.Changes[target] > count {
				return false, fmt.Errorf("%s callback repeated: %d > %d", target, s.Changes[target], count)
			}
			return s.Changes[target] == count && predicate(s), nil
		}
	}
	tests := []struct {
		target, kind, value string
		count               int
		accepted            func(acceptance.ControlState) bool
		nativeValue         any
	}{
		{"check", "set", "false", 1, func(s acceptance.ControlState) bool { return !s.Checked }, false},
		{"switch", "press", "", 1, func(s acceptance.ControlState) bool { return s.Checked }, true},
		{"toggle", "press", "", 1, func(s acceptance.ControlState) bool { return !s.Checked }, false},
		{"check", "set", "true", 2, func(s acceptance.ControlState) bool { return s.Checked }, true},
		{"slider", "set", "37", 1, func(s acceptance.ControlState) bool { return s.Number == 37 }, float64(37)},
		{"number", "increment", "", 1, func(s acceptance.ControlState) bool { return s.Number == 38 }, float64(38)},
		{"radio", "set", "a", 1, func(s acceptance.ControlState) bool { return s.Choice == "a" }, "A"},
		{"segment", "set", "b", 1, func(s acceptance.ControlState) bool { return s.Choice == "b" }, "B"},
		{"selector", "set", "a", 1, func(s acceptance.ControlState) bool { return s.Choice == "a" }, "A"},
	}
	for _, test := range tests {
		if err := run(test.target, test.kind, test.value, changed(test.target, test.count, test.accepted), valueMatches(test.target, test.nativeValue)); err != nil {
			return records, err
		}
	}
	if err := run("button", "press", "", func() (bool, error) {
		s := w.State()
		if s.Actions > 2 {
			return false, errors.New("native button repeated")
		}
		return s.Actions == 2, nil
	}, func(s cogent.AccessibilityInspection) bool {
		e, ok := nativeElement(s, "button")
		return ok && e.Enabled && e.Name == "Action"
	}); err != nil {
		return records, err
	}
	for i, open := range []bool{false, true} {
		kind := "collapse"
		if open {
			kind = "expand"
		}
		if err := run("disclosure", kind, "", changed("disclosure", i+1, func(s acceptance.ControlState) bool { return s.IndicatorsOpen == open }), func(s cogent.AccessibilityInspection) bool {
			e, ok := nativeElementRole(s, "disclosure", "AXDisclosureTriangle")
			return ok && e.Expanded == open
		}); err != nil {
			return records, err
		}
	}
	for i, tab := range []string{"about", "live"} {
		label := "About"
		if tab == "live" {
			label = "Workspace"
		}
		if err := run("tabs", "set", tab, changed("tabs", i+1, func(s acceptance.ControlState) bool { return s.Tab == tab }), valueMatches("tabs", label)); err != nil {
			return records, err
		}
	}

	if err := run("tabs", "set", "colors", changed("tabs", 3, func(s acceptance.ControlState) bool { return s.Tab == "colors" }), valueMatches("tabs", "Colors")); err != nil {
		return records, err
	}
	colorCompleted := func(want color.NRGBA) func(cogent.AccessibilityInspection) bool {
		return func(s cogent.AccessibilityInspection) bool { return colorEnvelopeCompleted(s, want) }
	}

	red := color.NRGBA{R: 12, G: 96, B: 224, A: 255}
	redID := ui.ColorFieldID("color", ui.ColorRed)
	if err := run(redID, "set", "12", changed(redID, 1, func(s acceptance.ControlState) bool { return s.Color == red }), colorCompleted(red)); err != nil {
		return records, err
	}
	hex := color.NRGBA{R: 53, G: 103, B: 181, A: 128}
	hexID := ui.ColorFieldID("color", ui.ColorHex)
	if err := run(hexID, "set", "#3567B580", changed(hexID, 1, func(s acceptance.ControlState) bool { return s.Color == hex }), colorCompleted(hex)); err != nil {
		return records, err
	}
	if err := run("tabs", "set", "live", changed("tabs", 4, func(s acceptance.ControlState) bool { return s.Tab == "live" }), valueMatches("tabs", "Workspace")); err != nil {
		return records, err
	}

	ownerTreeCompleted := false
	treeReady := func(selected string, disclosures, visible int, expanded bool) (bool, error) {
		ownerTreeCompleted = false
		s := w.State()
		if s.TreeSelections > 1 || s.Disclosures > disclosures {
			return false, errors.New("native tree callback repeated")
		}
		state, err := w.Tree.State()
		if err != nil {
			return false, err
		}
		accepted := state.Selected == selected && s.TreeSelections == 1 && s.Disclosures == disclosures
		geometry, err = cogent.InspectTree(a, w.Tree)
		if errors.Is(err, cogent.ErrTreeLayout) {
			return accepted, nil
		}
		if err != nil {
			return false, err
		}
		if accepted && !state.Pending && !geometry.Pending && geometry.Selected == selected && geometry.VisibleRows == visible && s.WorkerActive == 0 {
			for _, row := range geometry.Rows {
				if row.Key == "n0" {
					ownerTreeCompleted = row.HasChildren && row.Expanded == expanded
					break
				}
			}
		}
		// Callback/request acceptance is independent of asynchronous completion.
		return accepted, nil
	}
	nativeTreeReady := func(s cogent.AccessibilityInspection) bool {
		ready, _ := treeEnvelopeCompleted(geometry, s)
		return ownerTreeCompleted && ready
	}
	rowID := cogent.TreeRowID("outline", "n0")
	if err := run(rowID, "press", "", func() (bool, error) { return treeReady("n0", 0, 64, true) }, nativeTreeReady); err != nil {
		return records, err
	}
	if err := run(rowID, "collapse", "", func() (bool, error) { return treeReady("n0", 1, 60, false) }, nativeTreeReady); err != nil {
		return records, err
	}
	if err := run(rowID, "expand", "", func() (bool, error) { return treeReady("n0", 2, 64, true) }, nativeTreeReady); err != nil {
		return records, err
	}
	if err := run("text", "set", "Cocoa composed edit", changed("text", 1, func(acceptance.ControlState) bool { return w.Text() == "Cocoa composed edit" }), valueMatches("text", "Cocoa composed edit")); err != nil {
		return records, err
	}
	for _, open := range []bool{true, false} {
		target := "dialog-open"
		if !open {
			target = "dialog-close"
		}
		if err := run(target, "press", "", func() (bool, error) { return w.State().DialogOpen == open, nil }, func(s cogent.AccessibilityInspection) bool {
			_, present := nativeElement(s, "dialog-close")
			return present == open
		}); err != nil {
			return records, err
		}
	}
	return records, nil
}

func colorEnvelopeCompleted(s cogent.AccessibilityInspection, want color.NRGBA) bool {
	if !s.Reachable || !s.NativeWindow {
		return false
	}
	values := []any{float64(want.R), float64(want.G), float64(want.B), float64(want.A), ui.ColorHexString(want)}
	for i, c := range []ui.ColorChannel{ui.ColorRed, ui.ColorGreen, ui.ColorBlue, ui.ColorAlpha, ui.ColorHex} {
		e, ok := nativeElement(s, ui.ColorFieldID("color", c))
		role := "AXSlider"
		if c == ui.ColorHex {
			role = "AXTextField"
		}
		if !ok || e.ObjectIdentity == 0 || e.Role != role || !e.Enabled || e.Frame[2] <= 0 || e.Frame[3] <= 0 || e.Value != values[i] {
			return false
		}
	}
	e, ok := nativeElement(s, "color/color/preview")
	return ok && e.ObjectIdentity > 0 && e.Role == "AXImage" && e.Name == "Workspace color preview "+ui.ColorHexString(want) && e.Frame[2] > 0 && e.Frame[3] > 0
}
