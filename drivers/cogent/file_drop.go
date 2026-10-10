package cogent

import (
	"math"

	"cogentcore.org/core/base/fileinfo/mimedata"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

const fileDropRecordProperty = "vellum.file-drop.record"

// Install once, even without a callback, so retained callback updates need no
// remount. A dialog's visible host is moved into its own modal scene.
func bindFileDrop(r *record, a *app.Application, window string) {
	bind := func(widget core.Widget) {
		widget.AsTree().SetProperty(fileDropRecordProperty, r)
		widget.AsWidget().OnFirst(events.Drop, func(e events.Event) {
			dispatchFileDrop(r, widget, a, window, e)
		})
	}
	bind(r.widget)
	if r.control != nil && r.control.dialog != nil {
		bind(r.control.dialog.host)
	}
}

func externalFileDropPaths(e *events.DragDrop) ([]string, error) {
	if e.Source != nil {
		return nil, app.ErrUnavailable
	}
	data, ok := e.Data.(mimedata.Mimes)
	if !ok || len(data) == 0 || len(data) > ui.MaxFileDropPaths {
		return nil, app.ErrUnavailable
	}
	total := 0
	// Check byte bounds before converting bytes to strings or copying callbacks.
	for _, item := range data {
		if item == nil || item.Type != mimedata.TextPlain || len(item.Data) > ui.MaxFileDropBytes-total {
			return nil, app.ErrUnavailable
		}
		total += len(item.Data)
	}
	paths := make([]string, len(data))
	for i, item := range data {
		paths[i] = string(item.Data)
	}
	return ui.CopyFileDropPaths(paths)
}

// Called under the originating scene's native event lock and the model owner.
// Native visibility and bounds protect clipping; retained records additionally
// protect controlled disabled state, inactive pages, and closed modal hosts.
func fileDropTargetActive(r *record, widget core.Widget) bool {
	if tree.IsNil(widget) || !widget.AsWidget().IsVisible() || controlDisabled(r) {
		return false
	}
	for p := tree.Node(widget); p != nil; p = p.AsTree().Parent {
		if disabled, ok := p.AsTree().Property(disabledProperty).(bool); ok && disabled {
			return false
		}
		if active, ok := p.AsTree().Property("vellum.workbench.active").(bool); ok && !active {
			return false
		}
		if owner, ok := p.AsTree().Property(fileDropRecordProperty).(*record); ok {
			if owner.node.Disabled {
				return false
			}
			if owner.control != nil && (owner.control.retired || (owner.control.dialog != nil && (!owner.node.Open || owner.control.dialog.dismissed))) {
				return false
			}
			if tabs, ok := p.(*core.Tabs); ok && p != widget {
				current, _ := tabs.CurrentTab()
				if tree.IsNil(current) || widget.AsTree().ParentLevel(current) < 0 {
					return false
				}
				// A native tab click can optimistically select a page before its
				// queued controlled Change is accepted by the model.
				if owner.control != nil && owner.control.tabs != nil && current != owner.control.tabs.pages[owner.node.SelectedID] {
					return false
				}
				for _, option := range owner.node.Options {
					if option.ID == owner.node.SelectedID && option.Disabled {
						return false
					}
				}
			}
		}
	}
	return true
}

func dispatchFileDrop(r *record, widget core.Widget, a *app.Application, window string, event events.Event) {
	e, ok := event.(*events.DragDrop)
	if !ok || e.Type() != events.Drop || e.IsHandled() {
		return
	}
	// Synchronous admission preserves deepest-first native propagation. No model
	// callback is queued against a retained widget or a later window lifetime.
	_ = a.Do(func() {
		lifetime, err := a.WindowContext(window)
		if err != nil || lifetime.Err() != nil || r.node.FileDrop == nil || !fileDropTargetActive(r, widget) {
			return
		}
		w := widget.AsWidget()
		if !e.Pos().In(w.Geom.TotalBBox) {
			return
		}
		scale := w.Styles.UnitContext.Dp(1)
		if scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
			return
		}
		paths, err := externalFileDropPaths(e)
		if err != nil {
			return
		}
		pos := ui.Point{X: (float32(e.Pos().X) - w.Geom.Pos.Total.X) / scale, Y: (float32(e.Pos().Y) - w.Geom.Pos.Total.Y) / scale}
		if r.node.FileDrop(ui.FileDropEvent{ID: r.node.ID, WindowID: window, Paths: paths, Position: pos}) {
			e.Target = widget
			e.DropMod = events.DropCopy
			e.SetHandled()
			a.Invalidate(window)
		}
	})
}
