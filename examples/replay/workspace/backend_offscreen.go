//go:build offscreen

package workspace

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/replay"
	"github.com/dshills/vellum/ui"
)

// Backend is an example adapter for this declared fixture. Its physical
// operations enter the actual managed driver; commands remain a separate route.
// The caller owns it and explicitly closes its driver, owner, clock and executor.
type Backend struct {
	Fixture  *Fixture
	Driver   *cogent.Manual
	Clock    *execution.ManualClock
	Executor *execution.Manual
	manifest replay.Manifest
	bound    bool
}

func NewBackend(ctx context.Context, rows, ownerBatch int) (_ *Backend, err error) {
	if ctx == nil || ownerBatch < 1 || ownerBatch > 64 {
		return nil, replay.ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if err != nil {
		return nil, err
	}
	ex, err := execution.NewManual(8)
	if err != nil {
		clock.Close()
		return nil, err
	}
	a, err := app.New(app.Config{QueueCapacity: 8, CompletionCapacity: 4})
	if err != nil {
		ex.Close()
		clock.Close()
		return nil, err
	}
	var driver *cogent.Manual
	var f *Fixture
	constructed := false
	defer func() {
		if !constructed {
			if f != nil {
				f.Palette.Close()
			}
			if driver != nil {
				_ = driver.Close(context.Background())
			}
			a.Close()
			ex.Close()
			clock.Close()
		}
	}()
	f, err = New(a, ex, rows)
	if err != nil {
		return nil, err
	}
	theme := ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}
	config := cogent.ManualConfig{Clock: clock, Executor: ex, Commands: f.Commands, Theme: theme, OwnerBatch: ownerBatch, MaxWindows: 4, EventCapacity: 128, LogicalDPI: 160}
	var identities []replay.Font
	err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		data, err := fs.ReadFile(fonts.Default, path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(data)
		digest := hex.EncodeToString(h[:])
		config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, SHA256: digest, Data: data})
		identities = append(identities, replay.Font{Family: path, SHA256: digest})
		return nil
	})
	if err != nil {
		return nil, err
	}
	driver, err = cogent.NewManual(ctx, a, config)
	if err != nil {
		return nil, err
	}
	s, err := driver.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	themeBytes, err := json.Marshal(theme)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(themeBytes)
	m := replay.Manifest{Version: replay.Version, FixtureSHA256: FixtureSHA256(rows), ThemeSHA256: hex.EncodeToString(h[:]), LayoutVersion: "vellum-managed-layout-1", RendererVersion: "cogent-manual-software-1", Start: clock.Now(), Seed: 1, Fonts: identities, Windows: []replay.Window{}}
	for _, w := range s.Windows {
		m.Windows = append(m.Windows, replay.Window{ID: w.ID, Incarnation: w.Incarnation, Width: windowWidth(w.ID), Height: windowHeight(w.ID), DPI: 160})
	}
	constructed = true
	return &Backend{Fixture: f, Driver: driver, Clock: clock, Executor: ex, manifest: m}, nil
}

func (b *Backend) Manifest() replay.Manifest {
	m := b.manifest
	m.Windows = append([]replay.Window(nil), m.Windows...)
	m.Fonts = append([]replay.Font(nil), m.Fonts...)
	return m
}

