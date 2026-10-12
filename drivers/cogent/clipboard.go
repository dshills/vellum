package cogent

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/dshills/vellum/app"
)

const clipboardWireMax = 4 + app.ClipboardMaxFormats*(6+app.ClipboardMaxMIMEBytes) + app.ClipboardMaxTotalBytes
const clipboardPendingMax = 4

func (s *platformServices) ClipboardCapabilities() app.ClipboardCapabilities {
	return nativeClipboardCapabilities()
}
func (s *platformServices) ReadClipboardData(ctx context.Context, formats []string) ([]app.ClipboardRepresentation, error) {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	if err := app.ValidateClipboardFormats(formats); err != nil {
		return nil, err
	}
	request := make([]app.ClipboardRepresentation, len(formats))
	for i, mime := range formats {
		request[i].MIME = mime
	}
	wire, err := nativeClipboardData(scope, request, false)
	if err != nil {
		return nil, err
	}
	return decodeClipboardResponse(wire, formats)
}

func decodeClipboardResponse(wire []byte, formats []string) ([]app.ClipboardRepresentation, error) {
	data, err := decodeClipboardWire(wire)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err := app.ValidateClipboardData(data); err != nil {
			return nil, err
		}
	}
	// A native response must be an ordered subset of the requested formats.
	position := 0
	for _, r := range data {
		for position < len(formats) && formats[position] != r.MIME {
			position++
		}
		if position == len(formats) {
			return nil, fmt.Errorf("cogent: unexpected clipboard representation")
		}
		position++
	}
	return data, nil
}
func (s *platformServices) WriteClipboardData(ctx context.Context, data []app.ClipboardRepresentation) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err := app.ValidateClipboardData(data); err != nil {
		return err
	}
	_, err = nativeClipboardData(scope, data, true)
	return err
}

// Wire: little-endian count:u32, then MIME-length:u16, data-length:u32,
// ASCII MIME and opaque data. All producers and consumers enforce the same caps.
func encodeClipboardWire(data []app.ClipboardRepresentation) []byte {
	n := 4
	for _, r := range data {
		n += 6 + len(r.MIME) + len(r.Data)
	}
	wire := make([]byte, n)
	binary.LittleEndian.PutUint32(wire, uint32(len(data)))
	p := 4
	for _, r := range data {
		binary.LittleEndian.PutUint16(wire[p:], uint16(len(r.MIME)))
		binary.LittleEndian.PutUint32(wire[p+2:], uint32(len(r.Data)))
		p += 6
		p += copy(wire[p:], r.MIME)
		p += copy(wire[p:], r.Data)
	}
	return wire
}
func decodeClipboardWire(wire []byte) ([]app.ClipboardRepresentation, error) {
	bad := fmt.Errorf("cogent: invalid clipboard wire")
	if len(wire) < 4 || len(wire) > clipboardWireMax {
		return nil, bad
	}
	count32 := binary.LittleEndian.Uint32(wire)
	if count32 > app.ClipboardMaxFormats {
		return nil, bad
	}
	count := int(count32)
	data := make([]app.ClipboardRepresentation, 0, count)
	p, total := 4, 0
	for i := 0; i < count; i++ {
		if len(wire)-p < 6 {
			return nil, bad
		}
		m := int(binary.LittleEndian.Uint16(wire[p:]))
		n32 := binary.LittleEndian.Uint32(wire[p+2:])
		if n32 > app.ClipboardMaxRepresentationBytes {
			return nil, bad
		}
		n := int(n32)
		p += 6
		if m == 0 || m > app.ClipboardMaxMIMEBytes || n > app.ClipboardMaxRepresentationBytes || n > app.ClipboardMaxTotalBytes-total || m > len(wire)-p {
			return nil, bad
		}
		mime := string(wire[p : p+m])
		p += m
		if n > len(wire)-p {
			return nil, bad
		}
		data = append(data, app.ClipboardRepresentation{MIME: mime, Data: append([]byte(nil), wire[p:p+n]...)})
		p += n
		total += n
	}
	if p != len(wire) {
		return nil, bad
	}
	return data, nil
}

type clipboardResult struct {
	wire []byte
	err  error
}
type clipboardOperation struct {
	ctx    context.Context
	result chan clipboardResult
}
type clipboardDispatcher struct {
	sync.Mutex
	next    uint64
	closed  bool
	pending map[uint64]clipboardOperation
}

var clipboardDispatch clipboardDispatcher

// Canceled callers keep their credit until the native block acknowledges it.
// A stalled main queue therefore cannot accumulate unlimited retained payloads.
func (d *clipboardDispatcher) run(ctx context.Context, start func(uint64)) ([]byte, error) {
	if err := nativeServiceContextErr(ctx); err != nil {
		return nil, err
	}
	d.Lock()
	if d.closed {
		d.Unlock()
		return nil, app.ErrClosed
	}
	if len(d.pending) >= clipboardPendingMax {
		d.Unlock()
		return nil, app.ErrClipboardBusy
	}
	if d.next == ^uint64(0) {
		d.Unlock()
		return nil, app.ErrUnavailable
	}
	d.next++
	token := d.next
	op := clipboardOperation{ctx: ctx, result: make(chan clipboardResult, 1)}
	if d.pending == nil {
		d.pending = make(map[uint64]clipboardOperation)
	}
	d.pending[token] = op
	d.Unlock()
	start(token)
	select {
	case r := <-op.result:
		if err := nativeServiceContextErr(ctx); err != nil {
			return nil, err
		}
		return r.wire, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (d *clipboardDispatcher) active(token uint64) bool {
	d.Lock()
	defer d.Unlock()
	op, exists := d.pending[token]
	return exists && !d.closed && nativeServiceContextErr(op.ctx) == nil
}
func (d *clipboardDispatcher) finish(token uint64, wire []byte, err error) {
	d.Lock()
	defer d.Unlock()
	op, exists := d.pending[token]
	if !exists {
		return
	}
	delete(d.pending, token)
	if cause := nativeServiceContextErr(op.ctx); cause != nil {
		err = cause
		wire = nil
	}
	select {
	case op.result <- clipboardResult{wire, err}:
	default:
	}
}
func (d *clipboardDispatcher) stop() {
	d.Lock()
	defer d.Unlock()
	d.closed = true
	for _, op := range d.pending {
		select {
		case op.result <- clipboardResult{err: app.ErrClosed}:
		default:
		}
	}
}
