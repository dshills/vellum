package automation

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/dshills/vellum/app"
)

// Endpoint includes a bearer capability. Pass it privately to trusted clients;
// do not publish it in logs, audit events, files or UI semantics.
type Endpoint struct {
	Path    string `json:"path"`
	Token   string `json:"token"`
	Version int    `json:"version"`
}
type Config struct {
	Owner    *app.Application
	Window   string // Optional; closing this window retires the endpoint.
	Registry *Registry
	Timeout  time.Duration // Defaults to five seconds; maximum one minute.
}

// Event describes the terminal transport observation, not proof of rollback.
// If Started is true, a canceled/lost response may follow a committed action.
type Event struct {
	ID, Operation, Target, Code string
	Started                     bool
	Before, After               *uint64
}
type Audit struct {
	Events  []Event
	Dropped uint64
}

type Server struct {
	self       *Server
	owner      *app.Application
	registry   *Registry
	ctx        context.Context
	cancel     context.CancelFunc
	timeout    time.Duration
	endpoint   Endpoint
	listener   *net.UnixListener
	dir        string
	dirInfo    os.FileInfo
	socketInfo os.FileInfo
	mu         sync.Mutex
	conns      map[net.Conn]bool
	credits    chan struct{}
	closed     bool
	closeErr   error
	events     []Event
	dropped    uint64
	done       chan struct{}
}

