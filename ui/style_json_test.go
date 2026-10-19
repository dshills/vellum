package ui

import (
	"encoding/json"
	"testing"
)

// Legacy Style wire shape before the app-surface properties existed.
// Frozen replay manifests hash these exact bytes as part of their theme identity.
const legacyEmptyStyleJSON = `{"Padding":{"Set":false,"Value":{"Top":0,"Right":0,"Bottom":0,"Left":0}},"Gap":{"Set":false,"Value":0},"Align":{"Set":false,"Value":""},"Justify":{"Set":false,"Value":""},"MinSize":{"Set":false,"Value":{"Width":0,"Height":0}},"MaxSize":{"Set":false,"Value":{"Width":0,"Height":0}},"Grow":{"Set":false,"Value":0},"Foreground":{"Set":false,"Value":{"R":0,"G":0,"B":0,"A":0}},"Background":{"Set":false,"Value":{"R":0,"G":0,"B":0,"A":0}},"Typography":{"Set":false,"Value":{"Family":"","Size":0,"Weight":0,"Italic":false}}}`

func TestStyleJSONUnsetAdditionsPreserveLegacyIdentity(t *testing.T) {
	for _, style := range []Style{
		{},
		{GrowX: Optional[float32]{Value: 3}, GrowY: Optional[float32]{Value: 4},
			TextAlign: Optional[Alignment]{Value: AlignEnd}, CornerRadius: Optional[float32]{Value: 9},
			Border: Optional[Border]{Value: Border{Width: 2}}},
	} {
		data, err := json.Marshal(style)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != legacyEmptyStyleJSON {
			t.Fatalf("unset properties changed legacy identity:\n%s", data)
		}
	}
	var decoded Style
	if err := json.Unmarshal([]byte(legacyEmptyStyleJSON), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != (Style{}) {
		t.Fatal("legacy omitted properties no longer restore defaults", decoded)
	}
}

func TestStyleJSONExplicitZeroAdditionsSurviveRoundTrip(t *testing.T) {
	style := Style{GrowX: Some(float32(0)), GrowY: Some(float32(0)),
		TextAlign: Some(AlignStart), CornerRadius: Some(float32(0)), Border: Some(Border{})}
	data, err := json.Marshal(style)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"GrowX", "GrowY", "TextAlign", "CornerRadius", "Border"} {
		if _, ok := fields[name]; !ok {
			t.Fatal("explicit property disappeared", name)
		}
	}
	var decoded Style
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != style {
		t.Fatalf("explicit zero lost across JSON: got %+v, want %+v", decoded, style)
	}
}
