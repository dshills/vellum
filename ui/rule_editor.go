package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"
)

type RuleValueKind string

const (
	RuleText    RuleValueKind = "text"
	RuleNumber  RuleValueKind = "number"
	RuleBoolean RuleValueKind = "boolean"
)

type RuleOperator string

const (
	RuleEquals     RuleOperator = "equals"
	RuleContains   RuleOperator = "contains"
	RuleStartsWith RuleOperator = "starts-with"
	RuleLess       RuleOperator = "less"
	RuleGreater    RuleOperator = "greater"
	RuleIs         RuleOperator = "is"
)

type RuleMatch string

const (
	RuleAll RuleMatch = "all"
	RuleAny RuleMatch = "any"
)
const (
	MaxRules     = 64
	MaxRuleDepth = 3
)

type RuleField struct {
	ID, Label string
	Kind      RuleValueKind
}
type RuleValue struct {
	Text    string
	Number  float32
	Checked bool
}

// Rule is either a typed leaf or a group, never both. IDs are globally unique.
type Rule struct {
	ID, FieldID string
	Operator    RuleOperator
	Value       RuleValue
	Group       *RuleGroup
}
type RuleGroup struct {
	Match RuleMatch
	Rules []Rule
}
type RuleEditEvent struct {
	ID, RuleID string
	Root       Rule
}

// RuleEditor proposes a validated detached Root. Accept by assigning Root in
// Change; rejection leaves the current model intact. No predicate evaluation,
// arbitrary code or external data access is performed. Retain on one owner;
// do not copy or mutate concurrently after first use.
type RuleEditor struct {
	nextID uint64 // monotonic proposal identity; rejected additions also consume an ID

	ID                           string
	Fields                       []RuleField
	Root                         Rule
	Disabled                     bool
	Change                       func(RuleEditEvent)
	AddIconImage, GroupIconImage *canvas.TiledImage
	Style                        Style
}

func RuleControlID(root, rule, part string) string {
	return root + "/rules/" + base64.RawURLEncoding.EncodeToString([]byte(rule)) + "/" + part
}
func ruleOperators(kind RuleValueKind) []Option {
	switch kind {
	case RuleText:
		return []Option{{ID: string(RuleContains), Label: "contains"}, {ID: string(RuleEquals), Label: "is"}, {ID: string(RuleStartsWith), Label: "begins with"}}
	case RuleNumber:
		return []Option{{ID: string(RuleEquals), Label: "is"}, {ID: string(RuleLess), Label: "is less than"}, {ID: string(RuleGreater), Label: "is greater than"}}
	case RuleBoolean:
		return []Option{{ID: string(RuleIs), Label: "is"}}
	}
	return nil
}
func ruleField(fields []RuleField, id string) (RuleField, bool) {
	for _, f := range fields {
		if f.ID == id {
			return f, true
		}
	}
	return RuleField{}, false
}
func (e *RuleEditor) Validate() error {
	if e == nil || !validTokenText(e.ID) || len(e.Fields) == 0 || len(e.Fields) > 16 || e.Root.Group == nil {
		return fmt.Errorf("ui: rule editor needs ID, 1..16 fields and root group")
	}
	fields := map[string]bool{}
	for _, f := range e.Fields {
		if !validTokenText(f.ID) || !validTokenText(f.Label) || fields[f.ID] || len(ruleOperators(f.Kind)) == 0 {
			return fmt.Errorf("ui: invalid rule field")
		}
		fields[f.ID] = true
	}
	seen := map[string]bool{}
	count := 0
	var walk func(Rule, int) error
	walk = func(r Rule, depth int) error {
		count++
		if count > MaxRules || depth > MaxRuleDepth || !validTokenText(r.ID) || seen[r.ID] {
			return fmt.Errorf("ui: invalid, duplicate or excessive rule IDs/depth")
		}
		seen[r.ID] = true
		if r.Group != nil {
			if r.FieldID != "" || r.Operator != "" || r.Value != (RuleValue{}) || (r.Group.Match != RuleAll && r.Group.Match != RuleAny) {
				return fmt.Errorf("ui: invalid rule group")
			}
			for _, child := range r.Group.Rules {
				if err := walk(child, depth+1); err != nil {
					return err
				}
			}
			return nil
		}
		field, ok := ruleField(e.Fields, r.FieldID)
		if !ok {
			return fmt.Errorf("ui: rule field absent")
		}
		operator := false
		for _, o := range ruleOperators(field.Kind) {
			operator = operator || o.ID == string(r.Operator)
		}
		if !operator {
			return fmt.Errorf("ui: incompatible rule operator")
		}
		switch field.Kind {
		case RuleText:
			if len(r.Value.Text) > 4096 || !utf8.ValidString(r.Value.Text) || strings.ContainsRune(r.Value.Text, 0) || r.Value.Number != 0 || r.Value.Checked {
				return fmt.Errorf("ui: invalid text rule value")
			}
		case RuleNumber:
			if !finite(r.Value.Number) || r.Value.Number < -1e6 || r.Value.Number > 1e6 || r.Value.Text != "" || r.Value.Checked {
				return fmt.Errorf("ui: number rule outside finite +/-1e6 range")
			}
		case RuleBoolean:
			if r.Value.Text != "" || r.Value.Number != 0 {
				return fmt.Errorf("ui: invalid boolean rule value")
			}
		}
		return nil
	}
	return walk(e.Root, 1)
}
func cloneRule(r Rule) Rule {
	if r.Group != nil {
		g := *r.Group
		g.Rules = slices.Clone(g.Rules)
		for i := range g.Rules {
			g.Rules[i] = cloneRule(g.Rules[i])
		}
		r.Group = &g
	}
	return r
}
func locateRule(root *Rule, id, parent string, depth int) (*Rule, string, int) {
	if root.ID == id {
		return root, parent, depth
	}
	if root.Group != nil {
		for i := range root.Group.Rules {
			if r, p, d := locateRule(&root.Group.Rules[i], id, root.ID, depth+1); r != nil {
				return r, p, d
			}
		}
	}
	return nil, "", 0
}
func (e *RuleEditor) freshRuleID(root *Rule) string {
	for i := 0; i <= MaxRules; i++ {
		if e.nextID == ^uint64(0) {
			return ""
		}
		e.nextID++
		id := fmt.Sprintf("rule-%d", e.nextID)
		r, _, _ := locateRule(root, id, "", 1)
		if r == nil {
			return id
		}
	}
	return ""
}

