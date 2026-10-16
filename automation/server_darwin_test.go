//go:build darwin

package automation

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func fixture(t *testing.T, automatic bool, setup func(*app.Application, *Registry), timeout ...time.Duration) (*app.Application, *Server, chan struct{}) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(func() (uint64, error) { return 1, nil })
	setup(a, r)
	wake := make(chan struct{}, 64)
	a.SetWake(func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	})
	config := Config{Owner: a, Window: "main", Registry: r}
	if len(timeout) > 0 {
		config.Timeout = timeout[0]
	}
	s, err := Start(config)
	if err != nil {
		a.Close()
		t.Fatal(err)
	}
	loopDone := make(chan struct{})
	if automatic {
		go func() {
			defer close(loopDone)
			for {
				select {
				case <-a.Context().Done():
					return
				case <-wake:
					a.Drain()
				}
			}
		}()
	} else {
		close(loopDone)
	}
	t.Cleanup(func() { a.Close(); s.Close(); wait(t, s.Done()); wait(t, loopDone) })
	return a, s, wake
}
func wait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("bounded test wait expired")
	}
}
func request(op, target string, revision *uint64, params string) Request {
	return Request{Version: 1, ID: "r1", Operation: op, Target: target, ExpectedRevision: revision, Params: json.RawMessage(params)}
}
func call(t *testing.T, s *Server, r Request) Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	v, err := Call(ctx, s.Endpoint(), r)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func requireCode(t *testing.T, r Response, code string) {
	t.Helper()
	if r.Error == nil || r.Error.Code != code {
		t.Fatalf("want %s: %+v", code, r)
	}
}
func TestOwnerRevisionTypedActionsAndCommandParity(t *testing.T) {
	var revision uint64 = 1
	value := 10
	enabled := true
	_, s, _ := fixture(t, true, func(a *app.Application, r *Registry) {
		r.revision = func() (uint64, error) { return revision, nil }
		commands := command.New()
		if err := commands.Register(command.Command{ID: "raise", Label: "Raise", Available: func() error {
			if !enabled {
				return errors.New("private reason")
			}
			return nil
		}, Execute: func(context.Context) error { value++; revision++; return nil }}); err != nil {
			t.Fatal(err)
		}
		if err := RegisterCommand(r, commands, "raise"); err != nil {
			t.Fatal(err)
		}
		if err := RegisterQuery(r, "component", func(context.Context, struct{}) (any, error) {
			return struct {
				Value int    `json:"value"`
				Path  string `json:"path"`
			}{value, "[redacted]"}, nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := RegisterAction(r, "disable", func(context.Context, struct{}) (any, error) { enabled = false; revision++; return struct{}{}, nil }); err != nil {
			t.Fatal(err)
		}
	})
	first := call(t, s, request("inspect", "component", nil, `{}`))
	if *first.Revision != 1 || !strings.Contains(string(first.Result), `"value":10`) {
		t.Fatal(first)
	}
	requireCode(t, call(t, s, request("invoke", "raise", first.Revision, `{"unused":true}`)), "invalid_params")
	raised := call(t, s, request("invoke", "raise", first.Revision, `{}`))
	if raised.Error != nil || *raised.Revision != 2 {
		t.Fatal(raised)
	}
	requireCode(t, call(t, s, request("invoke", "raise", first.Revision, `{}`)), "stale_revision")
	disabled := call(t, s, request("invoke", "disable", raised.Revision, `{}`))
	requireCode(t, call(t, s, request("invoke", "raise", disabled.Revision, `{}`)), "rejected")
	requireCode(t, call(t, s, request("inspect", "missing", nil, `{}`)), "unknown_target")
	if string(call(t, s, request("inspect", "component", nil, `{}`)).Result) != `{"value":11,"path":"[redacted]"}` {
		t.Fatal("mutation/redaction")
	}
	for _, e := range s.Audit().Events {
		if strings.Contains(e.Code, "private") {
			t.Fatal("error leak")
		}
	}
	if err := RegisterCommand(s.registry, command.New(), "another"); err == nil {
		t.Fatal("late registration")
	}
}
func TestEndpointPermissionsCredentialsAndStrictWire(t *testing.T) {
	var calls atomic.Int32
	_, s, _ := fixture(t, true, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { calls.Add(1); return struct{}{}, nil })
	})
	ep := s.Endpoint()
	dir, err := os.Stat(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	socket, err := os.Lstat(ep.Path)
	if err != nil {
		t.Fatal(err)
	}
	if dir.Mode().Perm() != 0700 || socket.Mode().Perm() != 0600 || socket.Mode()&os.ModeSocket == 0 || len(ep.Token) != 64 {
		t.Fatal("private endpoint", dir.Mode(), socket.Mode())
	}
	for _, wire := range []string{strings.Repeat("0", 64) + "\n{}\n", ep.Token + "\n" + `{"version":1,"id":"r","operation":"inspect","target":"x","params":{},"params":{}}` + "\n"} {
		conn, err := net.Dial("unix", ep.Path)
		if err != nil {
			t.Fatal(err)
		}
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		conn.Write([]byte(wire))
		reader := bufio.NewReader(conn)
		raw, _ := reader.ReadBytes('\n')
		conn.Close()
		if len(raw) > 0 && !strings.Contains(string(raw), "invalid_request") {
			t.Fatal(string(raw))
		}
	}
	if calls.Load() != 0 {
		t.Fatal("unauthorized callback")
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	wait(t, s.Done())
	if _, err = os.Lstat(ep.Path); !os.IsNotExist(err) {
		t.Fatal("socket retained", err)
	}
	if _, err = os.Stat(s.dir); !os.IsNotExist(err) {
		t.Fatal("directory retained", err)
	}
}
func TestDisconnectRetainsOwnerCreditsAndSkipsActions(t *testing.T) {
	var calls atomic.Int32
	a, s, wake := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterAction(r, "x", func(context.Context, struct{}) (any, error) { calls.Add(1); return struct{}{}, nil })
	})
	<-wake // Initial Open wake, before any request.
	ep := s.Endpoint()
	wire := ep.Token + "\n" + `{"version":1,"id":"r","operation":"invoke","target":"x","expected_revision":1,"params":{}}` + "\n"
	for i := 0; i < MaxClients; i++ {
		conn, err := net.Dial("unix", ep.Path)
		if err != nil {
			t.Fatal(err)
		}
		conn.Write([]byte(wire))
		wait(t, wake)
		conn.Close()
	}
	conn, err := net.Dial("unix", ep.Path)
	if err != nil {
		t.Fatal(err)
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	conn.Write([]byte(wire))
	if _, err = bufio.NewReader(conn).ReadByte(); err == nil {
		t.Fatal("ninth admitted")
	}
	conn.Close()
	if len(s.credits) != MaxClients {
		t.Fatal("disconnected credit returned", len(s.credits))
	}
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Fatal("disconnected action ran")
	}
	// Closing the app retires any remaining transport work without another drain.
	a.Close()
	wait(t, s.Done())
	if len(s.credits) != 0 {
		t.Fatal("credit retained after owner closure")
	}
}
func TestCancellationBeforeOwnerAndTargetClosure(t *testing.T) {
	var calls atomic.Int32
	a, s, wake := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterAction(r, "x", func(context.Context, struct{}) (any, error) { calls.Add(1); return struct{}{}, nil })
	})
	<-wake
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan struct{})
	go func() { defer close(finished); Call(ctx, s.Endpoint(), request("invoke", "x", pointer(1), `{}`)) }()
	wait(t, wake)
	cancel()
	wait(t, finished)
	a.CloseWindow(automationWindowRef(t, a, "main"))
	wait(t, s.Done())
	if calls.Load() != 0 {
		t.Fatal("closed target invoked")
	}
}
func pointer(n uint64) *uint64 { return &n }
func TestOwnerSafeCloseCooperativeCancellationAndResponseLimit(t *testing.T) {
	entered := make(chan struct{})
	ended := make(chan struct{})
	var s *Server
	a, server, _ := fixture(t, true, func(_ *app.Application, r *Registry) {
		RegisterAction(r, "wait", func(ctx context.Context, _ struct{}) (any, error) {
			close(entered)
			<-ctx.Done()
			close(ended)
			return nil, ctx.Err()
		})
		RegisterAction(r, "close", func(context.Context, struct{}) (any, error) { return struct{}{}, s.Close() })
		RegisterQuery(r, "oversize", func(context.Context, struct{}) (any, error) { return strings.Repeat("x", MaxResponseBytes), nil })
		RegisterQuery(r, "secret", func(context.Context, struct{}) (any, error) { return nil, errors.New("/private/secret API_KEY=secret") })
	})
	s = server
	requireCode(t, call(t, s, request("inspect", "oversize", nil, `{}`)), "result_limit")
	secret := call(t, s, request("inspect", "secret", nil, `{}`))
	requireCode(t, secret, "rejected")
	if strings.Contains(secret.Error.Message, "secret") {
		t.Fatal(secret)
	}
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan struct{})
	go func() { defer close(finished); Call(ctx, s.Endpoint(), request("invoke", "wait", pointer(1), `{}`)) }()
	wait(t, entered)
	cancel()
	wait(t, finished)
	wait(t, ended)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if _, err := Call(ctx2, s.Endpoint(), request("invoke", "close", pointer(1), `{}`)); err == nil {
		t.Fatal("expected retired connection")
	}
	wait(t, s.Done())
	a.Close()
}
func TestAuditBoundDetachedAndDuplicateIDNotDeduplicated(t *testing.T) {
	var count int
	_, s, _ := fixture(t, true, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { count++; return count, nil })
	})
	for i := 0; i < MaxAuditEvents+3; i++ {
		if call(t, s, request("inspect", "x", nil, `{}`)).Error != nil {
			t.Fatal(i)
		}
	}
	a := s.Audit()
	if len(a.Events) != MaxAuditEvents || a.Dropped != 3 {
		t.Fatal(len(a.Events), a.Dropped)
	}
	a.Events[0].Code = "changed"
	*a.Events[0].Before = 900
	if b := s.Audit(); b.Events[0].Code == "changed" || *b.Events[0].Before == 900 {
		t.Fatal("borrowed audit")
	}
}
func TestQueueFullAndShutdown(t *testing.T) {
	a, s, _ := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { return struct{}{}, nil })
	})
	for i := 0; i < 32; i++ {
		if err := a.Post(func() {}); err != nil {
			t.Fatal(err)
		}
	}
	requireCode(t, call(t, s, request("inspect", "x", nil, `{}`)), "unavailable")
	a.Close()
	wait(t, s.Done())
	if len(s.credits) != 0 {
		t.Fatal("shutdown credit")
	}
}
func TestCallbackPanicClosesOwnerAndEndpoint(t *testing.T) {
	a, s, wake := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterAction(r, "panic", func(context.Context, struct{}) (any, error) { panic("application programmer error") })
	})
	<-wake
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		Call(context.Background(), s.Endpoint(), request("invoke", "panic", pointer(1), `{}`))
	}()
	wait(t, wake)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic swallowed")
			}
		}()
		a.Drain()
	}()
	wait(t, finished)
	wait(t, s.Done())
	if a.Context().Err() == nil {
		t.Fatal("owner not closed")
	}
}
func TestConcurrentClientsSerializeOwner(t *testing.T) {
	var active atomic.Int32
	var overlap atomic.Bool
	count := 0
	_, s, _ := fixture(t, true, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) {
			if active.Add(1) != 1 {
				overlap.Store(true)
			}
			count++
			active.Add(-1)
			return count, nil
		})
	})
	var workers sync.WaitGroup
	for i := 0; i < MaxClients; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := Call(ctx, s.Endpoint(), request("inspect", "x", nil, `{}`)); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	if overlap.Load() {
		t.Fatal("concurrent owner handlers")
	}
}

