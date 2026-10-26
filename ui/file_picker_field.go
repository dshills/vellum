package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type FilePickerKind string

const (
	FilePickerFile   FilePickerKind = "file"
	FilePickerFolder FilePickerKind = "folder"
)

type FilePickerAction string

const (
	FilePickerBrowse FilePickerAction = "browse"
	FilePickerClear  FilePickerAction = "clear"
	FilePickerReveal FilePickerAction = "reveal"
	FilePickerDrop   FilePickerAction = "drop"
)

type FilePickerEvent struct {
	ID, Path string
	Kind     FilePickerKind
	Action   FilePickerAction
}

// FilePickerField is controlled/owner-only. Request owns worker I/O, metadata
// verification and admission; the component checks path spelling/suffix only.
// Browse/Reveal must use worker-side platform services. It reads no files.
type FilePickerField struct {
	ID, Label, Path string
	Kind            FilePickerKind
	Extensions      []string
	Disabled, Busy  bool
	IconImage       *canvas.TiledImage
	Request         func(FilePickerEvent) error
	Style           Style
	message         string
}

func ValidatePickerFilters(kind FilePickerKind, extensions []string) error {
	if kind != FilePickerFile && kind != FilePickerFolder {
		return fmt.Errorf("ui: invalid picker kind")
	}
	if kind == FilePickerFolder && len(extensions) > 0 {
		return fmt.Errorf("ui: folders do not accept extension filters")
	}
	if len(extensions) > 32 {
		return fmt.Errorf("ui: at most32 picker extensions")
	}
	for _, ext := range extensions {
		if ext == "" || len(ext) > 32 || !utf8.ValidString(ext) || strings.ContainsAny(ext, ". /\\\x00\r\n") {
			return fmt.Errorf("ui: invalid picker extension")
		}
	}
	return nil
}
func ValidatePickerPath(path string, kind FilePickerKind, extensions []string) error {
	if e := ValidatePickerFilters(kind, extensions); e != nil {
		return e
	}
	if _, e := CopyFileDropPaths([]string{path}); e != nil {
		return e
	}
	if len(path) > 4096 {
		return fmt.Errorf("ui: picker path exceeds4096 bytes")
	}
	if kind == FilePickerFile && len(extensions) > 0 {
		match := false
		for _, ext := range extensions {
			match = match || strings.EqualFold(strings.TrimPrefix(filepath.Ext(path), "."), ext)
		}
		if !match {
			return fmt.Errorf("ui: file extension is not accepted")
		}
	}
	return nil
}
func (f *FilePickerField) View() (Node, error) {
	if f == nil || !validTokenText(f.ID) {
		return Node{}, fmt.Errorf("ui: picker requires an ID")
	}
	kind := f.Kind
	if kind == "" {
		kind = FilePickerFile
	}
	if f.Path != "" {
		if err := ValidatePickerPath(f.Path, kind, f.Extensions); err != nil {
			return Node{}, err
		}
	} else {
		if err := ValidatePickerFilters(kind, f.Extensions); err != nil {
			return Node{}, err
		}
	}
	id, value := f.ID, f.Path
	live := func() bool {
		current := f.Kind
		if current == "" {
			current = FilePickerFile
		}
		return f.ID == id && current == kind && !f.Disabled && !f.Busy && f.Request != nil
	}
	request := func(action FilePickerAction, path string) {
		if !live() || action != FilePickerDrop && f.Path != value {
			return
		}
		if action == FilePickerDrop {
			if err := ValidatePickerPath(path, kind, f.Extensions); err != nil {
				f.message = boundedWidgetMessage(err.Error())
				return
			}
		}
		err := f.Request(FilePickerEvent{ID: id, Path: path, Kind: kind, Action: action})
		f.message = ""
		if err != nil {
			f.message = boundedWidgetMessage(err.Error())
		}
	}
	label := f.Label
	if label == "" {
		label = "File"
		if kind == FilePickerFolder {
			label = "Folder"
		}
	}
	caption := "Choose a " + string(kind) + "…"
	if value != "" {
		caption = filepath.Base(value)
	}
	field := Node{ID: id + "/picker/path", Kind: Button, Text: caption, Grow: true, IconImage: f.IconImage, Disabled: f.Request == nil, Accessibility: Accessibility{Name: label, Help: value}, Tooltip: value, Activate: func() { request(FilePickerBrowse, f.Path) }}
	if field.IconImage == nil {
		field.Icon = SymbolFolder
	}
	row := Node{ID: id + "/picker/row", Kind: Row, Children: []Node{field, {ID: id + "/picker/browse", Kind: Button, Text: "Browse…", Disabled: f.Request == nil, Activate: func() { request(FilePickerBrowse, f.Path) }}, {ID: id + "/picker/reveal", Kind: Button, Text: "Reveal", Disabled: value == "" || f.Request == nil, Activate: func() { request(FilePickerReveal, f.Path) }}, {ID: id + "/picker/clear", Kind: Button, Icon: SymbolClose, Appearance: ButtonTool, Accessibility: Accessibility{Name: "Clear " + label}, Disabled: value == "" || f.Request == nil, Activate: func() { request(FilePickerClear, "") }}}}
	root := Node{ID: id, Kind: Column, Disabled: f.Disabled || f.Busy, Style: f.Style, Children: []Node{{ID: id + "/picker/label", Kind: Label, Text: label}, row}, FileDrop: func(e FileDropEvent) bool {
		if !live() {
			return false
		}
		if len(e.Paths) != 1 {
			f.message = "Choose one file or folder"
			return true
		}
		request(FilePickerDrop, e.Paths[0])
		return true
	}}
	if f.message != "" {
		root.Children = append(root.Children, Node{ID: id + "/picker/message", Kind: Label, Text: f.message})
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
