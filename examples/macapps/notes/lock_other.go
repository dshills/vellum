//go:build !darwin

package notes

import (
	"errors"
	"os"
)

// This internal macOS app fails closed for autosave on other platforms. Its
// drawing remains usable rather than sharing an unlocked recovery slot.
func lockRecoveryDirectory(string) (*os.File, error) {
	return nil, errors.New("recovery: directory locking is unavailable on this platform")
}
