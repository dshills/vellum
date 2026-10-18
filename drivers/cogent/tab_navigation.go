package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"sync/atomic"
)

// One latest intent is retained per strip until originating reconciliation
// publishes authoritative state after owner delivery. It is never accepted state.
type tabSelectionIntent struct {
	originID, id   string
	origin, target *tabBinding
	from, to       *tabCloseEpoch
	delivered      atomic.Bool
}

func tabSelectionIntentLive(r *record, intent *tabSelectionIntent) bool {
	return intent != nil && tabCloseEpochLive(intent.from) && tabCloseEpochLive(intent.to) && intent.origin.selectEpoch.Load() == intent.from && intent.target.selectEpoch.Load() == intent.to && tabSelectionEnabled(r, intent.originID, intent.origin) && tabSelectionEnabled(r, intent.id, intent.target) && r.node.Change != nil
}

// Runs only during originating reconciliation, under its native ownership.
func reconcileTabSelection(r *record) {
	tabs := r.control.tabs
	if intent := tabs.pending.Load(); intent != nil && (intent.delivered.Load() || !tabSelectionIntentLive(r, intent)) {
		tabs.pending.CompareAndSwap(intent, nil)
	}
	for id, binding := range tabs.bindings {
		if !binding.button.ContainsFocus() || tabSelectionEnabled(r, id, binding) && tabCloseEpochLive(binding.selectEpoch.Load()) {
			continue
		}
		if tabs.visible && !r.control.disabled {
			if target, err := focusWidget(r); err == nil {
				for targetID, targetBinding := range tabs.bindings {
					if targetBinding.button == target && tabSelectionEnabled(r, targetID, targetBinding) && tabCloseEpochLive(targetBinding.selectEpoch.Load()) {
						target.AsWidget().SetFocus()
						return
					}
				}
			}
		}
		// Skip the whole owned subtree, including private Parts. Child descriptor
		// sync follows this boundary, so their old enabled flags cannot admit focus.
		last := r.widget
		for {
			wb := last.AsWidget()
			if wb.Parts != nil && wb.Parts.HasChildren() {
				last = wb.Parts.Child(wb.Parts.NumChildren() - 1).(core.Widget)
				continue
			}
			if wb.HasChildren() {
				last = wb.Child(wb.NumChildren() - 1).(core.Widget)
				continue
			}
			break
		}
		if manager := binding.button.Events(); manager != nil {
			manager.FocusNextFrom(last)
		}
		return
	}
}

// The binding and its immutable visibility epoch belong to the originating
// strip. Owner delivery reads only owner values and atomic scope lifetimes.
func tabSelectionEnabled(r *record, id string, binding *tabBinding) bool {
	return binding != nil && r != nil && r.control != nil && !r.control.retired && !r.control.disabled && r.control.tabs != nil && r.control.tabs.visible && r.control.tabs.bindings[id] == binding && choiceOptionEnabled(r.node, id)
}
func queueTabSelection(r *record, originID string, origin *tabBinding, id string, target *tabBinding, repeat bool) {
	if origin == nil || target == nil {
		return
	}
	from, to := origin.selectEpoch.Load(), target.selectEpoch.Load()
	if !tabCloseEpochLive(from) || !tabCloseEpochLive(to) {
		return
	}
	// Freeze input intent against the latest live pending proposal. Focus-only
	// Tab traversal followed by activation can return to a still-rendered selection.
	tabs := r.control.tabs
	previous := tabs.pending.Load()
	if previous != nil && !tabSelectionIntentLive(r, previous) {
		tabs.pending.CompareAndSwap(previous, nil)
		previous = nil
	}
	baseline := r.node.SelectedID
	if previous != nil {
		baseline = previous.id
	}
	propose := r.node.Change != nil && (repeat || baseline != id)
	var intent *tabSelectionIntent
	if propose {
		intent = &tabSelectionIntent{originID: originID, id: id, origin: origin, target: target, from: from, to: to}
		tabs.pending.Store(intent)
	}
	admitted := postControl(r, func() {
		if intent != nil {
			defer intent.delivered.Store(true)
		}
		if !tabCloseEpochLive(from) || !tabCloseEpochLive(to) || origin.selectEpoch.Load() != from || target.selectEpoch.Load() != to || !tabSelectionEnabled(r, originID, origin) || !tabSelectionEnabled(r, id, target) || r.node.Change == nil {
			requestControlRestore(r)
			return
		}
		if propose {
			r.node.Change(ui.ChangeEvent{ID: r.node.ID, Kind: ui.Tabs, SelectedID: id})
			r.control.committed = true
		}
		requestControlRestore(r)
	})
	if !admitted && intent != nil {
		tabs.pending.CompareAndSwap(intent, previous)
	}
}
func tabNavigationKey(r *record, id string, binding *tabBinding, e events.Event) {
	if e.IsHandled() || e.Modifiers() != 0 || !tabSelectionEnabled(r, id, binding) || !tabCloseEpochLive(binding.selectEpoch.Load()) {
		return
	}
	step := 0
	edge := false
	target := id
	switch e.KeyCode() {
	case key.CodeLeftArrow:
		step = -1
	case key.CodeRightArrow:
		step = 1
	case key.CodeHome:
		step = 1
		edge = true
	case key.CodeEnd:
		step = -1
		edge = true
	case key.CodeSpacebar, key.CodeReturnEnter:
	default:
		return
	}
	if step != 0 {
		index := -1
		if !edge {
			for i, o := range r.node.Options {
				if o.ID == id {
					index = i
					break
				}
			}
		}
		if index < 0 && step < 0 {
			index = 0
		}
		target = ""
		for range len(r.node.Options) {
			index = (index + step + len(r.node.Options)) % len(r.node.Options)
			o := r.node.Options[index]
			if tabSelectionEnabled(r, o.ID, r.control.tabs.bindings[o.ID]) {
				target = o.ID
				break
			}
		}
	}
	if target == "" {
		return
	}
	next := r.control.tabs.bindings[target]
	if next == nil {
		return
	}
	e.SetHandled()
	next.button.SetFocus() // originating native event boundary only
	queueTabSelection(r, id, binding, target, next, target != id)
}
