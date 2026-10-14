//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
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
	if err != nil || performs != 1 || ownerCalls < 5 || record.Admissions != 1 || !record.OwnerAndNativeCompleted {
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
