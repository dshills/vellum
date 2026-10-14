package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func colorOwner(ctx context.Context, a *app.Application, fn func() error) error {
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
		active := phase == 1 || phase == 3
		mu.Unlock()
		if active {
			return errors.Join(ctx.Err(), <-reply)
		}
		return ctx.Err()
	}
}
func colorElement(s cogent.AccessibilityInspection, id string) (cogent.AccessibilityElement, bool) {
	for _, e := range s.Elements {
		if e.DomainID == id {
			return e, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func checkColor(ctx context.Context, a *app.Application, g *gallery) error {
	identities := map[string]uint64{}
	fields := []ui.ColorChannel{ui.ColorRed, ui.ColorGreen, ui.ColorBlue, ui.ColorAlpha, ui.ColorHex}
	await := func(want color.NRGBA, edits, invalid int, enabled bool) (cogent.AccessibilityInspection, error) {
		operation, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		for {
			accepted := false
			if err := colorOwner(operation, a, func() error {
				accepted = g.colorPicker.Color == want && g.colorEdits == edits && g.colorInvalid == invalid
				return nil
			}); err != nil {
				return cogent.AccessibilityInspection{}, err
			}
			s, err := cogent.InspectAccessibility(operation, "gallery")
			ready := err == nil && s.Reachable && s.NativeWindow && accepted
			values := []any{float64(want.R), float64(want.G), float64(want.B), float64(want.A), ui.ColorHexString(want)}
			for i, field := range fields {
				e, ok := colorElement(s, ui.ColorFieldID("ink", field))
				role := "AXSlider"
				if field == ui.ColorHex {
					role = "AXTextField"
				}
				ready = ready && ok && e.ObjectIdentity > 0 && e.Role == role && e.Value == values[i] && e.Enabled == enabled && e.Frame[2] > 0 && e.Frame[3] > 0
				if ok && identities[e.DomainID] != 0 && identities[e.DomainID] != e.ObjectIdentity {
					return s, fmt.Errorf("retained color field native identity replaced: %s", e.DomainID)
				}
			}
			preview, ok := colorElement(s, "ink/color/preview")
			ready = ready && ok && preview.Role == "AXImage" && preview.Name == "Ink color preview "+ui.ColorHexString(want) && preview.Frame[2] > 0 && preview.Frame[3] > 0
			_, message := colorElement(s, "ink/color/validation")
			ready = ready && (message == (invalid > 0 && edits == 4))
			if ready {
				return s, nil
			}
			select {
			case <-operation.Done():
				return s, fmt.Errorf("color publication: %w; want=%v edits=%d invalid=%d enabled=%t", errors.Join(operation.Err(), err), want, edits, invalid, enabled)
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	current := color.NRGBA{R: 32, G: 96, B: 224, A: 255}
	first, err := await(current, 0, 0, true)
	if err != nil {
		return err
	}
	for _, field := range fields {
		e, _ := colorElement(first, ui.ColorFieldID("ink", field))
		identities[e.DomainID] = e.ObjectIdentity
	}
	step := func(channel ui.ColorChannel, kind, value string, want color.NRGBA, edits, invalid int) error {
		operation, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		// This step requires a fresh native admission. Counter observation after
		// an owner barrier prevents repeated non-idempotent increment proposals.
		for {
			s, err := cogent.InspectAccessibility(operation, "gallery")
			if err != nil {
				return err
			}
			e, ok := colorElement(s, ui.ColorFieldID("ink", channel))
			if !ok {
				return errors.New("color field absent")
			}
			admitted, err := cogent.PerformAccessibilityAction(operation, "gallery", e.ID, kind, value)
			if err != nil {
				return err
			}
			accepted := false
			if admitted {
				err = colorOwner(operation, a, func() error {
					if g.colorEdits > edits || g.colorInvalid > invalid {
						return errors.New("color callback repeated")
					}
					accepted = g.colorEdits == edits && g.colorInvalid == invalid
					return nil
				})
				if err != nil {
					return err
				}
			}
			if accepted {
				break
			}
			select {
			case <-operation.Done():
				return operation.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
		_, err := await(want, edits, invalid, true)
		if err != nil {
			return err
		}
		current = want
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.color.step", "channel": channel, "kind": kind, "proposed": value, "accepted_hex": ui.ColorHexString(want), "edit_callbacks": edits, "invalid_callbacks": invalid, "owner_and_native_completed": true})
	}
	if err = step(ui.ColorRed, "set", "12", color.NRGBA{R: 12, G: 96, B: 224, A: 255}, 1, 0); err != nil {
		return err
	}
	if err = step(ui.ColorGreen, "increment", "", color.NRGBA{R: 12, G: 97, B: 224, A: 255}, 2, 0); err != nil {
		return err
	}
	if err = step(ui.ColorAlpha, "set", "0", color.NRGBA{R: 12, G: 97, B: 224}, 3, 0); err != nil {
		return err
	}
	if err = step(ui.ColorHex, "set", "#F001AB00", color.NRGBA{R: 240, G: 1, B: 171}, 4, 0); err != nil {
		return err
	}
	if err = step(ui.ColorHex, "set", "invalid", current, 4, 1); err != nil {
		return err
	}
	if err = step(ui.ColorRed, "set", "37.4", color.NRGBA{R: 37, G: 1, B: 171}, 5, 1); err != nil {
		return err
	}
	if err = colorOwner(ctx, a, func() error { g.colorReject = true; return nil }); err != nil {
		return err
	}
	if err = step(ui.ColorRed, "set", "99", current, 6, 1); err != nil {
		return err
	}
	if err = colorOwner(ctx, a, func() error { g.colorReject = false; g.colorNormalize = true; return nil }); err != nil {
		return err
	}
	if err = step(ui.ColorAlpha, "set", "128", color.NRGBA{R: 37, G: 1, B: 171, A: 255}, 7, 1); err != nil {
		return err
	}
	if err = colorOwner(ctx, a, func() error { g.colorPicker.Disabled = true; return a.Invalidate("gallery") }); err != nil {
		return err
	}
	disabled, err := await(current, 7, 1, false)
	if err != nil {
		return err
	}
	e, _ := colorElement(disabled, ui.ColorFieldID("ink", ui.ColorRed))
	admitted, err := cogent.PerformAccessibilityAction(ctx, "gallery", e.ID, "set", "77")
	if err != nil || admitted {
		return errors.Join(err, errors.New("disabled color field admitted"))
	}
	if err = colorOwner(ctx, a, func() error {
		if g.colorEdits != 7 || g.colorPicker.Color != current {
			return errors.New("disabled request mutated color")
		}
		g.tab = "about"
		return a.Invalidate("gallery")
	}); err != nil {
		return err
	}
	operation, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	for {
		s, err := cogent.InspectAccessibility(operation, "gallery")
		if err != nil {
			return err
		}
		if _, present := colorElement(s, ui.ColorFieldID("ink", ui.ColorHex)); !present {
			break
		}
		select {
		case <-operation.Done():
			return operation.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	hex, _ := colorElement(first, ui.ColorFieldID("ink", ui.ColorHex))
	admitted, err = cogent.PerformAccessibilityAction(ctx, "gallery", hex.ID, "set", "#FFFFFFFF")
	if err != nil || admitted {
		return errors.Join(err, errors.New("hidden color field admitted stale action"))
	}
	if err = colorOwner(ctx, a, func() error {
		if g.colorEdits != 7 || g.colorPicker.Color != current {
			return errors.New("hidden field changed accepted color")
		}
		return nil
	}); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.color.result", "result": "PASS_NATIVE_COLOR_PICKER", "steps": 8, "edit_callbacks": 7, "invalid_callbacks": 1, "stable_field_identities": true, "disabled_refused": true, "hidden_refused": true, "final_hex": ui.ColorHexString(current), "physical_input_qualified": false, "human_voiceover_qualified": false, "preferences_changed": false, "clipboard_access": false})
}
