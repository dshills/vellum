package main

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"math"
	"slices"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
)

// Components are immutable values. Applications can use their own payload type.
type component struct {
	Kind          string
	Width, Height float64
	Color         color.RGBA
}
type model struct {
	visibleScene                      *canvas.Scene
	scene                             *canvas.Scene
	palette                           *placement.Controller[component]
	target                            *placement.Target
	commands                          *command.Registry
	journal                           *history.Journal
	pending                           *placement.Request[component]
	next                              uint64
	snap, locked, rejectNotes         bool
	status                            string
	files                             *document.Session[drawingFile]
	recent                            *recentDrawings
	geometry                          *geometryDraft
	aspectLocked                      bool
	selectionSubtract                 bool
	pendingOpenPath                   string
	recovery                          *placementRecovery
	removeRecoveryOnClose             bool
	prompt                            string
	move                              *objectMove
	canvasEditor                      *canvasEditor
	selectionIDs                      []string
	selectionPrimary                  string
	groups                            map[string]string
	nextGroup                         uint64
	clipboardJob                      *clipboardOperation
	pasteText                         string
	pasteCount                        uint64
	renameEpoch                       uint64
	layerInfo                         map[string]layerMetadata
	layerFocus, renameID, renameDraft string
	layerAdd                          bool
	layerDrag                         *layerDrag
	layerEditor                       *layerEditor
}

func componentByKind(kind string) (component, bool) {
	for _, c := range components() {
		if c.Kind == kind {
			return c, true
		}
	}
	return component{}, false
}
func components() []component {
	return []component{{"block", 80, 50, color.RGBA{48, 117, 209, 255}}, {"connector", 120, 16, color.RGBA{35, 150, 110, 255}}, {"note", 90, 65, color.RGBA{235, 180, 50, 255}}}
}

func newModel() (*model, error) {
	scene, err := canvas.New(nil)
	if err != nil {
		return nil, err
	}
	journal, err := history.New(64)
	if err != nil {
		return nil, err
	}
	m := &model{scene: scene, journal: journal, commands: command.New(), snap: true, status: "Drag a tool onto the canvas, or focus a tool and press Return to place at center."}
	for _, c := range []command.Command{
		{ID: "place", Label: "Place component", Execute: func(ctx context.Context) error { return m.commit(ctx) }},
		{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", Available: func() error {
			if m.fileBusy() || m.move != nil || m.prompt != "" || m.palette.State().Active {
				return document.ErrBusy
			}
			if m.journal.Snapshot().UndoCount == 0 {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.journal.Undo() }},
		{ID: "redo", Label: "Redo", Shortcut: "Meta+Shift+Z", Available: func() error {
			if m.fileBusy() || m.move != nil || m.prompt != "" || m.palette.State().Active {
				return document.ErrBusy
			}
			if m.journal.Snapshot().RedoCount == 0 {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error { return m.journal.Redo() }},
	} {
		if err := m.commands.Register(c); err != nil {
			return nil, err
		}
	}
	for _, edit := range []struct {
		id, label, shortcut string
		dx, dy              float64
	}{
		{"move-left", "Move left", "Meta+Alt+LeftArrow", -1, 0},
		{"move-right", "Move right", "Meta+Alt+RightArrow", 1, 0},
		{"move-up", "Move up", "Meta+Alt+UpArrow", 0, -1},
		{"move-down", "Move down", "Meta+Alt+DownArrow", 0, 1},
		{"delete", "Delete selected", "Backspace", 0, 0},
		{"delete-forward", "Delete selected", "Delete", 0, 0},
	} {
		scope := command.WindowShortcuts
		if edit.id == "delete" || edit.id == "delete-forward" {
			scope = command.AfterFocusShortcuts
		}
		if err := m.commands.Register(command.Command{ID: edit.id, Label: edit.label, Shortcut: edit.shortcut, ShortcutScope: scope, Available: m.editAvailable, Execute: func(context.Context) error {
			if edit.id == "delete" || edit.id == "delete-forward" {
				return m.deleteSelected()
			}
			return m.nudge(edit.dx, edit.dy)
		}}); err != nil {
			return nil, err
		}
	}

	for _, resize := range []struct {
		id, label string
		dw, dh    float64
	}{
		{"width-less", "Width −", -1, 0}, {"width-more", "Width +", 1, 0},
		{"height-less", "Height −", 0, -1}, {"height-more", "Height +", 0, 1},
	} {
		if err := m.commands.Register(command.Command{ID: resize.id, Label: resize.label, Available: m.resizeAvailable, Execute: func(context.Context) error { return m.resizeSelected(resize.dw, resize.dh) }}); err != nil {
			return nil, err
		}
	}
	items := []placement.Item[component]{}
	for _, c := range components() {
		items = append(items, placement.Item[component]{ID: c.Kind, Label: map[string]string{"block": "Block", "connector": "Connector", "note": "Note"}[c.Kind], Payload: c, Preview: canvas.Rect{Width: c.Width, Height: c.Height, Color: c.Color}})
	}
	m.palette, err = placement.New(placement.Config[component]{Items: items, Command: func(ctx context.Context, r placement.Request[component]) error {
		m.pending = &r
		defer func() { m.pending = nil }()
		return m.commands.Invoke(ctx, "place")
	}})
	if err != nil {
		return nil, err
	}
	m.target, err = m.palette.NewTarget(placement.TargetConfig[component]{ID: "drawing", Snap: func(r placement.Request[component]) (canvas.Point, error) {
		p := r.Position
		if m.snap {
			p.X = math.Round(p.X/20) * 20
			p.Y = math.Round(p.Y/20) * 20
		}
		return p, nil
	}, Accept: func(r placement.Request[component]) error {
		if m.fileBusy() || m.move != nil || m.prompt != "" || m.locked || (m.rejectNotes && r.Payload.Kind == "note") || len(m.scene.Objects) >= 512 || math.Abs(r.Position.X) > 10000 || math.Abs(r.Position.Y) > 10000 {
			return placement.ErrRejected
		}
		return nil
	}})
	if err != nil {
		m.palette.Close()
		return nil, err
	}
	if err := m.bindGrouping(); err != nil {
		m.palette.Close()
		return nil, err
	}
	if err := m.bindStacking(); err != nil {
		m.palette.Close()
		return nil, err
	}
	if err := m.bindArrangement(); err != nil {
		m.palette.Close()
		return nil, err
	}
	return m, nil
}
func (m *model) commit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.pending == nil || m.next == math.MaxUint64 {
		return errors.New("no pending placement or object IDs exhausted")
	}
	r := *m.pending
	before := slices.Clone(m.scene.Objects)
	object := canvas.Rect{ID: fmt.Sprintf("%s-%d", r.Payload.Kind, m.next+1), X: r.Position.X, Y: r.Position.Y, Width: r.Payload.Width, Height: r.Payload.Height, Color: r.Payload.Color}
	after := append(slices.Clone(before), object)
	err := m.applyDocument("Place "+r.Payload.Kind, after, m.groups, []string{object.ID}, object.ID)
	if err == nil {
		m.next++
		m.status = fmt.Sprintf("Placed %s at (%.0f, %.0f)", object.ID, object.X, object.Y)
	}
	return err
}
