package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	// MaxFileDropPaths and MaxFileDropBytes bound a delivered file-drop payload.
	// They do not bound memory already allocated by the native drag source.
	MaxFileDropPaths = 256
	MaxFileDropBytes = 1 << 20
)

// Point uses device-independent pixels.
type Point struct{ X, Y float32 }

// FileDropEvent describes external file references. Position is relative to the
// target's unclipped outer layout origin, including margin and padding, in Dp.
// Paths is a private slice for this callback; spelling, order and duplicates are
// preserved. No file content is read and existence is not checked.
//
// Node.FileDrop runs on the application owner. Returning true handles the drop
// and invalidates WindowID; false lets an ancestor try, with a fresh payload.
// A declining callback should leave model state unchanged. Callbacks must not
// nest Application.Do/Drain or block on I/O. Acceptance does not move or delete
// source files. The current macOS backend acknowledges the OS drag before model
// admission: false controls framework propagation, not Finder feedback.
type FileDropEvent struct {
	ID, WindowID string
	Paths        []string
	Position     Point
}

// CopyFileDropPaths validates an entire batch before copying its slice. Paths
// must be absolute for the current host, nonempty valid UTF-8 without NUL. No
// normalization, URI decoding, file access, or partial admission is performed.
func CopyFileDropPaths(paths []string) ([]string, error) {
	if len(paths) == 0 || len(paths) > MaxFileDropPaths {
		return nil, fmt.Errorf("ui: file drop requires 1..%d paths", MaxFileDropPaths)
	}
	total := 0
	for _, path := range paths {
		if len(path) > MaxFileDropBytes-total {
			return nil, fmt.Errorf("ui: file drop exceeds %d bytes", MaxFileDropBytes)
		}
		total += len(path)
		if !utf8.ValidString(path) || strings.IndexByte(path, 0) >= 0 || !filepath.IsAbs(path) {
			return nil, fmt.Errorf("ui: file drop requires absolute UTF-8 paths without NUL")
		}
	}
	return append([]string(nil), paths...), nil
}
