package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func catalogueSelectAllModifiers() key.Modifiers {
	var mods key.Modifiers
	mods.SetFlag(true, key.Meta)
	return mods
}

func (g *gallery) acceptNumber(e ui.ChangeEvent) {
	g.numberRequests++
	if !g.numberReject {
		g.number = e.Number
	}
}
func (g *gallery) selectionControl() ui.Node {
	return ui.Node{ID: "selection", Kind: ui.Selector, SelectedID: g.selection, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) {
		g.selectionRequests++
		if !g.selectionReject {
			g.selection = e.SelectedID
		}
	}, Accessibility: ui.Accessibility{Name: "Preset"}}
}
func (g *gallery) numberControl() ui.Node {
	return ui.Node{ID: "number", Kind: ui.NumberField, Number: g.number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: g.acceptNumber, Accessibility: ui.Accessibility{Name: "Numeric level"}}
}

// The bounded fixture reuses the gallery's actual public control constructors.
// It keeps both visible; scrolling already has separate workspace acceptance.
func (g *gallery) controlCatalogueView() ui.Node {
	return ui.Node{ID: "control-catalogue", Kind: ui.Column, Grow: true, Style: ui.Style{Gap: ui.Some(float32(12))}, Children: []ui.Node{
		{ID: "heading", Kind: ui.Label, Text: "Selector and numeric field acceptance"},
		{ID: "fields", Kind: ui.FormGroup, Text: "Ordinary controls", Disabled: !g.enabled, Children: []ui.Node{g.selectionControl(), g.numberControl()}},
		{ID: "outside", Kind: ui.Button, Text: "Outside action", Activate: func() { g.outsideClicks++ }},
	}}
}

type catalogueModel struct {
	Selected                                         string
	Number                                           float32
	SelectionRequests, NumberRequests, OutsideClicks int
	Enabled                                          bool
}
type catalogueObservation struct {
	Label    string
	Model    catalogueModel
	Snapshot cogent.DesktopSnapshot
}
type catalogueInput struct {
	Label    string
	Before   cogent.DesktopSnapshot
	Input    cogent.DesktopInput
	Rejected bool
}
type catalogueReport struct {
	Event, Result, Error string
	Steps                []catalogueObservation
	Inputs               []catalogueInput
	Session              cogent.DesktopState
	OriginalRefClosed    bool
	PostCloseInputError  string
}

func catalogueOwner(ctx context.Context, a *app.Application, ref app.WindowRef, fn func() error) error {
	done := make(chan error, 1)
	if err := a.PostWindow(ref, func() { done <- fn() }); err != nil {
		return err
	}
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
func catalogueReady(ctx context.Context, s *cogent.DesktopSession, ref app.WindowRef) (cogent.DesktopSnapshot, error) {
	for range 4096 {
		err := s.Frame(ctx, ref)
		if err == nil {
			out, e := s.Snapshot(ctx, ref)
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return out, e
			}
		} else if !errors.Is(err, cogent.ErrDesktopPending) && !errors.Is(err, cogent.ErrDesktopBusy) && !(errors.Is(err, cogent.ErrDesktopClosed) && ref.Context().Err() == nil) {
			return cogent.DesktopSnapshot{}, err
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return cogent.DesktopSnapshot{}, ctx.Err()
		}
	}
	return cogent.DesktopSnapshot{}, errors.New("control readiness budget exceeded")
}
func catalogueElement(out cogent.DesktopSnapshot, id string) (cogent.DesktopElement, error) {
	for _, el := range out.Elements {
		if el.DomainID == id {
			return el, nil
		}
	}
	return cogent.DesktopElement{}, fmt.Errorf("control %q absent", id)
}
func catalogueMenu(out cogent.DesktopSnapshot) (cogent.DesktopPopup, error) {
	var found []cogent.DesktopPopup
	for _, p := range out.Popups {
		if p.Kind == core.MenuStage.String() {
			found = append(found, p)
		}
	}
	if len(found) != 1 || !found[0].Interactive || len(found[0].Items) != 2 {
		return cogent.DesktopPopup{}, fmt.Errorf("expected one actual two-item menu: %+v", out.Popups)
	}
	if found[0].Items[0].Name != "Basic" || found[0].Items[1].Name != "Advanced" {
		return cogent.DesktopPopup{}, errors.New("disabled/incorrect menu options")
	}
	return found[0], nil
}

func checkControlCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report catalogueReport, err error) {
	report.Event = "vellum.controls-native.result"
	report.Result = "FAILED_NATIVE_CONTROLS"
	defer func() {
		if err != nil {
			report.Error = err.Error()
		}
	}()
	want := catalogueModel{Selected: "basic", Number: 25, Enabled: true}
	setup := func(fn func()) error {
		return catalogueOwner(ctx, a, ref, func() error { fn(); return a.Invalidate(ref.ID()) })
	}
	observe := func(label string, menus int) error {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		var actual catalogueModel
		e = catalogueOwner(ctx, a, ref, func() error {
			actual = catalogueModel{g.selection, g.number, g.selectionRequests, g.numberRequests, g.outsideClicks, g.enabled}
			return nil
		})
		if e != nil {
			return e
		}
		if actual != want {
			return fmt.Errorf("%s model %+v want %+v", label, actual, want)
		}
		count := 0
		for _, p := range snap.Popups {
			if p.Kind == core.MenuStage.String() {
				count++
			}
		}
		if count != menus {
			return fmt.Errorf("%s menu count %d want %d", label, count, menus)
		}
		selector, e := catalogueElement(snap, "selection")
		if e != nil {
			return e
		}
		number, e := catalogueElement(snap, "number")
		if e != nil {
			return e
		}
		if selector.Enabled != want.Enabled || number.Enabled != want.Enabled || selector.Bounds.Empty() || number.Bounds.Empty() {
			return fmt.Errorf("%s enabled/visible control publication", label)
		}
		var selectedLabel string
		var numeric float32
		if json.Unmarshal(selector.Value, &selectedLabel) != nil || json.Unmarshal(number.Value, &numeric) != nil || numeric != want.Number || selectedLabel != map[string]string{"basic": "Basic", "advanced": "Advanced"}[want.Selected] {
			return fmt.Errorf("%s controlled semantics mismatch", label)
		}
		if menus == 1 {
			p, e := catalogueMenu(snap)
			if e != nil {
				return e
			}
			for _, item := range p.Items {
				if item.Selected != (item.Name == selectedLabel) || !item.Enabled {
					return fmt.Errorf("%s actual chooser projection mismatch", label)
				}
			}
		}
		report.Steps = append(report.Steps, catalogueObservation{label, actual, snap})
		return nil
	}
	send := func(label string, snap cogent.DesktopSnapshot, in cogent.DesktopInput, reject bool) error {
		in.Frame = snap.Frame
		e := s.Input(ctx, ref, in)
		report.Inputs = append(report.Inputs, catalogueInput{label, snap, in, reject})
		if reject {
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return fmt.Errorf("%s expected disabled/stale refusal: %v", label, e)
			}
			return nil
		}
		if e != nil {
			return fmt.Errorf("%s: %w", label, e)
		}
		return nil
	}
	click := func(label, id, menuItem string, outside bool) error {
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			snap, e := catalogueReady(ctx, s, ref)
			if e != nil {
				return e
			}
			var el cogent.DesktopElement
			if menuItem != "" {
				p, e := catalogueMenu(snap)
				if e != nil {
					return e
				}
				for _, item := range p.Items {
					if item.Name == menuItem {
						el = item
					}
				}
			} else {
				el, e = catalogueElement(snap, id)
				if e != nil {
					return e
				}
			}
			if el.ID == "" || el.Bounds.Empty() {
				return fmt.Errorf("%s actual pointer target missing", label)
			}
			pos := el.Bounds.Min.Add(el.Bounds.Size().Div(2))
			target := el.ID
			if outside {
				p, e := catalogueMenu(snap)
				if e != nil {
					return e
				}
				if pos.In(p.Bounds) {
					return errors.New("outside control overlaps menu")
				}
				target = ""
			}
			if e = send(label+"/"+typ.String(), snap, cogent.DesktopInput{Target: target, Type: typ, Button: events.Left, Position: pos}, false); e != nil {
				return e
			}
		}
		return nil
	}
	chord := func(label, id string, code key.Codes, r rune, mods key.Modifiers, menu bool) error {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		var el cogent.DesktopElement
		if menu {
			p, e := catalogueMenu(snap)
			if e != nil {
				return e
			}
			for _, item := range p.Items {
				if item.Focused {
					if el.ID != "" {
						return errors.New("multiple focused menu items")
					}
					el = item
				}
			}
		} else {
			el, e = catalogueElement(snap, id)
			if e != nil {
				return e
			}
		}
		if el.ID == "" || !el.Focused {
			return fmt.Errorf("%s target not actually focused", label)
		}
		return send(label, snap, cogent.DesktopInput{Target: el.ID, Type: events.KeyChord, Code: code, Rune: r, Modifiers: mods}, false)
	}
	typeNumber := func(label, text string) error {
		if keymap.Of(key.NewChord('a', key.CodeA, catalogueSelectAllModifiers())) != keymap.SelectAll {
			return errors.New("native Meta+A is not Select All")
		}
		if e := chord(label+"/select-all", "number", key.CodeA, 'a', catalogueSelectAllModifiers(), false); e != nil {
			return e
		}
		for i, r := range text {
			if e := chord(fmt.Sprintf("%s/rune-%d", label, i), "number", key.CodeUnknown, r, 0, false); e != nil {
				return e
			}
		}
		if e := chord(label+"/enter", "number", key.CodeReturnEnter, 0, 0, false); e != nil {
			return e
		}
		// Enter intentionally advances focus. Route it back through the same
		// public pointer path before the next authored numeric key operation.
		return click(label+"/refocus", "number", "", false)
	}
	if err = observe("initial", 0); err != nil {
		return
	}
	if err = click("selector-open", "selection", "", false); err != nil {
		return
	}
	if err = observe("pointer-menu-open", 1); err != nil {
		return
	}
	if err = click("choose-advanced", "", "Advanced", false); err != nil {
		return
	}
	want.Selected = "advanced"
	want.SelectionRequests = 1
	if err = observe("pointer-selection-once", 0); err != nil {
		return
	}
	if err = setup(func() { g.selectionReject = true }); err != nil {
		return
	}
	if err = chord("keyboard-open", "selection", key.CodeReturnEnter, 0, 0, false); err != nil {
		return
	}
	if err = observe("keyboard-menu-open", 1); err != nil {
		return
	}
	if err = chord("menu-up", "", key.CodeUpArrow, 0, 0, true); err != nil {
		return
	}
	if err = chord("menu-reject-enter", "", key.CodeReturnEnter, 0, 0, true); err != nil {
		return
	}
	want.SelectionRequests = 2
	if err = observe("rejected-selection-restored", 0); err != nil {
		return
	}
	if err = chord("reopen-rejected", "selection", key.CodeReturnEnter, 0, 0, false); err != nil {
		return
	}
	if err = observe("actual-rejected-choice", 1); err != nil {
		return
	}
	if err = chord("escape-menu", "", key.CodeEscape, 0, 0, true); err != nil {
		return
	}
	if err = observe("escape-cancels", 0); err != nil {
		return
	}
	if err = setup(func() { g.selectionReject = false }); err != nil {
		return
	}
	if err = chord("open-outside-cancel", "selection", key.CodeReturnEnter, 0, 0, false); err != nil {
		return
	}
	if err = click("outside-cancel", "outside", "", true); err != nil {
		return
	}
	if err = observe("outside-release-consumed", 0); err != nil {
		return
	}
	if err = click("outside-normal", "outside", "", false); err != nil {
		return
	}
	want.OutsideClicks = 1
	if err = observe("outside-action-normal", 0); err != nil {
		return
	}
	if err = chord("tab-to-selector", "outside", key.CodeTab, 0, 0, false); err != nil {
		return
	}
	if err = chord("open-accept-keyboard", "selection", key.CodeReturnEnter, 0, 0, false); err != nil {
		return
	}
	if err = chord("accept-menu-up", "", key.CodeUpArrow, 0, 0, true); err != nil {
		return
	}
	if err = chord("accept-menu-enter", "", key.CodeReturnEnter, 0, 0, true); err != nil {
		return
	}
	want.Selected = "basic"
	want.SelectionRequests = 3
	if err = observe("keyboard-selection-once", 0); err != nil {
		return
	}
	if err = click("number-focus", "number", "", false); err != nil {
		return
	}
	if err = typeNumber("number-42", "42"); err != nil {
		return
	}
	want.Number = 42
	want.NumberRequests = 1
	if err = observe("typed-number-once", 0); err != nil {
		return
	}
	if err = setup(func() { g.numberReject = true }); err != nil {
		return
	}
	if err = chord("rejected-step", "number", key.CodeUpArrow, 0, 0, false); err != nil {
		return
	}
	want.NumberRequests = 2
	if err = observe("rejected-number-restored", 0); err != nil {
		return
	}
	if err = setup(func() { g.numberReject = false }); err != nil {
		return
	}
	if err = chord("accepted-step", "number", key.CodeUpArrow, 0, 0, false); err != nil {
		return
	}
	want.Number = 43
	want.NumberRequests = 3
	if err = observe("step-after-rejection", 0); err != nil {
		return
	}
	if err = typeNumber("invalid-number", "bad"); err != nil {
		return
	}
	if err = observe("invalid-number-no-proposal", 0); err != nil {
		return
	}
	if err = typeNumber("corrected-number", "55"); err != nil {
		return
	}
	want.Number = 55
	want.NumberRequests = 4
	if err = observe("number-correction", 0); err != nil {
		return
	}
	if err = typeNumber("clamped-number", "101"); err != nil {
		return
	}
	want.Number = 100
	want.NumberRequests = 5
	if err = observe("number-clamped", 0); err != nil {
		return
	}
	if err = chord("step-down", "number", key.CodeDownArrow, 0, 0, false); err != nil {
		return
	}
	want.Number = 99
	want.NumberRequests = 6
	if err = observe("step-from-clamped", 0); err != nil {
		return
	}
	if err = setup(func() { g.enabled = false }); err != nil {
		return
	}
	want.Enabled = false
	if err = observe("disabled-publication", 0); err != nil {
		return
	}
	for _, id := range []string{"number", "selection"} {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			err = e
			return
		}
		el, e := catalogueElement(snap, id)
		if e != nil {
			err = e
			return
		}
		if err = send("disabled-target-"+id, snap, cogent.DesktopInput{Target: el.ID, Type: events.KeyChord, Code: key.CodeUpArrow}, true); err != nil {
			return
		}
	}
	snap, e := catalogueReady(ctx, s, ref)
	if e != nil {
		err = e
		return
	}
	if err = send("disabled-ordinary-key", snap, cogent.DesktopInput{Type: events.KeyChord, Code: key.CodeUpArrow}, false); err != nil {
		return
	}
	if err = observe("disabled-no-proposals", 0); err != nil {
		return
	}
	if err = setup(func() { g.enabled = true }); err != nil {
		return
	}
	want.Enabled = true
	if err = click("reenabled-number-focus", "number", "", false); err != nil {
		return
	}
	if err = chord("reenabled-step", "number", key.CodeUpArrow, 0, 0, false); err != nil {
		return
	}
	want.Number = 100
	want.NumberRequests = 7
	if err = observe("reenabled-number", 0); err != nil {
		return
	}
	if err = click("last-menu-open", "selection", "", false); err != nil {
		return
	}
	if err = observe("menu-before-close", 1); err != nil {
		return
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	p, e := catalogueMenu(last)
	if e != nil {
		err = e
		return
	}
	if err = catalogueOwner(ctx, a, ref, func() error { return a.CloseWindow(ref) }); err != nil {
		return
	}
	report.OriginalRefClosed = ref.Context().Err() != nil
	e = s.Input(ctx, ref, cogent.DesktopInput{Frame: last.Frame, Target: p.Items[0].ID, Type: events.MouseDown, Button: events.Left, Position: p.Items[0].Bounds.Min})
	if !report.OriginalRefClosed || !errors.Is(e, cogent.ErrDesktopClosed) {
		err = fmt.Errorf("closed original window admitted popup input: %v", e)
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_CONTROLS"
	return
}

func runControlCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	type outcome struct {
		report catalogueReport
		err    error
	}
	done := make(chan outcome, 1)
	go func() { r, e := checkControlCatalogue(ctx, a, g, s, ref); done <- outcome{r, e}; cancel() }()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	var result outcome
	select {
	case result = <-done:
	case <-time.After(6 * time.Second):
		// Do not take application/session locks when a worker failed to join.
		fmt.Fprintln(os.Stderr, "control catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("control catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_CONTROLS"
		result.report.Error = result.err.Error()
	}
	if err := json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}
