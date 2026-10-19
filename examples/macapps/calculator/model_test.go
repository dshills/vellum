package calculator

import (
	"context"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func press(m *Model, keys ...string) {
	for _, k := range keys {
		m.Press(k)
	}
}
func TestArithmeticAndRecovery(t *testing.T) {
	for _, tt := range []struct {
		keys []string
		want string
	}{
		{[]string{"2", "+", "3", "×", "4", "="}, "14"},
		{[]string{"0", ".", "1", "+", "0", ".", "2", "="}, "0.3"},
		{[]string{"8", "÷", "2", "=", "="}, "2"},
		{[]string{"2", "0", "0", "+", "1", "0", "%", "="}, "220"},
		{[]string{"9", "+", "−", "4", "="}, "5"},
		{[]string{"1", "±", "+", "2", "="}, "1"},
		{[]string{"1", "÷", "0", "="}, "Error"},
		{[]string{"1", "÷", "0", "=", "3"}, "3"},
		{[]string{"1", "2", "delete"}, "1"},
		{[]string{"8", "+", "delete"}, "8"},
	} {
		m := New()
		press(m, tt.keys...)
		if got := m.Display(); got != tt.want {
			t.Errorf("%v: %s want %s", tt.keys, got, tt.want)
		}
	}
}
func TestHistoryAndEntryAreBoundedAndDetached(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		m.Press("9")
	}
	if len(m.Display()) > maxEntry {
		t.Fatal("entry bound")
	}
	for i := 0; i < MaxHistory+9; i++ {
		press(m, "AC", "1", "+", "2", "=")
	}
	h := m.History()
	if len(h) != MaxHistory {
		t.Fatal(len(h))
	}
	h[0].Result = "changed"
	if m.History()[0].Result != "3" {
		t.Fatal("aliased history")
	}
	press(m, "AC")
	m.UseHistory(0)
	if m.Display() != "3" {
		t.Fatal(m.Display())
	}
	m.UseHistory(-1)
	m.UseHistory(MaxHistory)
	for i := 0; i < maxParts; i++ {
		press(m, "+", "1")
	}
	if len(m.parts) > maxParts {
		t.Fatal("unbounded expression")
	}
	if m.Press(strings.Repeat("1", 100000)) {
		t.Fatal("arbitrary token accepted")
	}
}
func find(n ui.Node, id string) *ui.Node {
	if n.ID == id {
		return &n
	}
	for _, c := range n.Children {
		if out := find(c, id); out != nil {
			return out
		}
	}
	return nil
}
func TestControlledViewAndKeyboardShareModel(t *testing.T) {
	m := New()
	if err := ui.Validate(m.View()); err != nil {
		t.Fatal(err)
	}
	if err := Theme().Validate(); err != nil {
		t.Fatal(err)
	}
	n := find(m.View(), "key-3-0")
	if n == nil {
		t.Fatal("one button missing")
	}
	n.Activate()
	r := m.Commands()
	for _, chord := range []string{"Shift++", "2", "ReturnEnter"} {
		used, err := r.Shortcut(context.Background(), chord)
		if !used || err != nil {
			t.Fatal(chord, used, err)
		}
	}
	if m.Display() != "3" {
		t.Fatal(m.Display())
	}
	m.HistoryOpen = true
	if err := ui.Validate(m.View()); err != nil {
		t.Fatal(err)
	}
}
func TestFormattedNumberCanBeReplacedOrDeleted(t *testing.T) {
	m := New()
	press(m, "0", ".", "0", "0", "0", "0", "0", "0", "0", "1", "%")
	if !strings.Contains(m.Display(), "e") {
		t.Fatal("fixture must use exponent", m.Display())
	}
	m.Press("2")
	if m.Display() != "2" {
		t.Fatal("typing after generated exponent must replace it", m.Display())
	}
	press(m, "AC", "1", "±", "delete")
	if m.Display() != "0" {
		t.Fatal(m.Display())
	}
}

func TestArithmeticMagnitudeIsBoundedBeforeParsingAndDuringOperations(t *testing.T) {
	for _, s := range []string{"1e999999999999999999999", "1e301", "1e-301", strings.Repeat("9", 65)} {
		if _, err := number(s); err == nil {
			t.Fatal("unbounded number accepted", s)
		}
	}
	m := New()
	for i := 0; i < 18; i++ {
		m.Press("9")
	}
	for i := 0; i < 32 && m.Display() != "Error"; i++ {
		press(m, "×", "=")
	}
	if m.Display() != "Error" {
		t.Fatal("exponential growth did not stop", m.Display())
	}
	if len(m.History()) > MaxHistory {
		t.Fatal("history bound")
	}
	press(m, "AC", "2", "+", "3", "=")
	if m.Display() != "5" {
		t.Fatal("budget error did not recover", m.Display())
	}
}
