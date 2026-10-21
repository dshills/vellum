//go:build !windows

package document

import "os"

func syncParentDirectory(directory *os.File) error { return directory.Sync() }
