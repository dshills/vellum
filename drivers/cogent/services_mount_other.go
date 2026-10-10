//go:build darwin || (!windows && (!linux || android)) || !cgo

package cogent

import "cogentcore.org/core/core"

func nativeMountServices(*core.Body, *platformServices) {}
func nativeRefreshServices(*core.Body)                  {}
func nativeUnmountServices(*core.Body)                  {}
