package main

import (
	"context"
	"errors"
	"example.com/vellum/app"
	"example.com/vellum/cogent"
	"example.com/vellum/command"
	"example.com/vellum/tasks"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	exit := flag.Duration("exit-after", 0, "close after duration for smoke checks")
	layoutPath := flag.String("layout", "", "layout file for explicit Save/Restore commands")
	smoke := flag.Bool("smoke", false, "exercise registered actions and worker completion")
	flag.Parse()
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	count := 0
	status := "Ready"
	var task *tasks.Task
	defaultLayout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "workspace", Axis: workbench.Horizontal, Ratio: .18, First: &workbench.Group{ID: "outline", Tabs: []string{"outline"}, Active: "outline"}, Second: &workbench.Group{ID: "center-right", Axis: workbench.Horizontal, Ratio: .75, First: &workbench.Group{ID: "documents", Tabs: []string{"board-a", "board-b"}, Active: "board-a"}, Second: &workbench.Group{ID: "inspector", Tabs: []string{"inspector"}, Active: "inspector"}}}}
	workspace, err := workbench.New(defaultLayout, []string{"outline", "board-a", "board-b", "inspector"})
	must(err)
	registry := command.New()
	manager, err := tasks.New(a, 4, func() { a.Invalidate("main") })
	must(err)
	must(registry.Register(command.Command{ID: "project.increment", Label: "Increment", Shortcut: "Meta+I", Execute: func(context.Context) error { count++; status = "Increment accepted"; return nil }}))
	must(registry.Register(command.Command{ID: "analysis.calculate", Label: "Calculate", Shortcut: "Meta+B", Available: func() error {
		if task != nil && task.Snapshot().Running {
			return errors.New("calculation already running")
		}
		return nil
	}, Execute: func(ctx context.Context) error {
		var err error
		task, err = tasks.Start(manager, ctx, func(ctx context.Context, report func(float64, string)) (int64, error) {
			var total int64
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				case <-time.After(10 * time.Millisecond):
				}
				for j := 0; j < 1000; j++ {
					total += int64(i*1000 + j)
				}
				report(float64(i+1)/100, "Calculating 100,000 records")
			}
			return total, nil
		}, func(total int64, err error) {
			if err != nil {
				status = err.Error()
			} else {
				status = fmt.Sprintf("Sum: %d", total)
			}
		})
		return err
	}}))
	must(registry.Register(command.Command{ID: "analysis.cancel", Label: "Cancel", Shortcut: "Meta+.", Available: func() error {
		if task == nil || !task.Snapshot().Running {
			return errors.New("no running task")
		}
		return nil
	}, Execute: func(context.Context) error { task.Cancel(); return nil }}))
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
				text = fmt.Sprintf("Inspector · R%d · 10k · 0603", count+1)
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
			a.Invalidate("main")
		})
	}
	if *exit > 0 {
		time.AfterFunc(*exit, a.Close)
	}
	must(a.Run(context.Background(), cogent.Driver{Commands: registry, OnError: func(err error) { status = err.Error() }}))
	fmt.Printf("workbench actions=%d status=%s\n", count, status)
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
