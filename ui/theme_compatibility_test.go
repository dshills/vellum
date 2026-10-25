package ui

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDefaultThemePreservesLegacySerializedIdentity(t *testing.T) {
	// The immutable replay records predate both optional palette and motion fields.
	type legacyTheme struct {
		Default Style
		Kinds   map[Kind]Style
	}
	for _, theme := range []Theme{{}, {Default: Style{Gap: Some(float32(9))}, Kinds: map[Kind]Style{Button: {CornerRadius: Some(float32(6))}}}} {
		legacy, err := json.Marshal(legacyTheme{theme.Default, theme.Kinds})
		if err != nil {
			t.Fatal(err)
		}
		current, err := json.Marshal(theme)
		if err != nil || !bytes.Equal(legacy, current) {
			t.Fatal("unchanged theme identity drift", err)
		}
		theme.ReducedMotion = true
		changed, err := json.Marshal(theme)
		if err != nil || bytes.Equal(legacy, changed) {
			t.Fatal("changed motion identity not represented", err)
		}
		var decoded Theme
		if err := json.Unmarshal(changed, &decoded); err != nil || !decoded.ReducedMotion || !decoded.Clone().ReducedMotion {
			t.Fatal("motion round trip", err)
		}
	}
}
func TestExplicitPaletteIncludingZeroRemainsSerialized(t *testing.T) {
	for _, theme := range []Theme{MacTheme(false), MacTheme(true), {Palette: Some(Palette{})}} {
		data, err := json.Marshal(theme)
		if err != nil {
			t.Fatal(err)
		}
		var decoded Theme
		if err := json.Unmarshal(data, &decoded); err != nil || decoded.Palette != theme.Palette || !decoded.Palette.Set {
			t.Fatal("palette presence/value lost", err)
		}
		absent := theme
		absent.Palette = Optional[Palette]{}
		other, err := json.Marshal(absent)
		if err != nil || bytes.Equal(data, other) {
			t.Fatal("explicit palette did not affect identity", err)
		}
	}
}
