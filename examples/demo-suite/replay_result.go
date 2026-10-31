package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func replayResult(dir string, out []byte) (string, error) {
	var record struct {
		Result struct {
			Applied      int
			QuiescentNow bool
			Checkpoints  []struct{ Label string }
		}
	}
	if err := json.Unmarshal(out, &record); err != nil {
		return "", fmt.Errorf("decode replay result: %w", err)
	}
	report := filepath.Join(dir, "replay-last.json")
	if err := os.WriteFile(report, out, 0600); err != nil {
		return "", fmt.Errorf("write replay report: %w", err)
	}
	labels := make([]string, 0, len(record.Result.Checkpoints))
	for _, checkpoint := range record.Result.Checkpoints {
		labels = append(labels, checkpoint.Label)
	}
	return fmt.Sprintf("Replay completed: %d operations · %d checkpoints · quiescent: %t\n%s\nFull result: %s", record.Result.Applied, len(labels), record.Result.QuiescentNow, strings.Join(labels, ", "), report), nil
}

func replayFailure(err error, out []byte) error {
	detail := []rune(string(out))
	if len(detail) > 1024 {
		detail = append(detail[:1024], '…')
	}
	return fmt.Errorf("replay run failed: %w: %s", err, string(detail))
}
