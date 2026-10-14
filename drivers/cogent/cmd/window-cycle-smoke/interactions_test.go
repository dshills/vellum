//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/color"
	"sync/atomic"
	"testing"
	"time"
)

func actionFixture() cogent.AccessibilityInspection {
	return cogent.AccessibilityInspection{Reachable: true, NativeWindow: true, Elements: []cogent.AccessibilityElement{{ID: "native", DomainID: "target", Enabled: true, Frame: [4]float64{0, 0, 10, 10}}}}
}
func TestNativeActionJoinsOwnerBeforeRetryAndWaitsForPublication(t *testing.T) {
	queued, accepted := false, false
	ownerCalls, performs, publications := 0, 0, 0
	record, err := runNativeAction(context.Background(), time.Second, "target", "press", "", nativeActionHooks{
		accepted: func(context.Context) (bool, error) {
			ownerCalls++
			if queued {
				queued = false
				accepted = true
			}
			return accepted, nil
		},
		inspect:   func(context.Context) (cogent.AccessibilityInspection, error) { return actionFixture(), nil },
		perform:   func(context.Context, string) (bool, error) { performs++; queued = true; return true, nil },
		completed: func(cogent.AccessibilityInspection) bool { publications++; return publications == 3 },
	})
	if err != nil || performs != 1 || ownerCalls < 5 || record.Admissions != 1 || record.ElapsedSeconds <= 0 || !record.OwnerAndNativeCompleted {
		t.Fatal("admission substituted for owner/publication or press repeated", record, err, ownerCalls, performs)
	}
}
func TestNativeActionRequiresItsOwnAdmission(t *testing.T) {
	performs := 0
	record, err := runNativeAction(context.Background(), time.Second, "target", "set", "value", nativeActionHooks{
		accepted:  func(context.Context) (bool, error) { return true, nil },
		inspect:   func(context.Context) (cogent.AccessibilityInspection, error) { return actionFixture(), nil },
		perform:   func(context.Context, string) (bool, error) { performs++; return true, nil },
		completed: func(cogent.AccessibilityInspection) bool { return true },
	})
	if err == nil || performs != 0 || record.OwnerAndNativeCompleted {
		t.Fatal("initial domain value counted as native acceptance", record, err)
	}
}

func TestNativeActionRetriesDrainedUnacceptedAdmission(t *testing.T) {
	queued, accepted := false, false
	performs, acknowledgements := 0, 0
	record, err := runNativeAction(context.Background(), time.Second, "target", "press", "", nativeActionHooks{
		accepted: func(context.Context) (bool, error) {
			if queued {
				queued = false
				acknowledgements++
				// The first admitted closure drained without accepting the model.
				accepted = acknowledgements == 2
			}
			return accepted, nil
		},
		inspect: func(context.Context) (cogent.AccessibilityInspection, error) { return actionFixture(), nil },
		perform: func(context.Context, string) (bool, error) {
			if queued {
				t.Fatal("retry preceded owner acknowledgement")
			}
			performs++
			queued = true
			return true, nil
		},
		completed: func(cogent.AccessibilityInspection) bool { return true },
	})
	if err != nil || performs != 2 || acknowledgements != 2 || record.Attempts != 2 || record.Admissions != 2 || !record.OwnerAndNativeCompleted {
		t.Fatal("drained unaccepted request prevented retry or bypassed acceptance", record, err, performs, acknowledgements)
	}
}

func TestNativeActionInspectionFailureAfterAdmissionStops(t *testing.T) {
	retirement := cogent.AccessibilityRetirement{Window: "test", Token: 8, Generation: 3, Source: "native-attachment-close"}
	closed := &cogent.AccessibilityClosedError{Retirement: retirement}
	performs, inspections := 0, 0
	record, err := runNativeAction(context.Background(), time.Second, "target", "press", "", nativeActionHooks{
		accepted: func(context.Context) (bool, error) { return performs > 0, nil },
		inspect: func(context.Context) (cogent.AccessibilityInspection, error) {
			inspections++
			if inspections > 1 {
				return cogent.AccessibilityInspection{}, closed
			}
			return actionFixture(), nil
		},
		perform:   func(context.Context, string) (bool, error) { performs++; return true, nil },
		completed: func(cogent.AccessibilityInspection) bool { t.Fatal("failed inspection qualified"); return true },
	})
	if !errors.Is(err, cogent.ErrAccessibilityClosed) || performs != 1 || record.Attempts != 1 || record.Admissions != 1 || record.OwnerAndNativeCompleted || record.AttachmentRetirement == nil || *record.AttachmentRetirement != retirement {
		t.Fatal("closed attachment swallowed, repeated or qualified", record, err, performs)
	}
}
func TestNativeAdmissionCannotQualifyWithoutOwnerAcceptance(t *testing.T) {
	ownerCalls, performs := 0, 0
	record, err := runNativeAction(context.Background(), 25*time.Millisecond, "target", "set", "value", nativeActionHooks{
		accepted:  func(context.Context) (bool, error) { ownerCalls++; return false, nil },
		inspect:   func(context.Context) (cogent.AccessibilityInspection, error) { return actionFixture(), nil },
		perform:   func(context.Context, string) (bool, error) { performs++; return true, nil },
		completed: func(cogent.AccessibilityInspection) bool { return true },
	})
	if !errors.Is(err, context.DeadlineExceeded) || record.OwnerAndNativeCompleted || performs < 1 || ownerCalls < 2*performs {
		t.Fatal("unaccepted native admission qualified or lacked barrier", record, err, ownerCalls, performs)
	}
}

