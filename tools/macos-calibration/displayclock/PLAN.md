# Existing-backend display scheduling investigation

Keep the Cogent/WebGPU backend, features, SDK/application, allocation/cache,
render/pacing defaults and the 33.3 ms release gate unchanged. Source audit shows
FIFO/vsync, two drawables and earliest-possible nontransactional presentation.
The separate-present topology hypothesis is already closed as insufficient.
Do not repeat those tests or introduce a direct Metal backend.

Reuse the reviewed joined-backend archive and exact private Core/three source
information overlays. Add a private NSView CADisplayLink observer (macOS 14+),
which owns no drawable, submits no GPU work, requests no frame and changes no
layer/display-link rate/presentation setting. Only record scalar native host
observation time, display-link timestamp/target/duration and window/display IDs.
Main-thread-only state, fixed8192non-evicting slots, one configure/start/stop,
explicit invalid/overflow state and invalidation before release bound ownership.
Use the view's default native display-link cadence. Timestamp/target are reported
Core Animation times, not independent optical scanout or guaranteed latch times.

One observer-off then one observer-on run of the same new binary/library;
248 inputs each (128 warmup,120 measured,120ms schedule), the existing routed
view/commands, receipts, source/raster/upload/native GPU/drawable joins, strict
foreground/configuration/geometry accounting and parent validation unchanged.
Both cases hold an identical50ms tail after the last presented state, outside
measurement, before stopping the observer and closing the window.

Off records no callback. On retains all callbacks and requires the timestamp
coverage to enclose every joined final-present GPU end and shown time. Correlate
nearest reported timestamps and targets without inventing missed-vsync counts,
physical scanout, compositor-only time or cause. Preserve callback gaps and late
callback observations: these are data, not reasons to discard samples. Do not fit
a timestamp lattice from presentation feedback and then call it independent proof.
The off/on pair is an observer-presence control, not repeated performance
qualification or a pure-cost/causal A/B proof. Do not pool or add quantiles.

First run focused scalar/ownership fixtures (no NSApplication/window), original
routed full pixels, classifier refusals, native consumer build and source/SDK/log
checks. Review with Prism Anthropic Claude Sonnet5.5 and commit before desktop
use. Then one newly approved quiet window, at most100s overall/40s per native
case, exclusive lease and owned executor. Stop on first invalidity; seal all raw
startup/callback/tail data with no unchanged retry. Review/commit results
separately. No ordinary-build equivalence, physical-input/displayed-pixel/GPU
allocation balance, shipping repair or release acceptance follows.
