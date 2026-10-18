package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"strings"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func structureID(kind string, parts ...string) string {
	for i, p := range parts {
		parts[i] = base64.RawURLEncoding.EncodeToString([]byte(p))
	}
	return kind + ":" + strings.Join(parts, ":")
}

type structureModel struct {
	Page, Document, Documents, Text                                                                         string
	Expanded                                                                                                bool
	Fraction                                                                                                float32
	DetailsRequests, SplitRequests, PageRequests, DocumentRequests, DocumentSelections, SectionActions      int
	DetailsReject, DetailsReadOnly, DetailsDisabled, SplitReject, SplitReadOnly, SplitDisabled, CloseReject bool
}

func (g *gallery) structureModel() structureModel {
	return structureModel{g.structurePage, g.documentSelected, strings.Join(g.documents, ","), g.text, g.details, g.split, g.detailsRequests, g.splitRequests, g.structurePageRequests, g.documentCloseRequests, g.documentSelectRequests, g.outsideClicks, g.detailsReject, g.detailsReadOnly, g.detailsDisabled, g.splitReject, g.splitReadOnly, g.splitDisabled, g.documentReject}
}
func structureModelsEqual(a, b structureModel) bool {
	if math.Abs(float64(a.Fraction-b.Fraction)) > 0.00001 {
		return false
	}
	a.Fraction, b.Fraction = 0, 0
	return a == b
}

type structureStep struct {
	Label, Kind, Target, Focus string
	Code                       key.Codes
	Rune                       rune
	Delta                      float32
	Want                       structureModel
	Setup                      func()
}

