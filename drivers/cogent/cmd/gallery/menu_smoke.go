package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
)

// checkMenu drives actual Cocoa selectors in one bounded process. Owner updates
// while the popup is modal deliberately exercise its custom-wake route.
func checkMenu(ctx context.Context, a *app.Application, g *gallery) error {
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	find := func(s cogent.AccessibilityInspection, domain, role string) (cogent.AccessibilityElement, bool) {
		for _, e := range s.Elements {
			if e.DomainID == domain && e.Role == role {
				return e, true
			}
		}
		return cogent.AccessibilityElement{}, false
	}
	await := func(open, enabled bool) (cogent.AccessibilityInspection, error) {
		for {
			s, err := cogent.InspectAccessibility(ctx, "gallery")
			root, ok := find(s, "menu-actions", "AXPopUpButton")
			chosen, item := find(s, "second", "AXMenuItem")
			geometryReady := !open || (item && chosen.Frame[2] > 0 && chosen.Frame[3] > 0 && chosen.HitTestID == chosen.ID)
			if err == nil && s.Reachable && s.NativeWindow && ok && root.Expanded == open && root.Enabled == enabled && item == open && geometryReady && root.Frame[2] > 0 && root.Frame[3] > 0 {
				return s, nil
			}
			select {
			case <-ctx.Done():
				_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.menu.timeout", "wanted_open": open, "wanted_enabled": enabled, "snapshot": s})
				return s, errors.Join(ctx.Err(), err)
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
	act := func(id, kind string) error {
		ok, err := cogent.PerformAccessibilityAction(ctx, "gallery", id, kind, "")
		if err != nil || !ok {
			return errors.Join(err, fmt.Errorf("%s action refused", kind))
		}
		return nil
	}
	first, err := await(false, true)
	if err != nil {
		return fmt.Errorf("initial menu: %w", err)
	}
	anchor, _ := find(first, "menu-actions", "AXPopUpButton")
	var previous uint64
	for cycle := 1; cycle <= 20; cycle++ {
		if err := act(anchor.ID, "expand"); err != nil {
			return fmt.Errorf("cycle %d expand: %w", cycle, err)
		}
		s, err := await(true, true)
		if err != nil {
			return fmt.Errorf("cycle %d open publication: %w", cycle, err)
		}
		current, _ := find(s, "menu-actions", "AXPopUpButton")
		chosen, _ := find(s, "second", "AXMenuItem")
		blocked, ok := find(s, "blocked", "AXMenuItem")
		if current.ObjectIdentity != anchor.ObjectIdentity || !ok || blocked.Enabled || chosen.Name != "Run action" || chosen.Frame[2] <= 0 || chosen.Frame[3] <= 0 || chosen.HitTestID != chosen.ID || chosen.ObjectIdentity == previous {
			return fmt.Errorf("cycle %d popup semantics/geometry/identity: %+v", cycle, chosen)
		}
		previous = chosen.ObjectIdentity
		admitted, err := cogent.PerformAccessibilityAction(ctx, "gallery", blocked.ID, "press", "")
		if err != nil || admitted {
			return errors.Join(err, errors.New("disabled menu item admitted"))
		}
		if err := act(chosen.ID, "press"); err != nil {
			return fmt.Errorf("cycle %d choose: %w", cycle, err)
		}
		if _, err = await(false, true); err != nil {
			return fmt.Errorf("cycle %d close: %w", cycle, err)
		}
		count, last := 0, ""
		if err := a.Do(func() { count = g.menuCount; last = g.menuLast }); err != nil {
			return err
		}
		if count != cycle || last != "second" {
			return fmt.Errorf("stable action failed: count %d last %q", count, last)
		}
		admitted, err = cogent.PerformAccessibilityAction(ctx, "gallery", chosen.ID, "press", "")
		if err != nil || admitted {
			return errors.Join(err, errors.New("retired item admitted"))
		}
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.menu.sample", "cycle": cycle, "snapshot": s}); err != nil {
			return err
		}
	}
	if err := act(anchor.ID, "expand"); err != nil {
		return err
	}
	if _, err := await(true, true); err != nil {
		return err
	}
	if err := a.PostWindow(ref, func() { g.menuDisabled = true; _ = a.Invalidate("gallery") }); err != nil {
		return err
	}
	if _, err := await(false, false); err != nil {
		return fmt.Errorf("modal wake/disable: %w", err)
	}
	if err := a.PostWindow(ref, func() { g.menuDisabled = false; _ = a.Invalidate("gallery") }); err != nil {
		return err
	}
	if _, err := await(false, true); err != nil {
		return err
	}
	if err := act(anchor.ID, "expand"); err != nil {
		return err
	}
	if _, err := await(true, true); err != nil {
		return err
	}
	if err := act(anchor.ID, "collapse"); err != nil {
		return err
	}
	if _, err := await(false, true); err != nil {
		return fmt.Errorf("AX collapse wake: %w", err)
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.menu.result", "result": "PASS_NATIVE_MENU_CYCLES", "selection_cycles": 20, "stable_action_id": "second", "disabled_action_refused": true, "retired_action_refused": true, "modal_owner_disable_completed": true, "cocoa_collapse_completed": true, "human_voiceover_qualified": false, "physical_input_qualified": false, "preferences_changed": false})
}
