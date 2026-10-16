package cogent

import (
	"reflect"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/svg"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func TestIconsBundledArtworkAndRetainedPresentation(t *testing.T) {
	for _, symbol := range ui.Symbols() {
		art := nativeSymbol(symbol)
		s := svg.NewSVG(math32.Vec2(24, 24))
		if !art.IsSet() || s.ReadXML(strings.NewReader(string(art))) != nil || s.Root == nil || !s.Root.HasChildren() {
			t.Fatal("bundled symbol has no usable artwork", symbol)
		}
	}
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("icons"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
		{ID: "decorative", Kind: ui.Icon, Icon: ui.SymbolInfo},
		{ID: "named", Kind: ui.Icon, Icon: ui.SymbolWarning, Text: "<Warning>"},
		{ID: "action", Kind: ui.Button, Icon: ui.SymbolSave, Text: "<Save>", Activate: func() {}},
		{ID: "toggle", Kind: ui.ToggleButton, Icon: ui.SymbolCheck, Checked: true, Accessibility: ui.Accessibility{Name: "Guides"}},
	}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	action := findRecord(r, "action").widget.(*core.Button)
	toggle := findRecord(r, "toggle").widget.(*core.Switch)
	named := findRecord(r, "named").widget.(*core.Icon)
	if action.Text != "&lt;Save&gt;" || action.Icon != icons.Save || action.Type != core.ButtonFilled || toggle.IconOn != icons.Check || toggle.IconOff != icons.Check || !toggle.IsChecked() {
		t.Fatal("native initial icon/presentation", action.Icon, action.Type, toggle.IconOn, toggle.IconOff)
	}
	for _, test := range []struct {
		appearance ui.ButtonAppearance
		want       core.ButtonTypes
	}{{ui.ButtonSecondary, core.ButtonOutlined}, {ui.ButtonQuiet, core.ButtonText}, {ui.ButtonTool, core.ButtonAction}, {ui.ButtonPrimary, core.ButtonFilled}, {"", core.ButtonFilled}} {
		n.Children[2].Appearance = test.appearance
		n.Children[2].Icon = ui.SymbolSearch
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		if findRecord(r, "action").widget != action || action.Type != test.want || action.Icon != icons.Search {
			t.Fatal("appearance change lost control identity/presentation", test)
		}
	}
	n.Children[2].Appearance = ui.ButtonTool
	n.Children[2].Style.Padding = ui.Some(ui.UniformInsets(17))
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	action.Style()
	if action.Styles.Padding.Left.Value != 17 || action.Styles.Padding.Top.Value != 17 {
		t.Fatal("explicit style lost to tool appearance", action.Styles.Padding)
	}
	n.Children[2].Style = ui.Style{}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	action.Style()
	if action.Styles.Padding.Left.Value != 6 || action.Styles.Padding.Top.Value != 6 {
		t.Fatal("removing style did not restore compact tool dimensions", action.Styles.Padding)
	}
	n.Children[1].Icon = ui.SymbolInfo
	n.Children[1].Style.MinSize = ui.Some(ui.Size{Width: 40, Height: 32})
	n.Children[2].Icon = ""
	n.Children[3].Icon = ""
	n.Children[3].Text = "Guides"
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	named.Style()
	if findRecord(r, "named").widget != named || named.Icon != icons.Info || named.Styles.Min.X.Value != 40 || named.Styles.Min.Y.Value != 32 || action.Icon.IsSet() || toggle.IconOn != icons.Check || toggle.IconOff.IsSet() || !toggle.IsChecked() {
		t.Fatal("icon removal/change or explicit sizing not retained")
	}
	n.Children[1].Style = ui.Style{}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	named.Style()
	if named.Styles.Min.X.Value != 24 || named.Styles.Min.Y.Value != 24 {
		t.Fatal("removing explicit icon dimensions did not restore defaults", named.Styles.Min)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	foundNamed, foundAction := false, false
	for _, node := range nodes {
		if node.DomainID == "decorative" {
			t.Fatal("decorative symbol entered accessible hierarchy")
		}
		if node.DomainID == "named" {
			foundNamed = true
			if node.Role != "image" || node.Name != "<Warning>" || node.Focusable || len(node.Actions) != 0 || handlers[node.ID] != nil || node.Value != nil {
				t.Fatal("named symbol semantics", node)
			}
		}
		if node.DomainID == "action" {
			foundAction = true
			if node.Role != "button" || node.Name != "<Save>" || !reflect.DeepEqual(node.Actions, []string{"focus", "press"}) {
				t.Fatal("icon action semantics", node)
			}
		}
	}
	if !foundNamed || !foundAction {
		t.Fatal("named icon/action missing from accessible hierarchy")
	}
	if got := ui.FocusOrder(n); !reflect.DeepEqual(got, []string{"action", "toggle"}) {
		t.Fatal(got)
	}
}

func TestIconsButtonOwnerDeliveryAndRetirement(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("icon buttons"))
	called := ""
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "action", Kind: ui.Button, Icon: ui.SymbolSave, Accessibility: ui.Accessibility{Name: "Save"}, Activate: func() { called = "old" }}}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w := findRecord(r, "action").widget.(*core.Button)
	w.Send(events.Click)
	if called != "" {
		t.Fatal("button bypassed owner delivery")
	}
	n.Children[0].Activate = func() { called = "latest" }
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if called != "latest" {
		t.Fatal("queued action used stale callback", called)
	}
	called = ""
	w.Send(events.Click)
	n.Disabled = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if called != "" {
		t.Fatal("disabled ancestor delivered queued action")
	}
	n.Disabled = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w.Send(events.Click)
	n.Children = nil
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if called != "" {
		t.Fatal("removed button delivered queued action")
	}
}

func TestIconsButtonWindowClosureRevokesPendingActivation(t *testing.T) {
	a, _ := newAXTest(t, 8)
	if err := a.Open(app.Window{ID: "other", Title: "Surviving window", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("closed button window"))
	called := 0
	r, err := reconcile(host, nil, ui.Node{ID: "action", Kind: ui.Button, Icon: ui.SymbolSave, Accessibility: ui.Accessibility{Name: "Save"}, Activate: func() { called++ }}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w := r.widget.(*core.Button)
	w.Send(events.Click)
	if err := a.CloseWindow(fixtureWindowRef(a, "ax")); err != nil {
		t.Fatal(err)
	}
	witness := 0
	if err := a.PostWindow(fixtureWindowRef(a, "other"), func() { witness++ }); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	w.Send(events.Click)
	a.Drain()
	if called != 0 || witness != 1 {
		t.Fatal("closed-window activation or surviving-owner admission", called, witness)
	}
}
