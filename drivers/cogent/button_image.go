package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

const buttonImageName = "vellum-color-icon"

// Artwork stays private to the same native button, including hit testing,
// focus, activation and accessibility. It uses the bounded shared-image painter.
func installButtonImage(r *record, button *core.Button) {
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
