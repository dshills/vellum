# Joined raster/upload/scheduling: prepared diagnostic

The [protocol](../../../tools/macos-calibration/joined/PLAN.md) prepares one bounded
native capture with absolute raster boundaries and request wake/check stamps.
It reuses the original-policy CPU-upload observer; no pool, cache mode, rendering,
pacing, SDK or application change is introduced. No new timing result is claimed.

The [renderer patch](renderer.patch) brackets unchanged rendering with a bounded
scalar journal, normal-return marker and safe reset. The
[fixture patch](native-main.patch) shows added scheduling clocks and receipt clock
bracket metadata; view/routing/model,20ms lateness and120ms response limits remain.
Native host stamps provide contiguous per-input segments; Go timer offsets and
nested work durations remain separate. No guessed cross-clock target conversion,
GPU-phase attribution or sum of independent phase percentiles is used.

Five recorder race methods, the original routed pixel/state fixture with the
clock enabled, and five classifier methods with23refusals pass. Sparse tags are
joined by identity; missing/ambiguous/unknown raster records are refused. The
native executable builds. Full recipe reproduction, original private archive/
reviewed overlay identities, sole Core renderer bracket delta, all8,321frozenSDK
files and owned build-process retirement are independently verified. Exact
[preparation checks](preparation-checks.json) and indexed evidence retain the basis.

One observer-on process/248requests,40seconds native/60overall, is planned within
the approved two-minute quiet desktop window after review/commit. All startup and
failure data remain retained; no unchanged retry or ordinary-build equivalence
claim. Features and33.3ms receipt-to-actual-presentation p95 gate unchanged;
M3–M7 open. This preparation does not fix latency or qualify a release.

Prism’s preparation review requested explicit begin-publication state and accurate
observer-on usage text. Both are corrected before native execution. A fifth race
method fault-injects End while Begin’s clock is blocked, refusing completion/reset
until publication and preserving invalid history. The final v2 race/pixel/build
receipts supersede the preserved v1 preparation; neither version has run natively.
