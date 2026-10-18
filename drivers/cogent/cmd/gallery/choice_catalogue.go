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

func (g *gallery) enableControl() ui.Node {
	return ui.Node{ID: "enable", Kind: ui.Checkbox, Text: "Enable editing controls", Checked: g.enabled, Change: func(e ui.ChangeEvent) { g.enabledRequests++; g.enabled = e.Checked }, Accessibility: ui.Accessibility{Help: "Toggle the disabled state of the controls below"}}
}
func (g *gallery) snappingControl() ui.Node {
	return ui.Node{ID: "snapping", Kind: ui.Switch, Text: "Snapping", Checked: g.snapping, Change: func(e ui.ChangeEvent) { g.snappingRequests++; g.snapping = e.Checked }}
}
func (g *gallery) pinnedControl() ui.Node {
	return ui.Node{ID: "pinned", Kind: ui.ToggleButton, Icon: ui.SymbolCheck, Text: "Pin panel", Checked: g.pinned, Change: func(e ui.ChangeEvent) { g.pinnedRequests++; g.pinned = e.Checked }}
}
func (g *gallery) choiceControl(id string, kind ui.Kind, name string) ui.Node {
	return ui.Node{ID: id, Kind: kind, SelectedID: g.selection, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) {
		g.selectionRequests++
		if !g.selectionReject {
			g.selection = e.SelectedID
		}
	}, Accessibility: ui.Accessibility{Name: name}}
}
func (g *gallery) radioControl() ui.Node {
	return g.choiceControl("radio-preset", ui.RadioGroup, "Preset radio choices")
}
func (g *gallery) segmentedControl() ui.Node {
	return g.choiceControl("segmented-preset", ui.SegmentedControl, "Preset segments")
}
func (g *gallery) sliderControl() ui.Node {
	return ui.Node{ID: "slider", Kind: ui.Slider, Number: g.number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: g.acceptNumber, MinWidth: 300, Accessibility: ui.Accessibility{Name: "Level", Help: "Drag or use arrow keys"}}
}
func (g *gallery) progressBarControl() ui.Node {
	return ui.Node{ID: "progress", Kind: ui.ProgressBar, Number: g.number / 100, Accessibility: ui.Accessibility{Name: "Operation progress"}}
}
func (g *gallery) progressShapesControl() ui.Node {
	return ui.Node{ID: "progress-shapes", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(24))}, Children: []ui.Node{
		{ID: "progress-ring", Kind: ui.ProgressRing, Number: g.number / 100, Text: fmt.Sprintf("%.0f%%", g.number), Accessibility: ui.Accessibility{Name: "Circular operation progress"}},
		{ID: "progress-gauge", Kind: ui.ProgressGauge, Number: g.number / 100, Text: fmt.Sprintf("%.0f%%", g.number), Accessibility: ui.Accessibility{Name: "Semicircular operation progress"}},
	}}
}

func (g *gallery) choiceCatalogueView() ui.Node {
	return ui.Node{ID: "choice-catalogue", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "enable", Sequence: 1}, Style: ui.Style{Gap: ui.Some(float32(8))}, Children: []ui.Node{
		{ID: "heading", Kind: ui.Row, Children: []ui.Node{{ID: "choice-icon", Kind: ui.Icon, Icon: ui.SymbolCheck, Text: "Choice check"}, {ID: "choice-title", Kind: ui.Label, Text: "Standard choices, toggles and progress"}}},
		g.enableControl(),
		{ID: "more-controls", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Control tools"}, Children: []ui.Node{g.snappingControl(), g.pinnedControl()}},
		{ID: "fields", Kind: ui.FormGroup, Text: "Ordinary choices", Disabled: !g.enabled, Children: []ui.Node{{ID: "choice-columns", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(24))}, Children: []ui.Node{g.radioControl(), {ID: "numeric-column", Kind: ui.Column, Grow: true, Children: []ui.Node{g.segmentedControl(), g.sliderControl()}}}}}},
		{ID: "controls-separator", Kind: ui.Separator}, g.progressBarControl(), g.progressShapesControl(),
		{ID: "choice-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Preset %s / level %.0f", g.selection, g.number)},
		{ID: "outside", Kind: ui.Button, Text: "Acknowledge", Activate: func() { g.outsideClicks++ }},
	}}
}

