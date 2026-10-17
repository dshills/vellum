// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/automation"
	"time"
)

type observation struct{ ID, Operation, Target, Code string }
type session struct {
	endpoint automation.Endpoint
	revision uint64
	events   []observation
	// Replaced only by regression controls; the runnable example uses public Call.
	send func(context.Context, automation.Endpoint, automation.Request) (automation.Response, error)
}

func (s *session) request(ctx context.Context, operation, target, params, wantCode string, expected *uint64) (automation.Response, error) {
	request := automation.Request{Version: automation.Version, ID: fmt.Sprintf("check:%02d", len(s.events)+1), Operation: operation, Target: target, Params: json.RawMessage(params), ExpectedRevision: expected}
	call := s.send
	if call == nil {
		call = automation.Call
	}
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	response, err := call(bounded, s.endpoint, request)
	if err != nil {
		return response, errors.New("public automation transport failed; invoke was not retried")
	}
	code := "ok"
	if response.Error != nil {
		code = response.Error.Code
	}
	if code != wantCode {
		return response, errors.New("unexpected semantic response code")
	}
	s.events = append(s.events, observation{request.ID, operation, target, code})
	if response.Revision != nil {
		s.revision = *response.Revision
	}
	return response, nil
}
func (s *session) invoke(ctx context.Context, target, params, code string, revision uint64) error {
	_, err := s.request(ctx, "invoke", target, params, code, &revision)
	return err
}

type component struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
	Value     uint32 `json:"value"`
	Package   string `json:"package"`
	Rotation  uint16 `json:"rotation"`
}

func (s *session) inspectComponent(ctx context.Context, params string) (component, error) {
	var c component
	r, err := s.request(ctx, "inspect", "component", params, "ok", nil)
	if err != nil {
		return c, err
	}
	if r.Revision == nil || *r.Revision == 0 || strict(r.Result, &c) != nil {
		return c, errors.New("component DTO or revision refused")
	}
	return c, nil
}
func (s *session) expectComponent(ctx context.Context, want component) error {
	got, err := s.inspectComponent(ctx, `{}`)
	if err != nil {
		return err
	}
	if got != want {
		return errors.New("component state mismatch")
	}
	return nil
}

type workbenchState struct {
	Selected     string `json:"selected"`
	Components   int    `json:"components"`
	DocumentPath string `json:"document_path"`
	References   int    `json:"references"`
	Commands     []struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	} `json:"commands"`
}

func (s *session) workbench(ctx context.Context, rows int, selected string, undo, redo bool) error {
	r, err := s.request(ctx, "inspect", "workbench", `{}`, "ok", nil)
	if err != nil {
		return err
	}
	var state workbenchState
	if r.Revision == nil || *r.Revision == 0 || strict(r.Result, &state) != nil || state.Selected != selected || state.Components != rows || state.DocumentPath != "[redacted]" || state.References != 0 || len(state.Commands) != 4 {
		return errors.New("workbench redaction, selection or count mismatch")
	}
	want := map[string]bool{"component.rotate": true, "component.raise-value": true, "edit.undo": undo, "edit.redo": redo}
	for _, v := range state.Commands {
		enabled, ok := want[v.ID]
		if !ok || v.Enabled != enabled {
			return errors.New("command availability mismatch")
		}
		delete(want, v.ID)
	}
	if len(want) != 0 {
		return errors.New("missing command availability")
	}
	return nil
}
func initialComponent(number int) component {
	index := number - 1
	pack := "0603"
	if index%2 == 1 {
		pack = "0805"
	}
	return component{fmt.Sprintf("component:%d", number), fmt.Sprintf("R%07d", number), uint32(100 + index%997), pack, uint16(index % 4 * 90)}
}

