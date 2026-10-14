package cogent

import "encoding/json"

// A nonzero-sized identity avoids counter wrap and zero-size pointer aliasing.
// The bridge replaces it whenever its ordinary identifier membership changes.
type accessibilityGeometryVersion struct{ identity byte }

// Called only while b.mu is held, after the complete semantic payload succeeds.
// Scoped grid/tree frames retain their separate atomic publication channel.
func (b *accessibilityBridge) updateGeometryMembership(nodes []accessibilityNode) {
	count, changed := 0, b.geometryVersion == nil
	for _, node := range nodes {
		if node.GridScope != "" {
			continue
		}
		count++
		if _, ok := b.geometryMembers[node.ID]; !ok {
			changed = true
		}
	}
	if !changed && count == len(b.geometryMembers) {
		return
	}
	members := make(map[string]struct{}, count)
	for _, node := range nodes {
		if node.GridScope == "" {
			members[node.ID] = struct{}{}
		}
	}
	b.geometryMembers = members
	b.geometryVersion = &accessibilityGeometryVersion{}
}

// Geometry and semantics enqueue under the same bridge lock. The single last
// payload belongs to the native bridge, so one host cannot reuse a stale value
// after a different main/modal host has overwritten the native geometry map.
func (b *accessibilityBridge) publishGeometry(bounds map[string][4]float32) {
	b.publishGeometryTo(bounds, nativeAccessibilityGeometry)
}

func (b *accessibilityBridge) publishGeometryTo(bounds map[string][4]float32, send func(uint64, string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed.Load() {
		return
	}
	version := b.geometryVersion
	for id := range bounds {
		if _, ok := b.geometryMembers[id]; !ok {
			version = nil
			break
		}
	}
	payload, err := json.Marshal(bounds)
	if err != nil {
		return
	}
	b.geometryCache.publish(b.token, version, string(payload), send)
}

const accessibilityGeometryCacheLimit = 64 << 10

// The bridge owns one bounded immutable last payload. Different render hosts
// share it because Cocoa retains their geometry in the same native map.
type accessibilityGeometryCache struct {
	token   uint64
	version *accessibilityGeometryVersion
	payload string
}

func (c *accessibilityGeometryCache) publish(token uint64, version *accessibilityGeometryVersion, payload string, send func(uint64, string)) {
	if version != nil && len(payload) <= accessibilityGeometryCacheLimit && c.token == token && c.version == version && c.payload == payload {
		return
	}
	send(token, payload)
	if version == nil || len(payload) > accessibilityGeometryCacheLimit {
		*c = accessibilityGeometryCache{}
		return
	}
	c.token, c.version, c.payload = token, version, payload
}
