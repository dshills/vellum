//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNativeAccessibilityUIntBoundaries(t *testing.T) {
	cases := []struct {
		literal        string
		positive, zero bool
	}{
		{"0", false, true}, {"0.0", false, true},
		{"1", true, true}, {"1.0", true, true}, {"1e3", true, true},
		{"127", true, true}, {"128", true, true}, {"255", true, true}, {"256", true, true},
		{"32767", true, true}, {"32768", true, true}, {"65535", true, true}, {"65536", true, true},
		{"2147483647", true, true}, {"2147483648", true, true},
		{"4294967295", true, true}, {"4294967296", true, true},
		{"9007199254740991", true, true}, {"9007199254740992", true, true},
		{"9223372036854775807", true, true}, {"9223372036854775808", true, true},
		{"18446744073709551615", true, true},
		{"-1", false, false}, {"-1.0", false, false}, {"1.5", false, false},
		{"1e-2", false, false}, {"18446744073709551616", false, false},
		{"1e20", false, false}, {"true", false, false}, {"false", false, false},
		{"null", false, false}, {`"1"`, false, false}, {"[]", false, false}, {"{}", false, false},
	}
	for _, c := range cases {
		t.Run(c.literal, func(t *testing.T) {
			payload := fmt.Sprintf(`{"generation":%s,"nodes":[]}`, c.literal)
			if got := nativeAccessibilitySnapshotValid(payload); got != c.positive {
				t.Fatalf("positive generation %s: %v, want %v", c.literal, got, c.positive)
			}
			for _, field := range []string{"grid_incarnation", "grid_target_generation", "grid_generation"} {
				node := map[string]any{"id": "table", "parent": "", "role": "table", "actions": []string{},
					"grid_scope": "table", "grid_incarnation": 1, "grid_target_generation": json.RawMessage("18446744073709551615"),
					"grid_generation": 0, "grid_pending": true, "disabled": true}
				node[field] = json.RawMessage(c.literal)
				bytes, err := json.Marshal(map[string]any{"generation": 1, "nodes": []any{node}})
				if err != nil {
					t.Fatal(err)
				}
				want := c.positive
				if field == "grid_generation" {
					want = c.zero
				}
				if got := nativeAccessibilitySnapshotValid(string(bytes)); got != want {
					t.Fatalf("%s=%s: %v, want %v", field, c.literal, got, want)
				}
			}
		})
	}
}

func BenchmarkNativeAccessibilitySnapshotUIntFields(b *testing.B) {
	frame := [4]float32{0, 0, 100, 20}
	makeNode := func(id, parent, role string) accessibilityNode {
		return accessibilityNode{ID: id, Parent: parent, Role: role, Actions: []string{}, GridScope: "table",
			GridIncarnation: 17, GridTargetGeneration: 12345, GridGeneration: 12345, GridFrame: &frame}
	}
	nodes := []accessibilityNode{makeNode("table", "", "table")}
	for row := 0; row < 50; row++ {
		id := fmt.Sprintf("row-%d", row)
		nodes = append(nodes, makeNode(id, "table", "grid-row"))
		for col := 0; col < 8; col++ {
			nodes = append(nodes, makeNode(fmt.Sprintf("%s-cell-%d", id, col), id, "grid-cell"))
		}
	}
	payload, err := json.Marshal(map[string]any{"generation": 12345, "nodes": nodes})
	if err != nil {
		b.Fatal(err)
	}
	text := string(payload)
	if !nativeAccessibilitySnapshotValid(text) {
		b.Fatal("benchmark snapshot refused")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !nativeAccessibilitySnapshotValid(text) {
			b.Fatal("snapshot refused")
		}
	}
}

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
	b.closeFrom("native-attachment-close")
	select {
	case err := <-done:
		var closed *AccessibilityClosedError
		if !errors.Is(err, ErrAccessibilityClosed) || !errors.As(err, &closed) || closed.Retirement.Window != "ax" || closed.Retirement.Token != b.token || closed.Retirement.Source != "native-attachment-close" {
			t.Fatal("closed request lacked first retirement provenance", err)
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

func TestNativeAccessibilityCompletedResultCannotHideRetirement(t *testing.T) {
	_, b := newAXTest(t, 8)
	_, err := accessibilityRequestRun(context.Background(), b, func(id uint64) {
		accessibilityResults.Lock()
		result := accessibilityResults.requests[id]
		accessibilityResults.Unlock()
		result <- "true"
		b.closeFrom("driver-window-cleanup")
	})
	var closed *AccessibilityClosedError
	if !errors.Is(err, ErrAccessibilityClosed) || !errors.As(err, &closed) || closed.Retirement.Source != "driver-window-cleanup" {
		t.Fatal("ready native result hid closed bridge", err)
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
