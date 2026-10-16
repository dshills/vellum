//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/tasks"
)

type soakConfig struct{ Duration, Quiet, Settle, Interval time.Duration }

func (s soakConfig) validate(c config) error {
	if !c.composed || c.cycles != 1 || c.every != 1 || s.Duration < 40*time.Second || s.Duration > 12*time.Hour || s.Quiet < time.Second || s.Settle < time.Second || s.Quiet > s.Duration/4 || s.Settle > s.Duration/4 || s.Quiet+s.Settle >= s.Duration/4 || s.Interval < s.Quiet+s.Settle || s.Interval > s.Duration/4 || c.timeout < s.Duration+30*time.Second || c.timeout > s.Duration+5*time.Minute {
		return errors.New("soak requires composed, cycles=sample-every=1, duration40s..12h, quiet/settle>=1s, quiet+settle<duration/4, quiet+settle<=interval<=duration/4 and timeout=duration+30s..5m")
	}
	return nil
}
func soakSleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
func soakCPU() (float64, error) {
	var r syscall.Rusage
	if e := syscall.Getrusage(syscall.RUSAGE_SELF, &r); e != nil {
		return 0, e
	}
	return float64(r.Utime.Sec+r.Stime.Sec) + float64(r.Utime.Usec+r.Stime.Usec)/1e6, nil
}
func soakWindow(ctx context.Context, title string) (cogent.NativeServiceWindowCheck, error) {
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	s, e := cogent.CheckNativeServices(bounded, cogent.NativeCheckSnapshot, 0)
	if e != nil {
		return cogent.NativeServiceWindowCheck{}, e
	}
	var found cogent.NativeServiceWindowCheck
	n := 0
	for _, w := range s.Windows {
		if w.Title == title {
			found = w
			n++
		}
	}
	if !s.MainThread || n != 1 {
		return found, fmt.Errorf("soak Cocoa window matches=%d main=%t", n, s.MainThread)
	}
	return found, nil
}
func soakState(s cogent.NativeServiceWindowCheck, phase string) bool {
	if phase != "visible" && phase != "occluded" && phase != "minimized" && phase != "background" {
		return false
	}
	for _, coordinate := range s.Frame {
		if math.IsNaN(coordinate) || math.IsInf(coordinate, 0) {
			return false
		}
	}
	if s.Key || s.Frame[2] <= 0 || s.Frame[3] <= 0 {
		return false
	}
	if phase == "minimized" {
		return s.Miniaturized && !s.OcclusionVisible
	}
	if !s.Visible || s.Miniaturized {
		return false
	}
	return s.OcclusionVisible == (phase != "occluded")
}

// Both DTOs come from one bounded main-thread snapshot. Requested client geometry
// is separate from observed Cocoa outer frames. Neither containment nor ordering
// substitutes for the target's strict effective occlusion predicate.
func soakPair(ctx context.Context, target, cover string) (cogent.NativeServiceWindowCheck, cogent.NativeServiceWindowCheck, error) {
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	s, err := cogent.CheckNativeServices(bounded, cogent.NativeCheckSnapshot, 0)
	if err != nil {
		return cogent.NativeServiceWindowCheck{}, cogent.NativeServiceWindowCheck{}, err
	}
	return soakPairState(s, target, cover)
}

func soakPairState(s cogent.NativeServiceCheckState, target, cover string) (a, b cogent.NativeServiceWindowCheck, err error) {
	na, nb := 0, 0
	for _, w := range s.Windows {
		if w.Title == target {
			a = w
			na++
		}
		if w.Title == cover {
			b = w
			nb++
		}
	}
	if !s.MainThread || target == cover || na != 1 || nb != 1 {
		err = fmt.Errorf("soak pair target=%d cover=%d main=%t", na, nb, s.MainThread)
	}
	return
}

func soakContains(outer, inner [4]float64) bool {
	for _, frame := range [][4]float64{outer, inner} {
		for _, v := range frame {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return false
			}
		}
		if frame[2] <= 0 || frame[3] <= 0 || math.IsInf(frame[0]+frame[2], 0) || math.IsInf(frame[1]+frame[3], 0) {
			return false
		}
	}
	return outer[0] <= inner[0] && outer[1] <= inner[1] && outer[0]+outer[2] >= inner[0]+inner[2] && outer[1]+outer[3] >= inner[1]+inner[3]
}

