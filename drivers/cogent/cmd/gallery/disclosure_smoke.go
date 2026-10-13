package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dshills/vellum/drivers/cogent"
)

// This bounded worker uses actual Cocoa selectors. It never synthesizes global
// input or changes OS preferences and does not claim human VoiceOver acceptance.
func checkDisclosure(ctx context.Context) error {
	find := func(s cogent.AccessibilityInspection, domain, role string) (cogent.AccessibilityElement, bool) {
		for _, e := range s.Elements {
			if e.DomainID == domain && e.Role == role {
				return e, true
			}
		}
		return cogent.AccessibilityElement{}, false
	}
	await := func(expanded bool) (cogent.AccessibilityInspection, error) {
		for {
			s, err := cogent.InspectAccessibility(ctx, "gallery")
			if err == nil {
				heading, ok := find(s, "details", "AXDisclosureTriangle")
				_, field := find(s, "project-name", "AXTextField")
				if s.Reachable && s.NativeWindow && ok && heading.Expanded == expanded && field == expanded && heading.Frame[2] > 0 && heading.Frame[3] > 0 {
					return s, nil
				}
			}
			select {
			case <-ctx.Done():
				return s, errors.Join(ctx.Err(), err)
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
	first, err := await(true)
	if err != nil {
		return fmt.Errorf("initial disclosure: %w", err)
	}
	heading, _ := find(first, "details", "AXDisclosureTriangle")
	field, _ := find(first, "project-name", "AXTextField")
	for cycle := 1; cycle <= 20; cycle++ {
		for _, action := range []struct {
			kind     string
			expanded bool
		}{{"collapse", false}, {"expand", true}} {
			accepted, err := cogent.PerformAccessibilityAction(ctx, "gallery", heading.ID, action.kind, "")
			if err != nil || !accepted {
				return fmt.Errorf("cycle%d %s admission: %w", cycle, action.kind, errors.Join(err, errors.New("action refused")))
			}
			s, err := await(action.expanded)
			if err != nil {
				return fmt.Errorf("cycle%d %s publication: %w", cycle, action.kind, err)
			}
			current, _ := find(s, "details", "AXDisclosureTriangle")
			if current.ObjectIdentity != heading.ObjectIdentity || current.ID != heading.ID {
				return errors.New("heading identity changed across collapse")
			}
			if !action.expanded {
				accepted, err := cogent.PerformAccessibilityAction(ctx, "gallery", field.ID, "set", "hidden write")
				if err != nil || accepted {
					return fmt.Errorf("hidden field admitted stale action: %w", errors.Join(err, errors.New("unexpected admission")))
				}
			} else {
				current, _ := find(s, "project-name", "AXTextField")
				if current.Value != "Vellum" {
					return errors.New("hidden action changed controlled text")
				}
			}
			if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.disclosure.sample", "cycle": cycle, "action": action.kind, "snapshot": s}); err != nil {
				return err
			}
		}
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.disclosure.result", "result": "PASS_NATIVE_DISCLOSURE_CYCLES", "cycles": 20, "actions": 40, "stable_heading_identity": true, "hidden_field_refused": true, "human_voiceover_qualified": false, "physical_input_qualified": false, "preferences_changed": false})
}
