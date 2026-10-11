//go:build offscreen

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	consumer "github.com/dshills/vellum/examples/replay"
)

func main() {
	batch := flag.Int("batch", 64, "combined owner/event dispatch batch (1..64)")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	proof, err := consumer.RunNative(ctx, *batch)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = json.NewEncoder(os.Stdout).Encode(proof); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
