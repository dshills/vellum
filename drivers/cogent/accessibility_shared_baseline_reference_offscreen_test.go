//go:build offscreen

// Frozen f641776 preparation, independent of lightweight baseline construction.
package cogent

import (
	"encoding/json"
	"math"
)

func (b *accessibilityBridge) legacySharedBaselinePrepare(nodes []accessibilityNode, handlers map[string]accessibilityHandler, owned bool, completedSemantic string) ([]accessibilityNode, map[string]accessibilityHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed.Load() {
		return nil, nil
	}
	// Direct preparation can change the cache independently of lastSnapshot.
	// Only a subsequent successful ready publication can certify reuse again.
	b.completedSnapshot = ""
	// Group input indices, rather than repeatedly copying large node records
	// while each scope grows. Completed cache records are still fully detached.
	scopes := map[string][]int{}
	for index, node := range nodes {
		if node.GridScope != "" {
			scopes[node.GridScope] = append(scopes[node.GridScope], index)
		}
	}
	cache := map[string][]accessibilityNode{}
	replacements := map[string][]accessibilityNode{}
	actions := handlers
	if !owned {
		actions = make(map[string]accessibilityHandler, len(handlers))
		for id, handler := range handlers {
			actions[id] = handler
		}
	}
	for scope, indices := range scopes {
		host := nodes[indices[0]]
		if host.ID != scope {
			continue // Framework traversal always emits the table host first.
		}
		if !host.GridPending {
			group := make([]accessibilityNode, len(indices))
			for i, index := range indices {
				group[i] = nodes[index]
				group[i].activationGuard = nil // Completed caches retain no owner/model closures.
			}
			detachAccessibilityNodes(group)
			cache[scope] = group
			continue
		}
		previous := b.gridPublications[scope]
		retained := nodes[indices[0] : indices[0]+1]
		if len(previous) > 0 && previous[0].GridIncarnation == host.GridIncarnation && previous[0].Parent == host.Parent && previous[0].Role == host.Role {
			cache[scope] = previous
			retained = previous
		}
		retained = cloneAccessibilityNodes(retained)
		for i := range retained {
			n := &retained[i]
			n.GridPending = true
			n.GridTargetGeneration = host.GridTargetGeneration
			n.Disabled, n.Focused = true, false
			n.Actions = []string{}
			n.TextClient, n.TextSelection, n.TextFocus = 0, nil, 0
			delete(actions, n.ID)
		}
		// Descriptor labels/help remain current even while cell data are read-only.
		retained[0].Name, retained[0].Help = host.Name, host.Help
		for _, index := range indices {
			delete(actions, nodes[index].ID)
		}
		replacements[scope] = retained
	}
	if owned && len(replacements) == 0 {
		b.gridPublications = cache
		b.completedSnapshot = completedSemantic
		return nodes, actions
	}
	out := make([]accessibilityNode, 0, len(nodes))
	for _, node := range nodes {
		if replacement, pending := replacements[node.GridScope]; pending {
			if node.ID == node.GridScope {
				out = append(out, replacement...)
			}
			continue
		}
		out = append(out, node)
	}
	b.gridPublications = cache
	b.completedSnapshot = completedSemantic
	return out, actions
}

func legacySharedBaselinePendingPatch(previous, next []accessibilityNode, generation uint64) string {
	if generation == 0 || generation == math.MaxUint64 || len(previous) == 0 || len(previous) != len(next) {
		return ""
	}
	patch := accessibilityPendingPublication{Previous: generation, Generation: generation + 1}
	scopes := make(map[string]accessibilityPendingScope)
	for _, n := range next {
		if n.GridScope != "" && n.ID == n.GridScope && n.GridPending {
			if n.GridIncarnation == 0 || n.GridGeneration == 0 || n.GridTargetGeneration < n.GridGeneration || n.GridFrame == nil {
				return ""
			}
			if _, exists := scopes[n.ID]; exists {
				return ""
			}
			scope := accessibilityPendingScope{n.ID, n.GridIncarnation, n.GridGeneration, n.GridTargetGeneration}
			scopes[n.ID] = scope
			patch.Scopes = append(patch.Scopes, scope)
		}
	}
	if len(scopes) == 0 {
		return ""
	}
	for i, n := range next {
		old := previous[i]
		if scope, pending := scopes[n.GridScope]; pending {
			if !n.GridPending || n.GridIncarnation != scope.Incarnation || n.GridGeneration != scope.Completed || n.GridTargetGeneration != scope.Target {
				return ""
			}
			old.GridPending, old.GridTargetGeneration = true, scope.Target
			old.Disabled, old.Focused, old.Actions = true, false, []string{}
			old.TextClient, old.TextSelection, old.TextFocus = 0, nil, 0
		}
		if !accessibilitySameWireNode(old, n) {
			return ""
		}
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		return ""
	}
	return string(encoded)
}
