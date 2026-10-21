package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

type helpDraft struct{ topic string }
type helpTopic struct {
	id, label  string
	paragraphs []string
}

var helpTopics = []helpTopic{
	{"gestures", "Mouse and canvas", []string{
		"Drag Block, Connector or Note from the palette to the canvas. Release over an accepted area to place once. Escape, outside release or focus loss cancels.",
		"Drag empty canvas to pan. Scroll to zoom at the pointer. Zoom offers In, Out, 100% and Fit Selection; Fit shows visible objects, Reset returns to origin at 100%.",
		"Click a component to select. Drag a selected component to move the selection. Drag a visible corner handle to resize it. Hold Shift on a selected corner for temporary proportional resizing.",
		"Drag layer rows to reorder; edge holding autoscrolls. Filtering disables layer dragging. Groups menu collapses or expands member rows without changing canvas visibility.",
	}},
	{"keyboard", "Keyboard shortcuts", []string{
		"Cmd+Shift+P: Search Commands. Type to filter; Up/Down chooses; Return runs; Escape cancels. Disabled results explain why they are unavailable. F1 opens this guide.",
		"Cmd+N: New. Cmd+O: Open. Cmd+S: Save. Cmd+Shift+S: Save As. Cmd+Z / Cmd+Shift+Z: Undo / Redo.",
		"Cmd+X / C / V: Cut / Copy / Paste. Cmd+D: Duplicate. Backspace or Delete: delete selected components. Focused text fields retain their usual editing shortcuts.",
		"Cmd+A / Cmd+Shift+A: Select All / Deselect All. Cmd+G / Cmd+Shift+G: Group / Ungroup. Cmd+Option+arrow: grid-aware movement; Nudge menu moves exactly 1 or 10 world units.",
		"Plain arrows pan. Option+Left/Right/Home/End selects components in paint order. Focus a palette tool and press Return to place at canvas center. Tab moves through enabled controls.",
	}},
	{"selection", "Selection and arrangement", []string{
		"Shift-click toggles complete components/groups. Shift-drag from empty canvas adds a marquee selection. With Subtract selection enabled, these gestures remove instead. Plain empty dragging continues to pan.",
		"Groups are flat interaction units: touching any member selects the whole group. Hide or Lock affects the whole group. Ungroup before editing a member independently.",
		"Geometry enters exact world X/Y/width/height. Lock aspect ratio preserves proportions. Grid visibility is independent of Snap; spacing controls snapped placement, moves and resizing.",
		"Match dimensions uses the primary component/group as reference and keeps it unchanged. Other units keep their top-left anchor; groups scale uniformly. Match size refuses incompatible group proportions.",
		"Exact spacing sorts units left-to-right or top-to-bottom, keeps the first fixed, and applies a nonnegative world-unit gap. Groups translate rigidly. Each changed arrangement is one Undo entry.",
	}},
	{"files", "Files, settings and export", []string{
		"Open and New ask before discarding unsaved changes. Cancel to keep the drawing and save first. Recent lists successful Opens, Saves and saved copies. Opening a drawing resets Undo history.",
		"More → Save a Copy writes a separate drawing without changing the active filename, dirty state or history. Copy destinations cannot refer to the active drawing. Native choosers handle overwrite confirmation.",
		"More → Export SVG writes visible rectangles in world coordinates and paint order, including locked objects. Hidden objects, grid and selection handles are omitted. SVG is output only; Open uses .vellum drawings.",
		"Private recovery offers the latest completed interrupted-session copy at startup. It is not a document version log. Explicit Save and Save a Copy remain the durable file workflow.",
		"Grid visibility/spacing, aspect lock and subtract mode are remembered separately from drawings. Snap is saved with the drawing. Filter/collapse state clears on New/Open.",
	}},
	{"limits", "Known limitations", []string{
		"This is an internal unsigned arm64 macOS test app. Framework scope is frozen and experimental rendering/pacing options remain disabled.",
		"Placement and editing are same-window. Groups are flat. Rotation, independent grouped-member editing, cross-window drag and SVG import are not available.",
		"Drawings are bounded to 512 components, origins within ±10000 and each component dimension from 8 to 2000 world units. Invalid multi-object edits reject the entire change.",
		"Undo retains up to 64 changes. Selection, navigation, settings and help do not add edit history. New/Open reset history; Save a Copy and export leave it intact.",
		"Automated routing and generated/native input checks are distinct from physical-device and cross-display qualification, which remain deferred.",
	}},
}

func (m *model) bindHelpGuide() error {
	return m.commands.Register(command.Command{ID: "help-guide", Label: "Help and shortcuts", Shortcut: "F1", Available: m.selectionAvailable, Execute: func(context.Context) error {
		if err := m.selectionAvailable(); err != nil {
			return err
		}
		if m.modalEpoch == ^uint64(0) {
			return errors.New("dialog focus sequence exhausted")
		}
		m.helpDraft = &helpDraft{topic: helpTopics[0].id}
		m.prompt = "help-guide"
		m.modalEpoch++
		return nil
	}})
}
func (m *model) helpGuideDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.helpDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.helpDraft == d && m.prompt == "help-guide"
	}
	close := func() {
		if active() {
			m.helpDraft = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	topic := "gestures"
	if d != nil {
		topic = d.topic
	}
	options := []ui.Option{}
	paragraphs := []ui.Node{}
	for _, t := range helpTopics {
		options = append(options, ui.Option{ID: t.id, Label: t.label})
		if t.id == topic {
			for i, text := range t.paragraphs {
				paragraphs = append(paragraphs, ui.Node{ID: "help-paragraph-" + string(rune('a'+i)), Kind: ui.Label, Text: text, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 560})}})
			}
		}
	}
	return ui.Node{ID: "help-guide-dialog", Kind: ui.Dialog, Text: "Vellum Placement Help", Open: active(), Dismiss: close, Children: []ui.Node{
		{ID: "help-topic", Kind: ui.Selector, Options: options, SelectedID: topic, Change: func(e ui.ChangeEvent) {
			if !active() {
				return
			}
			for _, t := range helpTopics {
				if t.id == e.SelectedID {
					d.topic = t.id
					_ = a.Invalidate("placement")
					return
				}
			}
		}},
		{ID: "help-content", Kind: ui.Column, Children: paragraphs},
		{ID: "help-close", Kind: ui.Button, Text: "Close help", Activate: close},
	}}
}
