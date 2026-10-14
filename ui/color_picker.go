package ui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/dshills/vellum/canvas"
)

// ColorChannel names a stable field and the component changed by a ColorEdit.
// ColorHex replaces the whole value; the byte channels preserve other bytes.
type ColorChannel string

const (
	ColorRed   ColorChannel = "red"
	ColorGreen ColorChannel = "green"
	ColorBlue  ColorChannel = "blue"
	ColorAlpha ColorChannel = "alpha"
	ColorHex   ColorChannel = "hex"
)

var ErrColorHex = errors.New("ui: color requires #RGB, #RGBA, #RRGGBB or #RRGGBBAA")
var ErrColorChannel = errors.New("ui: color channel requires a finite value within 0..255")

// ColorEdit is an owner-delivered proposal. Apply it to the current accepted
// value, rather than a view-captured value, to merge queued channel requests.
// Straight-alpha sRGB bytes retain RGB even when alpha is zero.
type ColorEdit struct {
	Channel ColorChannel
	Byte    uint8
	RGBA    color.NRGBA
}

func (e ColorEdit) Apply(current color.NRGBA) color.NRGBA {
	switch e.Channel {
	case ColorRed:
		current.R = e.Byte
	case ColorGreen:
		current.G = e.Byte
	case ColorBlue:
		current.B = e.Byte
	case ColorAlpha:
		current.A = e.Byte
	case ColorHex:
		return e.RGBA
	}
	return current
}
func ColorHexString(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// ParseColorHex accepts four bounded ASCII forms, with optional surrounding
// whitespace. Six/three digit forms are opaque. No named colors or implicit
// premultiplication/conversion is performed.
func ParseColorHex(text string) (color.NRGBA, error) {
	if len(text) > 32 {
		return color.NRGBA{}, ErrColorHex
	}
	text = strings.TrimSpace(text)
	if len(text) < 1 || text[0] != '#' {
		return color.NRGBA{}, ErrColorHex
	}
	text = text[1:]
	if len(text) != 3 && len(text) != 4 && len(text) != 6 && len(text) != 8 {
		return color.NRGBA{}, ErrColorHex
	}
	nibble := func(c byte) (uint8, bool) {
		switch {
		case c >= '0' && c <= '9':
			return c - '0', true
		case c >= 'a' && c <= 'f':
			return c - 'a' + 10, true
		case c >= 'A' && c <= 'F':
			return c - 'A' + 10, true
		}
		return 0, false
	}
	bytes := [4]uint8{0, 0, 0, 255}
	short := len(text) < 5
	count := len(text)
	if !short {
		count /= 2
	}
	for i := 0; i < count; i++ {
		position := i
		if !short {
			position *= 2
		}
		a, ok := nibble(text[position])
		if !ok {
			return color.NRGBA{}, ErrColorHex
		}
		b := a
		if !short {
			b, ok = nibble(text[position+1])
			if !ok {
				return color.NRGBA{}, ErrColorHex
			}
		}
		bytes[i] = a*16 + b
	}
	return color.NRGBA{R: bytes[0], G: bytes[1], B: bytes[2], A: bytes[3]}, nil
}

// ColorFieldID returns the generated managed input ID. The root owns the
// root+"/color/" namespace; ordinary Node validation detects caller collisions.
func ColorFieldID(root string, channel ColorChannel) string {
	return root + "/color/" + string(channel)
}

// ColorPicker is an owner-held reusable inline component. View uses ordinary
// managed nodes: four byte sliders, a native text field and a checkerboard
// preview. Retain this object to reuse unchanged immutable preview pixels.
// Change applies/rejects a proposal; Invalid reports invalid hex/channel input.
// ValidationMessage is caller-owned display text. Nil Change makes it read-only.
// A picker may be copied (its preview is immutable), but one instance must not
// be mutated or rendered concurrently by independent application owners.
type ColorPicker struct {
	ID, Label         string
	Color             color.NRGBA
	Disabled          bool
	Change            func(ColorEdit)
	Invalid           func(error)
	ValidationMessage string
	Style             Style
	preview           *canvas.TiledImage
	previewColor      color.NRGBA
	previewValid      bool
}

func (p *ColorPicker) View() Node {
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Color"
	}
	root := Node{ID: p.ID, Kind: FormGroup, Text: label, Disabled: p.Disabled || p.Change == nil, Style: p.Style, Accessibility: Accessibility{Name: label}}
	root.Children = append(root.Children, Node{ID: p.ID + "/color/preview-row", Kind: Row, Grow: true, Style: Style{Justify: Some(AlignStart)}, Children: []Node{
		{ID: p.ID + "/color/preview", Kind: Image, Image: p.swatch(), Text: "Color preview " + ColorHexString(p.Color), Accessibility: Accessibility{Name: label + " preview " + ColorHexString(p.Color)}, Style: Style{MinSize: Some(Size{Width: 32, Height: 32}), MaxSize: Some(Size{Width: 32, Height: 32})}},
		{ID: p.ID + "/color/value", Kind: Label, Text: ColorHexString(p.Color)},
	}})
	fields := []struct {
		channel ColorChannel
		name    string
		value   uint8
	}{{ColorRed, "Red", p.Color.R}, {ColorGreen, "Green", p.Color.G}, {ColorBlue, "Blue", p.Color.B}, {ColorAlpha, "Alpha", p.Color.A}}
	for _, field := range fields {
		channel := field.channel
		root.Children = append(root.Children, Node{ID: ColorFieldID(p.ID, channel) + "/row", Kind: Row, Children: []Node{
			{ID: ColorFieldID(p.ID, channel) + "/label", Kind: Label, Text: field.name, Style: Style{MinSize: Some(Size{Width: 48})}},
			{ID: ColorFieldID(p.ID, channel), Kind: Slider, Number: float32(field.value), Numeric: Numeric{Min: 0, Max: 255, Step: 1}, Grow: true, MinWidth: 160, Accessibility: Accessibility{Name: label + " " + field.name, Help: "Byte value from 0 to 255; arrow keys change one step"}, Change: func(e ChangeEvent) {
				if p.Disabled || p.Change == nil {
					return
				}
				if math.IsNaN(float64(e.Number)) || math.IsInf(float64(e.Number), 0) || e.Number < 0 || e.Number > 255 {
					p.invalid(ErrColorChannel)
					return
				}
				p.Change(ColorEdit{Channel: channel, Byte: uint8(math.Round(float64(e.Number)))})
			}},
		}})
	}
	root.Children = append(root.Children, Node{ID: p.ID + "/color/hex-row", Kind: Row, Children: []Node{
		{ID: p.ID + "/color/hex-label", Kind: Label, Text: "Hex", Style: Style{MinSize: Some(Size{Width: 48})}},
		{ID: ColorFieldID(p.ID, ColorHex), Kind: TextField, Text: ColorHexString(p.Color), Grow: true, MinWidth: 160, Accessibility: Accessibility{Name: label + " hex", Help: "Enter #RGB, #RGBA, #RRGGBB or #RRGGBBAA"}, Change: func(e ChangeEvent) {
			if p.Disabled || p.Change == nil {
				return
			}
			value, err := ParseColorHex(e.Text)
			if err != nil {
				p.invalid(err)
				return
			}
			p.Change(ColorEdit{Channel: ColorHex, RGBA: value})
		}},
	}})
	if p.ValidationMessage != "" {
		root.Children = append(root.Children, Node{ID: p.ID + "/color/validation", Kind: Label, Text: p.ValidationMessage, Accessibility: Accessibility{Name: p.ValidationMessage}})
	}
	return root
}
func (p *ColorPicker) invalid(err error) {
	if p.Invalid != nil {
		p.Invalid(err)
	}
}
func (p *ColorPicker) swatch() *canvas.TiledImage {
	if p.previewValid && p.previewColor == p.Color {
		return p.preview
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	c := p.Color
	for y := range 32 {
		for x := range 32 {
			background := uint16(240)
			if (x/8+y/8)%2 != 0 {
				background = 176
			}
			blend := func(v uint8) uint8 { return uint8((uint16(v)*uint16(c.A) + background*(255-uint16(c.A)) + 127) / 255) }
			pixels.SetNRGBA(x, y, color.NRGBA{R: blend(c.R), G: blend(c.G), B: blend(c.B), A: 255})
		}
	}
	var source *canvas.TiledImage
	var err error
	if p.preview == nil {
		source, err = canvas.NewTiledImage(pixels)
	} else {
		source, err = p.preview.WithPatch(pixels.Bounds(), pixels)
	}
	// The source is always a locally allocated 32x32 NRGBA image in bounds.
	if err != nil {
		panic(err)
	}
	p.preview, p.previewColor, p.previewValid = source, p.Color, true
	return source
}
