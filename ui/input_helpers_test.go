package ui

import (
	"image/color"
	"reflect"
	"strings"
	"testing"
)

func helperNode(t *testing.T, root Node, id string) Node {
	t.Helper()
	if root.ID == id {
		return root
	}
	for _, c := range root.Children {
		if n, ok := helperFind(c, id); ok {
			return n
		}
	}
	t.Fatal("missing node", id)
	return Node{}
}
func helperFind(root Node, id string) (Node, bool) {
	if root.ID == id {
		return root, true
	}
	for _, c := range root.Children {
		if n, ok := helperFind(c, id); ok {
			return n, true
		}
	}
	return Node{}, false
}
func TestSearchBoxControlledClearAndCurrentDelivery(t *testing.T) {
	s := &SearchBox{ID: "search", Label: "Find objects", Text: "Cube"}
	calls := 0
	s.Change = func(e SearchEvent) {
		calls++
		if e.ID != "search" {
			t.Fatal(e)
		}
	}
	n, err := s.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, SearchInputID(s.ID))
	clear := helperNode(t, n, SearchClearID(s.ID))
	field.Change(ChangeEvent{Text: "日本語"})
	clear.Activate()
	if calls != 2 || s.Text != "Cube" {
		t.Fatal("proposals mutated accepted text", s.Text, calls)
	}
	// A previously queued callback consults the current handler and accepted model.
	s.Change = func(e SearchEvent) { s.Text = e.Text }
	field.Change(ChangeEvent{Text: "Café"})
	if s.Text != "Café" {
		t.Fatal(s.Text)
	}
	clear.Activate()
	if s.Text != "" {
		t.Fatal(s.Text)
	}
	n, err = s.View()
	if err != nil {
		t.Fatal(err)
	}
	if !helperNode(t, n, SearchClearID(s.ID)).Disabled {
		t.Fatal("empty clear enabled")
	}
	if got := FocusOrder(n); !reflect.DeepEqual(got, []string{SearchInputID(s.ID)}) {
		t.Fatal(got)
	}
	s.Text = "keep"
	s.Disabled = true
	clear.Activate()
	field.Change(ChangeEvent{Text: "bad"})
	if s.Text != "keep" {
		t.Fatal("disabled request admitted")
	}
	s.Disabled = false
	s.ID = "replacement"
	field.Change(ChangeEvent{Text: "bad"})
	clear.Activate()
	if s.Text != "keep" {
		t.Fatal("old component identity admitted")
	}
	s.ID = "search"
	s.Change = nil
	clear.Activate()
	n, err = s.View()
	if err != nil || len(FocusOrder(n)) != 0 {
		t.Fatal("readonly focus", n, err)
	}
}
func TestInputHelpersValidationAndCollisions(t *testing.T) {
	var nilSearch *SearchBox
	if _, err := nilSearch.View(); err == nil {
		t.Fatal("nil search")
	}
	if _, err := (&SearchBox{}).View(); err == nil {
		t.Fatal("empty search ID")
	}
	s := &SearchBox{ID: "s", Style: Style{Gap: Some(float32(-1))}}
	if _, err := s.View(); err == nil {
		t.Fatal("invalid search style")
	}
	f := FormField{ID: "f", Label: "Name", Content: Node{ID: "name", Kind: TextField}}
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	root := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: "f/field/error", Kind: Label}}}
	f.ValidationMessage = "invalid"
	n, err = f.View()
	if err != nil {
		t.Fatal(err)
	}
	root.Children[0] = n
	if Validate(root) == nil {
		t.Fatal("generated collision accepted")
	}
	for _, kind := range []Kind{Label, Row, Tabs, SplitPane, Disclosure} {
		f.Content.Kind = kind
		if _, err := f.View(); err == nil {
			t.Fatal("non-leaf input accepted", kind)
		}
	}
}
func TestFormFieldAccessibleValidationWithoutMutation(t *testing.T) {
	f := FormField{ID: "name-field", Label: "Name", Required: true, Help: "Use a unique name", ValidationMessage: "Name is already used",
		Content: Node{ID: "name", Kind: TextField, Text: "Cube", Accessibility: Accessibility{Help: "Original help"}}}
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, "name")
	if field.Accessibility.Name != "Name" || !strings.Contains(field.Accessibility.Help, "Required") || !strings.Contains(field.Accessibility.Help, "Error: Name is already used") {
		t.Fatal(field.Accessibility)
	}
	if f.Content.Accessibility.Name != "" || f.Content.Accessibility.Help != "Original help" {
		t.Fatal("caller content mutated")
	}
	if order := FocusOrder(n); !reflect.DeepEqual(order, []string{"name"}) {
		t.Fatal(order)
	}
	f.Disabled = true
	n, err = f.View()
	if err != nil || len(FocusOrder(n)) != 0 {
		t.Fatal("disabled ancestry", err)
	}
	f.Content.Accessibility.Name = "Custom name"
	n, err = f.View()
	if err != nil {
		t.Fatal(err)
	}
	if helperNode(t, n, "name").Accessibility.Name != "Custom name" {
		t.Fatal("explicit name overwritten")
	}
}
func TestInlineNoticeActionsAndPassiveScope(t *testing.T) {
	n := &InlineNotice{ID: "message", Text: "<b>Literal & safe</b>", Tone: NoticeWarning}
	view, err := n.View()
	if err != nil {
		t.Fatal(err)
	}
	if len(FocusOrder(view)) != 0 || helperNode(t, view, "message/notice/text").Text != n.Text {
		t.Fatal("passive notice")
	}
	calls := 0
	n.ActionLabel = "Retry"
	n.Activate = func() { calls++ }
	view, err = n.View()
	if err != nil {
		t.Fatal(err)
	}
	action := helperNode(t, view, "message/notice/action")
	action.Activate()
	if calls != 1 {
		t.Fatal(calls)
	}
	n.Activate = func() { calls += 10 }
	action.Activate()
	if calls != 11 {
		t.Fatal("old action handler")
	}
	n.Disabled = true
	action.Activate()
	n.Disabled = false
	n.ID = "new"
	action.Activate()
	if calls != 11 {
		t.Fatal("stale/disabled action admitted")
	}
	n.ID = "message"
	n.Activate = nil
	action.Activate()
	if _, err = n.View(); err == nil {
		t.Fatal("action label without callback")
	}
	n.ActionLabel = ""
	n.Tone = "unsupported"
	if _, err = n.View(); err == nil {
		t.Fatal("unknown tone")
	}
	for _, tone := range []NoticeTone{"", NoticeInformation, NoticeSuccess, NoticeWarning, NoticeError} {
		n.Tone = tone
		view, err = n.View()
		if err != nil || Validate(view) != nil {
			t.Fatal(tone, err)
		}
	}
}
func TestMacThemeIsLocalValidAndOverridable(t *testing.T) {
	for _, dark := range []bool{false, true} {
		theme := MacTheme(dark)
		if err := theme.Validate(); err != nil {
			t.Fatal(err)
		}
		local := MacTheme(dark)
		theme.Kinds[TextField] = Style{}
		if !local.Kinds[TextField].Border.Set {
			t.Fatal("shared mutable theme")
		}
		custom := color.NRGBA{R: 80, G: 120, B: 160, A: 255}
		s, err := ResolveStyle(Node{ID: "input", Kind: TextField, Style: Style{Foreground: Some(custom)}}, local)
		if err != nil || s.Foreground.Value != custom || s.Typography.Value.Size != 13 {
			t.Fatal(s, err)
		}
	}
	if MacTheme(false).Kinds[TextField].Background == MacTheme(true).Kinds[TextField].Background {
		t.Fatal("dark palette unchanged")
	}
}

