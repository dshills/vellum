package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"path/filepath"
	"strings"
)

type FileDropZoneEvent struct {
	ID, WindowID string
	Paths        []string
}

// FileDropZone owns only bounded last-result feedback. Application Change owns
// file admission/I/O, returns an error to reject the entire batch, and must not
// partially mutate state before rejection. All callbacks/methods run on one
// owner. Choose requests a worker-side chooser; feed its result to ProposeFiles.
// Extension checks are filename hints, never content/security verification.
type FileDropZone struct {
	ID, Label, Detail              string
	Dark, Disabled, ChooseDisabled bool
	Extensions                     []string
	MaxFiles                       int
	IconImage                      *canvas.TiledImage
	Change                         func(FileDropZoneEvent) error
	Choose                         func()
	Style                          Style
	feedback                       string
	rejected                       bool
}

func DropZoneChooseID(id string) string { return id + "/drop/choose" }
func (z *FileDropZone) Feedback() string {
	if z == nil {
		return ""
	}
	return z.feedback
}
func boundedWidgetMessage(s string) string {
	if len(s) > 512 {
		s = s[:512]
	}
	s = strings.ToValidUTF8(s, "?")
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}
func (z *FileDropZone) config() error {
	if z == nil || z.ID == "" || len(z.Extensions) > 32 || z.MaxFiles < 0 || z.MaxFiles > MaxFileDropPaths {
		return fmt.Errorf("ui: invalid file drop zone configuration")
	}
	seen := map[string]bool{}
	for _, ext := range z.Extensions {
		lower := strings.ToLower(ext)
		if len(ext) == 0 || len(ext) > 32 || seen[lower] {
			return fmt.Errorf("ui: invalid or duplicate file extension")
		}
		for _, c := range lower {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
				return fmt.Errorf("ui: extensions omit dots and use ASCII letters/digits")
			}
		}
		seen[lower] = true
	}
	return nil
}

// ProposeFiles validates chooser results through the identical drop path. No
// file is opened, read, normalized or retained. Order/duplicates are preserved.
func (z *FileDropZone) ProposeFiles(paths []string) error { return z.propose(paths, "") }
func (z *FileDropZone) propose(paths []string, window string) (err error) {
	if err = z.config(); err != nil {
		return err
	}
	if z.Disabled || z.Change == nil {
		return fmt.Errorf("ui: file drop zone unavailable")
	}
	id := z.ID
	defer func() {
		if z.ID != id {
			return
		}
		z.rejected = err != nil
		if err != nil {
			z.feedback = boundedWidgetMessage(err.Error())
		} else {
			z.feedback = fmt.Sprintf("Accepted %d file(s)", len(paths))
		}
	}()
	copied, err := CopyFileDropPaths(paths)
	if err != nil {
		return err
	}
	limit := z.MaxFiles
	if limit == 0 {
		limit = 16
	}
	if len(copied) > limit {
		return fmt.Errorf("Choose no more than %d files", limit)
	}
	for _, path := range copied {
		if len(z.Extensions) > 0 {
			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			ok := false
			for _, allowed := range z.Extensions {
				ok = ok || strings.EqualFold(ext, allowed)
			}
			if !ok {
				return fmt.Errorf("File type is not accepted; choose %s", strings.Join(z.Extensions, ", "))
			}
		}
	}
	return z.Change(FileDropZoneEvent{ID: id, WindowID: window, Paths: copied})
}
func (z *FileDropZone) View() (Node, error) {
	if err := z.config(); err != nil {
		return Node{}, err
	}
	id := z.ID
	label := z.Label
	if label == "" {
		label = "Drop files here"
	}
	detail := z.Detail
	if detail == "" {
		detail = "Drag files from Finder"
		if len(z.Extensions) > 0 {
			detail += " · " + strings.Join(z.Extensions, ", ")
		}
	}
	p := MacPalette(z.Dark, color.NRGBA{})
	line := p.Line
	if z.feedback != "" {
		line = p.Accent
	}
	if z.rejected {
		line = color.NRGBA{R: 210, G: 45, B: 45, A: 255}
		if z.Dark {
			line = color.NRGBA{R: 255, G: 110, B: 110, A: 255}
		}
	}
	icon := Node{ID: id + "/drop/icon", Kind: Icon, Icon: SymbolFolder, Style: Style{MinSize: Some(Size{Width: 36, Height: 36}), MaxSize: Some(Size{Width: 36, Height: 36})}}
	if z.IconImage != nil {
		icon.Kind = Image
		icon.Icon = ""
		icon.Image = z.IconImage
		icon.Fit = ImageContain
		icon.Accessibility.Name = "File drop target"
	}
	root := Node{ID: id, Kind: Column, Disabled: z.Disabled || z.Change == nil, Accessibility: Accessibility{Name: label, Help: detail}, Style: Style{Padding: Some(UniformInsets(18)), Gap: Some(float32(9)), Align: Some(AlignCenter), Background: Some(p.Panel), Foreground: Some(p.Ink), Border: Some(Border{Width: 1, Color: line}), CornerRadius: Some(float32(9)), GrowX: Some(float32(1))}.Overlay(z.Style), FileDrop: func(e FileDropEvent) bool {
		if z.ID != id || z.Disabled || z.Change == nil {
			return false
		}
		_ = z.propose(e.Paths, e.WindowID)
		return true
	}}
	root.Children = []Node{icon, {ID: id + "/drop/label", Kind: Label, Text: label, Style: Style{Typography: Some(MacTypography(14, 600))}}, {ID: id + "/drop/detail", Kind: Label, Text: detail, Style: Style{Foreground: Some(p.Secondary)}}}
	if z.Choose != nil {
		root.Children = append(root.Children, Node{ID: DropZoneChooseID(id), Kind: Button, Text: "Choose File…", Disabled: z.ChooseDisabled, Activate: func() {
			if z.ID == id && !z.Disabled && !z.ChooseDisabled && z.Change != nil && z.Choose != nil {
				z.Choose()
			}
		}})
	}
	if z.feedback != "" {
		root.Children = append(root.Children, Node{ID: id + "/drop/feedback", Kind: Label, Text: z.feedback, Accessibility: Accessibility{Name: z.feedback}, Style: Style{Foreground: Some(line)}})
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
