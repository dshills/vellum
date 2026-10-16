//go:build darwin && cgo && !offscreen && vellum_native_checks

// workspace-smoke exercises the public professional workspace through actual
// Cocoa accessibility methods. It does not synthesize physical pointer input.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/replay/internal/smoke"
	"github.com/dshills/vellum/examples/replay/workspace"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/workbench"
)

type observation struct {
	Model   workspace.Snapshot
	Windows []string
	Grid    *cogent.GridInspection
	Tree    *cogent.TreeInspection
}
type reply struct {
	value observation
	err   error
}
type controller struct {
	a       *app.Application
	f       *workspace.Fixture
	records []smoke.Record
	last    observation
}
type report struct {
	Event                       string                         `json:"event"`
	Result                      string                         `json:"result"`
	Rows                        int                            `json:"stored_rows"`
	PlannedActions              int                            `json:"planned_actions"`
	GPUSceneRequested           bool                           `json:"gpu_scene_requested"`
	GeometryPersistenceDisabled bool                           `json:"geometry_persistence_disabled_in_fixture"`
	Actions                     []smoke.Record                 `json:"actions"`
	Final                       observation                    `json:"final"`
	NativeActivityBefore        cogent.NativeActivityCheck     `json:"native_activity_before"`
	NativeActivityAfter         cogent.NativeActivityCheck     `json:"native_activity_after"`
	Cleanup                     cogent.NativeServiceCheckState `json:"native_cleanup"`
	WorkerJoined                bool                           `json:"worker_joined"`
	ElapsedSeconds              float64                        `json:"elapsed_seconds"`
	Error                       string                         `json:"error,omitempty"`
	FullP9Qualified             bool                           `json:"full_p9_qualified"`
	FullP11Qualified            bool                           `json:"full_p11_qualified"`
	PhysicalPointerVerified     bool                           `json:"physical_pointer_verified"`
}

func activeDocument(g *workbench.Group) string {
	if g == nil {
		return ""
	}
	if g.ID == "documents" {
		return g.Active
	}
	if v := activeDocument(g.First); v != "" {
		return v
	}
	return activeDocument(g.Second)
}

// Only the posted envelope touches owner-held models. The detached reply owns
// its result; cancellation cannot race a late callback writing caller variables.
func (c *controller) read(ctx context.Context) (observation, error) {
	done := make(chan reply, 1)
	if err := c.a.PostWait(ctx, func() {
		if err := ctx.Err(); err != nil {
			done <- reply{err: err}
			return
		}
		out := observation{Model: c.f.Snapshot()}
		for _, w := range c.a.Windows() {
			out.Windows = append(out.Windows, w.ID)
		}
		if out.Model.Failure != "" {
			done <- reply{err: errors.New(out.Model.Failure)}
			return
		}
		if len(out.Windows) == 2 {
			if g, err := cogent.InspectTree(c.a, c.f.Tree); err == nil {
				out.Tree = &g
			} else if !errors.Is(err, cogent.ErrTreeLayout) && !errors.Is(err, cogent.ErrTreeClosed) {
				done <- reply{err: err}
				return
			}
			if activeDocument(out.Model.Layout.Root) == "grid" {
				if g, err := cogent.InspectGrid(c.a, c.f.Grid); err == nil {
					out.Grid = &g
				} else if !errors.Is(err, cogent.ErrGridLayout) && !errors.Is(err, cogent.ErrGridClosed) {
					done <- reply{err: err}
					return
				}
			}
		}
		done <- reply{value: out}
	}); err != nil {
		return observation{}, err
	}
	select {
	case r := <-done:
		return r.value, r.err
	case <-ctx.Done():
		return observation{}, ctx.Err()
	}
}

