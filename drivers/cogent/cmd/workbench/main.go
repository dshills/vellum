package main

import (
	"context"
	"example.com/vellum/app"
	"example.com/vellum/cogent"
	"example.com/vellum/ui"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	exit := flag.Duration("exit-after", 0, "close after duration for smoke checks")
	flag.Parse()
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	count := 0
	must(a.Open(app.Window{ID: "main", Title: "Vellum Workbench", Width: 1100, Height: 760, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "title", Kind: ui.Label, Text: "Vellum · Go native workbench"}, {ID: "counter", Kind: ui.Label, Text: fmt.Sprintf("Accepted actions: %d", count)}, {ID: "increment", Kind: ui.Button, Text: "Increment", Activate: func() { count++ }}, {ID: "close", Kind: ui.Button, Text: "Close", Activate: a.Close}}}
	}}))
	if *exit > 0 {
		time.AfterFunc(*exit, a.Close)
	}
	must(a.Run(context.Background(), cogent.Driver{}))
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
