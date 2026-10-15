// Package placement provides typed, owner-held palette placement gestures.
// Native drivers supply capture and measured geometry; applications supply one
// command callback, normally invoking their command registry and undo journal.
package placement

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/canvas"
)

var (
	ErrInvalid  = errors.New("placement: invalid configuration or geometry")
	ErrClosed   = errors.New("placement: controller or gesture closed")
	ErrBusy     = errors.New("placement: gesture or command already active")
	ErrRejected = errors.New("placement: target rejected placement")
)

const MaxItems = 256
const MaxTargets = 16

// Item is immutable for the controller lifetime. Payload is strongly typed and
// caller-owned; reference-bearing payloads must themselves remain immutable.
// Preview's width, height and color describe a proposed world-space footprint.
// Its ID and X/Y are ignored. The catalog is bounded by count, not payload bytes.
type Item[T any] struct {
	ID, Label string
	Payload   T
	Preview   canvas.Rect
}
type Request[T any] struct {
	ItemID, TargetID string
	Payload          T
	Position         canvas.Point
}
type Config[T any] struct {
	Items []Item[T]
	// Command runs synchronously on the UI owner, after final validation and
	// gesture retirement. It must either apply one complete undoable command or
	// return an error without changing the model. A callback panic propagates.
	Command func(context.Context, Request[T]) error
}
type TargetConfig[T any] struct {
	ID string
	// Snap precedes Accept for previews, pointer commit and keyboard placement.
	Snap   func(Request[T]) (canvas.Point, error)
	Accept func(Request[T]) error
}

// Geometry separates the unclipped content origin from the visible clip, all
// in window pixels. View is in logical canvas units; Scale is pixels per unit.
// A scrolled child can have an origin outside Clip. Apply display scale once.
type Geometry struct {
	Window string
	Origin canvas.Point
	Clip   canvas.Bounds
	Scale  float64
	View   canvas.Viewport
}

func (g Geometry) World(p canvas.Point) (canvas.Point, error) {
	if !identity(g.Window) || !point(p) || !point(g.Origin) || !point(g.Clip.Min) || !point(g.Clip.Max) || !finite(g.Scale) || g.Scale <= 0 || g.Scale > 64 || g.Clip.Max.X <= g.Clip.Min.X || g.Clip.Max.Y <= g.Clip.Min.Y {
		return canvas.Point{}, ErrInvalid
	}
	if _, err := g.View.Bounds(); err != nil {
		return canvas.Point{}, ErrInvalid
	}
	if p.X < g.Clip.Min.X || p.Y < g.Clip.Min.Y || p.X >= g.Clip.Max.X || p.Y >= g.Clip.Max.Y {
		return canvas.Point{}, ErrRejected
	}
	local := canvas.Point{X: (p.X - g.Origin.X) / g.Scale, Y: (p.Y - g.Origin.Y) / g.Scale}
	if local.X < 0 || local.Y < 0 || local.X >= g.View.Width || local.Y >= g.View.Height {
		return canvas.Point{}, ErrRejected
	}
	result := g.View.World(local)
	if !point(result) {
		return canvas.Point{}, ErrInvalid
	}
	return result, nil
}

// Source and Target are opaque transport endpoints for managed UI records.
// Obtain them from a Controller; zero values refuse. All operations share the
// controller's single UI owner and are neither concurrent nor reentrant.
type Source struct {
	self      *Source
	id, label string
	owner     any
	preview   canvas.Rect
	begin     func(string, canvas.Point, float64) (*Gesture, error)
	keyboard  func(context.Context, *Target, canvas.Point) error
	cancel    func()
}

func (s *Source) ID() string {
	if s == nil {
		return ""
	}
	return s.id
}
func (s *Source) Label() string {
	if s == nil {
		return ""
	}
	return s.label
}
func (s *Source) Valid() bool { return s != nil && s.self == s && s.begin != nil && s.owner != nil }
func (s *Source) Begin(window string, p canvas.Point, thresholdPixels float64) (*Gesture, error) {
	if !s.Valid() {
		return nil, ErrClosed
	}
	return s.begin(window, p, thresholdPixels)
}