func pause(ctx context.Context) error {
	t := time.NewTimer(20 * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
func await(ctx context.Context, check func(context.Context) (bool, error)) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ok, err := check(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if err = pause(ctx); err != nil {
			return err
		}
	}
}
func native(ctx context.Context, id string) (cogent.AccessibilityInspection, error) {
	s, err := cogent.InspectAccessibility(ctx, id)
	if err == nil && (!s.Reachable || !s.NativeWindow) {
		return s, errors.New("native window is not reachable")
	}
	return s, err
}
func find(s cogent.AccessibilityInspection, domain, role string) (cogent.AccessibilityElement, bool, error) {
	var result cogent.AccessibilityElement
	count := 0
	for _, e := range s.Elements {
		if e.DomainID == domain && e.Role == role {
			result = e
			count++
		}
	}
	if count > 1 {
		return result, false, fmt.Errorf("ambiguous native target %s/%s", domain, role)
	}
	return result, count == 1, nil
}
func usable(e cogent.AccessibilityElement) bool {
	return e.Enabled && e.ObjectIdentity != 0 && e.Frame[2] > 0 && e.Frame[3] > 0
}

// Completed geometry includes retained overscan cells whose clipped bounds are
// empty. Only a visible cell can become a usable native accessibility target.
func distinctVisibleCell(cells []cogent.GridCellGeometry, row, column string) (cogent.GridCellGeometry, bool) {
	for _, cell := range cells {
		if !cell.Bounds.Empty() && (cell.RowKey != row || cell.ColumnID != column) {
			return cell, true
		}
	}
	return cogent.GridCellGeometry{}, false
}

func (c *controller) action(ctx context.Context, name, domain, role, kind, value string, accepted func(observation) bool, completed func(cogent.AccessibilityInspection) bool) error {
	step, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	r, err := smoke.Action(step, name, smoke.Hooks{
		Ready: func(inner context.Context) (string, error) {
			s, e := native(inner, "workspace")
			if e != nil {
				return "", e
			}
			target, ok, e := find(s, domain, role)
			if e != nil {
				return "", e
			}
			if ok && usable(target) {
				return target.ID, nil
			}
			return "", nil
		},
		Perform: func(inner context.Context, id string) (bool, error) {
			return cogent.PerformAccessibilityAction(inner, "workspace", id, kind, value)
		},
		Accepted: func(inner context.Context) (bool, error) {
			o, e := c.read(inner)
			if e != nil {
				return false, e
			}
			c.last = o
			if !accepted(o) {
				return false, nil
			}
			if completed == nil {
				return true, nil
			}
			s, e := native(inner, "workspace")
			return e == nil && completed(s), e
		},
	})
	c.records = append(c.records, r)
	if e := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.workspace-native.action", "action": r}); e != nil {
		return errors.Join(err, e)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// This independent streaming oracle does not call the fixture's generator or
// retain a second dataset. Its recurrence and byte order define expected input.
func expectedData(ctx context.Context, rows int) (digest string, maximum uint32, filtered int, err error) {
	h := sha256.New()
	var block [workspace.Columns * 4]byte
	first := uint32(7)
	for row := 0; row < rows; row++ {
		if row%256 == 0 {
			if err = ctx.Err(); err != nil {
				return
			}
		}
		v := first
		for col := 0; col < workspace.Columns; col++ {
			binary.LittleEndian.PutUint32(block[col*4:], v)
			if col == 1 && strings.Contains(strconv.FormatUint(uint64(v), 10), "7") {
				filtered++
			}
			v = (v + 13) % 10007
		}
		h.Write(block[:])
		maximum = max(maximum, first)
		first = (first + 97) % 10007
	}
	return hex.EncodeToString(h.Sum(nil)), maximum, filtered, nil
}

func (c *controller) scenario(ctx context.Context, rows int) error {
	waitCtx, stop := context.WithTimeout(ctx, 20*time.Second)
	defer stop()
	// Startup absence is retryable only here. Later unexpected attachment errors
	// fail rather than being silently treated as an incomplete frame.
	var startupErr error
	if err := await(waitCtx, func(inner context.Context) (bool, error) {
		for _, id := range []string{"workspace", "inspector"} {
			if _, e := native(inner, id); e != nil {
				startupErr = e
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return errors.Join(err, startupErr)
	}
	digest, maximum, filtered, err := expectedData(ctx, rows)
	if err != nil {
		return err
	}
	readyWorker := func(o observation) bool {
		return o.Model.Task.WorkerDone && o.Model.Task.OwnerDone && o.Model.Task.Error == ""
	}
	value := func(domain, role string, want any) func(cogent.AccessibilityInspection) bool {
		return func(s cogent.AccessibilityInspection) bool {
			e, ok, err := find(s, domain, role)
			return err == nil && ok && usable(e) && e.Value == want
		}
	}
	press := func(name, domain string, accepted func(observation) bool) error {
		return c.action(ctx, name, domain, "AXButton", "press", "", accepted, nil)
	}
	if err = press("open-project", "open", func(o observation) bool {
		return o.Model.Loaded && readyWorker(o) && o.Model.StoredRows == rows && o.Model.StoredColumns == 20 && o.Model.StoredTreeNodes == workspace.TreeNodes && o.Model.DataSHA256 == digest && o.Tree != nil && len(o.Tree.Rows) > 0
	}); err != nil {
		return err
	}
	if err = c.action(ctx, "edit-project-name", "name", "AXTextField", "set", "Native workspace", func(o observation) bool { return o.Model.Name == "Native workspace" && o.Model.History.UndoCount == 1 }, value("name", "AXTextField", "Native workspace")); err != nil {
		return err
	}
	if err = c.action(ctx, "undo-name", "undo", "AXButton", "press", "", func(o observation) bool { return o.Model.Name == "North Star" && o.Model.History.RedoCount == 1 }, value("name", "AXTextField", "North Star")); err != nil {
		return err
	}
	if err = c.action(ctx, "redo-name", "redo", "AXButton", "press", "", func(o observation) bool { return o.Model.Name == "Native workspace" && o.Model.History.UndoCount == 1 }, value("name", "AXTextField", "Native workspace")); err != nil {
		return err
	}
	if err = press("place-block-through-palette-activation", "tool", func(o observation) bool { return len(o.Model.Canvas.Objects) == 1 && o.Model.History.UndoCount == 2 }); err != nil {
		return err
	}
	placed := c.last.Model.Canvas.Objects[0]
	if err = press("undo-placement", "undo", func(o observation) bool { return len(o.Model.Canvas.Objects) == 0 && o.Model.History.RedoCount == 1 }); err != nil {
		return err
	}
	if err = press("redo-same-placement", "redo", func(o observation) bool {
		return len(o.Model.Canvas.Objects) == 1 && o.Model.Canvas.Objects[0] == placed && o.Model.History.UndoCount == 2
	}); err != nil {
		return err
	}
	if err = press("sort-stored-table", "sort", func(o observation) bool {
		return o.Model.Sorted && readyWorker(o) && o.Model.FirstValue == strconv.FormatUint(uint64(maximum), 10)
	}); err != nil {
		return err
	}
	if err = press("filter-stored-table", "filter", func(o observation) bool {
		return o.Model.Filtered && readyWorker(o) && o.Model.Grid.Stamp.Rows == filtered
	}); err != nil {
		return err
	}
	if err = press("clear-table-filter", "filter", func(o observation) bool {
		return !o.Model.Filtered && readyWorker(o) && o.Model.Grid.Stamp.Rows == rows
	}); err != nil {
		return err
	}
	if err = press("filter-stored-outline", "tree-filter", func(o observation) bool {
		return o.Model.Tree.Filter == "Component 000100" && !o.Model.Tree.Pending && readyWorker(o) && o.Model.Tree.VisibleRows == 2 && o.Tree != nil && len(o.Tree.Rows) == 2
	}); err != nil {
		return err
	}
	if err = c.action(ctx, "show-table", "grid", "AXRadioButton", "press", "", func(o observation) bool {
		if activeDocument(o.Model.Layout.Root) != "grid" || o.Grid == nil {
			return false
		}
		_, ok := distinctVisibleCell(o.Grid.Cells, o.Model.Grid.Active.RowKey, o.Model.Grid.Active.ColumnID)
		return ok
	}, func(s cogent.AccessibilityInspection) bool {
		e, ok, err := find(s, "table", "AXTable")
		return err == nil && ok && usable(e) && !e.GridPending && e.GridGeneration > 0
	}); err != nil {
		return err
	}
	// Choose a visible cell different from the current cursor. Its measured
	// identity, native action and accepted model cursor must all agree.
	cell, found := distinctVisibleCell(c.last.Grid.Cells, c.last.Model.Grid.Active.RowKey, c.last.Model.Grid.Active.ColumnID)
	if !found {
		return errors.New("no distinct visible measured table cell")
	}
	if err = c.action(ctx, "select-table-cell", cell.ID, "AXCell", "press", "", func(o observation) bool {
		return o.Model.Grid.Active.RowKey == cell.RowKey && o.Model.Grid.Active.ColumnID == cell.ColumnID && o.Grid != nil
	}, nil); err != nil {
		return err
	}
	if err = press("open-modal", "modal", func(o observation) bool { return o.Model.Modal }); err != nil {
		return err
	}
	if err = press("close-modal", "close-modal", func(o observation) bool { return !o.Model.Modal }); err != nil {
		return err
	}
	for _, tab := range []string{"notes", "canvas"} {
		if err = c.action(ctx, "show-"+tab, tab, "AXRadioButton", "press", "", func(o observation) bool { return activeDocument(o.Model.Layout.Root) == tab }, func(s cogent.AccessibilityInspection) bool {
			e, ok, err := find(s, tab, "AXRadioButton")
			return err == nil && ok && usable(e) && e.Selected
		}); err != nil {
			return err
		}
	}
	// Observe the independently native inspector after all shared-model edits.
	if err = await(ctx, func(inner context.Context) (bool, error) {
		s, e := native(inner, "inspector")
		if e != nil {
			return false, e
		}
		n, ok, e := find(s, "project", "AXStaticText")
		return e == nil && ok && n.Name == "Native workspace" && n.ObjectIdentity != 0, e
	}); err != nil {
		return err
	}
	if err = press("close-main-preserve-inspector", "close-main", func(o observation) bool { return len(o.Windows) == 1 && o.Windows[0] == "inspector" }); err != nil {
		return err
	}
	_, err = native(ctx, "inspector")
	return err
}

func run(rows int, gpuScene bool, timeout time.Duration) (out report, err error) {
	started := time.Now()
	out = report{Event: "vellum.workspace-native.result", Result: "FAILED_NATIVE_WORKSPACE", Rows: rows, PlannedActions: 18, GPUSceneRequested: gpuScene}
	defer func() {
		out.ElapsedSeconds = time.Since(started).Seconds()
		if err != nil {
			out.Error = err.Error()
		}
	}()
	if (rows != 10_000 && rows != workspace.MaxRows) || timeout < 10*time.Second || timeout > 5*time.Minute {
		return out, errors.New("rows 10000 or 1000000 and timeout 10s..5m required")
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	defer func() { core.DebugSettings.DisableWindowGeometrySaver = previous }()
	out.GeometryPersistenceDisabled = true
	a, err := app.New(app.Config{QueueCapacity: 128, CompletionCapacity: 8})
	if err != nil {
		return out, err
	}
	defer a.Close()
	ex := &execution.Goroutine{}
	defer ex.Close()
	f, err := workspace.New(a, ex, rows)
	if err != nil {
		return out, err
	}
	defer f.Palette.Close()
	remove, err := cogent.ObserveNativeServiceCleanup()
	if err != nil {
		return out, err
	}
	defer remove()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := &controller{a: a, f: f}
	finished, scenarioDone := smoke.StartScenario(func() error { return c.scenario(ctx, rows) }, a.Close)
	nativeErrors := make(chan error, 1)
	out.NativeActivityBefore = cogent.NativeActivitySnapshot()
	runErr := a.Run(ctx, cogent.Driver{Commands: f.Commands, GPUSceneRendering: gpuScene, OnError: func(e error) {
		select {
		case nativeErrors <- e:
		default:
		}
	}})
	cancel()
	a.Close()
	var scenarioErr error
	joined := false
	select {
	case <-scenarioDone:
		scenarioErr = <-finished
		joined = true
	case <-time.After(6 * time.Second):
		scenarioErr = errors.New("native scenario did not join after driver exit")
	}
	select {
	case e := <-nativeErrors:
		scenarioErr = errors.Join(scenarioErr, e)
	default:
	}
	joinCtx, joinCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer joinCancel()
	joinErr := f.WaitWorker(joinCtx)
	out.WorkerJoined = joinErr == nil
	// The controller's result is read only after its worker has joined. On a
	// failed join its mutable fields are deliberately omitted from this report.
	if joined {
		out.Actions = c.records
		out.Final = c.last
	}
	if scenarioErr == nil {
		if len(out.Actions) != out.PlannedActions {
			scenarioErr = errors.New("native action plan incomplete")
		}
		for _, action := range out.Actions {
			if action.Attempts != 1 || action.Admissions != 1 || !action.OwnerCompleted {
				scenarioErr = errors.Join(scenarioErr, errors.New("native action receipt incomplete"))
				break
			}
		}
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	out.Cleanup = cleanup
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 || cleanup.FilePanels.Tracked != 0 || cleanup.Menu.Owned || cleanup.Tray.Present {
		scenarioErr = errors.Join(scenarioErr, errors.New("native terminal inventory not retired"))
	}
	out.NativeActivityAfter = cogent.NativeActivitySnapshot()
	if !out.NativeActivityBefore.Available || out.NativeActivityBefore.ProcessID != os.Getpid() || out.NativeActivityBefore.Saturated || !out.NativeActivityAfter.Available || out.NativeActivityAfter.ProcessID != os.Getpid() || out.NativeActivityAfter.Saturated || out.NativeActivityAfter.SurfacePresentAttempts <= out.NativeActivityBefore.SurfacePresentAttempts || out.NativeActivityAfter.EmptyEventPostAttempts < out.NativeActivityBefore.EmptyEventPostAttempts {
		scenarioErr = errors.Join(scenarioErr, errors.New("native present-attempt instrumentation unavailable"))
	}
	err = errors.Join(runErr, scenarioErr, joinErr)
	if err == nil {
		out.Result = "PASS_NATIVE_WORKSPACE"
	}
	return out, err
}
func main() {
	rows := flag.Int("rows", 10000, "stored rows: 10000 or 1000000")
	gpu := flag.Bool("gpu-scene", false, "request optional GPU main-scene paint route; fallback remains possible")
	timeout := flag.Duration("timeout", 3*time.Minute, "cooperative native scenario deadline: 10s..5m")
	flag.Parse()
	out, err := run(*rows, *gpu, *timeout)
	if e := json.NewEncoder(os.Stdout).Encode(out); e != nil {
		os.Exit(2)
	}
	if err != nil {
		os.Exit(1)
	}
}
