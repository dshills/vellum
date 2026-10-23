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

func TestSecureFieldIdentityChangeRequiresNewReveal(t *testing.T) {
	p := &SecureField{ID: "first", Text: "first-synthetic", AllowReveal: true, Change: func(ChangeEvent) {}}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	oldReveal := helperNode(t, n, "first/secure/reveal")
	oldReveal.Activate()
	n, _ = p.View()
	if !helperNode(t, n, SecureInputID("first")).Reveal {
		t.Fatal("initial reveal failed")
	}
	p.ID = "second"
	p.Text = "second-synthetic"
	n, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	if helperNode(t, n, SecureInputID("second")).Reveal {
		t.Fatal("new identity inherited reveal consent")
	}
	oldReveal.Activate()
	n, _ = p.View()
	if helperNode(t, n, SecureInputID("second")).Reveal {
		t.Fatal("obsolete reveal action changed new identity")
	}
}

func TestSecureFieldIdentityRoundTripRevokesOldCallbacks(t *testing.T) {
	calls := 0
	p := &SecureField{ID: "first", Text: "synthetic", AllowReveal: true, Change: func(ChangeEvent) { calls++ }}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	oldReveal := helperNode(t, n, "first/secure/reveal")
	oldInput := helperNode(t, n, SecureInputID("first"))
	p.ID = "second"
	if _, err = p.View(); err != nil {
		t.Fatal(err)
	}
	p.ID = "first"
	n, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	oldReveal.Activate()
	oldInput.Change(ChangeEvent{Text: "old queued proposal"})
	n, _ = p.View()
	if helperNode(t, n, SecureInputID("first")).Reveal || calls != 0 {
		t.Fatal("original identity callbacks survived replacement", calls)
	}
	helperNode(t, n, "first/secure/reveal").Activate()
	n, _ = p.View()
	if !helperNode(t, n, SecureInputID("first")).Reveal {
		t.Fatal("current identity cannot reveal")
	}
}
