//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNativeAccessibilitySnapshotValidationIsAtomic(t *testing.T) {
	valid := `{"generation":1,"nodes":[{"id":"root","parent":"","role":"column","actions":[]},{"id":"child","parent":"root","role":"button","actions":["press"]}]}`
	if !nativeAccessibilitySnapshotValid(valid) {
		t.Fatal("valid native snapshot refused")
	}
	for _, invalid := range []string{
		`{"generation":2,"nodes":[{"id":"root","parent":"","role":"column","actions":[]},{}]}`,
		`{"generation":2,"nodes":[{"id":"root","parent":"","role":"column","actions":[]},{"id":"root","parent":"","role":"button","actions":[]}]}`,
		`{"generation":2,"nodes":[{"id":"child","parent":"missing","role":"button","actions":[]}]}`,
		`{"generation":2,"nodes":[{"id":"a","parent":"b","role":"column","actions":[]},{"id":"b","parent":"a","role":"column","actions":[]}]}`,
		`{"generation":2,"nodes":[{"id":"a","parent":"","role":"button","actions":[1]}]}`,
		`{"generation":2,"nodes":[{"id":"a","parent":"","role":"button","actions":[],"value":{}}]}`,
	} {
		if nativeAccessibilitySnapshotValid(invalid) {
			t.Fatal("malformed native snapshot admitted", invalid)
		}
	}
}

func TestNativeAccessibilityRequestCancelsOnAttachmentClose(t *testing.T) {
	_, b := newAXTest(t, 8)
	started := make(chan uint64, 1)
	done := make(chan error, 1)
	go func() {
		_, err := accessibilityRequestRun(context.Background(), b, func(id uint64) { started <- id })
		done <- err
	}()
	request := <-started
	b.close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("closed attachment returned a native result")
		}
	case <-time.After(time.Second):
		t.Fatal("request survived attachment close")
	}
	accessibilityResults.Lock()
	_, retained := accessibilityResults.requests[request]
	accessibilityResults.Unlock()
	if retained {
		t.Fatal("closed request leaked result registration")
	}
}
func TestNativeAccessibilityRequestBoundsBackgroundWait(t *testing.T) {
	_, b := newAXTest(t, 8)
	start := time.Now()
	_, err := accessibilityRequestRun(context.Background(), b, func(uint64) {})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("missing native completion did not time out", err)
	}
	if elapsed := time.Since(start); elapsed < accessibilityRequestTimeout || elapsed > accessibilityRequestTimeout+time.Second {
		t.Fatal("unexpected default request bound", elapsed)
	}
}

func TestNativeAccessibilityHitTestRetainedTreeOrder(t *testing.T) {
	node := func(id, parent string) accessibilityNode {
		return accessibilityNode{ID: id, Parent: parent, Role: "button", Actions: []string{}}
	}
	cases := []struct {
		name   string
		nodes  []accessibilityNode
		frames map[string][4]float64
		point  [2]float64
		want   string
	}{
		{"equal parent and child", []accessibilityNode{node("root", ""), node("child", "root")}, map[string][4]float64{"root": {0, 0, 40, 40}, "child": {0, 0, 40, 40}}, [2]float64{20, 20}, "child"},
		{"later equal sibling", []accessibilityNode{node("root", ""), node("first", "root"), node("last", "root")}, map[string][4]float64{"root": {0, 0, 40, 40}, "first": {0, 0, 40, 40}, "last": {0, 0, 40, 40}}, [2]float64{20, 20}, "last"},
		{"reordered equal sibling", []accessibilityNode{node("root", ""), node("last", "root"), node("first", "root")}, map[string][4]float64{"root": {0, 0, 40, 40}, "first": {0, 0, 40, 40}, "last": {0, 0, 40, 40}}, [2]float64{20, 20}, "first"},
		{"zero frame ancestor", []accessibilityNode{node("root", ""), node("group", "root"), node("child", "group")}, map[string][4]float64{"root": {0, 0, 0, 0}, "group": {0, 0, 0, 0}, "child": {10, 10, 40, 40}}, [2]float64{20, 20}, "child"},
		{"later root overlays earlier descendant", []accessibilityNode{node("early", ""), node("child", "early"), node("late", "")}, map[string][4]float64{"early": {0, 0, 40, 40}, "child": {0, 0, 20, 20}, "late": {0, 0, 40, 40}}, [2]float64{10, 10}, "late"},
		{"noncontaining later sibling", []accessibilityNode{node("root", ""), node("first", "root"), node("last", "root")}, map[string][4]float64{"root": {0, 0, 40, 40}, "first": {0, 0, 40, 40}, "last": {100, 100, 40, 40}}, [2]float64{20, 20}, "first"},
		{"outside all frames", []accessibilityNode{node("root", ""), node("child", "root")}, map[string][4]float64{"root": {0, 0, 40, 40}, "child": {0, 0, 40, 40}}, [2]float64{100, 100}, ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			snapshot, err := json.Marshal(struct {
				Generation uint64              `json:"generation"`
				Nodes      []accessibilityNode `json:"nodes"`
			}{1, test.nodes})
			if err != nil {
				t.Fatal(err)
			}
			geometry, err := json.Marshal(test.frames)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 32; i++ {
				if got := nativeAccessibilityHitFixture(string(snapshot), string(geometry), test.point[0], test.point[1]); got != test.want {
					t.Fatalf("native hit %d=%q want %q", i, got, test.want)
				}
			}
		})
	}
}
