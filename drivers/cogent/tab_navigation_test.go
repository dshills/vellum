package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestTabSelectionQueuedCapabilitiesAndLatestOwner(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("tab selection"))
	calls, latest := 0, 0
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}, Change: func(ui.ChangeEvent) { calls++ }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer disposeControlTree(r)
	origin, target := r.control.tabs.bindings["a"], r.control.tabs.bindings["b"]
	queueTabSelection(r, "a", origin, "b", target, false)
	n.Change = func(e ui.ChangeEvent) {
		latest++
		if e.ID != "tabs" || e.Kind != ui.Tabs || e.SelectedID != "b" {
			t.Error("unstable selection payload", e)
		}
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 0 || latest != 1 {
		t.Fatal("latest owner callback", calls, latest)
	}
	// Hiding/showing revokes an already queued action even if IDs recur.
	queueTabSelection(r, "a", origin, "b", target, false)
	old := target.selectEpoch.Load()
	syncTabVisibility(r, false, nil)
	syncTabVisibility(r, true, nil)
	if target.selectEpoch.Load() == old {
		t.Fatal("hide/show reused navigation epoch")
	}
	a.Drain()
	if latest != 1 {
		t.Fatal("old hidden input mutated restored strip")
	}
	// Scope closure is atomic at actual dialog dismissal, before reconciliation.
	scope := &tabCloseScope{}
	syncTabVisibility(r, true, []*tabCloseScope{scope})
	queueTabSelection(r, "a", origin, "b", target, false)
	scope.closed.Store(true)
	a.Drain()
	if latest != 1 {
		t.Fatal("closed modal scope admitted selection")
	}
	syncTabVisibility(r, true, nil)
	queueTabSelection(r, "a", origin, "b", target, false)
	n.Options = n.Options[:1]
	n.Children = n.Children[:1]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if latest != 1 || target.selectEpoch.Load() != nil {
		t.Fatal("removed target retained authority")
	}
	// Direct stale callbacks and one-enabled-option keys cannot propose changes.
	e := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
	tabNavigationKey(r, "b", target, e)
	a.Drain()
	if latest != 1 {
		t.Fatal("retired header acted")
	}
	e = events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
	tabNavigationKey(r, "a", origin, e)
	a.Drain()
	if !e.IsHandled() || latest != 1 {
		t.Fatal("single-option navigation did not no-op")
	}
}

func TestTabKeyboardBurstPreservesAbsoluteReturnProposal(t *testing.T) {
	a, _ := newAXTest(t, 32)
	host := core.NewFrame(core.NewBody("burst"))
	selected := "a"
	var proposals []string
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}, Change: func(e ui.ChangeEvent) { proposals = append(proposals, e.SelectedID); selected = e.SelectedID }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer disposeControlTree(r)
	// Real tab listeners queue two key intents before either owner callback runs.
	right := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
	r.control.tabs.buttons["a"].HandleEvent(right)
	left := events.NewKey(events.KeyChord, 0, key.CodeLeftArrow, 0)
	r.control.tabs.buttons["b"].HandleEvent(left)
	if !right.IsHandled() || !left.IsHandled() || len(proposals) != 0 {
		t.Fatal("key/owner boundary was bypassed", proposals)
	}
	a.Drain()
	if len(proposals) != 2 || proposals[0] != "b" || proposals[1] != "a" || selected != "a" {
		t.Fatal("absolute return proposal dropped", proposals, selected)
	}
}

func TestTabSelectionReadOnlyTransitionsNeverReviveInput(t *testing.T) {
	for _, startReadOnly := range []bool{true, false} {
		t.Run(map[bool]string{true: "readonly-input", false: "writable-input-revoked"}[startReadOnly], func(t *testing.T) {
			a, _ := newAXTest(t, 32)
			host := core.NewFrame(core.NewBody("readonly transition"))
			calls := 0
			change := func(ui.ChangeEvent) { calls++ }
			n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}}
			if !startReadOnly {
				n.Change = change
			}
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			defer disposeControlTree(r)
			e := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
			r.control.tabs.buttons["a"].HandleEvent(e)
			if !e.IsHandled() {
				t.Fatal("actual tab key not consumed")
			}
			if !startReadOnly {
				n.Change = nil
				r, err = reconcile(host, r, n, a, "ax")
				if err != nil {
					t.Fatal(err)
				}
			}
			n.Change = change
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if calls != 0 {
				t.Fatal("queued input revived across readonly transition", calls)
			}
		})
	}
}

func TestTabPendingSelectionDeliveryRefusalAndRetirement(t *testing.T) {
	a, _ := newAXTest(t, 1)
	host := core.NewFrame(core.NewBody("pending selection"))
	var proposals []string
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}, Change: func(e ui.ChangeEvent) { proposals = append(proposals, e.SelectedID) }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer disposeControlTree(r)
	origin, target := r.control.tabs.bindings["a"], r.control.tabs.bindings["b"]
	queueTabSelection(r, "a", origin, "b", target, false)
	first := r.control.tabs.pending.Load()
	if first == nil {
		t.Fatal("missing admitted intent")
	}
	queueTabSelection(r, "b", target, "a", origin, true) // full queue refuses this proposal
	if r.control.tabs.pending.Load() != first {
		t.Fatal("refusal overwrote admitted intent")
	}
	a.Drain()
	if len(proposals) != 1 || proposals[0] != "b" || r.control.tabs.pending.Load() != first {
		t.Fatal("delivery prematurely retired input baseline", proposals)
	}
	// Acceptance/rejection has not yet been published to the native strip.
	queueTabSelection(r, "a", origin, "a", origin, false)
	a.Drain()
	if len(proposals) != 2 || proposals[1] != "a" {
		t.Fatal("activation after delivery but before publication lost", proposals)
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if r.control.tabs.pending.Load() != nil {
		t.Fatal("authoritative reconciliation retained delivered intent")
	}
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	queueTabSelection(r, "a", origin, "b", target, false)
	if r.control.tabs.pending.Load() != nil {
		t.Fatal("refused first input retained intent")
	}
	a.Drain()
	queueTabSelection(r, "a", origin, "b", target, false)
	syncTabVisibility(r, false, nil)
	if r.control.tabs.pending.Load() != nil {
		t.Fatal("hidden strip retained intent")
	}
	a.Drain()
	if len(proposals) != 2 {
		t.Fatal("hidden pending input acted", proposals)
	}
	syncTabVisibility(r, true, nil)
	queueTabSelection(r, "a", origin, "b", target, false)
	disposeControlTree(r)
	if r.control.tabs.pending.Load() != nil {
		t.Fatal("disposal retained intent")
	}
	a.Drain()
	if len(proposals) != 2 {
		t.Fatal("disposed pending input acted", proposals)
	}
}
