// Package replay defines bounded, portable input records for explicitly stepped
// application tests. It does not replace native routing or advance clocks itself.
package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const Version = 1

var (
	ErrInvalid  = errors.New("replay: invalid record")
	ErrCapacity = errors.New("replay: record limit reached")
)

type Limits struct {
	MaxOperations   int
	MaxBytes        int
	MaxPayloadBytes int
}

func (l Limits) normalized() (Limits, error) {
	if l.MaxOperations == 0 {
		l.MaxOperations = 10000
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 8 << 20
	}
	if l.MaxPayloadBytes == 0 {
		l.MaxPayloadBytes = 64 << 10
	}
	if l.MaxOperations < 1 || l.MaxOperations > 1000000 || l.MaxBytes < 1 || l.MaxBytes > 64<<20 || l.MaxPayloadBytes < 1 || l.MaxPayloadBytes > 1<<20 || l.MaxPayloadBytes > l.MaxBytes {
		return l, ErrInvalid
	}
	return l, nil
}

// Window fixes the logical viewport and initial target lifetime. Physical pixels
// are derived by the backend from DPI; a backend must refuse unsupported values.
type Window struct {
	ID          string `json:"id"`
	Incarnation uint64 `json:"incarnation"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	DPI         int    `json:"dpi"`
}
type Font struct {
	Family string `json:"family"`
	SHA256 string `json:"sha256"`
}

// Manifest records inputs required to reproduce a fixture. SHA256 identities
// refer to supplied fixture/font bytes; a backend checks them before starting.
// Records contain explicit input text, but never implicitly read OS clipboard,
// file paths, native pointers or ambient fonts/preferences.
type Manifest struct {
	Version         int       `json:"version"`
	FixtureSHA256   string    `json:"fixture_sha256"`
	LayoutVersion   string    `json:"layout_version"`
	RendererVersion string    `json:"renderer_version"`
	ThemeSHA256     string    `json:"theme_sha256"`
	Start           time.Time `json:"start"`
	Seed            uint64    `json:"seed"`
	Windows         []Window  `json:"windows"`
	Fonts           []Font    `json:"fonts"`
}

type Kind string

const (
	PointerDown Kind = "pointer-down"
	PointerUp   Kind = "pointer-up"
	PointerMove Kind = "pointer-move"
	KeyDown     Kind = "key-down"
	KeyUp       Kind = "key-up"
	KeyChord    Kind = "key-chord"
	Command     Kind = "command"
	Semantic    Kind = "semantic"
	Advance     Kind = "advance"
	Worker      Kind = "worker"
	Frame       Kind = "frame"
	Capture     Kind = "capture"
)

type Argument struct{ Key, Value string }

// Operation carries values rather than callbacks. At is elapsed logical time
// for validation/diagnostics; only Advance explicitly moves the backend clock.
// Sequence is contiguous from one. Window/Target incarnation and expected
// revision are checked by the live backend immediately before routing.
// Arguments carry typed command/semantic string parameters, copied on retention.
type Operation struct {
	Sequence          uint64        `json:"sequence"`
	At                time.Duration `json:"at"`
	Kind              Kind          `json:"kind"`
	Window            string        `json:"window,omitempty"`
	WindowIncarnation uint64        `json:"window_incarnation,omitempty"`
	Target            string        `json:"target,omitempty"`
	TargetIncarnation uint64        `json:"target_incarnation,omitempty"`
	CheckRevision     bool          `json:"check_revision,omitempty"`
	ExpectedRevision  uint64        `json:"expected_revision,omitempty"`
	X                 float64       `json:"x,omitempty"`
	Y                 float64       `json:"y,omitempty"`
	Button            int           `json:"button,omitempty"`
	Key               string        `json:"key,omitempty"`
	Modifiers         uint32        `json:"modifiers,omitempty"`
	Action            string        `json:"action,omitempty"`
	Arguments         []Argument    `json:"arguments,omitempty"`
	Delta             time.Duration `json:"delta,omitempty"`
	WorkerID          uint64        `json:"worker_id,omitempty"`
	Label             string        `json:"label,omitempty"`
}
type Record struct {
	Manifest   Manifest    `json:"manifest"`
	Operations []Operation `json:"operations"`
}

func text(s string, max int, empty bool) bool {
	return (empty || s != "") && len(s) <= max && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func hash(s string) bool {
	if len(s) != 2*sha256.Size {
		return false
	}
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}
func (m Manifest) validate() error {
	if m.Version != Version || !hash(m.FixtureSHA256) || !hash(m.ThemeSHA256) || !text(m.LayoutVersion, 128, false) || !text(m.RendererVersion, 128, false) || m.Start.IsZero() || m.Start.Location() != time.UTC || m.Start.Year() < 1 || m.Start.Year() > 9999 || len(m.Windows) < 1 || len(m.Windows) > 32 || len(m.Fonts) < 1 || len(m.Fonts) > 128 {
		return fmt.Errorf("%w: manifest", ErrInvalid)
	}
	ids := map[string]bool{}
	for _, w := range m.Windows {
		if !text(w.ID, 256, false) || ids[w.ID] || w.Incarnation == 0 || w.Width < 1 || w.Width > 32768 || w.Height < 1 || w.Height > 32768 || w.DPI < 1 || w.DPI > 1536 {
			return fmt.Errorf("%w: window", ErrInvalid)
		}
		ids[w.ID] = true
	}
	families := map[string]bool{}
	for _, f := range m.Fonts {
		if !text(f.Family, 256, false) || families[f.Family] || !hash(f.SHA256) {
			return fmt.Errorf("%w: font", ErrInvalid)
		}
		families[f.Family] = true
	}
	return nil
}
func cloneManifest(m Manifest) Manifest {
	m.FixtureSHA256 = strings.Clone(m.FixtureSHA256)
	m.LayoutVersion = strings.Clone(m.LayoutVersion)
	m.RendererVersion = strings.Clone(m.RendererVersion)
	m.ThemeSHA256 = strings.Clone(m.ThemeSHA256)
	m.Windows = append([]Window(nil), m.Windows...)
	for i := range m.Windows {
		m.Windows[i].ID = strings.Clone(m.Windows[i].ID)
	}
	m.Fonts = append([]Font(nil), m.Fonts...)
	for i := range m.Fonts {
		m.Fonts[i].Family = strings.Clone(m.Fonts[i].Family)
		m.Fonts[i].SHA256 = strings.Clone(m.Fonts[i].SHA256)
	}
	return m
}
func cloneOperation(o Operation) Operation {
	o.Kind = Kind(strings.Clone(string(o.Kind)))
	o.Window = strings.Clone(o.Window)
	o.Target = strings.Clone(o.Target)
	o.Key = strings.Clone(o.Key)
	o.Action = strings.Clone(o.Action)
	o.Label = strings.Clone(o.Label)
	o.Arguments = append([]Argument(nil), o.Arguments...)
	for i := range o.Arguments {
		o.Arguments[i].Key = strings.Clone(o.Arguments[i].Key)
		o.Arguments[i].Value = strings.Clone(o.Arguments[i].Value)
	}
	return o
}

func (o Operation) validate(m Manifest, l Limits) error {
	if o.Sequence == 0 || o.At < 0 || !text(o.Window, 256, true) || !text(o.Target, 256, true) || !text(o.Key, 128, true) || !text(o.Action, 256, true) || !text(o.Label, 256, true) || o.Modifiers > 15 || math.IsNaN(o.X) || math.IsNaN(o.Y) || math.IsInf(o.X, 0) || math.IsInf(o.Y, 0) || o.Delta < 0 {
		return ErrInvalid
	}
	if (o.Window == "") != (o.WindowIncarnation == 0) || (o.Target == "") != (o.TargetIncarnation == 0) || !o.CheckRevision && o.ExpectedRevision != 0 {
		return ErrInvalid
	}
	if o.Window != "" {
		found := false
		for _, w := range m.Windows {
			if w.ID == o.Window && w.Incarnation == o.WindowIncarnation {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%w: window lifetime", ErrInvalid)
		}
	}
	if len(o.Arguments) > 32 {
		return ErrInvalid
	}
	payload := len(o.Key) + len(o.Action) + len(o.Label)
	keys := map[string]bool{}
	for _, a := range o.Arguments {
		if !text(a.Key, 128, false) || !text(a.Value, 1<<20, true) || keys[a.Key] {
			return ErrInvalid
		}
		keys[a.Key] = true
		payload += len(a.Key) + len(a.Value)
	}
	if payload > l.MaxPayloadBytes {
		return ErrCapacity
	}
	// Reconstruct only the allowed fields for this variant. Unknown/irrelevant
	// fields refuse instead of silently changing interpretation across backends.
	allowed := Operation{Sequence: o.Sequence, At: o.At, Kind: o.Kind}
	window := func() { allowed.Window = o.Window; allowed.WindowIncarnation = o.WindowIncarnation }
	target := func() {
		allowed.Target = o.Target
		allowed.TargetIncarnation = o.TargetIncarnation
		allowed.CheckRevision = o.CheckRevision
		allowed.ExpectedRevision = o.ExpectedRevision
	}
	switch o.Kind {
	case PointerDown, PointerUp, PointerMove:
		window()
		target()
		allowed.X = o.X
		allowed.Y = o.Y
		allowed.Modifiers = o.Modifiers
		if o.Window == "" || o.X < 0 || o.Y < 0 {
			return ErrInvalid
		}
		for _, w := range m.Windows {
			if w.ID == o.Window && (o.X >= float64(w.Width) || o.Y >= float64(w.Height)) {
				return ErrInvalid
			}
		}
		if o.Kind != PointerMove {
			if o.Button < 1 || o.Button > 3 {
				return ErrInvalid
			}
			allowed.Button = o.Button
		}
	case KeyDown, KeyUp, KeyChord:
		window()
		target()
		allowed.Key = o.Key
		allowed.Modifiers = o.Modifiers
		if o.Window == "" || o.Key == "" {
			return ErrInvalid
		}
	case Command, Semantic:
		window()
		target()
		allowed.Action = o.Action
		allowed.Arguments = o.Arguments
		if o.Window == "" || o.Action == "" || o.Kind == Semantic && o.Target == "" {
			return ErrInvalid
		}
	case Advance:
		window()
		allowed.Delta = o.Delta
		if o.Window == "" {
			return ErrInvalid
		}
	case Worker:
		allowed.WorkerID = o.WorkerID
		if o.WorkerID == 0 {
			return ErrInvalid
		}
	case Frame:
		window()
		if o.Window == "" {
			return ErrInvalid
		}
	case Capture:
		window()
		allowed.Label = o.Label
		if o.Label == "" {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	if !reflect.DeepEqual(allowed, o) {
		return fmt.Errorf("%w: variant fields", ErrInvalid)
	}
	return nil
}

// Validate checks all portable bounds before a backend sees any operation.
// Live model/command/target checks are additional backend responsibilities.
func (r Record) Validate(l Limits) error {
	l, e := l.normalized()
	if e != nil {
		return e
	}
	if e = r.Manifest.validate(); e != nil {
		return e
	}
	if len(r.Operations) > l.MaxOperations {
		return ErrCapacity
	}
	// Count canonical bytes incrementally. A caller-built oversized Record must
	// refuse before marshaling its entire aggregate into an unbounded allocation.
	base, e := json.Marshal(Record{Manifest: r.Manifest, Operations: []Operation{}})
	if e != nil {
		return fmt.Errorf("%w: encoding", ErrInvalid)
	}
	size := len(base)
	if size > l.MaxBytes {
		return ErrCapacity
	}
	var at time.Duration
	for i, o := range r.Operations {
		if o.Sequence != uint64(i)+1 || o.At < at {
			return fmt.Errorf("%w: sequence/time", ErrInvalid)
		}
		if e = o.validate(r.Manifest, l); e != nil {
			return fmt.Errorf("operation %d: %w", i+1, e)
		}
		at = o.At
		encoded, e := json.Marshal(o)
		if e != nil {
			return fmt.Errorf("%w: encoding", ErrInvalid)
		}
		extra := len(encoded)
		if i > 0 {
			extra++
		}
		if extra > l.MaxBytes-size {
			return ErrCapacity
		}
		size += extra
	}
	return nil
}
func (r Record) Encode(l Limits) ([]byte, error) {
	if e := r.Validate(l); e != nil {
		return nil, e
	}
	if r.Operations == nil {
		r.Operations = []Operation{}
	}
	return json.Marshal(r)
}
func (r Record) Digest(l Limits) (string, error) {
	b, e := r.Encode(l)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// Recorder is safe for concurrent producers. Append assigns sequence only for
// accepted operations (a supplied nonzero sequence must match). Failed admission
// preserves the record exactly; Snapshot returns detached slices.
// Do not copy Recorder after construction.
type Recorder struct {
	mu     sync.Mutex
	limits Limits
	record Record
	bytes  int
	self   *Recorder
}

func NewRecorder(m Manifest, l Limits) (*Recorder, error) {
	l, e := l.normalized()
	if e != nil {
		return nil, e
	}
	if e = m.validate(); e != nil {
		return nil, e
	}
	m = cloneManifest(m)
	r := &Recorder{limits: l, record: Record{Manifest: m, Operations: []Operation{}}}
	r.self = r
	b, e := json.Marshal(r.record)
	if e != nil {
		return nil, e
	}
	r.bytes = len(b)
	if r.bytes > l.MaxBytes {
		return nil, ErrCapacity
	}
	return r, nil
}
func (r *Recorder) Append(o Operation) error {
	if r == nil || r.self != r {
		return ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.record.Operations) >= r.limits.MaxOperations {
		return ErrCapacity
	}
	next := uint64(len(r.record.Operations)) + 1
	if o.Sequence != 0 && o.Sequence != next {
		return fmt.Errorf("%w: sequence", ErrInvalid)
	}
	o.Sequence = next
	if len(r.record.Operations) > 0 && o.At < r.record.Operations[len(r.record.Operations)-1].At {
		return fmt.Errorf("%w: time reversal", ErrInvalid)
	}
	if e := o.validate(r.record.Manifest, r.limits); e != nil {
		return e
	}
	o = cloneOperation(o)
	b, e := json.Marshal(o)
	if e != nil {
		return e
	}
	extra := len(b)
	if len(r.record.Operations) > 0 {
		extra++
	}
	if extra > r.limits.MaxBytes-r.bytes {
		return ErrCapacity
	}
	r.bytes += extra
	r.record.Operations = append(r.record.Operations, o)
	return nil
}
func (r *Recorder) Snapshot() (Record, error) {
	if r == nil || r.self != r {
		return Record{}, ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v := Record{Manifest: cloneManifest(r.record.Manifest), Operations: make([]Operation, len(r.record.Operations))}
	for i, o := range r.record.Operations {
		v.Operations[i] = cloneOperation(o)
	}
	return v, nil
}