func TestNativeActionAcceptedRequestWaitsForAsyncCompletion(t *testing.T) {
	var requested, completed atomic.Bool
	var performs atomic.Int32
	entered := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := runNativeAction(context.Background(), time.Second, "target", "collapse", "", nativeActionHooks{
			accepted: func(context.Context) (bool, error) { return requested.Load(), nil },
			inspect:  func(context.Context) (cogent.AccessibilityInspection, error) { return actionFixture(), nil },
			perform: func(context.Context, string) (bool, error) {
				performs.Add(1)
				requested.Store(true)
				close(entered)
				return true, nil
			},
			completed: func(cogent.AccessibilityInspection) bool { return completed.Load() },
		})
		result <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request not admitted")
	}
	time.Sleep(35 * time.Millisecond)
	select {
	case err := <-result:
		t.Fatal("incomplete worker qualified", err)
	default:
	}
	if performs.Load() != 1 {
		t.Fatal("accepted pending request repeated")
	}
	completed.Store(true)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("completion not joined")
	}
}
func TestNativeActionSelectsActionableDisclosureHeading(t *testing.T) {
	native := actionFixture()
	group := native.Elements[0]
	group.ID = "group"
	group.Role = "AXGroup"
	heading := group
	heading.ID = "heading"
	heading.Role = "AXDisclosureTriangle"
	native.Elements = []cogent.AccessibilityElement{group, heading}
	accepted := false
	record, err := runNativeAction(context.Background(), time.Second, "target", "collapse", "", nativeActionHooks{
		targetRole: "AXDisclosureTriangle",
		accepted:   func(context.Context) (bool, error) { return accepted, nil },
		inspect:    func(context.Context) (cogent.AccessibilityInspection, error) { return native, nil },
		perform: func(_ context.Context, id string) (bool, error) {
			if id != "heading" {
				t.Fatal("noninteractive containing group targeted", id)
			}
			accepted = true
			return true, nil
		},
		completed: func(cogent.AccessibilityInspection) bool { return true },
	})
	if err != nil || record.NativeRole != "AXDisclosureTriangle" || record.Attempts != 1 {
		t.Fatal(record, err)
	}
}

func TestColorEnvelopeRequiresEveryNativeFieldValueAndPreview(t *testing.T) {
	want := color.NRGBA{R: 53, G: 103, B: 181, A: 128}
	native := cogent.AccessibilityInspection{Reachable: true, NativeWindow: true}
	values := []any{float64(53), float64(103), float64(181), float64(128), "#3567B580"}
	for i, c := range []ui.ColorChannel{ui.ColorRed, ui.ColorGreen, ui.ColorBlue, ui.ColorAlpha, ui.ColorHex} {
		role := "AXSlider"
		if c == ui.ColorHex {
			role = "AXTextField"
		}
		native.Elements = append(native.Elements, cogent.AccessibilityElement{DomainID: ui.ColorFieldID("color", c), ObjectIdentity: uint64(i + 1), Role: role, Enabled: true, Value: values[i], Frame: [4]float64{0, 0, 100, 20}})
	}
	native.Elements = append(native.Elements, cogent.AccessibilityElement{DomainID: "color/color/preview", ObjectIdentity: 6, Role: "AXImage", Name: "Workspace color preview #3567B580", Frame: [4]float64{0, 0, 32, 32}})
	if !colorEnvelopeCompleted(native, want) {
		t.Fatal("valid color view refused")
	}
	for name, change := range map[string]func(*cogent.AccessibilityInspection){
		"unreachable":     func(s *cogent.AccessibilityInspection) { s.Reachable = false },
		"non-native":      func(s *cogent.AccessibilityInspection) { s.NativeWindow = false },
		"missing-channel": func(s *cogent.AccessibilityInspection) { s.Elements = append(s.Elements[:1], s.Elements[2:]...) },
		"zero-identity":   func(s *cogent.AccessibilityInspection) { s.Elements[0].ObjectIdentity = 0 },
		"wrong-role":      func(s *cogent.AccessibilityInspection) { s.Elements[1].Role = "AXGroup" },
		"disabled":        func(s *cogent.AccessibilityInspection) { s.Elements[2].Enabled = false },
		"empty-frame":     func(s *cogent.AccessibilityInspection) { s.Elements[3].Frame = [4]float64{} },
		"stale-hex":       func(s *cogent.AccessibilityInspection) { s.Elements[4].Value = "#3567B5FF" },
		"stale-channel":   func(s *cogent.AccessibilityInspection) { s.Elements[0].Value = float64(99) },
		"stale-preview":   func(s *cogent.AccessibilityInspection) { s.Elements[5].Name = "Workspace color preview #3567B5FF" },
		"missing-preview": func(s *cogent.AccessibilityInspection) { s.Elements = s.Elements[:5] },
	} {
		t.Run(name, func(t *testing.T) {
			s := native
			s.Elements = append([]cogent.AccessibilityElement(nil), native.Elements...)
			change(&s)
			if colorEnvelopeCompleted(s, want) {
				t.Fatal("incomplete color view qualified")
			}
		})
	}
}
