package main

import (
	"context"
	"errors"
	"example.com/vellum/app"
	"example.com/vellum/cogent"
	"example.com/vellum/command"
	"example.com/vellum/tasks"
	"example.com/vellum/ui"
	"example.com/vellum/virtual"
	"example.com/vellum/workbench"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	exit := flag.Duration("exit-after", 0, "close after duration for smoke checks")
	treeCount := flag.Int("tree-rows", 100000, "component records indexed in the outline")
	rowCount := flag.Int("rows", 1000000, "actual component records")
	layoutPath := flag.String("layout", "", "layout file for explicit Save/Restore commands")
	smoke := flag.Bool("smoke", false, "exercise registered actions and worker completion")
	flag.Parse()
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
	revision := uint64(1)
	count := 0
	status := "Ready"
	var task *tasks.Task
	var sortTask *tasks.Task
	var table *virtual.Table
	defaultLayout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "workspace", Axis: workbench.Horizontal, Ratio: .18, First: &workbench.Group{ID: "outline", Tabs: []string{"outline"}, Active: "outline"}, Second: &workbench.Group{ID: "center-right", Axis: workbench.Horizontal, Ratio: .75, First: &workbench.Group{ID: "documents", Tabs: []string{"board-a", "board-b"}, Active: "board-a"}, Second: &workbench.Group{ID: "inspector", Tabs: []string{"inspector"}, Active: "inspector"}}}}
	workspace, err := workbench.New(defaultLayout, []string{"outline", "board-a", "board-b", "inspector"})
	must(err)
	registry := command.New()
	manager, err := tasks.New(a, 4, func() { a.Invalidate("main") })
	must(err)
	must(registry.Register(command.Command{ID: "project.increment", Label: "Next component", Shortcut: "Meta+I", Execute: func(context.Context) error {
		count++
		selectedIndex = (selectedIndex + 1) % parts.Len()
		table.Selected = parts.Row(selectedIndex).ID
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
		expectedRevision := revision
		var err error
		task, err = tasks.Start(manager, ctx, func(ctx context.Context, report func(float64, string)) (int64, error) {
			var total int64
			for i, p := range snapshot.parts {
				if i%1024 == 0 {
					if err := ctx.Err(); err != nil {
						return 0, err
					}
					report(float64(i)/float64(len(snapshot.parts)), "Summing component values")
				}
				total += int64(p.Value)
			}
			return total, nil
		}, func(total int64, err error) {
			if expectedRevision != revision {
				status = "Stale calculation discarded"
				return
			}
			if err != nil {
				status = err.Error()
			} else {
				status = fmt.Sprintf("Sum: %d", total)
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
	must(registry.Register(command.Command{ID: "app.close", Label: "Close", Shortcut: "Meta+Shift+X", Execute: func(context.Context) error { a.Close(); return nil }}))
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
	treeNodes := []virtual.TreeNode{{ID: "project", Label: fmt.Sprintf("Outline · %d of %d components", *treeCount, parts.Len())}}
	for group := 0; group < (*treeCount+999)/1000; group++ {
		id := fmt.Sprintf("group:%d", group)
		treeNodes = append(treeNodes, virtual.TreeNode{ID: id, Parent: "project", Label: fmt.Sprintf("Group %d", group+1)})
		for j := 0; j < 1000 && group*1000+j < *treeCount; j++ {
			index := group*1000 + j
			treeNodes = append(treeNodes, virtual.TreeNode{ID: "component:" + strconv.Itoa(index+1), Parent: id, Label: fmt.Sprintf("R%07d", index+1)})
		}
	}
	treeModel, err := virtual.NewTree(treeNodes)
	must(err)
	must(treeModel.Toggle("project"))
	selectID := func(id string) {
		if !strings.HasPrefix(id, "component:") {
			return
		}
		index, err := strconv.Atoi(strings.TrimPrefix(id, "component:"))
		if err == nil && index > 0 && index <= parts.Len() {
			selectedIndex = index - 1
			table.Selected = id
			status = "Selected " + parts.Row(selectedIndex).Cells[0]
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
		expectedRevision := revision
		var err error
		sortTask, err = tasks.Start(manager, ctx, func(ctx context.Context, _ func(float64, string)) (virtual.Projection, error) {
			return virtual.ProjectContext(ctx, snapshot, "", 1, false)
		}, func(indices virtual.Projection, err error) {
			if err != nil {
				status = err.Error()
				return
			}
			if revision != expectedRevision {
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
	must(registry.Register(command.Command{ID: "component.rotate", Label: "Rotate 90°", Shortcut: "Meta+R", Execute: func(context.Context) error {
		parts.parts[selectedIndex].Rotation = (parts.parts[selectedIndex].Rotation + 90) % 360
		table.Reset()
		revision++
		parts.revision++
		status = "Component rotated"
		return nil
	}}))
	must(registry.Register(command.Command{ID: "component.raise-value", Label: "Increase value", Execute: func(context.Context) error {
		parts.parts[selectedIndex].Value += 100
		table.Reset()
		revision++
		parts.revision++
		status = "Component value changed; sort reset"
		return nil
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
				return ui.Node{ID: "board-a.data", Kind: ui.Extension, Grow: true, Custom: tableView}
			}
			if id == "inspector" {
				row := parts.Row(selectedIndex)
				return ui.Node{ID: "inspector.root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "inspector.reference", Kind: ui.Label, Text: row.Cells[0]}, {ID: "inspector.value", Kind: ui.Label, Text: "Value: " + row.Cells[1]}, {ID: "inspector.package", Kind: ui.Label, Text: "Package: " + row.Cells[2]}, {ID: "inspector.rotation", Kind: ui.Label, Text: "Rotation: " + row.Cells[3]}, {ID: "inspector.rotate", Kind: ui.Button, Text: "Rotate 90°", Activate: func() {
					if err := registry.Invoke(a.Context(), "component.rotate"); err != nil {
						status = err.Error()
					}
				}}, {ID: "inspector.raise", Kind: ui.Button, Text: "Increase value", Activate: func() {
					if err := registry.Invoke(a.Context(), "component.raise-value"); err != nil {
						status = err.Error()
					}
				}}}}
			}
			return ui.Node{ID: id + ".root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: id + ".title", Kind: ui.Label, Text: text}, {ID: id + ".detail", Kind: ui.Label, Text: "Panel content · Vellum keyed UI"}}}
		}})
	}
	bar := &cogent.CommandBar{Owner: a, Registry: registry, Toolbar: []string{"project.increment", "analysis.calculate", "analysis.cancel"}, OnError: func(err error) { status = err.Error() }}
	must(a.Open(app.Window{ID: "main", Title: "Vellum Workbench", Width: 1100, Height: 760, View: func() ui.Node {
		progress := ""
		if task != nil {
			s := task.Snapshot()
			if s.Running {
				progress = fmt.Sprintf(" · %.0f%% %s", s.Progress*100, s.Message)
			}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "title", Kind: ui.Label, Text: "Vellum · Go native workbench"}, {ID: "commands", Kind: ui.Extension, Custom: bar}, {ID: "workbench", Kind: ui.Extension, Grow: true, MinHeight: 500, Custom: dock}, {ID: "status", Kind: ui.Label, Text: status + progress}}}
	}}))
	if *smoke {
		a.Post(func() {
			must(registry.Invoke(a.Context(), "project.increment"))
			must(registry.Invoke(a.Context(), "analysis.calculate"))
			must(registry.Invoke(a.Context(), "layout.move-inspector"))
			must(registry.Invoke(a.Context(), "layout.reset"))
			must(registry.Invoke(a.Context(), "table.next"))
			row, err := table.SelectVisible(3)
			must(err)
			selectID(row.ID)
			a.Invalidate("main")
		})
	}
	if *exit > 0 {
		time.AfterFunc(*exit, a.Close)
	}
	must(a.Run(context.Background(), cogent.Driver{Commands: registry, OnError: func(err error) { status = err.Error() }}))
	fmt.Printf("workbench actions=%d status=%s rows=%d pooled=%d tree_nodes=%d selected=%s\n", count, status, parts.Len(), table.VisibleCount(), treeModel.Len(), table.Selected)
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
