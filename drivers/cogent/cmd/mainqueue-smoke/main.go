// mainqueue-smoke exercises the real desktop event loop once per process.
// It opens no windows and does not change OS preferences or clipboard contents.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	_ "cogentcore.org/core/system/driver"
	"cogentcore.org/core/system/driver/base"
	"cogentcore.org/core/system/driver/desktop"
)

type outcome struct {
	returned bool
	value    any
}

func observe(f func()) (out outcome) {
	defer func() { out.value = recover() }()
	f()
	out.returned = true
	return
}

func main() {
	mode := flag.String("case", "normal", "native loop case: normal, panic, idle or stop")
	flag.Parse()
	if err := run(*mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(mode string) error {
	if mode != "normal" && mode != "panic" && mode != "idle" && mode != "stop" {
		return fmt.Errorf("unknown case %q", mode)
	}
	watchdog := time.AfterFunc(10*time.Second, func() { fmt.Fprintln(os.Stderr, "FAIL native queue watchdog"); os.Exit(1) })
	defer watchdog.Stop()
	a := desktop.TheApp
	a.InitMainQueue()
	sentinel := errors.New("native callback panic fixture")
	caller := make(chan outcome, 1)
	completed := 0 // Main writes; the acknowledgement publishes it to the caller.
	count, floodWorkers := 100, 4
	switch mode {
	case "panic":
		count = 1
	case "idle":
		count, floodWorkers = 3, 0
	case "stop":
		count, floodWorkers = 0, 0
	}
	var wakes sync.WaitGroup
	for i := 0; i < floodWorkers; i++ {
		wakes.Add(1)
		go func() {
			defer wakes.Done()
			for j := 0; j < 2500; j++ {
				a.SendEmptyEvent()
			}
		}()
	}
	go func() {
		caller <- observe(func() {
			if mode == "stop" {
				// No dispatch or wake flood precedes this worker-only stop.
				time.Sleep(250 * time.Millisecond)
			}
			for i := 0; i < count; i++ {
				if mode == "idle" {
					// Let the native loop enter its indefinite wait between
					// calls. This is liveness evidence, not a CPU benchmark.
					time.Sleep(250 * time.Millisecond)
				}
				if !a.RunOnMainIfOpen(func() {
					completed++
					if mode == "panic" {
						panic(sentinel)
					}
				}) {
					panic("native callback unexpectedly refused")
				}
				if completed != i+1 {
					panic("callback returned before output acknowledgement")
				}
			}
			a.StopMain()
		})
	}()
	loop := observe(a.MainLoop)
	var worker outcome
	select {
	case worker = <-caller:
	case <-time.After(3 * time.Second):
		return errors.New("native failure stranded caller")
	}
	if mode == "panic" {
		if loop.returned || worker.returned || loop.value != sentinel || worker.value != sentinel || completed != 1 {
			return fmt.Errorf("panic identity/completion changed: loop=%+v caller=%+v completed=%d", loop, worker, completed)
		}
	} else if !loop.returned || !worker.returned || completed != count {
		return fmt.Errorf("normal completion changed: loop=%+v caller=%+v completed=%d", loop, worker, completed)
	}
	done := make(chan struct{})
	go func() { wakes.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		return errors.New("event wakes survived native termination")
	}
	// Probe public desktop dispatch and creation after actual GLFW termination.
	for i := 0; i < 1000; i++ {
		a.SendEmptyEvent()
	}
	if !a.MainStopped() || a.RunOnMainIfOpen(func() { panic("terminal callback executed") }) {
		return errors.New("terminated loop still admitted callbacks")
	}
	window, err := a.NewWindow(nil)
	if window != nil || !errors.Is(err, base.ErrMainStopped) {
		return fmt.Errorf("terminal creation did not refuse: window=%v error=%v", window, err)
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Case                   string `json:"case"`
		Completed              int    `json:"callbacks_started"`
		WakeRequests           int    `json:"wake_requests"`
		QueueStopped           bool   `json:"queue_stopped"`
		CreationRefused        bool   `json:"native_creation_refused"`
		PanicIdentityPreserved bool   `json:"panic_identity_preserved"`
	}{mode, completed, floodWorkers*2500 + 1000, true, true, mode == "panic"})
}