func soakErrorText(text string) string {
	return strings.ToValidUTF8(text[:min(len(text), 512)], "")
}

// Quiet samples do not post owner callbacks, collect heaps or inspect Cocoa
// between the CPU/paint counter endpoints. Surrounding observations and the
// allocating background phase remain separate, explicitly measured activity.
func checkSoak(ctx context.Context, a *app.Application, survivor *cogent.GPUCanvas, model *canvas.Scene, shown *atomic.Bool, c config, p *progress) (result error) {
	cover, e := readback(ctx, a, survivor, model, shown)
	if e != nil {
		return e
	}
	gpu, m, ready, w, ref, e := openMode(a, "soak", "Vellum composed idle soak", true)
	if e != nil {
		return e
	}
	native, e := readback(ctx, a, gpu, m, ready)
	if e != nil {
		return e
	}
	win, ok := native.(*desktop.Window)
	if !ok || !win.IsOnDemand() {
		return errors.New("soak requires an on-demand desktop window")
	}
	if e = exerciseComposed(ctx, a, "soak", w, 0); e != nil {
		return e
	}
	emit := func(v any) error { return json.NewEncoder(os.Stdout).Encode(v) }
	manager, e := tasks.NewWithConfig(tasks.Config{Owner: a, Window: "soak", Limit: 1})
	if e != nil {
		return e
	}
	committed := 0
	var workerFailure error
	var worker *tasks.Task
	defer func() {
		if worker == nil {
			return
		}
		worker.Cancel()
		joined, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		for _, done := range []<-chan struct{}{worker.Done(), worker.OwnerDone()} {
			select {
			case <-done:
			case <-joined.Done():
				result = errors.Join(result, errors.New("soak worker did not join cancellation and owner retirement"))
				return
			}
		}
	}()
	soakStarted := time.Now()
	for _, phase := range []string{"visible", "occluded", "minimized", "background"} {
		native.Raise()
		if phase == "occluded" {
			// Equal decorated frames need not fully occlude their bounding boxes.
			// Keep the composed target size and surround it with a larger cover.
			cover.SetWinSize(soakWindowSize(win).Add(image.Pt(64, 64)))
			cover.SetPos(native.Position(nil).Sub(image.Pt(32, 32)), nil)
		} else {
			cover.SetWinSize(image.Pt(320, 160))
			cover.SetPos(image.Pt(0, 0), nil)
		}
		cover.Raise()
		if phase == "minimized" {
			native.Minimize()
		}
		var observedTarget, observedCover cogent.NativeServiceWindowCheck
		observed := false
		var observedUTC, queryError string
		e = wait(ctx, func(inner context.Context) (bool, error) {
			targetState, coverState, err := soakPair(inner, "Vellum composed idle soak", "Vellum cycle survivor")
			if err != nil {
				queryError = soakErrorText(err.Error())
				return false, err
			}
			observedTarget, observedCover, observed = targetState, coverState, true
			observedUTC, queryError = time.Now().UTC().Format(time.RFC3339Nano), ""
			return soakState(targetState, phase) && (phase != "occluded" || soakCoverState(targetState, coverState)), nil
		})
		transitionErr := emit(map[string]any{"event": "vellum.soak.transition", "phase": phase, "observed": observed, "observed_utc": observedUTC, "last_query_error": queryError, "target": observedTarget, "cover": observedCover, "cover_contains_target": observed && soakContains(observedCover.Frame, observedTarget.Frame), "state_accepted": e == nil})
		if e != nil || transitionErr != nil {
			return fmt.Errorf("soak %s state: %w", phase, errors.Join(e, transitionErr))
		}
		phaseStarted := time.Now()
		phaseCPU0, e := soakCPU()
		if e != nil {
			return e
		}
		deadline := phaseStarted.Add(c.soak.Duration / 4)
		samples := 0
		if e = emit(map[string]any{"event": "vellum.soak.phase", "phase": phase, "requested_seconds": (c.soak.Duration / 4).Seconds()}); e != nil {
			return e
		}
		for time.Now().Add(c.soak.Settle + c.soak.Quiet).Before(deadline) {
			started := time.Now()
			if phase == "background" {
				target := committed + 1
				operation, cancel := context.WithTimeout(ctx, 5*time.Second)
				reply, err := ownerProbe(operation, a, func() ownerReply {
					var e error
					worker, e = tasks.Start(manager, operation, func(ctx context.Context, _ func(float64, string)) (uint64, error) {
						return soakWork(ctx)
					}, func(v uint64, err error) {
						workerFailure = soakResult(v, err)
						if workerFailure == nil {
							committed++
							workerFailure = a.Invalidate("soak")
						}
					})
					return ownerReply{err: e}
				})
				if err != nil || reply.err != nil {
					cancel()
					return errors.Join(err, reply.err)
				}
				e = waitWithin(operation, 5*time.Second, func(inner context.Context) (bool, error) {
					v := false
					r, e := ownerProbe(inner, a, func() ownerReply {
						v = committed == target && manager.Pending().Active == 0
						return ownerReply{err: errors.Join(workerFailure, worker.Snapshot().Err)}
					})
					return v, errors.Join(e, r.err)
				})
				cancel()
				if e != nil {
					return e
				}
			}
			if e = soakSleep(ctx, c.soak.Settle); e != nil {
				return e
			}
			state0, cover0, e := soakPair(ctx, "Vellum composed idle soak", "Vellum cycle survivor")
			endpoint0 := soakEndpoint{target: state0, cover: cover0, err: e, at: time.Now()}
			if e != nil {
				return checkSoakEndpoints(phase, samples+1, endpoint0, soakEndpoint{}, time.Time{}, time.Time{}, emit)
			}
			activityStartRead := soakActivityReadWindow{StartUTC: time.Now()}
			activity0 := cogent.NativeActivitySnapshot()
			activityStartRead.EndUTC = time.Now()
			wakes := win.Paint.Wakes.Load()
			cpu0, e := soakCPU()
			if e != nil {
				return e
			}
			t0 := time.Now()
			if e = soakSleep(ctx, c.soak.Quiet); e != nil {
				return e
			}
			t1 := time.Now()
			elapsed := t1.Sub(t0).Seconds()
			wallElapsed := float64(t1.UnixNano()-t0.UnixNano()) / 1e9
			cpu1, e := soakCPU()
			if e != nil {
				return e
			}
			activityEndRead := soakActivityReadWindow{StartUTC: time.Now()}
			activity1 := cogent.NativeActivitySnapshot()
			activityEndRead.EndUTC = time.Now()
			intervalEligible, cpuPercent, e := soakInterval(c.soak.Quiet, elapsed, wallElapsed, cpu1-cpu0)
			if e != nil {
				return e
			}
			delta := win.Paint.Wakes.Load() - wakes
			state1, cover1, e := soakPair(ctx, "Vellum composed idle soak", "Vellum cycle survivor")
			endpoint1 := soakEndpoint{target: state1, cover: cover1, err: e, at: time.Now()}
			if err := checkSoakEndpoints(phase, samples+1, endpoint0, endpoint1, t0, t1, emit); err != nil {
				return err
			}
			activity, activityErr := soakActivityDelta(os.Getpid(), activity0, activity1, soakActivityBounds{StartRead: activityStartRead, EndRead: activityEndRead, QuietStart: t0, QuietEnd: t1})
			if activityErr != nil {
				return errors.Join(activityErr, emit(map[string]any{"event": "vellum.soak.activity-failure", "phase": phase, "sample": samples + 1, "start_utc": t0.UTC(), "end_utc": t1.UTC(), "framework_native_activity": activity}))
			}
			memory, e := sample(ctx, 0, soakStarted)
			if e != nil {
				return e
			}
			if e = memoryGuard(memory); e != nil {
				return e
			}
			if e = p.record(memory); e != nil {
				return e
			}
			samples++
			observation := map[string]any{"event": "vellum.soak.sample", "phase": phase, "sample": samples, "elapsed_seconds": elapsed, "wall_elapsed_seconds": wallElapsed, "start_utc": t0.UTC(), "end_utc": t1.UTC(), "process_cpu_seconds": cpu1 - cpu0, "one_cpu_percent": cpuPercent, "primary_paint_wakes": delta, "start_state": state0, "end_state": state1, "forced_gc_memory": memory, "background_commits": committed, "framework_native_activity": activity, "app_originated_wakeups_available": false, "present_count_available": false, "cpu_60s_interval": intervalEligible, "full_p11_qualified": false}
			if e = emit(observation); e != nil {
				return e
			}
			if rest := c.soak.Interval - time.Since(started); rest > 0 {
				if e = soakSleep(ctx, min(rest, time.Until(deadline))); e != nil {
					return e
				}
			}
		}
		if samples == 0 {
			return fmt.Errorf("soak %s produced no quiet sample", phase)
		}
		if e = soakSleep(ctx, max(time.Until(deadline), 0)); e != nil {
			return e
		}
		phaseCPU1, e := soakCPU()
		if e != nil {
			return e
		}
		if e = emit(map[string]any{"event": "vellum.soak.phase-end", "phase": phase, "elapsed_seconds": time.Since(phaseStarted).Seconds(), "process_cpu_seconds": phaseCPU1 - phaseCPU0, "samples": samples, "background_commits": committed}); e != nil {
			return e
		}
	}
	closing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r, e := ownerProbe(closing, a, func() ownerReply { return ownerReply{err: a.CloseWindow(ref)} })
	if e != nil || r.err != nil {
		return errors.Join(e, r.err)
	}
	p.advance()
	return nil
}

