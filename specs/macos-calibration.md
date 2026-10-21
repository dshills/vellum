# Matched macOS native/framework calibration

Authorized 21 October 2026. Features and the existing 33.3 ms / 60 Hz release
criteria remain frozen. Calibration cannot pass, relabel or waive a release gate.

## Predeclared comparison

Use the immutable bd3cbc9 SDK, its source-built Metal archive and default
scheduling/rendering. Compare a standalone AppKit/Metal reference with a Vellum
consumer using public `NewMeasuredDesktopSession` and `PresentationMetrics` APIs.
Both process a synthetic Space activation, toggle one opaque red/green rectangle
on white, and present the same framebuffer extent and rectangle. Native uses
cached reference textures; Vellum performs its normal owner/reconcile/paint/upload.
This measures the response of two implementations of the same small visible
operation. It is not an equal-algorithm CPU benchmark or a professional-workflow
qualification. Exclude the previous after-acquisition display-link reference.

Both receipts use the native host clock immediately before publication to the
application's originating input FIFO. Native posts an internal NSEvent key-down; Vellum routes a Core KeyChord
through its normal public command registry and existing FIFO. Record these input-method differences. Neither
includes hardware/OS input delivery; do not call the result input-to-photon or
subtract the native number from framework latency to claim pure framework cost.
Join every receipt to its actual changed model revision and armed presentation.

A bounded framework setup-only pilot determines actual client framebuffer and
payload bounds from completed layout, not timing results. Lock these values and
native pixel format, scale, drawable count, vsync and transaction settings in a
hashed match file before either timed implementation runs. Native must reproduce
them and every timed framework run must match; mismatch refuses the pair. Both
use next-drawable timeout disabled. Retain start/end actual foreground PID,
active/key/visible state and all pre-cohort-to-end focus-loss/foreign-activation
notifications. Framework monitoring begins after natural startup readiness,
before timed warmups; no complete startup foreground qualification is claimed.

## Populations, ordering and refusal

Each process retains all startup acquisitions and requires positive actual native
presentation before its first warmup receipt. The framework waits for natural
positive startup feedback before its one setup-only `Frame` barrier; that barrier
and its actual tags remain explicit in setup accounting. Only pre-receipt,
unarmed/tag-zero/generation-zero startup skips may be missing; a lost forced or
measured armed frame is invalid. No forced repaint substitutes for a timed input. Then use 128 warmups and 120
measured state changes, on fixed 120 ms deadlines. Every request must complete
CPU/state validation and actual presentation before the next request; no merged,
missing or replaced state is eligible. This deliberately sparse input calibration
does not test sustained 60 Hz cadence. Retain request lateness and all timestamps.
Refuse a worker wake more than 20 ms late; this is a calibration schedule bound,
not a revision of the release latency or resize limits. Report actual FIFO receipt
phase separately; framework request admission includes its existing control path.

Predeclare five pairs, orders N/F, F/N, N/F, F/N, N/F. Record per-run
p50/p95/p99/max, acquisition-to-presentation and receipt-to-acquisition components,
plus the five differences of per-run p95 values. Do not pool frame samples,
interpret differences of quantiles as quantiles of differences, extrapolate an
OS floor, or select a new budget from the most favorable run. Five pairs assess
repeatability for this fixture, not whole-framework release acceptance.

Use one exclusive native lease and an explicitly admitted quiet desktop window
of at most eight minutes; each timed process has a 45-second watchdog and the
setup pilot a 15-second watchdog. Stop on the first invalid capture, deadline
miss, lost state, focus/configuration change, source drift, process failure or
shutdown failure. Numeric latency over 33.3 ms is an expected observation in this
calibration and does not alone stop collection. This is a new declared purpose;
the earlier failed release campaigns stay sealed and failed. No blind retries.

## Checkpoints

1. Commit this protocol and diagnostic-only fixtures after focused model/render,
   classifier negative controls, optimized builds, internal review and Prism with
   Anthropic Claude Sonnet 5.5. No native run is implied by an offscreen pass.
2. Run the bounded pilot/pairs when the desktop is available, retain failures,
   validate provenance/configuration/lifecycle, review results with the same
   provider/model, and commit. If the pilot cannot match conditions, report the
   exact failed field and smallest repair; do not silently redefine the match.
3. Report calibration limits and a proposed next decision. Any release-gate
   revision requires the owner's separate approval and fresh representative
   framework measurements under the approved protocol. M3–M7 remain open.

Stop an unsuccessful calibration implementation after 90 minutes to reassess its
boundary. No new controls, renderer policies, public API or app features belong
in this work. Diagnostic sources live under `tools/macos-calibration`; they are
not included in the SDK or internal Placement application.
