// Grid is an independent consumer of Vellum's managed fixed-height data grid.
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

func run(smoke bool) error {
	program, err := newProgram()
	if err != nil {
		return err
	}
	defer program.owner.Close()
	// OnError runs under the owner: recording it must not reenter the owner.
	nativeErrors := make(chan error, 1)
	driver := cogent.Driver{OnError: func(err error) {
		select {
		case nativeErrors <- err:
		default: // Retain the first failure; later errors cannot hide it.
		}
	}}
	result := make(chan error, 1)
	if smoke {
		go func() {
			ctx, cancel := context.WithTimeout(program.owner.Context(), 45*time.Second)
			defer cancel()
			result <- desktopSmoke(ctx, program)
			program.owner.Close()
		}()
	}
	runErr := program.owner.Run(context.Background(), driver)
	var workerErr error
	if smoke {
		select {
		case workerErr = <-result:
		case <-time.After(2 * time.Second):
			workerErr = errors.New("grid smoke: native runtime ended before worker completion")
		}
	}
	select {
	case nativeErr := <-nativeErrors:
		runErr = errors.Join(runErr, nativeErr)
	default:
	}
	return errors.Join(runErr, workerErr)
}

func main() {
	smoke := flag.Bool("smoke", false, "check desktop measured geometry, model transitions and two-window lifetimes")
	flag.Parse()
	if err := run(*smoke); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
