//go:build darwin && vellum_native_checks

package cogent

import "cogentcore.org/core/base/nativeactivity"

// NativeActivityCheck is a detached instrumentation-build observation. It covers
// this framework's admitted desktop empty-event posts and held surface Present
// attempts across windows. Direct third-party calls, OS wakeups, GPU completion
// and compositor presentation are unavailable. Available does not mean the
// native backend is initialized. No capture/dispatch/clock/setting is changed.
type NativeActivityCheck = nativeactivity.Counters

// NativeActivitySnapshot reads counters without posting native/owner work.
// Values are process-local and loaded independently. Correlate ProcessID with
// the owned process receipt; PID equality alone is not lifetime provenance.
func NativeActivitySnapshot() NativeActivityCheck { return nativeactivity.Read() }
