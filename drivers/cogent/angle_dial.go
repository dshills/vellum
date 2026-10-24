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
	"math"
)

func angleAt(rect image.Rectangle, p image.Point) (float32, bool) {
	if rect.Empty() {
		return 0, false
	}
	x := float64(p.X) - float64(rect.Min.X+rect.Max.X)/2
	y := float64(p.Y) - float64(rect.Min.Y+rect.Max.Y)/2
	if math.Hypot(x, y) < float64(min(rect.Dx(), rect.Dy()))*.1 {
		return 0, false
	}
	v := math.Atan2(x, -y) * 180 / math.Pi
	if v < 0 {
		v += 360
	}
	return float32(v), true
}
func bindAngleDial(r *record, b *core.Button) {
	b.FinalMaker(func(p *tree.Plan) {
		if r.node.AngleDial == nil {
			return
		}
		p.Children = nil
		tree.AddAt(p, "angle-face", func(w *core.Canvas) {
			w.Styler(func(s *styles.Style) {
				size := float32(96)
				if d := r.node.AngleDial; d != nil {
					size = d.Size
				}
				s.Min.Set(units.Dp(size), units.Dp(size))
				s.Max = s.Min
				s.Grow.Set(0, 0)
				s.Padding.Zero()
				s.Margin.Zero()
				s.RenderBox = false
			})
			w.SetDraw(func(pc *paint.Painter) {
				d := r.node.AngleDial
				if d == nil {
					return
				}

				line, field, secondary, accent := colors.Scheme.OutlineVariant, colors.Scheme.SurfaceContainerLow, colors.Scheme.OnSurfaceVariant, colors.Scheme.Primary.Base
				if r.theme.Palette.Set {
					p := r.theme.Palette.Value
					line = colors.Uniform(p.Line)
					field = colors.Uniform(p.Field)
					secondary = colors.Uniform(p.Secondary)
					accent = colors.Uniform(p.Accent)
				}

				pc.Fill.Color = line
				pc.Stroke.Color = nil
				pc.Circle(.5, .5, .47)
				pc.Draw()
				pc.Fill.Color = field
				pc.Circle(.5, .5, .45)
				pc.Draw()
				pc.Fill.Color = nil
				pc.Stroke.Color = secondary
				pc.Stroke.Width.Dp(1)
				for tick := 0; tick < 24; tick++ {
					angle := float64(tick) * math.Pi / 12
					inner := .405
					if tick%6 == 0 {
						inner = .36
					}
					pc.Line(.5+float32(math.Sin(angle)*inner), .5-float32(math.Cos(angle)*inner), .5+float32(math.Sin(angle)*.435), .5-float32(math.Cos(angle)*.435))
					pc.Draw()
				}
				col := accent
				if controlDisabled(r) {
					col = secondary
				}
				pc.Stroke.Color = col
				pc.Stroke.Width.Dp(2.5)
				angle := float64(d.Degrees) * math.Pi / 180
				pc.Line(.5, .5, .5+float32(math.Sin(angle)*.31), .5-float32(math.Cos(angle)*.31))
				pc.Draw()
				pc.Stroke.Color = nil
				pc.Fill.Color = col
				pc.Circle(.5, .5, .055)
				pc.Draw()
			})
		})
	})
	queue := func(source *ui.AngleDialInput, fn func(*ui.AngleDialInput)) {
		postControl(r, func() {
			if !controlDisabled(r) && !r.control.retired && r.focusVisibility.Load() != 2 && source.SameSource(r.node.AngleDial) {
				fn(r.node.AngleDial)
				requestControlRestore(r)
			}
		})
	}
	b.OnFirst(events.KeyChord, func(e events.Event) {
		d := r.node.AngleDial
		if d == nil || e.IsHandled() || controlDisabled(r) || key.HasAnyModifier(e.Modifiers(), key.Meta, key.Control, key.Alt) {
			return
		}
		step := 1
		if key.HasAnyModifier(e.Modifiers(), key.Shift) {
			step = 10
		}
		switch e.KeyCode() {
		case key.CodeLeftArrow, key.CodeDownArrow:
			step = -step
		case key.CodeRightArrow, key.CodeUpArrow:
		case key.CodeHome:
			e.SetHandled()
			queue(d, func(current *ui.AngleDialInput) { current.Set(0) })
			return
		case key.CodeEnd:
			e.SetHandled()
			queue(d, func(current *ui.AngleDialInput) { current.Set(360 - current.Step) })
			return
		default:
			return
		}
		e.SetHandled()
		queue(d, func(current *ui.AngleDialInput) { current.Adjust(step) })
	})
	b.OnFirst(events.Click, func(e events.Event) {
		if r.node.AngleDial != nil && e.HasPos() {
			e.SetHandled()
		}
	})

	b.OnFirst(events.MouseDown, func(e events.Event) {
		d := r.node.AngleDial
		if d == nil || controlDisabled(r) || b.Scene == nil || e.MouseButton() != events.Left {
			return
		}
		face, ok := b.ChildByName("angle-face").(*core.Canvas)
		if !ok {
			return
		}
		rect := face.Geom.ContentBBox
		p := e.Pos()
		center := rect.Min.Add(rect.Size().Div(2))
		if math.Hypot(float64(p.X-center.X), float64(p.Y-center.Y)) > float64(min(rect.Dx(), rect.Dy()))*.5 {
			return
		}
		e.SetHandled()
		scene := b.Scene
		var token uint64
		propose := func(p image.Point) {
			if value, ok := angleAt(face.Geom.ContentBBox, p); ok {
				queue(d, func(current *ui.AngleDialInput) { current.Set(value) })
			}
		}
		next, err := scene.Events.CapturePointer(b, e, func(event events.Event) {
			if !d.SameSource(r.node.AngleDial) || controlDisabled(r) || r.control.retired || r.focusVisibility.Load() == 2 {
				scene.Events.ReleasePointer(token)
				return
			}
			switch event.Type() {
			case events.MouseMove, events.MouseDrag:
				propose(event.Pos())
			case events.MouseUp:
				scene.Events.ReleasePointer(token)
			}
		}, func() {})
		if err != nil {
			scene.Events.ConsumePointerRelease()
			return
		}
		token = next
		b.SetFocusQuiet()
		propose(p)
	})
}
