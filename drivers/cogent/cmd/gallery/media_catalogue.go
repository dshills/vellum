package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type mediaModel struct {
	Page, Path, Fit, Hex, Text                                           string
	Tools, Last, PathRequests, FitRequests, Replacements, Edits, Invalid int
	RejectPath, DisablePath, RejectColor, NormalizeColor, DisableColor   bool
}

func (g *gallery) mediaModel() mediaModel {
	ids := make([]string, len(g.breadcrumb.Items))
	for i, item := range g.breadcrumb.Items {
		ids[i] = item.ID
	}
	return mediaModel{g.mediaPage, strings.Join(ids, "/"), string(g.imageFit), ui.ColorHexString(g.colorPicker.Color), g.text, g.mediaTools, g.mediaLast, g.breadcrumbRequests, g.imageFitRequests, g.imageReplacements, g.colorEdits, g.colorInvalid, g.breadcrumbReject, g.breadcrumb.Disabled, g.colorReject, g.colorNormalize, g.colorPicker.Disabled}
}

type mediaStep struct {
	Label, Kind, Target, Option, Focus, Text string
	Code                                     key.Codes
	Want                                     mediaModel
	Setup                                    func()
}

// Expected state is authored independently of the callbacks being exercised.
func (g *gallery) mediaScript() []mediaStep {
	w := mediaModel{Page: "chrome", Path: "projects/workspace/document", Fit: "contain", Hex: "#2060E0FF", Text: "Vellum"}
	var steps []mediaStep
	add := func(label, kind, target, option, focus, text string, code key.Codes, setup func()) {
		steps = append(steps, mediaStep{label, kind, target, option, focus, text, code, w, setup})
	}
	workspace := ui.BreadcrumbItemID("project-path", "workspace")
	red, alpha, hex := ui.ColorFieldID("ink", ui.ColorRed), ui.ColorFieldID("ink", ui.ColorAlpha), ui.ColorFieldID("ink", ui.ColorHex)
	add("initial", "observe", "", "", "media-tool-0", "", 0, nil)
	add("toolbar-end", "key", "media-tool-0", "", "media-tool-7", "", key.CodeEnd, nil)
	w.Tools = 1
	w.Last = 7
	add("toolbar-action", "key", "media-tool-7", "", "media-tool-7", "", key.CodeReturnEnter, nil)
	add("toolbar-home", "key", "media-tool-7", "", "media-tool-0", "", key.CodeHome, nil)
	add("toolbar-skip-disabled", "key", "media-tool-0", "", "media-tool-2", "", key.CodeRightArrow, nil)
	add("toolbar-last", "key", "media-tool-2", "", "media-tool-7", "", key.CodeEnd, nil)
	add("form-traversal", "key", "media-tool-7", "", "media-name", "", key.CodeTab, nil)
	w.Path = "projects/workspace"
	w.PathRequests = 1
	add("breadcrumb-accept", "click", workspace, "", "", "", 0, nil)
	add("breadcrumb-stale-frame", "stale-path", "", "", "", "", 0, nil)
	w.Path = "projects/workspace/document"
	add("breadcrumb-reset", "click", "media-reset-path", "", "media-reset-path", "", 0, nil)
	w.RejectPath = true
	add("breadcrumb-reject-policy", "observe", "", "", "", "", 0, func() { g.breadcrumbReject = true })
	w.PathRequests = 2
	add("breadcrumb-rejected", "click", workspace, "", workspace, "", 0, nil)
	w.RejectPath = false
	w.DisablePath = true
	add("breadcrumb-disabled", "observe", "", "", "", "", 0, func() { g.breadcrumbReject = false; g.breadcrumb.Disabled = true })
	add("breadcrumb-disabled-refusal", "refuse", workspace, "", "", "", 0, nil)
	w.DisablePath = false
	add("breadcrumb-enabled", "observe", "", "", "", "", 0, func() { g.breadcrumb.Disabled = false })
	w.Path = "projects/workspace"
	w.PathRequests = 3
	add("breadcrumb-restored", "click", workspace, "", "", "", 0, nil)
	w.Page = "images"
	add("images-page", "click", "media-pages", "images", "media-pages", "", 0, nil)
	w.Fit = "cover"
	w.FitRequests = 1
	add("image-cover", "click", "image-fit", "cover", "image-fit", "", 0, nil)
	w.Fit = "stretch"
	w.FitRequests = 2
	add("image-stretch", "key", "image-fit", "", "image-fit", "", key.CodeEnd, nil)
	w.Fit = "contain"
	w.FitRequests = 3
	add("image-contain", "key", "image-fit", "", "image-fit", "", key.CodeHome, nil)
	w.Replacements = 1
	add("image-replace", "click", "image-replace", "", "image-replace", "", 0, nil)
	w.Page = "colors"
	add("colors-page", "click", "media-pages", "colors", "media-pages", "", 0, nil)
	add("color-tab", "key", "media-pages", "", red, "", key.CodeTab, nil)
	w.Hex = "#2160E0FF"
	w.Edits = 1
	add("color-red-step", "key", red, "", red, "", key.CodeRightArrow, nil)
	w.Hex = "#FF60E0FF"
	w.Edits = 2
	add("color-red-max", "key", red, "", red, "", key.CodeEnd, nil)
	w.RejectColor = true
	add("color-reject-policy", "observe", "", "", red, "", 0, func() { g.colorReject = true })
	w.Edits = 3
	add("color-rejected", "key", red, "", red, "", key.CodeHome, nil)
	w.RejectColor = false
	w.NormalizeColor = true
	add("color-normalize-policy", "observe", "", "", red, "", 0, func() { g.colorReject = false; g.colorNormalize = true })
	add("color-green-tab", "key", red, "", ui.ColorFieldID("ink", ui.ColorGreen), "", key.CodeTab, nil)
	add("color-blue-tab", "key", ui.ColorFieldID("ink", ui.ColorGreen), "", ui.ColorFieldID("ink", ui.ColorBlue), "", key.CodeTab, nil)
	add("color-alpha-tab", "key", ui.ColorFieldID("ink", ui.ColorBlue), "", alpha, "", key.CodeTab, nil)
	w.Edits = 4
	add("color-alpha-normalized", "key", alpha, "", alpha, "", key.CodeHome, nil)
	add("color-hex-tab", "key", alpha, "", hex, "", key.CodeTab, nil)
	w.Hex = "#AABBCCFF"
	w.Edits = 5
	add("color-hex-accepted", "text", hex, "", "", "#abc", 0, nil)
	w.Invalid = 1
	add("color-hex-invalid", "text", hex, "", "", "invalid", 0, nil)
	w.NormalizeColor = false
	add("color-straight-alpha-policy", "observe", "", "", "", "", 0, func() { g.colorNormalize = false })
	w.Hex = "#F001AB00"
	w.Edits = 6
	add("color-straight-alpha", "text", hex, "", "", "#F001AB00", 0, nil)
	w.DisableColor = true
	add("color-disabled", "observe", "", "", "", "", 0, func() { g.colorPicker.Disabled = true })
	add("color-disabled-refusal", "refuse", red, "", "", "", 0, nil)
	w.DisableColor = false
	add("color-reenabled", "observe", "", "", "", "", 0, func() { g.colorPicker.Disabled = false })
	w.Page = "chrome"
	add("chrome-return", "click", "media-pages", "chrome", "media-pages", "", 0, nil)
	add("hidden-color-refusal", "stale-color", "", "", "media-pages", "", 0, nil)
	return steps
}

