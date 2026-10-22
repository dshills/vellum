# Phase-matched presentation control — preparation

This prepares one native-reference then framework pair with pending drawable
feedback state. **No native capture has started.** The existing design, features,
SDK/app, renderer defaults and33.3ms release gate remain unchanged; M3–M7 open.

The [protocol](../../../tools/macos-calibration/presentationphase/PLAN.md) fixes
four reported-period phases (1/8,3/8,5/8,7/8), an eight-callback anchor stride,
128warmups and120measured requests. Each phase has30measured inputs. Actual
native receipt error must be0..2ms; each phase's p50/p95 error must match within
1ms between implementations. A mismatch refuses the comparison without dropping
or replacing observations. Request spacing is diagnostic input scheduling;
renderer pacing remains unchanged. The approximate Go/native target bridge is
recorded separately and cannot substitute for actual native receipt matching.

Shared scalar hooks record every acquisition start, success/unavailability,
receipt and presented callback, with pending IDs and acquiring counts before
and after. Every transition is replayed. Exact receipt/acquisition/drawable/shown
joins preserve startup and all248cohort states. Native GPU/submit observations
are added without changing its command order; the framework keeps its full
source/raster/upload/backend observer chain. A pending record means acquired and
awaiting callback, not a claim about Metal pool availability, reference lifetime,
compositor queue depth or physical scanout. The receipt-hook census has its own
observation bracket; earlier event-time state is reconstructed separately.

The native reference still uses an internal AppKit event and cached-texture blit;
framework still routes the public command and normal view/paint/upload. That
algorithm/API difference and extra diagnostic overhead prevent interpreting
subtraction as pure framework overhead. Both view observers use the same source;
reported timestamp and actual delivered callback clocks remain separate.

[Focused checks](preparation-checks.json) pass: six classifier methods plus two
controller methods, including phase/lifecycle/parent-refusal and fail-first controls;
headless native bounds/lifecycle/JSON tests, nine-pending overflow refusal and
four concurrent producers;
warning-free native reference/observer builds and private Go consumer build;
full routed pixels; exact recipe reproduction; source/log/process ownership
pins and all8,321SDK files. The Go build retains existing AppKit deprecation and
duplicate-objc link warnings. No Rust/native archive or SDK rebuild occurs.

Preparation inspection corrected the receipt hook to the actual measured-input
recorder path. An initial offscreen link failure is retained; native-only hook
overlays are excluded from the unchanged pixel fixture. Unused intermediate
preparations remain local and did not run a native application. The actual
consumer/core/measurement/observer snapshots are included for source review.

One approved quiet scope may run after tests, internal review, Prism using
Anthropic Claude Sonnet5.5 and commit: at most100seconds overall and40seconds
per process, native then framework, fixed80ms held tail, existing exclusive lease,
configuration/geometry and source/SDK guards. Stop first invalid, seal and retain;
no unchanged retry. Performance above33.3ms remains an observation. Results,
including rejection, require separate review and commit. No permanent link,
production scheduling change or release qualification follows.

The initial complete21file Prism review raised one census-overflow concern.
Recording now stops after overflow and marks the trace invalid, with a headless
nine-pending negative proving bounded serialization and retained failure prefix.
An internal correction keeps the receipt probe outside the original clock-read
bracket. Both consumers and focused checks pass after these changes; a scoped
follow-up review covers only the changed files. No native capture has started.

Prism follow-up is complete for seven changed files. Its failed-prefix concern
is dispositioned with an [actual headless audit](failed-prefix-audit.json): scalar
callbacks drain after recording fails, while invalid/overflow flags remain set
and the classifier refuses the trace before replay. Failed prefixes are retained
as rejection evidence, not complete accepted journals. No actionable findings
remain; see the [review receipt](review-completion.json).
