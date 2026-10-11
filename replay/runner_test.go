package replay

// This backend tests the runner protocol, bounds and ownership. It does not
// stand in for the separate actual managed widget/router acceptance.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

type protocolBackend struct {
	state      Readiness
	applied    []Operation
	quanta     []int
	dispatched int
	calls      int
	capture    Observation
	check      func(Manifest) error
	apply      func(Operation) error
	step       func(int) (Progress, error)
}

func protocol() *protocolBackend {
	return &protocolBackend{state: Readiness{Now: fixture().Start}, capture: Observation{Data: json.RawMessage(`{"count":1}`), ImageSHA256: strings.Repeat("d", 64)}}
}
func (b *protocolBackend) CheckManifest(_ context.Context, m Manifest) error {
	b.calls++
	if b.check != nil {
		return b.check(m)
	}
	return nil
}
func (b *protocolBackend) Apply(_ context.Context, o Operation) error {
	b.calls++
	b.applied = append(b.applied, o)
	if b.apply != nil {
		return b.apply(o)
	}
	if o.Kind == Advance {
		b.state.Now = b.state.Now.Add(o.Delta)
	} else {
		b.state.Ready += 9
	}
	return nil
}
func (b *protocolBackend) Readiness(context.Context) (Readiness, error) {
	b.calls++
	return b.state, nil
}
func (b *protocolBackend) Step(_ context.Context, n int) (Progress, error) {
	b.calls++
	b.quanta = append(b.quanta, n)
	if b.step != nil {
		return b.step(n)
	}
	count := min(n, b.state.Ready)
	b.state.Ready -= count
	b.dispatched += count
	return Progress{Units: count, Events: count}, nil
}
func (b *protocolBackend) Capture(context.Context, Operation) (Observation, error) {
	b.calls++
	return b.capture, nil
}
func recordWith(operations ...Operation) Record {
	for i := range operations {
		operations[i].Sequence = uint64(i) + 1
	}
	return Record{Manifest: fixture(), Operations: operations}
}

