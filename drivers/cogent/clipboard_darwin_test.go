//go:build darwin && cgo

package cogent

import (
	"bytes"
	"encoding/binary"
	"github.com/dshills/vellum/app"
	"testing"
)

// These calls resolve MIME types and parse native bytes, never NSPasteboard.
func TestClipboardObjectiveCParserOffscreen(t *testing.T) {
	valid := encodeClipboardWire([]app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte("Résumé 😀")}, {MIME: "image/png", Data: []byte{0, 255}}, {MIME: "text/html", Data: []byte("<b>Hello</b>")}, {MIME: "application/vnd.vellum+json", Data: []byte{1, 0, 2}}})
	if !nativeClipboardValidate(valid) {
		t.Fatal("valid native wire rejected")
	}
	for n := 0; n < len(valid); n++ {
		if nativeClipboardValidate(valid[:n]) {
			t.Fatalf("native truncation %d", n)
		}
	}
	for _, data := range [][]app.ClipboardRepresentation{
		{{MIME: "Text/plain"}}, {{MIME: "text/plain;charset=utf-8"}}, {{MIME: "text/plain", Data: []byte{0}}}, {{MIME: "text/plain", Data: []byte{255}}}, {{MIME: "text/plain"}, {MIME: "text/plain"}},
	} {
		if nativeClipboardValidate(encodeClipboardWire(data)) {
			t.Fatal("bad native wire accepted", data)
		}
	}
	for _, mutate := range []func([]byte){func(b []byte) { binary.LittleEndian.PutUint32(b, 17) }, func(b []byte) { binary.LittleEndian.PutUint16(b[4:], 128) }, func(b []byte) { binary.LittleEndian.PutUint32(b[6:], app.ClipboardMaxRepresentationBytes+1) }} {
		b := bytes.Clone(valid)
		mutate(b)
		if nativeClipboardValidate(b) {
			t.Fatal("bad native length accepted")
		}
	}
	if nativeClipboardValidate(append(valid, 0)) {
		t.Fatal("native trailing byte accepted")
	}
	if nativeClipboardValidate([]byte{0, 0, 0, 0}) {
		t.Fatal("native empty request accepted")
	}
}

func TestClipboardObjectiveCOwnershipAndNativeAliasesOffscreen(t *testing.T) {
	wire := encodeClipboardWire([]app.ClipboardRepresentation{{MIME: "application/one", Data: []byte{0, 255, 1}}, {MIME: "application/two", Data: []byte{2, 0}}})
	if nativeClipboardAliasValidate(wire) {
		t.Fatal("distinct MIME aliases accepted")
	}
	if got := nativeClipboardPayloadFixture(wire); got != 7 {
		t.Fatalf("producer/captured-consumer ownership flags=%d", got)
	}
}
