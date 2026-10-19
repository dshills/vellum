package ui

import (
	"fmt"
	"image/color"
	"math"
)

// Optional distinguishes an omitted property from an explicitly set zero value.
// Some(0) clears a dimension, gap or grow factor; Optional[T]{} restores defaults.
type Optional[T comparable] struct {
	Set   bool
	Value T
}

func Some[T comparable](value T) Optional[T] { return Optional[T]{Set: true, Value: value} }

// Insets and Size use device-independent pixels. Max size zero is unbounded.
type Insets struct{ Top, Right, Bottom, Left float32 }

func UniformInsets(value float32) Insets { return Insets{value, value, value, value} }

type Size struct{ Width, Height float32 }
type Alignment string

const (
	AlignStart  Alignment = "start"
	AlignCenter Alignment = "center"
	AlignEnd    Alignment = "end"
)

// Typography selects the backend default family when Family is empty. Weight
// is 0 (backend default) or a multiple of 100 from 100 through 900. Size must
// be positive. Font availability and shaping are supplied by the native driver.
type Typography struct {
	Family string
	Size   float32
	Weight int
	Italic bool
}

// Border is an app-local solid outline in device-independent pixels.
// A zero width explicitly removes a native widget's resting outline.
// Native keyboard-focus indicators retain precedence while a widget is focused.
type Border struct {
	Width float32
	Color color.NRGBA
}

// Style contains independently optional properties. Align controls children on
// the cross axis; Justify distributes children on the main axis. Grow sets both
// axes. Foreground and Background use straight-alpha sRGB bytes.
type Style struct {
	Padding          Optional[Insets]
	Gap              Optional[float32]
	Align, Justify   Optional[Alignment]
	MinSize, MaxSize Optional[Size]
	Grow             Optional[float32]
	// GrowX/GrowY override Grow on one axis, useful for sidebars and rows.
	GrowX, GrowY           Optional[float32]
	TextAlign              Optional[Alignment]
	CornerRadius           Optional[float32]
	Border                 Optional[Border]
	Foreground, Background Optional[color.NRGBA]
	Typography             Optional[Typography]
}

