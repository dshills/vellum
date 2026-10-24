package cogent

import (
	"bytes"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image"
	"image/png"
	"strings"
	"testing"
)

func prototypeFixture(t *testing.T) (*app.Application, *TitlebarToolbarPrototype, *int) {
	t.Helper()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	ref, err := a.OpenWindow(scopeDeclaration("prototype"))
	if err != nil {
		t.Fatal(err)
	}
	var image bytes.Buffer
	if err = png.Encode(&image, imageNewPrototype()); err != nil {
		t.Fatal(err)
	}
	count := new(int)
	return a, &TitlebarToolbarPrototype{Window: ref, Items: []TitlebarToolbarItem{{ID: "new", Label: "New", Kind: ui.Button, PNG: image.Bytes(), Activate: func() { *count++ }}, {ID: "space", Kind: ui.ToolbarSpace}, {ID: "flex", Kind: ui.ToolbarFlexibleSpace}, {ID: "delete", Label: "Delete", Kind: ui.Button, PNG: bytes.Clone(image.Bytes()), Disabled: true, Activate: func() { t.Error("disabled action") }}}}, count
}
func imageNewPrototype() *image.NRGBA { return image.NewNRGBA(image.Rect(0, 0, 8, 8)) }
func TestTitlebarPrototypeValidationAndImmutableSnapshot(t *testing.T) {
	a, spec, _ := prototypeFixture(t)
	p, err := prepareTitlebarToolbar(a, spec)
	if err != nil {
		t.Fatal(err)
	}
	saved := p.wire
	spec.Items[0].PNG[0] = 0
	spec.Items[0].Label = "Changed"
	spec.Items[0].Activate = func() { t.Fatal("replacement callback") }
	if p.wire != saved || p.ref != spec.Window || len(p.actions) != 1 {
		t.Fatal("configuration was not isolated")
	}
	var decoded []titlebarWireItem
	if err = json.Unmarshal([]byte(p.wire), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded[0].Label != "New" || decoded[0].PNG[0] != 137 || decoded[1].Kind != "space" || decoded[2].Kind != "flexible" {
		t.Fatal("wire lost artwork or spaces")
	}
	if !p.accepts(spec.Window) || p.accepts(app.WindowRef{}) {
		t.Fatal("opening guard")
	}
}
func TestTitlebarPrototypeRejectsUnsupportedConfiguration(t *testing.T) {
	cases := map[string]func(*TitlebarToolbarPrototype){
		"empty":      func(s *TitlebarToolbarPrototype) { s.Items = nil },
		"oversized":  func(s *TitlebarToolbarPrototype) { s.Items = make([]TitlebarToolbarItem, 17) },
		"duplicate":  func(s *TitlebarToolbarPrototype) { s.Items[1].ID = s.Items[0].ID },
		"control-id": func(s *TitlebarToolbarPrototype) { s.Items[0].ID = "bad\x00" },
		"bad-label":  func(s *TitlebarToolbarPrototype) { s.Items[0].Label = "" },
		"long-label": func(s *TitlebarToolbarPrototype) { s.Items[0].Label = strings.Repeat("x", 129) },
		"nil-action": func(s *TitlebarToolbarPrototype) { s.Items[0].Activate = nil },
		"bad-png":    func(s *TitlebarToolbarPrototype) { s.Items[0].PNG = []byte("not an image") },
		"large-png": func(s *TitlebarToolbarPrototype) {
			var b bytes.Buffer
			_ = png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 129, 8)))
			s.Items[0].PNG = b.Bytes()
		},
		"truncated-png":     func(s *TitlebarToolbarPrototype) { s.Items[0].PNG = s.Items[0].PNG[:40] },
		"decorative-action": func(s *TitlebarToolbarPrototype) { s.Items[1].Activate = func() {} },
		"unsupported-kind":  func(s *TitlebarToolbarPrototype) { s.Items[0].Kind = ui.MenuButton },
		"zero-ref":          func(s *TitlebarToolbarPrototype) { s.Window = app.WindowRef{} },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			a, spec, _ := prototypeFixture(t)
			change(spec)
			if _, err := prepareTitlebarToolbar(a, spec); err == nil {
				t.Fatal("accepted invalid config")
			}
		})
	}
	t.Run("multiple", func(t *testing.T) {
		a, spec, _ := prototypeFixture(t)
		if err := a.Open(scopeDeclaration("other")); err != nil {
			t.Fatal(err)
		}
		if _, err := prepareTitlebarToolbar(a, spec); err == nil {
			t.Fatal("accepted multiple windows")
		}
	})
	t.Run("unified", func(t *testing.T) {
		a, _ := app.New(app.Config{})
		defer a.Close()
		w := scopeDeclaration("unified")
		w.Chrome.Style = app.ChromeUnified
		ref, err := a.OpenWindow(w)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = prepareTitlebarToolbar(a, &TitlebarToolbarPrototype{Window: ref}); err == nil {
			t.Fatal("accepted fullsize chrome")
		}
	})
}
func TestTitlebarPrototypeQueuedActionsAndModalRetirement(t *testing.T) {
	for _, cancel := range []string{"none", "modal-round-trip", "detach", "window-close"} {
		t.Run(cancel, func(t *testing.T) {
			a, spec, count := prototypeFixture(t)
			p, err := prepareTitlebarToolbar(a, spec)
			if err != nil {
				t.Fatal(err)
			}
			b := newTitlebarToolbarBinding(a, p)
			if b.deliver("new") {
				t.Fatal("accepted before publication")
			}
			b.setAllowed(true)
			if b.deliver("delete") || b.deliver("space") || b.deliver("unknown") {
				t.Fatal("accepted non-action")
			}
			if !b.deliver("new") || *count != 0 {
				t.Fatal("callback reentered native main")
			}
			switch cancel {
			case "modal-round-trip":
				b.setAllowed(false)
				b.setAllowed(true)
			case "detach":
				b.close()
			case "window-close":
				if err = a.Open(scopeDeclaration("survivor")); err != nil {
					t.Fatal(err)
				}
				if err = a.CloseWindow(spec.Window); err != nil {
					t.Fatal(err)
				}
				if _, err = a.OpenWindow(scopeDeclaration("prototype")); err != nil {
					t.Fatal(err)
				}
			}
			if err = a.Drain(); err != nil {
				t.Fatal(err)
			}
			want := 0
			if cancel == "none" {
				want = 1
			}
			if *count != want {
				t.Fatal(cancel, *count)
			}
			if cancel == "none" || cancel == "modal-round-trip" {
				if !b.deliver("new") {
					t.Fatal("live delivery refused")
				}
				if err = a.Drain(); err != nil {
					t.Fatal(err)
				}
				if *count != want+1 {
					t.Fatal("wrong count")
				}
			}
			b.close()
			b.close()
			if b.deliver("new") {
				t.Fatal("closed binding admits action")
			}
		})
	}
}

