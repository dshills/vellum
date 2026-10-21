package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"slices"
	"unicode/utf8"
)

const fillStyleByteLimit = 256

type fillStyleClipboard struct {
	Format  string `json:"format"`
	Version int    `json:"version"`
	Color   string `json:"color"`
}

func encodeFillStyle(value color.RGBA) (string, error) {
	if value.A != 255 {
		return "", errors.New("fill style requires an opaque RGB color")
	}
	data, err := json.Marshal(fillStyleClipboard{"vellum-placement-fill", 1, fmt.Sprintf("#%02x%02x%02x", value.R, value.G, value.B)})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// This separate payload cannot be mistaken for copied components or plain hex
// text. Strict fields and bounds keep unrelated clipboard contents inert.
func decodeFillStyle(text string) (color.RGBA, error) {
	invalid := errors.New("clipboard does not contain a supported opaque Vellum fill style")
	if len(text) == 0 || len(text) > fillStyleByteLimit || !utf8.ValidString(text) {
		return color.RGBA{}, invalid
	}
	decoder := json.NewDecoder(bytes.NewBufferString(text))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return color.RGBA{}, invalid
	}
	var f fillStyleClipboard
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		name, ok := token.(string)
		if err != nil || !ok || seen[name] {
			return color.RGBA{}, invalid
		}
		seen[name] = true
		switch name {
		case "format":
			err = decoder.Decode(&f.Format)
		case "version":
			err = decoder.Decode(&f.Version)
		case "color":
			err = decoder.Decode(&f.Color)
		default:
			return color.RGBA{}, invalid
		}
		if err != nil {
			return color.RGBA{}, invalid
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(seen) != 3 || f.Format != "vellum-placement-fill" || f.Version != 1 || len(f.Color) != 7 || f.Color[0] != '#' {
		return color.RGBA{}, invalid
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return color.RGBA{}, invalid
	}
	rgb, err := hex.DecodeString(f.Color[1:])
	if err != nil || len(rgb) != 3 {
		return color.RGBA{}, invalid
	}
	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}, nil
}

func (m *model) copyFillStyle() (string, error) {
	if err := m.editAvailable(); err != nil {
		return "", err
	}
	index := m.selectedIndex()
	if index < 0 || !slices.Contains(m.selectedIDs(), m.scene.Selected) {
		return "", errNoSelection
	}
	return encodeFillStyle(m.scene.Objects[index].Color)
}

func (m *model) pasteFillStyle(text string) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	value, err := decodeFillStyle(text)
	if err != nil {
		return err
	}
	ids := m.selectedIDs()
	after := slices.Clone(m.scene.Objects)
	for i := range after {
		if slices.Contains(ids, after[i].ID) {
			after[i].Color = value
		}
	}
	// A style change can increase saved-file size; preflight the same codec
	// used by Save before admitting the single history operation.
	proposed := *m
	scene := *m.scene
	scene.Objects = after
	proposed.scene = &scene
	image, err := drawingCodec(&proposed).Capture()
	if err != nil {
		return err
	}
	if _, err = drawingCodec(&proposed).Encode(image); err != nil {
		return err
	}
	if err = m.applyObjects("Paste fill style", after, m.scene.Selected); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Pasted fill style to %d components", len(ids))
	return nil
}
