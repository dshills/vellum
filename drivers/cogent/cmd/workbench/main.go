package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	exit := flag.Duration("exit-after", 0, "close after duration for smoke checks")
	treeCount := flag.Int("tree-rows", 100000, "component records indexed in the outline")
	rowCount := flag.Int("rows", 1000000, "actual component records")
	layoutPath := flag.String("layout", "", "layout file for explicit Save/Restore commands")
	smoke := flag.Bool("smoke", false, "exercise registered actions and worker completion")
	documentSmoke := flag.Bool("document-smoke", false, "exercise native unsaved-close and temporary document save/open")
	enduranceDuration := flag.Duration("endurance", 0, "run composed workbench activity for 10s..12h (10m minimum for steady endurance evidence)")
	enduranceOutput := flag.String("endurance-output", "", "new absolute private directory for endurance observations and trace")
	automationEnabled := flag.Bool("automation", false, "enable private local semantic API and print its capability descriptor")
	flag.Parse()
	enduranceConfig := enduranceConfig{duration: *enduranceDuration, output: *enduranceOutput}
	if err := enduranceConfig.validate(); err != nil {
		must(err)
	}
	if *enduranceDuration > 0 && (*smoke || *documentSmoke || *exit > 0 || *automationEnabled) {
		must(errors.New("endurance requires a dedicated process"))
	}
	if *enduranceDuration > 0 {
		disableEnduranceGeometrySaver()
	}
	if *documentSmoke && (*smoke || *exit > 0) {
		must(errors.New("document-smoke requires its own native process without smoke/exit-after"))
	}
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	if *rowCount < 100000 || *rowCount > 1000000 {
		must(fmt.Errorf("rows must be 100,000..1,000,000"))
	}
	if *treeCount < 1 || *treeCount > *rowCount {
		must(fmt.Errorf("tree-rows must be 1..rows"))
	}
	parts := newParts(*rowCount)
	selectedIndex := 0
	inspectorExpanded := true
	selectionRevision := uint64(1)
	count := 0
	status := "Ready"
	var task *tasks.Task
	var sortTask *tasks.Task
	var table *virtual.Table
	var gpuVerified atomic.Bool
	var board *canvas.Scene
	var selectID func(string)
	defaultLayout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "workspace", Axis: workbench.Horizontal, Ratio: .18, First: &workbench.Group{ID: "outline", Tabs: []string{"outline"}, Active: "outline"}, Second: &workbench.Group{ID: "center-right", Axis: workbench.Horizontal, Ratio: .75, First: &workbench.Group{ID: "documents", Tabs: []string{"board-a", "board-b"}, Active: "board-a"}, Second: &workbench.Group{ID: "inspector", Tabs: []string{"inspector"}, Active: "inspector"}}}}
	workspace, err := workbench.New(defaultLayout, []string{"outline", "board-a", "board-b", "inspector"})
	must(err)
	registry := command.New()
	manager, err := tasks.New(a, 4, func() { a.Invalidate("main") })
	must(err)
	must(registry.Register(command.Command{ID: "project.increment", Label: "Next component", Shortcut: "Meta+I", Execute: func(context.Context) error {
		count++
		selectID(parts.Row((selectedIndex + 1) % parts.Len()).ID)
		status = "Selected " + parts.Row(selectedIndex).Cells[0]
		return nil
	}}))
	must(registry.Register(command.Command{ID: "analysis.calculate", Label: "Calculate", Shortcut: "Meta+B", Available: func() error {
		if task != nil && task.Snapshot().Running {
			return errors.New("calculation already running")
		}
		return nil
	}, Execute: func(ctx context.Context) error {
		snapshot := parts.Snapshot()
		expectedRevision := parts.Revision()
		var err error
		task, err = tasks.Start(manager, ctx, func(ctx context.Context, report func(float64, string)) (int64, error) {
			var total int64
			for i := 0; i < snapshot.Len(); i++ {
				p := snapshot.parts.at(i)
				if i%1024 == 0 {
					if err := ctx.Err(); err != nil {
						return 0, err
					}
					report(float64(i)/float64(snapshot.Len()), "Summing component values")
				}
				total += int64(p.Value)
			}
			return total, nil
		}, func(total int64, err error) {
			if expectedRevision != parts.Revision() {
				status = "Stale calculation discarded"
				return
			}
			if err != nil {
				status = err.Error()
			} else {
				status = fmt.Sprintf("Sum: %d", total)
				if *smoke {
					fmt.Printf("Snapshot calculation committed revision%d sum%d\n", expectedRevision, total)
				}
			}
		})
		return err
	}}))
	must(registry.Register(command.Command{ID: "analysis.cancel", Label: "Cancel", Shortcut: "Meta+.", Available: func() error {
		if (task == nil || !task.Snapshot().Running) && (sortTask == nil || !sortTask.Snapshot().Running) {
			return errors.New("no running task")
		}
		return nil
	}, Execute: func(context.Context) error {
		if task != nil {
			task.Cancel()
		}
		if sortTask != nil {
			sortTask.Cancel()
		}
		return nil
	}}))
	var mainRef app.WindowRef
	must(registry.Register(command.Command{ID: "app.close", Label: "Close", Shortcut: "Meta+Shift+X", Execute: func(context.Context) error {
		err := a.RequestCloseWindow(mainRef)
		if errors.Is(err, app.ErrCloseRefused) {
			return nil
		}
		return err
	}}))
	must(registry.Register(command.Command{ID: "layout.move-inspector", Label: "Dock inspector", Execute: func(context.Context) error { return workspace.Move("inspector", "documents", 0) }}))
	must(registry.Register(command.Command{ID: "layout.split-inspector", Label: "Split inspector", Execute: func(context.Context) error {
		return workspace.Split("inspector", "documents", "inspector-split", workbench.Vertical, false)
	}}))
	must(registry.Register(command.Command{ID: "layout.reset", Label: "Reset layout", Execute: func(context.Context) error { return workspace.Restore(defaultLayout) }}))
	must(registry.Register(command.Command{ID: "layout.save", Label: "Save layout", Available: func() error {
		if *layoutPath == "" {
			return errors.New("set -layout to choose a file")
		}
		return nil
	}, Execute: func(context.Context) error { return workspace.Save(*layoutPath) }}))
	must(registry.Register(command.Command{ID: "layout.restore", Label: "Restore layout", Available: func() error {
		if *layoutPath == "" {
			return errors.New("set -layout to choose a file")
		}
		return nil
	}, Execute: func(context.Context) error { return workspace.Load(*layoutPath) }}))
	table, err = virtual.NewTable(parts, 20)
	must(err)
	treeModel, err := buildOutline(*treeCount, parts.Len())
	must(err)
	selectComponent := func(id string) error {
		index, err := componentIndex(parts, id)
		if err != nil {
			return err
		}
		if index != selectedIndex && selectionRevision == math.MaxUint64 {
			return history.ErrRevisionExhausted
		}
		if index != selectedIndex {
			selectionRevision++
		}
		selectedIndex = index
		table.Selected = id
		if board != nil {
			board.Selected = id
		}
		status = "Selected " + parts.Row(selectedIndex).Cells[0]
		return nil
	}
	selectID = func(id string) {
		if !strings.HasPrefix(id, "component:") {
			return
		}
		if err := selectComponent(id); err != nil {
			status = err.Error()
		}
	}
	tableView := &cogent.VirtualTable{Owner: a, Window: "main", Model: table, Columns: []string{"Reference", "Value", "Package", "Rotation"}, Select: func(r virtual.Row) { selectID(r.ID) }}
	treeView := &cogent.VirtualTree{Owner: a, Window: "main", Model: treeModel, Count: 24, Select: selectID}
	must(registry.Register(command.Command{ID: "table.next", Label: "Next table page", Shortcut: "Meta+DownArrow", Execute: func(context.Context) error {
		return table.SetViewport(table.StartRow()+table.VisibleCount(), table.VisibleCount())
	}}))
	must(registry.Register(command.Command{ID: "table.previous", Label: "Previous table page", Execute: func(context.Context) error {
		start := table.StartRow() - table.VisibleCount()
		if start < 0 {
			start = 0
		}
		return table.SetViewport(start, table.VisibleCount())
	}}))
	must(registry.Register(command.Command{ID: "table.sort-value", Label: "Sort by value", Available: func() error {
		if sortTask != nil && sortTask.Snapshot().Running {
			return errors.New("sort already running")
		}
		return nil
	}, Execute: func(ctx context.Context) error {
		snapshot := parts.Snapshot()
		token := table.ProjectionToken()
		expectedRevision := parts.Revision()
		var err error
		sortTask, err = tasks.Start(manager, ctx, func(ctx context.Context, _ func(float64, string)) (virtual.Projection, error) {
			return virtual.ProjectContext(ctx, snapshot, "", 1, false)
		}, func(indices virtual.Projection, err error) {
			if err != nil {
				status = err.Error()
				return
			}
			if parts.Revision() != expectedRevision {
				status = "Stale projection discarded"
				return
			}
			if err := table.ApplyProjection(indices, token); err != nil {
				status = err.Error()
			} else {
				status = "Table sorted by value"
			}
		})
		return err
	}}))
	journal, err := history.New(256)
	must(err)
	editor := &partEditor{source: parts, table: table, journal: journal}
	references := &fileReferences{}
	referenceFiles := func(e ui.FileDropEvent) bool {
		if err := references.replace(journal, e.Paths); err != nil {
			status = err.Error()
			a.Invalidate(e.WindowID)
			return false
		}
		status = fmt.Sprintf("Referenced %d files", len(references.paths))
		return true
	}
	must(registry.Register(command.Command{ID: "component.rotate", Label: "Rotate 90°", Shortcut: "Meta+R", Execute: func(context.Context) error {
		next := parts.parts.at(selectedIndex)
		next.Rotation = (next.Rotation + 90) % 360
		if err := editor.change(selectedIndex, "Rotate component", next); err != nil {
			return err
		}
		status = "Component rotated"
		return nil
	}}))
	must(registry.Register(command.Command{ID: "component.raise-value", Label: "Increase value", Execute: func(context.Context) error {
		next := parts.parts.at(selectedIndex)
		if next.Value > math.MaxUint32-100 {
			return errors.New("component value would overflow")
		}
		next.Value += 100
		if err := editor.change(selectedIndex, "Increase component value", next); err != nil {
			return err
		}
		status = "Component value changed; sort reset"
		return nil
	}}))
	for _, operation := range []struct {
		id, label, chord string
		redo             bool
	}{{"edit.undo", "Undo", "Meta+Z", false}, {"edit.redo", "Redo", "Meta+Shift+Z", true}} {
		op := operation
		must(registry.Register(command.Command{ID: op.id, Label: op.label, Shortcut: op.chord, ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
			state := journal.Snapshot()
			if state.Invalid {
				return history.ErrInvalid
			}
			if state.Busy {
				return history.ErrBusy
			}
			if (!op.redo && state.UndoCount == 0) || (op.redo && state.RedoCount == 0) {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error {
			state := journal.Snapshot()
			label := state.UndoLabel
			replay := journal.Undo
			if op.redo {
				label = state.RedoLabel
				replay = journal.Redo
			}
			if err := replay(); err != nil {
				return err
			}
			status = op.label + ": " + label
			return nil
		}}))
	}

	objects := make([]canvas.Rect, 96)
	for i := range objects {
		objects[i] = canvas.Rect{ID: parts.Row(i).ID, X: float64(i%16) * 70, Y: float64(i/16) * 50, Width: 50, Height: 25, Color: color.RGBA{67, 197, 166, 255}}
	}
	board, err = canvas.New(objects)
	must(err)
	editor.board = board
	board.View.SetSize(800, 400)
	board.Fit()
	gpuCanvas := &cogent.GPUCanvas{Owner: a, Window: "main", Model: board, Select: selectID}
	must(registry.Register(command.Command{ID: "canvas.fit", Label: "Fit board", Shortcut: "Meta+Shift+F", Execute: func(context.Context) error { board.Fit(); status = "Board fitted"; return nil }}))
	must(registry.Register(command.Command{ID: "canvas.zoom-in", Label: "Zoom in", Shortcut: "Meta+Equal", Execute: func(context.Context) error {
		return board.View.ZoomAt(canvas.Point{X: board.View.Width / 2, Y: board.View.Height / 2}, 1.25)
	}}))
	must(registry.Register(command.Command{ID: "canvas.zoom-out", Label: "Zoom out", Shortcut: "Meta+Minus", Execute: func(context.Context) error {
		return board.View.ZoomAt(canvas.Point{X: board.View.Width / 2, Y: board.View.Height / 2}, .8)
	}}))
	dock := &cogent.Workbench{Owner: a, Window: "main", Workspace: workspace, OnError: func(err error) { status = err.Error() }}
	for _, id := range []string{"outline", "board-a", "board-b", "inspector"} {
		id := id
		title := id
		dock.Panels = append(dock.Panels, cogent.Panel{ID: id, Title: title, View: func() ui.Node {
			text := fmt.Sprintf("%s · accepted actions %d", id, count)
			if id == "outline" {
				text = "Project / Board A / Board B"
			}
			if id == "inspector" {
				text = "Inspector · " + strings.Join(parts.Row(selectedIndex).Cells, " · ")
			}
			if id == "outline" {
				return ui.Node{ID: "outline.data", Kind: ui.Extension, Grow: true, Custom: treeView}
			}
			if id == "board-a" {
				return ui.Node{ID: "board-a.root", Kind: ui.Column, Grow: true, FileDrop: referenceFiles, Children: []ui.Node{{ID: "board-a.canvas", Kind: ui.Extension, Grow: true, MinHeight: 240, Custom: gpuCanvas}, {ID: "board-a.table", Kind: ui.Extension, Grow: true, Custom: tableView}}}
			}
			if id == "inspector" {
				return inspectorView(parts, editor, inspectorExpanded, func(open bool) { inspectorExpanded = open }, func() int { return selectedIndex }, func(text string) { status = text })
			}
			return ui.Node{ID: id + ".root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: id + ".title", Kind: ui.Label, Text: text}, {ID: id + ".detail", Kind: ui.Label, Text: "Panel content · Vellum keyed UI"}}}
		}})
	}
	var persistence *document.Session[documentImage]
	workflow := &documentWorkflow{owner: a, window: "main", status: func(text string) { status = text }}
	bar := &cogent.CommandBar{Owner: a, Registry: registry, Toolbar: []string{"file.open", "file.save", "file.save-as", "edit.undo", "edit.redo", "project.increment", "analysis.calculate", "analysis.cancel"}, OnError: func(err error) { status = err.Error() }}
	mainRef, err = a.OpenWindow(app.Window{ID: "main", Title: "Vellum Workbench", Width: 1280, Height: 900, CloseRequest: workflow.requestClose, View: func() ui.Node {
		progress := ""
		if task != nil {
			s := task.Snapshot()
			if s.Running {
				progress = fmt.Sprintf(" · %.0f%% %s", s.Progress*100, s.Message)
			}
		}
		if journal.Snapshot().Dirty {
			progress += " · Unsaved changes"
		}
		fileStatus := references.summary()
		if persistence != nil {
			ds := persistence.State()
			if ds.Busy {
				progress += " · " + string(ds.Action) + " pending"
			}
			if ds.Path != "" {
				fileStatus = ds.Path + " · " + fileStatus
			}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "title", Kind: ui.Label, Text: "Vellum · Go native workbench"}, {ID: "commands", Kind: ui.Extension, Custom: bar}, {ID: "workbench", Kind: ui.Extension, Grow: true, MinHeight: 500, Custom: dock}, {ID: "files", Kind: ui.Label, Text: fileStatus}, {ID: "status", Kind: ui.Label, Text: status + progress}, workflow.view()}}
	}})
	must(err)
	workflow.windowRef = mainRef

	var documentStore document.Store
	var documentChoose func(context.Context, app.FileOptions) (string, error)
	var smokeDocumentPath string
	var smokeStore *documentSmokeStore
	var documentOutcome chan error
	if *documentSmoke {
		dir, err := os.MkdirTemp("", "vellum-document-native-*")
		must(err)
		defer os.RemoveAll(dir)
		smokeDocumentPath = filepath.Join(dir, "fixture.vellum")
		smokeStore = &documentSmokeStore{}
		documentStore = smokeStore
		documentChoose = func(context.Context, app.FileOptions) (string, error) { return smokeDocumentPath, nil }
		documentOutcome = make(chan error, 1)
	}
	persistence, err = document.New(document.Config[documentImage]{Store: documentStore, Choose: documentChoose, Owner: a, Window: "main", Journal: journal, Filename: "Untitled.vellum", Extensions: []string{"vellum"}, Codec: document.Codec[documentImage]{
		Capture: func() (documentImage, error) { return captureDocument(parts, board, references), nil },
		Encode:  encodeDocument,
		Decode:  func(data []byte) (documentImage, error) { return decodeDocument(data, *treeCount) },
		Install: func(image documentImage) error {
			if err := installDocument(parts, table, board, references, image); err != nil {
				return err
			}
			treeModel = image.outline
			treeView.Model = treeModel
			selectedIndex = min(selectedIndex, parts.Len()-1)
			selectID(parts.Row(selectedIndex).ID)
			return nil
		},
	}, OnComplete: workflow.complete})
	must(err)
	workflow.session = persistence
	fileAvailable := func() error {
		if persistence.State().Busy {
			return document.ErrBusy
		}
		if workflow.pending != "" {
			return app.ErrClosePending
		}
		return nil
	}
	must(registry.Register(command.Command{ID: "file.open", Label: "Open", Shortcut: "Meta+O", Available: fileAvailable, Execute: workflow.open}))
	must(registry.Register(command.Command{ID: "file.save", Label: "Save", Shortcut: "Meta+S", Available: fileAvailable, Execute: func(ctx context.Context) error { return persistence.Save(ctx, document.SaveOptions{}) }}))
	must(registry.Register(command.Command{ID: "file.save-as", Label: "Save As", Shortcut: "Meta+Shift+S", Available: fileAvailable, Execute: func(ctx context.Context) error { return persistence.Save(ctx, document.SaveOptions{Choose: true}) }}))
	if *automationEnabled {
		semantic, err := workbenchAutomation(a, registry, parts, references, func() (int, uint64) { return selectedIndex, selectionRevision }, selectComponent)
		must(err)
		server, err := automation.Start(automation.Config{Owner: a, Window: "main", Registry: semantic})
		must(err)
		defer server.Close()
		// Explicit opt-in writes this credential only to the invoking process's
		// stdout. Never include it in application semantics or audit records.
		must(json.NewEncoder(os.Stdout).Encode(server.Endpoint()))
	}
	if *documentSmoke {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			err := runDocumentSmoke(ctx, a, persistence, workflow, journal, registry, parts, board, references, smokeDocumentPath, smokeStore)
			documentOutcome <- err
			if err != nil {
				a.Close()
			}
		}()
	}
	if *smoke {
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 10*time.Second)
			defer cancel()
			tick := time.NewTicker(50 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
				}
				result := make(chan error, 1)
				if err := a.PostWait(ctx, func() {
					if board.Selected != "component:24" {
						result <- cogent.ErrFrameNotReady
						return
					}
					img, err := gpuCanvas.Snapshot()
					if err != nil {
						result <- err
						return
					}
					o := board.Objects[23]
					point := board.View.Screen(canvas.Point{X: o.X + o.Width/2, Y: o.Y + o.Height/2})
					if point.X < 0 || point.Y < 0 || point.X >= float64(img.Bounds().Dx()) || point.Y >= float64(img.Bounds().Dy()) {
						result <- fmt.Errorf("GPU object outside snapshot: %+v", point)
						return
					}
					pixel := img.NRGBAAt(int(point.X), int(point.Y))
					if pixel.R < 200 || pixel.G < 120 || pixel.B > 100 {
						result <- fmt.Errorf("GPU selected object absent: pixel=%v", pixel)
						return
					}
					fmt.Printf("GPU verified snapshot=%v viewport=%+v objects=%d pixel=%v\n", img.Bounds(), board.View, len(board.Objects), pixel)
					gpuVerified.Store(true)
					result <- nil
				}); err != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case err := <-result:
					if err == nil {
						return
					}
					if errors.Is(err, cogent.ErrFrameNotReady) {
						continue
					}
					fmt.Fprintf(os.Stderr, "GPU smoke failed: %v\n", err)
					a.Close()
					return
				}
			}
		}()

		a.Post(func() {
			must(registry.Invoke(a.Context(), "project.increment"))
			original := parts.parts.at(selectedIndex)
			must(registry.Invoke(a.Context(), "component.rotate"))
			must(registry.Invoke(a.Context(), "component.raise-value"))
			edited := parts.parts.at(selectedIndex)
			must(registry.Invoke(a.Context(), "edit.undo"))
			must(registry.Invoke(a.Context(), "edit.undo"))
			if parts.parts.at(selectedIndex) != original || journal.Snapshot().Dirty {
				must(errors.New("document undo failed"))
			}
			must(registry.Invoke(a.Context(), "edit.redo"))
			must(registry.Invoke(a.Context(), "edit.redo"))
			if parts.parts.at(selectedIndex) != edited || !journal.Snapshot().Dirty {
				must(errors.New("document redo failed"))
			}
			fmt.Println("Document history verified sparse property deltas, undo/redo and dirty state")
			must(registry.Invoke(a.Context(), "analysis.calculate"))
			must(registry.Invoke(a.Context(), "layout.move-inspector"))
			must(registry.Invoke(a.Context(), "layout.reset"))
			must(registry.Invoke(a.Context(), "canvas.fit"))
			must(registry.Invoke(a.Context(), "canvas.zoom-in"))
			must(registry.Invoke(a.Context(), "table.next"))
			row, err := table.SelectVisible(3)
			must(err)
			selectID(row.ID)
			a.Invalidate("main")
		})
	}
	var enduranceDone chan error
	runContext := context.Background()
	var endRun context.CancelFunc
	if *enduranceDuration > 0 {
		runContext, endRun = context.WithTimeout(runContext, *enduranceDuration+90*time.Second)
		defer endRun()
		enduranceDone = make(chan error, 1)
		go func() {
			enduranceDone <- runEndurance(runContext, a, enduranceConfig, registry, parts, table, board, gpuCanvas, manager, func() int { return selectedIndex })
			endRun()
		}()
	}
	if *exit > 0 {
		time.AfterFunc(*exit, a.Close)
	}
	runErr := a.Run(runContext, cogent.Driver{Theme: ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 16})}, Kinds: map[ui.Kind]ui.Style{ui.Row: {Gap: ui.Some(float32(8))}, ui.Column: {Gap: ui.Some(float32(8))}}}, Commands: registry, OnError: func(err error) { status = err.Error() }})
	if enduranceDone != nil {
		endRun()
		runErr = enduranceRunResult(runErr, enduranceDone, 10*time.Second)
	}
	must(runErr)
	if *documentSmoke {
		select {
		case err := <-documentOutcome:
			must(err)
		case <-time.After(90 * time.Second):
			must(errors.New("document smoke ended before outcome"))
		}
	}
	if *smoke && !gpuVerified.Load() {
		must(errors.New("GPU smoke ended before a frame was verified"))
	}
	fmt.Printf("workbench actions=%d status=%s rows=%d pooled=%d tree_nodes=%d selected=%s\n", count, status, parts.Len(), table.VisibleCount(), treeModel.Len(), table.Selected)
	fmt.Printf("canvas objects=%d zoom=%.4f selected=%s\n", len(board.Objects), board.View.Zoom, board.Selected)
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
