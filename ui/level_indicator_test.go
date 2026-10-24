package ui

import (
	"image/color"
	"math"
	"reflect"
	"strings"
	"testing"
)

func levelNode(n Node, id string) Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := levelNode(c, id); found.ID != "" {
			return found
		}
	}
	return Node{}
}
func levelFixture(mode LevelMode) *LevelIndicator {
	p := &LevelIndicator{ID: "level", Label: "Storage", Mode: mode, Max: 100, Value: 50}
	if mode == LevelRating {
		p.Max = 5
		p.Value = 3
		p.AllowClear = true
		p.Change = func(LevelEvent) {}
	}
	return p
}

func TestLevelIndicatorCapacityPixelsThresholdsAndImmutability(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, mode := range []LevelMode{LevelContinuous, LevelDiscrete} {
			t.Run(string(mode)+map[bool]string{false: "-light", true: "-dark"}[dark], func(t *testing.T) {
				p := levelFixture(mode)
				p.Dark = dark
				p.Warning = Some(float32(75))
				p.Critical = Some(float32(90))
				n, err := p.View()
				if err != nil {
					t.Fatal(err)
				}
				first := levelNode(n, "level/level/meter").Image
				if first == nil || len(FocusOrder(n)) != 0 {
					t.Fatal("capacity must be a passive image", FocusOrder(n))
				}
				normal := first.NRGBAAt(20, 14)
				rail := first.NRGBAAt(450, 14)
				if normal == rail {
					t.Fatal("filled portion does not differ from track")
				}
				n, _ = p.View()
				if levelNode(n, "level/level/meter").Image != first {
					t.Fatal("unchanged preview reallocated")
				}
				p.Value = 75
				n, _ = p.View()
				warning := levelNode(n, "level/level/meter").Image
				if warning == first || warning.NRGBAAt(20, 14) == normal || !strings.Contains(n.Accessibility.Name, "Warning") {
					t.Fatal("warning threshold not rendered")
				}
				p.Value = 90
				n, _ = p.View()
				critical := levelNode(n, "level/level/meter").Image
				if critical.NRGBAAt(20, 14) == warning.NRGBAAt(20, 14) || !strings.Contains(n.Accessibility.Name, "Critical") {
					t.Fatal("critical threshold not rendered")
				}
				if first.NRGBAAt(20, 14) != normal {
					t.Fatal("old immutable image mutated")
				}
				p.Value = 0
				n, _ = p.View()
				if got := levelNode(n, "level/level/meter").Image.NRGBAAt(20, 14); got != rail {
					t.Fatal("minimum does not empty meter", got, rail)
				}
				p.Value = 100
				n, _ = p.View()
				if got := levelNode(n, "level/level/meter").Image.NRGBAAt(450, 14); got == rail {
					t.Fatal("maximum does not fill meter")
				}
				if !strings.Contains(n.Accessibility.Name, "range 0 to 100") {
					t.Fatal("missing accessible range")
				}
			})
		}
	}
}

func TestLevelIndicatorDiscreteQuantizationAndOffsetBounds(t *testing.T) {
	p := levelFixture(LevelDiscrete)
	p.Min = -10
	p.Max = 10
	p.Value = -9
	p.Segments = 4
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	im := levelNode(n, "level/level/meter").Image
	if im.NRGBAAt(60, 14) == im.NRGBAAt(180, 14) {
		t.Fatal("partial first segment not lit, or unoccupied second segment lit")
	}
	if im.NRGBAAt(60, 14) == im.NRGBAAt(120, 14) {
		t.Fatal("segment separator missing")
	}
	p.Value = 10
	n, _ = p.View()
	if levelNode(n, "level/level/meter").Image.NRGBAAt(420, 14) == im.NRGBAAt(420, 14) {
		t.Fatal("maximum failed to light last segment")
	}
	p.Min = -math.MaxFloat32
	p.Max = math.MaxFloat32
	p.Value = 0
	if _, err = p.View(); err != nil {
		t.Fatal("finite float32 range should use double fraction arithmetic", err)
	}
}

