package notes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dshills/vellum/document"
)

const recoveryLockFilename = ".notes.lock"

// Repository owns one private local notebook/recovery slot. Construct before Run;
// Close only after document workers have retired. The OS retires a crashed lock.
type Repository struct {
	exists                        bool
	Directory, Path, RecoveryPath string
	lock                          *os.File
}

func OpenRepository(directory string) (*Repository, error) {
	if !filepath.IsAbs(directory) {
		return nil, fmt.Errorf("notes: data-dir must be an absolute path")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("notes: private directory (0700) required")
	}
	lock, err := lockRecoveryDirectory(directory)
	if err != nil {
		return nil, err
	}
	return &Repository{Directory: directory, Path: filepath.Join(directory, "Notebook.vellumnotes"), RecoveryPath: filepath.Join(directory, "Recovery.vellumnotes"), lock: lock}, nil
}
func (r *Repository) Close() error {
	if r == nil || r.lock == nil {
		return nil
	}
	err := r.lock.Close()
	r.lock = nil
	return err
}

// Load runs before the UI owner starts. Invalid data is preserved and refused.
func (r *Repository) Load(m *Model) (*Recovery, error) {
	store := document.FileStore{}
	data, err := store.Read(context.Background(), r.Path, MaxNotebookBytes)
	if err == nil {
		b, e := DecodeNotebook(data)
		if e != nil {
			return nil, e
		}
		if e = m.journal.Replace(true, func() error { return m.installNotebook(b) }); e != nil {
			return nil, e
		}
		r.exists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	data, err = store.Read(context.Background(), r.RecoveryPath, MaxNotebookBytes)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	recovery, err := decodeRecovery(data)
	if err != nil {
		return nil, err
	}
	if recovery.Notebook == nil {
		return nil, nil
	}
	return &recovery, nil
}

// A nil Notebook is an atomic clean marker; recovery never marks the main journal saved.
type Recovery struct {
	Version  int       `json:"version"`
	Notebook *Notebook `json:"notebook"`
}

func encodeRecovery(r Recovery) ([]byte, error) {
	if r.Version != 1 {
		return nil, ErrInvalid
	}
	if r.Notebook != nil {
		if err := r.Notebook.Validate(); err != nil {
			return nil, err
		}
	}
	return json.MarshalIndent(r, "", "  ")
}
func decodeRecovery(data []byte) (Recovery, error) {
	var r Recovery
	if err := decodeDocument(data, &r); err != nil {
		return r, err
	}
	if r.Version != 1 {
		return r, ErrInvalid
	}
	if r.Notebook != nil {
		return r, r.Notebook.Validate()
	}
	return r, nil
}