// Close owns this fixture's executor: revoke application lifetimes first, then
// explicitly finish its at-most-eight queued canceled workers. Manual.Close
// itself still never chooses work from its borrowed executor.
func (b *Backend) Close(ctx context.Context) error {
	if ctx == nil {
		return replay.ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	b.Fixture.Owner.Close()
	ids := b.Executor.PendingIDs()
	var cleanupErr error
	for _, id := range ids {
		cleanupErr = errors.Join(cleanupErr, b.Executor.Step(ctx, id))
	}
	if b.Fixture.task != nil {
		select {
		case <-b.Fixture.task.Done():
		case <-ctx.Done():
			return errors.Join(cleanupErr, ctx.Err())
		}
	}
	if !b.Executor.Snapshot().Idle() {
		return errors.Join(cleanupErr, fmt.Errorf("workspace: accepted workers remain during closure"))
	}
	if err := b.Fixture.WaitWorker(ctx); err != nil {
		return errors.Join(cleanupErr, err)
	}
	if err := b.Driver.Close(ctx); err != nil {
		return errors.Join(cleanupErr, err)
	}
	b.Fixture.Palette.Close()
	b.Executor.Close()
	b.Clock.Close()
	return cleanupErr
}
func (b *Backend) CheckManifest(ctx context.Context, m replay.Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !reflect.DeepEqual(m, b.manifest) {
		return fmt.Errorf("%w: fixture/theme/font/viewport identity", replay.ErrBackend)
	}
	if b.Fixture.Owner.Context().Err() != nil {
		return app.ErrClosed
	}
	// Later segments bind the same immutable identity, while their live-window
	// set may legitimately reflect recorded closure. Each input checks liveness.
	if b.bound {
		return nil
	}
	native, err := b.Driver.Snapshot(ctx)
	if err != nil {
		return err
	}
	if len(native.Windows) != len(m.Windows) {
		return replay.ErrBackend
	}
	for i, w := range native.Windows {
		if w.ID != m.Windows[i].ID || w.Incarnation != m.Windows[i].Incarnation || w.Closed {
			return cogent.ErrManualStale
		}
	}
	live := b.Fixture.Owner.Windows()
	if len(live) != len(m.Windows) {
		return replay.ErrBackend
	}
	for i, w := range live {
		if w.ID != m.Windows[i].ID || w.Width != float32(m.Windows[i].Width) || w.Height != float32(m.Windows[i].Height) {
			return replay.ErrBackend
		}
		lifetime, err := b.Fixture.Owner.WindowContext(w.ID)
		if err != nil {
			return err
		}
		if lifetime.Err() != nil {
			return app.ErrWindowClosed
		}
	}
	b.bound = true
	return nil
}
func (b *Backend) Model() (s Snapshot, err error) {
	err = b.Fixture.Owner.Do(func() { s = b.Fixture.Snapshot() })
	return
}

// Readiness includes origin FIFO, ready paint, owner delivery and admitted worker
// tails. Unrendered layout requires an explicit Frame operation, hence Blocked.
func (b *Backend) Readiness(ctx context.Context) (replay.Readiness, error) {
	return b.Driver.ReplayReadiness(ctx)
}
func (b *Backend) Step(ctx context.Context, n int) (replay.Progress, error) {
	r, err := b.Driver.Step(ctx, n)
	return replay.Progress{Units: r.Events + r.OwnerCallbacks, Events: r.Events, Owners: r.OwnerCallbacks, Frames: r.Frames}, err
}

func (b *Backend) checkLive(ctx context.Context, o replay.Operation) (cogent.ManualWindowSnapshot, error) {
	if !b.bound {
		return cogent.ManualWindowSnapshot{}, replay.ErrBackend
	}
	if err := ctx.Err(); err != nil {
		return cogent.ManualWindowSnapshot{}, err
	}
	if b.Fixture.Owner.Context().Err() != nil {
		return cogent.ManualWindowSnapshot{}, app.ErrClosed
	}
	if o.CheckRevision {
		s, err := b.Model()
		if err != nil {
			return cogent.ManualWindowSnapshot{}, err
		}
		if s.HistoryRevision != o.ExpectedRevision {
			return cogent.ManualWindowSnapshot{}, replay.ErrBackend
		}
	}
	if o.Window == "" {
		return cogent.ManualWindowSnapshot{}, nil
	}
	lifetime, err := b.Fixture.Owner.WindowContext(o.Window)
	if err != nil {
		return cogent.ManualWindowSnapshot{}, err
	}
	if lifetime.Err() != nil {
		return cogent.ManualWindowSnapshot{}, app.ErrWindowClosed
	}
	s, err := b.Driver.Snapshot(ctx)
	if err != nil {
		return cogent.ManualWindowSnapshot{}, err
	}
	for _, w := range s.Windows {
		if w.ID != o.Window {
			continue
		}
		if w.Incarnation != o.WindowIncarnation {
			return w, cogent.ErrManualStale
		}
		if o.Target != "" {
			if w.LayoutPending {
				return w, cogent.ErrManualLayout
			}
			for _, e := range w.Elements {
				if e.ID != o.Target {
					continue
				}
				if e.Incarnation != o.TargetIncarnation {
					return w, cogent.ErrManualStale
				}
				if !e.Enabled || !e.Visible {
					return w, cogent.ErrManualDisabled
				}
				if (o.Kind == replay.PointerDown || o.Kind == replay.PointerUp || o.Kind == replay.Scroll) && !image.Pt(int(o.X), int(o.Y)).In(e.Bounds) {
					return w, replay.ErrInvalid
				}
				if (o.Kind == replay.KeyDown || o.Kind == replay.KeyUp || o.Kind == replay.KeyChord) && !e.Focused {
					return w, cogent.ErrManualStale
				}
				return w, nil
			}
			return w, cogent.ErrManualStale
		}
		return w, nil
	}
	return cogent.ManualWindowSnapshot{}, cogent.ErrManualClosed
}

// Key names are this fixture's declared transport: one Unicode scalar for text,
// or one of the explicit special codes. Modifier bits are Shift/Control/Alt/Meta.
func transportKey(value string) (rune, key.Codes, error) {
	codes := map[string]key.Codes{"Enter": key.CodeReturnEnter, "Escape": key.CodeEscape, "Tab": key.CodeTab, "Backspace": key.CodeBackspace, "Delete": key.CodeDelete, "Left": key.CodeLeftArrow, "Right": key.CodeRightArrow, "Up": key.CodeUpArrow, "Down": key.CodeDownArrow, "Home": key.CodeHome, "End": key.CodeEnd, "F2": key.CodeF2}
	if code, ok := codes[value]; ok {
		return 0, code, nil
	}
	if utf8.RuneCountInString(value) != 1 {
		return 0, 0, replay.ErrInvalid
	}
	r, _ := utf8.DecodeRuneInString(value)
	code := key.CodeUnknown
	if r == 'a' || r == 'A' {
		code = key.CodeA
	}
	if r == 'z' || r == 'Z' {
		code = key.CodeZ
	}
	return r, code, nil
}
func transportModifiers(bits uint32) key.Modifiers {
	var m key.Modifiers
	for i, flag := range []key.Modifiers{key.Shift, key.Control, key.Alt, key.Meta} {
		if bits&(1<<i) != 0 {
			m.SetFlag(true, flag)
		}
	}
	return m
}
func (b *Backend) Apply(ctx context.Context, o replay.Operation) error {
	if _, err := b.checkLive(ctx, o); err != nil {
		return err
	}
	switch o.Kind {
	case replay.Frame:
		return b.Driver.RequestFrame(ctx, o.Window)
	case replay.Advance:
		_, err := b.Clock.AdvanceContext(ctx, o.Delta)
		return err
	case replay.Worker:
		return b.Executor.Step(ctx, execution.ID(o.WorkerID))
	case replay.Command:
		value := ""
		for _, a := range o.Arguments {
			if a.Key != "value" {
				return ErrUnavailable
			}
			value = a.Value
		}
		var callErr error
		err := b.Fixture.Owner.Do(func() {
			callErr = b.Fixture.Invoke(ctx, o.Action, value)
		})
		return errors.Join(err, callErr)
	case replay.Semantic:
		return cogent.ErrManualUnsupported
	case replay.PointerDown, replay.PointerUp, replay.PointerMove, replay.Scroll, replay.KeyDown, replay.KeyUp, replay.KeyChord:
		input := cogent.ManualInput{Window: o.Window, WindowIncarnation: o.WindowIncarnation, Position: image.Pt(int(o.X), int(o.Y)), Modifiers: transportModifiers(o.Modifiers)}
		switch o.Kind {
		case replay.PointerDown:
			input.Type = events.MouseDown
		case replay.PointerUp:
			input.Type = events.MouseUp
		case replay.PointerMove:
			input.Type = events.MouseMove
		case replay.Scroll:
			input.Type = events.Scroll
			input.Delta = math32.Vec2(float32(o.DeltaX), float32(o.DeltaY))
		case replay.KeyDown:
			input.Type = events.KeyDown
		case replay.KeyUp:
			input.Type = events.KeyUp
		case replay.KeyChord:
			input.Type = events.KeyChord
		}
		if input.Type.IsKey() {
			var err error
			input.Rune, input.Code, err = transportKey(o.Key)
			if err != nil {
				return err
			}
		} else if o.Kind != replay.PointerMove && o.Kind != replay.Scroll {
			input.Button = events.Buttons(o.Button)
		}
		return b.Driver.Input(ctx, input)
	default:
		return replay.ErrInvalid
	}
}

//go:embed records/*.json
var recordFiles embed.FS

func FixtureBytes(rows int) []byte {
	return []byte(fmt.Sprintf("{\"schema\":1,\"rows\":%d,\"columns\":20,\"tree_nodes\":100000,\"values\":\"(r*97+c*13+7)%%10007\",\"parent\":\"(node-1)/100\",\"seed\":1}", rows))
}
func FixtureSHA256(rows int) string {
	h := sha256.Sum256(FixtureBytes(rows))
	return hex.EncodeToString(h[:])
}
func windowWidth(id string) int {
	if id == "inspector" {
		return 500
	}
	return Width
}
func windowHeight(id string) int {
	if id == "inspector" {
		return 320
	}
	return Height
}
func Record(rows int) (replay.Record, error) {
	var record replay.Record
	if rows != 10000 && rows != MaxRows {
		return record, replay.ErrInvalid
	}
	data, err := recordFiles.ReadFile(fmt.Sprintf("records/%d.json", rows))
	if err != nil {
		return record, err
	}
	if len(data) > 8<<20 {
		return record, replay.ErrCapacity
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return record, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return record, replay.ErrInvalid
	}
	return record, record.Validate(replay.Limits{})
}
func (b *Backend) Capture(ctx context.Context, o replay.Operation) (replay.Observation, error) {
	if _, err := b.checkLive(ctx, o); err != nil {
		return replay.Observation{}, err
	}
	native, err := b.Driver.Snapshot(ctx)
	if err != nil {
		return replay.Observation{}, err
	}
	model, err := b.Model()
	if err != nil {
		return replay.Observation{}, err
	}
	if model.Failure != "" {
		return replay.Observation{}, fmt.Errorf("workspace: %s", model.Failure)
	}
	capture := Capture{Model: model}
	type imageIdentity struct {
		Window string
		Bounds image.Rectangle
		Stride int
		SHA256 string
	}
	var images []imageIdentity
	for _, w := range native.Windows {
		pixels, err := b.Driver.Capture(ctx, w.ID)
		if err != nil {
			return replay.Observation{}, err
		}
		digest := sha256.Sum256(pixels.Pix)
		images = append(images, imageIdentity{w.ID, pixels.Bounds(), pixels.Stride, hex.EncodeToString(digest[:])})
		// Input capability tokens and transport frame counters are checked at input
		// admission, but are not semantic equality across independent instances.
		for i := range w.Elements {
			w.Elements[i].Incarnation = 0
		}
		for i := range w.Grids {
			g := &w.Grids[i]
			g.Incarnation = 0
			g.Generation = 0
			g.Frames = 0
			g.CellIncarnation = 0
		}
		for i := range w.Trees {
			t := &w.Trees[i]
			t.Incarnation = 0
			t.Generation = 0
			t.Frames = 0
			t.RowIncarnation = 0
		}
		capture.Windows = append(capture.Windows, WindowCapture{ID: w.ID, FocusedID: w.FocusedID, Elements: w.Elements, Grids: w.Grids, Trees: w.Trees, Bounds: pixels.Bounds()})
	}
	data, err := json.Marshal(capture)
	if err != nil {
		return replay.Observation{}, err
	}
	identity, err := json.Marshal(images)
	if err != nil {
		return replay.Observation{}, err
	}
	digest := sha256.Sum256(identity)
	return replay.Observation{Data: data, ImageSHA256: hex.EncodeToString(digest[:])}, nil
}