// PlaceKeyboard uses exactly the pointer drop's snap, validation and command.
// The application chooses an explicit target and world position (e.g. center).
func (s *Source) PlaceKeyboard(ctx context.Context, target *Target, p canvas.Point) error {
	if !s.Valid() {
		return ErrClosed
	}
	return s.keyboard(ctx, target, p)
}

// Cancel retires a gesture originating from this source, e.g. on record removal.
func (s *Source) Cancel() {
	if s.Valid() && s.cancel != nil {
		s.cancel()
	}
}

type proposal struct {
	position canvas.Point
	commit   func(context.Context) error
}
type Target struct {
	self    *Target
	id      string
	owner   any
	propose func(*Source, canvas.Point) (proposal, error)
	cancel  func()
	state   func() State
}

func (t *Target) ID() string {
	if t == nil {
		return ""
	}
	return t.id
}
func (t *Target) Valid() bool { return t != nil && t.self == t && t.owner != nil && t.propose != nil }
func (t *Target) Cancel() {
	if t.Valid() && t.cancel != nil {
		t.cancel()
	}
}

// State is detached value data; it retains no payload, command or widget.
func (t *Target) State() State {
	if !t.Valid() || t.state == nil {
		return State{}
	}
	return t.state()
}

type State struct {
	Active, Dragging, Accepted bool
	ItemID, TargetID           string
	Pointer, Position          canvas.Point
	Preview                    canvas.Rect
	Reason                     string
}
type Gesture struct {
	self      *Gesture
	source    *Source
	window    string
	start     canvas.Point
	threshold float64
	target    *Target
	candidate *proposal
	state     State
	finish    func()
	closed    bool
	mutating  bool
}

func (g *Gesture) State() State {
	if g == nil || g.self != g {
		return State{}
	}
	return g.state
}

// Move evaluates the current target against fresh measured geometry. Moving
// below threshold never produces acceptance or a preview. Invalid/outside
// positions clear the previous candidate; they cannot commit a stale preview.
func (g *Gesture) Move(target *Target, geometry Geometry, p canvas.Point) error {
	if g == nil || g.self != g || g.closed {
		return ErrClosed
	}
	if g.mutating {
		return ErrBusy
	}
	g.mutating = true
	defer func() { g.mutating = false }()
	return g.move(target, geometry, p)
}
func (g *Gesture) move(target *Target, geometry Geometry, p canvas.Point) error {
	if !point(p) {
		g.Cancel()
		return ErrInvalid
	}
	g.state.Pointer = p
	g.state.Dragging = g.state.Dragging || math.Hypot(p.X-g.start.X, p.Y-g.start.Y) >= g.threshold
	g.target = nil
	g.candidate = nil
	g.state.Accepted = false
	g.state.TargetID = ""
	g.state.Reason = ""
	g.state.Preview = g.source.preview
	if !g.state.Dragging {
		return nil
	}
	if !target.Valid() || target.owner != g.source.owner {
		g.state.Reason = "No accepting canvas"
		return nil
	}
	g.target = target
	g.state.TargetID = target.id
	if geometry.Window != g.window {
		g.state.Reason = "Cross-window placement is unsupported"
		return nil
	}
	world, err := geometry.World(p)
	if err != nil {
		g.state.Reason = reason(err)
		return nil
	}
	g.state.Position = world
	g.state.Preview.X = world.X
	g.state.Preview.Y = world.Y
	candidate, err := target.propose(g.source, world)
	if g.closed {
		return ErrClosed
	}
	if point(candidate.position) {
		g.state.Position = candidate.position
		g.state.Preview.X = candidate.position.X
		g.state.Preview.Y = candidate.position.Y
	}
	if err != nil {
		g.state.Reason = reason(err)
		return nil
	}
	g.candidate = &candidate
	g.state.Accepted = true
	return nil
}

