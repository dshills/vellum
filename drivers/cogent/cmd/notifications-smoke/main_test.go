//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestOutcomeKeepsCompletedWorkerAndEarlyExitReports(t *testing.T) {
	t.Run("completed", func(t *testing.T) {
		finished := make(chan outcome, 1)
		finished <- outcome{report: map[string]any{"result": "PASS", "owned_cleanup_verified": true}}
		stopped := false
		result := awaitOutcome(finished, func() { stopped = true }, make(chan time.Time), nil)
		if stopped || result.err != nil || result.report["result"] != "PASS" || result.report["owned_cleanup_verified"] != true {
			t.Fatalf("completed outcome changed: %+v", result)
		}
	})
	t.Run("cleanup-reply-after-stop", func(t *testing.T) {
		finished := make(chan outcome, 1)
		workerError := errors.New("worker failure")
		result := awaitOutcome(finished, func() {
			finished <- outcome{report: map[string]any{"worker_detail": "retained", "owned_cleanup_verified": true, "delivery_qualified": true}, err: workerError}
		}, make(chan time.Time), nil)
		if !errors.Is(result.err, workerError) || result.report["worker_detail"] != "retained" || result.report["owned_cleanup_verified"] != false || result.report["delivery_qualified"] != false || result.report["native_loop_ended_early"] != true {
			t.Fatalf("early-loop outcome lost/overclaimed: %+v", result)
		}
	})
	t.Run("no-worker-reply", func(t *testing.T) {
		deadline := make(chan time.Time, 1)
		deadline <- time.Time{}
		stopped := false
		result := awaitOutcome(make(chan outcome), func() { stopped = true }, deadline, map[string]any{"nonce": strings.Repeat("a", 64)})
		if !stopped || result.err == nil || result.report["owned_cleanup_verified"] != false || result.report["result"] != "FAILED_NATIVE_NOTIFICATION_EARLY_LOOP_EXIT" {
			t.Fatalf("timeout did not retain an explicit failed report: %+v", result)
		}
		wire, err := json.Marshal(result.report)
		if err != nil || !strings.Contains(string(wire), `"nonce"`) || !strings.Contains(string(wire), `"owned_cleanup_verified":false`) {
			t.Fatal("failed outcome is not serializable with its scoped identity")
		}
	})
}
