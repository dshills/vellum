//go:build darwin && cgo

package main

import (
	"encoding/json"
	"example.com/vellum/foundationintegration/inputbridge"
	"strings"
)

func deltaContracts() map[string]any {
	pool := poolBegin()
	defer poolEnd(pool)
	view := bareViewNew()
	defer bareViewFree(view)
	shadow := strings.Repeat("a", 2<<20)
	pending := []inputbridge.Update{}
	maxBytes := 0
	in := inputbridge.Attach(view, func(u inputbridge.Update) bool {
		b, _ := json.Marshal(u)
		if len(b) > maxBytes {
			maxBytes = len(b)
		}
		pending = append(pending, u)
		return len(b) < 64<<10
	}, func() {})
	if in == nil {
		panic("delta attachment")
	}
	defer in.Close()
	in.EnableDeltas()
	if !in.UpdateClient(40, shadow, 1500000, 1500000, -1, -1, true, 0, [4]float64{10, 10, 0, 23}) {
		panic("large native publication")
	}
	for _, step := range []int{1, 2, 3} {
		if in.Probe(step) == "" {
			panic("delta protocol probe")
		}
	}
	if len(pending) != 3 {
		panic("queued delta count")
	}
	for _, u := range pending {
		if !u.Delta {
			panic("whole snapshot queued")
		}
		next, e := inputbridge.ExpandDelta(u, shadow)
		if e != nil {
			panic(e)
		}
		shadow = next.Text
	}
	if shadow != strings.Repeat("a", 1500000)+"日本"+strings.Repeat("a", (2<<20)-1500000) {
		panic("large native queued edits lost bytes")
	}
	last := pending[2]
	if !in.UpdateClient(40, shadow, last.Anchor, last.Focus, -1, -1, true, last.Sequence, [4]float64{10, 10, 0, 23}) {
		panic("delta acknowledgement")
	}
	before := in.Probe(0)
	in.Probe(10)
	after := in.Probe(0)
	if before != after || len(pending) != 3 {
		panic("oversized delta mutated native state")
	}
	return map[string]any{"document_bytes": 2 << 20, "accepted": len(pending), "max_encoded_message_bytes": maxBytes, "oversized_refused_without_mutation": true, "global_utf16_anchor": last.Focus, "scope": "actual hidden Cocoa client; deterministic protocol, not live IME"}
}