// Actions use current revisions, except explicit stale-revision probes. Every
// invoke is sent once. A transport failure can follow a commit and stops the run.
func (s *session) exercise(ctx context.Context, rows int) error {
	if err := s.workbench(ctx, rows, "component:1", false, false); err != nil {
		return err
	}
	if err := s.invoke(ctx, "edit.undo", `{}`, "rejected", s.revision); err != nil {
		return err
	}
	original, err := s.inspectComponent(ctx, `{"id":"component:17"}`)
	if err != nil {
		return err
	}
	if original != initialComponent(17) {
		return errors.New("initial component mismatch")
	}
	last, err := s.inspectComponent(ctx, fmt.Sprintf(`{"id":"component:%d"}`, rows))
	if err != nil {
		return err
	}
	if last != initialComponent(rows) {
		return errors.New("virtual component mismatch")
	}
	for _, tc := range []struct{ op, target, params, code string }{
		{"inspect", "component", `{"id":"component:017"}`, "rejected"},
		{"inspect", "component", `{"id":17}`, "invalid_params"},
		{"invoke", "component.select", `{"id":"component:17","unknown":true}`, "invalid_params"},
		{"invoke", "file.save", `{}`, "unknown_target"},
	} {
		revision := s.revision
		var expected *uint64
		if tc.op == "invoke" {
			expected = &revision
		}
		if _, err := s.request(ctx, tc.op, tc.target, tc.params, tc.code, expected); err != nil {
			return err
		}
	}
	if err := s.workbench(ctx, rows, "component:1", false, false); err != nil {
		return err
	}
	if err := s.invoke(ctx, "component.select", `{"id":"component:17"}`, "ok", s.revision); err != nil {
		return err
	}
	if err := s.expectComponent(ctx, original); err != nil {
		return err
	}
	beforeRotate := s.revision
	if err := s.invoke(ctx, "component.rotate", `{}`, "ok", beforeRotate); err != nil {
		return err
	}
	rotated := original
	rotated.Rotation = (original.Rotation + 90) % 360
	if err := s.expectComponent(ctx, rotated); err != nil {
		return err
	}
	if err := s.invoke(ctx, "component.raise-value", `{}`, "stale_revision", beforeRotate); err != nil {
		return err
	}
	if err := s.expectComponent(ctx, rotated); err != nil {
		return err
	}
	if err := s.invoke(ctx, "component.raise-value", `{}`, "ok", s.revision); err != nil {
		return err
	}
	edited := rotated
	edited.Value += 100
	if err := s.expectComponent(ctx, edited); err != nil {
		return err
	}
	afterRaise := s.revision
	if err := s.workbench(ctx, rows, "component:17", true, false); err != nil {
		return err
	}
	for _, want := range []component{rotated, original} {
		if err := s.invoke(ctx, "edit.undo", `{}`, "ok", s.revision); err != nil {
			return err
		}
		if err := s.expectComponent(ctx, want); err != nil {
			return err
		}
	}
	if err := s.workbench(ctx, rows, "component:17", false, true); err != nil {
		return err
	}
	for _, want := range []component{rotated, edited} {
		if err := s.invoke(ctx, "edit.redo", `{}`, "ok", s.revision); err != nil {
			return err
		}
		if err := s.expectComponent(ctx, want); err != nil {
			return err
		}
	}
	// Equal domain values after Undo/Redo cannot make an old revision reusable.
	if err := s.invoke(ctx, "component.rotate", `{}`, "stale_revision", afterRaise); err != nil {
		return err
	}
	if err := s.expectComponent(ctx, edited); err != nil {
		return err
	}
	beforeSelection := s.revision
	for _, id := range []string{"component:1", "component:17"} {
		if err := s.invoke(ctx, "component.select", fmt.Sprintf(`{"id":%q}`, id), "ok", s.revision); err != nil {
			return err
		}
	}
	if err := s.invoke(ctx, "component.rotate", `{}`, "stale_revision", beforeSelection); err != nil {
		return err
	}
	if err := s.expectComponent(ctx, edited); err != nil {
		return err
	}
	for _, want := range []component{rotated, original} {
		if err := s.invoke(ctx, "edit.undo", `{}`, "ok", s.revision); err != nil {
			return err
		}
		if err := s.expectComponent(ctx, want); err != nil {
			return err
		}
	}
	if err := s.workbench(ctx, rows, "component:17", false, true); err != nil {
		return err
	}
	if err := s.invoke(ctx, "edit.undo", `{}`, "rejected", s.revision); err != nil {
		return err
	}
	return s.expectComponent(ctx, original)
}
