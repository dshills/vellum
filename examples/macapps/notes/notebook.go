package notes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dshills/vellum/editor"
)

const MaxNotebookBytes int64 = 16 << 20

// Notebook is a detached, versioned application document. History is session-local.
type Notebook struct {
	Version  int    `json:"version"`
	NextID   uint64 `json:"next_id"`
	Notes    []Note `json:"notes"`
	Selected string `json:"selected"`
	Folder   string `json:"folder"`
}

func (b Notebook) Validate() error {
	if b.Version != 1 || len(b.Notes) > MaxNotes || b.NextID < 10 || !validFolder(b.Folder) {
		return ErrInvalid
	}
	ids := map[string]bool{}
	for _, n := range b.Notes {
		if !strings.HasPrefix(n.ID, "n") {
			return ErrInvalid
		}
		id, err := strconv.ParseUint(strings.TrimPrefix(n.ID, "n"), 10, 64)
		if err != nil || id == 0 || "n"+strconv.FormatUint(id, 10) != n.ID || id >= b.NextID || ids[n.ID] || !validFolder(n.FolderID) || n.FolderID == "all" || !validText(n.Title, MaxTitleBytes) || !validText(n.Body, MaxBodyBytes) {
			return ErrInvalid
		}
		ids[n.ID] = true
	}
	if b.Selected != "" && !ids[b.Selected] {
		return ErrInvalid
	}
	return nil
}
func validFolder(id string) bool {
	return id == "all" || id == "notes" || id == "projects" || id == "personal"
}
func (m *Model) CaptureNotebook() (Notebook, error) {
	if err := m.flush(); err != nil {
		return Notebook{}, err
	}
	return Notebook{Version: 1, NextID: m.nextID, Notes: m.Items(), Selected: m.selected, Folder: m.selectedFolder}, nil
}

// installNotebook validates and allocates everything before changing live state.
// The caller owns the journal's replacement transaction.
func (m *Model) installNotebook(b Notebook) error {
	if err := b.Validate(); err != nil {
		return err
	}
	items := make([]*note, 0, len(b.Notes))
	for _, v := range b.Notes {
		e, err := editor.New(v.Body, editor.Config{MaxBytes: MaxBodyBytes, HistoryLimit: 1})
		if err != nil {
			return err
		}
		items = append(items, &note{value: v, editor: e})
	}
	m.notes = items
	m.nextID = b.NextID
	m.selected = b.Selected
	m.selectedFolder = b.Folder
	m.query = ""
	m.focus.ID = ""
	m.touched = ""
	return nil
}
func EncodeNotebook(b Notebook) ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(b, "", "  ")
}
func DecodeNotebook(data []byte) (Notebook, error) {
	var b Notebook
	if err := decodeDocument(data, &b); err != nil {
		return b, err
	}
	return b, b.Validate()
}

// Reject duplicate keys, excessive nesting, unknown fields, trailing objects and
// invalid encodings before a decoded document can reach the owner.
func decodeDocument(data []byte, target any) error {
	if int64(len(data)) > MaxNotebookBytes || !validText(string(data), int(MaxNotebookBytes)) {
		return ErrInvalid
	}
	scan := json.NewDecoder(bytes.NewReader(data))
	var value func(int) error
	value = func(depth int) error {
		if depth > 12 {
			return ErrInvalid
		}
		t, err := scan.Token()
		if err != nil {
			return err
		}
		d, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch d {
		case '{':
			seen := map[string]bool{}
			for scan.More() {
				key, err := scan.Token()
				if err != nil {
					return err
				}
				s, ok := key.(string)
				s = strings.ToLower(s)
				if !ok || seen[s] {
					return ErrInvalid
				}
				seen[s] = true
				if err = value(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for scan.More() {
				if err = value(depth + 1); err != nil {
					return err
				}
			}
		default:
			return ErrInvalid
		}
		_, err = scan.Token()
		return err
	}
	if err := value(0); err != nil {
		return fmt.Errorf("notes: invalid document: %w", err)
	}
	if _, err := scan.Token(); err != io.EOF {
		return ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("notes: decode: %w", err)
	}
	return nil
}
