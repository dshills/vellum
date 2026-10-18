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
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func (g *gallery) navigationCatalogueView() ui.Node {
	return ui.Node{ID: "navigation-catalogue", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "layers", Sequence: 1}, Style: ui.Style{Gap: ui.Some(float32(12))}, Children: []ui.Node{
		{ID: "navigation-title", Kind: ui.Label, Text: "Layers and native action menus"}, g.layerListControl(),
		{ID: "navigation-tools", Kind: ui.Toolbar, Children: []ui.Node{g.actionMenuControl(), g.viewMenuControl()}},
		{ID: "navigation-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Layer %s / action %s / grid %t", g.listSelection, g.menuLast, g.menuChecked)},
		{ID: "outside", Kind: ui.Button, Text: "Acknowledge", Activate: func() { g.outsideClicks++ }},
	}}
}

type navigationModel struct {
	List, Action                                                   string
	ListRequests, ActionRequests, CheckRequests, OutsideActions    int
	ReadOnly, Reject, Disabled, MenuDisabled, Checked, CheckReject bool
}

func (g *gallery) navigationModel() navigationModel {
	return navigationModel{g.listSelection, g.menuLast, g.listRequests, g.menuCount, g.viewMenuRequests, g.outsideClicks, g.listReadOnly, g.listReject, g.listDisabled, g.menuDisabled, g.menuChecked, g.viewMenuReject}
}

type navigationStep struct {
	Label, Kind, Target, Option string
	Code                        key.Codes
	Rune                        rune
	Popup                       string
	Focus                       string
	Want                        navigationModel
	Setup                       func()
}

