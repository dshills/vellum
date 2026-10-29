package cogent

import (
	"github.com/dshills/vellum/app"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

// OpenFilesEvent is a detached bounded Finder request. Initial means it was
// captured before Driver.Run bound any native window (no native input existed).
// Applications must validate contents and resolve dirty/recovery decisions.
type OpenFilesEvent struct {
	Paths   []string
	Initial bool
}
type openFilesReceiver struct {
	mu                     sync.Mutex
	owner                  *app.Application
	callback               func(OpenFilesEvent)
	queue                  []OpenFilesEvent
	count                  int
	pending, ready, closed bool
}

func (r *openFilesReceiver) enqueue(event OpenFilesEvent) bool {
	if len(event.Paths) == 0 || len(event.Paths) > 16 {
		return false
	}
	paths := append([]string(nil), event.Paths...)
	total := 0
	for _, p := range paths {
		total += len(p)
		if !filepath.IsAbs(p) || len(p) > 4096 || !utf8.ValidString(p) || strings.ContainsRune(p, 0) {
			return false
		}
	}
	if total > 64<<10 {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.count+len(paths) > 16 {
		return false
	}
	r.queue = append(r.queue, OpenFilesEvent{paths, event.Initial})
	r.count += len(paths)
	r.startLocked()
	return true
}
func (r *openFilesReceiver) startLocked() {
	if r.pending || len(r.queue) == 0 || !r.ready && !r.queue[0].Initial {
		return
	}
	r.pending = true
	go r.drain()
}
func (r *openFilesReceiver) drain() {
	for {
		delivered := false
		err := r.owner.Do(func() {
			r.mu.Lock()
			if r.closed || len(r.queue) == 0 || !r.ready && !r.queue[0].Initial {
				r.pending = false
				r.mu.Unlock()
				return
			}
			event := r.queue[0]
			r.queue[0] = OpenFilesEvent{}
			r.queue = r.queue[1:]
			r.count -= len(event.Paths)
			r.mu.Unlock()
			delivered = true
			r.callback(event)
		})
		if err != nil {
			r.close()
			return
		}
		if !delivered {
			return
		}
	}
}
func (r *openFilesReceiver) markReady() { r.mu.Lock(); r.ready = true; r.startLocked(); r.mu.Unlock() }
func (r *openFilesReceiver) close() {
	r.mu.Lock()
	r.closed = true
	r.queue = nil
	r.count = 0
	r.mu.Unlock()
}
