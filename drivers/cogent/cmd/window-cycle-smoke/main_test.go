//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"image"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
)

func TestProbeDeadlineReachesBlockingCheck(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := waitWithin(parent, 20*time.Millisecond, func(inner context.Context) (bool, error) {
		if _, ok := inner.Deadline(); !ok {
			t.Fatal("probe received unbounded parent context")
		}
		<-inner.Done()
		return false, inner.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) || parent.Err() != nil {
		t.Fatalf("probe deadline/caller lifetime: %v %v", err, parent.Err())
	}
}

func TestOwnerAdmissionCannotSubstituteForAcknowledgement(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	called := false
	_, err = ownerProbe(ctx, a, func() ownerReply { called = true; return ownerReply{} })
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("admitted but undrained callback qualified: %v called=%t", err, called)
	}
	// An expired request may still be queued. It must skip native work and its
	// buffered late acknowledgment must not strand the application owner.
	drained := make(chan struct{})
	go func() { a.Drain(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("late acknowledgment blocked owner")
	}
	if called || !a.Pending().Idle() {
		t.Fatal("expired owner probe executed or remained pending")
	}
}

func TestOwnerAcknowledgementCarriesCompletedFailure(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	marker := errors.New("completed owner check")
	finished := make(chan ownerReply, 1)
	go func() {
		value, e := ownerProbe(ctx, a, func() ownerReply { return ownerReply{err: marker} })
		if e != nil {
			value.err = e
		}
		finished <- value
	}()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case value := <-finished:
			if !errors.Is(value.err, marker) {
				t.Fatalf("lost completed owner result: %v", value.err)
			}
			return
		case <-ticker.C:
			a.Drain()
		case <-ctx.Done():
			t.Fatal("acknowledged probe did not join")
		}
	}
}

