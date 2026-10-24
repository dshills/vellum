package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/dshills/vellum/canvas"
)

// LevelMode selects a passive capacity meter or an editable star rating.
type LevelMode string

const (
	LevelContinuous LevelMode = "continuous"
	LevelDiscrete   LevelMode = "discrete"
	LevelRating     LevelMode = "rating"
)

// LevelEvent proposes a rating. Unrated is distinct from a rating of zero;
// an unrated proposal always carries Value zero. The application accepts a
// proposal by updating Value and Unrated. Rejected proposals change no state.
type LevelEvent struct {
	ID      string
	Value   float32
	Unrated bool
}

// LevelIndicator is an owner-held, custom-rendered Mac-style capacity/rating
// component. Retain one instance on one application owner. Capacity uses finite
// Min < Max and Value within that range. Warning/Critical are optional absolute
// values: a value at or above each threshold changes the meter color and text.
// Discrete capacity defaults to 10 segments (maximum 40); a partly occupied
// segment is lit. Capacity is passive and does not accept Change.
//
// Rating requires Min=0 and integral Max in 1..10, with integral Value. Each
// star is a regular accessible button: Tab navigates, Return/Space proposes its
// value. The separate zero button proposes a rated zero. AllowClear adds a
// distinct Clear action; Unrated is permitted only with AllowClear and Value=0.
// Nil Change makes ratings read-only. This is not AppKit NSLevelIndicator and
// does not implement arrow-key or pointer-drag rating adjustment.
type LevelIndicator struct {
	ID, Label           string
	Mode                LevelMode
	Min, Max, Value     float32
	Warning, Critical   Optional[float32]
	Segments            int
	Unrated, AllowClear bool
	Dark, Disabled      bool
	Change              func(LevelEvent)
	Style               Style
	preview             *canvas.TiledImage
	previewKey          levelPreviewKey
	previewValid        bool
	revision            uint64
	identity            levelIdentity
	identityValid       bool
}
type levelIdentity struct {
	id       string
	mode     LevelMode
	min, max float32
	clear    bool
	segments int
}
type levelPreviewKey struct {
	mode              LevelMode
	min, max, value   float32
	warning, critical Optional[float32]
	segments          int
	dark, disabled    bool
}

func LevelStarID(root string, value int) string { return fmt.Sprintf("%s/level/star/%d", root, value) }
func LevelZeroID(root string) string            { return root + "/level/zero" }
func LevelClearID(root string) string           { return root + "/level/clear" }

func (p *LevelIndicator) validate() error {
	if p == nil || strings.TrimSpace(p.ID) == "" || !finite(p.Min) || !finite(p.Max) || !finite(p.Value) || p.Min >= p.Max || p.Value < p.Min || p.Value > p.Max {
		return fmt.Errorf("ui: level indicator requires ID, finite ordered bounds and value within range")
	}
	if p.Mode != LevelContinuous && p.Mode != LevelDiscrete && p.Mode != LevelRating {
		return fmt.Errorf("ui: level indicator requires continuous, discrete or rating mode")
	}
	if p.Mode == LevelRating {
		if p.Min != 0 || p.Max > 10 || p.Max < 1 || p.Max != float32(math.Trunc(float64(p.Max))) || p.Value != float32(math.Trunc(float64(p.Value))) || p.Segments != 0 || p.Warning.Set || p.Critical.Set || (p.Unrated && (!p.AllowClear || p.Value != 0)) {
			return fmt.Errorf("ui: rating requires integral 0..1-10 bounds/value, and no capacity settings; unrated requires AllowClear and zero value")
		}
	} else {
		if p.Change != nil || p.Unrated || p.AllowClear || p.Segments < 0 || p.Segments > 40 || (p.Mode == LevelContinuous && p.Segments != 0) {
			return fmt.Errorf("ui: capacity is passive, with at most 40 segments in discrete mode")
		}
		for _, threshold := range []Optional[float32]{p.Warning, p.Critical} {
			if threshold.Set && (!finite(threshold.Value) || threshold.Value < p.Min || threshold.Value > p.Max) {
				return fmt.Errorf("ui: level threshold must be finite and within range")
			}
		}
		if p.Warning.Set && p.Critical.Set && p.Warning.Value > p.Critical.Value {
			return fmt.Errorf("ui: warning threshold must not exceed critical threshold")
		}
	}
	return p.Style.Validate()
}

