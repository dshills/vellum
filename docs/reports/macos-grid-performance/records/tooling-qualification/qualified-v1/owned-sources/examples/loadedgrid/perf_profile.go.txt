package main

import (
	"errors"
	"io"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
)

const perfProfileMaxBytes = 64 << 20

var errPerfProfileBudget = errors.New("loadedgrid perf: CPU profile byte budget exceeded")

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