func TestTitlebarPrototypeBridgeRetirementRevokesPendingActions(t *testing.T) {
	a, spec, count := prototypeFixture(t)
	config, err := prepareTitlebarToolbar(a, spec)
	if err != nil {
		t.Fatal(err)
	}
	bridge := newAccessibilityBridge(a)
	binding := newTitlebarToolbarBinding(a, config)
	bridge.titlebar.Store(binding)
	binding.setAllowed(true)
	if !deliverTitlebarToolbarAction(bridge.token, "new") {
		t.Fatal("native token not admitted")
	}
	bridge.closeFrom("titlebar-test")
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if *count != 0 || deliverTitlebarToolbarAction(bridge.token, "new") || deliverTitlebarToolbarAction(0, "new") {
		t.Fatal("retired native token invoked command")
	}
	if !binding.closed.Load() {
		t.Fatal("pending input not revoked until native detach")
	}
}

func TestTitlebarPrototypeCustomizationIdentity(t *testing.T) {
	a, spec, _ := prototypeFixture(t)
	spec.CustomizationID = "com.example.documents.light-v1"
	prepared, err := prepareTitlebarToolbar(a, spec)
	if err != nil {
		t.Fatal(err)
	}
	spec.CustomizationID = "changed"
	if prepared.customizationID != "com.example.documents.light-v1" {
		t.Fatal("settings identity was not copied")
	}
	for _, id := range []string{"bad\x00id", "é", "has space", strings.Repeat("x", 129)} {
		spec.CustomizationID = id
		if _, err = prepareTitlebarToolbar(a, spec); err == nil {
			t.Fatal("invalid customization ID accepted", id)
		}
	}
	spec.CustomizationID = ""
	plain, err := prepareTitlebarToolbar(a, spec)
	if err != nil || plain.customizationID != "" {
		t.Fatal("legacy unsaved configuration changed", err)
	}
}
func TestTitlebarPrototypeCustomizationScopeAndMembership(t *testing.T) {
	for _, revoke := range []string{"palette-round-trip", "items-changed", "managed-while-palette", "close-while-palette"} {
		t.Run(revoke, func(t *testing.T) {
			a, spec, count := prototypeFixture(t)
			p, err := prepareTitlebarToolbar(a, spec)
			if err != nil {
				t.Fatal(err)
			}
			binding := newTitlebarToolbarBinding(a, p)
			binding.setAllowed(true)
			if !binding.deliver("new") {
				t.Fatal("initial input refused")
			}
			switch revoke {
			case "items-changed":
				binding.itemsChanged()
			case "palette-round-trip":
				if !binding.setCustomizing(true) || binding.deliver("new") {
					t.Fatal("palette allowed source command")
				}
				active := binding.permission.Load()
				if !binding.setCustomizing(true) || binding.permission.Load() != active {
					t.Fatal("duplicate native palette entry was refused or retired twice")
				}
				if !binding.setCustomizing(false) {
					t.Fatal("palette exit refused")
				}
				restored := binding.permission.Load()
				if !binding.setCustomizing(false) || binding.permission.Load() != restored {
					t.Fatal("duplicate native palette exit retired restored permission")
				}

			case "managed-while-palette":
				binding.setCustomizing(true)
				binding.setAllowed(false)
				binding.setCustomizing(false)
				if binding.deliver("new") || binding.setCustomizing(true) {
					t.Fatal("native palette bypassed managed modal")
				}
			case "close-while-palette":
				binding.setCustomizing(true)
				binding.close()
				binding.setCustomizing(false)
			}
			if err = a.Drain(); err != nil {
				t.Fatal(err)
			}
			if *count != 0 {
				t.Fatal("stale input survived customization")
			}
			if revoke != "close-while-palette" {
				binding.setAllowed(true)
				if !binding.deliver("new") {
					t.Fatal("input not restored")
				}
				if err = a.Drain(); err != nil {
					t.Fatal(err)
				}
				if *count != 1 {
					t.Fatal("restored action not exactly once")
				}
			}
		})
	}
}
func TestTitlebarPrototypeCustomizationRequestRejectsForeignOrPlainWindow(t *testing.T) {
	a, spec, _ := prototypeFixture(t)
	p, err := prepareTitlebarToolbar(a, spec)
	if err != nil {
		t.Fatal(err)
	}
	bridge := newAccessibilityBridge(a)
	defer bridge.closeFrom("customization-request-test")
	bridge.bindScoped(spec.Window)
	bridge.titlebar.Store(newTitlebarToolbarBinding(a, p))
	if err = CustomizeTitlebarToolbar(spec.Window); err == nil {
		t.Fatal("plain toolbar accepted palette request")
	}
	if err = CustomizeTitlebarToolbar(app.WindowRef{}); err == nil {
		t.Fatal("zero window accepted")
	}
	other, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	foreign, err := other.OpenWindow(scopeDeclaration(spec.Window.ID()))
	if err != nil {
		t.Fatal(err)
	}
	if err = CustomizeTitlebarToolbar(foreign); err == nil {
		t.Fatal("foreign window accepted")
	}
}
