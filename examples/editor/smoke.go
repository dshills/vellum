package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"cogentcore.org/core/system"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
)

// smokeStore delegates to the real atomic FileStore. One explicitly armed
// write can be held to admit a native preedit during IO without timing guesses.
type smokeWrite struct{ entered, release chan struct{} }
type smokeStore struct{ pending atomic.Pointer[smokeWrite] }

func (s *smokeStore) arm() *smokeWrite {
	gate := &smokeWrite{entered: make(chan struct{}), release: make(chan struct{})}
	s.pending.Store(gate)
	return gate
}
func (s *smokeStore) Read(ctx context.Context, path string, max int64) ([]byte, error) {
	return (document.FileStore{}).Read(ctx, path, max)
}
func (s *smokeStore) Write(ctx context.Context, path string, data []byte, max int64) (bool, error) {
	if gate := s.pending.Swap(nil); gate != nil {
		close(gate.entered)
		select {
		case <-gate.release:
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
	return (document.FileStore{}).Write(ctx, path, data, max)
}

type nativeInput struct {
	Client         uint64     `json:"client_id"`
	Sequence       uint64     `json:"sequence"`
	Text           string     `json:"text"`
	SelectedStart  int        `json:"selected_start"`
	SelectedLength int        `json:"selected_length"`
	HasMarked      bool       `json:"has_marked"`
	Caret          [4]float64 `json:"caret_screen"`
	Active         bool       `json:"active"`
	Cells          int        `json:"cell_count"`
	Carets         int        `json:"caret_count"`
	Viewport       [4]float64 `json:"viewport"`
}
type nativeQuery struct {
	Requested      [2]int     `json:"requested"`
	Substring      string     `json:"substring"`
	SubstringRange [2]int     `json:"substring_range"`
	RectRange      [2]int     `json:"rect_range"`
	Rect           [4]float64 `json:"screen_rect"`
	Index          int        `json:"point_index"`
}

func owner(a *app.Application, fn func() error) error {
	var result error
	if err := a.Do(func() { result = fn() }); err != nil {
		return err
	}
	return result
}
func waitOwner(ctx context.Context, a *app.Application, fn func() (bool, error)) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var ready bool
		err := owner(a, func() error { var e error; ready, e = fn(); return e })
		if err != nil {
			return err
		}
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func readNative(ctx context.Context, w *editorWindow, step int) (nativeInput, error) {
	raw, err := cogent.ProbeTextInput(ctx, w.id, step)
	if err != nil {
		return nativeInput{}, err
	}
	var state nativeInput
	if err = json.Unmarshal([]byte(raw), &state); err != nil {
		return state, fmt.Errorf("native state %q: %w", raw, err)
	}
	return state, nil
}
func waitNativeWhere(ctx context.Context, w *editorWindow, predicate func(nativeInput) bool) (nativeInput, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	var last nativeInput
	var lastErr error
	for {
		last, lastErr = readNative(ctx, w, 0)
		if lastErr == nil && last.Client != 0 && predicate(last) {
			return last, nil
		}
		select {
		case <-ctx.Done():
			return last, fmt.Errorf("native publication wait: got text %q client %d: %w", last.Text, last.Client, errors.Join(lastErr, ctx.Err()))
		case <-ticker.C:
		}
	}
}
func waitNative(ctx context.Context, w *editorWindow, text string) (nativeInput, error) {
	return waitNativeWhere(ctx, w, func(s nativeInput) bool { return s.Text == text })
}

func focus(ctx context.Context, a *app.Application, w *editorWindow, text string) (nativeInput, error) {
	if err := owner(a, func() error { w.focusID = w.editorID; w.focusSequence++; w.invalidate(); return nil }); err != nil {
		return nativeInput{}, err
	}
	return waitNative(ctx, w, text)
}
func query(ctx context.Context, w *editorWindow, start, length int, point [2]float64) (nativeQuery, error) {
	raw, err := cogent.QueryTextInput(ctx, w.id, start, length, point)
	if err != nil {
		return nativeQuery{}, err
	}
	var q nativeQuery
	err = json.Unmarshal([]byte(raw), &q)
	return q, err
}
func waitQuery(ctx context.Context, w *editorWindow, start, length int, predicate func(nativeQuery) bool) (nativeQuery, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	var q nativeQuery
	var err error
	for {
		q, err = query(ctx, w, start, length, [2]float64{})
		if err == nil && predicate(q) {
			return q, nil
		}
		select {
		case <-ctx.Done():
			return q, fmt.Errorf("native geometry not ready: %+v: %w", q, errors.Join(err, ctx.Err()))
		case <-ticker.C:
		}
	}
}
func capture(ctx context.Context, a *app.Application, w *editorWindow) (editor.Snapshot, error) {
	var s editor.Snapshot
	err := owner(a, func() error { s = w.model.Snapshot(); return nil })
	return s, err
}
func waitText(ctx context.Context, a *app.Application, w *editorWindow, text string, composing bool) (editor.Snapshot, error) {
	var s editor.Snapshot
	err := waitOwner(ctx, a, func() (bool, error) {
		s = w.model.Snapshot()
		return s.VisibleText == text && s.Composing == composing, nil
	})
	if err != nil {
		return s, fmt.Errorf("owner editor wait for %q composing %t: got %+v: %w", text, composing, s, err)
	}
	return s, nil
}

// actionState is the owner fingerprint for the native semantic publication.
// Geometry changes do not affect it; model/status/action transitions do.
type actionState struct {
	Editor                                     editor.Snapshot
	Status, Prompt, Filename, Search, EditorID string
	Operation                                  uint64
	Settling, Busy                             bool
}

func (w *editorWindow) actionState() actionState {
	session := w.session.State()
	filename := "Untitled"
	if session.Path != "" {
		filename = filepath.Base(session.Path)
	}
	if session.Dirty {
		filename += " • Unsaved"
	}
	return actionState{Editor: w.model.Snapshot(), Status: w.status, Prompt: w.prompt,
		Filename: filename, Search: w.search, EditorID: w.editorID, Operation: w.operation,
		Settling: w.settling, Busy: session.Busy}
}
func actionSelection(snapshot editor.Snapshot) ([2]int, bool) {
	runes := []rune(snapshot.VisibleText)
	a, f := snapshot.Selection.Anchor, snapshot.Selection.Focus
	if a < 0 || f < 0 || a > len(runes) || f > len(runes) {
		return [2]int{}, false
	}
	anchor, focus := utf16len(string(runes[:a])), utf16len(string(runes[:f]))
	return [2]int{min(anchor, focus), max(anchor, focus) - min(anchor, focus)}, true
}

// actionReady compares accepted owner data to the actual published Cocoa tree.
// A matching value alone is insufficient: unmark can change only selection,
// status and action availability, while a stale admitted action can be dropped.
func actionReady(tree cogent.AccessibilityInspection, state actionState, id string) bool {
	if !tree.Reachable || !tree.NativeWindow || tree.Generation == 0 {
		return false
	}
	elements := map[string]cogent.AccessibilityElement{}
	for _, e := range tree.Elements {
		elements[e.DomainID] = e
	}
	action, ok := elements[id]
	if !ok || !action.Enabled || action.Role != "AXButton" {
		return false
	}
	if state.Prompt != "" {
		_, dialog := elements["decision"]
		_, document := elements[state.EditorID]
		if !dialog || document {
			return false
		}
		return id == "keep" || state.Prompt == "open" && id == "discard-open"
	}
	selection, valid := actionSelection(state.Editor)
	document, documentPresent := elements[state.EditorID]
	text, textPresent := document.Value.(string)
	if !valid || !documentPresent || !textPresent || text != state.Editor.VisibleText || document.TextSelection != selection {
		return false
	}
	for id, value := range map[string]string{"status": state.Status, "filename": state.Filename, "find": state.Search} {
		e, present := elements[id]
		text, textPresent := e.Value.(string)
		if !present || !textPresent || text != value {
			return false
		}
	}
	busy := state.Settling || state.Busy
	for id, enabled := range map[string]bool{
		"open": !busy, "save": !busy, "save-as": !busy,
		"undo": !busy && (state.Editor.History.UndoCount > 0 || state.Editor.Composing),
		"redo": !busy && state.Editor.History.RedoCount > 0, "cancel-operation": busy,
	} {
		e, present := elements[id]
		if !present || e.Enabled != enabled {
			return false
		}
	}
	return true
}
func actionDelivered(id string, before, after actionState) bool {
	switch id {
	case "undo", "redo", "discard-open":
		return after.Operation == before.Operation+1
	case "keep":
		return before.Prompt != "" && after.Prompt == "" && after.Operation == before.Operation
	default:
		return false
	}
}
func waitActionReady(ctx context.Context, w *editorWindow, id string) (actionState, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	var expected, current actionState
	var inspection cogent.AccessibilityInspection
	var inspectionErr error
	var lastSuccessful cogent.AccessibilityInspection
	successfulInspections := 0
	for {
		if err := owner(w.application, func() error { expected = w.actionState(); return nil }); err != nil {
			return expected, err
		}
		inspection, inspectionErr = cogent.InspectAccessibility(ctx, w.id)
		if inspectionErr == nil {
			lastSuccessful = inspection
			successfulInspections++
		}
		if inspectionErr == nil && actionReady(inspection, expected, id) {
			if err := owner(w.application, func() error { current = w.actionState(); return nil }); err != nil {
				return expected, err
			}
			if current == expected {
				return expected, nil
			}
		}
		select {
		case <-ctx.Done():
			// A final timed-out query returns a zero tree; preserve the actual
			// prior publication so readiness failures identify stale fields.
			fmt.Printf("EDITOR_ACTION_READINESS_DIAGNOSTIC %s\n", mustJSON(map[string]any{"action": id, "owner": expected, "successful_inspections": successfulInspections, "last_successful": lastSuccessful, "last_inspection_error": fmt.Sprint(inspectionErr)}))
			return expected, fmt.Errorf("native action %q publication not ready: owner %+v generation %d: %w", id, expected, inspection.Generation, errors.Join(inspectionErr, ctx.Err()))
		case <-ticker.C:
		}
	}
}
func press(ctx context.Context, w *editorWindow, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	before, err := waitActionReady(ctx, w, id)
	if err != nil {
		return err
	}
	ok, err := cogent.PerformAccessibilityAction(ctx, w.id, id, "press", "")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("native button %q refused", id)
	}
	// A FIFO owner fence follows the admitted action. Observe delivery once;
	// do not repeat Undo if its semantic generation became stale before execution.
	fence := make(chan actionState, 1)
	if err := w.application.PostWait(ctx, func() { fence <- w.actionState() }); err != nil {
		return err
	}
	select {
	case after := <-fence:
		if !actionDelivered(id, before, after) {
			return fmt.Errorf("native button %q admitted but not delivered: before %+v after %+v", id, before, after)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func settle(ctx context.Context, a *app.Application, w *editorWindow) error {
	result := make(chan error, 1)
	if err := owner(a, func() error { return cogent.SettleEditor(ctx, a, w.id, w.model, func(err error) { result <- err }) }); err != nil {
		return err
	}
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
func utf16len(text string) int { return len(utf16.Encode([]rune(text))) }
func validRect(rect [4]float64) bool {
	for _, v := range rect {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return rect[2] > 0 && rect[3] > 0
}

func waitInitialNative(ctx context.Context, w *editorWindow) error {
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := waitNative(readyCtx, w, ""); err != nil {
		inspectionCtx, inspectionCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer inspectionCancel()
		inspection, inspectionErr := cogent.InspectAccessibility(inspectionCtx, w.id)
		fmt.Printf("EDITOR_STARTUP_DIAGNOSTIC %s\n", mustJSON(map[string]any{"window": w.id, "accessibility": inspection, "inspection_error": fmt.Sprint(inspectionErr)}))
		return err
	}
	return nil
}

func editorSmoke(ctx context.Context, a *app.Application, first, second *editorWindow) error {
	stage := func(name string) { fmt.Println("stage", name) }
	stage("native-ready")
	if err := waitInitialNative(ctx, first); err != nil {
		return err
	}
	if err := waitInitialNative(ctx, second); err != nil {
		return err
	}
	// Native window names are looked up only after a successful active text-client
	// probe, after creation has published its immutable name under the backend lock.
	firstNative := system.TheApp.WindowByName("Vellum document editor — " + first.id)
	secondNative := system.TheApp.WindowByName("Vellum document editor — " + second.id)
	if firstNative == nil || secondNative == nil || firstNative == secondNative {
		return errors.New("independent native windows unavailable")
	}
	base := "A😀e\u0301 אבג العربية\r\nlast\r\n\r\n"
	compositionPrefix := strings.TrimSuffix(base, "\r\n\r\n")
	compositionText := compositionPrefix + "日本\r\n\r\n"
	if err := owner(a, func() error {
		if err := first.model.Replace(base, editor.Selection{Anchor: utf8.RuneCountInString(compositionPrefix), Focus: utf8.RuneCountInString(compositionPrefix)}); err != nil {
			return err
		}
		first.invalidate()
		if err := second.model.Replace("Independent\r\n", editor.Selection{}); err != nil {
			return err
		}
		second.invalidate()
		return nil
	}); err != nil {
		return err
	}
	initial, err := focus(ctx, a, first, base)
	if err != nil {
		return err
	}
	if initial.SelectedStart != utf16len(compositionPrefix) {
		return fmt.Errorf("scalar/UTF16 selection mismatch: %+v", initial)
	}
	before, err := capture(ctx, a, first)
	if err != nil {
		return err
	}
	stage("settle-admission-refusal")
	canceled, cancelAdmission := context.WithCancel(ctx)
	cancelAdmission()
	callbackCount := 0
	if err = owner(a, func() error {
		cb := func(error) { callbackCount++ }
		if e := cogent.SettleEditor(nil, a, first.id, first.model, cb); e == nil {
			return errors.New("nil settle context admitted")
		}
		if e := cogent.SettleEditor(ctx, a, first.id, first.model, nil); e == nil {
			return errors.New("nil settle callback admitted")
		}
		if e := cogent.SettleEditor(canceled, a, first.id, first.model, cb); !errors.Is(e, context.Canceled) {
			return fmt.Errorf("precanceled settle: %v", e)
		}
		if callbackCount != 0 {
			return errors.New("refused settle called completion")
		}
		return nil
	}); err != nil {
		return err
	}
	stage("admitted-settle-cancellation")
	cancelCtx, cancelSettle := context.WithCancel(ctx)
	cancelResult := make(chan error, 1)
	cancelCalls := 0
	if err = owner(a, func() error {
		e := cogent.SettleEditor(cancelCtx, a, first.id, first.model, func(e error) { cancelCalls++; cancelResult <- e })
		if e == nil {
			cancelSettle()
		}
		return e
	}); err != nil {
		cancelSettle()
		return err
	}
	select {
	case e := <-cancelResult:
		if !errors.Is(e, context.Canceled) {
			return fmt.Errorf("admitted settle cancellation: %v", e)
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	if err = owner(a, func() error {
		if cancelCalls != 1 {
			return fmt.Errorf("settle cancellation completion count %d", cancelCalls)
		}
		return nil
	}); err != nil {
		return err
	}
	stage("controlled-field-focus-handoff")
	if err = owner(a, func() error { first.focusID = "find"; first.focusSequence++; first.invalidate(); return nil }); err != nil {
		return err
	}
	field, err := waitNative(ctx, first, "")
	if err != nil {
		return err
	}
	fmt.Printf("field_focus %s\n", mustJSON(field))
	if field.Client == initial.Client {
		return errors.New("TextField handoff reused multiline client")
	}
	stage("field-mark1")
	if _, err = readNative(ctx, first, 1); err != nil {
		return err
	}
	if _, err = waitNative(ctx, first, "に"); err != nil {
		return err
	}
	if err = owner(a, func() error {
		if first.search != "" || first.searchChanges != 0 || first.model.Snapshot() != before {
			return errors.New("controlled field preedit committed Change or edited document")
		}
		return nil
	}); err != nil {
		return err
	}
	stage("field-mark2")
	if _, err = readNative(ctx, first, 2); err != nil {
		return err
	}
	if _, err = waitNative(ctx, first, "日本"); err != nil {
		return err
	}
	stage("field-commit")
	if field, err = readNative(ctx, first, 3); err != nil {
		return err
	}
	if _, err = waitNative(ctx, first, "日本"); err != nil {
		return err
	}
	fmt.Printf("field_committed_native %s\n", mustJSON(field))
	// Existing managed TextField Change commits on Enter or FocusLost, not on
	// Cocoa unmark alone. Handoff back to the editor supplies that real boundary.
	if err = owner(a, func() error {
		if first.search != "" || first.searchChanges != 0 {
			return errors.New("TextField unmark prematurely committed controlled Change")
		}
		return nil
	}); err != nil {
		return err
	}
	stage("field-return-to-editor")
	back, err := focus(ctx, a, first, base)
	if err != nil {
		return err
	}
	if back.Client != initial.Client || back.Client == field.Client {
		return errors.New("focus handoff failed to restore multiline incarnation")
	}
	stage("field-controlled-change")
	var fieldSearch, fieldStatus string
	var fieldChanges uint64
	var fieldModel editor.Snapshot
	if err = waitOwner(ctx, a, func() (bool, error) {
		fieldSearch, fieldChanges, fieldStatus = first.search, first.searchChanges, first.status
		fieldModel = first.model.Snapshot()
		return fieldSearch == "日本" && fieldChanges == 1, nil
	}); err != nil {
		return fmt.Errorf("controlled field commit wait: search %q changes %d status %q editor %+v: %w", fieldSearch, fieldChanges, fieldStatus, fieldModel, err)
	}
	if err = owner(a, func() error {
		if first.model.Snapshot() != before {
			return errors.New("controlled field Change edited multiline journal")
		}
		return nil
	}); err != nil {
		return err
	}
	stage("composition-one-history-entry")
	if _, err = readNative(ctx, first, 1); err != nil {
		return err
	}
	if _, err = waitText(ctx, a, first, compositionPrefix+"に\r\n\r\n", true); err != nil {
		return err
	}
	if _, err = readNative(ctx, first, 2); err != nil {
		return err
	}
	preview, err := waitText(ctx, a, first, compositionText, true)
	if err != nil {
		return err
	}
	if preview.Text != base || preview.History != before.History || preview.Revision != before.Revision {
		return errors.New("preedit changed committed bytes/history")
	}
	if _, err = readNative(ctx, first, 3); err != nil {
		return err
	}
	committed, err := waitText(ctx, a, first, compositionText, false)
	if err != nil {
		return err
	}
	if committed.History.UndoCount != before.History.UndoCount+1 || committed.VisibleEpoch != preview.VisibleEpoch || committed.Revision != before.Revision+1 {
		return fmt.Errorf("no-visible-delta commit not atomic: before %+v preview %+v committed %+v", before, preview, committed)
	}
	if err = press(ctx, first, "undo"); err != nil {
		return err
	}
	if _, err = waitText(ctx, a, first, base, false); err != nil {
		return err
	}
	if err = press(ctx, first, "redo"); err != nil {
		return err
	}
	if _, err = waitText(ctx, a, first, compositionText, false); err != nil {
		return err
	}
	if _, err = waitNative(ctx, first, compositionText); err != nil {
		return err
	}
	stage("cancel-preedit")
	prior, err := capture(ctx, a, first)
	if err != nil {
		return err
	}
	// Step 2 selected its marked text; restore an insertion caret explicitly.
	if err = owner(a, func() error {
		n := utf8.RuneCountInString(prior.Text)
		if err := first.model.Select(editor.Selection{Anchor: n, Focus: n}); err != nil {
			return err
		}
		first.invalidate()
		return nil
	}); err != nil {
		return err
	}
	if _, err = waitNativeWhere(ctx, first, func(s nativeInput) bool {
		return s.Text == prior.Text && s.SelectedStart == utf16len(prior.Text) && s.SelectedLength == 0
	}); err != nil {
		return err
	}
	if _, err = readNative(ctx, first, 4); err != nil {
		return err
	}
	if _, err = waitText(ctx, a, first, prior.Text+"取消", true); err != nil {
		return err
	}
	if _, err = readNative(ctx, first, 5); err != nil {
		return err
	}
	afterCancel, err := waitText(ctx, a, first, prior.Text, false)
	if err != nil {
		return err
	}
	if afterCancel.History != prior.History || afterCancel.Revision != prior.Revision {
		return errors.New("cancel changed history")
	}
	stage("settled-save-exact-bytes")
	// Explicit replacementRange {1,2} addresses the supplementary emoji in UTF16.
	if _, err = readNative(ctx, first, 6); err != nil {
		return err
	}
	expected := strings.Replace(prior.Text, "😀", "中文", 1)
	marked, err := waitText(ctx, a, first, expected, true)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "vellum-editor-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "document.txt")
	var completion uint64
	if err = owner(a, func() error { completion = first.resultCount + 1; first.start(actionSave, path); return nil }); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return first.resultCount >= completion, nil }); err != nil {
		return err
	}
	saved, err := capture(ctx, a, first)
	if err != nil {
		return err
	}
	if saved.Composing || saved.Text != expected || saved.History.Dirty || saved.History.UndoCount != marked.History.UndoCount+1 {
		return fmt.Errorf("save did not settle native unmark: %+v", saved)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(bytes) != expected {
		return fmt.Errorf("saved bytes differ: %q", bytes)
	}
	stage("open-install-native-refresh")
	if err = owner(a, func() error {
		if err := first.model.Replace("discard this owner edit", editor.Selection{}); err != nil {
			return err
		}
		first.invalidate()
		completion = first.resultCount + 1
		first.start(actionOpen, path)
		return nil
	}); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return first.prompt == "open", nil }); err != nil {
		return err
	}
	if err = press(ctx, first, "discard-open"); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return first.resultCount >= completion, nil }); err != nil {
		return err
	}
	installed, err := capture(ctx, a, first)
	if err != nil {
		return err
	}
	if installed.Text != expected || installed.History.Dirty || installed.History.UndoCount != 0 || installed.History.RedoCount != 0 {
		return fmt.Errorf("open did not install exact clean sole-journal document: %+v", installed)
	}
	if _, err = waitNative(ctx, first, expected); err != nil {
		return err
	}
	stage("bounded-native-refusal")
	other, err := capture(ctx, a, second)
	if err != nil {
		return err
	}
	if _, err = focus(ctx, a, second, other.Text); err != nil {
		return err
	}
	refused, err := readNative(ctx, second, 10)
	if err != nil {
		return err
	}
	if refused.Text != other.Text {
		return errors.New("native configured budget mutated text before refusal")
	}
	if err = settle(ctx, a, second); err != nil {
		return err
	}
	otherAfter, err := capture(ctx, a, second)
	if err != nil {
		return err
	}
	if otherAfter != other {
		return errors.New("refused native edit changed second model/history")
	}
	stage("owner-edit-remount-and-focus-handoff")
	// Replace is an external owner edit, including bytes outside the most recent
	// native delta. It must become the published native authority without a second
	// history stack. Root adapter tests deterministically hold stale delta delivery.
	programText := "owner replacement 😀 אבג\r\n\r\n"
	if err = owner(a, func() error {
		if err := first.model.Replace(programText, editor.Selection{}); err != nil {
			return err
		}
		first.editorID = "document-remounted"
		first.focusID = first.editorID
		first.focusSequence++
		first.invalidate()
		return nil
	}); err != nil {
		return err
	}
	remounted, err := waitNative(ctx, first, programText)
	if err != nil {
		return err
	}
	if remounted.Client == initial.Client {
		return errors.New("remount reused native client incarnation")
	}
	if err = settle(ctx, a, first); err != nil {
		return err
	}
	independent, err := capture(ctx, a, second)
	if err != nil {
		return err
	}
	if independent != otherAfter {
		return errors.New("first document changed second window model")
	}
	stage("native-range-geometry-wrap-bidi")
	prefix := "A😀e\u0301 abc אבג العربية "
	geometryText := prefix + strings.Repeat("wrap ", 80) + "\r\nlast\r\n\r\n"
	if err = owner(a, func() error {
		if err := first.model.Replace(geometryText, editor.Selection{}); err != nil {
			return err
		}
		first.invalidate()
		return nil
	}); err != nil {
		return err
	}
	geometry, err := focus(ctx, a, first, geometryText)
	if err != nil {
		return err
	}
	emoji, err := waitQuery(ctx, first, 1, 2, func(q nativeQuery) bool { return q.Substring == "😀" && validRect(q.Rect) })
	if err != nil {
		return err
	}
	if emoji.Substring != "😀" || emoji.SubstringRange != [2]int{1, 2} || !validRect(emoji.Rect) {
		return fmt.Errorf("supplementary native range geometry: %+v", emoji)
	}
	point, err := query(ctx, first, 1, 2, [2]float64{emoji.Rect[0] + emoji.Rect[2]/2, emoji.Rect[1] + emoji.Rect[3]/2})
	if err != nil {
		return err
	}
	if point.Index < 1 || point.Index > 3 || point.Index == 2 {
		return fmt.Errorf("native point splits supplementary scalar: %+v", point)
	}
	rtlStart := utf16len("A😀e\u0301 abc ")
	rtl, err := query(ctx, first, rtlStart, 3, [2]float64{})
	if err != nil {
		return err
	}
	if rtl.Substring != "אבג" || !validRect(rtl.Rect) {
		return fmt.Errorf("bidi native range: %+v", rtl)
	}
	rtlFirst, err := query(ctx, first, rtlStart, 1, [2]float64{})
	if err != nil {
		return err
	}
	rtlSecond, err := query(ctx, first, rtlStart+1, 1, [2]float64{})
	if err != nil {
		return err
	}
	if !validRect(rtlFirst.Rect) || !validRect(rtlSecond.Rect) || rtlFirst.Rect[0] <= rtlSecond.Rect[0] {
		return fmt.Errorf("Hebrew visual cell order is not RTL: first %+v second %+v", rtlFirst, rtlSecond)
	}
	wrapStart := utf16len(prefix)
	wrapped, err := query(ctx, first, wrapStart, utf16len(strings.Repeat("wrap ", 80)), [2]float64{})
	if err != nil {
		return err
	}
	if wrapped.RectRange[0] < 0 || wrapped.RectRange[1] <= 0 || wrapped.RectRange[1] >= 400 || !validRect(wrapped.Rect) {
		return fmt.Errorf("wrap fragment range: %+v", wrapped)
	}
	next, err := query(ctx, first, wrapped.RectRange[0]+wrapped.RectRange[1], 1, [2]float64{})
	if err != nil {
		return err
	}
	if !validRect(next.Rect) || math.Abs(next.Rect[1]-wrapped.Rect[1]) < wrapped.Rect[3]/2 {
		return fmt.Errorf("wrapped next-line geometry unchanged: first %+v next %+v", wrapped, next)
	}
	fmt.Printf("native_geometry %s\n", mustJSON(map[string]any{"state": geometry, "emoji": emoji, "point": point, "bidi": rtl, "bidiFirst": rtlFirst, "bidiSecond": rtlSecond, "wrap": wrapped, "next": next}))
	stage("save-close-native-preedit-during-write")
	delayed, ok := first.store.(*smokeStore)
	if !ok {
		return errors.New("native delayed FileStore fixture missing")
	}
	if err = owner(a, func() error {
		n := utf8.RuneCountInString(geometryText)
		if err := first.model.Select(editor.Selection{Anchor: n, Focus: n}); err != nil {
			return err
		}
		first.invalidate()
		return nil
	}); err != nil {
		return err
	}
	if _, err = waitNativeWhere(ctx, first, func(s nativeInput) bool {
		return s.Text == geometryText && s.SelectedStart == utf16len(geometryText) && s.SelectedLength == 0
	}); err != nil {
		return err
	}
	gate := delayed.arm()
	if err = owner(a, func() error { completion = first.resultCount + 1; first.start(actionSaveClose, path); return nil }); err != nil {
		return err
	}
	select {
	case <-gate.entered:
	case <-ctx.Done():
		return ctx.Err()
	}
	if _, err = readNative(ctx, first, 1); err != nil {
		return err
	}
	if _, err = waitText(ctx, a, first, geometryText+"に", true); err != nil {
		return err
	}
	close(gate.release)
	if err = waitOwner(ctx, a, func() (bool, error) {
		s := first.model.Snapshot()
		return first.resultCount >= completion && !first.settling && !first.session.State().Busy && s.Text == geometryText+"に" && !s.Composing, nil
	}); err != nil {
		return err
	}
	if err = owner(a, func() error {
		if first.lifetime.Err() != nil || !first.session.State().Dirty {
			return errors.New("SaveClose lost preedit admitted during its write")
		}
		return nil
	}); err != nil {
		return err
	}
	older, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(older) != geometryText {
		return errors.New("delayed save did not write its captured exact bytes")
	}
	stage("save-close-refusal-and-terminal-teardown")
	if err = owner(a, func() error { first.start(actionClose, ""); return nil }); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return first.prompt == "close", nil }); err != nil {
		return err
	}
	if first.lifetime.Err() != nil {
		return errors.New("dirty native close accepted without decision")
	}
	if err = press(ctx, first, "keep"); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return first.prompt == "", nil }); err != nil {
		return err
	}
	if err = owner(a, func() error { second.start(actionDiscardClose, ""); return nil }); err != nil {
		return err
	}
	if err = waitOwner(ctx, a, func() (bool, error) { return second.lifetime.Err() != nil, nil }); err != nil {
		return err
	}
	if err = waitNativeClosed(ctx, secondNative); err != nil {
		return err
	}
	if first.lifetime.Err() != nil {
		return errors.New("closing second canceled first")
	}
	// The final asynchronous settle commits the dirty-close decision before last
	// registration cancellation; native disposal then proceeds outside the owner.
	if err = owner(a, func() error { first.start(actionDiscardClose, ""); return nil }); err != nil {
		return err
	}
	terminal, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-first.lifetime.Done():
	case <-terminal.Done():
		return terminal.Err()
	}
	if len(a.Windows()) != 0 || a.Context().Err() == nil {
		return errors.New("last-window domain lifetime not closed")
	}
	if err = waitNativeClosed(terminal, firstNative); err != nil {
		return err
	}
	fmt.Println("EDITOR_SMOKE_PASS: actual Cocoa protocol; atomic composition/history; exact UTF8/CRLF; bounded native refusal; independent windows; remount; range/wrap/bidi; settled save/close; native teardown")
	return nil
}
func mustJSON(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("JSON error: %v", err)
	}
	return string(b)
}
func waitNativeClosed(ctx context.Context, native system.Window) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for !native.IsClosed() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	// Acquiring/rechecking this native mutex waits for the compositor teardown
	// callback. The process watchdog bounds Lock itself; ctx bounds polling only.
	if native.Lock() {
		native.Unlock()
		return errors.New("closed native window accepted render lock")
	}
	return nil
}
