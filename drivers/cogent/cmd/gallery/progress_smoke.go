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
	"github.com/dshills/vellum/ui"
)

func (g *gallery) progressView() ui.Node {
	return ui.Node{ID: "progress-root", Kind: ui.Column, Children: []ui.Node{
		{ID: "progress-title", Kind: ui.Label, Text: "Vellum progress shapes"},
		{ID: "progress-shapes", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(24))}, Children: []ui.Node{
			{ID: "progress-ring", Kind: ui.ProgressRing, Number: g.number / 100, Text: fmt.Sprintf("%.0f%%", g.number), Accessibility: ui.Accessibility{Name: "Circular operation progress"}},
			{ID: "progress-gauge", Kind: ui.ProgressGauge, Number: g.number / 100, Text: fmt.Sprintf("%.0f%%", g.number), Accessibility: ui.Accessibility{Name: "Semicircular operation progress"}},
		}},
	}}
}

func checkProgress(ctx context.Context, a *app.Application, g *gallery) error {
	identities := map[string]uint64{}
	for stage, percent := range []float32{0, 25, 75, 100} {
		step, stop := context.WithTimeout(ctx, 5*time.Second)
		// Cancellable owner assignment runs once; publication polling never repeats it.
		if err := colorOwner(step, a, func() error { g.number = percent; return a.Invalidate("gallery") }); err != nil {
			stop()
			return err
		}
		var found map[string]cogent.AccessibilityElement
		for {
			if err := step.Err(); err != nil {
				stop()
				return fmt.Errorf("progress stage %d: %w", stage, err)
			}
			snapshot, err := cogent.InspectAccessibility(step, "gallery")
			if err == nil && snapshot.Reachable && snapshot.NativeWindow {
				found = map[string]cogent.AccessibilityElement{}
				for _, e := range snapshot.Elements {
					if e.DomainID != "progress-ring" && e.DomainID != "progress-gauge" {
						continue
					}
					value, ok := e.Value.(float64)
					expectedName := "Circular operation progress"
					if e.DomainID == "progress-gauge" {
						expectedName = "Semicircular operation progress"
					}
					if ok && value == float64(percent/100) && e.Role == "AXProgressIndicator" && e.Name == expectedName && e.Enabled && e.ObjectIdentity != 0 && e.Frame[2] > 0 && e.Frame[3] > 0 {
						found[e.DomainID] = e
					}
				}
				if len(found) == 2 {
					break
				}
			}
			select {
			case <-step.Done():
			case <-time.After(10 * time.Millisecond):
			}
		}
		for id, e := range found {
			if previous := identities[id]; previous != 0 && previous != e.ObjectIdentity {
				stop()
				return fmt.Errorf("progress identity changed: %s", id)
			}
			identities[id] = e.ObjectIdentity
			admitted, err := cogent.PerformAccessibilityAction(step, "gallery", e.ID, "set", "0.5")
			if err != nil || admitted {
				stop()
				return errors.Join(err, fmt.Errorf("read-only progress setter admitted: %s", id))
			}
		}
		if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.progress.sample", "stage": stage, "fraction": percent / 100, "elements": found, "readonly_setters_refused": true}); err != nil {
			stop()
			return err
		}
		stop()
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.progress.completed", "stages": 4, "retained_shapes": 2, "physical_input_qualified": false, "human_voiceover_qualified": false})
}

// Only the exact successful-stop sentinel is normalized. Joined or wrapped
// errors retain their identity and any additional driver failure.
func progressRunError(err error) error {
	if err == context.Canceled {
		return nil
	}
	return err
}
