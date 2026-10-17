// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
)

// Independent tiny application fixture, using the real public Registry, owner
// queue, Unix endpoint and Call. It opens no native windows. Faults prove the
// example's assertions distinguish incorrect domain/history/DTO behavior.
type model struct {
	rows, selected, cursor int
	revision               uint64
	states                 []component
	fault                  string
}
type componentParam struct {
	ID string `json:"id"`
}

func (m *model) registry(t *testing.T) *automation.Registry {
	t.Helper()
	m.selected = 1
	m.revision = 1
	m.states = []component{initialComponent(17)}
	r := automation.NewRegistry(func() (uint64, error) {
		if m.fault == "revision-aba" {
			return uint64(m.cursor + m.selected), nil
		}
		return m.revision, nil
	})
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	index := func(id string) (int, error) {
		n, err := strconv.Atoi(strings.TrimPrefix(id, "component:"))
		if err != nil || n < 1 || n > m.rows || fmt.Sprintf("component:%d", n) != id {
			return 0, errors.New("identity")
		}
		return n, nil
	}
	check(automation.RegisterQuery(r, "component", func(_ context.Context, p componentParam) (any, error) {
		n := m.selected
		var err error
		if p.ID != "" {
			n, err = index(p.ID)
		}
		if err != nil {
			return nil, err
		}
		value := initialComponent(n)
		if n == 17 {
			value = m.states[m.cursor]
		}
		if m.fault == "virtual" && n == m.rows {
			value.Reference = "wrong"
		}
		return value, nil
	}))
	check(automation.RegisterQuery(r, "workbench", func(context.Context, struct{}) (any, error) {
		path := "[redacted]"
		if m.fault == "redaction" {
			path = "private-path"
		}
		state := workbenchState{Selected: fmt.Sprintf("component:%d", m.selected), Components: m.rows, DocumentPath: path}
		for _, id := range []string{"component.rotate", "component.raise-value", "edit.undo", "edit.redo"} {
			enabled := true
			if id == "edit.undo" {
				enabled = m.cursor > 0
			}
			if id == "edit.redo" {
				enabled = m.cursor < len(m.states)-1
			}
			if m.fault == "availability" && id == "edit.undo" {
				enabled = !enabled
			}
			state.Commands = append(state.Commands, struct {
				ID      string `json:"id"`
				Enabled bool   `json:"enabled"`
			}{id, enabled})
		}
		return state, nil
	}))
	check(automation.RegisterAction(r, "component.select", func(_ context.Context, p componentParam) (any, error) {
		n, err := index(p.ID)
		if err == nil && n != m.selected {
			m.selected = n
			m.revision++
		}
		return struct{}{}, err
	}))
	for _, target := range []string{"component.rotate", "component.raise-value", "edit.undo", "edit.redo"} {
		check(automation.RegisterAction(r, target, func(context.Context, struct{}) (any, error) {
			switch target {
			case "component.rotate", "component.raise-value":
				if m.selected != 17 {
					return nil, errors.New("fixture requires selected component17")
				}
				next := m.states[m.cursor]
				if target == "component.rotate" {
					next.Rotation = (next.Rotation + 90) % 360
				} else {
					next.Value += 100
				}
				m.states = append(m.states[:m.cursor+1], next)
				m.cursor++
			case "edit.undo":
				if m.cursor == 0 {
					return nil, errors.New("empty history")
				}
				if m.fault != "undo" {
					m.cursor--
				}
			case "edit.redo":
				if m.cursor == len(m.states)-1 {
					return nil, errors.New("empty history")
				}
				if m.fault != "redo" {
					m.cursor++
				}
			}
			m.revision++
			return struct{}{}, nil
		}))
	}
	return r
}
func testModel(t *testing.T, rows int, fault string) error {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("Mac endpoint")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m := &model{rows: rows, fault: fault}
	r := m.registry(t)
	wake := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	})
	server, err := automation.Start(automation.Config{Owner: a, Registry: r})
	if err != nil {
		a.Close()
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-a.Context().Done():
				return
			case <-wake:
				a.Drain()
			}
		}
	}()
	defer func() {
		a.Close()
		server.Close()
		for _, ch := range []<-chan struct{}{done, server.Done()} {
			select {
			case <-ch:
			case <-time.After(5 * time.Second):
				t.Error("headless fixture did not join")
			}
		}
	}()
	s := &session{endpoint: server.Endpoint()}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = s.exercise(ctx, rows)
	if err == nil {
		if len(s.events) != 40 {
			t.Fatalf("unexpected coverage count: %d", len(s.events))
		}
		restored := false
		if err := a.Do(func() { restored = m.states[m.cursor] == initialComponent(17) && m.cursor == 0 && m.selected == 17 }); err != nil {
			t.Fatal(err)
		}
		if !restored {
			t.Fatal("final state not restored")
		}
		audit := server.Audit()
		if audit.Dropped != 0 || len(audit.Events) != len(s.events) {
			t.Fatal("audit coverage mismatch")
		}
		for i, e := range audit.Events {
			if e.ID != s.events[i].ID || e.Code != s.events[i].Code {
				t.Fatal("audit response mismatch")
			}
		}
	}
	return err
}
func TestPublicClientHistoryAndVirtualQueries(t *testing.T) {
	for _, rows := range []int{100000, 1000000} {
		t.Run(strconv.Itoa(rows), func(t *testing.T) {
			if err := testModel(t, rows, ""); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWorkflowRejectsIndependentFaults(t *testing.T) {
	for _, fault := range []string{"redaction", "virtual", "availability", "undo", "redo", "revision-aba"} {
		t.Run(fault, func(t *testing.T) {
			want := map[string]string{"redaction": "workbench redaction, selection or count mismatch", "virtual": "virtual component mismatch", "availability": "command availability mismatch", "undo": "component state mismatch", "redo": "component state mismatch", "revision-aba": "unexpected semantic response code"}[fault]
			if err := testModel(t, 100000, fault); err == nil || err.Error() != want {
				t.Fatalf("fault was not refused for expected reason %q: %v", want, err)
			}
		})
	}
}
