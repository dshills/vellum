# Native presentation feedback (stage 88)

This optional observation supports the owner's measured 60 Hz release policy.
A Vellum-owned CAMetalLayer subclass calls the original nextDrawable path and
adds a Metal presented handler only after a bounded recorder reserves a sample.
The default path samples no clock and installs no handlers. No global production
method replacement, extra acquire/present, readback or queue wait is introduced.

The drawable callback carries only scalar subscription/sequence IDs and uses the
callback's drawable argument. It retains no window, layer, Go pointer or app.
Go registry tokens are never reused; close withdraws the token before native
surface/view teardown. The sink independently revokes its scope, including for
callbacks looked up just before removal. Registry and native stamp locks are
released before invoking sinks. Unresolved records retire without claiming a
presentation. Nil drawables, zero-time callbacks, invalid clocks, exhaustion and
pending feedback remain explicit. Drawable ID zero is valid.

CACurrentMediaTime and MTLDrawable.presentedTime share the native host-time
boundary. Receipt sampling happens before FIFO send without a render-window
lock or Scene/Stage lookup. Receipts currently cover routed grid/tree requests;
they do not establish hardware input-to-photon latency. macOS 10.15.4 or newer
is checked when installing the observer. Other platforms refuse installation.

Measured tracing snapshots generation and armed state at composition start;
late owner work cannot reclassify an earlier frame. Native composition ordinals
cover ordinary, measured and cached-resize composition. An armed stamp alone is
not a successful trace: qualification must join the matching completed/error-free
FrameMetrics generation and actual presentation, count missing feedback and
retain all intervals. The recorder is bounded by counts, not a process-memory
budget, and is disabled for long endurance runs unless separately sized.

The pure Go tests cover shared capacity, detached snapshots, clock validation,
callback order, nil acquisition and close races. The native check.py compiles the
actual override with controlled drawable stand-ins, verifying the original
acquire call, disabled/refused registration, pre-arm metadata, zero time/ID and
late scalar callbacks. Its temporary method replacement exists only inside that
owned fixture process. It creates no NSApplication/window/device/real drawable.
The separate property-ownership fixture links this implementation and preserves
view ownership across local pool drainage. None of these tests qualifies actual
Metal callback delivery, hardware input, refresh pacing or native teardown; those
require a quiet live campaign on the declared Mac configuration.
