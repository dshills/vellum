package cogent

import (
	"math"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

type numberFieldState struct {
	buttons           ui.NumberFieldButtons
	direction         float32
	leading, trailing func(events.Event)
	step              func(*core.Spinner, float32) float32
}

func bindNumberField(r *record, sp *core.Spinner) {
	r.control.number = &numberFieldState{direction: 1, leading: sp.LeadingIconOnClick, trailing: sp.TrailingIconOnClick, step: sp.StepFunc}
	sp.Maker(func(p *tree.Plan) {
		if r.node.NumberButtons != ui.NumberButtonsStepper || sp.IsReadOnly() || sp.IsDisabled() {
			return
		}
		tree.AddInit(p, "lead-icon", func(button *core.Button) {
			button.Styler(func(s *styles.Style) {
				if r.node.NumberButtons == ui.NumberButtonsStepper {
					s.Display = styles.DisplayNone
				}
			})
		})
		tree.AddInit(p, "trail-icon", func(button *core.Button) {
			button.SetTooltip("Increase above; decrease below")
			button.OnFirst(events.MouseDown, func(e events.Event) {
				if r.node.NumberButtons != ui.NumberButtonsStepper {
					return
				}
				box := button.Geom.TotalBBox
				r.control.number.direction = 1
				if e.Pos().Y >= box.Min.Y+box.Dy()/2 {
					r.control.number.direction = -1
				}
			})
			button.Styler(func(s *styles.Style) {
				if r.node.NumberButtons != ui.NumberButtonsStepper {
					return
				}
				s.Min.Set(units.Dp(18), units.Dp(24))
				s.Max.Set(units.Dp(18), units.Dp(24))
				s.IconSize.Set(units.Dp(16))
				s.Margin.Zero()
				s.Padding.Zero()
				s.Color = sp.Styles.Color
				s.Background = sp.Styles.Background
				s.Border.Width.Set(units.Dp(1))
				s.Border.Style.Set(styles.BorderSolid)
				s.Border.Color = sp.Styles.Border.Color
				s.Border.Radius.Set(units.Dp(4))
			})
		})
	})
}

// Native fields and their event handlers remain on the existing UI owner.
// The single private arrow button reserves the ordinary trailing-icon space;
// its two halves use the same clamped increment function as keyboard stepping.
func syncNumberField(r *record, sp *core.Spinner) {
	state := r.control.number
	if state.buttons == r.node.NumberButtons {
		return
	}
	state.buttons = r.node.NumberButtons
	if state.buttons == ui.NumberButtonsStepper {
		sp.SetLeadingIcon(icons.Remove)
		sp.StepFunc = func(s *core.Spinner, steps float32) float32 {
			value := float64(s.Value) + float64(steps)*float64(s.Step)
			return float32(math.Max(float64(s.Min), math.Min(float64(s.Max), value)))
		}
		sp.SetTrailingIcon(icons.UnfoldMore, func(e events.Event) {
			if sp.IsReadOnly() || sp.IsDisabled() {
				return
			}
			button := sp.ChildByName("trail-icon", 0)
			if button == nil {
				return
			}
			box := button.(core.Widget).AsWidget().Geom.TotalBBox
			delta := state.direction
			if e.MouseButton() != events.NoButton {
				delta = 1
				if e.Pos().Y >= box.Min.Y+box.Dy()/2 {
					delta = -1
				}
			}
			before := sp.Value
			sp.SetValue(sp.StepFunc(sp, delta))
			if sp.Value != before {
				sp.SendChange()
			}
		})
	} else {
		sp.SetLeadingIcon(icons.Remove, state.leading)
		sp.SetTrailingIcon(icons.Add, state.trailing)
		sp.StepFunc = state.step
		if button := sp.ChildByName("trail-icon", 0); button != nil {
			button.(*core.Button).SetTooltip("")
		}
	}
	if part := sp.ChildByName("lead-icon", 0); part != nil {
		part.(core.Widget).AsWidget().SetState(state.buttons == ui.NumberButtonsStepper, states.Invisible)
	}
	sp.Update()
}
