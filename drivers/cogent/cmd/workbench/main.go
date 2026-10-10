package main

import (
	"context"
	"errors"
	"example.com/vellum/app"
	"example.com/vellum/cogent"
	"example.com/vellum/command"
	"example.com/vellum/tasks"
	"example.com/vellum/ui"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	exit := flag.Duration("exit-after", 0, "close after duration for smoke checks")
	smoke := flag.Bool("smoke", false, "exercise registered actions and worker completion")
	flag.Parse()
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	count := 0
	status := "Ready"
	var task *tasks.Task
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
	bar := &cogent.CommandBar{Owner: a, Registry: registry, OnError: func(err error) { status = err.Error() }}
	must(a.Open(app.Window{ID: "main", Title: "Vellum Workbench", Width: 1100, Height: 760, View: func() ui.Node {
		progress := ""
		if task != nil {
			s := task.Snapshot()
			if s.Running {
				progress = fmt.Sprintf(" · %.0f%% %s", s.Progress*100, s.Message)
			}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "title", Kind: ui.Label, Text: "Vellum · Go native workbench"}, {ID: "commands", Kind: ui.Extension, Custom: bar}, {ID: "counter", Kind: ui.Label, Text: fmt.Sprintf("Accepted actions: %d", count)}, {ID: "status", Kind: ui.Label, Text: status + progress}}}
	}}))
	if *smoke {
		a.Post(func() {
			must(registry.Invoke(a.Context(), "project.increment"))
			must(registry.Invoke(a.Context(), "analysis.calculate"))
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
