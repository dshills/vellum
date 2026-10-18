package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type textModel struct {
	Name, Note, Visible                                                            string
	NameRequests, Edits, Undo, Redo, Opens, Dismisses, Applies                     int
	Modal, RejectName, DisableName, DisableEditor, Dirty, Composing, Busy, Invalid bool
}

func (g *gallery) textModel() textModel {
	s := g.notes.Snapshot()
	return textModel{g.text, s.Text, s.VisibleText, g.textRequests, g.editRequests, s.History.UndoCount, s.History.RedoCount, g.dialogOpens, g.dialogDismisses, g.dialogApplies, g.dialog, g.textReject, g.textDisabled, g.notesDisabled, s.History.Dirty, s.Composing, s.History.Busy, s.History.Invalid}
}

type textStep struct {
	Label, Kind, Target, Focus, Text string
	Code                             key.Codes
	Rune                             rune
	Modifiers                        key.Modifiers
	Want                             textModel
	Setup                            func()
}

func (g *gallery) textScript() []textStep {
	w := textModel{Name: "Vellum", Note: "Draft\nsecond line\n", Visible: "Draft\nsecond line\n"}
	var steps []textStep
	add := func(label, kind, target, focus, text string, code key.Codes, r rune, mods key.Modifiers, setup func()) {
		steps = append(steps, textStep{label, kind, target, focus, text, code, r, mods, w, setup})
	}
	meta := catalogueSelectAllModifiers()
	redo := meta
	redo.SetFlag(true, key.Shift)
	add("initial", "observe", "", "text-name", "", 0, 0, 0, nil)
	w.Name = "Atlas"
	w.NameRequests = 1
	add("name-accept", "replace", "text-name", "", "Atlas", 0, 0, 0, nil)
	w.RejectName = true
	add("name-reject-policy", "observe", "", "", "", 0, 0, 0, func() { g.textReject = true })
	w.NameRequests = 2
	add("name-rejected", "replace", "text-name", "", "Rejected", 0, 0, 0, nil)
	w.RejectName = false
	w.DisableName = true
	add("name-disabled", "observe", "", "", "", 0, 0, 0, func() { g.textReject = false; g.textDisabled = true })
	add("name-disabled-refusal", "refuse", "text-name", "", "", 0, 0, 0, nil)
	w.DisableName = false
	add("name-enabled", "observe", "", "", "", 0, 0, 0, func() { g.textDisabled = false })
	add("editor-focus", "click", "text-notes", "text-notes", "", 0, 0, 0, nil)
	add("editor-select-all", "key", "text-notes", "text-notes", "", key.CodeA, 'a', meta, nil)
	w.Note = "A"
	w.Visible = w.Note
	w.Edits = 1
	w.Undo = 1
	w.Dirty = true
	add("editor-replace", "key", "text-notes", "text-notes", "", key.CodeA, 'A', 0, nil)
	w.Note = "A😀"
	w.Visible = w.Note
	w.Edits = 2
	w.Undo = 2
	add("editor-supplementary", "key", "text-notes", "text-notes", "", 0, '😀', 0, nil)
	w.Note = "A😀\n"
	w.Visible = w.Note
	w.Edits = 3
	w.Undo = 3
	add("editor-newline", "key", "text-notes", "text-notes", "", key.CodeReturnEnter, 0, 0, nil)
	w.Note = "A😀\ne"
	w.Visible = w.Note
	w.Edits = 4
	w.Undo = 4
	add("editor-base", "key", "text-notes", "text-notes", "", key.CodeE, 'e', 0, nil)
	w.Note = "A😀\ne\u0301"
	w.Visible = w.Note
	w.Edits = 5
	w.Undo = 5
	add("editor-combining", "key", "text-notes", "text-notes", "", 0, '\u0301', 0, nil)
	w.Note = "A😀\n"
	w.Visible = w.Note
	w.Edits = 6
	w.Undo = 6
	add("editor-grapheme-delete", "key", "text-notes", "text-notes", "", key.CodeBackspace, 0, 0, nil)
	w.Note = "A😀\ne\u0301"
	w.Visible = w.Note
	w.Edits = 7
	w.Undo = 5
	w.Redo = 1
	add("editor-undo-delete", "key", "text-notes", "text-notes", "", key.CodeZ, 'z', meta, nil)
	w.Note = "A😀\ne"
	w.Visible = w.Note
	w.Edits = 8
	w.Undo = 4
	w.Redo = 2
	add("editor-undo-combining", "key", "text-notes", "text-notes", "", key.CodeZ, 'z', meta, nil)
	w.Note = "A😀\ne\u0301"
	w.Visible = w.Note
	w.Edits = 9
	w.Undo = 5
	w.Redo = 1
	add("editor-redo-combining", "key", "text-notes", "text-notes", "", key.CodeZ, 'Z', redo, nil)
	add("editor-tab", "key", "text-notes", "text-preferences", "", key.CodeTab, 0, 0, nil)
	w.DisableEditor = true
	add("editor-disabled", "observe", "", "text-preferences", "", 0, 0, 0, func() { g.notesDisabled = true })
	add("editor-disabled-refusal", "refuse", "text-notes", "text-preferences", "", 0, 0, 0, nil)
	w.DisableEditor = false
	add("editor-enabled", "observe", "", "text-preferences", "", 0, 0, 0, func() { g.notesDisabled = false })
	w.Modal = true
	w.Opens = 1
	add("modal-open", "click", "text-preferences", "dialog-name", "", 0, 0, 0, nil)
	add("modal-background-refusal", "background", "", "dialog-name", "", 0, 0, 0, nil)
	w.Name = "Modal"
	w.NameRequests = 3
	add("modal-name-accept", "replace", "dialog-name", "", "Modal", 0, 0, 0, nil)
	add("modal-refocus", "click", "dialog-name", "dialog-name", "", 0, 0, 0, nil)
	w.Modal = false
	w.Dismisses = 1
	add("modal-escape", "key", "dialog-name", "", "", key.CodeEscape, 0, 0, nil)
	add("modal-retired-refusal", "retired", "", "", "", 0, 0, 0, nil)
	w.Modal = true
	w.Opens = 2
	add("modal-retained-reopen", "click", "text-preferences", "dialog-name", "", 0, 0, 0, nil)
	w.Name = "Final"
	w.NameRequests = 4
	add("modal-name-second", "replace", "dialog-name", "", "Final", 0, 0, 0, nil)
	w.Modal = false
	w.Applies = 1
	add("modal-apply", "click", "apply", "", "", 0, 0, 0, nil)
	add("modal-stale-frame-refusal", "stale", "", "", "", 0, 0, 0, nil)
	w.Modal = true
	w.Opens = 3
	add("modal-before-close", "click", "text-preferences", "dialog-name", "", 0, 0, 0, nil)
	return steps
}

