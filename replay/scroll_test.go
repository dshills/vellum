package replay

import (
	"errors"
	"math"
	"testing"
)

func TestVersionedScrollAndLegacyVocabulary(t *testing.T) {
	base := Operation{Sequence: 1, Kind: Scroll, Window: "main", WindowIncarnation: 1, X: 200, Y: 150, DeltaX: -.5, DeltaY: 120}
	current := fixture()
	if err := (Record{Manifest: current, Operations: []Operation{base}}).Validate(Limits{}); err != nil {
		t.Fatal(err)
	}
	legacy := current
	legacy.Version = LegacyVersion
	if err := (Record{Manifest: legacy, Operations: []Operation{base}}).Validate(Limits{}); !errors.Is(err, ErrInvalid) {
		t.Fatal("legacy record admitted new vocabulary", err)
	}
	move := base
	move.Kind = PointerMove
	move.DeltaX = 0
	move.DeltaY = 0
	if err := (Record{Manifest: legacy, Operations: []Operation{move}}).Validate(Limits{}); err != nil {
		t.Fatal("legacy record no longer readable", err)
	}
	for _, change := range []func(*Operation){
		func(o *Operation) { o.DeltaX = math.NaN() }, func(o *Operation) { o.DeltaY = math.Inf(1) }, func(o *Operation) { o.DeltaX = 1e6 + 1 }, func(o *Operation) { o.DeltaY = -1e6 - 1 }, func(o *Operation) { o.DeltaX = 0; o.DeltaY = 0 }, func(o *Operation) { o.Button = 1 }, func(o *Operation) { o.Key = "A" }, func(o *Operation) { o.X = 800 }, func(o *Operation) { o.WindowIncarnation = 2 },
	} {
		o := base
		change(&o)
		if err := (Record{Manifest: current, Operations: []Operation{o}}).Validate(Limits{}); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid scroll accepted", o, err)
		}
	}
	for _, o := range []Operation{
		{Sequence: 1, Kind: PointerMove, Window: "main", WindowIncarnation: 1},
		{Sequence: 1, Kind: KeyChord, Window: "main", WindowIncarnation: 1, Key: "A"},
		{Sequence: 1, Kind: Frame, Window: "main", WindowIncarnation: 1},
		{Sequence: 1, Kind: Advance, Window: "main", WindowIncarnation: 1},
		{Sequence: 1, Kind: Capture, Label: "capture"},
	} {
		if err := (Record{Manifest: current, Operations: []Operation{o}}).Validate(Limits{}); err != nil {
			t.Fatal("invalid test baseline", err)
		}
		o.DeltaY = 1
		if err := (Record{Manifest: current, Operations: []Operation{o}}).Validate(Limits{}); !errors.Is(err, ErrInvalid) {
			t.Fatal("wheel payload admitted in different variant", o.Kind, err)
		}
	}
}
