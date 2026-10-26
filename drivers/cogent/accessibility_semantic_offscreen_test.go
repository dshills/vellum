//go:build offscreen

// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestAccessibilityBuiltinSemanticSchema(t *testing.T) {
	expected := [][3]string{{"ID", "string", "id"},
		{"DomainID", "string", "domain_id"},
		{"Parent", "string", "parent"},
		{"Role", "string", "role"},
		{"Secure", "bool", "secure,omitempty"},
		{"Name", "string", "name"},
		{"Help", "string", "help"},
		{"Value", "interface {}", "value,omitempty"},
		{"Disabled", "bool", "disabled"},
		{"Focusable", "bool", "focusable"},
		{"Focused", "bool", "focused"},
		{"Selected", "bool", "selected"},
		{"Actions", "[]string", "actions"},
		{"Orientation", "string", "orientation,omitempty"},
		{"RowIndex", "int", "row_index,omitempty"},
		{"RowCount", "int", "row_count,omitempty"},
		{"ColumnIndex", "int", "column_index,omitempty"},
		{"ColumnCount", "int", "column_count,omitempty"},
		{"GridGeneration", "uint64", "grid_generation,omitempty"},
		{"GridScope", "string", "grid_scope,omitempty"},
		{"GridIncarnation", "uint64", "grid_incarnation,omitempty"},
		{"GridPending", "bool", "grid_pending,omitempty"},
		{"GridTargetGeneration", "uint64", "grid_target_generation,omitempty"},
		{"GridFrame", "*[4]float32", "grid_frame,omitempty"},
		{"Level", "int", "level,omitempty"},
		{"Expanded", "*bool", "expanded,omitempty"},
		{"Min", "*float32", "min,omitempty"},
		{"TextClient", "uint64", "text_client,omitempty"},
		{"TextSelection", "*[2]int", "text_selection,omitempty"},
		{"TextFocus", "int", "text_focus,omitempty"},
		{"NativeContent", "*ui.NativeContentInput", "native_content,omitempty"},
		{"Max", "*float32", "max,omitempty"}}
	typ := reflect.TypeOf(accessibilityNode{})
	var actual [][3]string
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.IsExported() {
			actual = append(actual, [3]string{f.Name, f.Type.String(), f.Tag.Get("json")})
		}
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatal("fixed wire schema changed; audit serializer", actual)
	}
	for _, v := range []any{accessibilityNode{}, &accessibilityNode{}} {
		if _, ok := v.(json.Marshaler); ok {
			t.Fatal("node custom JSON marshaler requires audit")
		}
		if _, ok := v.(encoding.TextMarshaler); ok {
			t.Fatal("node custom text marshaler requires audit")
		}
	}
}
func assertBuiltinEncoding(t *testing.T, nodes []accessibilityNode) {
	t.Helper()
	want, werr := json.Marshal(nodes)
	got, err := accessibilitySemantic(nodes)
	if (werr == nil) != (err == nil) || err == nil && got != string(want) {
		t.Fatalf("wire mismatch\nwant=%q\ngot =%q\nwantErr=%v gotErr=%v", want, got, werr, err)
	}
	if err != nil && (got != "" || reflect.TypeOf(err) != reflect.TypeOf(werr) || err.Error() != werr.Error()) {
		t.Fatal("fallback error or partial output changed", err, werr)
	}
}
func TestAccessibilityBuiltinSemanticStringsAndScalars(t *testing.T) {
	assertBuiltinEncoding(t, []accessibilityNode{{ID: "secure", Role: "text-field", Secure: true, Name: "Synthetic protected entry", Actions: []string{}}})
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	for _, s := range []string{"", "<>&\"\\\n\u2028\u2029\xff", string(all), "A😀עברית�", "\xed\xa0\x80\xc0\x80", "abc\x00def\x7f"} {
		assertBuiltinEncoding(t, []accessibilityNode{{ID: s, DomainID: s, Parent: s, Role: s, Name: s, Help: s, Value: s, Actions: []string{s}, Orientation: s, GridScope: s}})
	}
	values := []any{nil, false, true, "", int(0), int(-1), int8(-128), int16(-32768), int32(math.MinInt32), int64(math.MinInt64), uint(0), uint8(255), uint16(65535), uint32(math.MaxUint32), uint64(math.MaxUint64), uintptr(1), float32(0), float32(math.Copysign(0, -1)), float32(1e-6), float32(1e21), float32(math.SmallestNonzeroFloat32), math.Copysign(0, -1), 1e-6, 1e-7, 1e21, math.SmallestNonzeroFloat64, math.MaxFloat64, math.Inf(1), math.NaN(), json.RawMessage(` {"x":"<>&"} `), func() {}}
	zero := float32(0)
	no := false
	sel := [2]int{0, 0}
	frame := [4]float32{0, math.Float32frombits(1), math.Float32frombits(0x80000000), float32(1e21)}
	for _, v := range values {
		for _, actions := range [][]string{nil, {}, {"focus", "press"}} {
			assertBuiltinEncoding(t, []accessibilityNode{{Value: v, Actions: actions, Min: &zero, Max: &zero, Expanded: &no, TextSelection: &sel, GridFrame: &frame}})
		}
	}
}
func TestAccessibilityBuiltinSemanticRandomizedWire(t *testing.T) {
	r := rand.New(rand.NewSource(1469))
	text := func() string {
		b := make([]byte, r.Intn(48))
		for i := range b {
			b[i] = byte(r.Intn(256))
		}
		return string(b)
	}
	finite32 := func() float32 {
		for {
			f := math.Float32frombits(r.Uint32())
			if accessibilityFinite(float64(f)) {
				return f
			}
		}
	}
	finite64 := func() float64 {
		for {
			f := math.Float64frombits(r.Uint64())
			if accessibilityFinite(f) {
				return f
			}
		}
	}
	for i := 0; i < 2500; i++ {
		f := [4]float32{finite32(), finite32(), finite32(), finite32()}
		lo, hi := finite32(), finite32()
		expanded := r.Intn(2) == 0
		selection := [2]int{int(r.Int63()), -int(r.Int63())}
		values := []any{nil, text(), r.Intn(2) == 0, int(r.Int63()), int8(r.Intn(255) - 128), int16(r.Intn(65535) - 32768), int32(r.Uint32()), int64(r.Uint64()), uint(r.Uint64()), uint8(r.Uint32()), uint16(r.Uint32()), r.Uint32(), r.Uint64(), uintptr(r.Uint64()), finite32(), finite64()}
		n := accessibilityNode{ID: text(), DomainID: text(), Parent: text(), Role: text(), Name: text(), Help: text(), Value: values[i%len(values)], Disabled: r.Intn(2) == 0, Focusable: r.Intn(2) == 0, Focused: r.Intn(2) == 0, Selected: r.Intn(2) == 0, Actions: []string{text(), text()}, Orientation: text(), RowIndex: int(r.Int63()), RowCount: int(r.Int63()), ColumnIndex: int(r.Int63()), ColumnCount: int(r.Int63()), GridGeneration: r.Uint64(), GridScope: text(), GridIncarnation: r.Uint64(), GridPending: r.Intn(2) == 0, GridTargetGeneration: r.Uint64(), GridFrame: &f, Level: int(r.Int63()), Expanded: &expanded, Min: &lo, TextClient: r.Uint64(), TextSelection: &selection, TextFocus: int(r.Int63()), Max: &hi}
		if i%2 == 0 {
			n.Actions = nil
			n.GridFrame = nil
			n.Min = nil
			n.Max = nil
			n.Expanded = nil
			n.TextSelection = nil
		}
		if i%3 == 0 {
			n.Orientation = ""
			n.GridScope = ""
			n.RowIndex = 0
			n.RowCount = 0
			n.ColumnIndex = 0
			n.ColumnCount = 0
			n.GridGeneration = 0
			n.GridIncarnation = 0
			n.GridTargetGeneration = 0
			n.Level = 0
			n.TextClient = 0
			n.TextFocus = 0
		}
		assertBuiltinEncoding(t, []accessibilityNode{n, {Actions: []string{}, Value: false}})
	}
	// Every exponent range and exact random finite bit pattern must match, including
	// float32 cutovers that differ from comparisons after widening to float64.
	for i := 0; i < 10000; i++ {
		assertBuiltinEncoding(t, []accessibilityNode{{Value: finite32()}, {Value: finite64()}})
	}
}