// This is a bounded authored consumer script, not a framework input interpreter.
// Only policy flags use Setup. Tested values/counters change via public callbacks.
func (g *gallery) navigationScript() []navigationStep {
	want := navigationModel{List: "layer-00"}
	steps := []navigationStep{}
	add := func(label, kind, target, option string, code key.Codes, r rune, popup, focus string, setup func()) {
		steps = append(steps, navigationStep{label, kind, target, option, code, r, popup, focus, want, setup})
	}
	add("initial", "observe", "", "", 0, 0, "", "layer-00", nil)
	want.List = "layer-01"
	want.ListRequests = 1
	add("list-down", "key", "layers", "", key.CodeDownArrow, 0, "", "layer-01", nil)
	want.List = "layer-03"
	want.ListRequests = 2
	add("list-disabled-skip", "key", "layers", "", key.CodeDownArrow, 0, "", "layer-03", nil)
	want.List = "layer-03"
	want.ListRequests = 3
	add("list-pointer", "click", "layers", "layer-03", 0, 0, "", "layer-03", nil)
	want.List = "layer-23"
	want.ListRequests = 4
	add("list-end", "key", "layers", "", key.CodeEnd, 0, "", "layer-23", nil)
	add("list-boundary", "key", "layers", "", key.CodeDownArrow, 0, "", "layer-23", nil)
	want.List = "layer-00"
	want.ListRequests = 5
	add("list-home", "key", "layers", "", key.CodeHome, 0, "", "layer-00", nil)
	want.List = "layer-23"
	want.ListRequests = 6
	add("list-initial-letter", "key", "layers", "", key.CodeA, 'a', "", "layer-23", nil)
	want.Reject = true
	want.ListRequests = 7
	add("list-rejected", "key", "layers", "", key.CodeHome, 0, "", "layer-00", func() { g.listReject = true })
	want.Reject = false
	want.ReadOnly = true
	add("list-readonly-end", "key", "layers", "", key.CodeEnd, 0, "", "layer-23", func() { g.listReject = false; g.listReadOnly = true })
	add("list-readonly-home", "key", "layers", "", key.CodeHome, 0, "", "layer-00", nil)
	add("list-readonly-activate", "key", "layers", "", key.CodeSpacebar, ' ', "", "layer-00", nil)
	want.ReadOnly = false
	want.Disabled = true
	add("list-disabled", "observe", "", "", 0, 0, "", "", func() { g.listReadOnly = false; g.listDisabled = true })
	add("list-disabled-refusal", "disabled", "layers", "", key.CodeDownArrow, 0, "", "", nil)
	want.Disabled = false
	want.List = "layer-01"
	want.ListRequests = 8
	add("list-reenabled-pointer", "click", "layers", "layer-01", 0, 0, "", "layer-01", func() { g.listDisabled = false })
	add("list-tab-to-menu", "key", "layers", "", key.CodeTab, 0, "", "menu-actions", nil)
	add("action-menu-open", "key", "menu-actions", "", key.CodeDownArrow, 0, "menu-actions", "first", nil)
	add("action-disabled-refusal", "popup-disabled", "", "", 0, 0, "menu-actions", "first", nil)
	want.Action = "second"
	want.ActionRequests = 1
	add("action-second-pointer", "popup-click", "", "", 0, 0, "", "menu-actions", nil)
	add("action-stale-refusal", "stale", "", "", 0, 0, "", "menu-actions", nil)
	add("action-keyboard-open", "key", "menu-actions", "", key.CodeReturnEnter, 0, "menu-actions", "first", nil)
	add("action-keyboard-next", "popup-key", "", "", key.CodeDownArrow, 0, "menu-actions", "second", nil)
	want.ActionRequests = 2
	add("action-second-keyboard", "popup-key", "", "", key.CodeReturnEnter, 0, "", "menu-actions", nil)
	add("action-escape-open", "key", "menu-actions", "", key.CodeReturnEnter, 0, "menu-actions", "first", nil)
	add("action-escape", "popup-key", "", "", key.CodeEscape, 0, "", "menu-actions", nil)
	add("action-outside-open", "key", "menu-actions", "", key.CodeReturnEnter, 0, "menu-actions", "first", nil)
	add("action-outside-consumed", "outside", "", "", 0, 0, "", "", nil)
	want.OutsideActions = 1
	add("outside-normal", "click", "outside", "", 0, 0, "", "outside", nil)
	add("action-before-disable", "click", "menu-actions", "", 0, 0, "menu-actions", "first", nil)
	want.MenuDisabled = true
	add("action-disabled-close", "observe", "", "", 0, 0, "", "", func() { g.menuDisabled = true })
	add("action-anchor-refused", "disabled", "menu-actions", "", key.CodeDownArrow, 0, "", "", nil)
	want.MenuDisabled = false
	add("action-reenabled-open", "click", "menu-actions", "", 0, 0, "menu-actions", "first", func() { g.menuDisabled = false })
	add("action-tab", "popup-key", "", "", key.CodeTab, 0, "", "view-menu", nil)
	add("check-menu-open", "key", "view-menu", "", key.CodeDownArrow, 0, "view-menu", "grid", nil)
	want.Checked = true
	want.CheckRequests = 1
	add("check-accepted", "popup-click", "", "", 0, 0, "", "view-menu", nil)
	add("check-menu-reopen", "key", "view-menu", "", key.CodeReturnEnter, 0, "view-menu", "grid", nil)
	want.CheckReject = true
	want.CheckRequests = 2
	add("check-rejected", "popup-key", "", "", key.CodeReturnEnter, 0, "", "view-menu", func() { g.viewMenuReject = true })
	want.CheckReject = false
	add("check-reset-open", "key", "view-menu", "", key.CodeDownArrow, 0, "view-menu", "grid", func() { g.viewMenuReject = false })
	add("check-reset-next", "popup-key", "", "", key.CodeDownArrow, 0, "view-menu", "reset", nil)
	want.Checked = false
	want.CheckRequests = 3
	add("check-reset", "popup-key", "", "", key.CodeReturnEnter, 0, "", "view-menu", nil)
	add("menu-before-close", "click", "menu-actions", "", 0, 0, "menu-actions", "first", nil)
	return steps
}

type navigationObservation struct {
	Label    string
	Model    navigationModel
	Snapshot cogent.DesktopSnapshot
}
type navigationReport struct {
	Event, Result, Error string
	Failed               *navigationObservation
	Steps                []navigationObservation
	Inputs               []catalogueInput
	Session              cogent.DesktopState
	OriginalRefClosed    bool
	PostCloseInputError  string
}

