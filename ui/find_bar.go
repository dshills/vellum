package ui

import (
	"errors"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/editor"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrFindNoMatch = errors.New("ui: no matches")
var ErrFindLimit = errors.New("ui: refine search; document/query/match limit exceeded")

const MaxFindMatches = 10000

// FindOptions use literal Unicode scalar matching; case-insensitive search uses
// simple Unicode folding, not locale-specific expansion or regular expressions.
type FindOptions struct{ CaseSensitive, WholeWord bool }

func foldFind(r rune) rune {
	lowest := r
	for s := unicode.SimpleFold(r); s != r; s = unicode.SimpleFold(s) {
		if s < lowest {
			lowest = s
		}
	}
	return lowest
}
func FindMatches(text, query string, o FindOptions) ([]editor.Range, error) {
	if !utf8.ValidString(text) || !utf8.ValidString(query) || strings.ContainsRune(query, 0) {
		return nil, editor.ErrInvalidText
	}
	if len(text) > editor.DefaultMaxBytes || utf8.RuneCountInString(query) > 1024 {
		return nil, ErrFindLimit
	}
	if query == "" {
		return nil, nil
	}
	source, needle := []rune(text), []rune(query)
	if !o.CaseSensitive {
		for i := range needle {
			needle[i] = foldFind(needle[i])
		}
	}
	prefix := make([]int, len(needle))
	for i, j := 1, 0; i < len(needle); i++ {
		for j > 0 && needle[i] != needle[j] {
			j = prefix[j-1]
		}
		if needle[i] == needle[j] {
			j++
		}
		prefix[i] = j
	}
	word := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
	out := []editor.Range{}
	j := 0
	for i, ch := range source {
		if !o.CaseSensitive {
			ch = foldFind(ch)
		}
		for j > 0 && ch != needle[j] {
			j = prefix[j-1]
		}
		if ch == needle[j] {
			j++
		}
		if j == len(needle) {
			start, end := i+1-j, i+1
			if !o.WholeWord || (start == 0 || !word(source[start-1])) && (end == len(source) || !word(source[end])) {
				out = append(out, editor.Range{Start: start, End: end})
				if len(out) > MaxFindMatches {
					return nil, ErrFindLimit
				}
				j = 0
			} else {
				j = prefix[j-1]
			}
		}
	}
	return out, nil
}

type FindAction uint8

const (
	FindNext FindAction = iota
	FindPrevious
	FindReplace
	FindReplaceAll
)

type FindRequest struct {
	Action             FindAction
	Query, Replacement string
	Options            FindOptions
}

// Apply runs on the model owner. Replace All is one journal transaction; no
// document mutation occurs while composing or when the bounded search refuses.
func (q FindRequest) Apply(m *editor.Model) error {
	if m == nil {
		return editor.ErrInvalidModel
	}
	s := m.Snapshot()
	if s.Composing {
		return editor.ErrComposition
	}
	if q.Action > FindReplaceAll {
		return fmt.Errorf("ui: unknown find action")
	}
	if !utf8.ValidString(q.Replacement) || strings.ContainsRune(q.Replacement, 0) {
		return editor.ErrInvalidText
	}
	if len(q.Replacement) > 64<<10 {
		return ErrFindLimit
	}
	matches, err := FindMatches(s.Text, q.Query, q.Options)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return ErrFindNoMatch
	}
	lo, hi := s.Selection.Anchor, s.Selection.Focus
	if lo > hi {
		lo, hi = hi, lo
	}
	runes := []rune(s.Text)
	if q.Action == FindReplaceAll {
		size := int64(len(s.Text))
		for _, m := range matches {
			size += int64(len(q.Replacement) - len(string(runes[m.Start:m.End])))
		}
		if size > editor.DefaultMaxBytes {
			return editor.ErrTooLarge
		}
		var b strings.Builder
		b.Grow(int(size))
		from := 0
		for _, m := range matches {
			b.WriteString(string(runes[from:m.Start]))
			b.WriteString(q.Replacement)
			from = m.End
		}
		b.WriteString(string(runes[from:]))
		replacementRunes := utf8.RuneCountInString(q.Replacement)
		translate := func(pos int) int {
			delta := 0
			for _, match := range matches {
				if pos <= match.Start {
					break
				}
				if pos < match.End {
					return match.Start + delta + min(pos-match.Start, replacementRunes)
				}
				delta += replacementRunes - (match.End - match.Start)
			}
			return pos + delta
		}
		return m.Replace(b.String(), editor.Selection{Anchor: translate(s.Selection.Anchor), Focus: translate(s.Selection.Focus)})
	}
	if q.Action == FindReplace {
		for _, match := range matches {
			if match.Start == lo && match.End == hi {
				end := lo + utf8.RuneCountInString(q.Replacement)
				return m.Apply(editor.Edit{BaseRevision: s.Revision, BaseVisibleEpoch: s.VisibleEpoch, Range: match, Before: string(runes[lo:hi]), Insert: q.Replacement, Selection: editor.Selection{Anchor: lo, Focus: end}})
			}
		}
	}
	target := matches[0]
	if q.Action == FindPrevious {
		target = matches[len(matches)-1]
		for i := len(matches) - 1; i >= 0; i-- {
			if matches[i].End <= lo {
				target = matches[i]
				break
			}
		}
	} else {
		for _, match := range matches {
			if match.Start >= hi {
				target = match
				break
			}
		}
	}
	return m.Select(editor.Selection{Anchor: target.Start, Focus: target.End})
}

