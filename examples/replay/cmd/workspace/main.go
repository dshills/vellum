// Workspace is an independent consumer of Vellum's public UI APIs.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/replay/workspace"
	"github.com/dshills/vellum/execution"
)

func run(rows int) error {
	a, err := app.New(app.Config{QueueCapacity: 128, CompletionCapacity: 8})
	if err != nil {
		return err
	}
	defer a.Close()
	ex := &execution.Goroutine{}
	defer ex.Close()
	f, err := workspace.New(a, ex, rows)
	if err != nil {
		return err
	}
	defer f.Palette.Close()
	nativeErrors := make(chan error, 1)
	runErr := a.Run(context.Background(), cogent.Driver{Commands: f.Commands, OnError: func(err error) {
		select {
		case nativeErrors <- err:
		default:
		}
	}})
	a.Close()
	// All windows have retired; joining the sole task is explicit and bounded.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var nativeErr error
	select {
	case nativeErr = <-nativeErrors:
	default:
	}
	return errors.Join(runErr, nativeErr, f.WaitWorker(ctx))
}
func main() {
	rows := flag.Int("rows", 10000, "stored rows: 10000 or 1000000")
	flag.Parse()
	if err := run(*rows); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