func TestRunnerProtocolBatchesAndExplicitTime(t *testing.T) {
	var expected Result
	for _, batch := range []int{1, 7, 64} {
		b := protocol()
		r, err := NewRunner(RunnerConfig{EventBatch: batch}, b)
		if err != nil {
			t.Fatal(err)
		}
		record := recordWith(action(), Operation{Kind: Advance, Window: "main", WindowIncarnation: 1, Delta: time.Second}, Operation{Kind: Frame, Window: "main", WindowIncarnation: 1, At: time.Second}, Operation{Kind: Capture, Window: "main", WindowIncarnation: 1, At: time.Second, Label: "final"})
		got, err := r.Run(context.Background(), record)
		if err != nil {
			t.Fatal(err)
		}
		if got.Applied != 4 || got.Progress.Units != 18 || b.dispatched != 18 || !got.QuiescentNow || !b.state.Now.Equal(fixture().Start.Add(time.Second)) {
			t.Fatalf("batch %d: %+v", batch, got)
		}
		for _, quantum := range b.quanta {
			if quantum < 1 || quantum > batch {
				t.Fatal("unbounded quantum", quantum)
			}
		}
		if batch == 1 {
			expected = got
		} else if !reflect.DeepEqual(expected, got) {
			t.Fatalf("batch %d differs", batch)
		}
	}
}
func TestRunnerFutureOnlyAndUnreadyWorkRemainVisible(t *testing.T) {
	for _, kind := range []string{"future", "due", "running", "awaiting", "blocked"} {
		t.Run(kind, func(t *testing.T) {
			b := protocol()
			switch kind {
			case "future":
				b.state.FutureTimers = 2
				b.state.NextDeadline = b.state.Now.Add(time.Hour)
			case "due":
				b.state.DueTimers = 1
				b.state.NextDeadline = b.state.Now
			case "running":
				b.state.Running = 1
			case "awaiting":
				b.state.Awaiting = 1
			case "blocked":
				b.state.Ready = 4
				b.state.Blocked = true
				b.step = func(int) (Progress, error) { return Progress{}, nil }
			}
			r, _ := NewRunner(RunnerConfig{}, b)
			got, err := r.RunUntilIdle(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got.QuiescentNow != (kind == "future") || !got.Residual.Now.Equal(fixture().Start) || got.Progress.Units != 0 {
				t.Fatal(got)
			}
			if kind != "blocked" && len(b.quanta) != 0 {
				t.Fatal("hidden step")
			}
		})
	}
	if (Readiness{}).QuiescentNow() {
		t.Fatal("uninitialized state called quiescent")
	}
}
func TestRunnerCanCancelAndSelectAwaitingWorkerExplicitly(t *testing.T) {
	b := protocol()
	b.apply = func(o Operation) error {
		switch o.Action {
		case "start":
			b.state.Awaiting = 1
		case "cancel":
			if b.state.Awaiting != 1 {
				t.Fatal("missing admitted work")
			}
		}
		if o.Kind == Worker {
			b.state.Awaiting = 0
			b.state.Ready = 1
		}
		return nil
	}
	r, _ := NewRunner(RunnerConfig{}, b)
	start := action()
	start.Action = "start"
	cancel := action()
	cancel.Action = "cancel"
	got, err := r.Run(context.Background(), recordWith(start, cancel, Operation{Kind: Worker, WorkerID: 7}))
	if err != nil || got.Applied != 3 || got.Progress.Units != 1 || !got.QuiescentNow || len(b.applied) != 3 {
		t.Fatal(got, err)
	}
}
func TestRunnerBoundsRefuseBeforeBackendAndReportAppliedPrefix(t *testing.T) {
	b := protocol()
	r, _ := NewRunner(RunnerConfig{MaxSteps: 2}, b)
	bad := action()
	bad.Action = "bad\x00"
	if _, err := r.Run(context.Background(), recordWith(action(), bad)); !errors.Is(err, ErrInvalid) || b.calls != 0 {
		t.Fatal("invalid record reached backend", err, b.calls)
	}
	got, err := r.Run(context.Background(), recordWith(action()))
	if !errors.Is(err, ErrBudget) || got.Applied != 1 || got.Progress.Units != 1 || got.Residual.Ready != 8 || got.QuiescentNow {
		t.Fatal(got, err)
	}
}
func TestRunnerRecordAndCaptureStorageAreDetached(t *testing.T) {
	b := protocol()
	record := recordWith(action(), Operation{Kind: Capture, Window: "main", WindowIncarnation: 1, Label: "before"}, action())
	b.check = func(m Manifest) error {
		m.Fonts[0].Family = "backend mutation"
		record.Operations[0].Action = "caller mutation"
		record.Operations[0].Arguments[0].Value = "caller mutation"
		return nil
	}
	b.apply = func(o Operation) error {
		if len(b.applied) == 2 {
			b.capture.Data[9] = '2'
		}
		return nil
	}
	r, _ := NewRunner(RunnerConfig{}, b)
	got, err := r.Run(context.Background(), record)
	if err != nil || len(b.applied) != 2 || b.applied[0].Action != "save" || b.applied[0].Arguments[0].Value != "A😀\r\n" || string(got.Checkpoints[0].Data) != `{"count":1}` {
		t.Fatal(got, err, b.applied)
	}
	got.Checkpoints[0].Data[9] = '3'
	if string(b.capture.Data) != `{"count":2}` {
		t.Fatal("result data aliases backend")
	}
}
func TestRunnerCaptureBudgetsAndInvalidData(t *testing.T) {
	capture := Operation{Kind: Capture, Window: "main", WindowIncarnation: 1, Label: "sample"}
	b := protocol()
	point := Checkpoint{Sequence: 1, Window: "main", Label: "sample", Data: b.capture.Data, ImageSHA256: b.capture.ImageSHA256}
	encoded, _ := json.Marshal(point)
	for _, which := range []string{"exact", "checkpoint-byte", "payload-byte", "count", "json", "hash", "utf8"} {
		t.Run(which, func(t *testing.T) {
			b := protocol()
			config := RunnerConfig{MaxCaptureBytes: len(b.capture.Data), MaxCheckpointBytes: len(encoded)}
			record := recordWith(capture)
			switch which {
			case "checkpoint-byte":
				config.MaxCheckpointBytes--
			case "payload-byte":
				config.MaxCaptureBytes--
			case "count":
				config.MaxCaptures = 1
				record = recordWith(capture, capture)
			case "json":
				b.capture.Data = json.RawMessage(`{`)
			case "hash":
				b.capture.ImageSHA256 = "bad"
			case "utf8":
				b.capture.Data = json.RawMessage{0xff}
			}
			r, err := NewRunner(config, b)
			if err != nil {
				t.Fatal(err)
			}
			got, err := r.Run(context.Background(), record)
			if which == "exact" {
				if err != nil || len(got.Checkpoints) != 1 {
					t.Fatal(got, err)
				}
				return
			}
			want := ErrCapacity
			if which == "json" || which == "hash" || which == "utf8" {
				want = ErrBackend
			}
			if !errors.Is(err, want) {
				t.Fatal(which, got, err)
			}
			if which != "count" && len(got.Checkpoints) != 0 {
				t.Fatal("refused capture retained")
			}
		})
	}
}
func TestRunnerRefusesHiddenTimeAndImpossibleProgress(t *testing.T) {
	for _, which := range []string{"manifest-time", "operation-time", "step-time", "record-time", "overspend", "negative", "category", "no-progress", "invalid-readiness"} {
		t.Run(which, func(t *testing.T) {
			b := protocol()
			record := recordWith(action())
			want := ErrBackend
			switch which {
			case "manifest-time":
				b.check = func(Manifest) error { b.state.Now = b.state.Now.Add(time.Second); return nil }
				want = ErrTime
			case "operation-time":
				b.apply = func(Operation) error { b.state.Now = b.state.Now.Add(time.Second); return nil }
				want = ErrTime
			case "step-time":
				b.step = func(int) (Progress, error) {
					b.state.Now = b.state.Now.Add(time.Second)
					return Progress{Units: 1}, nil
				}
				want = ErrTime
			case "record-time":
				record.Operations[0].At = time.Second
				want = ErrTime
			case "overspend":
				b.step = func(n int) (Progress, error) { return Progress{Units: n + 1}, nil }
			case "negative":
				b.step = func(int) (Progress, error) { return Progress{Units: -1}, nil }
			case "category":
				b.step = func(int) (Progress, error) { return Progress{Units: 1, Frames: 2}, nil }
			case "no-progress":
				b.step = func(int) (Progress, error) { return Progress{}, nil }
				want = ErrNoProgress
			case "invalid-readiness":
				b.state.FutureTimers = 1
			}
			r, _ := NewRunner(RunnerConfig{}, b)
			_, err := r.Run(context.Background(), record)
			if !errors.Is(err, want) {
				t.Fatal(which, err)
			}
		})
	}
}
func TestRunnerCancellationAndReentrantAdmission(t *testing.T) {
	b := protocol()
	b.state.Ready = 2
	r, _ := NewRunner(RunnerConfig{}, b)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.RunUntilIdle(canceled); !errors.Is(err, context.Canceled) || b.calls != 0 {
		t.Fatal(err, b.calls)
	}
	if _, err := r.Run(nil, recordWith(action())); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	b.step = func(int) (Progress, error) {
		if _, err := r.RunUntilIdle(context.Background()); !errors.Is(err, ErrBusy) {
			t.Fatal(err)
		}
		b.state.Ready = 0
		return Progress{Units: 2, Events: 2}, nil
	}
	if got, err := r.RunUntilIdle(context.Background()); err != nil || !got.QuiescentNow {
		t.Fatal(got, err)
	}
	b.state.Ready = 2
	ctx, cancel := context.WithCancel(context.Background())
	b.step = func(int) (Progress, error) { cancel(); b.state.Ready--; return Progress{Units: 1}, ctx.Err() }
	got, err := r.RunUntilIdle(ctx)
	if !errors.Is(err, context.Canceled) || got.Progress.Units != 1 || got.Residual.Ready != 2 || got.QuiescentNow {
		t.Fatal(got, err)
	}
}
func TestRunnerConcurrentAndCopiedAdmission(t *testing.T) {
	b := protocol()
	b.state.Ready = 1
	r, _ := NewRunner(RunnerConfig{}, b)
	copyValue := reflect.New(reflect.TypeOf(r).Elem())
	copyValue.Elem().Set(reflect.ValueOf(r).Elem())
	copied := copyValue.Interface().(*Runner)
	if _, err := copied.RunUntilIdle(context.Background()); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	b.step = func(int) (Progress, error) {
		close(entered)
		<-release
		b.state.Ready = 0
		return Progress{Units: 1}, nil
	}
	go func() { _, err := r.RunUntilIdle(context.Background()); done <- err }()
	<-entered
	if _, err := r.Run(context.Background(), recordWith(action())); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type inspectionBackend struct {
	*protocolBackend
	reads   int
	inspect func(int) (Readiness, error)
}

func (b *inspectionBackend) Readiness(context.Context) (Readiness, error) {
	b.reads++
	if b.inspect != nil {
		return b.inspect(b.reads)
	}
	return b.state, nil
}

func TestRunnerPreservesLastInspectionWhenDrainCannotRead(t *testing.T) {
	for _, empty := range []bool{false, true} {
		for _, canceled := range []bool{false, true} {
			t.Run(fmt.Sprintf("empty=%v/canceled=%v", empty, canceled), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				b := &inspectionBackend{protocolBackend: protocol()}
				lastRead := 4 // initial, manifest-bound, before Apply, after Apply
				record := recordWith(action())
				if empty {
					lastRead = 2 // initial, manifest-bound
					record = recordWith()
					b.state.Ready = 3
				}
				failure := errors.New("inspection unavailable")
				wantErr := failure
				if canceled {
					wantErr = context.Canceled
				}
				var observed Readiness
				b.inspect = func(n int) (Readiness, error) {
					if n == lastRead {
						observed = b.state
						if canceled {
							cancel()
						}
					}
					if n > lastRead {
						return Readiness{}, failure
					}
					return b.state, nil
				}
				r, err := NewRunner(RunnerConfig{}, b)
				if err != nil {
					t.Fatal(err)
				}
				got, err := r.Run(ctx, record)
				if !errors.Is(err, wantErr) || got.Residual != observed || got.QuiescentNow || observed.Ready == 0 {
					t.Fatal(got, err, observed)
				}
				if !empty && got.Applied != 1 {
					t.Fatal("lost applied prefix", got)
				}
			})
		}
	}
}

func TestRunnerReadinessTimeMustSerialize(t *testing.T) {
	for _, year := range []int{0, 10000} {
		for _, deadline := range []bool{false, true} {
			b := protocol()
			bad := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			if deadline {
				b.state.NextDeadline = bad
				if bad.Before(b.state.Now) {
					b.state.DueTimers = 1
				} else {
					b.state.FutureTimers = 1
				}
			} else {
				b.state.Now = bad
			}
			r, _ := NewRunner(RunnerConfig{}, b)
			got, err := r.RunUntilIdle(context.Background())
			if !errors.Is(err, ErrBackend) || got.QuiescentNow || !got.Residual.Now.IsZero() {
				t.Fatal(got, err)
			}
		}
	}
	b := protocol()
	b.state.Now = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	r, _ := NewRunner(RunnerConfig{}, b)
	got, err := r.RunUntilIdle(context.Background())
	if err != nil || !got.QuiescentNow {
		t.Fatal(got, err)
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal("accepted readiness cannot encode", err)
	}
}

func TestRunnerCheckpointBudgetIsExplicitRetentionBound(t *testing.T) {
	b := protocol()
	op := Operation{Kind: Capture, Window: "main", WindowIncarnation: 1, Label: "sample"}
	point := Checkpoint{Sequence: 1, Window: "main", Label: "sample", Data: b.capture.Data, ImageSHA256: b.capture.ImageSHA256}
	encoded, err := json.Marshal(point)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRunner(RunnerConfig{MaxCaptureBytes: len(b.capture.Data), MaxCheckpointBytes: len(encoded)}, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.Run(context.Background(), recordWith(op))
	if err != nil || len(got.Checkpoints) != 1 {
		t.Fatal(got, err)
	}
	report, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if len(report) <= len(encoded) {
		t.Fatal("expected Result wrapper overhead")
	}
	checkpoint, err := json.Marshal(got.Checkpoints[0])
	if err != nil || len(checkpoint) != len(encoded) {
		t.Fatal(err, string(checkpoint))
	}
}

func TestRunnerCheckpointBudgetChargesStoredCompactData(t *testing.T) {
	b := protocol()
	b.capture.Data = json.RawMessage(strings.Repeat(" ", 65536) + "{ \"count\" : 1 }" + strings.Repeat("\n", 65536))
	op := Operation{Kind: Capture, Window: "main", WindowIncarnation: 1, Label: "sample"}
	point := Checkpoint{Sequence: 1, Window: "main", Label: "sample", Data: json.RawMessage(`{"count":1}`), ImageSHA256: b.capture.ImageSHA256}
	encoded, err := json.Marshal(point)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRunner(RunnerConfig{MaxCaptureBytes: len(b.capture.Data), MaxCheckpointBytes: 2 * len(b.capture.Data)}, b)
	if err != nil {
		t.Fatal(err)
	}
	// The raw admission limit is independent of compact retained bytes. Each
	// payload must fit before compaction; a record can retain only charged bytes.
	got, err := r.Run(context.Background(), recordWith(op, op))
	if err != nil || len(got.Checkpoints) != 2 {
		t.Fatal(got, err)
	}
	for _, p := range got.Checkpoints {
		if string(p.Data) != `{"count":1}` || cap(p.Data) != len(p.Data) {
			t.Fatal("whitespace retained", len(p.Data))
		}
		charged, err := json.Marshal(p)
		if err != nil || len(charged) != len(encoded) || len(p.Data) > len(charged) {
			t.Fatal(err, len(p.Data), len(charged))
		}
	}
	b.capture.Data[0] = '\t'
	if string(got.Checkpoints[0].Data) != `{"count":1}` {
		t.Fatal("compact data borrowed")
	}
}
