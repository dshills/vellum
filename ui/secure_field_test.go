package ui

import "testing"

func TestSecureFieldControlledRevealAndDisabled(t *testing.T) {
	p := &SecureField{ID: "password", Text: "synthetic", AllowReveal: true}
	calls := 0
	p.Change = func(e ChangeEvent) {
		calls++
		if e.ID != "password" {
			t.Fatal(e)
		}
	}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, SecureInputID(p.ID))
	if !field.Secure || field.Reveal {
		t.Fatal("not masked")
	}
	field.Change(ChangeEvent{Text: "new"})
	if calls != 1 || p.Text != "synthetic" {
		t.Fatal("not controlled")
	}
	helperNode(t, n, "password/secure/reveal").Activate()
	n, _ = p.View()
	field = helperNode(t, n, SecureInputID(p.ID))
	if !field.Reveal || !field.Secure {
		t.Fatal("reveal lost protection")
	}
	p.Disabled = true
	n, _ = p.View()
	field = helperNode(t, n, SecureInputID(p.ID))
	if field.Reveal {
		t.Fatal("disabled reveal retained")
	}
	field.Change(ChangeEvent{Text: "ignored"})
	if calls != 1 {
		t.Fatal(calls)
	}
	if Validate(Node{ID: "bad", Kind: Label, Secure: true}) == nil || Validate(Node{ID: "bad", Kind: TextField, Reveal: true}) == nil {
		t.Fatal("invalid secure flags")
	}
}
