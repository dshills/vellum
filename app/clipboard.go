package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	ClipboardMaxFormats             = 16
	ClipboardMaxMIMEBytes           = 127
	ClipboardMaxRepresentationBytes = 8 << 20
	ClipboardMaxTotalBytes          = 16 << 20
)

var (
	ErrClipboardBusy     = errors.New("vellum: clipboard dispatch capacity exhausted")
	ErrClipboardChanged  = errors.New("vellum: clipboard changed during read")
	ErrClipboardTooLarge = errors.New("vellum: clipboard data exceeds limit")
)

// ClipboardRepresentation is one exact byte representation of one clipboard
// item. MIME is a lowercase concrete media type without parameters. PNG and HTML
// are transported as bytes; the framework does not decode or sanitize their data.
// Do not mutate input slices until the service method returns. Returned slices
// are detached from native storage and owned by the caller.
type ClipboardRepresentation struct {
	MIME string
	Data []byte
}

// ClipboardCapabilities describes implemented transports, not runtime acceptance.
// CustomMIME maps valid types through the OS MIME registry; an unmappable type or
// two MIME names mapping to one native type return ErrUnavailable before a write.
type ClipboardCapabilities struct{ Text, PNG, HTML, CustomMIME bool }

// ClipboardServices is an optional extension of Services. Call from workers,
// outside UI-owner callbacks. Reads request explicit formats from the first native
// item, preserve request order, and omit absent representations. Writes replace
// the clipboard with one item containing all representations. Cancellation can
// race an already executing OS call; writes cannot promise rollback or atomicity.
type ClipboardServices interface {
	ClipboardCapabilities() ClipboardCapabilities
	ReadClipboardData(context.Context, []string) ([]ClipboardRepresentation, error)
	WriteClipboardData(context.Context, []ClipboardRepresentation) error
}

func (a *Application) Clipboard() (ClipboardServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	c, ok := s.(ClipboardServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return c, nil
}

// ValidateClipboardFormats validates the bounded request without copying data.
func ValidateClipboardFormats(formats []string) error {
	if len(formats) == 0 || len(formats) > ClipboardMaxFormats {
		return fmt.Errorf("vellum: clipboard needs 1..%d formats", ClipboardMaxFormats)
	}
	seen := make(map[string]bool, len(formats))
	for _, mime := range formats {
		if len(mime) > ClipboardMaxMIMEBytes || strings.Count(mime, "/") != 1 {
			return fmt.Errorf("vellum: invalid clipboard MIME type")
		}
		parts := strings.Split(mime, "/")
		for _, part := range parts {
			if part == "" {
				return fmt.Errorf("vellum: invalid clipboard MIME type")
			}
			for _, c := range part {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'+-.^_`|~", c)) {
					return fmt.Errorf("vellum: invalid clipboard MIME type")
				}
			}
		}
		if seen[mime] {
			return fmt.Errorf("vellum: duplicate clipboard MIME type")
		}
		seen[mime] = true
	}
	return nil
}

// ValidateClipboardData checks lengths before any payload copying. Plain text
// follows the existing Services UTF-8/no-NUL contract; other formats are opaque.
func ValidateClipboardData(data []ClipboardRepresentation) error {
	if len(data) == 0 || len(data) > ClipboardMaxFormats {
		return fmt.Errorf("vellum: invalid clipboard representation count")
	}
	formats := make([]string, len(data))
	total := 0
	for i, r := range data {
		formats[i] = r.MIME
		if len(r.Data) > ClipboardMaxRepresentationBytes || len(r.Data) > ClipboardMaxTotalBytes-total {
			return ErrClipboardTooLarge
		}
		total += len(r.Data)
		if r.MIME == "text/plain" && (!utf8.Valid(r.Data) || bytes.IndexByte(r.Data, 0) >= 0) {
			return fmt.Errorf("vellum: invalid clipboard text")
		}
	}
	return ValidateClipboardFormats(formats)
}
