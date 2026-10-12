//go:build darwin && cgo && !offscreen

package cogent

/*
#include <stdint.h>
int vellumShutdownOnMain(void);
*/
import "C"

import (
	"cogentcore.org/core/system/driver/desktop"
	"fmt"
	"github.com/dshills/vellum/app"
)

func configureServiceCleanup(s *platformServices) (func(), error) {
	return desktop.TheApp.AddMainCleanup(func() {
		s.stopped.Store(true)
		stopNativeCallbacks()
		if C.vellumShutdownOnMain() != 0 {
			err := fmt.Errorf("cogent: native service cleanup failed: %w", app.ErrUnavailable)
			s.cleanupErr.Store(&err)
		}
	})
}
