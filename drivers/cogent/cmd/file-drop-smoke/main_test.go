package main

import (
	"errors"
	"testing"
	"time"
)

func TestOutcomePreservesDeliveredDiagnostic(t *testing.T) {
	for _, diagnostic := range []error{nil, errors.New("native geometry missing")} {
		outcome := make(chan error, 1)
		outcome <- diagnostic
		if got := awaitOutcome(outcome, 0); got != diagnostic {
			t.Fatalf("delivered diagnostic replaced: got %v, want %v", got, diagnostic)
		}
	}
}

func TestOutcomeWaitsForFinishingWorker(t *testing.T) {
	outcome := make(chan error)
	diagnostic := errors.New("worker cancelled during inspection")
	go func() { outcome <- diagnostic }()
	if got := awaitOutcome(outcome, time.Second); got != diagnostic {
		t.Fatalf("finishing worker diagnostic lost: %v", got)
	}
}

func TestOutcomeRefusesStuckWorkerAtGraceDeadline(t *testing.T) {
	if got := awaitOutcome(make(chan error), 0); !errors.Is(got, errWorkerOutcome) {
		t.Fatalf("stuck worker was not bounded: %v", got)
	}
}
