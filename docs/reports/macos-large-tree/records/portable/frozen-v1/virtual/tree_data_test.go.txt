package virtual

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func outlineData(t *testing.T, identity string, revision uint64, nodes []TreeNode, limits TreeLimits) *TreeData {
	t.Helper()
	d, err := NewTreeData(context.Background(), identity, revision, nodes, limits)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func outlineIndex(t *testing.T, data *TreeData, options TreeIndexOptions) *TreeIndex {
	t.Helper()
	i, err := BuildTreeIndex(context.Background(), data, options)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func outlineBuilt(t *testing.T, request OutlineRequest, quantum int) *TreeIndex {
	t.Helper()
	b, err := NewOutlineBuilder(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return outlineSteps(t, b, quantum)
}

func outlineSteps(t *testing.T, b *TreeIndexBuilder, quantum int) *TreeIndex {
	t.Helper()
	initial, err := b.Progress()
	if err != nil {
		t.Fatal(err)
	}
	before := initial.Operations
	for {
		done, err := b.Step(quantum)
		if err != nil {
			t.Fatal(err)
		}
		p, err := b.Progress()
		if err != nil || p.Operations < before || p.Operations-before > uint64(quantum) {
			t.Fatalf("progress %+v after %d: %v", p, before, err)
		}
		before = p.Operations
		if done {
			if !p.Done {
				t.Fatal("completed step without completed progress")
			}
			i, err := b.Result()
			if err != nil {
				t.Fatal(err)
			}
			return i
		}
	}
}

func TestTreeDataMalformedAndBudgetRefusal(t *testing.T) {
	tests := []struct {
		name  string
		nodes []TreeNode
		limit TreeLimits
		want  error
	}{
		{"empty-id", []TreeNode{{Label: "bad"}}, TreeLimits{}, ErrTreeSource},
		{"duplicate", []TreeNode{{ID: "a"}, {ID: "a"}}, TreeLimits{}, ErrTreeSource},
		{"missing", []TreeNode{{ID: "a", Parent: "missing"}}, TreeLimits{}, ErrTreeSource},
		{"self-cycle", []TreeNode{{ID: "a", Parent: "a"}}, TreeLimits{}, ErrTreeSource},
		{"disconnected-cycle", []TreeNode{{ID: "root"}, {ID: "a", Parent: "b"}, {ID: "b", Parent: "a"}}, TreeLimits{}, ErrTreeSource},
		{"nul", []TreeNode{{ID: "a", Label: "x\x00y"}}, TreeLimits{}, ErrTreeSource},
		{"invalid-utf8", []TreeNode{{ID: "a", Parent: "\xff"}}, TreeLimits{}, ErrTreeSource},
		{"node-count", []TreeNode{{ID: "a"}, {ID: "b"}}, TreeLimits{MaxNodes: 1}, ErrTreeCapacity},
		{"storage-before-copy", []TreeNode{{ID: "a"}}, TreeLimits{MaxStorageBytes: 100}, ErrTreeCapacity},
		{"payload", []TreeNode{{ID: "a", Label: strings.Repeat("x", 256)}}, TreeLimits{MaxStorageBytes: 200}, ErrTreeCapacity},
		{"node-text", []TreeNode{{ID: "long"}}, TreeLimits{MaxNodeTextBytes: 3}, ErrTreeSource},
		{"work", []TreeNode{{ID: "a"}}, TreeLimits{MaxWorkOperations: 1}, ErrTreeCapacity},
		{"depth", []TreeNode{{ID: "a"}, {ID: "b", Parent: "a"}, {ID: "c", Parent: "b"}}, TreeLimits{MaxDepth: 1}, ErrTreeCapacity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewTreeData(context.Background(), "test", 1, tc.nodes, tc.limit)
			if !errors.Is(err, tc.want) || d != nil {
				t.Fatalf("data=%v err=%v want=%v", d, err, tc.want)
			}
		})
	}
	for _, id := range []string{"", "\xff", "a\x00b", strings.Repeat("i", MaxTreeQueryBytes+1)} {
		if _, err := NewTreeData(context.Background(), id, 1, nil, TreeLimits{}); !errors.Is(err, ErrTreeSource) {
			t.Fatalf("identity %q: %v", id, err)
		}
	}
	if _, err := NewTreeData(context.Background(), "test", 0, nil, TreeLimits{}); !errors.Is(err, ErrTreeSource) {
		t.Fatal(err)
	}
	if _, err := NewTreeData(nil, "test", 1, nil, TreeLimits{}); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	for _, l := range []TreeLimits{{MaxNodes: -1}, {MaxDepth: MaxTreeDepth + 1}, {MaxStorageBytes: MaxTreeStorageBytes + 1}, {MaxWorkOperations: MaxTreeWorkOperations + 1}, {MaxVisibleNodes: -1}, {MaxQueryBytes: -1}, {MaxNodeTextBytes: -1}} {
		if _, err := NewTreeData(context.Background(), "test", 1, nil, l); !errors.Is(err, ErrTreeCapacity) {
			t.Fatalf("limits %+v: %v", l, err)
		}
	}
}

func TestTreeDataOrderAndDetachedPayload(t *testing.T) {
	large := strings.Repeat("padding", 1<<18)
	id := large[12:18]
	parent := large[24:29]
	label := large[31:43]
	identity := large[50:57]
	nodes := []TreeNode{{ID: id, Parent: parent, Label: label}, {ID: "second", Parent: parent}, {ID: parent}, {ID: "other"}}
	d := outlineData(t, identity, 1, nodes, TreeLimits{})
	nodes[0] = TreeNode{ID: "mutated"}
	assertDetached := func(s string) {
		t.Helper()
		if len(s) > 0 {
			p, base := uintptr(unsafe.Pointer(unsafe.StringData(s))), uintptr(unsafe.Pointer(unsafe.StringData(large)))
			if p >= base && p < base+uintptr(len(large)) {
				t.Fatal("retained tiny payload aliases caller's large string backing")
			}
		}
	}
	owned, err := d.Node(id)
	if err != nil || owned.ID != id || owned.Parent != parent || owned.Label != label {
		t.Fatalf("owned %+v: %v", owned, err)
	}
	for _, s := range []string{owned.ID, owned.Parent, owned.Label, d.Stamp().Identity} {
		assertDetached(s)
	}
	i := outlineIndex(t, d, TreeIndexOptions{ExpandAll: true})
	var keys []string
	for _, n := range i.visible {
		keys = append(keys, d.nodes[n].ID)
	}
	if !reflect.DeepEqual(keys, []string{parent, id, "second", "other"}) {
		t.Fatalf("order %v", keys)
	}
	copyData := *d
	if _, err := copyData.Node(id); !errors.Is(err, ErrTreeInvalid) || copyData.Len() != 0 {
		t.Fatalf("value copy admitted: %v", err)
	}
	if _, err := d.Node(id); err != nil {
		t.Fatal("copy refusal changed original", err)
	}
}

type treeCancelAfter struct {
	context.Context
	calls atomic.Int64
	limit int64
}

func (c *treeCancelAfter) Err() error {
	if c.calls.Add(1) >= c.limit {
		return context.Canceled
	}
	return nil
}

func TestTreeConstructionAndBuilderCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if d, err := NewTreeData(ctx, "x", 1, []TreeNode{{ID: "a"}}, TreeLimits{}); d != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("preallocation cancellation: %v %v", d, err)
	}
	nodes := make([]TreeNode, 128)
	for n := range nodes {
		nodes[n].ID = fmt.Sprint(n)
	}
	counted := &treeCancelAfter{Context: context.Background(), limit: 30}
	if d, err := NewTreeData(counted, "x", 1, nodes, TreeLimits{}); d != nil || !errors.Is(err, context.Canceled) || counted.calls.Load() != 30 {
		t.Fatalf("construction mid-work cancellation: %v %v calls=%d", d, err, counted.calls.Load())
	}
	d := outlineData(t, "x", 1, nodes, TreeLimits{})
	ctx, cancel = context.WithCancel(context.Background())
	b, err := NewTreeIndexBuilder(ctx, d, TreeIndexOptions{ExpandAll: true})
	if err != nil {
		t.Fatal(err)
	}
	if done, err := b.Step(7); err != nil || done {
		t.Fatal(done, err)
	}
	before, _ := b.Progress()
	cancel()
	if _, err := b.Step(4096); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	after, _ := b.Progress()
	if after != before {
		t.Fatal("canceled step performed work", before, after)
	}
	if result, err := b.Result(); result != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(result, err)
	}
	if _, err := NewTreeIndexBuilder(ctx, d, TreeIndexOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	readyCtx, readyCancel := context.WithCancel(context.Background())
	ready, err := NewTreeIndexBuilder(readyCtx, d, TreeIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	completed := outlineSteps(t, ready, 256)
	readyCancel()
	if _, err := ready.Result(); !errors.Is(err, context.Canceled) || completed.Len() != len(nodes) {
		t.Fatal("cancellation changed an already returned immutable result", err)
	}
}

type treeHeldContext struct {
	context.Context
	block   atomic.Bool
	entered chan struct{}
	release chan struct{}
}

func (c *treeHeldContext) Err() error {
	if c.block.Swap(false) {
		close(c.entered)
		<-c.release
	}
	return nil
}

func TestTreeBuilderAdmissionAndTerminalBounds(t *testing.T) {
	d := outlineData(t, "x", 1, []TreeNode{{ID: "root"}, {ID: "child", Parent: "root"}}, TreeLimits{})
	c := &treeHeldContext{Context: context.Background(), entered: make(chan struct{}), release: make(chan struct{})}
	b, err := NewTreeIndexBuilder(c, d, TreeIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Result(); !errors.Is(err, ErrTreeNotReady) {
		t.Fatal(err)
	}
	for _, n := range []int{-1, 0, MaxTreeStepOperations + 1} {
		if _, err := b.Step(n); !errors.Is(err, ErrTreeCapacity) {
			t.Fatal(err)
		}
	}
	// Copy the atomic-bearing builder reflectively so vet can still flag real
	// accidental lock copies in production code.
	copyBuilder := reflect.New(reflect.TypeOf(b).Elem())
	copyBuilder.Elem().Set(reflect.ValueOf(b).Elem())
	if _, err := copyBuilder.Interface().(*TreeIndexBuilder).Step(1); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	c.block.Store(true)
	done := make(chan error, 1)
	go func() { _, err := b.Step(1); done <- err }()
	select {
	case <-c.entered:
	case <-time.After(time.Second):
		t.Fatal("worker did not enter counted step")
	}
	for _, call := range []func() error{func() error { _, e := b.Step(1); return e }, func() error { _, e := b.Progress(); return e }, func() error { _, e := b.Result(); return e }} {
		if err := call(); !errors.Is(err, ErrTreeBusy) {
			t.Fatal(err)
		}
	}
	close(c.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	i := outlineSteps(t, b, 7)
	prior := i.Stats()
	if done, err := b.Step(1); !done || err != nil || i.Stats() != prior {
		t.Fatal("terminal step changed index", done, err)
	}
	copyIndex := *i
	if _, err := NewOutline(&copyIndex, OutlineConfig{}, OutlineViewport{Width: 100, Height: 24}); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	limited := outlineData(t, "limited", 1, []TreeNode{{ID: "root"}, {ID: "child", Parent: "root"}}, TreeLimits{MaxVisibleNodes: 1})
	if _, err := BuildTreeIndex(context.Background(), limited, TreeIndexOptions{ExpandAll: true}); !errors.Is(err, ErrTreeCapacity) {
		t.Fatal(err)
	}
	// Data construction fits its work budget, but index construction does not.
	limited = outlineData(t, "work", 1, []TreeNode{{ID: "root"}}, TreeLimits{MaxWorkOperations: 4})
	if _, err := BuildTreeIndex(context.Background(), limited, TreeIndexOptions{}); !errors.Is(err, ErrTreeCapacity) {
		t.Fatal(err)
	}
	for _, q := range []string{"\xff", "x\x00y"} {
		if _, err := NewTreeIndexBuilder(context.Background(), d, TreeIndexOptions{Filter: q}); !errors.Is(err, ErrTreeSource) {
			t.Fatal(err)
		}
	}
	if _, err := NewTreeIndexBuilder(context.Background(), d, TreeIndexOptions{Filter: strings.Repeat("x", MaxTreeQueryBytes+1)}); !errors.Is(err, ErrTreeCapacity) {
		t.Fatal(err)
	}
}

func TestTreeHundredThousandWideAndDeepBoundedSteps(t *testing.T) {
	const count = 100_000
	for _, deep := range []bool{false, true} {
		name := "wide"
		if deep {
			name = "deep"
		}
		t.Run(name, func(t *testing.T) {
			nodes := make([]TreeNode, count)
			for n := range nodes {
				nodes[n].ID = fmt.Sprintf("node:%06d", n)
				if n > 0 {
					parent := 0
					if deep {
						parent = n - 1
					}
					nodes[n].Parent = nodes[parent].ID
				}
			}
			nodes[count-1].Label = "matching last node"
			d := outlineData(t, name, 1, nodes, TreeLimits{})
			if d.Len() != count || d.Stats().ConstructionOperations > 4*count+1 || deep && d.Stats().Depth != count-1 {
				t.Fatalf("data stats %+v", d.Stats())
			}
			var baseline *TreeIndex
			for _, quantum := range []int{1, 7, MaxTreeStepOperations} {
				b, err := NewTreeIndexBuilder(context.Background(), d, TreeIndexOptions{ExpandAll: true})
				if err != nil {
					t.Fatal(err)
				}
				i := outlineSteps(t, b, quantum)
				if i.Len() != count || i.Stats().Operations > 5*count+5 {
					t.Fatalf("index stats %+v", i.Stats())
				}
				if baseline != nil && (!reflect.DeepEqual(i.visible, baseline.visible) || !reflect.DeepEqual(i.positions, baseline.positions) || i.Stats() != baseline.Stats()) {
					t.Fatal("operation budget changed completed index")
				}
				baseline = i
			}
			o, err := NewOutline(baseline, OutlineConfig{MaxRows: 32}, OutlineViewport{Width: 640, Height: 480, ScrollY: (count - 20) * 24})
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := o.Snapshot()
			if err != nil || len(snapshot.Rows) != 20 || snapshot.Rows[0].ID != nodes[count-20].ID || snapshot.Rows[19].ID != nodes[count-1].ID {
				t.Fatalf("far snapshot rows=%d err=%v", len(snapshot.Rows), err)
			}
			if deep && snapshot.Rows[19].Depth != count-1 {
				t.Fatal("deep semantic level was truncated")
			}
			filter := outlineIndex(t, d, TreeIndexOptions{Filter: "matching last"})
			want := 2
			if deep {
				want = count
			}
			if filter.Len() != want || filter.Stats().Operations > 5*count+5 {
				t.Fatalf("filtered ancestors %+v want=%d", filter.Stats(), want)
			}
			collapsed := outlineIndex(t, d, TreeIndexOptions{})
			if collapsed.Len() != 1 {
				t.Fatal("default tree did not collapse its branch")
			}
		})
	}
}