type choiceModel struct {
	Enabled, Snapping, Pinned                                                                          bool
	Selected                                                                                           string
	Number                                                                                             float32
	EnableRequests, SnappingRequests, PinnedRequests, SelectionRequests, NumberRequests, OutsideClicks int
}
type choiceObservation struct {
	Label    string
	Model    choiceModel
	Snapshot cogent.DesktopSnapshot
}
type choiceReport struct {
	Event, Result, Error string
	Steps                []choiceObservation
	Inputs               []catalogueInput
	Session              cogent.DesktopState
	OriginalRefClosed    bool
	PostCloseInputError  string
}

func choiceElement(s cogent.DesktopSnapshot, id, option string) (cogent.DesktopElement, error) {
	group, err := catalogueElement(s, id)
	if err != nil || option == "" {
		return group, err
	}
	for _, el := range s.Elements {
		if el.Parent == group.ID && el.DomainID == option {
			return el, nil
		}
	}
	return cogent.DesktopElement{}, fmt.Errorf("choice %s/%s absent", id, option)
}
func checkChoiceCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report choiceReport, err error) {
	report.Event = "vellum.choices-native.result"
	report.Result = "FAILED_NATIVE_CHOICES"
	defer func() {
		if err != nil {
			report.Error = err.Error()
		}
	}()
	want := choiceModel{Enabled: true, Selected: "basic", Number: 25}
	setup := func(fn func()) error {
		return catalogueOwner(ctx, a, ref, func() error { fn(); return a.Invalidate(ref.ID()) })
	}
	observe := func(label string) error {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		var actual choiceModel
		e = catalogueOwner(ctx, a, ref, func() error {
			actual = choiceModel{g.enabled, g.snapping, g.pinned, g.selection, g.number, g.enabledRequests, g.snappingRequests, g.pinnedRequests, g.selectionRequests, g.numberRequests, g.outsideClicks}
			return nil
		})
		if e != nil {
			return e
		}
		if actual != want {
			return fmt.Errorf("%s model %+v want %+v", label, actual, want)
		}
		if len(snap.Popups) != 0 {
			return fmt.Errorf("%s unexpected popup", label)
		}
		for _, id := range []string{"enable", "snapping", "pinned", "radio-preset", "segmented-preset", "slider", "progress", "progress-ring", "progress-gauge", "choice-status", "outside", "choice-icon", "choice-title"} {
			el, e := choiceElement(snap, id, "")
			if e != nil {
				return e
			}
			if el.Bounds.Empty() || !el.Bounds.In(image.Rectangle{Max: snap.Size}) {
				return fmt.Errorf("%s %s not fully visible", label, id)
			}
			enabled := true
			if id == "radio-preset" || id == "segmented-preset" || id == "slider" {
				enabled = want.Enabled
			}
			if el.Enabled != enabled {
				return fmt.Errorf("%s %s enabled publication", label, id)
			}
		}
		for id, value := range map[string]bool{"enable": want.Enabled, "snapping": want.Snapping, "pinned": want.Pinned} {
			el, _ := choiceElement(snap, id, "")
			var v bool
			if json.Unmarshal(el.Value, &v) != nil || v != value {
				return fmt.Errorf("%s boolean %s", label, id)
			}
		}
		for _, id := range []string{"radio-preset", "segmented-preset"} {
			el, _ := choiceElement(snap, id, "")
			var value string
			if json.Unmarshal(el.Value, &value) != nil || value != map[string]string{"basic": "Basic", "advanced": "Advanced"}[want.Selected] {
				return fmt.Errorf("%s selection %s", label, id)
			}
			for _, opt := range []string{"basic", "advanced", "unavailable"} {
				item, e := choiceElement(snap, id, opt)
				if e != nil {
					return e
				}
				if item.Selected != (opt == want.Selected) || item.Enabled != (want.Enabled && opt != "unavailable") || item.Bounds.Empty() || !item.Bounds.In(image.Rectangle{Max: snap.Size}) {
					return fmt.Errorf("%s option %s/%s", label, id, opt)
				}
			}
		}
		for _, id := range []string{"slider", "progress", "progress-ring", "progress-gauge"} {
			el, _ := choiceElement(snap, id, "")
			var value float32
			expected := want.Number
			if id != "slider" {
				expected /= 100
			}
			if json.Unmarshal(el.Value, &value) != nil || value != expected {
				return fmt.Errorf("%s value %s", label, id)
			}
		}
		status, _ := choiceElement(snap, "choice-status", "")
		var caption string
		if json.Unmarshal(status.Value, &caption) != nil || caption != fmt.Sprintf("Preset %s / level %.0f", want.Selected, want.Number) {
			return errors.New("controlled status caption")
		}
		report.Steps = append(report.Steps, choiceObservation{label, actual, snap})
		return nil
	}
	send := func(label string, snap cogent.DesktopSnapshot, in cogent.DesktopInput, reject bool) error {
		in.Frame = snap.Frame
		e := s.Input(ctx, ref, in)
		report.Inputs = append(report.Inputs, catalogueInput{label, snap, in, reject})
		if reject {
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return fmt.Errorf("%s disabled refusal: %v", label, e)
			}
			return nil
		}
		if e != nil {
			return fmt.Errorf("%s: %w", label, e)
		}
		return nil
	}
	click := func(label, id, option string) error {
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			snap, e := catalogueReady(ctx, s, ref)
			if e != nil {
				return e
			}
			el, e := choiceElement(snap, id, option)
			if e != nil {
				return e
			}
			if e = send(label+"/"+typ.String(), snap, cogent.DesktopInput{Target: el.ID, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}, false); e != nil {
				return e
			}
		}
		return nil
	}
	chord := func(label, id string, code key.Codes) error {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		el, e := choiceElement(snap, id, "")
		if e != nil {
			return e
		}
		if !el.Focused {
			return fmt.Errorf("%s actual focused target missing", label)
		}
		in := cogent.DesktopInput{Target: el.ID, Type: events.KeyChord, Code: code}
		if code == key.CodeSpacebar {
			in.Rune = ' '
		} // printable Go key events carry their rune
		return send(label, snap, in, false)
	}
	refuse := func(label, id, option string, typ events.Types) error {
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			return e
		}
		el, e := choiceElement(snap, id, option)
		if e != nil {
			return e
		}
		if el.Enabled || el.Bounds.Empty() {
			return fmt.Errorf("%s disabled target not published", label)
		}
		return send(label, snap, cogent.DesktopInput{Target: el.ID, Type: typ, Code: key.CodeRightArrow, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}, true)
	}
	if err = observe("initial"); err != nil {
		return
	}
	if err = chord("disable-editing", "enable", key.CodeSpacebar); err != nil {
		return
	}
	want.Enabled = false
	want.EnableRequests = 1
	if err = observe("editing-disabled"); err != nil {
		return
	}
	if err = refuse("disabled-slider", "slider", "", events.KeyChord); err != nil {
		return
	}
	if err = refuse("disabled-radio", "radio-preset", "advanced", events.MouseDown); err != nil {
		return
	}
	snap, e := catalogueReady(ctx, s, ref)
	if e != nil {
		err = e
		return
	}
	if err = send("disabled-ordinary-up", snap, cogent.DesktopInput{Type: events.KeyChord, Code: key.CodeUpArrow}, false); err != nil {
		return
	}
	if err = observe("disabled-no-proposals"); err != nil {
		return
	}
	if err = click("reenable-editing", "enable", ""); err != nil {
		return
	}
	want.Enabled = true
	want.EnableRequests = 2
	if err = observe("editing-reenabled"); err != nil {
		return
	}
	if err = click("switch-on", "snapping", ""); err != nil {
		return
	}
	want.Snapping = true
	want.SnappingRequests = 1
	if err = observe("switch-pointer"); err != nil {
		return
	}
	if err = chord("switch-off", "snapping", key.CodeSpacebar); err != nil {
		return
	}
	want.Snapping = false
	want.SnappingRequests = 2
	if err = observe("switch-keyboard"); err != nil {
		return
	}
	if err = click("pin-on", "pinned", ""); err != nil {
		return
	}
	want.Pinned = true
	want.PinnedRequests = 1
	if err = observe("toggle-pointer"); err != nil {
		return
	}
	if err = chord("pin-off", "pinned", key.CodeSpacebar); err != nil {
		return
	}
	want.Pinned = false
	want.PinnedRequests = 2
	if err = observe("toggle-keyboard"); err != nil {
		return
	}
	if err = click("radio-advanced", "radio-preset", "advanced"); err != nil {
		return
	}
	want.Selected = "advanced"
	want.SelectionRequests = 1
	if err = observe("radio-pointer"); err != nil {
		return
	}
	if err = chord("radio-wrap", "radio-preset", key.CodeRightArrow); err != nil {
		return
	}
	want.Selected = "basic"
	want.SelectionRequests = 2
	if err = observe("radio-wrap-skips-disabled"); err != nil {
		return
	}
	if err = setup(func() { g.selectionReject = true }); err != nil {
		return
	}
	if err = chord("radio-reject", "radio-preset", key.CodeRightArrow); err != nil {
		return
	}
	want.SelectionRequests = 3
	if err = observe("radio-rejected"); err != nil {
		return
	}
	if err = setup(func() { g.selectionReject = false }); err != nil {
		return
	}
	if err = click("segments-advanced", "segmented-preset", "advanced"); err != nil {
		return
	}
	want.Selected = "advanced"
	want.SelectionRequests = 4
	if err = observe("segments-pointer-shared-selection"); err != nil {
		return
	}
	if err = chord("segments-wrap", "segmented-preset", key.CodeRightArrow); err != nil {
		return
	}
	want.Selected = "basic"
	want.SelectionRequests = 5
	if err = observe("segments-wrap-skips-disabled"); err != nil {
		return
	}
	if err = chord("segments-end", "segmented-preset", key.CodeEnd); err != nil {
		return
	}
	want.Selected = "advanced"
	want.SelectionRequests = 6
	if err = observe("segments-end"); err != nil {
		return
	}
	if err = chord("segments-noop", "segmented-preset", key.CodeSpacebar); err != nil {
		return
	}
	if err = observe("segments-no-duplicate"); err != nil {
		return
	}
	if err = refuse("disabled-option", "segmented-preset", "unavailable", events.MouseDown); err != nil {
		return
	}
	if err = observe("disabled-option-refused"); err != nil {
		return
	}
	if err = chord("tab-to-slider", "segmented-preset", key.CodeTab); err != nil {
		return
	}
	if err = observe("slider-keyboard-focus"); err != nil {
		return
	}
	if err = chord("slider-increment", "slider", key.CodeRightArrow); err != nil {
		return
	}
	want.Number = 26
	want.NumberRequests = 1
	if err = observe("slider-accepted"); err != nil {
		return
	}
	if err = setup(func() { g.numberReject = true }); err != nil {
		return
	}
	if err = chord("slider-reject", "slider", key.CodeRightArrow); err != nil {
		return
	}
	want.NumberRequests = 2
	if err = observe("slider-rejected"); err != nil {
		return
	}
	if err = setup(func() { g.numberReject = false }); err != nil {
		return
	}
	if err = chord("slider-min", "slider", key.CodeHome); err != nil {
		return
	}
	want.Number = 0
	want.NumberRequests = 3
	if err = observe("slider-min-progress-zero"); err != nil {
		return
	}
	if err = chord("slider-max", "slider", key.CodeEnd); err != nil {
		return
	}
	want.Number = 100
	want.NumberRequests = 4
	if err = observe("slider-max-progress-full"); err != nil {
		return
	}
	if err = chord("slider-boundary", "slider", key.CodeRightArrow); err != nil {
		return
	}
	want.NumberRequests = 5
	if err = observe("slider-boundary-once"); err != nil {
		return
	}
	if err = chord("tab-to-action", "slider", key.CodeTab); err != nil {
		return
	}
	if err = chord("acknowledge", "outside", key.CodeSpacebar); err != nil {
		return
	}
	want.OutsideClicks = 1
	if err = observe("ordinary-action-once"); err != nil {
		return
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	target, e := choiceElement(last, "slider", "")
	if e != nil {
		err = e
		return
	}
	if err = catalogueOwner(ctx, a, ref, func() error { return a.CloseWindow(ref) }); err != nil {
		return
	}
	report.OriginalRefClosed = ref.Context().Err() != nil
	e = s.Input(ctx, ref, cogent.DesktopInput{Frame: last.Frame, Target: target.ID, Type: events.KeyChord, Code: key.CodeRightArrow})
	if !report.OriginalRefClosed || !errors.Is(e, cogent.ErrDesktopClosed) {
		err = fmt.Errorf("closed original window admitted choice input: %v", e)
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_CHOICES"
	return
}

func runChoiceCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	type outcome struct {
		report choiceReport
		err    error
	}
	done := make(chan outcome, 1)
	go func() { r, e := checkChoiceCatalogue(ctx, a, g, s, ref); done <- outcome{r, e}; cancel() }()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	var result outcome
	select {
	case result = <-done:
	case <-time.After(6 * time.Second):
		fmt.Fprintln(os.Stderr, "choice catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("choice catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_CHOICES"
		result.report.Error = result.err.Error()
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}