func (g *gallery) structureScript() []structureStep {
	want := structureModel{Page: "sections", Document: "sketch", Documents: "sketch,notes", Text: "Vellum", Expanded: true, Fraction: 0.35}
	var steps []structureStep
	n := func(id string) string { return structureID("node", id) }
	heading, handle, pages := structureID("disclosure", "details"), structureID("split", "panes"), n("structure-pages")
	add := func(label, kind, target, focus string, code key.Codes, r rune, delta float32, setup func()) {
		steps = append(steps, structureStep{label, kind, target, focus, code, r, delta, want, setup})
	}
	add("initial", "observe", "", heading, 0, 0, 0, nil)
	want.Expanded = false
	want.DetailsRequests = 1
	add("section-collapse", "key", heading, heading, key.CodeLeftArrow, 0, 0, nil)
	add("section-tab-skips-hidden", "key", heading, n("section-after"), key.CodeTab, 0, 0, nil)
	want.Expanded = true
	want.DetailsRequests = 2
	add("section-pointer-reopen", "click", heading, heading, 0, 0, 0, nil)
	add("section-tab-draft", "key", heading, n("section-draft"), key.CodeTab, 0, 0, nil)
	want.SectionActions = 1
	add("section-child-action", "click", n("section-action"), n("section-action"), 0, 0, 0, nil)
	want.Expanded = false
	add("section-owner-collapse", "observe", "", heading, 0, 0, 0, func() { g.details = false })
	want.Expanded = true
	want.DetailsRequests = 3
	add("section-right", "key", heading, heading, key.CodeRightArrow, 0, 0, nil)
	want.Expanded = false
	want.DetailsRequests = 4
	add("section-enter", "key", heading, heading, key.CodeReturnEnter, 0, 0, nil)
	want.Expanded = true
	want.DetailsRequests = 5
	add("section-space", "key", heading, heading, key.CodeSpacebar, ' ', 0, nil)
	want.DetailsReject = true
	want.DetailsRequests = 6
	add("section-rejected", "key", heading, heading, key.CodeLeftArrow, 0, 0, func() { g.detailsReject = true })
	want.DetailsReject = false
	want.DetailsReadOnly = true
	add("section-readonly", "key", heading, heading, key.CodeLeftArrow, 0, 0, func() { g.detailsReject = false; g.detailsReadOnly = true })
	want.DetailsReadOnly = false
	want.DetailsDisabled = true
	add("section-disabled", "observe", "", "", 0, 0, 0, func() { g.detailsReadOnly = false; g.detailsDisabled = true })
	add("section-disabled-refusal", "refuse", heading, "", key.CodeLeftArrow, 0, 0, nil)
	want.DetailsDisabled = false
	want.Expanded = false
	want.DetailsRequests = 7
	add("section-reenabled", "click", heading, heading, 0, 0, 0, func() { g.detailsDisabled = false })
	add("page-disabled-refusal", "refuse", structureID("option", "structure-pages", "unavailable"), heading, key.CodeRightArrow, 0, 0, nil)
	want.Page = "split-page"
	want.PageRequests = 1
	add("page-split-pointer", "click", structureID("option", "structure-pages", "split-page"), pages, 0, 0, 0, nil)
	want.Page = "sections"
	want.PageRequests = 2
	add("page-home", "key", pages, pages, key.CodeHome, 0, 0, nil)
	want.Page = "split-page"
	want.PageRequests = 3
	add("page-disabled-skip", "key", pages, pages, key.CodeRightArrow, 0, 0, nil)
	add("split-focus", "click", handle, handle, 0, 0, 0, nil)
	want.Fraction = 0.36
	want.SplitRequests = 1
	add("split-right", "key", handle, handle, key.CodeRightArrow, 0, 0, nil)
	want.Fraction = 0.35
	want.SplitRequests = 2
	add("split-left", "key", handle, handle, key.CodeLeftArrow, 0, 0, nil)
	want.Fraction = 0.05
	want.SplitRequests = 3
	add("split-home", "key", handle, handle, key.CodeHome, 0, 0, nil)
	want.Fraction = 0.95
	want.SplitRequests = 4
	add("split-end", "key", handle, handle, key.CodeEnd, 0, 0, nil)
	want.SplitReject = true
	want.SplitRequests = 5
	add("split-rejected", "key", handle, handle, key.CodeHome, 0, 0, func() { g.splitReject = true })
	want.SplitReject = false
	want.SplitReadOnly = true
	add("split-readonly", "key", handle, handle, key.CodeHome, 0, 0, func() { g.splitReject = false; g.splitReadOnly = true })
	want.SplitReadOnly = false
	want.Fraction = 0.05
	want.SplitRequests = 6
	add("split-writable", "key", handle, handle, key.CodeHome, 0, 0, func() { g.splitReadOnly = false })
	want.Fraction = 0.35
	add("split-reset", "click", n("split-reset"), n("split-reset"), 0, 0, 0, nil)
	add("scroll-bottom", "scroll", n("split-workspace"), "", 0, 0, 100000, nil)
	add("scroll-boundary", "scroll", n("split-workspace"), "", 0, 0, 100000, nil)
	add("scroll-top", "scroll", n("split-workspace"), "", 0, 0, -100000, nil)
	want.SplitDisabled = true
	add("split-disabled", "observe", "", "", 0, 0, 0, func() { g.splitDisabled = true })
	add("split-disabled-refusal", "refuse", handle, "", key.CodeHome, 0, 0, nil)
	want.SplitDisabled = false
	add("split-reenabled", "click", handle, handle, 0, 0, 0, func() { g.splitDisabled = false })
	want.Page = "documents"
	want.PageRequests = 4
	add("page-documents", "click", structureID("option", "structure-pages", "documents"), pages, 0, 0, 0, nil)
	want.Document = "notes"
	want.DocumentSelections = 1
	add("document-notes", "click", structureID("option", "document-tabs", "notes"), n("document-tabs"), 0, 0, 0, nil)
	want.CloseReject = true
	want.DocumentRequests = 1
	add("document-pointer-rejected", "click", structureID("tab-close", "document-tabs", "notes"), n("document-tabs"), 0, 0, 0, func() { g.documentReject = true })
	want.DocumentRequests = 2
	add("document-key-rejected", "key", n("document-tabs"), n("document-tabs"), key.CodeDelete, 0, 0, nil)
	want.CloseReject = false
	want.DocumentRequests = 3
	want.Documents = "sketch"
	want.Document = "sketch"
	add("document-key-accepted", "key", n("document-tabs"), "", key.CodeDelete, 0, 0, func() { g.documentReject = false })
	add("document-stale-close", "stale", "", "", 0, 0, 0, nil)
	want.Page = "sections"
	want.PageRequests = 5
	add("page-retained-section", "click", structureID("option", "structure-pages", "sections"), pages, 0, 0, 0, nil)
	want.Expanded = true
	want.DetailsRequests = 8
	add("section-final-reopen", "click", heading, heading, 0, 0, 0, nil)
	return steps
}

