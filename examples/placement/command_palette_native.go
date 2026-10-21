package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
)

// This app-local extension supplies live search and dialog keys using the
// existing pinned native event API; no framework input behavior is modified.
type commandPaletteKeys struct {
	a   *app.Application
	ref app.WindowRef
	m   *model
}

func (e *commandPaletteKeys) Mount(parent tree.Node) core.Widget {
	w := core.NewFrame(parent)
	w.Styler(func(s *styles.Style) { s.Display = styles.DisplayNone })
	return w
}
func (e *commandPaletteKeys) Sync(w core.Widget) error {
	parent := w.AsTree().Parent
	if parent == nil {
		return nil
	}
	node := parent.AsTree().ChildByName("command-query")
	if node == nil {
		return nil
	}
	field, ok := node.(*core.TextField)
	if !ok {
		return nil
	}
	const prop = "vellum-placement-command-keys"
	field.SetProperty(prop+"-draft", e.m.commandDraft)
	update := func(run bool, delta int) {
		if _, _, active := field.NativeCompositionRange(); active {
			return
		}
		text := field.EditingState().Text
		expected, _ := field.Property(prop + "-draft").(*commandPaletteDraft)
		_ = e.a.Do(func() {
			if e.ref.Context().Err() != nil {
				return
			}
			d := expected
			if !e.m.commandDraftActive(d) {
				return
			}
			if !e.m.updateCommandQuery(d, text) {
				_ = e.a.Invalidate("placement")
				return
			}
			if delta != 0 {
				e.m.moveCommandChoice(d, delta)
			}
			if run {
				if err := e.m.runCommandChoice(e.ref.Context(), d); err != nil {
					e.m.status = err.Error()
				}
			}
			_ = e.a.Invalidate("placement")
		})
	}
	if field.Property(prop) != e {
		field.SetProperty(prop, e)
		field.OnInput(func(events.Event) { update(false, 0) })
	}
	var bind func(tree.Node)
	bind = func(n tree.Node) {
		base := core.AsWidget(n)
		if base != nil && base.Property(prop+"-key") != e {
			base.SetProperty(prop+"-key", e)
			base.OnFirst(events.KeyChord, func(event events.Event) {
				if event.IsHandled() || event.Modifiers() != 0 || e.ref.Context().Err() != nil {
					return
				}
				if _, _, active := field.NativeCompositionRange(); active {
					return
				}
				switch event.KeyCode() {
				case key.CodeReturnEnter, key.CodeKeypadEnter:
					event.SetHandled()
					update(true, 0)
				case key.CodeUpArrow:
					if base == &field.WidgetBase {
						event.SetHandled()
						update(false, -1)
					}
				case key.CodeDownArrow:
					if base == &field.WidgetBase {
						event.SetHandled()
						update(false, 1)
					}
				}
			})
		}
		for _, child := range n.AsTree().Children {
			bind(child)
		}
	}
	bind(field)
	if results := parent.AsTree().ChildByName("command-results"); results != nil {
		bind(results)
	}
	return nil
}
