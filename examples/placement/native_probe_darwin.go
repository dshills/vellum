//go:build darwin && cgo && !offscreen && vellum_placement_probe

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/system"
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

// A dedicated executable follows the ordinary desktop initialization and original
// locked OS thread. A go-test binary would select upstream's offscreen driver.
func main() { os.Exit(nativePlacementProbeMain()) }
func nativePlacementProbeMain() (code int) {
	defer func() {
		if v := recover(); v != nil {
			fmt.Fprintln(os.Stderr, "native probe panic:", v)
			code = 1
		}
	}()
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "native probe accepts no arguments")
		return 2
	}
	if err := nativePlacementPlatform(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := runNativePlacementDesktopProbe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
func nativePlacementPlatform() error {
	if system.TheApp != desktop.TheApp || core.TheApp.App != system.TheApp || system.TheApp.Platform() != system.MacOS {
		return fmt.Errorf("native placement probe requires consistent macOS desktop app identities")
	}
	return nil
}

// This probe does not generate pointer/key events or invoke placement commands.
// An external desktop operator supplies those inputs to the ordinary demo view.
// Its diagnostic buttons model source/target removal during a real active drag.
type nativePlacementProbe struct {
	m                                          *model
	a                                          *app.Application
	ref                                        app.WindowRef
	sourceActivations, canvasChanges, removals int
	hideSource, hideTarget                     bool
	removeOnDrag                               string
}

func (p *nativePlacementProbe) view() ui.Node {
	root := view(p.a, p.ref, p.m)
	var wrap func(ui.Node) ui.Node
	wrap = func(n ui.Node) ui.Node {
		if n.PlacementSource != nil {
			old := n.Activate
			n.Activate = func() { p.sourceActivations++; old() }
		}
		if n.ID == "palette" && p.hideSource {
			n.Children = nil
		}
		if n.ID == "drawing" {
			if p.hideTarget {
				return ui.Node{ID: "probe-no-canvas", Kind: ui.Label, Text: "Canvas removed", Grow: true}
			}
			old := n.Change
			n.Change = func(e ui.ChangeEvent) { p.canvasChanges++; old(e) }
		}
		for i := range n.Children {
			n.Children[i] = wrap(n.Children[i])
		}
		return n
	}
	root = wrap(root)
	bar := ui.Node{ID: "probe-controls", Kind: ui.Toolbar}
	for _, mode := range []string{"source", "target", "window"} {
		bar.Children = append(bar.Children, ui.Node{ID: "probe-arm-" + mode, Kind: ui.Button, Text: "Remove " + mode + " on next drag", Activate: func() { p.removeOnDrag = mode; _ = p.a.Invalidate("placement") }})
	}
	bar.Children = append(bar.Children, ui.Node{ID: "probe-restore", Kind: ui.Button, Text: "Restore bindings", Activate: func() { p.hideSource, p.hideTarget, p.removeOnDrag = false, false, ""; _ = p.a.Invalidate("placement") }})
	root.Children = append(root.Children, bar)
	return root
}

type nativePlacementState struct {
	Objects                                               []canvas.Rect
	View                                                  canvas.Viewport
	Selected, Status                                      string
	History                                               history.State
	Gesture                                               placement.State
	SourceActivations, CanvasChanges, Removals            int
	HiddenSource, HiddenTarget, Snap, Locked, RejectNotes bool
	RemoveOnDrag                                          string
	Next                                                  uint64
	Reserved                                              int
}

func (p *nativePlacementProbe) state() nativePlacementState {
	return nativePlacementState{Objects: slices.Clone(p.m.scene.Objects), View: p.m.scene.View, Selected: p.m.scene.Selected, Status: p.m.status, History: p.m.journal.Snapshot(), Gesture: p.m.palette.State(), SourceActivations: p.sourceActivations, CanvasChanges: p.canvasChanges, Removals: p.removals, HiddenSource: p.hideSource, HiddenTarget: p.hideTarget, Snap: p.m.snap, Locked: p.m.locked, RejectNotes: p.m.rejectNotes, RemoveOnDrag: p.removeOnDrag, Next: p.m.next, Reserved: p.a.Pending().Reserved}
}

const nativeProbeByteLimit = 8 << 20

type nativeProbeWriter struct {
	root  string
	log   *os.File
	bytes int
}

func newNativeProbeWriter(root string) (*nativeProbeWriter, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, fmt.Errorf("probe requires a clean absolute directory")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("probe requires a private directory")
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil || resolved != root {
		return nil, fmt.Errorf("probe directory must not use symlinks")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	if len(entries) != 0 {
		return nil, fmt.Errorf("probe directory must be empty")
	}
	log, err := os.OpenFile(filepath.Join(root, "events.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &nativeProbeWriter{root: root, log: log}, nil
}
func (w *nativeProbeWriter) write(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if len(data) > 256<<10 || w.bytes+len(data) > nativeProbeByteLimit {
		return fmt.Errorf("probe output bound")
	}
	n, err := w.log.Write(data)
	w.bytes += n
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("short probe log write")
	}
	tmp := filepath.Join(w.root, "latest.tmp")
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	n, writeErr := f.Write(data)
	closeErr := f.Close()
	if err = errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if n != len(data) {
		_ = os.Remove(tmp)
		return fmt.Errorf("short probe snapshot write")
	}
	return os.Rename(tmp, filepath.Join(w.root, "latest.json"))
}

// done closes only after panic recovery, error publication and observer return.
// Timeout never authorizes shared model/writer/global teardown.
var errNativeObserverNotJoined = errors.New("native probe observer did not join")

type nativeProbeObserver struct {
	done chan struct{}
	err  error
}

func startNativeProbeObserver(cancel func(), work func() error) *nativeProbeObserver {
	o := &nativeProbeObserver{done: make(chan struct{})}
	go func() {
		defer close(o.done)
		defer func() {
			if v := recover(); v != nil {
				o.err = fmt.Errorf("native probe observer panic: %v", v)
			}
			if o.err != nil {
				cancel()
			}
		}()
		o.err = work()
	}()
	return o
}
func retireNativeProbe(cancel func(), o *nativeProbeObserver, timeout time.Duration, cleanup func() error) error {
	cancel()
	if o != nil {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-o.done:
		case <-timer.C:
			return errNativeObserverNotJoined
		}
	}
	var observerErr error
	if o != nil {
		observerErr = o.err
	}
	return errors.Join(observerErr, cleanup())
}

func runNativePlacementDesktopProbe() (result error) {
	w, err := newNativeProbeWriter(os.Getenv("VELLUM_NATIVE_PLACEMENT_PROBE"))
	if err != nil {
		return err
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	var m *model
	var a *app.Application
	var observer *nativeProbeObserver
	defer func() {
		err := retireNativeProbe(cancel, observer, 3*time.Second, func() error {
			if a != nil {
				a.Close()
			}
			if m != nil {
				m.palette.Close()
			}
			core.DebugSettings.DisableWindowGeometrySaver = previous
			return w.log.Close()
		})
		if err != nil {
			// Keep secondary failures even when a main-thread panic unwinds past result.
			fmt.Fprintln(os.Stderr, "native probe retirement failure:", err)
		}
		if errors.Is(err, errNativeObserverNotJoined) {
			// Do not unwind shared teardown while a worker may still access it. This
			// explicit probe fails at its process boundary; the owned outer watchdog
			// reaps the group and records failure, never a clean observer shutdown.
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		result = errors.Join(result, err)
	}()
	m, err = newModel()
	if err != nil {
		return err
	}
	a, err = app.New(app.Config{})
	if err != nil {
		return err
	}
	p := &nativePlacementProbe{m: m, a: a}
	p.ref, err = a.OpenWindow(app.Window{ID: "placement", Title: "Vellum native placement probe", Width: 1100, Height: 800, View: p.view})
	if err != nil {
		return err
	}
	observer = startNativeProbeObserver(cancel, func() error {
		var last []byte
		var ax cogent.AccessibilityInspection
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
			for _, marker := range []string{filepath.Join(w.root, "stop-live"), "/private/tmp/vellum-macos-same-name-live-ddd23c6/stop-live"} {
				if _, err := os.Stat(marker); err == nil {
					return fmt.Errorf("probe stopped by marker")
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			var state nativePlacementState
			if err := a.Do(func() {
				state = p.state()
				if state.Gesture.Active && state.Gesture.Dragging && p.removeOnDrag != "" {
					mode := p.removeOnDrag
					p.removeOnDrag = ""
					p.removals++
					switch mode {
					case "source":
						p.hideSource = true
					case "target":
						p.hideTarget = true
					case "window":
						a.Close()
					}
					_ = a.Invalidate("placement")
				}
			}); err != nil {
				if errors.Is(err, app.ErrClosed) {
					continue
				}
				return err
			}
			encoded, err := json.Marshal(state)
			if err != nil {
				return err
			}
			if bytes.Equal(encoded, last) && ax.Reachable {
				continue
			}
			query, stop := context.WithTimeout(ctx, 2*time.Second)
			next, axErr := cogent.InspectAccessibility(query, "placement")
			stop()
			if axErr == nil {
				ax = next
			}
			if err := w.write(map[string]any{"event": "placement.probe.sample", "observed_utc": time.Now().UTC(), "state": state, "accessibility": ax, "accessibility_error": fmt.Sprint(axErr), "native_platform": system.TheApp.Platform().String(), "geometry_persistence_disabled": core.DebugSettings.DisableWindowGeometrySaver}); err != nil {
				return err
			}
			last = encoded
		}
	})
	nativeErr := a.Run(ctx, cogent.Driver{Commands: m.commands})
	deadline := ctx.Err()
	cancel()
	// Join before reading the final caller-owned model or writing a final record.
	select {
	case <-observer.done:
	case <-time.After(3 * time.Second):
		return errNativeObserverNotJoined
	}
	final := p.state()
	writeErr := w.write(map[string]any{"event": "placement.probe.closed", "state": final, "native_error": fmt.Sprint(nativeErr), "observer_error": fmt.Sprint(observer.err), "deadline_error": fmt.Sprint(deadline), "native_platform": system.TheApp.Platform().String(), "geometry_persistence_disabled": core.DebugSettings.DisableWindowGeometrySaver})
	if nativeErr != nil || observer.err != nil || deadline != nil || final.Gesture.Active || final.Reserved != 0 {
		return errors.Join(writeErr, fmt.Errorf("native=%v observer=%v deadline=%v state=%+v", nativeErr, observer.err, deadline, final))
	}
	return writeErr
}
