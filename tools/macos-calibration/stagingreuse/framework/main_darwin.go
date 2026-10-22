//go:build darwin && !offscreen

package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/color"
	"os"
	"time"
)

func good(d map[string]any) bool {
	return d["active"] == true && d["key"] == true && d["visible"] == true && d["pid"] == d["front_pid"] && d["losses"] == float64(0) && d["drawables"] == float64(2) && d["vsync"] == true && d["transactions"] == false && d["timeout"] == false
}
func run() (result error) {
	if len(os.Args) != 4 || (os.Args[1] != "pilot" && os.Args[1] != "timed") {
		return errors.New("pilot|timed MATCH.json off|on")
	}
	if os.Args[3] != "off" && os.Args[3] != "on" {
		return errors.New("off|on observer required")
	}
	if err := reuseConfigure(os.Args[3] == "on"); err != nil {
		return err
	}
	if err := probeConfigure(true); err != nil {
		return err
	}
	pilot := os.Args[1] == "pilot"
	core.DebugSettings.DisableWindowGeometrySaver = true
	a, p, e := newFixture()
	if e != nil {
		return e
	}
	defer a.Close()
	session, e := cogent.NewMeasuredDesktopSession(256)
	if e != nil {
		return e
	}
	metrics, e := cogent.NewPresentationMetrics(4096)
	if e != nil {
		return e
	}
	out := map[string]any{"kind": "framework", "release_qualified": false, "diagnostic_mode": "source-cost-only"}
	done := make(chan error, 1)
	worker := func() (failure error) {
		ctx, cancel := context.WithTimeout(a.Context(), 40*time.Second)
		defer cancel()
		defer func() {
			out["end"] = observe(2)
			out["before_close"] = metrics.Snapshot()
			out["measurements"] = session.Measurements()
		}()
		startupDeadline := time.Now().Add(5 * time.Second)
		for {
			s := metrics.Snapshot()
			ready := false
			for _, r := range s.Samples {
				if r.State == "presented" && r.Presented > 0 {
					ready = true
				}
			}
			if ready {
				break
			}
			if time.Now().After(startupDeadline) {
				return errors.New("initial natural presentation deadline")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond):
			}
		}
		if e := session.Frame(ctx, p.ref); e != nil {
			return e
		}
		snap, e := session.Snapshot(ctx, p.ref)
		if e != nil {
			return e
		}
		deadline := time.Now().Add(2 * time.Second)
		for {
			s := metrics.Snapshot()
			positive := false
			for _, r := range s.Samples {
				if r.State == "presented" && r.Presented > 0 {
					positive = true
				}
			}
			if positive {
				break
			}
			if time.Now().After(deadline) {
				return errors.New("startup positive presentation deadline")
			}
			time.Sleep(time.Millisecond)
		}
		settleDeadline := time.Now().Add(2 * time.Second)
		for {
			s := metrics.Snapshot()
			if s.Failed || s.Overflow {
				return errors.New("startup recorder failure")
			}
			if s.Pending == 0 {
				break
			}
			if time.Now().After(settleDeadline) {
				return errors.New("startup callback drain deadline")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond):
			}
		}
		begin := observe(1)
		out["begin"] = begin
		if !good(begin) {
			return errors.New("foreground/native configuration refused")
		}
		var rect [4]int
		for _, el := range snap.Elements {
			if el.DomainID == "patch" || el.ID == "patch" {
				rect = [4]int{el.Bounds.Min.X, el.Bounds.Min.Y, el.Bounds.Max.X, el.Bounds.Max.Y}
			}
		}
		if rect[2] <= rect[0] || rect[3] <= rect[1] {
			return fmt.Errorf("missing patch bounds: %+v", snap.Elements)
		}
		match := map[string]any{"width": begin["width"], "height": begin["height"], "scale": begin["scale"], "pixel_format": begin["pixel_format"], "rect": rect, "drawables": begin["drawables"], "vsync": begin["vsync"], "transactions": begin["transactions"], "timeout": begin["timeout"]}
		out["match"] = match
		if pilot {
			f, e := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if e != nil {
				return e
			}
			e = json.NewEncoder(f).Encode(match)
			return errors.Join(e, f.Close())
		}
		raw, e := os.ReadFile(os.Args[2])
		if e != nil {
			return e
		}
		var expected map[string]any
		if e = json.Unmarshal(raw, &expected); e != nil {
			return e
		}
		actual, _ := json.Marshal(match)
		want, _ := json.Marshal(expected)
		if string(actual) != string(want) {
			return errors.New("timed match drift")
		}
		base := time.Now().Add(120 * time.Millisecond)
		lateness := make([]int64, 0, 248)
		for i := 1; i <= 248; i++ {
			target := base.Add(time.Duration(i-1) * 120 * time.Millisecond)
			if delay := time.Until(target); delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}
			if !good(observe(0)) {
				return errors.New("foreground/configuration changed during cohort")
			}
			late := time.Since(target).Nanoseconds()
			lateness = append(lateness, late)
			out["lateness_ns"] = lateness
			if late > 20_000_000 {
				return errors.New("request deadline lateness over20ms")
			}
			s, e := session.Snapshot(ctx, p.ref)
			if e != nil {
				return e
			}
			sample, e := session.MeasureInput(ctx, p.ref, cogent.DesktopInput{Frame: s.Frame, Type: events.KeyChord, Code: key.CodeSpacebar, Rune: ' '}, func() (uint64, error) { return p.expect(uint64(i)) })
			if e != nil {
				return e
			}
			receiptOffsets, _ := out["receipt_lateness_ns"].([]int64)
			out["receipt_lateness_ns"] = append(receiptOffsets, sample.ReceiptStartedAt.Sub(target).Nanoseconds())
			for {
				snapshot := metrics.Snapshot()
				found := false
				if snapshot.Failed || snapshot.Overflow {
					return errors.New("native presentation recorder refused")
				}
				for _, r := range snapshot.Samples {
					if r.Scope == sample.Scope && r.Tag == sample.Frame.Tag && r.Generation == sample.Frame.Generation && r.Armed {
						if r.State == "skipped" || r.State == "retired" {
							return errors.New("lost armed presentation")
						}
						if r.State == "presented" && r.Presented > 0 {
							found = true
						}
					}
				}
				if found {
					break
				}
				if time.Now().After(target.Add(120 * time.Millisecond)) {
					return errors.New("response missed next request deadline")
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Millisecond):
				}
			}
			if time.Now().After(target.Add(120 * time.Millisecond)) {
				return errors.New("response late after presentation")
			}
		}
		return nil
	}
	go func() { err := worker(); done <- err; a.Close() }()
	native := a.Run(context.Background(), cogent.Driver{DesktopSession: session, PresentationMetrics: metrics, Commands: p.commands, Theme: ui.Theme{Default: ui.Style{Background: ui.Some(color.NRGBA{A: 255}), Foreground: ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 15})}}, OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
	select {
	case e = <-done:
	case <-time.After(2 * time.Second):
		e = errors.New("worker did not retire")
	}
	out["after_close"] = metrics.Snapshot()
	trace, traceErr := probeSnapshot()
	out["native_upload_trace"] = trace
	out["observer_mode"] = "on"
	out["reuse_mode"] = os.Args[3]
	pool, poolErr := reuseSnapshot()
	out["reuse_trace"] = pool
	e = errors.Join(e, poolErr)
	e = errors.Join(e, traceErr)
	out["error"] = fmt.Sprint(errors.Join(native, e))
	out["runtime_returned"] = true
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		return err
	}
	return errors.Join(native, e)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
