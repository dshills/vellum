// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

// This example consumes only Vellum's public, standard-library-only API.
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dshills/vellum/automation"
)

const maxStreamBytes = 8 << 20
const maxLineBytes = 64 << 10

// No child output is published, including diagnostics on malformed/duplicate
// endpoint frames. Only explicitly checked scalar observations leave this type.
// Capability and path stay in memory and are never placed in argv or environment.
type capture struct {
	mu                          sync.Mutex
	pending                     []byte
	total                       int
	endpoint                    automation.Endpoint
	haveEndpoint, ready, closed bool
	audit                       automation.Audit
	failure                     error
	changed                     chan struct{}
}

func newCapture() *capture { return &capture{changed: make(chan struct{}, 1)} }
func (c *capture) fail() error {
	c.failure = errors.New("native output framing refused")
	select {
	case c.changed <- struct{}{}:
	default:
	}
	return c.failure
}
func (c *capture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failure != nil {
		return 0, c.failure
	}
	c.total += len(p)
	if c.total > maxStreamBytes {
		return 0, c.fail()
	}
	// Process incrementally, without allocating proportional to a malicious Write.
	for _, b := range p {
		if b == '\n' {
			if err := c.line(c.pending); err != nil {
				return 0, c.fail()
			}
			c.pending = c.pending[:0]
		} else {
			if len(c.pending) >= maxLineBytes {
				return 0, c.fail()
			}
			c.pending = append(c.pending, b)
		}
	}
	return len(p), nil
}
func strict(data []byte, out any) error {
	if err := uniqueJSON(data); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return errors.New("invalid typed record")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("trailing typed record")
	}
	return nil
}
func (c *capture) line(line []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(line, &fields) != nil {
		return nil
	} // unretained native diagnostics
	_, token := fields["token"]
	_, path := fields["path"]
	_, version := fields["version"]
	if token || path || version {
		var ep automation.Endpoint
		if c.haveEndpoint || c.ready || c.closed || strict(line, &ep) != nil {
			return errors.New("endpoint framing")
		}
		encoded, _ := json.Marshal(ep)
		credential, err := hex.DecodeString(ep.Token)
		if !bytes.Equal(line, encoded) || err != nil || len(credential) != 32 || strings.ToLower(ep.Token) != ep.Token || ep.Version != automation.Version || !filepath.IsAbs(ep.Path) || filepath.Base(ep.Path) != "api.sock" || !strings.HasPrefix(filepath.Base(filepath.Dir(ep.Path)), "vellum-api-") {
			return errors.New("endpoint descriptor")
		}
		c.endpoint = ep
		c.haveEndpoint = true
		return nil
	}
	var event string
	if json.Unmarshal(fields["event"], &event) != nil {
		return nil
	}
	switch event {
	case "vellum.automation.native.ready":
		var record struct {
			Event    string `json:"event"`
			Isolated bool   `json:"geometry_persistence_disabled_in_fixture"`
			Native   bool   `json:"native_window"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
		}
		if !c.haveEndpoint || c.ready || c.closed || strict(line, &record) != nil || !record.Isolated || !record.Native || record.Width <= 0 || record.Height <= 0 {
			return errors.New("native readiness")
		}
		c.ready = true
	case "vellum.automation.native.closed":
		var record struct {
			Event string           `json:"event"`
			Done  bool             `json:"transport_done"`
			Audit automation.Audit `json:"audit"`
		}
		if !c.ready || c.closed || strict(line, &record) != nil || !record.Done || len(record.Audit.Events) > automation.MaxAuditEvents || record.Audit.Dropped != 0 {
			return errors.New("native closure")
		}
		c.closed = true
		c.audit = record.Audit
	default:
		return nil
	}
	select {
	case c.changed <- struct{}{}:
	default:
	}
	return nil
}
func (c *capture) waitReady(ctx context.Context) (automation.Endpoint, error) {
	for {
		c.mu.Lock()
		ep, ready, err := c.endpoint, c.ready, c.failure
		c.mu.Unlock()
		if err != nil {
			return automation.Endpoint{}, err
		}
		if ready {
			return ep, nil
		}
		select {
		case <-ctx.Done():
			return automation.Endpoint{}, errors.New("native readiness deadline")
		case <-c.changed:
		}
	}
}
func (c *capture) complete(events []observation) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failure != nil || !c.closed || len(c.pending) != 0 || len(c.audit.Events) != len(events) {
		return errors.New("native terminal evidence incomplete")
	}
	for i, e := range c.audit.Events {
		want := events[i]
		if e.ID != want.ID || e.Operation != want.Operation || e.Target != want.Target || e.Code != want.Code {
			return errors.New("native audit does not match client responses")
		}
	}
	return nil
}

type discardBounded struct{ n int }

func (d *discardBounded) Write(p []byte) (int, error) {
	d.n += len(p)
	if d.n > maxStreamBytes {
		return 0, errors.New("native stderr limit")
	}
	return len(p), nil
}

// run owns exactly one direct child. It inherits our process group so an outer
// test executor can retire both; we never signal an invented child process group.
func run(ctx context.Context, binary string, rows int) (result summary, returned error) {
	if runtime.GOOS != "darwin" {
		return result, automation.ErrUnsupported
	}
	if !filepath.IsAbs(binary) || (rows != 100000 && rows != 1000000) {
		return result, errors.New("absolute workbench path and rows=100000 or 1000000 required")
	}
	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	output := newCapture()
	stderr := &discardBounded{}
	cmd := exec.Command(binary, "-automation", "-automation-check", "-exit-after=20s", fmt.Sprintf("-rows=%d", rows), "-tree-rows=100000")
	cmd.Stdout = output
	cmd.Stderr = stderr
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		return result, errors.New("native workbench did not start")
	}
	done := make(chan struct{})
	var waitErr error
	go func() { waitErr = cmd.Wait(); close(done) }()
	joined := false
	defer func() {
		if !joined {
			returned = errors.Join(returned, retireChild(func() error { return cmd.Process.Signal(syscall.SIGTERM) }, cmd.Process.Kill, done, 2*time.Second))
		}
	}()
	readyCtx, stop := context.WithTimeout(ctx, 12*time.Second)
	endpoint, err := output.waitReady(readyCtx)
	stop()
	if err != nil {
		return result, err
	}
	client := &session{endpoint: endpoint}
	if err = client.exercise(ctx, rows); err != nil {
		return result, err
	}
	select {
	case <-done:
		joined = true
		err = waitErr
	case <-ctx.Done():
		return result, errors.New("native closure deadline")
	}
	if err != nil {
		return result, errors.New("native workbench exit failure")
	}
	if err = output.complete(client.events); err != nil {
		return result, err
	}
	for _, p := range []string{endpoint.Path, filepath.Dir(endpoint.Path)} {
		if _, err = os.Lstat(p); !os.IsNotExist(err) {
			return result, errors.New("owned endpoint remains after closure")
		}
	}
	if err := probeClosedEndpoint(ctx, endpoint, automation.Call); err != nil {
		return result, err
	}
	return summary{Status: "PASS", Rows: rows, Requests: len(client.events), NativeReady: true, TransportJoined: true, EndpointRemoved: true, PostCloseRefused: true, RawNativeOutputPublished: false}, nil
}

type summary struct {
	Status                   string `json:"status"`
	Rows                     int    `json:"rows"`
	Requests                 int    `json:"requests"`
	NativeReady              bool   `json:"native_ready"`
	TransportJoined          bool   `json:"transport_joined"`
	EndpointRemoved          bool   `json:"endpoint_removed"`
	PostCloseRefused         bool   `json:"post_close_refused"`
	RawNativeOutputPublished bool   `json:"raw_native_output_published"`
}

func main() {
	binary := flag.String("workbench", "", "absolute path to built Vellum workbench")
	rows := flag.Int("rows", 100000, "100000 or 1000000 component records")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	result, err := run(ctx, *binary, *rows)
	if err != nil {
		fmt.Fprintln(os.Stderr, "automation example:", err)
		os.Exit(1)
	}
	if json.NewEncoder(os.Stdout).Encode(result) != nil {
		os.Exit(1)
	}
}

// Signal, escalation and final bounded join are independently attempted. Errors
// use fixed descriptions, without child output, paths or credential material.
func retireChild(term, kill func() error, done <-chan struct{}, wait time.Duration) error {
	select {
	case <-done:
		return nil
	default:
	}
	var err error
	if e := term(); e != nil && !errors.Is(e, os.ErrProcessDone) {
		err = errors.New("native child TERM failed")
	}
	timer := time.NewTimer(wait)
	select {
	case <-done:
		timer.Stop()
		return err
	case <-timer.C:
	}
	if e := kill(); e != nil && !errors.Is(e, os.ErrProcessDone) {
		err = errors.Join(err, errors.New("native child Kill failed"))
	}
	timer = time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-done:
		return err
	case <-timer.C:
		return errors.Join(err, errors.New("native child final join unproven"))
	}
}

// Typed encoding/json decoding alone silently accepts duplicate object keys.
// Reject those at every depth before interpreting readiness, audit or DTO fields.
func uniqueJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	var value func(int) error
	value = func(depth int) error {
		if depth > 32 {
			return errors.New("typed record depth")
		}
		token, err := d.Token()
		if err != nil {
			return errors.New("typed record framing")
		}
		delim, container := token.(json.Delim)
		if !container {
			return nil
		}
		switch delim {
		case '{':
			keys := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return errors.New("typed record key")
				}
				name, ok := key.(string)
				if !ok || keys[name] {
					return errors.New("duplicate typed record key")
				}
				keys[name] = true
				if err = value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim('}') {
				return errors.New("typed record object")
			}
		case '[':
			for d.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return errors.New("typed record array")
			}
		default:
			return errors.New("typed record delimiter")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing typed record")
	}
	return nil
}

func probeClosedEndpoint(runContext context.Context, endpoint automation.Endpoint, call func(context.Context, automation.Endpoint, automation.Request) (automation.Response, error)) error {
	if runContext.Err() != nil {
		return errors.New("automation run canceled before final probe")
	}
	// The probe must actually attempt the missing socket, independently of run
	// cancellation. Neither a canceled context nor a protocol error proves refusal.
	probe, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := call(probe, endpoint, automation.Request{Version: automation.Version, ID: "after-close", Operation: "inspect", Target: "component", Params: json.RawMessage(`{}`)})
	if probe.Err() != nil || !errors.Is(err, os.ErrNotExist) {
		return errors.New("closed endpoint refusal unproven")
	}
	if runContext.Err() != nil {
		return errors.New("automation run canceled during final probe")
	}
	return nil
}