// Release revalidates the final pointer, consumes the gesture before invoking
// the command, and invokes it at most once. Rejected or canceled releases are
// normal non-committing outcomes; command/context errors are returned.
func (g *Gesture) Release(ctx context.Context, target *Target, geometry Geometry, p canvas.Point) (bool, error) {
	if g == nil || g.self != g || g.closed {
		return false, ErrClosed
	}
	if g.mutating {
		return false, ErrBusy
	}
	g.mutating = true
	defer func() { g.mutating = false }()
	if ctx == nil {
		g.Cancel()
		return false, ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		g.Cancel()
		return false, err
	}
	if err := g.move(target, geometry, p); err != nil {
		return false, err
	}
	candidate := g.candidate
	g.retire()
	if candidate == nil {
		return false, nil
	}
	if err := candidate.commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
func (g *Gesture) Cancel() {
	if g != nil && g.self == g && !g.closed {
		g.retire()
	}
}
func (g *Gesture) retire() {
	finish := g.finish
	g.finish = nil
	g.source = nil
	g.target = nil
	g.candidate = nil
	g.window = ""
	g.state = State{}
	g.closed = true
	if finish != nil {
		finish()
	}
}

// Controller owns at most one gesture, 256 immutable items and sixteen targets.
// Keep it on the application's UI owner; close it when the caller retires the
// palette. Drivers borrow it and cancel gestures when their bindings disappear.
type Controller[T any] struct {
	self         *Controller[T]
	items        map[string]Item[T]
	sources      map[string]*Source
	targets      []*Target
	command      func(context.Context, Request[T]) error
	active       *Gesture
	busy, closed bool
}

func New[T any](config Config[T]) (*Controller[T], error) {
	if len(config.Items) < 1 || len(config.Items) > MaxItems || config.Command == nil {
		return nil, ErrInvalid
	}
	c := &Controller[T]{items: map[string]Item[T]{}, sources: map[string]*Source{}, command: config.Command}
	c.self = c
	for _, item := range config.Items {
		if !identity(item.ID) || !identity(item.Label) || c.sources[item.ID] != nil || !finite(item.Preview.Width) || !finite(item.Preview.Height) || item.Preview.Width <= 0 || item.Preview.Height <= 0 || item.Preview.Width > 1e6 || item.Preview.Height > 1e6 {
			return nil, ErrInvalid
		}
		item.Preview.ID = ""
		item.Preview.X = 0
		item.Preview.Y = 0
		c.items[item.ID] = item
		source := &Source{id: item.ID, label: item.Label, owner: c, preview: item.Preview}
		source.self = source
		source.begin = func(window string, p canvas.Point, threshold float64) (*Gesture, error) {
			if c.closed {
				return nil, ErrClosed
			}
			if c.busy || c.active != nil {
				return nil, ErrBusy
			}
			if !identity(window) || !point(p) || !finite(threshold) || threshold <= 0 || threshold > 4096 {
				return nil, ErrInvalid
			}
			g := &Gesture{source: source, window: window, start: p, threshold: threshold, state: State{Active: true, ItemID: source.id, Pointer: p}}
			g.self = g
			g.finish = func() {
				if c.active == g {
					c.active = nil
				}
			}
			c.active = g
			return g, nil
		}
		source.keyboard = func(ctx context.Context, target *Target, p canvas.Point) error {
			if c.closed {
				return ErrClosed
			}
			if c.busy || c.active != nil {
				return ErrBusy
			}
			if ctx == nil || !point(p) {
				return ErrInvalid
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if !target.Valid() || target.owner != c {
				return ErrRejected
			}
			proposal, err := target.propose(source, p)
			if err != nil {
				return err
			}
			return proposal.commit(ctx)
		}
		source.cancel = func() {
			if c.active != nil && c.active.source == source {
				c.active.Cancel()
			}
		}
		c.sources[item.ID] = source
	}
	return c, nil
}
func (c *Controller[T]) Source(id string) *Source {
	if c == nil || c.self != c || c.closed {
		return nil
	}
	return c.sources[id]
}
func (c *Controller[T]) State() State {
	if c == nil || c.self != c || c.active == nil {
		return State{}
	}
	return c.active.State()
}
func (c *Controller[T]) NewTarget(config TargetConfig[T]) (*Target, error) {
	if c == nil || c.self != c || c.closed || c.command == nil {
		return nil, ErrClosed
	}
	if c.busy {
		return nil, ErrBusy
	}
	if !identity(config.ID) || len(c.targets) >= MaxTargets {
		return nil, ErrInvalid
	}
	for _, t := range c.targets {
		if t.id == config.ID {
			return nil, ErrInvalid
		}
	}
	target := &Target{id: config.ID, owner: c}
	target.self = target
	target.propose = func(source *Source, p canvas.Point) (proposal, error) {
		if c.closed {
			return proposal{}, ErrClosed
		}
		if c.busy {
			return proposal{}, ErrBusy
		}
		if !source.Valid() || source.owner != c || !point(p) {
			return proposal{}, ErrInvalid
		}
		c.busy = true
		defer func() { c.busy = false }()
		origin := c.active
		item := c.items[source.id]
		request := Request[T]{ItemID: item.ID, TargetID: target.id, Payload: item.Payload, Position: p}
		if config.Snap != nil {
			snapped, err := config.Snap(request)
			if err != nil {
				return proposal{position: p}, err
			}
			if !point(snapped) {
				return proposal{position: p}, ErrInvalid
			}
			request.Position = snapped
		}
		if c.closed || origin != nil && (origin.closed || c.active != origin) {
			return proposal{}, ErrClosed
		}
		proposed := proposal{position: request.Position}
		if config.Accept != nil {
			if err := config.Accept(request); err != nil {
				return proposed, err
			}
		}
		if c.closed {
			return proposal{}, ErrClosed
		}
		proposed.commit = func(ctx context.Context) error {
			if ctx == nil {
				return ErrInvalid
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if c.closed {
				return ErrClosed
			}
			if c.busy {
				return ErrBusy
			}
			c.busy = true
			defer func() { c.busy = false }()
			return c.command(ctx, request)
		}
		return proposed, nil
	}
	target.cancel = func() {
		if c.active != nil && c.active.target == target {
			c.active.Cancel()
		}
	}
	target.state = func() State {
		if c.active != nil && c.active.target == target {
			return c.active.State()
		}
		return State{}
	}
	c.targets = append(c.targets, target)
	return target, nil
}
func (c *Controller[T]) Close() {
	if c == nil || c.self != c || c.closed {
		return
	}
	c.closed = true
	if c.active != nil {
		c.active.Cancel()
	}
	for _, s := range c.sources {
		s.owner = nil
		s.begin = nil
		s.keyboard = nil
		s.cancel = nil
	}
	for _, t := range c.targets {
		t.owner = nil
		t.propose = nil
		t.cancel = nil
		t.state = nil
	}
	clear(c.items)
	clear(c.sources)
	clear(c.targets)
	c.items = nil
	c.sources = nil
	c.targets = nil
	c.command = nil
}
func identity(s string) bool {
	return len(s) > 0 && len(s) <= 256 && utf8.ValidString(s) && !strings.ContainsRune(s, 0) && strings.TrimSpace(s) != ""
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func point(p canvas.Point) bool {
	return finite(p.X) && finite(p.Y) && math.Abs(p.X) <= 1e12 && math.Abs(p.Y) <= 1e12
}
func reason(err error) string {
	s := fmt.Sprint(err)
	if len(s) > 512 {
		s = s[:512]
	}
	return strings.ToValidUTF8(s, "?")
}