// FindBarChange edits one controlled bar field without replacing other values.
type FindBarChange struct {
	ID, Field, Text string
	Checked         bool
}
type FindBar struct {
	cacheModel                         *editor.Model
	cacheRevision                      uint64
	cacheQuery                         string
	cacheOptions                       FindOptions
	cacheMatches                       []editor.Range
	cacheError                         error
	cacheValid                         bool
	ID, Label, Query, Replacement      string
	Open, ReplaceOpen, Disabled, Small bool
	Options                            FindOptions
	Model                              *editor.Model
	IconImage                          *canvas.TiledImage
	Change                             func(FindBarChange)
	Request                            func(FindRequest)
	Close                              func()
	Style                              Style
}

func FindBarInputID(id string) string { return id + "/find/query" }
func (f *FindBar) View() (Node, error) {
	if f == nil || !validTokenText(f.ID) || f.Model == nil || f.Model.Validate() != nil || !utf8.ValidString(f.Query) || !utf8.ValidString(f.Replacement) || strings.ContainsRune(f.Query+f.Replacement, 0) || utf8.RuneCountInString(f.Query) > 1024 || len(f.Replacement) > 64<<10 {
		return Node{}, fmt.Errorf("ui: find bar requires bounded fields and editor model")
	}
	label := f.Label
	if label == "" {
		label = "Find and replace"
	}
	id := f.ID
	root := Node{ID: id, Kind: Column, Accessibility: Accessibility{Name: label}, Style: Style{Gap: Some(float32(6))}.Overlay(f.Style)}
	if !f.Open {
		f.cacheModel = nil
		f.cacheMatches = nil
		f.cacheValid = false
		return root, nil
	}
	live := func() bool { return f.ID == id && f.Open && !f.Disabled }
	change := func(field string, text string, checked bool) {
		if !utf8.ValidString(text) || strings.ContainsRune(text, 0) || (field == "query" && utf8.RuneCountInString(text) > 1024) || (field == "replacement" && len(text) > 64<<10) {
			return
		}
		if live() && f.Change != nil {
			f.Change(FindBarChange{ID: id, Field: field, Text: text, Checked: checked})
		}
	}
	snapshot := f.Model.Snapshot()
	if !f.cacheValid || f.cacheModel != f.Model || f.cacheRevision != snapshot.Revision || f.cacheQuery != f.Query || f.cacheOptions != f.Options {
		f.cacheMatches, f.cacheError = FindMatches(snapshot.Text, f.Query, f.Options)
		f.cacheModel = f.Model
		f.cacheRevision = snapshot.Revision
		f.cacheQuery = f.Query
		f.cacheOptions = f.Options
		f.cacheValid = true
	}
	matches, err := f.cacheMatches, f.cacheError
	count := fmt.Sprintf("%d matches", len(matches))
	if err != nil {
		count = "Refine search"
	}
	sel := snapshot.Selection
	for i, m := range matches {
		if (sel.Anchor == m.Start && sel.Focus == m.End) || (sel.Focus == m.Start && sel.Anchor == m.End) {
			count = fmt.Sprintf("%d of %d", i+1, len(matches))
		}
	}
	height := float32(30)
	if f.Small {
		height = 24
	}
	style := Style{MinSize: Some(Size{Height: height}), GrowX: Some(float32(0))}
	request := func(action FindAction) {
		if live() && f.Request != nil {
			f.Request(FindRequest{Action: action, Query: f.Query, Replacement: f.Replacement, Options: f.Options})
		}
	}
	action := func(suffix, label string, a FindAction) Node {
		return Node{ID: id + "/find/" + suffix, Kind: Button, Text: label, Appearance: ButtonTool, Style: style, Disabled: f.Disabled || f.Request == nil || len(matches) == 0 || err != nil || snapshot.Composing, Activate: func() { request(a) }}
	}
	row := Node{ID: id + "/find/row", Kind: Row, Children: []Node{{ID: FindBarInputID(id), Kind: TextField, Text: f.Query, Placeholder: "Find", Style: Style{MinSize: Some(Size{Height: height, Width: 130}), GrowX: Some(float32(1))}, Accessibility: Accessibility{Name: "Find text"}, Disabled: f.Disabled || f.Change == nil, Change: func(e ChangeEvent) { change("query", e.Text, false) }}, action("previous", "Previous", FindPrevious), action("next", "Next", FindNext), {ID: id + "/find/count", Kind: Label, Text: count}, {ID: id + "/find/close", Kind: Button, Text: "Done", Appearance: ButtonTool, Style: style, Disabled: f.Disabled || f.Close == nil, Activate: func() {
		if live() && f.Close != nil {
			f.Close()
		}
	}}}}
	if f.IconImage != nil {
		row.Children = append([]Node{{ID: id + "/find/art", Kind: Image, Image: f.IconImage, Text: "Find and replace", Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 20, Height: 20}), MaxSize: Some(Size{Width: 20, Height: 20})}}}, row.Children...)
	}
	root.Children = []Node{row, {ID: id + "/find/options", Kind: Row, Children: []Node{{ID: id + "/find/case", Kind: Checkbox, Text: "Match case", Checked: f.Options.CaseSensitive, Disabled: f.Disabled || f.Change == nil, Change: func(e ChangeEvent) { change("case", "", e.Checked) }}, {ID: id + "/find/word", Kind: Checkbox, Text: "Whole words", Checked: f.Options.WholeWord, Disabled: f.Disabled || f.Change == nil, Change: func(e ChangeEvent) { change("word", "", e.Checked) }}, {ID: id + "/find/replace-visible", Kind: Checkbox, Text: "Replace", Checked: f.ReplaceOpen, Disabled: f.Disabled || f.Change == nil, Change: func(e ChangeEvent) { change("replace-open", "", e.Checked) }}}}}
	if f.ReplaceOpen {
		root.Children = append(root.Children, Node{ID: id + "/find/replacement-row", Kind: Row, Children: []Node{{ID: id + "/find/replacement", Kind: TextField, Text: f.Replacement, Placeholder: "Replace with", Style: Style{MinSize: Some(Size{Width: 150, Height: height}), GrowX: Some(float32(1))}, Accessibility: Accessibility{Name: "Replacement text"}, Disabled: f.Disabled || f.Change == nil, Change: func(e ChangeEvent) { change("replacement", e.Text, false) }}, action("replace", "Replace", FindReplace), action("replace-all", "Replace All", FindReplaceAll)}})
	}
	return root, Validate(root)
}