func TestHalfCloseAndSecondMessageCancelQueuedAction(t *testing.T) {
	for _, mode := range []string{"half_close", "second_message"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			a, s, wake := fixture(t, false, func(_ *app.Application, r *Registry) {
				RegisterAction(r, "x", func(context.Context, struct{}) (any, error) { calls.Add(1); return struct{}{}, nil })
			}, time.Minute)
			<-wake
			conn, err := net.Dial("unix", s.Endpoint().Path)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			wire := s.Endpoint().Token + "\n" + `{"version":1,"id":"r","operation":"invoke","target":"x","expected_revision":1,"params":{}}` + "\n"
			conn.Write([]byte(wire))
			wait(t, wake)
			if mode == "half_close" {
				if err = conn.(*net.UnixConn).CloseWrite(); err != nil {
					t.Fatal(err)
				}
			} else {
				conn.Write([]byte("{}\n"))
			}
			if _, err = bufio.NewReader(conn).ReadByte(); !errors.Is(err, io.EOF) {
				t.Fatal("framing did not cancel before owner", err)
			}
			if err = a.Drain(); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 0 {
				t.Fatal("canceled framing executed")
			}
		})
	}
}

func TestClosePreservesChangedOwnedPaths(t *testing.T) {
	for _, mode := range []string{"socket", "directory"} {
		t.Run(mode, func(t *testing.T) {
			_, s, _ := fixture(t, false, func(_ *app.Application, r *Registry) {
				RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { return nil, nil })
			})
			if mode == "socket" {
				if err := os.Remove(s.Endpoint().Path); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(s.Endpoint().Path, []byte("preserve"), 0600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Remove(s.Endpoint().Path); os.Remove(s.dir) })
			} else {
				moved := s.dir + "-original"
				if err := os.Rename(s.dir, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(s.dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(s.dir, "marker"), []byte("preserve"), 0600); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					os.Remove(filepath.Join(s.dir, "marker"))
					os.Remove(s.dir)
					os.Remove(filepath.Join(moved, "api.sock"))
					os.Remove(moved)
				})
			}
			if err := s.Close(); err == nil {
				t.Fatal("changed owned path accepted")
			}
			if s.Close() == nil {
				t.Fatal("lost cleanup error")
			}
			if s.Err() == nil {
				t.Fatal("unobservable cleanup error")
			}
			wait(t, s.Done())
			marker := s.Endpoint().Path
			if mode == "directory" {
				marker = filepath.Join(s.dir, "marker")
			}
			if b, err := os.ReadFile(marker); err != nil || string(b) != "preserve" {
				t.Fatal("non-owned path changed", err)
			}
		})
	}
}

