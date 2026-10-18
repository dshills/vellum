# Typed unavailable Frame results

`DesktopSession.Frame` now returns a detached `DesktopCompositionError` for failed
composition. `errors.Is(err, cogent.ErrDesktopFrameUnavailable)` identifies a
checked known surface yield; `errors.As` exposes its original `Failure` text and
`Unavailable` flag. Existing error text is preserved. Context cancellation and
window/session closure keep their own identities. Native errors already reduced
to trace strings are not reconstructed.

Frame keeps its existing first-composition contract. An unavailable or failed
composition terminal publishes no frame capability. Each request consumes one
actual terminal and releases capacity exactly once. No automatic retry or requeue
is added; existing bounded acquisition and trace
polling remain unchanged.
An application can explicitly request a later frame after a meaningful lifecycle
or application event; neither shown/focused state nor this sentinel proves a
drawable is ready. Architecture review found no authoritative availability event
supporting safe automatic pending/resumption. That feature remains deferred.

Stage86 classifies only coherent existing acquisition sequences: one identical
first/last Occluded or Timeout status, or two attempts starting Lost/Outdated and
ending Occluded/Timeout/Lost/Outdated. Unknown, backend/view/encoder/pass errors,
success-null, offscreen and legacy outcomes remain generic failures. The scalar
propagates through measured composer, Core and `FrameSample`; false values are
omitted from JSON. Ordinary acquisition/release/ownership and measured timing
regions are unchanged. Metadata work remains charged as UI work. Contradictory
submitted/unavailable metadata fails closed.

An actual pre-change Manual FIFO regression reproduces lost `context.Canceled`
identity. The fix checks cancellation after taking the actual terminal trace,
before returning composition status or publishing a token. Queued cancellation
keeps capacity busy through input/owner terminals. The same regression then passes.
Cancellation does not retract dispatched or committed work. A late cancellation
can race an already-committed reply; publication is prevented when cancellation
is observed before terminal collection, rather than promised for every canceled wait.
Labelled synthetic measured-drawer Manual fixtures cover unavailable/no token,
no automatic requeue, stale tokens, a separately requested successful frame,
fatal errors, cancellation before/after composition, retirement, and detached
metrics copying. These exercise real software FIFO/composition paths with
synthetic measured results; they do not establish native GPU acceptance.

Initial internal review found source-detail failures could manufacture or retain
the availability flag. Exact counterfactual tests reproduce seven failing
negative subcases with three passing controls. The fix captures eligibility from
original terminal metadata and clears it on Core/composer diagnostic failure,
preserving existing first-error text and draw/end order. Those controls and final
formatted focused race checks pass. The unapproved first review, first candidate,
negative/fixed receipts and controller syntax failure before child creation are
retained; no failure is relabelled as a pass. Direct Surface.TryCreateView factory
failure injection remains a coverage gap; renderer-reported view errors are tested.

Complete driver-root package (`.`) and gallery race suites, focused GPU/composer/
Core/protocol regressions, affected-package default vet, native gallery build,
prepare unit tests and two deterministic offline stage86 preparations pass.
Prepared output has6276 files. Default broader Core vet still fails with the same
12 retained baseline diagnostics (10 unkeyed composites,2 unreachable code); no
new diagnostic appears. No unfiltered Core test suite is run. See [results](records/results.json).

No new native case runs at this checkpoint. The [older diagnostic run](../macos-acquisition-diagnostics/README.md)
failed16/15 with actual Occluded196609 before these typed changes. It does not
qualify the new public error mapping, full50/42 structural flow, physical input,
native pixels, presentation, IME/VoiceOver, display, performance, memory or soak.
Old failed native scopes stay sealed. Full M2/M5 and M3–M7 remain open.

Evidence uses Go1.25.6 ARM64/GOMAXPROCS2 offscreen checks and the controlled
source-built native archive786a for the build only. Current non-prototype Go,
native support and prepared bytes match the checked immutable-SDK overlay; the
SDK's native-library override does not qualify the checkout default cached library.
The7880-file retained SDK stays unchanged. No SDK refresh or public release occurs.
Prism Anthropic Claude Sonnet5.5 review, independent dispositions and final bounded
archive/commit review are required before commit. The [archive](records/checkpoint.tar.gz)
and [manifest](records/manifest.json) will retain source, checks and reviews without
executables or SDK copies. Follow the [Mac delivery plan](../../../specs/macos-delivery.md).
