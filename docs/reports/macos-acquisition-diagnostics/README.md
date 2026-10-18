# Measured GPU acquisition diagnostics

Stage85 records a detached acquisition outcome on the measured composition path:
source, stage, actual attempt count and first/last backend statuses. An unavailable
render pass now produces a specific failure instead of the generic missing-frame
message. Existing acquisition/release/configuration decisions, the single allowed
lost/outdated retry, ownership, timing boundaries and strict completed-frame
criteria are unchanged. Ordinary drawing adds no observation. Existing measured
methods retain their return conventions; optional detail methods support legacy
renderers with unrecorded status and offscreen textures with no invented surface
attempts. Metadata has bounded scalar descriptions and retains no handles/history.
Existing arbitrary backend error strings are not newly bounded.

The fresh, separately admitted native structural case **FAILED** after16 acknowledged
Go FIFO inputs and15 source-asserted observations, last `section-reenabled`.
The refused `page-disabled-refusal` input is acknowledged at Frame76; the following
Frame fails before a Snapshot/model observation. The actual acquisition reports
`source=surface stage=unavailable attempts=1 first=occluded(196609)
last=occluded(196609)`. No failed snapshot, disabled-page observation, later split/
document result or full50-input/42-state pass is inferred. The raw producer exits1
in3.581s. Session cleanup reports Closed=true, Pending=false, LiveWindows=0 after79
issued requests; original-window close was not reached. Owned process group72988
is absent. Raw logs are complete and pinned inputs unchanged. A separate complete
post-failure source/prepared/SDK/binary boundary check passes; the controller's
post-validation boundary did not execute. This scope is permanently sealed.

This identifies the unavailable surface in this run only. It does not establish
why the backend reported it occluded, nor diagnose the older41/35 failure, which has no
acquisition detail. The next product action is to define and implement bounded
visibility-aware scheduling of explicit Frame requests: preserve strict completion,
never spin/present a fabricated frame, keep cancellation/closure responsive, and
design a supported signal to resume eligible pending work when drawable availability
returns. Current shown/nonminimized state alone supplies no drawable-availability
notification. Exercise hidden/
occluded-to-visible and cancellation paths with deterministic injected outcomes
before admitting a distinct native scope. Do not repeatedly retry sealed cases.

The original geometry oracle and failure stay immutable. A separate strengthened
oracle covers only160/162DPI: disabled heading/draft/action heights shrink6/2/6px,
and their restored enabled metrics equal the original enabled metrics. Actual
public Manual consumer tests pass at both DPIs, including retained children,
keys, pointers, split/scroll/document behavior. One positive and40 negative
synthetic oracle controls pass. These software fixtures are not native evidence;
the new native case has no complete independent oracle pass.

The complete driver-root package (`.`) and gallery race suites, focused acquisition/composition/
FrameTrace regressions, native gallery build,11 prepare tests and two identical
offline preparations pass. Prepared output has6274 files. Injected acquisition
cases assert exact ordinary versus measured calls and independently expected
release/configure order for occluded/timeout/error/lost/outdated/recovery/null/
suboptimal/unknown outcomes. GraphicsSystem, gpudraw and Core tests preserve error
identity and reject failed frames even with Submitted=true.

Default driver/gallery/GPU vet passes. Default broader Core vet **fails** with12
diagnostics reproduced exactly on pre-stage85 source:10 unkeyed composites and2
unreachable-code diagnostics. Both failures are retained; remaining Core analyzers
pass with just those two inherited analyzer classes excluded. The initial prepare
test also failed because its patch-number endpoint excluded85; the corrected
endpoint and all11 tests pass. No unrelated Core source is changed to hide these
failures, and no unfiltered Core test suite is run.

Internal source/controller and exact-plan reviews precede native execution.
Prism provider review, independent dispositions and final archive/commit review
are required before commit. The bounded [archive](records/checkpoint.tar.gz) and
[manifest](records/manifest.json) will retain raw checks, failures, source,
controllers, plan/admission and review payloads without executables or SDK copies.
See [results](records/results.json).

Evidence uses Go1.25.6 ARM64/GOMAXPROCS2 CPU-scene, retained source-built native
archive786a and the immutable7880-file SDK. All current non-prototype Go, native
support and prepared bytes match the effective SDK overlay. The SDK native-library
override is scoped; the checkout default cached library is not qualified by this
run. No SDK refresh or public release occurs. Go FIFO establishes no physical/OS
input, native pixels, presentation, IME/VoiceOver, split dragging, arbitrary display
transitions, performance or native/GPU memory/endurance. Full structural M2,
physical M5 and M3–M7 stay open. Follow the [Mac delivery plan](../../../specs/macos-delivery.md).
