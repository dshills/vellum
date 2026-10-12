package automation

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStrictRequestRefusals(t *testing.T) {
	base := `{"version":1,"id":"request.1","operation":"inspect","target":"component","params":{}}`
	if _, err := decodeRequest([]byte(base)); err != nil {
		t.Fatal(err)
	}
	bad := []string{
		strings.Replace(base, `"version":1`, `"version":2`, 1),
		strings.Replace(base, `"id":"request.1"`, `"id":"request.1","id":"other"`, 1),
		strings.Replace(base, `"params":{}`, `"params":{"nested":{"a":1,"a":2}}`, 1),
		strings.Replace(base, `"version":1`, `"Version":1`, 1),
		strings.Replace(base, `"params":{}`, `"params":null`, 1),
		strings.Replace(base, `"params":{}`, `"params":[]`, 1),
		strings.Replace(base, `"params":{}`, `"params":{},"expected_revision":null`, 1),
		strings.Replace(base, `"operation":"inspect"`, `"operation":"invoke"`, 1),
		strings.Replace(base, `"operation":"inspect"`, `"operation":"unknown"`, 1),
		strings.Replace(base, `"id":"request.1"`, `"id":"bad id"`, 1),
		base + ` {}`, strings.Replace(base, `{}}`, `{"n":`+strings.Repeat("[", 33)+`0`+strings.Repeat("]", 33)+`}}`, 1),
		strings.Replace(base, `"params":{}`, `"params":{"x":"`+string([]byte{0xff})+`"}`, 1),
		strings.Replace(base, `"params":{}`, `"params":{"x":"\ud800"}`, 1),
		strings.Replace(base, `"params":{}`, `"params":{"x":"\udc00"}`, 1),
		strings.Replace(base, `"params":{}`, `"params":{"x":"`+strings.Repeat("a", MaxRequestBytes)+`"}`, 1),
	}
	for _, params := range []string{`{"x":"\ud83d\ude00"}`, `{"x":"\\ud800"}`, `{"x":"😀"}`} {
		if _, err := decodeRequest([]byte(strings.Replace(base, `"params":{}`, `"params":`+params, 1))); err != nil {
			t.Fatal("valid scalar/literal", params, err)
		}
	}
	for i, raw := range bad {
		if _, err := decodeRequest([]byte(raw)); err == nil {
			t.Fatalf("accepted invalid request %d", i)
		}
	}
	var params struct {
		Number int `json:"number"`
	}
	for _, raw := range []string{`{"number":2,"unknown":3}`, `{"number":"2"}`, `{"number":1} {}`} {
		if strictJSON([]byte(raw), &params) == nil {
			t.Fatal(raw)
		}
	}
	// Caller-owned bytes are detached by wire decoding.
	raw := []byte(base)
	r, err := decodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-2] = 'x'
	if !json.Valid(r.Params) {
		t.Fatal("borrowed params")
	}
}

func TestClientDeadlineIncludesDial(t *testing.T) {
	ep := Endpoint{Path: "unused", Token: strings.Repeat("a", 64), Version: Version}
	req := Request{Version: Version, ID: "r", Operation: "inspect", Target: "x", Params: json.RawMessage(`{}`)}
	// This observer never dials or waits: deadline admission is inspected directly.
	for _, parent := range []context.Context{context.Background(), func() context.Context {
		c, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
		t.Cleanup(cancel)
		return c
	}()} {
		start := time.Now()
		errFixture := errors.New("dial observer")
		_, err := callWithDial(parent, ep, req, func(ctx context.Context, _ string) (net.Conn, error) {
			deadline, ok := ctx.Deadline()
			if !ok || deadline.After(time.Now().Add(time.Minute)) || deadline.Before(start) {
				t.Error("dial without maximum deadline")
			}
			if expected, ok := parent.Deadline(); ok && !deadline.Equal(expected) {
				t.Error("caller deadline changed")
			}
			return nil, errFixture
		})
		if !errors.Is(err, errFixture) {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := callWithDial(ctx, ep, req, func(c context.Context, _ string) (net.Conn, error) { return nil, c.Err() })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRegistryRefusesUninitializedAndInvalidBindings(t *testing.T) {
	query := func(context.Context, struct{}) (any, error) { return struct{}{}, nil }
	for _, r := range []*Registry{nil, {}, NewRegistry(nil)} {
		if RegisterQuery(r, "x", query) == nil {
			t.Fatal("invalid registry")
		}
	}
	r := NewRegistry(func() (uint64, error) { return 1, nil })
	if err := RegisterQuery(r, "x", query); err != nil {
		t.Fatal(err)
	}
	if RegisterQuery(r, "x", query) == nil || RegisterQuery(r, "bad id", query) == nil || RegisterQuery[struct{}](r, "nil", nil) == nil {
		t.Fatal("invalid binding")
	}
}
