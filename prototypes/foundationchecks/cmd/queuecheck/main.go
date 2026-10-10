// queuecheck probes the actual nativeinput event transport with a stalled consumer.
package main

import (
	"cogentcore.org/core/events"
	"encoding/json"
	"example.com/vellum/nativeinput/eventqueue"
	"flag"
	"os"
	"runtime"
)

func main() {
	flag.Bool("nogui", true, "offscreen initialization")
	n := flag.Int("events", 100000, "unique actions before consumer resumes")
	flag.Parse()
	if *n < 1 {
		panic("positive events required")
	}
	source := &events.Source{}
	runtime.GC()
	var before, queued runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < *n; i++ {
		eventqueue.Send(source, i)
	}
	runtime.GC()
	runtime.ReadMemStats(&queued)
	depth := len(source.Deque.Back)
	for i := 0; i < *n; i++ {
		ev, ok := source.Deque.PollEvent()
		if !ok || ev.(*events.CustomEvent).Data.(int) != i {
			panic("action loss/reordering")
		}
	}
	_, extra := source.Deque.PollEvent()
	if extra {
		panic("unexpected event")
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"events": *n, "queue_depth_before_drain": depth, "heap_delta_bytes": int64(queued.HeapAlloc) - int64(before.HeapAlloc), "fifo_pass": true, "bounded_admission": false, "scope": "actual nativeinput Send path, stalled consumer, integer payloads; no native event delivery or queue-limit implementation"})
}
