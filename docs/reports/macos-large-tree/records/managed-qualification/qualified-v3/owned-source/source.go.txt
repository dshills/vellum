package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/dshills/vellum/virtual"
)

const maxNodes = 100_000
const smallNodes = 10_000
const buildQuantum = 256

var errData = errors.New("loadedtree: node count must be 1..100000")

func nodeKey(i int) string { return fmt.Sprintf("node-%06d", i) }

// Stored nodes, not a procedural native row source. The first three quarters
// form one huge branch; the final quarter is a separate deep chain. Construction
// is iterative and cancellation checked. TreeData then takes a detached copy.
func generate(ctx context.Context, count int) ([]virtual.TreeNode, error) {
	if ctx == nil {
		return nil, errData
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if count < 1 || count > maxNodes {
		return nil, errData
	}
	out := make([]virtual.TreeNode, count)
	split := max(1, count*3/4)
	for i := range out {
		if i%1024 == 0 {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
		}
		parent := ""
		kind := "Branch leaf"
		if i > 0 && i < split {
			parent = nodeKey(0)
		}
		if i > split {
			parent = nodeKey(i - 1)
			kind = "Chain node"
		}
		if i == 0 {
			kind = "Huge branch root"
		}
		if i == split {
			kind = "Deep chain root"
		}
		out[i] = virtual.TreeNode{ID: nodeKey(i), Parent: parent, Label: fmt.Sprintf("%s %06d", kind, i)}
	}
	return out, ctx.Err()
}
func treeLimits() virtual.TreeLimits {
	return virtual.TreeLimits{MaxNodes: maxNodes, MaxDepth: maxNodes, MaxVisibleNodes: maxNodes, MaxStorageBytes: 128 << 20, MaxWorkOperations: 10_000_000, MaxQueryBytes: 4096, MaxNodeTextBytes: 1024}
}
func makeData(ctx context.Context, id string, count int) (*virtual.TreeData, error) {
	nodes, e := generate(ctx, count)
	if e != nil {
		return nil, e
	}
	return virtual.NewTreeData(ctx, id, 1, nodes, treeLimits())
}
func buildIndex(ctx context.Context, request virtual.OutlineRequest, progress func(float64, string)) (*virtual.TreeIndex, error) {
	b, e := virtual.NewOutlineBuilder(ctx, request)
	if e != nil {
		return nil, e
	}
	progress(0, "Bounded-step immutable tree indexing")
	for {
		done, e := b.Step(buildQuantum)
		if e != nil {
			return nil, e
		}
		if done {
			return b.Result()
		}
	}
}
