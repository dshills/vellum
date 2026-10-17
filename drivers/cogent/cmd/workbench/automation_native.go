// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"github.com/dshills/vellum/drivers/cogent"
)

func validateAutomationCheck(enabled, api bool, exit time.Duration, conflicting bool) error {
	if !enabled {
		return nil
	}
	if !api || exit < 5*time.Second || exit > 30*time.Second || conflicting {
		return errors.New("automation-check requires automation and exit-after=5s..30s without other smoke/endurance modes")
	}
	return nil
}

// This explicit fixture waits for a real native window and a completed public
// GPU snapshot. It does not assert presentation, pixel contents or frame budgets.
func beginAutomationNativeCheck(a *app.Application, gpu *cogent.GPUCanvas) <-chan error {
	result := make(chan error, 1)
	go func() {
		defer close(result)
		ctx, cancel := context.WithTimeout(a.Context(), 10*time.Second)
		defer cancel()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		var err error
		for {
			err = enduranceOwner(ctx, a, func() error {
				image, err := gpu.Snapshot()
				if err != nil {
					return err
				}
				if image == nil || image.Bounds().Empty() || gpu.NativeWindow() == nil {
					return cogent.ErrFrameNotReady
				}
				return json.NewEncoder(os.Stdout).Encode(struct {
					Event            string `json:"event"`
					GeometryIsolated bool   `json:"geometry_persistence_disabled_in_fixture"`
					NativeWindow     bool   `json:"native_window"`
					Width            int    `json:"width"`
					Height           int    `json:"height"`
				}{"vellum.automation.native.ready", core.DebugSettings.DisableWindowGeometrySaver, true, image.Bounds().Dx(), image.Bounds().Dy()})
			})
			if !errors.Is(err, cogent.ErrFrameNotReady) {
				break
			}
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-ticker.C:
				continue
			}
			break
		}
		result <- err
		if err != nil {
			a.Close()
		}
	}()
	return result
}

// A failed descriptor write must close/join the already-created endpoint before
// main's exit helper, which deliberately does not run deferred functions.
func writeAutomationEndpoint(w io.Writer, server *automation.Server) error {
	if err := json.NewEncoder(w).Encode(server.Endpoint()); err != nil {
		return errors.Join(err, closeAutomationTransport(server))
	}
	return nil
}

func closeAutomationTransport(server *automation.Server) error {
	if server == nil {
		return errors.New("automation-check missing server")
	}
	err := server.Close()
	select {
	case <-server.Done():
		return errors.Join(err, server.Err())
	case <-time.After(5 * time.Second):
		return errors.Join(err, errors.New("automation-check transport did not join"))
	}
}

// Join the worker separately from Done, which only joins transport goroutines.
// Close the transport even when readiness failed. Audit contains metadata only.
func finishAutomationNativeCheck(ready <-chan error, server *automation.Server) error {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	err := joinAutomationReadiness(ready, timer.C)
	err = errors.Join(err, closeAutomationTransport(server))
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Event         string           `json:"event"`
		TransportDone bool             `json:"transport_done"`
		Audit         automation.Audit `json:"audit"`
	}{"vellum.automation.native.closed", true, server.Audit()})
}

// The result alone is insufficient: producer closure follows a.Close, ticker
// stop and context cancellation. No worker-owned operation follows that closure.
func joinAutomationReadiness(ready <-chan error, deadline <-chan time.Time) error {
	if ready == nil {
		return errors.New("automation-check missing readiness")
	}
	var err error
	select {
	case value, ok := <-ready:
		if !ok {
			return errors.New("automation-check empty readiness")
		}
		err = value
	case <-deadline:
		return errors.New("automation-check readiness did not join")
	}
	select {
	case _, ok := <-ready:
		if ok {
			return errors.Join(err, errors.New("automation-check repeated readiness result"))
		}
		return err
	case <-deadline:
		return errors.Join(err, errors.New("automation-check worker did not retire"))
	}
}
