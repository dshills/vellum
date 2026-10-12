package cogent

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/dshills/vellum/app"
	"reflect"
	"testing"
	"time"
)

func TestClipboardWireDetachedAndBounded(t *testing.T) {
	want := []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte("Résumé 😀")}, {MIME: "application/vnd.vellum+json", Data: []byte{0, 255, 1}}}
	wire := encodeClipboardWire(want)
	got, err := decodeClipboardWire(wire)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal(got, err)
	}
	for i := range wire {
		wire[i] = 0
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("returned bytes alias wire")
	}
	valid := encodeClipboardWire(want)
	for n := 0; n < len(valid); n++ {
		if _, err := decodeClipboardWire(valid[:n]); err == nil {
			t.Fatalf("truncation %d", n)
		}
	}
	for _, mutate := range []func([]byte){
		func(b []byte) { binary.LittleEndian.PutUint32(b, 17) },
		func(b []byte) { binary.LittleEndian.PutUint32(b, ^uint32(0)) },
		func(b []byte) { binary.LittleEndian.PutUint16(b[4:], 128) },
		func(b []byte) { binary.LittleEndian.PutUint32(b[6:], app.ClipboardMaxRepresentationBytes+1) },
		func(b []byte) { binary.LittleEndian.PutUint32(b[6:], ^uint32(0)) },
	} {
		b := bytes.Clone(valid)
		mutate(b)
		if _, err := decodeClipboardWire(b); err == nil {
			t.Fatal("invalid length accepted")
		}
	}
	if _, err := decodeClipboardWire(append(valid, 0)); err == nil {
		t.Fatal("trailing byte accepted")
	}
	if got, err := decodeClipboardWire([]byte{0, 0, 0, 0}); err != nil || len(got) != 0 {
		t.Fatal(got, err)
	}
}

func TestClipboardCanceledQueueRetainsAdmissionUntilAcknowledged(t *testing.T) {
	var d clipboardDispatcher
	tokens := make([]uint64, clipboardPendingMax)
	for i := range tokens {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := d.run(ctx, func(token uint64) { tokens[i] = token; cancel() })
		if !errors.Is(err, context.Canceled) || d.active(tokens[i]) {
			t.Fatal(err)
		}
	}
	if _, err := d.run(context.Background(), func(uint64) { t.Fatal("over capacity dispatched") }); !errors.Is(err, app.ErrClipboardBusy) {
		t.Fatal(err)
	}
	d.finish(tokens[0], nil, nil)
	want := []byte{1, 2, 3}
	got, err := d.run(context.Background(), func(token uint64) { d.finish(token, want, nil) })
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal(got, err)
	}
	for _, token := range tokens {
		d.finish(token, nil, nil)
	}
	if len(d.pending) != 0 {
		t.Fatal("credits retained")
	}
}
func TestClipboardShutdownAndAppLifetime(t *testing.T) {
	var d clipboardDispatcher
	started := make(chan uint64, 1)
	done := make(chan error, 1)
	go func() { _, err := d.run(context.Background(), func(token uint64) { started <- token }); done <- err }()
	token := <-started
	d.stop()
	select {
	case err := <-done:
		if !errors.Is(err, app.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown stranded caller")
	}
	if d.active(token) {
		t.Fatal("active after shutdown")
	}
	if _, err := d.run(context.Background(), func(uint64) { t.Fatal("dispatch after shutdown") }); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
	d.finish(token, nil, nil)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	s := &platformServices{owner: a}
	ctx, cancel, err := s.serviceContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	var other clipboardDispatcher
	_, err = other.run(ctx, func(token uint64) {
		a.Close()
		if other.active(token) {
			t.Error("app-close active")
		}
		other.finish(token, nil, nil)
	})
	if err == nil {
		t.Fatal("app shutdown ignored")
	}
}
func TestClipboardPreflightNeverAccessesNativePasteboard(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	s := &platformServices{owner: a}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.ReadClipboardData(ctx, []string{"text/plain"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.WriteClipboardData(ctx, []app.ClipboardRepresentation{{MIME: "text/plain"}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := s.ReadClipboardData(context.Background(), []string{"Text/plain"}); err == nil {
		t.Fatal("bad MIME")
	}
	if err := s.WriteClipboardData(context.Background(), []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte{0}}}); err == nil {
		t.Fatal("bad text")
	}
	a.Close()
	if _, err := s.ReadClipboardData(context.Background(), []string{"text/plain"}); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}

func TestClipboardResponseOrderAndTextValidation(t *testing.T) {
	formats := []string{"image/png", "text/html", "text/plain"}
	for _, bad := range [][]app.ClipboardRepresentation{
		{{MIME: "text/plain"}, {MIME: "image/png"}}, {{MIME: "a/unknown"}}, {{MIME: "text/plain"}, {MIME: "text/plain"}}, {{MIME: "text/plain", Data: []byte{255}}},
	} {
		if _, err := decodeClipboardResponse(encodeClipboardWire(bad), formats); err == nil {
			t.Fatal("bad response accepted", bad)
		}
	}
	got, err := decodeClipboardResponse(encodeClipboardWire([]app.ClipboardRepresentation{{MIME: "text/html", Data: []byte("<b>Hi</b>")}, {MIME: "text/plain", Data: []byte("Hi")}}), formats)
	if err != nil || len(got) != 2 {
		t.Fatal(got, err)
	}
}
