package main

import (
	"cogentcore.org/core/text/lines"
	"context"
	"encoding/json"
	"example.com/vellum/foundationcontracts/mailbox"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

func main() {
	flag.Bool("nogui", true, "offscreen diagnostic")
	count := flag.Int("messages", 100000, "ordered worker messages")
	flag.Parse()
	if *count < 1 {
		panic("positive messages required")
	}
	var textResults []map[string]any
	for _, text := range []string{"", "one\ntwo", "one\ntwo\n", "one\ntwo\n\n", "\n", "one\n\ntwo", "日本語\n👩‍💻", "a\r\nb\r\n"} {
		ls := lines.NewLinesExact().SetString(text)
		got := ls.String()
		ls.Close()
		textResults = append(textResults, map[string]any{"input": text, "observed": got, "pass": got == text})
		if got != text {
			panic("exact roundtrip")
		}
	}
	const slots = 128
	limits := mailbox.Limits{Slots: slots, PayloadBytes: 4096, MessageBytes: 32}
	q, err := mailbox.New(limits)
	if err != nil {
		panic(err)
	}
	// Fill a stalled-consumer control, proving preflight refuses capacity before
	// hypothetical external mutation and inflight messages still consume credit.
	var rejected int
	for i := 0; i < slots; i++ {
		if err = q.Submit(context.Background(), []byte(fmt.Sprintf("%08d", i))); err != nil {
			panic(err)
		}
	}
	for range 10000 {
		if _, err = q.TryReserve(8); err != mailbox.ErrFull {
			panic("admission boundary")
		}
		rejected++
	}
	stalled := q.Stats()
	for i := 0; i < slots; i++ {
		d, ok := q.TryNext()
		if !ok || string(d.Data) != fmt.Sprintf("%08d", i) {
			panic("control order")
		}
		if err = q.Ack(d.ID); err != nil {
			panic(err)
		}
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	result := make(chan error, 1)
	go func() {
		for i := 0; i < *count; i++ {
			if err := q.Submit(context.Background(), []byte(fmt.Sprintf("%08d", i))); err != nil {
				result <- err
				return
			}
		}
		result <- nil
	}()
	var peak mailbox.Stats
	var id uint64
	for i := 0; i < *count; i++ {
		d, err := q.Next(context.Background())
		if err != nil || string(d.Data) != fmt.Sprintf("%08d", i) || d.ID <= id {
			panic("loss/order failure")
		}
		id = d.ID
		s := q.Stats()
		peak.InFlight = max(peak.InFlight, s.InFlight)
		peak.Outstanding = max(peak.Outstanding, s.Outstanding)
		peak.PayloadBytes = max(peak.PayloadBytes, s.PayloadBytes)
		if err = q.Ack(d.ID); err != nil {
			panic(err)
		}
	}
	if err = <-result; err != nil {
		panic(err)
	}
	q.Close()
	runtime.GC()
	runtime.ReadMemStats(&after)
	json.NewEncoder(os.Stdout).Encode(map[string]any{"scope": "exact opt-in multiline export and standalone reserved mailbox; no native preflight/event-pump integration or workbench expansion", "text_results": textResults, "limits": limits, "stalled_control": stalled, "preflight_rejections": rejected, "delivered_fifo": *count, "observed_peak": peak, "after_drain": q.Stats(), "seconds": time.Since(start).Seconds(), "heap_delta_after_gc": int64(after.HeapAlloc) - int64(before.HeapAlloc), "total_alloc_bytes": after.TotalAlloc - before.TotalAlloc, "accepted_loss": 0})
}
