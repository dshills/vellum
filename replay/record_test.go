package replay

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func fixture() Manifest {
	return Manifest{Version: Version, FixtureSHA256: strings.Repeat("a", 64), ThemeSHA256: strings.Repeat("b", 64), LayoutVersion: "layout-1", RendererVersion: "software-1", Start: time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), Seed: 17, Windows: []Window{{ID: "main", Incarnation: 1, Width: 800, Height: 600, DPI: 96}}, Fonts: []Font{{Family: "fixture", SHA256: strings.Repeat("c", 64)}}}
}
func action() Operation {
	return Operation{Kind: Command, Window: "main", WindowIncarnation: 1, Target: "save", TargetIncarnation: 9, CheckRevision: true, ExpectedRevision: 3, Action: "save", Arguments: []Argument{{Key: "name", Value: "A😀\r\n"}}}
}
func TestRetentionAndRefusalAtomicity(t *testing.T) {
	m := fixture()
	r, e := NewRecorder(m, Limits{MaxOperations: 2})
	if e != nil {
		t.Fatal(e)
	}
	m.Windows[0].ID = "caller mutation"
	m.Fonts[0].Family = "caller mutation"
	o := action()
	if e = r.Append(o); e != nil {
		t.Fatal(e)
	}
	o.Arguments[0].Value = "caller mutation"
	before, _ := r.Snapshot()
	hash, _ := before.Digest(Limits{})
	bad := action()
	bad.WindowIncarnation = 2
	if e = r.Append(bad); !errors.Is(e, ErrInvalid) {
		t.Fatal(e)
	}
	after, _ := r.Snapshot()
	got, _ := after.Digest(Limits{})
	if got != hash {
		t.Fatal("refusal mutated record")
	}
	if after.Manifest.Windows[0].ID != "main" || after.Manifest.Fonts[0].Family != "fixture" || after.Operations[0].Arguments[0].Value != "A😀\r\n" {
		t.Fatal("caller alias retained")
	}
	after.Manifest.Windows[0].ID = "snapshot mutation"
	after.Operations[0].Arguments[0].Value = "snapshot mutation"
	unchanged, _ := r.Snapshot()
	got, _ = unchanged.Digest(Limits{})
	if got != hash {
		t.Fatal("snapshot aliases record")
	}
	if e = r.Append(Operation{Kind: Worker, WorkerID: 7}); e != nil {
		t.Fatal(e)
	}
	if e = r.Append(Operation{Kind: Worker, WorkerID: 8}); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
}
func TestEveryVariantAndCrossVariantFields(t *testing.T) {
	cases := []Operation{
		{Kind: PointerDown, Window: "main", WindowIncarnation: 1, X: 4, Y: 5, Button: 1},
		{Kind: PointerUp, Window: "main", WindowIncarnation: 1, X: 4, Y: 5, Button: 1},
		{Kind: PointerMove, Window: "main", WindowIncarnation: 1, X: 4, Y: 5},
		{Kind: KeyDown, Window: "main", WindowIncarnation: 1, Key: "A"},
		{Kind: KeyUp, Window: "main", WindowIncarnation: 1, Key: "A"},
		{Kind: KeyChord, Window: "main", WindowIncarnation: 1, Key: "Meta+Z"}, action(),
		{Kind: Semantic, Window: "main", WindowIncarnation: 1, Target: "save", TargetIncarnation: 9, Action: "activate"},
		{Kind: Advance, Window: "main", WindowIncarnation: 1, Delta: time.Second},
		{Kind: Worker, WorkerID: 2}, {Kind: Frame, Window: "main", WindowIncarnation: 1}, {Kind: Capture, Label: "final"},
	}
	for _, o := range cases {
		t.Run(string(o.Kind), func(t *testing.T) {
			r, e := NewRecorder(fixture(), Limits{})
			if e != nil {
				t.Fatal(e)
			}
			if e = r.Append(o); e != nil {
				t.Fatal(e)
			}
			o.WorkerID = 99
			if o.Kind == Worker {
				o.Action = "unrelated"
			}
			if e = r.Append(o); !errors.Is(e, ErrInvalid) {
				t.Fatalf("variant field accepted: %v", e)
			}
		})
	}
}
func TestInvalidPortableBoundsAndSequence(t *testing.T) {
	mutations := []func(*Operation){
		func(o *Operation) { o.Kind = "unknown" }, func(o *Operation) { o.Sequence = 9 }, func(o *Operation) { o.At = -1 }, func(o *Operation) { o.X = math.NaN() }, func(o *Operation) { o.Y = math.Inf(1) }, func(o *Operation) { o.Window = "unregistered" }, func(o *Operation) { o.WindowIncarnation = 0 }, func(o *Operation) { o.TargetIncarnation = 0 }, func(o *Operation) { o.CheckRevision = false }, func(o *Operation) { o.Action = "NUL\x00" }, func(o *Operation) { o.Action = string([]byte{0xff}) }, func(o *Operation) { o.Arguments = append(o.Arguments, o.Arguments[0]) }, func(o *Operation) { o.Arguments = make([]Argument, 33) }, func(o *Operation) { o.Modifiers = 16 }, func(o *Operation) { o.Delta = -1 },
	}
	for i, mutate := range mutations {
		r, _ := NewRecorder(fixture(), Limits{})
		o := action()
		mutate(&o)
		if e := r.Append(o); !errors.Is(e, ErrInvalid) {
			t.Fatalf("case %d: %v", i, e)
		}
		v, _ := r.Snapshot()
		if len(v.Operations) != 0 {
			t.Fatal("partial admission")
		}
	}
	r, _ := NewRecorder(fixture(), Limits{})
	o := action()
	o.At = 2
	if e := r.Append(o); e != nil {
		t.Fatal(e)
	}
	o.At = 1
	if e := r.Append(o); !errors.Is(e, ErrInvalid) {
		t.Fatal(e)
	}
	for _, o := range []Operation{{Kind: PointerDown, Window: "main", WindowIncarnation: 1, X: 800, Y: 0, Button: 1}, {Kind: PointerUp, Window: "main", WindowIncarnation: 1, X: 0, Y: 600, Button: 1}, {Kind: PointerMove, Window: "main", WindowIncarnation: 1, X: -1}} {
		if e := r.Append(o); !errors.Is(e, ErrInvalid) {
			t.Fatal(e)
		}
	}
}
func TestExactCanonicalByteLimit(t *testing.T) {
	rec := Record{Manifest: fixture(), Operations: []Operation{}}
	base, _ := json.Marshal(rec)
	o := action()
	o.Sequence = 1
	b, _ := json.Marshal(o)
	maximum := len(base) + len(b)
	r, e := NewRecorder(fixture(), Limits{MaxBytes: maximum, MaxPayloadBytes: 64})
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Append(action()); e != nil {
		t.Fatal(e)
	}
	v, _ := r.Snapshot()
	encoded, e := v.Encode(Limits{MaxBytes: maximum, MaxPayloadBytes: 64})
	if e != nil || len(encoded) != maximum {
		t.Fatalf("canonical size %d want %d err %v", len(encoded), maximum, e)
	}
	if e = r.Append(action()); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	r, e = NewRecorder(fixture(), Limits{MaxBytes: maximum - 1, MaxPayloadBytes: 64})
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Append(action()); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	v.Operations = append(v.Operations, v.Operations[0])
	v.Operations[1].Sequence = 2
	if e = v.Validate(Limits{MaxBytes: maximum, MaxPayloadBytes: 64}); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	empty := Record{Manifest: fixture()}
	a, e := empty.Encode(Limits{})
	if e != nil {
		t.Fatal(e)
	}
	if string(a) != string(base) {
		t.Fatal("nil operations have distinct encoding")
	}
}
func TestPayloadAndManifestRefusal(t *testing.T) {
	r, _ := NewRecorder(fixture(), Limits{MaxPayloadBytes: 64})
	o := action()
	o.Arguments[0].Value = strings.Repeat("z", 65)
	if e := r.Append(o); !errors.Is(e, ErrInvalid) && !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	for _, mutate := range []func(*Manifest){func(m *Manifest) { m.Version++ }, func(m *Manifest) { m.FixtureSHA256 = "abc" }, func(m *Manifest) { m.Fonts[0].SHA256 = strings.Repeat("C", 64) }, func(m *Manifest) { m.Windows[0].DPI = 0 }, func(m *Manifest) { m.Windows = append(m.Windows, m.Windows[0]) }, func(m *Manifest) { m.Fonts = append(m.Fonts, m.Fonts[0]) }, func(m *Manifest) { m.Start = m.Start.In(time.FixedZone("ambient", 3600)) }} {
		m := fixture()
		mutate(&m)
		if _, e := NewRecorder(m, Limits{}); !errors.Is(e, ErrInvalid) {
			t.Fatal(e)
		}
	}
	for _, l := range []Limits{{MaxOperations: -1}, {MaxBytes: 65 << 20}, {MaxPayloadBytes: 2 << 20}, {MaxBytes: 1, MaxPayloadBytes: 2}} {
		if _, e := NewRecorder(fixture(), l); !errors.Is(e, ErrInvalid) {
			t.Fatal(e)
		}
	}
	var zero Recorder
	if e := zero.Append(action()); !errors.Is(e, ErrInvalid) {
		t.Fatal(e)
	}
	r, _ = NewRecorder(fixture(), Limits{})
	var copy Recorder
	reflect.ValueOf(&copy).Elem().Set(reflect.ValueOf(r).Elem())
	if e := copy.Append(action()); !errors.Is(e, ErrInvalid) {
		t.Fatal(e)
	}
}
func TestConcurrentProducersContiguousOrder(t *testing.T) {
	r, _ := NewRecorder(fixture(), Limits{MaxOperations: 1000})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := r.Append(Operation{Kind: Worker, WorkerID: 1}); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	rec, _ := r.Snapshot()
	if len(rec.Operations) != 100 {
		t.Fatal(len(rec.Operations))
	}
	if e := rec.Validate(Limits{}); e != nil {
		t.Fatal(e)
	}
	for i, o := range rec.Operations {
		if o.Sequence != uint64(i+1) {
			t.Fatal("noncontiguous order")
		}
	}
}

