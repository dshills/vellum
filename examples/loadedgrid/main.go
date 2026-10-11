// Loadedgrid uses actual stored immutable records and bounded worker projections.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"os"
	"time"
)

func run(rows int, variable, smoke bool) error {
	p, e := newProgram(rows, variable, nil)
	if e != nil {
		return e
	}
	defer p.owner.Close()
	for _, w := range p.windows {
		w.settle = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleGridEdit(ctx, p.owner, w.id, w.grid, cb)
		}
		w.settleFilter = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleTextField(ctx, p.owner, w.id, "filter-text", cb)
		}
	}
	failures := make(chan error, 1)
	driver := cogent.Driver{OnError: func(e error) {
		select {
		case failures <- e:
		default:
		}
	}}
	result := make(chan error, 1)
	if smoke {
		go func() {
			ctx, cancel := context.WithTimeout(p.owner.Context(), 90*time.Second)
			defer cancel()
			result <- desktopSmoke(ctx, p)
			p.owner.Close()
		}()
	}
	native := p.owner.Run(context.Background(), driver)
	var worker error
	if smoke {
		select {
		case worker = <-result:
		case <-time.After(2 * time.Second):
			worker = errors.New("loadedgrid: native runtime ended before smoke completion")
		}
	}
	select {
	case e = <-failures:
		native = errors.Join(native, e)
	default:
	}
	return errors.Join(native, worker)
}
func main() {
	rows := flag.Int("rows", maxRows, "stored record count (1..1000000); first window shows at most10000")
	variable := flag.Bool("variable-rows", false, "preserve indexed stable-key variable heights through projections")
	smoke := flag.Bool("smoke", false, "bounded Cocoa model-command/settlement smoke, no synthesized desktop input")
	perf := flag.Bool("perf", false, "opt-in release-only stored-table native scroll measurements")
	perfSamples := flag.Int("perf-samples", 120, "measured native scroll frames (120..2048)")
	perfWarmup := flag.Int("perf-warmup", 128, "unmeasured closed-trajectory frames (multiple64)")
	perfNovel := flag.Bool("perf-novel", false, "separate newly visited row strings; not cold-font or warm P6 results")
	perfProfile := flag.String("perf-cpu-profile", "", "diagnostic process-wide CPU profile to a new file (64MiB cap); excluded from warm qualification")
	flag.Parse()
	if !*perf && *perfProfile != "" {
		fmt.Fprintln(os.Stderr, "loadedgrid: -perf-cpu-profile requires -perf")
		os.Exit(1)
	}
	if *perf {
		if *smoke {
			fmt.Fprintln(os.Stderr, "loadedgrid: -perf and -smoke are separate workflows")
			os.Exit(1)
		}
		if err := runPerf(perfConfig{Rows: *rows, Variable: *variable, Samples: *perfSamples, Warmup: *perfWarmup, Novel: *perfNovel, CPUProfile: *perfProfile}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if e := run(*rows, *variable, *smoke); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
