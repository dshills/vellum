package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dshills/vellum/ui"
)

func TestProgressSmokePreservesJoinedDriverFailures(t *testing.T) {
	if progressRunError(nil) != nil || progressRunError(context.Canceled) != nil {
		t.Fatal("successful stop not normalized")
	}
	failure := errors.New("driver failed")
	for _, err := range []error{failure, errors.Join(context.Canceled, failure), fmt.Errorf("wrapped driver: %w", failure), fmt.Errorf("wrapped cancellation: %w", context.Canceled)} {
		if progressRunError(err) != err {
			t.Fatal("driver error changed or discarded", err)
		}
	}
}

func TestProgressSmokeViewUsesBothExportedKinds(t *testing.T) {
	g := newGallery()
	for _, value := range []float32{0, 25, 75, 100} {
		g.number = value
		n := g.progressView()
		if err := ui.Validate(n); err != nil {
			t.Fatal(err)
		}
		shapes := n.Children[1].Children
		if len(shapes) != 2 || shapes[0].Kind != ui.ProgressRing || shapes[1].Kind != ui.ProgressGauge || shapes[0].Number != value/100 || shapes[1].Number != value/100 {
			t.Fatal("native fixture descriptors drifted", shapes)
		}
	}
}