func TestTextPlaceholderScope(t *testing.T) {
	n := Node{ID: "text", Kind: TextField, Placeholder: "Search 日本語"}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []Kind{Label, Button, NumberField, Column, MultilineText} {
		if err := ValidateControl(Node{ID: "bad", Kind: kind, Placeholder: "hint"}); err == nil {
			t.Fatal("placeholder accepted on unsupported kind", kind)
		}
	}
}

func TestPlaceholderRejectedThroughFullTreeValidation(t *testing.T) {
	for _, kind := range []Kind{Column, Row, Label, Button, FormGroup} {
		err := Validate(Node{ID: "unsupported", Kind: kind, Placeholder: "hint"})
		if err == nil || !strings.Contains(err.Error(), "Placeholder") {
			t.Fatal("full traversal missed placeholder", kind, err)
		}
	}
	nested := Node{ID: "root", Kind: Column, Children: []Node{{ID: "label", Kind: Label, Placeholder: "hint"}}}
	if err := Validate(nested); err == nil || !strings.Contains(err.Error(), "Placeholder") {
		t.Fatal("nested traversal missed placeholder", err)
	}
}

func TestNumberStepperPresentationScope(t *testing.T) {
	n := Node{ID: "copies", Kind: NumberField, Number: 1, Numeric: Numeric{Min: 0, Max: 3, Step: 1}, NumberButtons: NumberButtonsStepper}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	n.NumberButtons = "unknown"
	if Validate(n) == nil {
		t.Fatal("unknown numeric appearance")
	}
	n.Kind = TextField
	n.NumberButtons = NumberButtonsStepper
	if Validate(n) == nil {
		t.Fatal("numeric appearance on text field")
	}
}