func navigationPopup(s cogent.DesktopSnapshot) (cogent.DesktopPopup, error) {
	if len(s.Popups) != 1 || s.Popups[0].Kind != "MenuStage" || !s.Popups[0].Interactive {
		return cogent.DesktopPopup{}, errors.New("one interactive actual menu required")
	}
	return s.Popups[0], nil
}
func navigationObserve(s cogent.DesktopSnapshot, w navigationStep) error {
	if w.Popup == "" {
		if len(s.Popups) != 0 {
			return errors.New("unexpected popup")
		}
		for _, id := range []string{"layers", "menu-actions", "view-menu", "outside", "navigation-title", "navigation-status"} {
			el, e := catalogueElement(s, id)
			if e != nil {
				return e
			}
			if el.Bounds.Empty() || !el.Bounds.In(image.Rectangle{Max: s.Size}) {
				return fmt.Errorf("%s not fully visible", id)
			}
			enabled := !(id == "layers" && w.Want.Disabled || id == "menu-actions" && w.Want.MenuDisabled)
			if el.Enabled != enabled {
				return fmt.Errorf("%s enabled publication", id)
			}
		}
		list, _ := catalogueElement(s, "layers")
		var label string
		options := galleryLayerOptions(false)
		for _, o := range options {
			row, e := choiceElement(s, "layers", o.ID)
			if e != nil {
				return e
			}
			if row.Selected != (o.ID == w.Want.List) || row.Enabled != (!w.Want.Disabled && !o.Disabled) {
				return fmt.Errorf("row %s accepted/enabled state", o.ID)
			}
			if o.ID == w.Want.List {
				label = o.Label
			}
		}
		var actual string
		if json.Unmarshal(list.Value, &actual) != nil || actual != label {
			return errors.New("list accepted semantic value")
		}
		status, _ := catalogueElement(s, "navigation-status")
		if json.Unmarshal(status.Value, &actual) != nil || actual != fmt.Sprintf("Layer %s / action %s / grid %t", w.Want.List, w.Want.Action, w.Want.Checked) {
			return errors.New("controlled navigation status")
		}
		if w.Label == "list-end" || w.Label == "list-boundary" || w.Label == "list-initial-letter" || w.Label == "list-readonly-end" {
			first, e := choiceElement(s, "layers", "layer-00")
			if e != nil {
				return e
			}
			last, e := choiceElement(s, "layers", "layer-23")
			if e != nil {
				return e
			}
			if !first.Bounds.Empty() || last.Bounds.Empty() {
				return errors.New("End did not scroll actual list")
			}
		}
	} else {
		p, e := navigationPopup(s)
		if e != nil {
			return e
		}
		ids, names := []string{"blocked", "first", "second"}, []string{"Unavailable", "Run action", "Run action"}
		if w.Popup == "view-menu" {
			ids, names = []string{"grid", "reset"}, []string{"Show grid", "Reset view"}
		}
		if len(p.Items) != len(ids) {
			return errors.New("actual menu item/separator count")
		}
		for i, id := range ids {
			item := p.Items[i]
			if item.Name != names[i] || item.Enabled != (id != "blocked") || item.Bounds.Empty() || !item.Bounds.In(p.Bounds) {
				return errors.New("actual menu option geometry/state")
			}
			semantic, e := catalogueElement(s, id)
			if e != nil {
				return e
			}
			if semantic.Role != "menu-item" || semantic.Enabled != item.Enabled {
				return errors.New("menu semantic option")
			}
			if id == "grid" {
				var checked bool
				if json.Unmarshal(semantic.Value, &checked) != nil || checked != w.Want.Checked {
					return errors.New("accepted check action value")
				}
			}
		}
	}
	if w.Focus != "" {
		el, e := catalogueElement(s, w.Focus)
		if e != nil {
			return e
		}
		if !el.Focused || el.Bounds.Empty() {
			return fmt.Errorf("actual focus %s absent", w.Focus)
		}
	}
	return nil
}
func checkNavigationCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report navigationReport, err error) {
	report.Event = "vellum.navigation-native.result"
	report.Result = "FAILED_NATIVE_NAVIGATION"
	defer func() {
		if err != nil {
			report.Error = err.Error()
		}
	}()
	var stale cogent.DesktopInput
	var outside image.Point
	send := func(label string, before cogent.DesktopSnapshot, in cogent.DesktopInput, reject bool) error {
		in.Frame = before.Frame
		e := s.Input(ctx, ref, in)
		report.Inputs = append(report.Inputs, catalogueInput{label, before, in, reject})
		if reject {
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return fmt.Errorf("%s expected refusal: %v", label, e)
			}
			return nil
		}
		return e
	}
	for _, step := range g.navigationScript() {
		if step.Setup != nil {
			if err = catalogueOwner(ctx, a, ref, func() error { step.Setup(); return a.Invalidate(ref.ID()) }); err != nil {
				return
			}
		}
		if step.Kind != "observe" {
			types := []events.Types{events.KeyChord}
			if step.Kind == "click" || step.Kind == "popup-click" || step.Kind == "outside" {
				types = []events.Types{events.MouseDown, events.MouseUp}
			}
			for _, typ := range types {
				before, e := catalogueReady(ctx, s, ref)
				if e != nil {
					err = e
					return
				}
				in := cogent.DesktopInput{Type: typ, Code: step.Code, Rune: step.Rune}
				reject := step.Kind == "disabled" || step.Kind == "popup-disabled" || step.Kind == "stale"
				switch step.Kind {
				case "key", "click", "disabled":
					el, e := choiceElement(before, step.Target, step.Option)
					if e != nil {
						err = e
						return
					}
					in.Target = el.ID
					if typ == events.KeyChord && step.Kind != "disabled" && !el.Focused {
						err = fmt.Errorf("%s key target lacks actual focus", step.Label)
						return
					}
					if typ != events.KeyChord {
						in.Button = events.Left
						in.Position = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
					}
				case "popup-click", "popup-key", "popup-disabled":
					p, e := navigationPopup(before)
					if e != nil {
						err = e
						return
					}
					index := -1
					if step.Kind == "popup-click" {
						index = len(p.Items) - 1
						if step.Label == "check-accepted" {
							index = 0
						}
					} else if step.Kind == "popup-disabled" {
						index = 0
					} else {
						for i, item := range p.Items {
							if item.Focused {
								if index != -1 {
									err = errors.New("multiple actual popup focus")
									return
								}
								index = i
							}
						}
					}
					if index < 0 {
						err = errors.New("actual focused menu item absent")
						return
					}
					el := p.Items[index]
					in.Target = el.ID
					if step.Kind != "popup-key" {
						in.Type = events.MouseDown
						if step.Kind == "popup-click" {
							in.Type = typ
						}
						in.Button = events.Left
						in.Position = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
						in.Code = key.CodeUnknown
					}
					if step.Kind == "popup-click" {
						stale = in
					}
				case "outside":
					p, e := navigationPopup(before)
					if typ == events.MouseDown && (e != nil || outside.In(p.Bounds)) {
						err = errors.New("outside dismissal lacks actual outside point")
						return
					}
					in.Button = events.Left
					in.Position = outside
				case "stale":
					in = stale
					in.Type = events.MouseDown
				default:
					err = errors.New("unknown navigation script action")
					return
				}
				label := step.Label
				if len(types) == 2 {
					label += "/" + typ.String()
				}
				if err = send(label, before, in, reject); err != nil {
					err = fmt.Errorf("%s: %w", step.Label, err)
					return
				}
			}
		}
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			err = e
			return
		}
		var actual navigationModel
		if err = catalogueOwner(ctx, a, ref, func() error { actual = g.navigationModel(); return nil }); err != nil {
			return
		}
		report.Failed = &navigationObservation{step.Label, actual, snap}
		if actual != step.Want {
			err = fmt.Errorf("%s model %+v want %+v", step.Label, actual, step.Want)
			return
		}
		if err = navigationObserve(snap, step); err != nil {
			err = fmt.Errorf("%s: %w", step.Label, err)
			return
		}
		if step.Popup == "" {
			el, e := catalogueElement(snap, "outside")
			if e != nil {
				err = e
				return
			}
			outside = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
		}
		report.Steps = append(report.Steps, navigationObservation{step.Label, actual, snap})
		report.Failed = nil
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	finalPopup, e := navigationPopup(last)
	if e != nil {
		err = e
		return
	}
	finalTarget := finalPopup.Items[1].ID
	report.OriginalRefClosed, err = catalogueFinishWindow(ctx, a, s, ref, cogent.DesktopInput{Frame: last.Frame, Target: finalTarget, Type: events.KeyChord, Code: key.CodeReturnEnter})
	if err != nil {
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_NAVIGATION"
	return
}
func runNavigationCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	done := make(chan catalogueOutcome[navigationReport], 1)
	go func() {
		r, e := checkNavigationCatalogue(ctx, a, g, s, ref)
		done <- catalogueOutcome[navigationReport]{r, e}
		cancel()
	}()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	result, joined := catalogueWorkerResult(cancel, done)
	if !joined {
		fmt.Fprintln(os.Stderr, "navigation catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("navigation catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_NAVIGATION"
		result.report.Error = result.err.Error()
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}
