//go:build darwin && cgo

package main

import (
	"encoding/json"
	"example.com/vellum/foundationintegration/inputbridge"
)

func clientContracts() map[string]any {
	pool := poolBegin()
	defer poolEnd(pool)
	view := bareViewNew()
	defer bareViewFree(view)
	accept := true
	var updates []inputbridge.Update
	in := inputbridge.Attach(view, func(u inputbridge.Update) bool {
		if !accept {
			return false
		}
		if _, err := inputbridge.Decode(u); err != nil {
			panic(err)
		}
		updates = append(updates, u)
		return true
	}, func() {})
	if in == nil {
		panic("client attachment")
	}
	defer in.Close()
	rect := [4]float64{10, 10, 1, 20}
	if !in.UpdateClient(8, "A👩Z", 1, 1, -1, -1, true, 0, rect) {
		panic("field binding")
	}
	for _, bad := range []struct {
		text   string
		anchor int
	}{{"A\x00B", 0}, {string([]byte{0xff}), 0}, {"A", 1 << 32}} {
		if in.UpdateClient(8, bad.text, bad.anchor, 0, -1, -1, true, 0, rect) {
			panic("invalid client text/offset accepted")
		}
	}
	in.Probe(1)
	if len(updates) != 1 || updates[0].ClientID != 8 || updates[0].BeforeText != "A👩Z" {
		panic("field provenance")
	}
	accept = false
	if in.UpdateClient(40, "B\n\n", 0, 0, -1, -1, true, 1, rect) {
		panic("refused handoff succeeded")
	}
	var refused map[string]any
	if err := json.Unmarshal([]byte(in.Probe(0)), &refused); err != nil {
		panic(err)
	}
	if refused["text"] != "Aに👩Z" || refused["has_marked"] != true || len(updates) != 1 {
		panic("refused handoff mutated old client")
	}
	accept = true
	if in.UpdateClient(40, "B\n\n", 0, 0, -1, -1, true, 1, rect) {
		panic("handoff bypassed acknowledgement")
	}
	if len(updates) != 2 || updates[1].ClientID != 8 || updates[1].Kind != "unmark" || updates[1].BeforeText != "Aに👩Z" {
		panic("handoff lost old-client commit")
	}
	if !in.UpdateClient(40, "B\n\n", 0, 0, -1, -1, true, 2, rect) {
		panic("acknowledged editor binding")
	}
	in.Probe(1)
	if len(updates) != 3 || updates[2].ClientID != 40 || updates[2].BeforeText != "B\n\n" {
		panic("editor provenance")
	}
	if in.UpdateClient(8, "old", 0, 0, -1, -1, true, 2, rect) || len(updates) != 3 {
		panic("stale binding mutated state")
	}
	if in.UpdateClient(40, updates[2].Text, updates[2].Focus, updates[2].Focus, -1, -1, false, 3, rect) {
		panic("focus loss bypassed commit acknowledgement")
	}
	if len(updates) != 4 || updates[3].ClientID != 40 || updates[3].Kind != "unmark" {
		panic("focus loss lost editor commit")
	}
	if !in.UpdateClient(40, updates[3].Text, updates[3].Focus, updates[3].Focus, -1, -1, false, 4, rect) || in.ProbeInactive() != "inactive_optional_super_queries_passed" {
		panic("inactive editor binding")
	}
	return map[string]any{"updates": updates, "refused_handoff_preserves_old_client": true, "handoff_waits_for_old_commit_ack": true, "origin_client_and_before_text_preserved": true, "stale_binding_rejected": true, "focus_loss_waits_for_commit_ack": true, "inactive_super_queries": true, "invalid_client_publication_atomic": true}
}
