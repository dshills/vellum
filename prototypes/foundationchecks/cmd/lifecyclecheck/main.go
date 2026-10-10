//go:build darwin && cgo

// lifecyclecheck exercises adapter attachment/teardown on owned GLFW views.
// It does not establish a supported Cogent before-show hook or GPU lifetime.
package main

import (
	"encoding/json"
	"example.com/vellum/nativeinput/inputbridge"
	"example.com/vellum/nativeinput/macbridge"
	"example.com/vellum/nativeinput/semantics"
	"flag"
	"github.com/go-gl/glfw/v3.4/glfw"
	"os"
	"runtime"
	"time"
	"unsafe"
)

func measurement(cycle int) map[string]any {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return map[string]any{"cycle": cycle, "heap_alloc": m.HeapAlloc, "resident": macbridge.ResidentBytes(), "goroutines": runtime.NumGoroutine()}
}
func main() {
	n := flag.Int("cycles", 1000, "adapter attach/detach cycles")
	fresh := flag.Bool("fresh-windows", false, "destroy and recreate owned hidden views each cycle")
	pooled := flag.Bool("pooled", true, "drain a Cocoa autorelease pool each cycle; false is a harness control")
	flag.Parse()
	if *n < 1 {
		panic("cycles must be positive")
	}
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	makeWindow := func() *glfw.Window {
		w, err := glfw.CreateWindow(400, 250, "Vellum foundation lifecycle check", nil, nil)
		if err != nil {
			panic(err)
		}
		return w
	}
	w := makeWindow()
	var samples []map[string]any
	samples = append(samples, measurement(0))
	start := time.Now()
	callbacks := 0
	queries := 0
	snap := semantics.Snapshot{Focus: 2, Nodes: []semantics.Node{{ID: 1, Role: "window", Label: "Lifecycle fixture", Children: []uint64{2}, Bounds: [4]float64{0, 0, 400, 250}}, {ID: 2, Role: "text", Label: "Field", Value: "A👩Z", Actions: []string{"focus", "set_value"}, Bounds: [4]float64{10, 10, 200, 35}}}}
	if err := snap.Validate(); err != nil {
		panic(err)
	}
	for cycle := 1; cycle <= *n; cycle++ {
		var pool unsafe.Pointer
		if *pooled {
			pool = poolBegin()
		}
		ax := macbridge.Attach(w.GetCocoaView(), w.GetCocoaWindow(), snap, func(semantics.Action) { callbacks++ })
		in := inputbridge.Attach(w.GetCocoaView(), func(inputbridge.Update) { callbacks++ })
		if in == nil {
			panic("input attachment")
		}
		in.Update("A👩Z", 1, 1, -1, -1, true, 0, [4]float64{10, 10, 1, 23})
		for _, step := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
			var p map[string]any
			if err := json.Unmarshal([]byte(in.Probe(step)), &p); err != nil {
				panic(err)
			}
			queries++
			if step == 9 && p["text"] != "A中文漢👩Z" {
				panic("native state lost")
			}
		}
		in.Close()
		in.Close()
		ax.Close()
		glfw.PollEvents()
		if *fresh {
			w.Destroy()
			if cycle < *n {
				w = makeWindow()
			}
		}
		if *pooled {
			poolEnd(pool)
		}
		if cycle%100 == 0 || cycle == *n {
			samples = append(samples, measurement(cycle))
		}
	}
	if !*fresh {
		w.Destroy()
	}
	samples = append(samples, measurement(*n))
	if callbacks != *n*10 || queries != *n*9 {
		panic("callback/query count mismatch")
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{"cycles": *n, "autorelease_pools": *pooled, "fresh_windows": *fresh, "callbacks": callbacks, "protocol_queries": queries, "seconds": time.Since(start).Seconds(), "samples": samples, "scope": "owned hidden GLFW view native adapter lifetime; no OS settings changed; not Cogent/GPU lifecycle or a live IME/VoiceOver test"}); err != nil {
		panic(err)
	}
}
