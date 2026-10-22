# Existing-backend display scheduling: preparation

The source audit confirms the existing backend uses FIFO/vsync, two drawables
and ordinary earliest-possible nontransactional presentation. It does not set a
minimum display duration or target presentation time in this path. The separate
present-command-buffer hypothesis was already tested and found insufficient;
transaction and supplied-drawable display-link variants are not repeated.

The next diagnostic observes the existing window's Core Animation display clock
through `NSView.displayLink(target:selector:)`. It leaves the Cogent/WebGPU
renderer, allocation/cache and scheduling policies unchanged. It acquires no
drawable, requests no frame and calls no GPU submit/present method. It records
only scalar timestamp, target, duration, host observation time and window/display
identity with fixed, non-evicting storage. Stop invalidates the link and releases
owned objects before application close. The private diagnostic requires macOS 14;
it does not change the SDK/application deployment target.

This differs from the earlier supplied-drawable CAMetalDisplayLink experiment:
the observer does not supply or own the drawable. Apple's view API follows the
view's display, and its timestamp/target are reported Core Animation clocks.
They do not establish optical scanout, guaranteed latch opportunities or missed
vsyncs. Retrospective matching does not imply callback delivery before GPU end.
[View API](https://developer.apple.com/documentation/appkit/nsview/displaylink%28target%3Aselector%3A%29),
[timestamp](https://developer.apple.com/documentation/quartzcore/cadisplaylink/timestamp),
[duration](https://developer.apple.com/documentation/quartzcore/cadisplaylink/duration),
[command-buffer presentation](https://developer.apple.com/documentation/metal/mtlcommandbuffer/present%28_%3A%29).
The [source audit](source-audit.json) pins the actual native sources and SDK headers.

Eight focused classifier methods and 20 negative cases pass. Missing coverage,
invalid GPU timing, callback identities/count/order, terminal states and inherited
backend errors are refused; callback gaps and lateness remain data. The native
recorder/ownership fixture passes without creating an NSApplication or window,
compiled warning-free with `-Wall -Wextra -Werror`. The original routed complete
pixel/state fixture and fresh private consumer build pass. The exact reviewed
native archive, private Core and source-detail overlays are reused, with no Rust
or SDK rebuild. Full source/log/owned-process checks, exact consumer/module/overlay
reproduction and all 8,321 SDK files verify unchanged. Existing AppKit deprecation
and duplicate-library warnings remain visible in the Go build.

The [protocol](../../../tools/macos-calibration/displayclock/PLAN.md) predeclares
one observer-off then one observer-on run of the same binary. Each retains 128
warmup and 120 measured inputs, with a common 50 ms held tail outside the cohort.
The newly approved quiet window is capped at 100 seconds overall and 40 seconds
per native case, after review and commit. Stop on first invalidity; no unchanged
retry. This is an observer-presence control, not repeated qualification, calibrated
overhead or a pure-cost causal comparison. No native capture has run yet.

Features, original backend, SDK/app, render/pacing defaults and the 33.3 ms release
gate remain unchanged. M3–M7 remain open; no production repair or release claim.

Before any desktop run, internal review found numeric JSON booleans in the native
snapshot. Explicit Boolean boxing and real native JSON-type checks now pass;
the original implementation is refused by the final fixture with an ordinary
exit, without a crash or window. The final fixture also exercises the callback
selector and identity/nil-link start-failure cleanup. The generated worker's
unconditional cleanup defer precedes startup. Prism's boundary concern led to an
exact timestamp equality test; key/focus changes remain deliberate refusals.
All initial sources/build receipts and complete review dispositions are retained;
only final v2 is eligible for execution.

The private consumer commands are `framework timed MATCH.json on off` and
`framework timed MATCH.json on on`. The fixed third argument keeps CPU-upload
and GPU-submission observation enabled in both cases; only the final argument
toggles the display clock. Disabling the other observers would remove the joins
needed for this experiment. Both complete Prism reviews and their source-backed
dispositions are retained. No unresolved actionable finding remains.