type textObservation struct {
	Label    string
	Model    textModel
	Snapshot cogent.DesktopSnapshot
}
type textReport struct {
	Event, Result, Error, ActiveStep string
	Steps                            []textObservation
	Inputs                           []catalogueInput
	Failed                           *textObservation
	Unavailable                      bool
	Session                          cogent.DesktopState
	OriginalRefClosed                bool
	PostCloseInputError              string
}

func textObserve(s cogent.DesktopSnapshot, step textStep) error {
	visible := func(id string) (cogent.DesktopElement, error) {
		e, err := catalogueElement(s, id)
		if err == nil && (e.Bounds.Empty() || !e.Bounds.In(image.Rectangle{Max: s.Size})) {
			err = fmt.Errorf("%s not fully visible", id)
		}
		return e, err
	}
	if step.Focus != "" {
		e, err := visible(step.Focus)
		if err != nil {
			return err
		}
		if !e.Focused {
			return fmt.Errorf("actual focus %s absent", step.Focus)
		}
	}
	if step.Want.Modal {
		for _, id := range []string{"preferences", "dialog-name", "apply"} {
			if _, err := visible(id); err != nil {
				return err
			}
		}
		e, _ := visible("dialog-name")
		var v string
		if json.Unmarshal(e.Value, &v) != nil || v != step.Want.Name || !e.Enabled {
			return errors.New("modal accepted text publication")
		}
		for _, id := range []string{"text-name", "text-notes", "text-preferences"} {
			if e, err := catalogueElement(s, id); err == nil && e.Enabled {
				return fmt.Errorf("modal enabled background %s", id)
			}
		}
	} else {
		for _, id := range []string{"text-title", "text-name", "text-notes", "text-preferences", "text-status"} {
			if _, err := visible(id); err != nil {
				return err
			}
		}
		for _, id := range []string{"dialog-name", "apply"} {
			if e, err := catalogueElement(s, id); err == nil && !e.Bounds.Empty() && e.Enabled {
				return errors.New("retired dialog target remains eligible")
			}
		}
		for id, want := range map[string]string{"text-name": step.Want.Name, "text-notes": step.Want.Note, "text-status": fmt.Sprintf("Project %s / edits %d / undo %d / redo %d", step.Want.Name, step.Want.Edits, step.Want.Undo, step.Want.Redo)} {
			e, err := visible(id)
			if err != nil {
				return err
			}
			var v string
			if json.Unmarshal(e.Value, &v) != nil || v != want {
				return fmt.Errorf("%s accepted publication", id)
			}
			if id == "text-name" && e.Enabled == step.Want.DisableName || id == "text-notes" && e.Enabled == step.Want.DisableEditor {
				return errors.New("text disabled publication")
			}
		}
	}
	return nil
}
func checkTextCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report textReport, err error) {
	report.Event = "vellum.text-native.result"
	report.Result = "FAILED_NATIVE_TEXT"
	defer func() { report.Unavailable = errors.Is(err, cogent.ErrDesktopFrameUnavailable) }()
	var background, retired, stale cogent.DesktopInput
	send := func(label string, before cogent.DesktopSnapshot, in cogent.DesktopInput, reject, keepFrame bool) error {
		if !keepFrame {
			in.Frame = before.Frame
		}
		e := s.Input(ctx, ref, in)
		report.Inputs = append(report.Inputs, catalogueInput{label, before, in, reject})
		if reject {
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return errors.Join(fmt.Errorf("%s expected refusal", label), e)
			}
			return nil
		}
		return e
	}
	click := func(label, target string) error {
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			before, e := catalogueReady(ctx, s, ref)
			if e != nil {
				return e
			}
			el, e := catalogueElement(before, target)
			if e != nil {
				return e
			}
			if el.Bounds.Empty() {
				return errors.New("empty text pointer target")
			}
			in := cogent.DesktopInput{Target: el.ID, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}
			if e = send(label+"/"+typ.String(), before, in, false, false); e != nil {
				return e
			}
			if target == "text-preferences" && typ == events.MouseDown {
				background = in
			}
			if target == "dialog-name" && typ == events.MouseDown {
				retired = in
			}
			if target == "apply" && typ == events.MouseDown {
				stale = in
				stale.Frame = before.Frame
			}
		}
		return nil
	}
	keyInput := func(label, target string, code key.Codes, r rune, mods key.Modifiers) error {
		before, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		el, e := catalogueElement(before, target)
		if e != nil {
			return e
		}
		if !el.Focused {
			return fmt.Errorf("%s current key target lacks focus", label)
		}
		return send(label, before, cogent.DesktopInput{Target: el.ID, Type: events.KeyChord, Code: code, Rune: r, Modifiers: mods}, false, false)
	}
	for _, step := range g.textScript() {
		report.ActiveStep = step.Label
		if step.Setup != nil {
			if err = catalogueOwner(ctx, a, ref, func() error { step.Setup(); return a.Invalidate(ref.ID()) }); err != nil {
				return
			}
		}
		switch step.Kind {
		case "observe":
		case "click":
			err = click(step.Label, step.Target)
		case "key":
			err = keyInput(step.Label, step.Target, step.Code, step.Rune, step.Modifiers)
		case "replace":
			if keymap.Of(key.NewChord('a', key.CodeA, catalogueSelectAllModifiers())) != keymap.SelectAll {
				err = errors.New("native Meta+A is not Select All")
				return
			}
			if err = click(step.Label, step.Target); err != nil {
				return
			}
			if err = keyInput(step.Label+"/select-all", step.Target, key.CodeA, 'a', catalogueSelectAllModifiers()); err != nil {
				return
			}
			for i, r := range step.Text {
				if err = keyInput(fmt.Sprintf("%s/rune-%d", step.Label, i), step.Target, 0, r, 0); err != nil {
					return
				}
			}
			err = keyInput(step.Label+"/commit", step.Target, key.CodeReturnEnter, 0, 0)
		case "refuse", "background", "retired", "stale":
			var before cogent.DesktopSnapshot
			before, err = catalogueReady(ctx, s, ref)
			if err != nil {
				return
			}
			in := background
			if step.Kind == "refuse" {
				var el cogent.DesktopElement
				el, err = catalogueElement(before, step.Target)
				if err != nil {
					return
				}
				in = cogent.DesktopInput{Target: el.ID, Type: events.MouseDown, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}
			}
			if step.Kind == "retired" {
				in = retired
			}
			if step.Kind == "stale" {
				in = stale
			}
			if in.Target == "" {
				err = errors.New("missing text refusal capability")
				return
			}
			err = send(step.Label, before, in, true, step.Kind == "stale")
		default:
			err = errors.New("unknown text step")
		}
		if err != nil {
			return
		}
		var snap cogent.DesktopSnapshot
		snap, err = catalogueReady(ctx, s, ref)
		if err != nil {
			return
		}
		var actual textModel
		if err = catalogueOwner(ctx, a, ref, func() error { actual = g.textModel(); return nil }); err != nil {
			return
		}
		report.Failed = &textObservation{step.Label, actual, snap}
		if actual != step.Want {
			err = fmt.Errorf("%s model %+v want %+v", step.Label, actual, step.Want)
			return
		}
		if err = textObserve(snap, step); err != nil {
			return
		}
		report.Steps = append(report.Steps, *report.Failed)
		report.Failed = nil
		report.ActiveStep = ""
	}
	if len(report.Steps) == 0 {
		err = errors.New("empty text record")
		return
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	target, e := catalogueElement(last, "dialog-name")
	if e != nil {
		err = e
		return
	}
	report.OriginalRefClosed, err = catalogueFinishWindow(ctx, a, s, ref, cogent.DesktopInput{Frame: last.Frame, Target: target.ID, Type: events.KeyChord, Code: key.CodeReturnEnter})
	if err != nil {
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_TEXT"
	return
}
func runTextCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	done := make(chan catalogueOutcome[textReport], 1)
	go func() {
		r, e := checkTextCatalogue(ctx, a, g, s, ref)
		done <- catalogueOutcome[textReport]{r, e}
		cancel()
	}()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	result, joined := catalogueWorkerResult(cancel, done)
	if !joined {
		fmt.Fprintln(os.Stderr, "text catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("text catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_TEXT"
		result.report.Error = result.err.Error()
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}
