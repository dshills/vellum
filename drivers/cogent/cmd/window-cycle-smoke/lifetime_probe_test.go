//go:build vellum_native_checks

package main

import (
	"errors"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func lifetimeWindow(name string) app.Window {
	return app.Window{ID: name, Title: name, Width: 320, Height: 240, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: name} }}
}

func openLifetimeWindow(t *testing.T, a *app.Application, name string) app.WindowRef {
	t.Helper()
	ref, err := a.OpenWindow(lifetimeWindow(name))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestSameNameConfigurationAndIdentity(t *testing.T) {
	if cycleWindowID(0, false) != "native-cycle-0000" || cycleWindowID(999, false) != "native-cycle-0999" || cycleWindowID(0, true) != cycleWindowID(999, true) {
		t.Fatal("unique/same-name workload identities changed")
	}
	if validateSameNameMode(true, true, 2, 0) != nil || validateSameNameMode(false, false, 1, time.Minute) != nil {
		t.Fatal("valid mode refused")
	}
	for _, mode := range []struct {
		composed bool
		cycles   int
		soak     time.Duration
	}{{false, 2, 0}, {true, 1, 0}, {true, 2, time.Minute}} {
		if validateSameNameMode(true, mode.composed, mode.cycles, mode.soak) == nil {
			t.Fatal("invalid same-name mode accepted", mode)
		}
	}
}

func TestSameNameProbePreservesReplacementAndDrainsOnlyFreshWork(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	openLifetimeWindow(t, a, "survivor")
	old, err := a.OpenWindow(lifetimeWindow("reused"))
	if err != nil {
		t.Fatal(err)
	}
	cleaned := 0
	ticket, err := a.ReserveWindowCompletion(old, func(e error) {
		if !errors.Is(e, app.ErrWindowClosed) {
			t.Error(e)
		}
		cleaned++
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	fresh := lifetimeWindow("reused")
	policies := 0
	fresh.CloseRequest = func() bool { policies++; return true }
	current, err := a.OpenWindow(fresh)
	if err != nil {
		t.Fatal(err)
	}
	var probe *reopenedLifetimeProbe
	var probeErr error
	if err = a.Do(func() { probe, probeErr = probeReopenedLifetime(a, old, current, ticket) }); err != nil {
		t.Fatal(err)
	}
	if probeErr != nil || probe == nil {
		t.Fatal("probe did not complete", probeErr)
	}
	if !a.TakeDirty(current) {
		t.Fatal("probe did not rearm native repaint")
	}
	delivered := 0
	if err = a.PostWindow(current, func() { delivered++ }); err != nil {
		t.Fatal(err)
	}
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err = probe.afterOwnerBarrier(a, current); err != nil {
		t.Fatal(err)
	}
	if delivered != 1 || cleaned != 1 || policies != 0 || len(a.Windows()) != 2 || !a.Pending().Idle() {
		t.Fatal("old probes changed replacement or leaked work", delivered, cleaned, policies, a.Pending())
	}
}

func TestSameNameProbeRefusesMissingOrNonreopenedEvidence(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	openLifetimeWindow(t, a, "survivor")
	old := openLifetimeWindow(t, a, "same")
	ticket, err := a.ReserveWindowCompletion(old, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	current := openLifetimeWindow(t, a, "same")
	other := openLifetimeWindow(t, a, "other")
	for _, pair := range [][2]app.WindowRef{{{}, current}, {old, {}}, {old, old}, {current, current}, {old, other}, {current, old}} {
		if probe, err := probeReopenedLifetime(a, pair[0], pair[1], ticket); err == nil || probe != nil {
			t.Fatal("nonreopened evidence accepted", pair, probe, err)
		}
	}
	if probe, err := probeReopenedLifetime(a, old, current, nil); err == nil || probe != nil {
		t.Fatal("missing original completion accepted", probe, err)
	}
}

func TestSameNameBarrierRefusesLateCallbackOrReplacementRetirement(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	openLifetimeWindow(t, a, "survivor")
	current := openLifetimeWindow(t, a, "same")
	probe := &reopenedLifetimeProbe{}
	probe.unexpected.Store(true)
	if probe.afterOwnerBarrier(a, current) == nil {
		t.Fatal("late old callback accepted")
	}
	probe.unexpected.Store(false)
	if err := a.CloseWindow(current); err != nil {
		t.Fatal(err)
	}
	openLifetimeWindow(t, a, "same")
	if probe.afterOwnerBarrier(a, current) == nil {
		t.Fatal("replaced opening accepted")
	}
	var missing *reopenedLifetimeProbe
	if missing.afterOwnerBarrier(a, current) == nil {
		t.Fatal("missing probe accepted")
	}
}