type structureObservation struct {
	Label    string
	Model    structureModel
	Snapshot cogent.DesktopSnapshot
}
type structureReport struct {
	Event, Result, Error string
	Steps                []structureObservation
	Inputs               []catalogueInput
	Failed               *structureObservation
	Session              cogent.DesktopState
	OriginalRefClosed    bool
	PostCloseInputError  string
}

func structureElement(s cogent.DesktopSnapshot, id string) (cogent.DesktopElement, error) {
	for _, e := range s.Elements {
		if e.ID == id {
			return e, nil
		}
	}
	return cogent.DesktopElement{}, fmt.Errorf("structure target %q absent", id)
}
func structureObserve(s cogent.DesktopSnapshot, step structureStep) error {
	if len(s.Popups) != 0 {
		return errors.New("unexpected structure popup")
	}
	find := func(kind string, parts ...string) (cogent.DesktopElement, error) {
		return structureElement(s, structureID(kind, parts...))
	}
	visible := func(e cogent.DesktopElement) bool {
		return !e.Bounds.Empty() && e.Bounds.In(image.Rectangle{Max: s.Size})
	}
	for _, id := range []string{"structure-title", "structure-pages", "structure-status"} {
		e, err := find("node", id)
		if err != nil {
			return err
		}
		if !visible(e) {
			return fmt.Errorf("structure %s bounds", id)
		}
	}
	for _, id := range []string{"sections", "unavailable", "split-page", "documents"} {
		e, err := find("option", "structure-pages", id)
		if err != nil {
			return err
		}
		if !visible(e) || e.Enabled != (id != "unavailable") || e.Selected != (id == step.Want.Page) {
			return fmt.Errorf("structure page %s projection", id)
		}
	}
	n, err := find("node", "structure-status")
	if err != nil {
		return err
	}
	var value string
	count := len(strings.Split(step.Want.Documents, ","))
	if json.Unmarshal(n.Value, &value) != nil || value != fmt.Sprintf("Page %s / expanded %t / split %.2f / documents %d", step.Want.Page, step.Want.Expanded, step.Want.Fraction, count) {
		return errors.New("structure status accepted state")
	}
	switch step.Want.Page {
	case "sections":
		h, err := find("disclosure", "details")
		if err != nil {
			return err
		}
		var expanded bool
		if !visible(h) || h.Enabled == step.Want.DetailsDisabled || json.Unmarshal(h.Value, &expanded) != nil || expanded != step.Want.Expanded {
			return errors.New("disclosure accepted geometry/state")
		}
		after, err := find("node", "section-after")
		if err != nil || !visible(after) {
			return errors.New("after-section geometry")
		}
		for _, id := range []string{"section-draft", "section-action"} {
			e, err := find("node", id)
			if !step.Want.Expanded {
				if err == nil {
					return errors.New("collapsed child target retained")
				}
				continue
			}
			if err != nil || !visible(e) || e.Enabled == step.Want.DetailsDisabled {
				return fmt.Errorf("expanded child %s bounds/state", id)
			}
		}
		if step.Want.Expanded {
			e, err := find("node", "section-draft")
			if err != nil {
				return err
			}
			if json.Unmarshal(e.Value, &value) != nil || value != step.Want.Text {
				return errors.New("retained section text")
			}
		}
	case "split-page":
		pane, err := find("node", "panes")
		if err != nil {
			return err
		}
		h, err := find("split", "panes")
		if err != nil {
			return err
		}
		var f float32
		if !visible(h) || !h.Bounds.In(pane.Bounds) || h.Enabled == step.Want.SplitDisabled || json.Unmarshal(h.Value, &f) != nil || math.Abs(float64(f-step.Want.Fraction)) > 0.00001 {
			return errors.New("split accepted divider geometry/state")
		}
		wantX := float64(pane.Bounds.Min.X) + float64(step.Want.Fraction)*float64(pane.Bounds.Dx()-h.Bounds.Dx())
		if math.Abs(float64(h.Bounds.Min.X)-wantX) > 4 {
			return fmt.Errorf("split divider position %d want %.2f", h.Bounds.Min.X, wantX)
		}
		if step.Label == "scroll-bottom" || step.Label == "scroll-boundary" || step.Label == "scroll-top" || step.Label == "split-reset" {
			first, err := find("node", "split-line-00")
			if err != nil {
				return err
			}
			last, err := find("node", "split-line-23")
			if err != nil {
				return err
			}
			bottom := step.Label == "scroll-bottom" || step.Label == "scroll-boundary"
			if first.Bounds.Empty() != bottom || last.Bounds.Empty() == bottom {
				return errors.New("scroll clipped endpoint projection")
			}
		}
	case "documents":
		for _, id := range []string{"sketch", "notes"} {
			e, err := find("option", "document-tabs", id)
			present := strings.Contains(","+step.Want.Documents+",", ","+id+",")
			if !present {
				if err == nil {
					return errors.New("retired document option retained")
				}
				if _, err = find("tab-close", "document-tabs", id); err == nil {
					return errors.New("retired document close retained")
				}
				continue
			}
			if err != nil || !visible(e) || !e.Enabled || e.Selected != (id == step.Want.Document) {
				return errors.New("document tab projection")
			}
			c, err := find("tab-close", "document-tabs", id)
			if err != nil || !visible(c) || !c.Enabled || c.Parent != e.ID {
				return errors.New("document close target projection")
			}
		}
		e, err := find("node", "document-content-"+step.Want.Document)
		if err != nil || !visible(e) {
			return errors.New("active document content bounds")
		}
		want := map[string]string{"sketch": "Untitled sketch", "notes": "Working notes"}[step.Want.Document]
		if json.Unmarshal(e.Value, &value) != nil || value != want {
			return errors.New("retained document draft")
		}
	default:
		return errors.New("unexpected structure page")
	}
	if step.Focus != "" {
		e, err := structureElement(s, step.Focus)
		if err != nil || !e.Focused || !visible(e) {
			return fmt.Errorf("structure focus %s absent", step.Focus)
		}
	}
	return nil
}
func checkStructureCatalogue(ctx context.Context, a *app.Application, g *gallery, s *cogent.DesktopSession, ref app.WindowRef) (report structureReport, err error) {
	report.Event = "vellum.structure-native.result"
	report.Result = "FAILED_NATIVE_STRUCTURE"
	var stale cogent.DesktopInput
	for _, step := range g.structureScript() {
		if step.Setup != nil {
			if err = catalogueOwner(ctx, a, ref, func() error { step.Setup(); return a.Invalidate(ref.ID()) }); err != nil {
				return
			}
		}
		if step.Kind != "observe" {
			types := []events.Types{events.KeyChord}
			if step.Kind == "click" {
				types = []events.Types{events.MouseDown, events.MouseUp}
			}
			if step.Kind == "scroll" {
				types = []events.Types{events.Scroll}
			}
			if step.Kind == "stale" {
				types = []events.Types{events.MouseDown}
			}
			for _, typ := range types {
				before, e := catalogueReady(ctx, s, ref)
				if e != nil {
					err = e
					return
				}
				in := cogent.DesktopInput{Type: typ, Code: step.Code, Rune: step.Rune}
				reject := step.Kind == "refuse" || step.Kind == "stale"
				if step.Kind == "stale" {
					in = stale
					in.Type = events.MouseDown
					if in.Target == "" {
						err = errors.New("missing retired close receipt")
						return
					}
				} else {
					el, e := structureElement(before, step.Target)
					if e != nil {
						err = e
						return
					}
					in.Target = el.ID
					if typ == events.KeyChord && !reject && !el.Focused {
						err = fmt.Errorf("%s current key target lacks focus", step.Label)
						return
					}
					if typ != events.KeyChord {
						if el.Bounds.Empty() {
							err = fmt.Errorf("%s pointer/scroll target bounds empty", step.Label)
							return
						}
						in.Position = el.Bounds.Min.Add(el.Bounds.Size().Div(2))
						if typ == events.Scroll {
							in.Delta = math32.Vec2(0, step.Delta)
						} else {
							in.Button = events.Left
						}
					}
					if step.Label == "document-pointer-rejected" {
						stale = in
					}
				}
				in.Frame = before.Frame
				e = s.Input(ctx, ref, in)
				label := step.Label
				if len(types) == 2 {
					label += "/" + typ.String()
				}
				report.Inputs = append(report.Inputs, catalogueInput{label, before, in, reject})
				if reject {
					if !errors.Is(e, cogent.ErrDesktopPending) {
						err = fmt.Errorf("%s expected refusal: %v", step.Label, e)
						return
					}
				} else if e != nil {
					err = fmt.Errorf("%s input: %w", step.Label, e)
					return
				}
			}
		}
		snap, e := catalogueReady(ctx, s, ref)
		if e != nil {
			err = e
			return
		}
		var actual structureModel
		if err = catalogueOwner(ctx, a, ref, func() error { actual = g.structureModel(); return nil }); err != nil {
			return
		}
		report.Failed = &structureObservation{step.Label, actual, snap}
		if !structureModelsEqual(actual, step.Want) {
			err = fmt.Errorf("%s model %+v want %+v", step.Label, actual, step.Want)
			return
		}
		if e = structureObserve(snap, step); e != nil {
			err = fmt.Errorf("%s: %w", step.Label, e)
			return
		}
		report.Steps = append(report.Steps, structureObservation{step.Label, actual, snap})
		report.Failed = nil
	}
	if len(report.Steps) == 0 {
		err = errors.New("empty structural record")
		return
	}
	last := report.Steps[len(report.Steps)-1].Snapshot
	target, e := structureElement(last, structureID("disclosure", "details"))
	if e != nil {
		err = e
		return
	}
	report.OriginalRefClosed, err = catalogueFinishWindow(ctx, a, s, ref, cogent.DesktopInput{Frame: last.Frame, Target: target.ID, Type: events.KeyChord, Code: key.CodeReturnEnter})
	if err != nil {
		return
	}
	report.PostCloseInputError = "closed"
	report.Result = "PASS_NATIVE_STRUCTURE"
	return
}
func runStructureCatalogue(parent context.Context, a *app.Application, g *gallery, theme ui.Theme) error {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	s := cogent.NewDesktopSession()
	ref, err := a.WindowRef("gallery")
	if err != nil {
		return err
	}
	done := make(chan catalogueOutcome[structureReport], 1)
	go func() {
		r, e := checkStructureCatalogue(ctx, a, g, s, ref)
		done <- catalogueOutcome[structureReport]{r, e}
		cancel()
	}()
	runErr := a.Run(ctx, cogent.Driver{Theme: theme, DesktopSession: s})
	result, joined := catalogueWorkerResult(cancel, done)
	if !joined {
		fmt.Fprintln(os.Stderr, "structure catalogue worker did not join after native exit")
		os.Exit(3)
	}
	result.report.Session = s.State()
	if !result.report.Session.Closed || result.report.Session.Pending || result.report.Session.LiveWindows != 0 {
		result.err = errors.Join(result.err, errors.New("structure catalogue retained desktop state"))
	}
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		result.err = errors.Join(result.err, runErr)
	}
	if result.err != nil {
		result.report.Result = "FAILED_NATIVE_STRUCTURE"
		result.report.Error = result.err.Error()
	}
	if err = json.NewEncoder(os.Stdout).Encode(result.report); err != nil {
		return err
	}
	return result.err
}