type semanticMutator struct {
	calls *int
	fix   func()
}

func (v semanticMutator) MarshalJSON() ([]byte, error) {
	*v.calls++
	v.fix()
	return []byte(`"changed"`), nil
}
func TestAccessibilityBuiltinSemanticWholeSnapshotFallback(t *testing.T) {
	fixture := func() ([]accessibilityNode, *int) {
		calls := 0
		nodes := []accessibilityNode{{}, {Value: math.NaN()}}
		nodes[0].Value = semanticMutator{&calls, func() { nodes[1].Value = "repaired" }}
		return nodes, &calls
	}
	oracle, oc := fixture()
	actual, ac := fixture()
	if _, ok := accessibilityBuiltinSemantic(actual); ok || *ac != 0 {
		t.Fatal("custom value invoked by eligibility scan")
	}
	want, werr := json.Marshal(oracle)
	got, err := accessibilitySemantic(actual)
	if werr != nil || err != nil || got != string(want) || *oc != 1 || *ac != 1 {
		t.Fatal("whole-call fallback changed marshaler order/effects", werr, err, *oc, *ac)
	}
	for _, nodes := range [][]accessibilityNode{nil, {}, {{Actions: []string{}}}, gridAXPayloadFixture(400)} {
		got, err := accessibilitySemantic(nodes)
		if err != nil {
			t.Fatal(err)
		}
		want, oracleErr := json.Marshal(nodes)
		if oracleErr != nil {
			t.Fatal(oracleErr)
		}
		for range 4 {
			accessibilitySemantic(gridAXPayloadFixture(17))
		}
		if got != string(want) {
			t.Fatal("retained semantic string overwritten")
		}
	}
}
func BenchmarkAccessibilitySemanticBuiltin(b *testing.B) {
	for _, count := range []int{1, 400} {
		nodes := gridAXPayloadFixture(count)
		wire, _ := json.Marshal(nodes)
		hint := len(wire) + 16*len(nodes)
		b.Run(fmt.Sprintf("stdlib/%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var out strings.Builder
				if err := json.NewEncoder(&out).Encode(nodes); err != nil {
					b.Fatal(err)
				}
				accessibilityPayloadBenchmarkSink = strings.TrimSuffix(out.String(), "\n")
			}
		})
		b.Run(fmt.Sprintf("builtin/%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s, err := accessibilitySemanticHinted(nodes, hint)
				if err != nil {
					b.Fatal(err)
				}
				accessibilityPayloadBenchmarkSink = s
			}
		})
	}
}

func TestAccessibilitySemanticAllocationHintIsOnlyCapacity(t *testing.T) {
	nodes := gridAXPayloadFixture(4)
	b := &accessibilityBridge{lastSnapshot: "prior", pendingBaseline: nodes}
	if b.semanticAllocationHint(len(nodes)+1) != 0 || b.semanticAllocationHint(0) != 0 || b.semanticAllocationHint(len(nodes)) == 0 {
		t.Fatal("mismatched membership supplied capacity")
	}
	want, _ := json.Marshal(nodes)
	for _, hint := range []int{-1, 0, 1, len(want), 1 << 20, int(^uint(0) >> 1)} {
		got, err := accessibilitySemanticHinted(nodes, hint)
		if err != nil || got != string(want) {
			t.Fatal("capacity hint altered wire", hint, err)
		}
	}
	// Ready references also certify only count. The capacity read dereferences no
	// cached record, so stale callbacks or a detached snapshot cannot gain authority.
	b.pendingBaseline = nil
	b.readyBaseline = make([]accessibilityNodeRef, len(nodes))
	if b.semanticAllocationHint(len(nodes)) > 1<<20 {
		t.Fatal("unbounded hint")
	}
}