// Start is explicit; no imported package or driver opens an endpoint by default.
// Register callbacks first, then start before Run or on the owner. Close never
// waits for owner work, so a command may close the server safely.
func Start(c Config) (*Server, error) {
	if !localSupported() {
		return nil, ErrUnsupported
	}
	if c.Owner == nil || c.Registry == nil {
		return nil, errors.New("automation: owner and registry required")
	}
	if c.Registry.self != c.Registry {
		return nil, errors.New("automation: copied or uninitialized registry")
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	if c.Timeout < 0 || c.Timeout > time.Minute {
		return nil, errors.New("automation: invalid timeout")
	}
	parent := c.Owner.Context()
	if c.Window != "" {
		var err error
		parent, err = c.Owner.WindowContext(c.Window)
		if err != nil {
			return nil, err
		}
	}
	if parent == nil {
		return nil, errors.New("automation: initialized application required")
	}
	if err := parent.Err(); err != nil {
		return nil, err
	}
	c.Registry.mu.Lock()
	defer c.Registry.mu.Unlock()
	if c.Registry.frozen || c.Registry.revision == nil {
		return nil, errors.New("automation: unusable registry")
	}
	var capability [32]byte
	if _, err := rand.Read(capability[:]); err != nil {
		return nil, err
	}
	// Keep the socket path below Darwin's short sockaddr_un limit. This newly
	// owned directory is private even under a permissive process umask.
	dir, err := os.MkdirTemp("/tmp", "vellum-api-")
	if err != nil {
		return nil, err
	}
	cleanup := func() { os.Remove(filepath.Join(dir, "api.sock")); os.Remove(dir) }
	if err = os.Chmod(dir, 0700); err != nil {
		cleanup()
		return nil, err
	}
	path := filepath.Join(dir, "api.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		cleanup()
		return nil, err
	}
	listener.SetUnlinkOnClose(false)
	if err = os.Chmod(path, 0600); err != nil {
		listener.Close()
		cleanup()
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		listener.Close()
		cleanup()
		return nil, err
	}
	dirInfo, err := os.Lstat(dir)
	if err != nil {
		listener.Close()
		cleanup()
		return nil, err
	}
	ctx, cancel := context.WithCancel(parent)
	s := &Server{owner: c.Owner, registry: c.Registry, ctx: ctx, cancel: cancel, timeout: c.Timeout, endpoint: Endpoint{path, hex.EncodeToString(capability[:]), Version}, listener: listener, dir: dir, dirInfo: dirInfo, socketInfo: info, conns: map[net.Conn]bool{}, credits: make(chan struct{}, MaxClients), done: make(chan struct{})}
	s.self = s
	c.Registry.frozen = true
	go s.accept()
	context.AfterFunc(ctx, func() { s.Close() })
	return s, nil
}

func (s *Server) Endpoint() Endpoint {
	if s == nil || s.self != s {
		return Endpoint{}
	}
	return s.endpoint
}

// Done returns nil for a copied/uninitialized handle. Do not copy Server.
func (s *Server) Done() <-chan struct{} {
	if s == nil || s.self != s {
		return nil
	}
	return s.done
}

// Close retires listener/clients and removes only the owned socket/empty
// directory. Done closes after transport goroutines end; admitted owner
// envelopes remain bounded by the application's mailbox until drained/closed.
func (s *Server) Close() error {
	return s.closeWithCause(nil)
}

// Err reports fatal accept/owned-cleanup errors. Normal closure returns nil.
func (s *Server) Err() error {
	if s == nil || s.self != s {
		return errors.New("automation: copied or uninitialized server")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeErr
}

func (s *Server) closeWithCause(cause error) error {
	if s == nil || s.self != s {
		return errors.New("automation: copied or uninitialized server")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		if cause != nil {
			s.closeErr = errors.Join(s.closeErr, cause)
		}
		return s.closeErr
	}
	s.closed = true
	s.cancel()
	err := errors.Join(cause, s.listener.Close())
	for conn := range s.conns {
		conn.Close()
	}
	dirInfo, dirErr := os.Lstat(s.dir)
	if dirErr != nil || !os.SameFile(dirInfo, s.dirInfo) {
		if !os.IsNotExist(dirErr) {
			err = errors.Join(err, errors.New("automation: owned directory changed"))
		}
		s.closeErr = err
		return err
	}
	info, statErr := os.Lstat(s.endpoint.Path)
	if statErr == nil && os.SameFile(info, s.socketInfo) {
		err = errors.Join(err, os.Remove(s.endpoint.Path))
	} else if statErr == nil {
		err = errors.Join(err, errors.New("automation: owned socket changed"))
	} else if statErr != nil && !os.IsNotExist(statErr) {
		err = errors.Join(err, statErr)
	}
	if removeErr := os.Remove(s.dir); removeErr != nil && !os.IsNotExist(removeErr) {
		err = errors.Join(err, removeErr)
	}
	s.closeErr = err
	return err
}

func (s *Server) Audit() Audit {
	if s == nil || s.self != s {
		return Audit{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	a := Audit{Events: append([]Event(nil), s.events...), Dropped: s.dropped}
	for i := range a.Events {
		e := &a.Events[i]
		if e.Before != nil {
			v := *e.Before
			e.Before = &v
		}
		if e.After != nil {
			v := *e.After
			e.After = &v
		}
	}
	return a
}
func (s *Server) audit(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.events) == MaxAuditEvents {
		copy(s.events, s.events[1:])
		s.events = s.events[:MaxAuditEvents-1]
		if s.dropped != ^uint64(0) {
			s.dropped++
		}
	}
	s.events = append(s.events, e)
}

func (s *Server) accept() {
	var workers sync.WaitGroup
	defer func() { workers.Wait(); close(s.done) }()
	for {
		conn, err := nextAccepted(s.ctx, func() (net.Conn, error) { return s.listener.AcceptUnix() }, acceptBackoff)
		if err != nil {
			if s.ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				s.Close()
			} else {
				s.closeWithCause(err)
			}
			return
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			conn.Close()
			continue
		}
		select {
		case s.credits <- struct{}{}:
		default:
			s.mu.Unlock()
			conn.Close()
			continue
		}
		s.conns[conn] = true
		s.mu.Unlock()
		lease := &clientLease{refs: 1, credits: s.credits}
		workers.Add(1)
		go func() { defer workers.Done(); s.serve(conn, lease) }()
	}
}

// A successful accept resets backoff. Sustained resource pressure never spins
// and remains cancellable; permanent errors retain a terminal cause for Err.
func nextAccepted(ctx context.Context, accept func() (net.Conn, error), wait func(context.Context, time.Duration) error) (net.Conn, error) {
	delay := 5 * time.Millisecond
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, err := accept()
		if err == nil {
			return conn, nil
		}
		if errors.Is(err, net.ErrClosed) || ctx.Err() != nil || !retryableAccept(err) {
			return nil, err
		}
		if err = wait(ctx, delay); err != nil {
			return nil, err
		}
		delay = min(delay*2, time.Second)
	}
}

func retryableAccept(err error) bool {
	for _, code := range []syscall.Errno{syscall.EINTR, syscall.EAGAIN, syscall.ECONNABORTED, syscall.EMFILE, syscall.ENFILE} {
		if errors.Is(err, code) {
			return true
		}
	}
	var temporary net.Error
	return errors.As(err, &temporary) && (temporary.Timeout() || temporary.Temporary())
}

func acceptBackoff(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type callState struct {
	mu                sync.Mutex
	terminal, started bool
	before            *uint64
}

func (c *callState) begin(ctx context.Context, revision uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.terminal || ctx.Err() != nil {
		return false
	}
	c.started = true
	c.before = &revision
	return true
}
func (c *callState) finish() (bool, *uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.terminal = true
	return c.started, c.before
}

type clientLease struct {
	mu      sync.Mutex
	refs    int
	credits chan struct{}
}

func (l *clientLease) retain() { l.mu.Lock(); l.refs++; l.mu.Unlock() }
func (l *clientLease) release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refs--
	if l.refs == 0 {
		<-l.credits
	}
}

func (s *Server) serve(conn net.Conn, lease *clientLease) {
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer lease.release()
	defer func() { cancel(); conn.Close(); s.mu.Lock(); delete(s.conns, conn); s.mu.Unlock() }()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	conn.SetDeadline(time.Now().Add(s.timeout))
	reader := bufio.NewReaderSize(conn, MaxRequestBytes+512)
	// Credential prelude is a fixed lowercase hex line, before any JSON parsing.
	credential, err := reader.ReadSlice('\n')
	if err != nil || len(credential) != 65 || subtle.ConstantTimeCompare(credential[:len(credential)-1], []byte(s.endpoint.Token)) != 1 {
		s.audit(Event{Code: "unauthorized"})
		return
	}
	raw, err := reader.ReadSlice('\n')
	if err != nil {
		s.audit(Event{Code: "invalid_request"})
		return
	}
	r, err := decodeRequest(raw[:len(raw)-1])
	if err != nil {
		s.write(conn, failure("", "invalid_request"))
		s.audit(Event{Code: "invalid_request"})
		return
	}
	state := &callState{}
	event := Event{ID: r.ID, Operation: r.Operation, Target: r.Target}
	response := failure(r.ID, "canceled")
	// Continue reading while waiting: EOF/disconnect or a second request cancels
	// the first. Clients keep their write side open until the response arrives.
	monitorDone := make(chan struct{})
	go func() { defer close(monitorDone); reader.ReadByte(); cancel() }()
	defer func() { conn.Close(); <-monitorDone }()
	// Publish a normal terminal observation before final EOF can reach Call.
	defer func() {
		event.Code = "ok"
		if response.Error != nil {
			event.Code = response.Error.Code
		}
		event.Started, event.Before = state.finish()
		event.After = response.Revision
		s.audit(event)
	}()
	result := make(chan Response, 1)
	// Disconnect never returns an admitted owner credit. App closure drops the
	// queued envelopes and may retire that credit; ordinary cancellation waits
	// for the callback's explicit skip/return.
	lease.retain()
	var releaseOnce sync.Once
	releaseOwner := func() { releaseOnce.Do(lease.release) }
	stopOwner := context.AfterFunc(s.owner.Context(), releaseOwner)
	err = s.owner.Post(func() {
		defer releaseOwner()
		defer stopOwner()
		if ctx.Err() != nil {
			result <- failure(r.ID, "canceled")
			return
		}
		revision, err := s.registry.revision()
		if err != nil {
			result <- failure(r.ID, "revision_unavailable")
			return
		}
		entries := s.registry.queries
		if r.Operation == "invoke" {
			entries = s.registry.actions
		}
		h, exists := entries[r.Target]
		if !exists {
			result <- failure(r.ID, "unknown_target")
			return
		}
		if r.ExpectedRevision != nil && *r.ExpectedRevision != revision {
			out := failure(r.ID, "stale_revision")
			out.Revision = &revision
			result <- out
			return
		}
		if !state.begin(ctx, revision) {
			result <- failure(r.ID, "canceled")
			return
		}
		value, err := h(ctx, r.Params)
		out := Response{Version: Version, ID: r.ID}
		if err != nil {
			code := "rejected"
			if errors.Is(err, errParams) {
				code = "invalid_params"
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				code = "canceled"
			}
			out = failure(r.ID, code)
		} else {
			encoded, marshalErr := json.Marshal(value)
			if marshalErr != nil || len(encoded) > MaxResponseBytes-1024 {
				out = failure(r.ID, "result_limit")
			} else {
				out.Result = encoded
			}
		}
		if after, err := s.registry.revision(); err == nil {
			out.Revision = &after
		}
		result <- out
	})
	if err != nil {
		stopOwner()
		releaseOwner()
		response = failure(r.ID, "unavailable")
		s.write(conn, response)
		return
	}
	select {
	case response = <-result:
		if err := s.write(conn, response); err != nil {
			response = failure(r.ID, "transport_error")
		}
	case <-ctx.Done():
	}
}

func (s *Server) write(conn net.Conn, r Response) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if len(data) > MaxResponseBytes {
		return errors.New("automation: response limit")
	}
	data = append(data, '\n')
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

// Call uses a new connection. Cancellation closes it; a returned transport error
// does not prove an action did not execute. Never automatically retry an invoke.
func Call(ctx context.Context, endpoint Endpoint, r Request) (Response, error) {
	return callWithDial(ctx, endpoint, r, dialLocal)
}

func callWithDial(ctx context.Context, endpoint Endpoint, r Request, dial func(context.Context, string) (net.Conn, error)) (Response, error) {
	if ctx == nil || endpoint.Version != Version || len(endpoint.Token) != 64 {
		return Response{}, errors.New("automation: invalid client configuration")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if len(r.Params) > MaxRequestBytes || !identity(r.ID) || !identity(r.Target) {
		return Response{}, errors.New("automation: invalid client request")
	}
	if _, err := hex.DecodeString(endpoint.Token); err != nil {
		return Response{}, err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return Response{}, err
	}
	if _, err = decodeRequest(data); err != nil {
		return Response{}, err
	}
	conn, err := dial(ctx, endpoint.Path)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	conn.SetDeadline(deadline)
	wire := append([]byte(endpoint.Token+"\n"), data...)
	wire = append(wire, '\n')
	for len(wire) > 0 {
		n, err := conn.Write(wire)
		if err != nil {
			return Response{}, err
		}
		wire = wire[n:]
	}
	reader := bufio.NewReaderSize(conn, MaxResponseBytes+1)
	raw, err := reader.ReadSlice('\n')
	if err != nil {
		return Response{}, err
	}
	if len(raw)-1 > MaxResponseBytes {
		return Response{}, errors.New("automation: response limit")
	}
	var response Response
	if err = strictJSON(raw[:len(raw)-1], &response); err != nil {
		return Response{}, err
	}
	if response.Version != Version || response.ID != r.ID || response.Error == nil && response.Result == nil || response.Error != nil && response.Result != nil {
		return Response{}, errors.New("automation: invalid response")
	}
	if trailing, err := reader.ReadByte(); err != io.EOF {
		return Response{}, errors.New("automation: trailing response data " + string(trailing))
	}
	return response, nil
}