func TestFatalAcceptCauseRemainsObservable(t *testing.T) {
	_, s, _ := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { return struct{}{}, nil })
	})
	cause := errors.New("injected permanent accept failure")
	if !errors.Is(s.closeWithCause(cause), cause) {
		t.Fatal("lost cause")
	}
	wait(t, s.Done())
	if !errors.Is(s.Err(), cause) || !errors.Is(s.Close(), cause) {
		t.Fatal("terminal cause not retained")
	}
}

func TestWindowScopeClosurePreservesOtherWindow(t *testing.T) {
	var calls atomic.Int32
	a, s, wake := fixture(t, false, func(_ *app.Application, r *Registry) {
		RegisterAction(r, "x", func(context.Context, struct{}) (any, error) { calls.Add(1); return struct{}{}, nil })
	})
	if err := a.Open(app.Window{ID: "other", Title: "Other headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "other.root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	<-wake
	<-wake
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Call(ctx, s.Endpoint(), request("invoke", "x", pointer(1), `{}`))
	}()
	wait(t, wake)
	if err := a.CloseWindow(automationWindowRef(t, a, "main")); err != nil {
		t.Fatal(err)
	}
	current, err := a.OpenWindow(app.Window{ID: "main", Title: "Reopened headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "reopened", Kind: ui.Label} }})
	if err != nil {
		t.Fatal(err)
	}
	wait(t, s.Done())
	wait(t, finished)
	if a.Context().Err() != nil || current.Context().Err() != nil || len(a.Windows()) != 2 {
		t.Fatal("unrelated window retired")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 || len(s.credits) != 0 {
		t.Fatal("closed window action or retained credit")
	}
}

func TestUninitializedOwnerRefusesBeforeRegistryFreeze(t *testing.T) {
	r := NewRegistry(func() (uint64, error) { return 1, nil })
	if _, err := Start(Config{Owner: &app.Application{}, Registry: r}); err == nil {
		t.Fatal("zero owner admitted")
	}
	if r.frozen {
		t.Fatal("invalid owner froze registry")
	}
}

func TestCopiedHandlesRefuseWithoutChangingOriginal(t *testing.T) {
	a, s, _ := fixture(t, true, func(_ *app.Application, r *Registry) {
		RegisterQuery(r, "x", func(context.Context, struct{}) (any, error) { return struct{}{}, nil })
	})
	clone := &Registry{}
	s.registry.mu.Lock()
	reflect.ValueOf(clone).Elem().Set(reflect.ValueOf(s.registry).Elem())
	s.registry.mu.Unlock()
	if RegisterQuery(clone, "new", func(context.Context, struct{}) (any, error) { return nil, nil }) == nil {
		t.Fatal("copied registry mutated")
	}
	if _, err := Start(Config{Owner: a, Registry: clone}); err == nil {
		t.Fatal("copied registry start")
	}
	// Copy shared resources under the server lock without copying its live mutex.
	// Copying an active mutex itself is unsupported and would race its waiters.
	s.mu.Lock()
	copyServer := &Server{self: s, owner: s.owner, registry: s.registry, ctx: s.ctx, cancel: s.cancel,
		endpoint: s.endpoint, listener: s.listener, dir: s.dir, dirInfo: s.dirInfo, socketInfo: s.socketInfo,
		conns: s.conns, credits: s.credits, done: s.done}
	s.mu.Unlock()
	if copyServer.Close() == nil || copyServer.Endpoint().Token != "" || copyServer.Done() != nil {
		t.Fatal("copied server accepted")
	}
	if v := call(t, s, request("inspect", "x", nil, `{}`)); v.Error != nil {
		t.Fatal("original retired", v)
	}
}

func automationWindowRef(t *testing.T, a *app.Application, name string) app.WindowRef {
	t.Helper()
	ref, err := a.WindowRef(name)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
