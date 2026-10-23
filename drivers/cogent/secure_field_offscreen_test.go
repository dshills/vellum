//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestSecureFieldManagedPrivacyAndReveal(t *testing.T) {
	p := &ui.SecureField{ID: "password", Label: "Password", Text: "synthetic😀", AllowReveal: true}
	calls := 0
	p.Change = func(e ui.ChangeEvent) { calls++; p.Text = e.Text }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 925), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	r := findRecord(root, ui.SecureInputID(p.ID))
	field := r.widget.(*core.TextField)
	if !field.NoEcho {
		t.Fatal("clear glyphs")
	}
	check := func() {
		t.Helper()
		nodes, _ := accessibilitySnapshot(root, nil)
		found := false
		for _, n := range nodes {
			if n.DomainID == r.node.ID {
				found = true
				if !n.Secure || n.Value != "" || n.TextClient != 0 || n.TextSelection != nil {
					t.Fatal("AX exposure")
				}
				for _, action := range n.Actions {
					if action == "selection" {
						t.Fatal("selection action exposed")
					}
				}
			}
		}
		if !found {
			t.Fatal("missing secure AX node")
		}
		if _, ok := accessibilityTextGeometry(root)[accessibilityManagedID(r.node.ID)]; ok {
			t.Fatal("geometry exposure")
		}
		semantic, err := accessibilitySemantic(nodes)
		if err != nil || strings.Contains(semantic, p.Text) || !strings.Contains(semantic, `"secure":true`) {
			t.Fatal("wire protection", err)
		}
		_, element := imageManualElement(t, m, r.node.ID)
		if element.Value != "" {
			t.Fatal("manual snapshot exposure")
		}
	}
	check()
	// Real managed text commit reaches the owner; semantics remain protected.
	field.SetText("replacement").SendChange()
	a.Drain()
	unitSettle(t, m)
	if calls != 1 || p.Text != "replacement" {
		t.Fatal("commit")
	}
	root = m.implementation.find("one").state.root
	check()
	findRecord(root, "password/secure/reveal").widget.(*core.Button).Send(events.Click)
	a.Drain()
	unitSettle(t, m)
	root = m.implementation.find("one").state.root
	if field.NoEcho {
		t.Fatal("explicit reveal failed")
	}
	check()
	if err := a.Do(func() { p.Disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !field.NoEcho {
		t.Fatal("disable failed to remask")
	}
}

func TestSecureFieldRawNodeWireProjections(t *testing.T) {
	for _, reveal := range []bool{false, true} {
		name := "masked"
		if reveal {
			name = "revealed"
		}
		t.Run(name, func(t *testing.T) {
			view := func() ui.Node {
				return ui.Node{ID: "secure", Kind: ui.TextField, Text: "fixture-private-text", Secure: true, Reveal: reveal, Change: func(ui.ChangeEvent) {}}
			}
			m, _ := simpleManual(t, manualConfig(t, 927), view)
			unitSettle(t, m)
			root := m.implementation.find("one").state.root
			nodes, _ := accessibilitySnapshot(root, nil)
			if len(nodes) != 1 || nodes[0].Name != "Password" || !nodes[0].Secure {
				t.Fatal("default label protection")
			}
			reflected, err := json.Marshal(nodes)
			if err != nil {
				t.Fatal(err)
			}
			semantic, err := accessibilitySemantic(nodes)
			if err != nil {
				t.Fatal(err)
			}
			for _, wire := range []string{string(reflected), semantic, accessibilityPayload(1, semantic)} {
				if strings.Contains(wire, "fixture-private-text") {
					t.Fatal("protected value reached wire")
				}
			}
			if len(accessibilityTextGeometry(root)) != 0 {
				t.Fatal("text geometry exposure")
			}
			_, e := imageManualElement(t, m, "secure")
			if e.Value != "" {
				t.Fatal("manual exposure")
			}
		})
	}
}