func countRules(r Rule) int {
	n := 1
	if r.Group != nil {
		for _, c := range r.Group.Rules {
			n += countRules(c)
		}
	}
	return n
}
func (e *RuleEditor) View() (Node, error) {
	if err := e.Validate(); err != nil {
		return Node{}, err
	}
	id := e.ID
	shown := cloneRule(e.Root)
	fields := slices.Clone(e.Fields)
	edit := func(target string, change func(*Rule, *Rule, int) bool) {
		if e.ID != id || e.Disabled || e.Change == nil || e.Validate() != nil || !slices.Equal(fields, e.Fields) {
			return
		}
		old, oldParent, _ := locateRule(&shown, target, "", 1)
		current, parent, depth := locateRule(&e.Root, target, "", 1)
		if old == nil || current == nil || oldParent != parent || (old.Group == nil) != (current.Group == nil) {
			return
		}
		candidate := cloneRule(e.Root)
		node, _, _ := locateRule(&candidate, target, "", 1)
		if !change(node, &candidate, depth) || reflect.DeepEqual(candidate, e.Root) {
			return
		}
		check := *e
		check.Root = candidate
		if check.Validate() != nil {
			return
		}
		e.Change(RuleEditEvent{ID: id, RuleID: target, Root: candidate})
	}
	fieldsOptions := []Option{}
	for _, f := range fields {
		fieldsOptions = append(fieldsOptions, Option{ID: f.ID, Label: f.Label})
	}
	compact := Style{MinSize: Some(Size{Width: 116, Height: 28}), MaxSize: Some(Size{Width: 150}), GrowX: Some(float32(0)), Padding: Some(Insets{Top: 3, Bottom: 3, Left: 6, Right: 6})}
	var build func(Rule, int, bool) Node
	build = func(r Rule, depth int, isRoot bool) Node {
		key := func(part string) string { return RuleControlID(id, r.ID, part) }
		remove := Node{ID: key("remove"), Kind: Button, Text: "−", Accessibility: Accessibility{Name: "Remove rule " + r.ID}, Disabled: isRoot, Style: Style{MinSize: Some(Size{Width: 28, Height: 28}), MaxSize: Some(Size{Width: 28, Height: 28}), Padding: Some(UniformInsets(2))}, Activate: func() {
			edit(r.ID, func(node, root *Rule, _ int) bool {
				_, parent, _ := locateRule(root, node.ID, "", 1)
				p, _, _ := locateRule(root, parent, "", 1)
				if p == nil || p.Group == nil {
					return false
				}
				for i, c := range p.Group.Rules {
					if c.ID == node.ID {
						p.Group.Rules = slices.Delete(p.Group.Rules, i, i+1)
						return true
					}
				}
				return false
			})
		}}
		if r.Group != nil {
			match := Node{ID: key("match"), Kind: Selector, Options: []Option{{ID: "all", Label: "All conditions"}, {ID: "any", Label: "Any condition"}}, SelectedID: string(r.Group.Match), Accessibility: Accessibility{Name: "Match mode for " + r.ID}, Style: compact, Change: func(c ChangeEvent) {
				edit(r.ID, func(node, _ *Rule, _ int) bool {
					if c.SelectedID != "all" && c.SelectedID != "any" {
						return false
					}
					node.Group.Match = RuleMatch(c.SelectedID)
					return true
				})
			}}
			add := func(group bool) Node {
				label, part, art := "Add Rule", "add-rule", e.AddIconImage
				if group {
					label, part, art = "Add Group", "add-group", e.GroupIconImage
				}
				disabled := countRules(shown) >= MaxRules || depth >= MaxRuleDepth
				if group {
					disabled = disabled || depth >= MaxRuleDepth-1
				}
				return Node{ID: key(part), Kind: Button, Text: label, IconImage: art, Disabled: disabled, Style: Style{MinSize: Some(Size{Height: 28}), MaxSize: Some(Size{Height: 28}), GrowX: Some(float32(0)), Padding: Some(Insets{Top: 2, Bottom: 2, Left: 6, Right: 6})}, Activate: func() {
					edit(r.ID, func(node, root *Rule, level int) bool {
						if countRules(*root) >= MaxRules || level >= MaxRuleDepth || group && level >= MaxRuleDepth-1 {
							return false
						}
						newID := e.freshRuleID(root)
						if newID == "" {
							return false
						}
						child := Rule{ID: newID, FieldID: e.Fields[0].ID, Operator: RuleOperator(ruleOperators(e.Fields[0].Kind)[0].ID)}
						if group {
							child.FieldID = ""
							child.Operator = ""
							child.Group = &RuleGroup{Match: RuleAll}
						}
						node.Group.Rules = append(node.Group.Rules, child)
						return true
					})
				}}
			}
			header := Node{ID: key("header"), Kind: Row, Children: []Node{match, add(false), add(true)}}
			if !isRoot {
				header.Children = append(header.Children, remove)
			}
			children := []Node{header}
			for _, child := range r.Group.Rules {
				children = append(children, build(child, depth+1, false))
			}
			return Node{ID: key("group"), Kind: FormGroup, Text: "", Style: Style{Padding: Some(UniformInsets(8)), Gap: Some(float32(6))}, Children: children}
		}
		field, _ := ruleField(fields, r.FieldID)
		criterion := Node{ID: key("field"), Kind: Selector, Options: fieldsOptions, SelectedID: r.FieldID, Style: compact, Accessibility: Accessibility{Name: "Criterion for " + r.ID}, Change: func(c ChangeEvent) {
			edit(r.ID, func(node, _ *Rule, _ int) bool {
				f, ok := ruleField(e.Fields, c.SelectedID)
				if !ok {
					return false
				}
				if node.FieldID == f.ID {
					return false
				}
				node.FieldID = f.ID
				node.Operator = RuleOperator(ruleOperators(f.Kind)[0].ID)
				node.Value = RuleValue{}
				return true
			})
		}}
		operator := Node{ID: key("operator"), Kind: Selector, Options: ruleOperators(field.Kind), SelectedID: string(r.Operator), Style: compact, Accessibility: Accessibility{Name: "Operator for " + r.ID}, Change: func(c ChangeEvent) {
			edit(r.ID, func(node, _ *Rule, _ int) bool {
				if node.FieldID != r.FieldID {
					return false
				}
				for _, o := range ruleOperators(field.Kind) {
					if o.ID == c.SelectedID {
						node.Operator = RuleOperator(c.SelectedID)
						return true
					}
				}
				return false
			})
		}}
		value := Node{ID: key("value"), Kind: TextField, Text: r.Value.Text, Style: compact, Accessibility: Accessibility{Name: "Value for " + r.ID}}
		switch field.Kind {
		case RuleNumber:
			value.Kind = NumberField
			value.Text = ""
			value.Number = r.Value.Number
			value.Numeric = Numeric{Min: -1e6, Max: 1e6, Step: 1}
			value.NumberButtons = NumberButtonsStepper
		case RuleBoolean:
			value.Kind = Checkbox
			value.Text = "True"
			value.Checked = r.Value.Checked
		}
		value.Change = func(c ChangeEvent) {
			edit(r.ID, func(node, _ *Rule, _ int) bool {
				if node.FieldID != r.FieldID || node.Operator != r.Operator {
					return false
				}
				switch field.Kind {
				case RuleText:
					node.Value.Text = c.Text
				case RuleNumber:
					node.Value.Number = c.Number
				case RuleBoolean:
					node.Value.Checked = c.Checked
				}
				return true
			})
		}
		return Node{ID: key("row"), Kind: Row, Style: Style{Gap: Some(float32(6))}, Children: []Node{criterion, operator, value, remove}}
	}
	root := Node{ID: id, Kind: Scroll, Disabled: e.Disabled || e.Change == nil, Accessibility: Accessibility{Name: "Filter rules"}, Style: Style{MinSize: Some(Size{Width: 320, Height: 160}), MaxSize: Some(Size{Height: 400}), Padding: Some(Insets{}), GrowX: Some(float32(1))}.Overlay(e.Style), Children: []Node{build(shown, 1, true)}}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