func TestLevelIndicatorRatingControlledZeroClearAndExactlyOnce(t *testing.T) {
	p := levelFixture(LevelRating)
	calls := []LevelEvent{}
	p.Change = func(e LevelEvent) { calls = append(calls, e) }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{LevelStarID(p.ID, 1), LevelStarID(p.ID, 2), LevelStarID(p.ID, 3), LevelStarID(p.ID, 4), LevelStarID(p.ID, 5), LevelZeroID(p.ID), LevelClearID(p.ID)}
	if got := FocusOrder(n); !reflect.DeepEqual(got, want) {
		t.Fatal("ordinary buttons must provide keyboard access", got)
	}
	set := levelNode(n, LevelStarID(p.ID, 4))
	set.Activate()
	if len(calls) != 1 || calls[0] != (LevelEvent{ID: p.ID, Value: 4}) || p.Value != 3 {
		t.Fatal("controlled rejection mutated value", calls, p.Value)
	}
	p.Change = func(e LevelEvent) { calls = append(calls, e); p.Value, p.Unrated = e.Value, e.Unrated }
	set.Activate()
	set.Activate()
	if p.Value != 4 || len(calls) != 2 {
		t.Fatal("accepted value or duplicate guard", calls, p.Value)
	}
	levelNode(n, LevelZeroID(p.ID)).Activate()
	if p.Value != 0 || p.Unrated {
		t.Fatal("zero mistaken for unrated")
	}
	levelNode(n, LevelClearID(p.ID)).Activate()
	if p.Value != 0 || !p.Unrated || !calls[len(calls)-1].Unrated {
		t.Fatal("clear not distinct")
	}
	levelNode(n, LevelClearID(p.ID)).Activate()
	if len(calls) != 4 {
		t.Fatal("redundant clear emitted")
	}
	n, _ = p.View()
	if !strings.Contains(n.Accessibility.Name, "Unrated") {
		t.Fatal("unrated not exposed")
	}
	for _, c := range levelNode(n, "level/level/stars").Children[:5] {
		if c.Text != "☆" || !strings.Contains(c.Accessibility.Name, "current Unrated") {
			t.Fatal("unrated stars/accessibility", c)
		}
	}
}

func TestLevelIndicatorRatingStaleDisabledAndInvalidGuards(t *testing.T) {
	for _, mode := range []string{"disabled", "readonly", "identity", "range", "mode", "clear-policy", "invalid-style", "invalid-view"} {
		t.Run(mode, func(t *testing.T) {
			p := levelFixture(LevelRating)
			calls := 0
			p.Change = func(LevelEvent) { calls++ }
			n, _ := p.View()
			star := levelNode(n, LevelStarID(p.ID, 4))
			clear := levelNode(n, LevelClearID(p.ID))
			switch mode {
			case "disabled":
				p.Disabled = true
			case "readonly":
				p.Change = nil
			case "identity":
				p.ID = "other"
			case "range":
				p.Max = 4
			case "mode":
				p.Mode = LevelContinuous
			case "clear-policy":
				p.AllowClear = false
			case "invalid-style":
				p.Style.Gap = Some(float32(-1))
			case "invalid-view":
				p.Style.Gap = Some(float32(-1))
				if _, err := p.View(); err == nil {
					t.Fatal("invalid description accepted")
				}
				p.Style = Style{}
			}
			star.Activate()
			clear.Activate()
			if calls != 0 {
				t.Fatal("stale action admitted", mode, calls)
			}
		})
	}
	p := levelFixture(LevelRating)
	calls := 0
	p.Change = func(LevelEvent) { calls++ }
	p.Disabled = true
	n, _ := p.View()
	p.Disabled = false
	levelNode(n, LevelStarID(p.ID, 4)).Activate()
	if calls != 0 {
		t.Fatal("disabled description reactivated before new view")
	}
	n, _ = p.View()
	levelNode(n, LevelStarID(p.ID, 4)).Activate()
	if calls != 1 {
		t.Fatal("fresh description did not activate")
	}
	p.Change = nil
	n, _ = p.View()
	if len(FocusOrder(n)) != 0 {
		t.Fatal("read-only rating retained focus targets")
	}
}

