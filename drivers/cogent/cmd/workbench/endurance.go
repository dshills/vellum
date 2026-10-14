package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"sync"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/system"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/virtual"
)

type enduranceConfig struct {
	duration time.Duration
	output   string
}

func (c enduranceConfig) validate() error {
	if c.duration == 0 && c.output == "" {
		return nil
	}
	if c.duration < 10*time.Second || c.duration > 12*time.Hour || !filepath.IsAbs(c.output) || filepath.Clean(c.output) != c.output {
		return errors.New("endurance requires duration10s..12h and a clean absolute new output directory")
	}
	return nil
}
func disableEnduranceGeometrySaver() { core.DebugSettings.DisableWindowGeometrySaver = true }

// PostWait admits work; the reply proves owner execution actually finished.
func enduranceOwner(ctx context.Context, a *app.Application, fn func() error) error {
	reply := make(chan error, 1)
	var mu sync.Mutex
	phase := 0 // queued, executing, revoked, finished
	if err := a.PostWait(ctx, func() {
		mu.Lock()
		if phase == 2 || ctx.Err() != nil {
			phase = 2
			mu.Unlock()
			reply <- ctx.Err()
			return
		}
		phase = 1
		mu.Unlock()
		err := fn()
		mu.Lock()
		phase = 3
		mu.Unlock()
		reply <- err
	}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		mu.Lock()
		if phase == 0 {
			phase = 2
			mu.Unlock()
			return ctx.Err()
		}
		active := phase == 1 || phase == 3
		mu.Unlock()
		// Join running callbacks before closing their resources. The outer
		// native-process watchdog bounds a broken owner operation.
		if active {
			return errors.Join(ctx.Err(), <-reply)
		}
		return ctx.Err()
	}
}

type cappedTrace struct {
	mu        sync.Mutex
	file      *os.File
	remaining int64
	err       error
}

func (w *cappedTrace) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return 0, w.err
	}
	if int64(len(b)) > w.remaining {
		w.err = errors.New("diagnostic output exceeds configured byte limit")
		return 0, w.err
	}
	n, err := w.file.Write(b)
	w.remaining -= int64(n)
	w.err = err
	return n, err
}
func (w *cappedTrace) failure() error { w.mu.Lock(); defer w.mu.Unlock(); return w.err }

type enduranceState struct {
	Steps, NativeEdits, Resizes, WorkersStarted, WorkersCommitted, WorkersStale int
	Worker                                                                      *tasks.Task
	Failure                                                                     error
}

