package cogent

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"math"
	"slices"
)

func trackFraction(rect image.Rectangle, p image.Point, width float32) (float32, bool) {
	if rect.Empty() || width <= 24 {
		return 0, false
	}
	inset := float64(rect.Dx()) * 12 / float64(width)
	return max(0, min(1, float32((float64(p.X-rect.Min.X)-inset)/(float64(rect.Dx())-2*inset)))), true
}
func bindScalarTrack(r *record, b *core.Button) {
	b.FinalMaker(func(p *tree.Plan) {
		if r.node.Track == nil {
			return
		}
		p.Children = nil
		tree.AddAt(p, "scalar-track", func(w *core.Canvas) {
			w.Styler(func(s *styles.Style) {
				if t := r.node.Track; t != nil {
					s.Min.Set(units.Dp(t.Width), units.Dp(t.Height))
					s.Max = s.Min
					s.Grow.Set(0, 0)
					s.Padding.Zero()
					s.Margin.Zero()
					s.RenderBox = false
				}
			})
			w.SetDraw(func(pc *paint.Painter) {
				t := r.node.Track
				if t == nil {
					return
				}
				line, field, accent := colors.Scheme.OutlineVariant, colors.Scheme.SurfaceContainerLow, colors.Scheme.Primary.Base
				if r.theme.Palette.Set {
					p := r.theme.Palette.Value
					line = colors.Uniform(p.Line)
					field = colors.Uniform(p.Field)
					accent = colors.Uniform(p.Accent)
				}
				if controlDisabled(r) {
					accent = line
				}
				inset := float32(12) / t.Width
				span := 1 - 2*inset
				pc.Stroke.Color = nil
				if t.Gradient {
					for i := 0; i < 128; i++ {
						c := t.ColorAt(float32(i) / 127)
						for y := 0; y < 2; y++ {
							base := uint16(232)
							if (i/4+y)%2 != 0 {
								base = 190
							}
							blend := func(v uint8) uint8 { return uint8((uint16(v)*uint16(c.A) + base*(255-uint16(c.A)) + 127) / 255) }
							pc.Fill.Color = colors.Uniform(color.NRGBA{R: blend(c.R), G: blend(c.G), B: blend(c.B), A: 255})
							pc.Rectangle(inset+float32(i)/128*span, .1+float32(y)*.2, span/128+.001, .2)
							pc.Draw()
						}
					}
				} else {
					pc.Fill.Color = line
					pc.Rectangle(inset, .44, span, .12)
					pc.Draw()
					pc.Fill.Color = accent
					pc.Rectangle(inset+t.Handles[0].Position*span, .44, (t.Handles[len(t.Handles)-1].Position-t.Handles[0].Position)*span, .12)
					pc.Draw()
				}
				for _, h := range t.Handles {
					cx, cy := inset+h.Position*span, float32(.5)
					if t.Gradient {
						cy = .76
					}
					radius := float32(6)
					pc.Fill.Color = line
					if h.ID == t.ActiveID {
						pc.Fill.Color = accent
						radius = 8
					}
					pc.Ellipse(cx, cy, radius/t.Width, radius/t.Height)
					pc.Draw()
					pc.Fill.Color = field
					if t.Gradient {
						pc.Fill.Color = colors.Uniform(h.Color)
					}
					pc.Ellipse(cx, cy, (radius-2)/t.Width, (radius-2)/t.Height)
					pc.Draw()
				}
			})
		})
	})
	queue := func(origin *ui.TrackInput, fn func(*ui.TrackInput)) {
		postControl(r, func() {
			if !controlDisabled(r) && !r.control.retired && r.focusVisibility.Load() != 2 && origin.SameSource(r.node.Track) {
				fn(r.node.Track)
				requestControlRestore(r)
			}
		})
	}
	b.OnFirst(events.KeyChord, func(e events.Event) {
		t := r.node.Track
		if t == nil || e.IsHandled() || controlDisabled(r) || key.HasAnyModifier(e.Modifiers(), key.Meta, key.Control, key.Alt) {
			return
		}
		step := 1
		if key.HasAnyModifier(e.Modifiers(), key.Shift) {
			step = 10
		}
		switch e.KeyCode() {
		case key.CodeLeftArrow:
			step = -step
		case key.CodeRightArrow:
		case key.CodeUpArrow, key.CodeDownArrow:
			e.SetHandled()
			delta := 1
			if e.KeyCode() == key.CodeUpArrow {
				delta = -1
			}
			queue(t, func(c *ui.TrackInput) {
				i := 0
				for n, h := range c.Handles {
					if h.ID == c.Current() {
						i = n
					}
				}
				i = (i + delta + len(c.Handles)) % len(c.Handles)
				c.Select(c.Handles[i].ID)
			})
			return
		case key.CodeHome:
			e.SetHandled()
			queue(t, func(c *ui.TrackInput) { c.Set(c.Current(), 0) })
			return
		case key.CodeEnd:
			e.SetHandled()
			queue(t, func(c *ui.TrackInput) { c.Set(c.Current(), 1) })
			return
		default:
			return
		}
		e.SetHandled()
		queue(t, func(c *ui.TrackInput) { c.Adjust(c.Current(), step) })
	})
	b.OnFirst(events.Click, func(e events.Event) {
		if r.node.Track != nil && e.HasPos() {
			e.SetHandled()
		}
	})
	b.OnFirst(events.MouseDown, func(e events.Event) {
		t := r.node.Track
		if t == nil || controlDisabled(r) || b.Scene == nil || e.MouseButton() != events.Left {
			return
		}
		face, ok := b.ChildByName("scalar-track").(*core.Canvas)
		if !ok {
			return
		}
		fraction, ok := trackFraction(face.Geom.ContentBBox, e.Pos(), t.Width)
		if !ok {
			return
		}
		e.SetHandled()
		chosen := t.ActiveID
		distance := float32(math.Inf(1))
		for _, h := range t.Handles {
			d := float32(math.Abs(float64(h.Position - fraction)))
			if d < distance || d == distance && h.ID == t.ActiveID {
				chosen = h.ID
				distance = d
			}
		}
		scene := b.Scene
		var token uint64
		propose := func(point image.Point) {
			if value, ok := trackFraction(face.Geom.ContentBBox, point, t.Width); ok {
				queue(t, func(c *ui.TrackInput) { c.Set(chosen, value) })
			}
		}
		next, err := scene.Events.CapturePointer(b, e, func(event events.Event) {
			if !t.SameSource(r.node.Track) || controlDisabled(r) || r.control.retired || r.focusVisibility.Load() == 2 {
				scene.Events.ReleasePointer(token)
				return
			}
			switch event.Type() {
			case events.MouseMove, events.MouseDrag:
				propose(event.Pos())
			case events.MouseUp:
				propose(event.Pos())
				scene.Events.ReleasePointer(token)
			}
		}, func() {})
		if err != nil {
			scene.Events.ConsumePointerRelease()
			return
		}
		token = next
		b.SetFocusQuiet()
		queue(t, func(c *ui.TrackInput) { c.Select(chosen) })
		propose(e.Pos())
	})
}

func scalarTrackChanged(a, b *ui.TrackInput) bool {
	if a == nil || b == nil {
		return a != b
	}
	return !a.SameSource(b) || a.ActiveID != b.ActiveID || a.Min != b.Min || a.Max != b.Max || a.Step != b.Step || a.Width != b.Width || a.Height != b.Height || a.Gradient != b.Gradient || !slices.Equal(a.Handles, b.Handles)
}
