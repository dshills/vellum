// Package automation exposes explicitly registered application semantics over
// an opt-in, capability-protected local endpoint. It never synthesizes UI input.
package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dshills/vellum/command"
)

const (
	Version          = 1
	MaxRequestBytes  = 32 << 10
	MaxResponseBytes = 64 << 10
	MaxClients       = 8
	MaxAuditEvents   = 256
)

var ErrUnsupported = errors.New("automation: local endpoint requires macOS")

// Request IDs correlate one connection; they do not provide retry deduplication.
// Params must be a JSON object, even for argument-free commands (use {}).
type Request struct {
	Version          int             `json:"version"`
	ID               string          `json:"id"`
	Operation        string          `json:"operation"`
	Target           string          `json:"target"`
	ExpectedRevision *uint64         `json:"expected_revision,omitempty"`
	Params           json.RawMessage `json:"params"`
}

type Failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Response struct {
	Version  int             `json:"version"`
	ID       string          `json:"id"`
	Revision *uint64         `json:"revision,omitempty"`
	Result   json.RawMessage `json:"result,omitempty"`
	Error    *Failure        `json:"error,omitempty"`
}

type handler func(context.Context, json.RawMessage) (any, error)

// Registry is populated before Start. Start freezes it; subsequent registration
// refuses. Callbacks and Revision run only on the application owner. Revision
// must change whenever any exposed state or command availability changes.
type Registry struct {
	self             *Registry
	mu               sync.Mutex
	frozen           bool
	revision         func() (uint64, error)
	queries, actions map[string]handler
}

func NewRegistry(revision func() (uint64, error)) *Registry {
	r := &Registry{revision: revision, queries: map[string]handler{}, actions: map[string]handler{}}
	r.self = r
	return r
}

// RegisterQuery registers a bounded, detached, explicitly redacted DTO query.
// Application callbacks own domain validation, result size and their allocations.
func RegisterQuery[T any](r *Registry, target string, fn func(context.Context, T) (any, error)) error {
	return register(r, target, fn, false)
}

// RegisterAction registers a typed domain action. Reuse the same validation/action
// function from UI entry points; cancellation is cooperative, without rollback.
func RegisterAction[T any](r *Registry, target string, fn func(context.Context, T) (any, error)) error {
	return register(r, target, fn, true)
}

// RegisterCommand invokes the existing registry, including its availability
// check. Only this named command becomes externally callable.
func RegisterCommand(r *Registry, commands *command.Registry, id string) error {
	if commands == nil {
		return errors.New("automation: command registry required")
	}
	return RegisterAction(r, id, func(ctx context.Context, _ struct{}) (any, error) {
		return struct{}{}, commands.Invoke(ctx, id)
	})
}

func register[T any](r *Registry, target string, fn func(context.Context, T) (any, error), action bool) error {
	if r == nil || r.self != r || !identity(target) || fn == nil {
		return errors.New("automation: registry, target and callback required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.revision == nil || r.queries == nil || r.actions == nil {
		return errors.New("automation: use NewRegistry with a revision callback")
	}
	if r.frozen {
		return errors.New("automation: registry is frozen")
	}
	entries := r.queries
	if action {
		entries = r.actions
	}
	if _, exists := entries[target]; exists {
		return errors.New("automation: duplicate target")
	}
	if len(r.queries)+len(r.actions) >= 256 {
		return errors.New("automation: registry limit")
	}
	entries[target] = func(ctx context.Context, raw json.RawMessage) (any, error) {
		var params T
		if err := strictJSON(raw, &params); err != nil {
			return nil, errParams
		}
		return fn(ctx, params)
	}
	return nil
}

func identity(s string) bool {
	if len(s) < 1 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("._:-", c)) {
			return false
		}
	}
	return true
}

var errParams = errors.New("automation: invalid parameters")

// strictJSON rejects duplicate keys at any depth before typed decoding. JSON
// decoding's UTF-8 replacement and case-insensitive field matching are also
// refused for the wire Request by its canonical key allowlist.
func strictJSON(raw []byte, output any) error {
	if !utf8.Valid(raw) {
		return errors.New("invalid UTF-8")
	}
	if !validEscapedScalars(raw) {
		return errors.New("invalid Unicode escape")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 32 {
			return errors.New("JSON depth")
		}
		t, err := d.Token()
		if err != nil {
			return err
		}
		if delim, ok := t.(json.Delim); ok {
			switch delim {
			case '{':
				keys := map[string]bool{}
				for d.More() {
					key, err := d.Token()
					if err != nil {
						return err
					}
					s, ok := key.(string)
					if !ok || keys[s] {
						return errors.New("duplicate JSON key")
					}
					keys[s] = true
					if err := walk(depth + 1); err != nil {
						return err
					}
				}
			case '[':
				for d.More() {
					if err := walk(depth + 1); err != nil {
						return err
					}
				}
			default:
				return errors.New("invalid JSON delimiter")
			}
			_, err = d.Token()
			return err
		}
		return nil
	}
	if err := walk(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing JSON")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(output)
}

// Refuse JSON's lossy replacement of lone escaped UTF-16 surrogates.
func validEscapedScalars(raw []byte) bool {
	inString := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) {
			return false
		}
		if raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return false
		}
		n, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if n >= 0xDC00 && n <= 0xDFFF {
			return false
		}
		if n >= 0xD800 && n <= 0xDBFF {
			if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
			if err != nil || low < 0xDC00 || low > 0xDFFF {
				return false
			}
			i += 6
		}
	}
	return true
}

func decodeRequest(raw []byte) (Request, error) {
	var r Request
	if len(raw) > MaxRequestBytes {
		return r, errors.New("request limit")
	}
	if err := strictJSON(raw, &r); err != nil {
		return Request{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Request{}, err
	}
	for key := range fields {
		switch key {
		case "version", "id", "operation", "target", "expected_revision", "params":
		default:
			return Request{}, errors.New("noncanonical field")
		}
	}
	if r.Version != Version || !identity(r.ID) || !identity(r.Target) || !object(r.Params) {
		return Request{}, errors.New("invalid request")
	}
	if r.Operation != "inspect" && r.Operation != "invoke" || r.Operation == "invoke" && r.ExpectedRevision == nil || r.Operation == "inspect" && (r.ExpectedRevision != nil || fields["expected_revision"] != nil) {
		return Request{}, errors.New("invalid operation fields")
	}
	return r, nil
}

func object(raw []byte) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) >= 2 && raw[0] == '{' && raw[len(raw)-1] == '}'
}

func failure(id, code string) Response {
	// Never transmit callback/error text, including filesystem paths and secrets.
	return Response{Version: Version, ID: id, Error: &Failure{Code: code, Message: fmt.Sprintf("Request %s", code)}}
}