func runEndurance(ctx context.Context, a *app.Application, c enduranceConfig, commands *command.Registry, parts *partSource, table *virtual.Table, board *canvas.Scene, gpu *cogent.GPUCanvas, manager *tasks.Manager, selected func() int) (result error) {
	if err := c.validate(); err != nil {
		return err
	}
	if err := os.Mkdir(c.output, 0700); err != nil {
		return err
	}
	out, err := os.OpenFile(filepath.Join(c.output, "observations.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, out.Close()) }()
	emit := func(v any) error { return json.NewEncoder(out).Encode(v) }
	probe, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var native system.Window
	var rowCount int
	for {
		err = enduranceOwner(probe, a, func() error {
			_, e := gpu.Snapshot()
			if e != nil {
				return e
			}
			native = gpu.NativeWindow()
			rowCount = parts.Len()
			if native == nil {
				return cogent.ErrFrameNotReady
			}
			return nil
		})
		if err == nil {
			break
		}
		if !errors.Is(err, cogent.ErrFrameNotReady) {
			return err
		}
		select {
		case <-probe.Done():
			return probe.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	pixels, managerUnits, geometryErr := enduranceGeometry(native)
	if geometryErr != nil {
		return geometryErr
	}
	if err = emit(map[string]any{"pid": os.Getpid(), "event": "vellum.endurance.ready", "duration_seconds": c.duration.Seconds(), "rows": rowCount, "go_version": runtime.Version(), "window_pixels": pixels, "window_manager_units": managerUnits, "scope": "profiled composed workbench; actual Cocoa value requests, owner domain/history/table/canvas activity, native resize and bounded allocating workers; no physicalinput, frame/presentation metrics or fullP11 qualification"}); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(c.output, "execution.trace"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	writer := &cappedTrace{file: file, remaining: 256 << 20}
	if err = trace.Start(writer); err != nil {
		file.Close()
		return err
	}
	tracing := true
	defer func() {
		if tracing {
			trace.Stop()
		}
		result = errors.Join(result, writer.failure(), file.Close())
	}()
	state := &enduranceState{}
	workContext, stopWorkers := context.WithCancel(ctx)
	var workerMu sync.Mutex
	var admittedWorker *tasks.Task
	defer func() {
		stopWorkers()
		workerMu.Lock()
		worker := admittedWorker
		workerMu.Unlock()
		if worker != nil {
			joined, stop := context.WithTimeout(context.Background(), 10*time.Second)
			defer stop()
			for _, done := range []<-chan struct{}{worker.Done(), worker.OwnerDone()} {
				select {
				case <-done:
				case <-joined.Done():
					result = errors.Join(result, errors.New("endurance worker cleanup did not join"))
					return
				}
			}
		}
	}()
	started := time.Now()
	deadline := started.Add(c.duration)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
		stepStart := time.Now()
		step := state.Steps + 1
		operation, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = enduranceOwner(operation, a, func() error {
			if state.Failure != nil {
				return state.Failure
			}
			if err := commands.Invoke(operation, "project.increment"); err != nil {
				return err
			}
			original := parts.parts.at(selected())
			for _, id := range []string{"component.rotate", "component.raise-value", "edit.undo", "edit.undo"} {
				if err := commands.Invoke(operation, id); err != nil {
					return err
				}
			}
			if parts.parts.at(selected()) != original {
				return errors.New("endurance document replay mismatch")
			}
			if err := table.SetViewport((step*17)%(parts.Len()-table.VisibleCount()), table.VisibleCount()); err != nil {
				return err
			}
			if step%2 == 0 {
				board.Fit()
			} else {
				if err := board.View.ZoomAt(canvas.Point{X: board.View.Width / 2, Y: board.View.Height / 2}, 1.1); err != nil {
					return err
				}
			}
			if step%12 == 0 {
				native.SetWinSize(image.Pt(1100+(step/12%2)*180, 780+(step/12%2)*120))
				state.Resizes++
			}
			if state.Worker == nil || (state.Worker.Snapshot().WorkerDone && state.Worker.Snapshot().OwnerDone) {
				revision := parts.Revision()
				token := step
				worker, e := tasks.Start(manager, workContext, func(ctx context.Context, _ func(float64, string)) (uint64, error) {
					var sum uint64
					for i := 0; i < 16; i++ {
						if err := ctx.Err(); err != nil {
							return 0, err
						}
						b := enduranceChunk()
						for j := range b {
							b[j] = byte(j + token + i)
						}
						for _, v := range b {
							sum += uint64(v)
						}
						runtime.KeepAlive(b)
					}
					select {
					case <-ctx.Done():
						return 0, ctx.Err()
					case <-time.After(enduranceWorkerDelay(token)):
					}
					return sum, nil
				}, func(sum uint64, e error) {
					if e != nil {
						if !errors.Is(e, context.Canceled) {
							state.Failure = e
						}
						return
					}
					if sum != 16*64*1024*255/2 {
						state.Failure = errors.New("endurance worker checksum mismatch")
						return
					}
					if revision != parts.Revision() {
						state.WorkersStale++
					} else {
						state.WorkersCommitted++
					}
				})
				if e != nil {
					return e
				}
				state.Worker = worker
				workerMu.Lock()
				admittedWorker = worker
				workerMu.Unlock()
				state.WorkersStarted++
			}
			state.Steps = step
			return a.Invalidate("main")
		})
		if err == nil && step%4 == 0 {
			value := uint32(1000 + step)
			// Reinspect and retry requests revoked by a newer semantic generation.
			// Queue admission alone never counts as an accepted domain edit.
			for {
				visible := false
				err = enduranceOwner(operation, a, func() error { visible = parts.parts.at(selected()).Value == value; return nil })
				if err != nil || visible {
					break
				}
				inspection, e := cogent.InspectAccessibility(operation, "main")
				if e != nil {
					err = e
					break
				}
				target := ""
				for _, element := range inspection.Elements {
					if element.DomainID == "inspector.value" && element.Role == "AXTextField" {
						target = element.ID
						break
					}
				}
				if target == "" {
					err = errors.New("endurance inspector AX text target missing")
					break
				}
				_, err = cogent.PerformAccessibilityAction(operation, "main", target, "set", strconv.FormatUint(uint64(value), 10))
				if err != nil {
					break
				}
				select {
				case <-operation.Done():
					err = operation.Err()
				case <-time.After(25 * time.Millisecond):
				}
				if err != nil {
					break
				}
			}
			if err == nil {
				err = enduranceOwner(operation, a, func() error { state.NativeEdits++; return nil })
			}
		}
		if err == nil && step%20 == 0 {
			var sample map[string]any
			pixels, managerUnits, geometryErr := enduranceGeometry(native)
			if geometryErr != nil {
				cancel()
				return geometryErr
			}
			err = enduranceOwner(operation, a, func() error {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				sample = map[string]any{"event": "vellum.endurance.sample", "step": state.Steps, "elapsed_seconds": time.Since(started).Seconds(), "owner_roundtrip_ns": time.Since(stepStart).Nanoseconds(), "heap_alloc": mem.HeapAlloc, "heap_inuse": mem.HeapInuse, "total_alloc": mem.TotalAlloc, "go_sys": mem.Sys, "gc_cycles": mem.NumGC, "goroutines": runtime.NumGoroutine(), "pending": a.Pending(), "workers": manager.Pending().Active, "window_pixels": pixels, "window_manager_units": managerUnits, "domain_value": parts.parts.at(selected()).Value}
				return nil
			})
			if err == nil {
				err = emit(sample)
			}
		}
		cancel()
		if err != nil {
			return fmt.Errorf("endurance step%d: %w", step, err)
		}
	}
	// Worker termination and owner completion precede native shutdown.
	finish, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var worker *tasks.Task
	if err = enduranceOwner(finish, a, func() error { worker = state.Worker; return nil }); err != nil {
		return err
	}
	if worker != nil {
		select {
		case <-worker.Done():
		case <-finish.Done():
			return finish.Err()
		}
		select {
		case <-worker.OwnerDone():
		case <-finish.Done():
			return finish.Err()
		}
	}
	if err = enduranceOwner(finish, a, func() error {
		if state.Failure != nil {
			return state.Failure
		}
		if manager.Pending().Active != 0 || state.WorkersStarted != state.WorkersCommitted+state.WorkersStale || state.WorkersCommitted == 0 || state.WorkersStale == 0 {
			return errors.New("endurance worker completion lost")
		}
		return emit(map[string]any{"event": "vellum.endurance.result", "result": "PASS_COMPOSED_ACTIVITY", "elapsed_seconds": time.Since(started).Seconds(), "steps": state.Steps, "native_edits": state.NativeEdits, "resizes": state.Resizes, "workers_started": state.WorkersStarted, "workers_committed": state.WorkersCommitted, "workers_stale_discarded": state.WorkersStale, "pending": a.Pending(), "workers": manager.Pending().Active, "full_p11_qualified": false})
	}); err != nil {
		return err
	}
	trace.Stop()
	tracing = false
	if err = writer.failure(); err != nil {
		return err
	}
	runtime.GC()
	heap, err := os.OpenFile(filepath.Join(c.output, "heap.pprof"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	heapWriter := &cappedTrace{file: heap, remaining: 8 << 20}
	return errors.Join(pprof.WriteHeapProfile(heapWriter), heapWriter.failure(), heap.Close())
}

//go:noinline
func enduranceChunk() []byte { return make([]byte, 64<<10) }

func enduranceWorkerDelay(token int) time.Duration {
	if token%4 < 2 {
		return 100 * time.Millisecond
	}
	return 300 * time.Millisecond
}

// Desktop resize workers update geometry under this same window mutex.
func enduranceGeometry(native system.Window) (pixels, managerUnits image.Point, err error) {
	if !native.Lock() {
		return pixels, managerUnits, app.ErrWindowClosed
	}
	defer native.Unlock()
	return native.Size(), native.WinSize(), nil
}

// The native-loop cause precedes any secondary controller cancellation/error.
func enduranceRunResult(runErr error, done <-chan error, limit time.Duration) error {
	select {
	case workerErr := <-done:
		return errors.Join(runErr, workerErr)
	case <-time.After(limit):
		return errors.Join(runErr, errors.New("endurance worker failed to join"))
	}
}
