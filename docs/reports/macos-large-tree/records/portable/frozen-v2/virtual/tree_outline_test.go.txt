package virtual

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func testOutline(t *testing.T, nodes []TreeNode, expanded bool) *Outline {
	t.Helper()
	d := outlineData(t, "outline", 1, nodes, TreeLimits{})
	o, err := NewOutline(outlineIndex(t, d, TreeIndexOptions{ExpandAll: expanded}), OutlineConfig{RowHeight: 10, MaxRows: 8}, OutlineViewport{Width: 80, Height: 25})
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestOutlineVisibleOrderAgainstIndependentTraversal(t *testing.T) {
	rng := rand.New(rand.NewSource(781))
	nodes := make([]TreeNode, 128)
	children := map[string][]TreeNode{}
	for n := range nodes {
		nodes[n].ID = fmt.Sprintf("key:%03d", n)
		nodes[n].Label = fmt.Sprintf("label:%d", n%7)
		if n > 0 && rng.Intn(5) != 0 {
			nodes[n].Parent = nodes[rng.Intn(n)].ID
		}
		children[nodes[n].Parent] = append(children[nodes[n].Parent], nodes[n])
	}
	o := testOutline(t, nodes, false)
	expanded := map[string]bool{}
	filter := ""
	for turn := 0; turn < 120; turn++ {
		node := nodes[rng.Intn(len(nodes))]
		change := OutlineChange{Kind: OutlineToggle, ID: node.ID}
		switch turn % 11 {
		case 0:
			change = OutlineChange{Kind: OutlineFilter, Filter: fmt.Sprintf("label:%d", rng.Intn(7))}
			filter = change.Filter
		case 1:
			change = OutlineChange{Kind: OutlineFilter}
			filter = ""
		case 2:
			change = OutlineChange{Kind: OutlineExpandAll}
			for _, n := range nodes {
				expanded[n.ID] = true
			}
		case 3:
			change = OutlineChange{Kind: OutlineCollapseAll}
			clear(expanded)
		default:
			expanded[node.ID] = !expanded[node.ID]
		}
		applyOutline(t, o, change)
		// This small oracle intentionally uses recursive traversal/ancestor walks,
		// independent of the worker's linked index and reverse-preorder algorithm.
		included := map[string]bool{}
		for _, n := range nodes {
			if filter == "" || strings.Contains(n.ID, filter) || strings.Contains(n.Label, filter) {
				for key := n.ID; key != ""; {
					included[key] = true
					for _, candidate := range nodes {
						if candidate.ID == key {
							key = candidate.Parent
							break
						}
					}
				}
			}
		}
		var want []string
		var visit func(string)
		visit = func(parent string) {
			for _, n := range children[parent] {
				if !included[n.ID] {
					continue
				}
				want = append(want, n.ID)
				if filter != "" || expanded[n.ID] {
					visit(n.ID)
				}
			}
		}
		visit("")
		var got []string
		for at := 0; at < o.VisibleLen(); at += 8 {
			rows, err := o.View(at, 8)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				got = append(got, row.ID)
				if row.Expanded != expanded[row.ID] {
					t.Fatal("stored expansion diverged", turn, row)
				}
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("visible order diverged", turn, got, want)
		}
	}
}

func outlineState(t *testing.T, o *Outline) OutlineState {
	t.Helper()
	s, err := o.State()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func applyOutline(t *testing.T, o *Outline, change OutlineChange) {
	t.Helper()
	r, err := o.Prepare(change)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Complete(r, outlineBuilt(t, r, 7)); err != nil {
		t.Fatal(err)
	}
}

func TestOutlineSnapshotNavigationAndReferenceLifetime(t *testing.T) {
	nodes := []TreeNode{{ID: "root"}, {ID: "a", Parent: "root"}, {ID: "b", Parent: "a"}, {ID: "c", Parent: "root"}, {ID: "other"}}
	o := testOutline(t, nodes, true)
	snap, err := o.Snapshot()
	if err != nil || len(snap.Rows) != 3 || snap.Rows[2].Clip.Height != 5 || snap.Rows[2].Rect.Y != 20 {
		t.Fatalf("snapshot %+v: %v", snap, err)
	}
	ref := snap.Rows[0].Ref
	if _, err := o.Row(ref); err != nil {
		t.Fatal(err)
	}
	request, err := o.Prepare(OutlineChange{Kind: OutlineCollapse, ID: "root"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := o.Row(ref); err != nil || o.VisibleLen() != 5 {
		t.Fatal("prepare retired accepted rows", err)
	}
	if err := o.Cancel(request); err != nil {
		t.Fatal(err)
	}
	if err := o.SelectRef(ref); err != nil {
		t.Fatal(err)
	}
	if _, err := o.Row(ref); !errors.Is(err, ErrTreeStale) {
		t.Fatal("selection did not retire gesture", err)
	}
	current, err := o.Current()
	if err != nil || current.ID != "root" {
		t.Fatal(current, err)
	}
	generation := outlineState(t, o).Generation
	if err := o.Select("root"); err != nil || outlineState(t, o).Generation != generation {
		t.Fatal("no-op changed generation", err)
	}
	if err := o.FirstChild(); err != nil || outlineState(t, o).Selected != "a" {
		t.Fatal(err)
	}
	if err := o.FirstChild(); err != nil || outlineState(t, o).Selected != "b" {
		t.Fatal(err)
	}
	if err := o.Parent(); err != nil || outlineState(t, o).Selected != "a" {
		t.Fatal(err)
	}
	if err := o.End(); err != nil || outlineState(t, o).Selected != "other" {
		t.Fatal(err)
	}
	if err := o.Move(math.MinInt); err != nil || outlineState(t, o).Selected != "root" {
		t.Fatal(err)
	}
	if err := o.Page(math.MaxInt); err != nil || outlineState(t, o).Selected != "other" {
		t.Fatal(err)
	}
	if err := o.Page(math.MinInt); err != nil || outlineState(t, o).Selected != "root" {
		t.Fatal(err)
	}
	if err := o.Page(1); err != nil || outlineState(t, o).Selected != "b" {
		t.Fatal(err)
	}
	if err := o.Home(); err != nil || outlineState(t, o).Selected != "root" {
		t.Fatal(err)
	}
	foreign := testOutline(t, nodes, true)
	snap, _ = foreign.Snapshot()
	if _, err := o.Row(snap.Rows[0].Ref); !errors.Is(err, ErrTreeStale) {
		t.Fatal("foreign reference admitted", err)
	}
	copyOutline := *o
	if _, err := copyOutline.Prepare(OutlineChange{}); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	if _, err := copyOutline.State(); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	if _, err := o.Current(); err != nil {
		t.Fatal("copy refusal changed original", err)
	}
	// A detached row ref only retains its identity token and exact owned key.
	typeGraph := reflect.TypeOf(OutlineRef{})
	for n := 0; n < typeGraph.NumField(); n++ {
		field := typeGraph.Field(n).Type
		if field.Kind() == reflect.Interface || field.Kind() == reflect.Func || field == reflect.TypeOf(o) || field == reflect.TypeOf(o.index) || field == reflect.TypeOf(o.index.data) {
			t.Fatal("row reference retains mutable domain", field)
		}
	}
}

func TestOutlineHiddenSelectionFilterAndStableDataReplacement(t *testing.T) {
	nodes := []TreeNode{{ID: "root"}, {ID: "a", Parent: "root"}, {ID: "b", Parent: "a", Label: "find me"}, {ID: "other"}}
	o := testOutline(t, nodes, true)
	if err := o.Select("b"); err != nil {
		t.Fatal(err)
	}
	applyOutline(t, o, OutlineChange{Kind: OutlineCollapseAll})
	s := outlineState(t, o)
	if s.Selected != "b" || s.SelectedVisible || s.VisibleRows != 2 {
		t.Fatal(s)
	}
	if _, err := o.Current(); !errors.Is(err, ErrTreeHidden) {
		t.Fatal(err)
	}
	if err := o.Reveal("b"); !errors.Is(err, ErrTreeHidden) {
		t.Fatal(err)
	}
	applyOutline(t, o, OutlineChange{Kind: OutlineFilter, Filter: "find me"})
	s = outlineState(t, o)
	if !s.SelectedVisible || s.Selected != "b" || s.VisibleRows != 3 || s.Filter != "find me" {
		t.Fatal(s)
	}
	rows, _ := o.View(0, 3)
	if rows[0].Expanded || rows[1].Expanded || rows[2].ID != "b" {
		t.Fatal("filter changed stored expansion", rows)
	}
	applyOutline(t, o, OutlineChange{Kind: OutlineFilter})
	if s = outlineState(t, o); s.SelectedVisible || s.Selected != "b" || s.VisibleRows != 2 {
		t.Fatal("clearing filter did not restore stored collapse", s)
	}
	applyOutline(t, o, OutlineChange{Kind: OutlineExpand, ID: "root"})
	applyOutline(t, o, OutlineChange{Kind: OutlineExpand, ID: "a"})
	// Change sibling input order and label without disturbing stable-key state.
	next := outlineData(t, "outline", 2, []TreeNode{{ID: "other"}, {ID: "b", Parent: "a", Label: "changed"}, {ID: "a", Parent: "root"}, {ID: "root"}}, TreeLimits{})
	r, err := o.PrepareData(next, OutlineChange{})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Complete(r, outlineBuilt(t, r, 1)); err != nil {
		t.Fatal(err)
	}
	if s = outlineState(t, o); s.Selected != "b" || !s.SelectedVisible || s.Stamp.Revision != 2 || s.VisibleRows != 4 {
		t.Fatal(s)
	}
	row, _ := o.Current()
	if row.Label != "changed" || row.Depth != 2 {
		t.Fatal(row)
	}
	if _, err := o.PrepareData(outlineData(t, "outline", 2, nodes, TreeLimits{}), OutlineChange{}); !errors.Is(err, ErrTreeStale) {
		t.Fatal("distinct same revision data admitted", err)
	}
	newIncarnation := outlineData(t, "different", 1, nodes, TreeLimits{})
	r, err = o.PrepareData(newIncarnation, OutlineChange{})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Complete(r, outlineBuilt(t, r, 7)); err != nil {
		t.Fatal(err)
	}
	if s = outlineState(t, o); s.Selected != "" || s.VisibleRows != 2 {
		t.Fatal("new incarnation inherited old state", s)
	}
}

func TestOutlineNewestRequestAtomicAdmissionAndUsableOldIndex(t *testing.T) {
	o := testOutline(t, []TreeNode{{ID: "a"}, {ID: "b", Parent: "a"}, {ID: "c", Parent: "b"}, {ID: "d"}}, false)
	snap, _ := o.Snapshot()
	oldRef := snap.Rows[0].Ref
	first, err := o.Prepare(OutlineChange{Kind: OutlineExpandAll})
	if err != nil {
		t.Fatal(err)
	}
	firstResult := outlineBuilt(t, first, 1)
	second, err := o.Prepare(OutlineChange{Kind: OutlineFilter, Filter: "c"})
	if err != nil {
		t.Fatal(err)
	}
	before := outlineState(t, o)
	if err := o.Complete(first, firstResult); !errors.Is(err, ErrTreeStale) || outlineState(t, o) != before {
		t.Fatal("stale completion changed state", err)
	}
	if err := o.Cancel(first); !errors.Is(err, ErrTreeStale) || outlineState(t, o) != before {
		t.Fatal("stale cancel changed state", err)
	}
	if _, err := o.Row(oldRef); err != nil {
		t.Fatal("new request retired old snapshot", err)
	}
	if err := o.Select("d"); err != nil {
		t.Fatal(err)
	}
	if err := o.SetViewport(OutlineViewport{Width: 90, Height: 25}); err != nil {
		t.Fatal(err)
	}
	result := outlineBuilt(t, second, 7)
	before = outlineState(t, o)
	foreignResult := outlineIndex(t, o.index.data, TreeIndexOptions{})
	if err := o.Complete(second, foreignResult); !errors.Is(err, ErrTreeStale) || outlineState(t, o) != before {
		t.Fatal("foreign result changed admission", err)
	}
	if err := o.Complete(second, result); err != nil {
		t.Fatal(err)
	}
	s := outlineState(t, o)
	if s.Pending || s.Selected != "d" || s.SelectedVisible || s.Viewport.Width != 90 || s.VisibleRows != 3 || s.Generation != before.Generation+1 {
		t.Fatal("completion lost concurrent owner navigation", s)
	}
	if _, err := o.Row(oldRef); !errors.Is(err, ErrTreeStale) {
		t.Fatal(err)
	}
	// Only an admitted new request supersedes pending work; malformed input does
	// not retire it or change accepted references/state.
	r, err := o.Prepare(OutlineChange{})
	if err != nil {
		t.Fatal(err)
	}
	before = outlineState(t, o)
	for _, bad := range []OutlineChange{{Kind: OutlineExpand, ID: "missing"}, {Kind: OutlineFilter, Filter: "\xff"}, {ID: "\x00"}, {Kind: OutlineKeep, Filter: "\xff"}, {Kind: OutlineFilter, Filter: strings.Repeat("x", MaxTreeQueryBytes+1)}, {Kind: OutlineChangeKind(255)}} {
		if _, err := o.Prepare(bad); err == nil || outlineState(t, o) != before {
			t.Fatal("failed request retired existing work", bad, err)
		}
	}
	if err := o.Cancel(r); err != nil || outlineState(t, o).Pending {
		t.Fatal(err)
	}
	if err := o.Complete(r, outlineBuilt(t, r, 7)); !errors.Is(err, ErrTreeStale) {
		t.Fatal(err)
	}
}

func TestOutlineAnchorAndCapacityAdmissionAreAtomic(t *testing.T) {
	nodes := []TreeNode{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}, {ID: "f"}}
	o := testOutline(t, nodes, true)
	if err := o.SetViewport(OutlineViewport{Width: 80, Height: 25, ScrollY: 13}); err != nil {
		t.Fatal(err)
	}
	next := outlineData(t, "outline", 2, []TreeNode{{ID: "z"}, {ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}, {ID: "f"}}, TreeLimits{})
	r, err := o.PrepareData(next, OutlineChange{})
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Complete(r, outlineBuilt(t, r, 7)); err != nil || outlineState(t, o).Viewport.ScrollY != 23 {
		t.Fatal("stable anchor/subrow offset lost", err, outlineState(t, o))
	}
	data := outlineData(t, "cap", 1, []TreeNode{{ID: "root"}, {ID: "a", Parent: "root"}, {ID: "b", Parent: "root"}}, TreeLimits{})
	limited, err := NewOutline(outlineIndex(t, data, TreeIndexOptions{}), OutlineConfig{RowHeight: 10, MaxRows: 1}, OutlineViewport{Width: 80, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	r, _ = limited.Prepare(OutlineChange{Kind: OutlineExpandAll})
	result := outlineBuilt(t, r, 7)
	before := outlineState(t, limited)
	if err := limited.Complete(r, result); !errors.Is(err, ErrTreeCapacity) || outlineState(t, limited) != before {
		t.Fatal("visible row capacity refusal partially published", err)
	}
	if err := limited.SetViewport(OutlineViewport{Width: 80, Height: 10}); err != nil {
		t.Fatal(err)
	}
	if err := limited.Complete(r, result); err != nil || limited.VisibleLen() != 3 {
		t.Fatal("capacity recovery cannot install existing result", err)
	}
	before = outlineState(t, limited)
	for _, v := range []OutlineViewport{{Height: 10}, {Width: 1}, {Width: 1, Height: 1, ScrollY: -1}, {Width: 80, Height: 11}} {
		if err := limited.SetViewport(v); err == nil || outlineState(t, limited) != before {
			t.Fatal("invalid viewport changed state", v, err)
		}
	}
	limited.generation = math.MaxUint64
	before = outlineState(t, limited)
	if err := limited.Select("b"); !errors.Is(err, ErrTreeExhausted) || outlineState(t, limited) != before {
		t.Fatal(err)
	}
	if _, err := limited.Prepare(OutlineChange{}); !errors.Is(err, ErrTreeExhausted) || outlineState(t, limited) != before {
		t.Fatal(err)
	}
	// Cancellation is still usable at the terminal generation for recovery.
	limited.generation--
	r, _ = limited.Prepare(OutlineChange{})
	result = outlineBuilt(t, r, 7)
	limited.generation = math.MaxUint64
	before = outlineState(t, limited)
	if err := limited.Complete(r, result); !errors.Is(err, ErrTreeExhausted) || outlineState(t, limited) != before {
		t.Fatal("exhausted completion changed state", err)
	}
	if err := limited.Cancel(r); err != nil {
		t.Fatal(err)
	}
}

func TestOutlineEmptyAndOverflowCoordinates(t *testing.T) {
	d := outlineData(t, "empty", 1, nil, TreeLimits{})
	i := outlineIndex(t, d, TreeIndexOptions{})
	o, err := NewOutline(i, OutlineConfig{RowHeight: 1}, OutlineViewport{Width: math.MaxInt, Height: math.MaxInt})
	if err != nil {
		t.Fatal(err)
	}
	s, err := o.Snapshot()
	if err != nil || len(s.Rows) != 0 || s.State.VisibleRows != 0 || s.State.TotalHeight != 0 {
		t.Fatal(s, err)
	}
	for _, nav := range []func() error{func() error { return o.Move(1) }, o.Home, o.End, o.Parent, o.FirstChild} {
		if nav() == nil {
			t.Fatal("empty navigation unexpectedly succeeded")
		}
	}
	if _, err := NewOutline(i, OutlineConfig{MaxRows: MaxTreeViewRows + 1}, OutlineViewport{Width: 10, Height: 10}); !errors.Is(err, ErrTreeCapacity) {
		t.Fatal(err)
	}
	if math.MaxInt <= math.MaxInt32 {
		// A 32-bit owner must refuse expanded extent before retained-anchor math.
		nodes := make([]TreeNode, 4096)
		for n := range nodes {
			nodes[n].ID = string(rune(n+1)) + "x"
			if n > 0 {
				nodes[n].Parent = nodes[0].ID
			}
		}
		d = outlineData(t, "narrow", 1, nodes, TreeLimits{})
		o, err = NewOutline(outlineIndex(t, d, TreeIndexOptions{}), OutlineConfig{RowHeight: 1 << 20, MaxRows: 1}, OutlineViewport{Width: 1, Height: 1})
		if err != nil {
			t.Fatal(err)
		}
		r, _ := o.Prepare(OutlineChange{Kind: OutlineExpandAll})
		before := outlineState(t, o)
		if err := o.Complete(r, outlineBuilt(t, r, 7)); !errors.Is(err, ErrTreeCapacity) || outlineState(t, o) != before {
			t.Fatal("32-bit extent overflow published", err)
		}
	}
	if _, err := NewOutline(nil, OutlineConfig{}, OutlineViewport{Width: 10, Height: 10}); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
	if _, err := NewOutlineBuilder(context.Background(), OutlineRequest{}); !errors.Is(err, ErrTreeInvalid) {
		t.Fatal(err)
	}
}

func TestOutlinePublicKeysBoundBeforeOwnerLookup(t *testing.T) {
	d := outlineData(t, "keys", 1, []TreeNode{{ID: "root"}, {ID: "kid", Parent: "root"}}, TreeLimits{MaxNodeTextBytes: 4})
	o, err := NewOutline(outlineIndex(t, d, TreeIndexOptions{ExpandAll: true}), OutlineConfig{}, OutlineViewport{Width: 100, Height: 24})
	if err != nil {
		t.Fatal(err)
	}
	request, err := o.Prepare(OutlineChange{})
	if err != nil {
		t.Fatal(err)
	}
	before := outlineState(t, o)
	for _, tc := range []struct {
		key  string
		want error
	}{
		{strings.Repeat("x", 8<<20), ErrTreeCapacity},
		{"12345", ErrTreeCapacity},
		{"\xff", ErrTreeSource},
		{"a\x00b", ErrTreeSource},
		{"", ErrTreeRange},
		{"nope", ErrTreeRange},
	} {
		if _, err := d.Node(tc.key); !errors.Is(err, tc.want) {
			t.Fatalf("Node key length=%d: %v want %v", len(tc.key), err, tc.want)
		}
		if err := o.Select(tc.key); !errors.Is(err, tc.want) || outlineState(t, o) != before {
			t.Fatalf("Select key length=%d: %v want %v; state changed=%v", len(tc.key), err, tc.want, outlineState(t, o) != before)
		}
		if err := o.Reveal(tc.key); !errors.Is(err, tc.want) || outlineState(t, o) != before {
			t.Fatalf("Reveal key length=%d: %v want %v; state changed=%v", len(tc.key), err, tc.want, outlineState(t, o) != before)
		}
	}
	if err := o.Cancel(request); err != nil {
		t.Fatal("refused public keys retired pending work", err)
	}
	if err := o.Reveal("kid"); err != nil || outlineState(t, o).Selected != "kid" {
		t.Fatal("bounded valid key refused", err)
	}
}