func (p *LevelIndicator) View() (result Node, err error) {
	if p == nil {
		return Node{}, fmt.Errorf("ui: nil level indicator")
	}
	// A failed description or replacement invalidates retained action closures.
	defer func() {
		if err != nil {
			p.revision++
			p.identityValid = false
		}
	}()
	if err = p.validate(); err != nil {
		return Node{}, err
	}
	identity := levelIdentity{p.ID, p.Mode, p.Min, p.Max, p.AllowClear, p.Segments}
	if !p.identityValid || p.identity != identity {
		p.revision++
		p.identity = identity
		p.identityValid = true
	}
	id, revision := p.ID, p.revision
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Level"
	}
	value := fmt.Sprintf("%g of %g (range %g to %g)", p.Value, p.Max, p.Min, p.Max)
	status := "Normal"
	if p.Warning.Set && p.Value >= p.Warning.Value {
		status = "Warning"
	}
	if p.Critical.Set && p.Value >= p.Critical.Value {
		status = "Critical"
	}
	if p.Mode == LevelRating {
		value = fmt.Sprintf("%g of %g stars", p.Value, p.Max)
		if p.Unrated {
			value = "Unrated"
		}
	}
	root := Node{ID: id, Kind: Column, Style: Style{Gap: Some(float32(5)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}.Overlay(p.Style), Accessibility: Accessibility{Name: label + ", " + value}}
	root.Children = append(root.Children, Node{ID: id + "/level/value", Kind: Label, Text: label + ": " + value, Accessibility: Accessibility{Name: label + ", " + value}})
	if p.Mode != LevelRating {
		description := label + ", " + value + ", " + status
		root.Children[0].Text += " — " + status
		root.Children[0].Accessibility.Name = description
		root.Accessibility.Name = description
		preview, previewErr := p.capacityImage()
		if previewErr != nil {
			return Node{}, previewErr
		}
		root.Children = append(root.Children, Node{ID: id + "/level/meter", Kind: Image, Image: preview, Text: description, Accessibility: Accessibility{Name: description}, Style: Style{MinSize: Some(Size{Width: 240, Height: 14}), MaxSize: Some(Size{Width: 240, Height: 14}), Grow: Some(float32(0))}})
	} else {
		renderedAvailable := !p.Disabled && p.Change != nil
		available := func() bool {
			return renderedAvailable && p.revision == revision && p.identityValid && p.identity == identity && p.ID == id && p.Mode == LevelRating && !p.Disabled && p.Change != nil && p.validate() == nil && p.Min == identity.min && p.Max == identity.max && p.AllowClear == identity.clear
		}
		propose := func(value float32, unrated bool) func() {
			return func() {
				if available() && (!unrated || p.AllowClear) && (p.Value != value || p.Unrated != unrated) {
					p.Change(LevelEvent{ID: id, Value: value, Unrated: unrated})
				}
			}
		}
		row := Node{ID: id + "/level/stars", Kind: Row, Style: Style{Gap: Some(float32(2)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}}
		ink := color.NRGBA{R: 128, G: 128, B: 133, A: 255}
		filled := color.NRGBA{R: 208, G: 137, B: 0, A: 255}
		if p.Dark {
			ink = color.NRGBA{R: 161, G: 161, B: 166, A: 255}
			filled = color.NRGBA{R: 255, G: 202, B: 71, A: 255}
		}
		if p.Disabled {
			filled = ink
		}
		for i := 1; i <= int(p.Max); i++ {
			text, fg := "☆", ink
			if !p.Unrated && float32(i) <= p.Value {
				text = "★"
				fg = filled
			}
			row.Children = append(row.Children, Node{ID: LevelStarID(id, i), Kind: Button, Appearance: ButtonQuiet, Text: text, Disabled: !available(), Accessibility: Accessibility{Name: fmt.Sprintf("%s: set %d of %g stars; current %s", label, i, p.Max, value), Help: "Press Return or Space to set this rating"}, Style: Style{Padding: Some(UniformInsets(0)), MinSize: Some(Size{Width: 28, Height: 28}), MaxSize: Some(Size{Width: 28, Height: 28}), Border: Some(Border{}), Background: Some(color.NRGBA{}), Foreground: Some(fg), Typography: Some(Typography{Size: 23}), Grow: Some(float32(0))}, Activate: propose(float32(i), false)})
		}
		row.Children = append(row.Children, Node{ID: LevelZeroID(id), Kind: Button, Appearance: ButtonQuiet, Text: "0", Disabled: !available(), Accessibility: Accessibility{Name: label + ": set zero stars", Help: "Zero is a rating; Clear makes it unrated"}, Activate: propose(0, false)})
		if p.AllowClear {
			row.Children = append(row.Children, Node{ID: LevelClearID(id), Kind: Button, Appearance: ButtonQuiet, Text: "Clear", Disabled: !available(), Accessibility: Accessibility{Name: "Clear " + label, Help: "Set this rating to Unrated, distinct from zero stars"}, Activate: propose(0, true)})
		}
		root.Children = append(root.Children, row)
	}
	if err = Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}

func (p *LevelIndicator) capacityImage() (*canvas.TiledImage, error) {
	key := levelPreviewKey{p.Mode, p.Min, p.Max, p.Value, p.Warning, p.Critical, p.Segments, p.Dark, p.Disabled}
	if p.previewValid && key == p.previewKey {
		return p.preview, nil
	}
	// Two source pixels per logical pixel keep rounded edges crisp at Retina scale.
	pixels := image.NewNRGBA(image.Rect(0, 0, 480, 28))
	rail := color.NRGBA{R: 225, G: 225, B: 229, A: 255}
	edge := color.NRGBA{R: 174, G: 174, B: 180, A: 255}
	fill := color.NRGBA{R: 52, G: 177, B: 79, A: 255}
	if p.Warning.Set && p.Value >= p.Warning.Value {
		fill = color.NRGBA{R: 237, G: 173, B: 20, A: 255}
	}
	if p.Critical.Set && p.Value >= p.Critical.Value {
		fill = color.NRGBA{R: 229, G: 69, B: 65, A: 255}
	}
	if p.Dark {
		rail = color.NRGBA{R: 57, G: 57, B: 61, A: 255}
		edge = color.NRGBA{R: 102, G: 102, B: 107, A: 255}
	}
	if p.Disabled {
		fill = edge
	}
	fraction := (float64(p.Value) - float64(p.Min)) / (float64(p.Max) - float64(p.Min))
	count := 1
	if p.Mode == LevelDiscrete {
		count = p.Segments
		if count == 0 {
			count = 10
		}
	}
	lit := int(math.Ceil(fraction * float64(count)))
	for y := 0; y < 28; y++ {
		for x := 0; x < 480; x++ {
			// A bounded rounded outer track, with a one-logical-pixel inset fill.
			dx := math.Max(math.Max(6-float64(x)-0.5, float64(x)+0.5-474), 0)
			dy := math.Max(math.Max(6-float64(y)-0.5, float64(y)+0.5-22), 0)
			if dx*dx+dy*dy > 36 {
				continue
			}
			c := edge
			if x >= 2 && x < 478 && y >= 2 && y < 26 {
				c = rail
				if p.Mode == LevelContinuous {
					if float64(x-2) < fraction*476 {
						c = fill
					}
				} else {
					cell := float64(x-2) * float64(count) / 476
					segment := int(cell)
					gap := cell - float64(segment)
					if segment < lit && gap >= 0.06 && gap <= 0.94 {
						c = fill
					}
				}
			}
			pixels.SetNRGBA(x, y, c)
		}
	}
	result, err := canvas.NewTiledImage(pixels)
	if err != nil {
		return nil, err
	}
	p.preview, p.previewKey, p.previewValid = result, key, true
	return result, nil
}
