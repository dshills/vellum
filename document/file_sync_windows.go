package document

import (
	"errors"
	"os"
	"syscall"
)

// Windows FlushFileBuffers rejects read-only directory handles with
// ERROR_ACCESS_DENIED. Contents were synced before rename; the portable API
// cannot promise directory durability. Preserve unexpected errors (for example
// a closed handle) rather than turning every post-rename failure into success.
func syncParentDirectory(directory *os.File) error {
	err := directory.Sync()
	if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		return nil
	}
	return err
}