func TestOversizedManifestRefusesBeforeBulkCopy(t *testing.T) {
	m := fixture()
	m.Windows = make([]Window, 100000)
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := NewRecorder(m, Limits{})
	runtime.ReadMemStats(&after)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	// A 100k caller window slice is far outside the32-window admission limit.
	// Refusal must not allocate another copy of that oversized backing storage.
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatal("refusal copied oversized caller manifest")
	}
	runtime.KeepAlive(m)
}
func TestRetainedTextDoesNotPinOversizedCallerStorage(t *testing.T) {
	largeHash := strings.Repeat("a", 1<<20)
	largeID := "main" + strings.Repeat("z", 1<<20)
	largeText := strings.Repeat("s", 1<<20)
	m := fixture()
	m.FixtureSHA256 = largeHash[:64]
	m.Windows[0].ID = largeID[:4]
	r, err := NewRecorder(m, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	o := action()
	o.Window = largeID[:4]
	o.Action = largeText[:4]
	o.Arguments[0].Value = largeText[:8]
	if err = r.Append(o); err != nil {
		t.Fatal(err)
	}
	// Same-package inspection avoids Snapshot's separate defensive clone masking
	// the retained-storage boundary that admission itself must satisfy.
	retained := r.record
	for _, pair := range [][2]string{{retained.Manifest.FixtureSHA256, m.FixtureSHA256}, {retained.Manifest.Windows[0].ID, m.Windows[0].ID}, {retained.Operations[0].Window, o.Window}, {retained.Operations[0].Action, o.Action}, {retained.Operations[0].Arguments[0].Value, o.Arguments[0].Value}} {
		if pair[0] != pair[1] || unsafe.StringData(pair[0]) == unsafe.StringData(pair[1]) {
			t.Fatal("retained text aliases caller backing storage")
		}
	}
}

func TestOversizedHashRefusesBeforeDecodeAllocation(t *testing.T) {
	m := fixture()
	m.FixtureSHA256 = strings.Repeat("a", 8<<20)
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := NewRecorder(m, Limits{})
	runtime.ReadMemStats(&after)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatal("refusal decoded oversized hash")
	}
	runtime.KeepAlive(m)
}
func TestRetainedKindDoesNotPinCallerStorage(t *testing.T) {
	large := "frame" + strings.Repeat("x", 1<<20)
	o := Operation{Kind: Kind(large[:5]), Window: "main", WindowIncarnation: 1}
	r, err := NewRecorder(fixture(), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = r.Append(o); err != nil {
		t.Fatal(err)
	}
	if unsafe.StringData(string(r.record.Operations[0].Kind)) == unsafe.StringData(string(o.Kind)) {
		t.Fatal("kind aliases caller storage")
	}
}
