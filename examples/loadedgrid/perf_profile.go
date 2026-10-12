package main

import (
	"errors"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
)

const perfProfileMaxBytes = 64 << 20

var errPerfProfileBudget = errors.New("loadedgrid perf: profile byte budget exceeded")

// The profiler is diagnostic and process-wide. A profiled process is never a
// warm-budget qualification run, and the ordinary runner never enables it.
type perfProfileWriter struct {
	mu        sync.Mutex
	target    io.Writer
	remaining int
	err       error
}

func (w *perfProfileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return 0, w.err
	}
	if len(p) > w.remaining {
		w.err = errPerfProfileBudget
		return 0, w.err
	}
	n, err := w.target.Write(p)
	w.remaining -= n
	if n != len(p) && err == nil {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}
func (w *perfProfileWriter) failure() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.err
}
func validPerfProfilePath(path string) bool {
	return len(path) <= 4096 && !strings.ContainsRune(path, 0)
}
func startPerfCPUProfile(path string) (func() error, error) {
	if path == "" {
		return func() error { return nil }, nil
	}
	if !validPerfProfilePath(path) {
		return nil, errors.New("loadedgrid perf: invalid CPU profile path")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	w := &perfProfileWriter{target: f, remaining: perfProfileMaxBytes}
	if err = pprof.StartCPUProfile(w); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	var once sync.Once
	var terminal error
	return func() error {
		once.Do(func() {
			// Stop waits for the profiler's writer before observing its error.
			pprof.StopCPUProfile()
			terminal = errors.Join(w.failure(), f.Close())
		})
		return terminal
	}, nil
}

func validPerfAllocationPrefix(prefix string) bool {
	return validPerfProfilePath(prefix+".before.pprof") && validPerfProfilePath(prefix+".after.pprof")
}

// Differential cumulative allocs profiles exclude startup and warmup. Runtime
// heap accounting can lag two GC cycles, so both snapshots force two collections
// outside the measured events. MemProfileRate stays at its process default.
// Samples still include observers and profile encoding; never qualify latency
// or native/GPU memory from these diagnostic processes.
func startPerfAllocationProfile(prefix string) (func() error, error) {
	return startPerfAllocationProfileWithCapture(prefix, func(w io.Writer) error {
		runtime.GC()
		runtime.GC()
		return pprof.Lookup("allocs").WriteTo(w, 0)
	})
}

func startPerfAllocationProfileWithCapture(prefix string, capture func(io.Writer) error) (func() error, error) {
	if prefix == "" {
		return func() error { return nil }, nil
	}
	if !validPerfAllocationPrefix(prefix) {
		return nil, errors.New("loadedgrid perf: invalid allocation profile prefix")
	}
	before, err := os.OpenFile(prefix+".before.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	after, err := os.OpenFile(prefix+".after.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.Join(err, before.Close())
	}
	w := &perfProfileWriter{target: before, remaining: perfProfileMaxBytes}
	err = errors.Join(capture(w), w.failure(), before.Close())
	if err != nil {
		return nil, errors.Join(err, after.Close())
	}
	var once sync.Once
	var terminal error
	return func() error {
		once.Do(func() {
			w := &perfProfileWriter{target: after, remaining: perfProfileMaxBytes}
			terminal = errors.Join(capture(w), w.failure(), after.Close())
		})
		return terminal
	}, nil
}
