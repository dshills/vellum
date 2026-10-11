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
	flag.Parse()
	if e := run(*nodes, *smoke); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