func finiteNonnegative(value float32) bool {
	return value >= 0 && !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
func (s Style) Validate() error {
	for name, property := range map[string]Optional[float32]{"gap": s.Gap, "grow": s.Grow, "horizontal grow": s.GrowX, "vertical grow": s.GrowY, "corner radius": s.CornerRadius} {
		if property.Set && !finiteNonnegative(property.Value) {
			return fmt.Errorf("ui: %s must be finite and nonnegative", name)
		}
	}
	if s.Border.Set && !finiteNonnegative(s.Border.Value.Width) {
		return fmt.Errorf("ui: border width must be finite and nonnegative")
	}
	if s.TextAlign.Set && s.TextAlign.Value != AlignStart && s.TextAlign.Value != AlignCenter && s.TextAlign.Value != AlignEnd {
		return fmt.Errorf("ui: unsupported text alignment %q", s.TextAlign.Value)
	}
	if s.Padding.Set {
		p := s.Padding.Value
		for _, v := range []float32{p.Top, p.Right, p.Bottom, p.Left} {
			if !finiteNonnegative(v) {
				return fmt.Errorf("ui: padding must be finite and nonnegative")
			}
		}
	}
	for name, property := range map[string]Optional[Size]{"min size": s.MinSize, "max size": s.MaxSize} {
		if property.Set && (!finiteNonnegative(property.Value.Width) || !finiteNonnegative(property.Value.Height)) {
			return fmt.Errorf("ui: %s must be finite and nonnegative", name)
		}
	}
	if s.Align.Set && s.Align.Value != AlignStart && s.Align.Value != AlignCenter && s.Align.Value != AlignEnd {
		return fmt.Errorf("ui: unsupported cross-axis alignment %q", s.Align.Value)
	}
	if s.Justify.Set {
		switch s.Justify.Value {
		case AlignStart, AlignCenter, AlignEnd:
		default:
			return fmt.Errorf("ui: unsupported justification %q", s.Justify.Value)
		}
	}
	if s.Typography.Set {
		p := s.Typography.Value
		if !finiteNonnegative(p.Size) || p.Size == 0 || (p.Weight != 0 && (p.Weight < 100 || p.Weight > 900 || p.Weight%100 != 0)) {
			return fmt.Errorf("ui: typography requires positive finite size and weight 0 or 100..900 in steps of 100")
		}
	}
	if s.MinSize.Set && s.MaxSize.Set {
		min, max := s.MinSize.Value, s.MaxSize.Value
		if (max.Width > 0 && max.Width < min.Width) || (max.Height > 0 && max.Height < min.Height) {
			return fmt.Errorf("ui: maximum size is smaller than minimum size")
		}
	}
	return nil
}

// Overlay returns a value snapshot, replacing only properties set by override.
func (s Style) Overlay(o Style) Style {
	if o.Padding.Set {
		s.Padding = o.Padding
	}
	if o.Gap.Set {
		s.Gap = o.Gap
	}
	if o.Align.Set {
		s.Align = o.Align
	}
	if o.Justify.Set {
		s.Justify = o.Justify
	}
	if o.MinSize.Set {
		s.MinSize = o.MinSize
	}
	if o.MaxSize.Set {
		s.MaxSize = o.MaxSize
	}
	if o.Grow.Set {
		s.Grow = o.Grow
	}
	if o.GrowX.Set {
		s.GrowX = o.GrowX
	}
	if o.GrowY.Set {
		s.GrowY = o.GrowY
	}
	if o.TextAlign.Set {
		s.TextAlign = o.TextAlign
	}
	if o.CornerRadius.Set {
		s.CornerRadius = o.CornerRadius
	}
	if o.Border.Set {
		s.Border = o.Border
	}
	if o.Foreground.Set {
		s.Foreground = o.Foreground
	}
	if o.Background.Set {
		s.Background = o.Background
	}
	if o.Typography.Set {
		s.Typography = o.Typography
	}
	return s
}

// Theme supplies application-local defaults and optional kind overrides. A
// driver takes a Clone at startup; callers must not mutate a live driver theme.
// No process-wide palette, font, or operating-system appearance is changed.
type Theme struct {
	Default Style
	Kinds   map[Kind]Style
}

func (t Theme) Clone() Theme {
	r := Theme{Default: t.Default}
	if t.Kinds != nil {
		r.Kinds = make(map[Kind]Style, len(t.Kinds))
		for k, s := range t.Kinds {
			r.Kinds[k] = s
		}
	}
	return r
}
func (t Theme) Validate() error {
	if err := t.Default.Validate(); err != nil {
		return fmt.Errorf("ui: default theme: %w", err)
	}
	for k, s := range t.Kinds {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("ui: theme kind %q: %w", k, err)
		}
		if err := t.Default.Overlay(s).Validate(); err != nil {
			return fmt.Errorf("ui: resolved theme kind %q: %w", k, err)
		}
	}
	return nil
}

// ResolveStyle applies theme default, theme kind, legacy dimensions/grow, then
// explicit node properties. Legacy zero dimensions/Grow=false leave theme
// values intact; Some(Size{}) and Some(float32(0)) explicitly clear them.
func ResolveStyle(n Node, theme Theme) (Style, error) {
	if !finiteNonnegative(n.MinWidth) || !finiteNonnegative(n.MinHeight) {
		return Style{}, fmt.Errorf("ui: node %q legacy minimum size must be finite and nonnegative", n.ID)
	}
	s := theme.Default.Overlay(theme.Kinds[n.Kind])
	if n.MinWidth > 0 || n.MinHeight > 0 {
		min := Size{}
		if s.MinSize.Set {
			min = s.MinSize.Value
		}
		if n.MinWidth > 0 {
			min.Width = n.MinWidth
		}
		if n.MinHeight > 0 {
			min.Height = n.MinHeight
		}
		s.MinSize = Some(min)
	}
	if n.Grow {
		s.Grow = Some(float32(1))
	}
	s = s.Overlay(n.Style)
	if err := s.Validate(); err != nil {
		return Style{}, fmt.Errorf("ui: node %q: %w", n.ID, err)
	}
	return s, nil
}
