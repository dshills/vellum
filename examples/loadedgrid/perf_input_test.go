package main

import (
	"encoding/json"
	"errors"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func TestPerfWarmupRefusesOffsetsExtraCallbacksAndLayoutDrift(t *testing.T) {
	initial := cogent.GridInspection{Viewport: virtual.GridViewport{Width: 768, Height: 624}, Bounds: image.Rect(0, 0, 1296, 1053), Scale: 1.6875}
	if err := checkPerfWarmup(initial, initial, 128, 128); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*cogent.GridInspection){func(g *cogent.GridInspection) { g.Viewport.ScrollY = 31 }, func(g *cogent.GridInspection) { g.Viewport.ScrollX = 1 }, func(g *cogent.GridInspection) { g.Scale = 2 }, func(g *cogent.GridInspection) { g.Bounds.Max.Y++ }, func(g *cogent.GridInspection) { g.Viewport.Height++ }} {
		final := initial
		mutate(&final)
		if checkPerfWarmup(initial, final, 128, 128) == nil {
			t.Fatal("accepted changed warm workload")
		}
	}
	if checkPerfWarmup(initial, initial, 129, 128) == nil {
		t.Fatal("accepted extra transition")
	}
	initial.Viewport.ScrollY = 31
	if checkPerfWarmup(initial, initial, 128, 128) == nil {
		t.Fatal("accepted noncanonical starting offset")
	}
}

func TestPerfInputLedgerExclusiveFailureEvidence(t *testing.T) {
	m, _ := cogent.NewFrameMetricsWithGridInputDetails(130)
	path := filepath.Join(t.TempDir(), "input.json")
	finish, err := startPerfInputDetails(path, m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = startPerfInputDetails(path, m); err == nil {
		t.Fatal("overwrote existing evidence")
	}
	if err = finish(errors.New("warmup refused")); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err = finish(nil); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("repeated finish altered failed evidence")
	}
	var saved struct {
		Schema, Result string
		Details        cogent.GridInputDetails
	}
	if err = json.Unmarshal(before, &saved); err != nil || saved.Schema != "vellum-grid-input-v1" || saved.Result != "FAILED_RUN_DIAGNOSTIC" || saved.Details.Capacity != 646 {
		t.Fatal(err, saved)
	}
	c := perfConfig{Rows: 10000, Samples: 120, Warmup: 128, InputDetails: path}
	if err = c.validate(); err != nil {
		t.Fatal(err)
	}
	if _, err = newPerfFrameMetrics(c); err != nil {
		t.Fatal(err)
	}
	c.SourceDetails = true
	if c.validate() == nil {
		t.Fatal("combined diagnostic modes accepted")
	}
}