type mediaObservation struct {
	Label    string
	Model    mediaModel
	Snapshot cogent.DesktopSnapshot
}
type mediaReport struct {
	Event, Result, Error string
	ActiveStep           string `json:",omitempty"`
	Steps                []mediaObservation
	Inputs               []catalogueInput
	Failed               *mediaObservation
	Unavailable          bool
	Session              cogent.DesktopState
	OriginalRefClosed    bool
	PostCloseInputError  string
}

func mediaObserve(s cogent.DesktopSnapshot, step mediaStep) error {
	if step.Label == "hidden-color-refusal" {
		_, e := catalogueElement(s, ui.ColorFieldID("ink", ui.ColorHex))
		if e == nil {
			return errors.New("hidden color retains visible input bounds")
		}
	}
	if len(s.Popups) != 0 {
		return errors.New("unexpected popup")
	}
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
			return fmt.Errorf("focus %s absent", step.Focus)
		}
	}
	if step.Want.Page == "chrome" {
		toolbar, e := visible("media-tools")
		if e != nil {
			return e
		}
		if strings.HasPrefix(step.Focus, "media-tool-") {
			tool, e := visible(step.Focus)
			if e != nil {
				return e
			}
			if !tool.Bounds.In(toolbar.Bounds) {
				return errors.New("focused tool outside toolbar viewport")
			}
		}
		if step.Label == "initial" {
			last, e := catalogueElement(s, "media-tool-7")
			if e != nil {
				return e
			}
			if !last.Bounds.Empty() && last.Bounds.In(toolbar.Bounds) {
				return errors.New("fixture does not overflow toolbar")
			}
		}

		form, err := visible("media-form")
		if err != nil {
			return err
		}
		if form.Name != "<Project>" {
			return fmt.Errorf("literal form name %q", form.Name)
		}
		status, err := visible("media-status")
		if err != nil {
			return err
		}
		var text string
		want := fmt.Sprintf("Tool %d / actions %d / path %d", step.Want.Last, step.Want.Tools, step.Want.PathRequests)
		if json.Unmarshal(status.Value, &text) != nil || text != want {
			return fmt.Errorf("status %s want %s", text, want)
		}
		workspace, err := catalogueElement(s, ui.BreadcrumbItemID("project-path", "workspace"))
		if err != nil {
			return err
		}
		wantRole := string(ui.Button)
		if step.Want.Path == "projects/workspace" {
			wantRole = string(ui.Label)
		}
		if workspace.Role != wantRole {
			return fmt.Errorf("breadcrumb current role %s want %s", workspace.Role, wantRole)
		}
		if workspace.Enabled == step.Want.DisablePath {
			return errors.New("breadcrumb enabled publication")
		}
	} else if step.Want.Page == "images" {
		image, err := visible("image-preview")
		if err != nil {
			return err
		}
		if image.Role != string(ui.Image) {
			return errors.New("image role")
		}
		fit, err := visible("image-fit")
		if err != nil {
			return err
		}
		var value string
		if json.Unmarshal(fit.Value, &value) != nil || value != map[string]string{"contain": "Contain", "cover": "Cover", "stretch": "Stretch"}[step.Want.Fit] {
			return errors.New("image fit publication")
		}
	} else {
		rgba, err := ui.ParseColorHex(step.Want.Hex)
		if err != nil {
			return err
		}
		for channel, want := range map[ui.ColorChannel]uint8{ui.ColorRed: rgba.R, ui.ColorGreen: rgba.G, ui.ColorBlue: rgba.B, ui.ColorAlpha: rgba.A} {
			e, err := visible(ui.ColorFieldID("ink", channel))
			if err != nil {
				return err
			}
			var value float32
			if json.Unmarshal(e.Value, &value) != nil || value != float32(want) || e.Enabled == step.Want.DisableColor {
				return fmt.Errorf("color field %s publication", channel)
			}
		}
		e, err := visible("ink/color/preview")
		if err != nil {
			return err
		}
		if e.Name != "Ink color preview "+step.Want.Hex {
			return errors.New("color preview publication")
		}
		_, err = catalogueElement(s, "ink/color/validation")
		wantMessage := step.Want.Invalid > 0 && step.Want.Edits == 5
		if (err == nil) != wantMessage {
			return errors.New("color validation publication")
		}
	}
	return nil
}
func checkMediaCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report mediaReport, err error) {
	report.Event = "vellum.media-native.result"
	report.Result = "FAILED_NATIVE_MEDIA"
	defer func() {
		if err != nil {
			report.Error = err.Error()
			report.Unavailable = errors.Is(err, cogent.ErrDesktopFrameUnavailable)
		}
	}()
	var stalePath, staleColor cogent.DesktopInput
	send := func(label string, before cogent.DesktopSnapshot, in cogent.DesktopInput, refuse bool) error {
		if in.Frame == 0 {
			in.Frame = before.Frame
		}
		e := s.Input(ctx, ref, in)
		report.Inputs = append(report.Inputs, catalogueInput{label, before, in, refuse})
		if refuse {
			if !errors.Is(e, cogent.ErrDesktopPending) {
				return fmt.Errorf("%s refusal: %w", label, e)
			}
			return nil
		}
		return e
	}
	for _, step := range g.mediaScript() {
		report.ActiveStep = step.Label
		if step.Setup != nil {
			if err = catalogueOwner(ctx, a, ref, func() error { step.Setup(); return a.Invalidate(ref.ID()) }); err != nil {
				return
			}
		}
		types := []events.Types{events.KeyChord}
		if step.Kind == "observe" {
			types = nil
		}
		if step.Kind == "click" || step.Kind == "text" {
			types = []events.Types{events.MouseDown, events.MouseUp}
		}
		if step.Kind == "refuse" {
			types = []events.Types{events.MouseDown}
		}
		for _, typ := range types {
			before, e := catalogueReady(ctx, s, ref)
			if e != nil {
				err = e
				return
			}
			in := cogent.DesktopInput{Type: typ, Code: step.Code}
			reject := step.Kind == "refuse" || step.Kind == "stale-path" || step.Kind == "stale-color"
			if step.Kind == "stale-path" {
				in = stalePath
				in.Type = events.MouseDown
			} else if step.Kind == "stale-color" {
				in = staleColor
				in.Type = events.MouseDown
			} else {
				el, e := choiceElement(before, step.Target, step.Option)
				if e != nil {
					err = e
					return
				}
				in.Target = el.ID
				if typ == events.KeyChord && !el.Focused {
					err = fmt.Errorf("%s target lacks focus", step.Label)
					return
				}
				if typ != events.KeyChord {
					in.Button = events.Left
					in.Position = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
				}
				if step.Target == ui.BreadcrumbItemID("project-path", "workspace") && typ == events.MouseDown && !reject {
					stalePath = in
					stalePath.Frame = before.Frame
				}
				if step.Target == ui.ColorFieldID("ink", ui.ColorHex) && typ == events.MouseDown {
					staleColor = in
				}
			}
			if err = send(step.Label+"/"+typ.String(), before, in, reject); err != nil {
				return
			}
		}
		if step.Kind == "text" {
			if keymap.Of(key.NewChord('a', key.CodeA, catalogueSelectAllModifiers())) != keymap.SelectAll {
				err = errors.New("native Meta+A is not Select All")
				return
			}
			for i, r := range append(append([]rune{'a'}, []rune(step.Text)...), '\n') {
				before, e := catalogueReady(ctx, s, ref)
				if e != nil {
					err = e
					return
				}
				el, e := catalogueElement(before, step.Target)
				if e != nil || !el.Focused {
					err = fmt.Errorf("%s text target lacks focus: %v", step.Label, e)
					return
				}
				in := cogent.DesktopInput{Target: el.ID, Type: events.KeyChord, Rune: r}
				if i == 0 {
					in.Code = key.CodeA
					in.Modifiers = catalogueSelectAllModifiers()
				} else if r == '\n' {
					in.Rune = 0
					in.Code = key.CodeReturnEnter
				}
				if err = send(fmt.Sprintf("%s/text-%d", step.Label, i), before, in, false); err != nil {
					return
				}
			}
		}
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			err = e
			return
		}
		var actual mediaModel
		if err = catalogueOwner(ctx, a, ref, func() error { actual = g.mediaModel(); return nil }); err != nil {
			return
		}
		report.Failed = &mediaObservation{step.Label, actual, snap}
		if actual != step.Want {
			err = fmt.Errorf("%s model %+v want %+v", step.Label, actual, step.Want)
			return
		}
		if err = mediaObserve(snap, step); err != nil {
			err = fmt.Errorf("%s: %w", step.Label, err)
			return
		}
		report.Steps = append(report.Steps, *report.Failed)
		report.Failed = nil
		report.ActiveStep = ""
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	target, e := catalogueElement(last, "media-pages")
	if e != nil {
		err = e
		return
	}
	report.OriginalRefClosed, err = mediaFinishWindow(ctx, a, s, ref, cogent.DesktopInput{Frame: last.Frame, Target: target.ID, Type: events.KeyChord, Code: key.CodeEnd})
	if err != nil {
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_MEDIA"
	return
}
func runMediaCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	done := make(chan mediaOutcome, 1)
	go func() { r, e := checkMediaCatalogue(ctx, a, g, s, ref); done <- mediaOutcome{r, e}; cancel() }()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	result, joined := mediaWorkerResult(cancel, done)
	if !joined {
		fmt.Fprintln(os.Stderr, "media catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("media catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_MEDIA"
		result.report.Error = result.err.Error()
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}

// Application retirement can drop an admitted owner callback. Cancel the
// scenario after Run returns so its private reply wait also loses authority.
type mediaOutcome struct {
	report mediaReport
	err    error
}

func mediaWorkerResult(cancel context.CancelFunc, done <-chan mediaOutcome) (mediaOutcome, bool) {
	cancel() // Release owner waits even when Run exits before their callbacks.
	select {
	case result := <-done:
		return result, true
	case <-time.After(6 * time.Second):
		return mediaOutcome{}, false
	}
}

// Final teardown is already authorized after all observations. Keep its close
// acknowledgement and closed-reference check bounded but independent of the
// scenario cancellation triggered when closing the last window ends Run.
func mediaFinishWindow(ctx context.Context, a *app.Application, s *cogent.DesktopSession, ref app.WindowRef, in cogent.DesktopInput) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	if err := catalogueOwner(finalCtx, a, ref, func() error { return a.CloseWindow(ref) }); err != nil {
		return ref.Context().Err() != nil, err
	}
	closed := ref.Context().Err() != nil
	err := s.Input(finalCtx, ref, in)
	if !closed || !errors.Is(err, cogent.ErrDesktopClosed) {
		return closed, fmt.Errorf("closed original admitted media input: %v", err)
	}
	return closed, nil
}
