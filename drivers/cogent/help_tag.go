package cogent

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"html"
	"image"
	"image/color"
)

func closeHelpTag(r *record) {
	if r == nil || r.helpTag == nil {
		return
	}
	sc := r.helpTag
	r.helpTag = nil
	if sc.Stage != nil {
		sc.Close()
	}
}
func showHelpTag(r *record) bool {
	w := r.widget.AsWidget()
	if r.node.Tooltip == "" || r.scope.Context().Err() != nil || controlDisabled(r) || !fileDropTargetActive(r, r.widget) || w.Scene == nil || w.Geom.TotalBBox.Empty() {
		return false
	}
	closeHelpTag(r)
	sc := core.NewScene(r.node.ID + "-help-tag")
	sc.SceneGeom.Pos = w.DefaultTooltipPos()
	sc.SceneGeom.Size = image.Pt(min(w.Geom.TotalBBox.Dx(), 320), w.Geom.TotalBBox.Dy())
	p := r.theme.Palette.Value
	if !r.theme.Palette.Set {
		p = ui.MacPalette(false, color.NRGBA{})
	}
	sc.Styler(func(s *styles.Style) {
		s.Padding.Set(units.Dp(7))
		s.Border.Radius.Set(units.Dp(5))
		s.Grow.Set(1, 1)
		s.Background = colors.Uniform(p.Popup)
		s.Color = colors.Uniform(p.Ink)
	})
	core.NewText(sc).SetText(html.EscapeString(r.node.Tooltip)).Styler(func(s *styles.Style) { s.SetTextWrap(true); s.Max.X = units.Dp(320); s.Font.Size = units.Dp(12) })
	r.helpTag = sc
	core.NewPopupStage(core.TooltipStage, sc, r.widget).Run()
	return true
}
func bindHelpTag(r *record) {
	w := r.widget.AsWidget()
	w.OnFirst(events.LongHoverStart, func(e events.Event) {
		if !e.IsHandled() && showHelpTag(r) {
			e.SetHandled()
		}
	})
	for _, typ := range []events.Types{events.LongHoverEnd, events.MouseLeave, events.MouseDown, events.FocusLost} {
		w.OnFirst(typ, func(events.Event) { closeHelpTag(r) })
	}
	w.OnFirst(events.KeyChord, func(e events.Event) {
		if e.IsHandled() {
			return
		}
		k, ok := e.(*events.Key)
		if !ok {
			return
		}
		if k.Code == key.CodeEscape && r.helpTag != nil {
			closeHelpTag(r)
			e.SetHandled()
		}
		if k.Code == key.CodeF1 && showHelpTag(r) {
			e.SetHandled()
		}
	})
}