func soakCoverState(target, cover cogent.NativeServiceWindowCheck) bool {
	return cover.Visible && !cover.Miniaturized && cover.Opaque && cover.Alpha == 1 &&
		cover.Level == target.Level && cover.OrderedIndex >= 0 && target.OrderedIndex > cover.OrderedIndex &&
		soakContains(cover.Frame, target.Frame)
}

func soakResult(value uint64, err error) error {
	if err != nil {
		return err
	}
	if value != 133693440 {
		return fmt.Errorf("soak background checksum mismatch: %d", value)
	}
	return nil
}

// Keep acceptance eligibility separate from measured CPU. A short preflight or
// suspended/clock-shifted interval can exercise the fixture without qualifying
// the original sixty-second idle budget. The caller still records both values.
func soakInterval(requested time.Duration, elapsed, wallElapsed, cpuSeconds float64) (bool, float64, error) {
	for _, value := range []float64{elapsed, wallElapsed, cpuSeconds} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false, 0, errors.New("soak CPU interval is nonfinite")
		}
	}
	if elapsed <= 0 || wallElapsed <= 0 || cpuSeconds < 0 {
		return false, 0, errors.New("soak CPU interval is nonpositive or reversed")
	}
	percent := 100 * cpuSeconds / elapsed
	if math.IsNaN(percent) || math.IsInf(percent, 0) {
		return false, 0, errors.New("soak CPU percentage overflow")
	}
	eligible := requested >= time.Minute && elapsed >= 60 && elapsed <= (requested+max(time.Second, requested/10)).Seconds() && math.Abs(wallElapsed-elapsed) <= 1
	return eligible, percent, nil
}

func soakWork(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	data := make([]byte, 1<<20)
	var sum uint64
	for i := range data {
		if i%4096 == 0 && ctx.Err() != nil {
			return 0, ctx.Err()
		}
		data[i] = byte(i*31 + 7)
	}
	for i, v := range data {
		if i%4096 == 0 && ctx.Err() != nil {
			return 0, ctx.Err()
		}
		sum += uint64(v)
	}
	runtime.KeepAlive(data)
	return sum, nil
}

// Desktop geometry callbacks update the inherited WinSize field under Mu. Copy
// only while holding that mutex; dispatching a native operation under it could
// deadlock the main-thread callback that needs the same window ownership.
func soakWindowSize(window *desktop.Window) image.Point {
	window.Mu.Lock()
	size := window.WinSize()
	window.Mu.Unlock()
	return size
}