func TestLevelIndicatorValidation(t *testing.T) {
	cases := []func(*LevelIndicator){func(p *LevelIndicator) { p.ID = "" }, func(p *LevelIndicator) { p.ID = "menu:reserved" }, func(p *LevelIndicator) { p.Mode = "unknown" }, func(p *LevelIndicator) { p.Value = float32(math.NaN()) }, func(p *LevelIndicator) { p.Min = float32(math.Inf(-1)) }, func(p *LevelIndicator) { p.Max = p.Min }, func(p *LevelIndicator) { p.Value = 101 }, func(p *LevelIndicator) { p.Warning = Some(float32(-1)) }, func(p *LevelIndicator) { p.Critical = Some(float32(math.Inf(1))) }, func(p *LevelIndicator) { p.Warning = Some(float32(90)); p.Critical = Some(float32(75)) }, func(p *LevelIndicator) { p.Change = func(LevelEvent) {} }, func(p *LevelIndicator) { p.AllowClear = true }, func(p *LevelIndicator) { p.Segments = 41 }, func(p *LevelIndicator) { p.Style.MinSize = Some(Size{Width: -1}) }}
	for i, mutate := range cases {
		p := levelFixture(LevelContinuous)
		mutate(p)
		if _, err := p.View(); err == nil {
			t.Fatal("invalid capacity accepted", i)
		}
	}
	if _, err := (*LevelIndicator)(nil).View(); err == nil {
		t.Fatal("nil accepted")
	}
	for i, mutate := range []func(*LevelIndicator){func(p *LevelIndicator) { p.Min = 1 }, func(p *LevelIndicator) { p.Max = 11 }, func(p *LevelIndicator) { p.Max = 5.5 }, func(p *LevelIndicator) { p.Value = 2.5 }, func(p *LevelIndicator) { p.Warning = Some(float32(2)) }, func(p *LevelIndicator) { p.Segments = 5 }, func(p *LevelIndicator) { p.Unrated = true }, func(p *LevelIndicator) { p.Value = 0; p.Unrated = true; p.AllowClear = false }} {
		p := levelFixture(LevelRating)
		mutate(p)
		if _, err := p.View(); err == nil {
			t.Fatal("invalid rating accepted", i)
		}
	}
	p := levelFixture(LevelRating)
	p.AllowClear = false
	n, err := p.View()
	if err != nil || levelNode(n, LevelClearID(p.ID)).ID != "" {
		t.Fatal("clear action exists without AllowClear", err)
	}
	root := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: LevelStarID(p.ID, 1), Kind: Label}}}
	if Validate(root) == nil {
		t.Fatal("generated ID collision accepted")
	}
}

func TestLevelIndicatorDarkAndDisabledPalette(t *testing.T) {
	p := levelFixture(LevelContinuous)
	n, _ := p.View()
	light := levelNode(n, "level/level/meter").Image
	p.Dark = true
	n, _ = p.View()
	dark := levelNode(n, "level/level/meter").Image
	if light.NRGBAAt(400, 14) == dark.NRGBAAt(400, 14) {
		t.Fatal("dark track not adapted")
	}
	p.Disabled = true
	n, _ = p.View()
	disabled := levelNode(n, "level/level/meter").Image
	if disabled.NRGBAAt(20, 14) == dark.NRGBAAt(20, 14) {
		t.Fatal("disabled capacity not muted")
	}
	if light.NRGBAAt(0, 0) != (color.NRGBA{}) {
		t.Fatal("rounded corner should be transparent")
	}
}