func TestSampledMemoryGuardsRetainBoundary(t *testing.T) {
	if memoryGuard(observation{HeapAlloc: 512 << 20, RSSBytes: 1 << 30}) != nil {
		t.Fatal("exact sampled bounds refused")
	}
	for _, s := range []observation{{HeapAlloc: 512<<20 + 1}, {RSSBytes: 1<<30 + 1}} {
		if memoryGuard(s) == nil {
			t.Fatal("sample above safety bound accepted")
		}
	}
}
func TestCycleConfigurationBounds(t *testing.T) {
	if err := (config{cycles: 1000, every: 100, timeout: 12 * time.Minute}).validate(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []config{{cycles: 0, every: 1, timeout: time.Minute}, {cycles: 1001, every: 100, timeout: time.Minute}, {cycles: 10, every: 0, timeout: time.Minute}, {cycles: 10, every: 11, timeout: time.Minute}, {cycles: 1, every: 1, timeout: time.Second}, {cycles: 1, every: 1, timeout: 21 * time.Minute}, {cycles: 1, every: 1, timeout: time.Minute, heapProfile: "relative.profile"}} {
		if c.validate() == nil {
			t.Fatalf("accepted invalid config: %+v", c)
		}
	}
}

func TestHeapProfileRefusesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.profile")
	if err := os.WriteFile(path, []byte("preserve existing bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeHeapProfile(path); !errors.Is(err, os.ErrExist) {
		t.Fatalf("profile overwrote/refused wrongly: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "preserve existing bytes" {
		t.Fatal("existing profile bytes changed")
	}
}

func TestOwnerProbeJoinsExecutingCallback(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	done := make(chan error, 1)
	go func() {
		_, e := ownerProbe(ctx, a, func() ownerReply { close(entered); <-release; return ownerReply{} })
		done <- e
	}()
	deadline := time.After(time.Second)
	for a.Pending().Ordinary == 0 {
		select {
		case <-deadline:
			t.Fatal("owner callback not admitted")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	drained := make(chan struct{})
	go func() { a.Drain(); close(drained) }()
	<-entered
	cancel()
	select {
	case <-done:
		t.Fatal("running callback detached at cancellation")
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("callback did not join")
	}
	<-drained
}
func TestComposedVisibilityRequiresCompleteInsideFrames(t *testing.T) {
	b := [4]float64{10, 20, 100, 80}
	if !frameInside([4]float64{20, 30, 20, 20}, b) {
		t.Fatal("contained frame refused")
	}
	for _, f := range [][4]float64{{20, 30, 0, 20}, {9, 30, 20, 20}, {90, 30, 30, 20}, {20, 90, 20, 20}} {
		if frameInside(f, b) {
			t.Fatal("empty/clipped frame qualified", f)
		}
	}
}

func TestComposedDeadlineHasBoundedExtendedBudget(t *testing.T) {
	if err := (config{cycles: 1000, every: 25, timeout: 45 * time.Minute, composed: true}).validate(); err != nil {
		t.Fatal(err)
	}
	if (config{cycles: 1000, every: 25, timeout: 45*time.Minute + time.Second, composed: true}).validate() == nil {
		t.Fatal("unbounded composed run admitted")
	}
}

func TestTreeEnvelopeRequiresMatchingCompletedNativeView(t *testing.T) {
	g := cogent.TreeInspection{ID: "outline", Incarnation: 17, Generation: 3, VisibleRows: 1000, Rows: []cogent.TreeRowGeometry{{ID: "row-id", Label: "Row", Index: 7, Selected: true, Bounds: image.Rect(0, 0, 100, 20)}}}
	host := cogent.AccessibilityElement{ID: "host", DomainID: "outline", ObjectIdentity: 1, Role: "AXOutline", GridScope: "host", GridIncarnation: 17, GridGeneration: 3, GridTargetGeneration: 3, RowCount: 1000, Frame: [4]float64{0, 0, 100, 200}}
	row := host
	row.ID = "row"
	row.DomainID = "row-id"
	row.ObjectIdentity = 2
	row.Role = "AXRow"
	row.Parent = "host"
	row.Name = "Row"
	row.Selected = true
	row.RowIndex = [2]int{7, 1}
	native := cogent.AccessibilityInspection{Reachable: true, NativeWindow: true, Elements: []cogent.AccessibilityElement{host, row}}
	if ready, count := treeEnvelopeCompleted(g, native); !ready || count != 1 {
		t.Fatal("valid completed view refused", count)
	}
	// Production treeAccessibility omits empty-bounds rows. The checker accepts
	// that bounded view and refuses a native row that should have been omitted.
	clipped := g
	clipped.Rows = append(append([]cogent.TreeRowGeometry(nil), g.Rows...), cogent.TreeRowGeometry{ID: "clipped-row", Label: "Clipped", Index: 8})
	if ready, count := treeEnvelopeCompleted(clipped, native); !ready || count != 1 {
		t.Fatal("production view with omitted clipped row refused", count)
	}
	extra := native
	extra.Elements = append([]cogent.AccessibilityElement(nil), native.Elements...)
	clippedNative := row
	clippedNative.ID, clippedNative.DomainID = "clipped-native", "clipped-row"
	clippedNative.ObjectIdentity = 4
	clippedNative.Frame = [4]float64{}
	extra.Elements = append(extra.Elements, clippedNative)
	if ready, _ := treeEnvelopeCompleted(clipped, extra); ready {
		t.Fatal("out-of-contract native clipped row qualified")
	}
	cases := map[string]func(*cogent.AccessibilityInspection){
		"pending":      func(s *cogent.AccessibilityInspection) { s.Elements[1].GridPending = true },
		"generation":   func(s *cogent.AccessibilityInspection) { s.Elements[1].GridGeneration = 2 },
		"target":       func(s *cogent.AccessibilityInspection) { s.Elements[1].GridTargetGeneration = 4 },
		"incarnation":  func(s *cogent.AccessibilityInspection) { s.Elements[1].GridIncarnation = 18 },
		"scope":        func(s *cogent.AccessibilityInspection) { s.Elements[1].GridScope = "foreign" },
		"selection":    func(s *cogent.AccessibilityInspection) { s.Elements[1].Selected = false },
		"empty-frame":  func(s *cogent.AccessibilityInspection) { s.Elements[1].Frame = [4]float64{} },
		"row-count":    func(s *cogent.AccessibilityInspection) { s.Elements[1].RowCount = 999 },
		"index-length": func(s *cogent.AccessibilityInspection) { s.Elements[1].RowIndex[1] = 2 },
		"duplicate": func(s *cogent.AccessibilityInspection) {
			duplicate := s.Elements[1]
			duplicate.ID = "duplicate-row"
			duplicate.ObjectIdentity = 3
			s.Elements = append(s.Elements, duplicate)
		},
		"missing": func(s *cogent.AccessibilityInspection) { s.Elements = s.Elements[:1] },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := native
			candidate.Elements = append([]cogent.AccessibilityElement(nil), native.Elements...)
			change(&candidate)
			if ready, _ := treeEnvelopeCompleted(g, candidate); ready {
				t.Fatal("stale/missing native rows qualified")
			}
		})
	}
}
