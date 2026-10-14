package cogent

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestAccessibilityGeometryCacheSemanticPublication(t *testing.T) {
	_, b := newAXTest(t, 8)
	ordinary := accessibilityNode{ID: "button", Role: "button", Name: "Action"}
	b.updateOwned([]accessibilityNode{ordinary}, nil)
	version := b.geometryVersion
	if version == nil {
		t.Fatal("semantic publication did not establish geometry membership")
	}
	b.updateOwned([]accessibilityNode{ordinary}, nil)
	ordinary.Name = "Changed"
	b.updateOwned([]accessibilityNode{ordinary}, nil)
	if b.geometryVersion != version {
		t.Fatal("unchanged ordinary membership retired geometry")
	}
	invalid := ordinary
	invalid.Value = math.NaN()
	b.updateOwned([]accessibilityNode{invalid}, nil)
	if b.geometryVersion != version {
		t.Fatal("refused encoding mutated geometry membership")
	}
	b.updateOwned([]accessibilityNode{}, nil)
	b.updateOwned([]accessibilityNode{ordinary}, nil)
	if b.geometryVersion == version {
		t.Fatal("semantic remove/readd did not invalidate geometry")
	}
	sends := 0
	b.publishGeometryTo(map[string][4]float32{"button": {1, 2, 3, 4}}, func(uint64, string) { sends++ })
	if sends != 1 || b.geometryCache.payload == "" {
		t.Fatal("known member was not cached")
	}
	b.close()
	if b.geometryVersion != nil || b.geometryMembers != nil || b.geometryCache.payload != "" {
		t.Fatal("retired bridge retained geometry cache")
	}
}

func TestAccessibilityGeometryCacheRetirementAndSharedHosts(t *testing.T) {
	b := &accessibilityBridge{token: 17}
	set := func(nodes ...accessibilityNode) {
		b.mu.Lock()
		b.updateGeometryMembership(nodes)
		b.mu.Unlock()
	}
	ordinary := accessibilityNode{ID: "button", Role: "button"}
	set(ordinary)
	native := map[string][4]float32{}
	sends := 0
	send := func(token uint64, payload string) {
		sends++
		if token != b.token {
			t.Fatal("wrong native bridge")
		}
		var bounds map[string][4]float32
		if err := json.Unmarshal([]byte(payload), &bounds); err != nil {
			t.Fatal(err)
		}
		for id, rect := range bounds {
			native[id] = rect
		}
	}
	first := map[string][4]float32{"button": {1, 2, 3, 4}}
	second := map[string][4]float32{"button": {5, 6, 7, 8}}
	b.publishGeometryTo(first, send)
	b.publishGeometryTo(first, send)
	if sends != 1 {
		t.Fatal("unchanged ordinary frame was republished", sends)
	}
	// A second host can write the same ID. Returning to the first host must
	// restore its bounds rather than consulting an independent host-local cache.
	b.publishGeometryTo(second, send)
	b.publishGeometryTo(first, send)
	if sends != 3 || !reflect.DeepEqual(native, first) {
		t.Fatal("shared native map kept another host's frame", sends, native)
	}
	version := b.geometryVersion
	ordinary.Role, ordinary.Parent, ordinary.Name = "label", "other", "Changed"
	set(ordinary, accessibilityNode{ID: "cell-new", GridScope: "grid", GridIncarnation: 2})
	b.publishGeometryTo(first, send)
	if sends != 3 || b.geometryVersion != version {
		t.Fatal("ordinary ID reuse or scoped grid data invalidated cached geometry")
	}
	// Cocoa removes the old geometry on semantic retirement. No render between
	// removal and restoration may allow the identical previous payload to skip.
	set()
	delete(native, "button")
	set(ordinary)
	b.publishGeometryTo(first, send)
	if sends != 4 || !reflect.DeepEqual(native, first) {
		t.Fatal("same-ID restoration lost its frame", sends, native)
	}
	set(accessibilityNode{ID: "button", GridScope: "grid", GridIncarnation: 3})
	delete(native, "button")
	set(ordinary)
	b.publishGeometryTo(first, send)
	if sends != 5 || !reflect.DeepEqual(native, first) {
		t.Fatal("ordinary/scoped transition reused retired geometry", sends, native)
	}
	// A hidden/presemantic host is outside the current membership. Cocoa can
	// prune its cached geometry on any semantic publication, so always resend.
	unknown := map[string][4]float32{"hidden": {9, 10, 11, 12}}
	b.publishGeometryTo(unknown, send)
	delete(native, "hidden")
	b.publishGeometryTo(unknown, send)
	if sends != 7 || native["hidden"] != unknown["hidden"] {
		t.Fatal("unknown identifier was cached across semantic pruning")
	}
	b.closed.Store(true)
	b.publishGeometryTo(second, send)
	if sends != 7 {
		t.Fatal("closed bridge published geometry")
	}
}

func TestAccessibilityGeometryCacheBoundAndToken(t *testing.T) {
	var c accessibilityGeometryCache
	v := &accessibilityGeometryVersion{}
	sends := 0
	send := func(uint64, string) { sends++ }
	boundary := strings.Repeat("x", accessibilityGeometryCacheLimit)
	c.publish(1, v, boundary, send)
	c.publish(1, v, boundary, send)
	if sends != 1 || len(c.payload) != accessibilityGeometryCacheLimit {
		t.Fatal("exact cache boundary was refused")
	}
	c.publish(2, v, boundary, send)
	if sends != 2 {
		t.Fatal("different bridge token reused old publication")
	}
	oversized := boundary + "x"
	c.publish(2, v, oversized, send)
	c.publish(2, v, oversized, send)
	if sends != 4 || c.payload != "" || c.version != nil {
		t.Fatal("oversized geometry was refused or retained")
	}
	c.publish(2, nil, "{}", send)
	c.publish(2, nil, "{}", send)
	if sends != 6 || c.payload != "" {
		t.Fatal("unversioned geometry was cached")
	}
}
