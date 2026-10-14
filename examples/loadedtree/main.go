package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dshills/vellum/drivers/cogent"
)

func run(nodes int, smoke bool) error {
	p, e := newProgram(nodes, nil)
	if e != nil {
		return e
	}
	defer p.owner.Close()
	for _, w := range p.windows {
		w.settleFilter = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleTextField(ctx, p.owner, w.id, "filter", cb)
		}
	}
	failures := make(chan error, 1)
	result := make(chan error, 1)
	driver := cogent.Driver{OnError: func(e error) {
		select {
		case failures <- e:
		default:
		}
	}}
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
			worker = errors.New("loadedtree: native runtime ended before smoke completion")
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
	nodes := flag.Int("nodes", maxNodes, "stored nodes (1..100000); first window at most10000")
	smoke := flag.Bool("smoke", false, "bounded Cocoa owner-command geometry/settlement smoke")
	perf := flag.Bool("perf", false, "routed completed-frame warm tree performance")
	samples := flag.Int("perf-samples", 240, "measured native scroll samples")
	warmup := flag.Int("perf-warmup", 128, "unmeasured complete warm trajectory events")
	deep := flag.Bool("perf-deep", false, "warm deep-chain viewport instead of wide branch")
	gpuScene := flag.Bool("perf-gpu-scene", false, "opt into GPU scene rendering for this performance run")
	sourceDetails := flag.Bool("perf-source-details", false, "additional diagnostic source clocks/counters; exclude from ordinary timing")
	flag.Parse()
	if *perf && *smoke || !*perf && (*deep || *gpuScene || *sourceDetails || *samples != 240 || *warmup != 128) {
		fmt.Fprintln(os.Stderr, "loadedtree: perf flags require -perf and exclude -smoke")
		os.Exit(1)
	}
	if *perf {
		if e := runTreePerf(treePerfConfig{Nodes: *nodes, Samples: *samples, Warmup: *warmup, Deep: *deep, GPUSceneRendering: *gpuScene, SourceDetails: *sourceDetails}); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		return
	}

	if e := run(*nodes, *smoke); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
