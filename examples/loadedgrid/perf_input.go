package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/dshills/vellum/drivers/cogent"
)

// Write only after Run has retired native dispatch, including failed warmups.
// O_EXCL preserves prior evidence and refuses existing files and symlinks.
func startPerfInputDetails(path string, metrics *cogent.FrameMetrics) (func(error) error, error) {
	if path == "" {
		return func(error) error { return nil }, nil
	}
	if !validPerfProfilePath(path) {
		return nil, errors.New("loadedgrid perf: invalid input details path")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	var once sync.Once
	var terminal error
	return func(runError error) error {
		once.Do(func() {
			result := "CAPTURED_DIAGNOSTIC"
			if runError != nil {
				result = "FAILED_RUN_DIAGNOSTIC"
			}
			details := metrics.GridInputDetails()
			var detailError error
			if details.Failure != "" {
				result = "REFUSED_INPUT_DETAILS"
				detailError = errors.New(details.Failure)
			}
			w := &perfProfileWriter{target: f, remaining: perfProfileMaxBytes}
			err := json.NewEncoder(w).Encode(struct {
				Schema, Result string
				Details        cogent.GridInputDetails
			}{"vellum-grid-input-v1", result, details})
			terminal = errors.Join(detailError, err, w.failure(), f.Close())
		})
		return terminal
	}, nil
}
