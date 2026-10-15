//go:build offscreen

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dshills/vellum/examples/replay/workspace"
	"github.com/dshills/vellum/replay"
)

func run(rows, batch int) (err error) {
	record, err := workspace.Record(rows)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	b, err := workspace.NewBackend(ctx, rows, batch)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if e := b.Close(cleanup); err == nil {
			err = e
		}
	}()
	runner, err := replay.NewRunner(replay.RunnerConfig{EventBatch: batch, MaxSteps: 100000, MaxCheckpointBytes: 32 << 20}, b)
	if err != nil {
		return err
	}
	result, err := runner.Run(ctx, record)
	if err != nil {
		return err
	}
	if !result.QuiescentNow || result.Applied != len(record.Operations) {
		return fmt.Errorf("workspace replay incomplete: applied %d; %+v", result.Applied, result.Residual)
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	rows := flag.Int("rows", 10000, "stored rows: 10000 or 1000000")
	batch := flag.Int("batch", 64, "event and owner dispatch batch: 1..64")
	flag.Parse()
	if err := run(*rows, *batch); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
