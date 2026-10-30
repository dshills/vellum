package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"slices"
)

const buttonImageName = "vellum-color-icon"

// Artwork stays private to the same native button, including hit testing,
// focus, activation and accessibility. It uses the bounded shared-image painter.
func installButtonImage(r *record, button *core.Button) {
	button.Styler(func(s *styles.Style) {
		if r.node.IconPlacement == ui.IconAboveText {
			s.Direction = styles.Column
			s.IconSize.Set(units.Em(2))
			s.Gap.Set(units.Dp(4))
		} else {
			s.Direction = styles.Row
			if r.node.IconImage != nil && r.node.Text != "" {
				s.Gap.Set(units.Dp(6))
			}
		}
	})
	button.FinalMaker(func(p *tree.Plan) {
		if r.node.IconPlacement != ui.IconAboveText {
			return
		}
		// The pinned native Button inserts an inline icon/text spacer. Stacked
		// content uses its managed gap instead; the next inline plan restores it.
		p.Children = slices.DeleteFunc(p.Children, func(item *tree.PlanItem) bool { return item.Name == "space" })
	})
	button.FirstMaker(func(p *tree.Plan) {
		if r.control.retired || r.node.IconImage == nil {
			return
		}
		tree.AddAt(p, buttonImageName, func(w *managedImage) {
			w.Styler(func(s *styles.Style) {
				s.Grow.Set(0, 0)
				s.Padding.Zero()
				s.Margin.Zero()
				s.Min = s.IconSize
				s.Max = s.IconSize
			})
			w.Updater(func() {
				if r.control.retired {
					w.source = nil
					return
				}
				w.source = r.node.IconImage
				w.fit = ui.ImageContain
			})
		})
		// FirstMaker precedes the native button label and indicator makers.

	})
}
func disposeButtonImage(r *record) {
	if r.node.IconImage == nil {
		return
	}
	r.node.IconImage = nil
	if button, ok := r.widget.(*core.Button); ok {
		if image, ok := button.ChildByName(buttonImageName).(*managedImage); ok {
			image.source = nil
		}
	}
}
