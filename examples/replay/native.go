//go:build offscreen

package replayconsumer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
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

// NativeBackend is an example adapter for this declared fixture. Its physical
// operations enter the actual managed driver; commands remain a separate route.
// The caller owns it and explicitly closes its driver, owner, clock and executor.
type NativeBackend struct {
	Fixture  *Fixture
	Driver   *cogent.Manual
	Clock    *execution.ManualClock
	Executor *execution.Manual
	manifest replay.Manifest
	bound    bool
}

func NewNative(ctx context.Context, ownerBatch int) (_ *NativeBackend, err error) {
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
	constructed := false
	defer func() {
		if !constructed {
			if driver != nil {
				_ = driver.Close(context.Background())
			}
			a.Close()
			ex.Close()
			clock.Close()
		}
	}()
	f, err := New(Config{Owner: a, Executor: ex})
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
	m := replay.Manifest{Version: replay.LegacyVersion, FixtureSHA256: FixtureSHA256(), ThemeSHA256: hex.EncodeToString(h[:]), LayoutVersion: "vellum-managed-layout-1", RendererVersion: "cogent-manual-software-1", Start: clock.Now(), Seed: 1, Fonts: identities, Windows: []replay.Window{}}
	for _, w := range s.Windows {
		m.Windows = append(m.Windows, replay.Window{ID: w.ID, Incarnation: w.Incarnation, Width: ViewportWidth, Height: ViewportHeight, DPI: 160})
	}
	constructed = true
	return &NativeBackend{Fixture: f, Driver: driver, Clock: clock, Executor: ex, manifest: m}, nil
}

func (b *NativeBackend) Manifest() replay.Manifest {
	m := b.manifest
	m.Windows = append([]replay.Window(nil), m.Windows...)
	m.Fonts = append([]replay.Font(nil), m.Fonts...)
	return m
}
func (b *NativeBackend) Close(ctx context.Context) error {
	err := b.Driver.Close(ctx)
	if err != nil {
		return err
	}
	b.Fixture.Owner.Close()
	b.Executor.Close()
	b.Clock.Close()
	return nil
}
func (b *NativeBackend) CheckManifest(ctx context.Context, m replay.Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !reflect.DeepEqual(m, b.manifest) {
		return fmt.Errorf("%w: fixture/theme/font/viewport identity", replay.ErrBackend)
	}
	if b.Fixture.Owner.Context().Err() != nil {
		return app.ErrClosed
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
func (b *NativeBackend) Model() (s Snapshot, err error) {
	err = b.Fixture.Owner.Do(func() { s = b.Fixture.Snapshot() })
	return
}

// Readiness includes origin FIFO, ready paint, owner delivery and admitted worker
// tails. Unrendered layout requires an explicit Frame operation, hence Blocked.
func (b *NativeBackend) Readiness(ctx context.Context) (replay.Readiness, error) {
	return b.Driver.ReplayReadiness(ctx)
}
func (b *NativeBackend) Step(ctx context.Context, n int) (replay.Progress, error) {
	r, err := b.Driver.Step(ctx, n)
	return replay.Progress{Units: r.Events + r.OwnerCallbacks, Events: r.Events, Owners: r.OwnerCallbacks, Frames: r.Frames}, err
}

func (b *NativeBackend) checkLive(ctx context.Context, o replay.Operation) (cogent.ManualWindowSnapshot, error) {
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
			return cogent.ManualWindowSnapshot{}, ErrStale
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
	codes := map[string]key.Codes{"Enter": key.CodeReturnEnter, "Escape": key.CodeEscape, "Tab": key.CodeTab, "Backspace": key.CodeBackspace, "Delete": key.CodeDelete, "Left": key.CodeLeftArrow, "Right": key.CodeRightArrow, "Up": key.CodeUpArrow, "Down": key.CodeDownArrow}
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
func (b *NativeBackend) Apply(ctx context.Context, o replay.Operation) error {
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
				return ErrAction
			}
			value = a.Value
		}
		var callErr error
		err := b.Fixture.Owner.Do(func() {
			callErr = b.Fixture.Invoke(ctx, Action{ID: o.Action, Value: value, CheckRevision: o.CheckRevision, ExpectedRevision: o.ExpectedRevision})
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

type NativeCapture struct {
	Model       Snapshot
	Window      string
	Incarnation uint64
	FocusedID   string
	Elements    []cogent.ManualElement
	Viewport    image.Rectangle
}

func (b *NativeBackend) Capture(ctx context.Context, o replay.Operation) (replay.Observation, error) {
	w, err := b.checkLive(ctx, o)
	if err != nil {
		return replay.Observation{}, err
	}
	pixels, err := b.Driver.Capture(ctx, o.Window)
	if err != nil {
		return replay.Observation{}, err
	}
	model, err := b.Model()
	if err != nil {
		return replay.Observation{}, err
	}
	data, err := json.Marshal(NativeCapture{Model: model, Window: w.ID, Incarnation: w.Incarnation, FocusedID: w.FocusedID, Elements: w.Elements, Viewport: pixels.Bounds()})
	if err != nil {
		return replay.Observation{}, err
	}
	h := sha256.Sum256(pixels.Pix)
	return replay.Observation{Data: data, ImageSHA256: hex.EncodeToString(h[:])}, nil
}

type NativeProof struct {
	Kind, FixtureSHA256, RecordSHA256 string
	Batch                             int
	Result                            replay.Result
}

// RunNative executes the same bundled record through the real managed router.
// All clocks, workers, frame requests and capture points come from that record.
func RunNative(ctx context.Context, batch int) (proof NativeProof, err error) {
	record, err := NativeRecord()
	if err != nil {
		return proof, err
	}
	b, err := NewNative(ctx, batch)
	if err != nil {
		return proof, err
	}
	defer func() { err = errors.Join(err, b.Close(context.Background())) }()
	r, err := replay.NewRunner(replay.RunnerConfig{EventBatch: batch}, b)
	if err != nil {
		return proof, err
	}
	proof = NativeProof{Kind: "actual-managed-software-replay", FixtureSHA256: FixtureSHA256(), RecordSHA256: NativeRecordSHA256(), Batch: batch}
	proof.Result, err = r.Run(ctx, record)
	if err == nil && (!proof.Result.QuiescentNow || proof.Result.Applied != len(record.Operations)) {
		err = fmt.Errorf("consumer: native record has residual current work %+v", proof.Result.Residual)
	}
	return proof, err
}
