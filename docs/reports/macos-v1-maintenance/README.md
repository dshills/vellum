# Offline v1 maintenance

The owner asked for autonomous highest-value work while offline on 25 October 2026.
Keep the existing feature/backend/performance freeze and private-release limits.
Other projects and their desktop activity are outside this maintenance scope.

## Completed integration

The reviewed v1 closeout was fast-forwarded to main and pushed at 9f05a37. Runtime
and SDK bytes remain the qualified 382e0e4/build9 delivery; report-only integration
does not relabel them. No tags or public app assets were published. The existing
main workflows also passed after integration: [Framework40254929774](https://github.com/dshills/vellum/actions/runs/40254929774) and [Platform adapters40254929791](https://github.com/dshills/vellum/actions/runs/40254929791).
The release report retains the original exact-payload green runs.

## Root API guard checkpoint

The separately versioned tools/apicompat module adds no dependency to the root or
native driver module. It checks the 13 stable root packages with Go's pinned
semantic diff and an additional conservative generic-constraint check. The pilot
caught a constraint-narrowing omission in the pinned upstream tool; the regression
remains asserted. Constant/type/function/field/interface/method/package cases,
read-only baseline checks, hash/membership corruption, policy shrink, module
origin, child environment, build-excluded cgo and dirty/hidden source refusal are covered.

The reviewed export baseline was captured from clean stable source at 9f05a37
(same public API as the 382e0e4 release), hashes to
`a75e6d0cc398e905648495e3e97a0e3f6177e06c8216f4edb4833d39f9c65697`,
and has no local source paths. Semantic roundtrip/current comparison passes.
Prism reviews text/code and the baseline manifest; the generated binary snapshot
is checked by its digest, source identity, membership and semantic comparison.
It is not an LLM-reviewed binary or a cryptographic source attestation.

The guard covers root source compatibility only. Native facade, behavior,
errors.Is, file formats, UI and human/hardware qualification retain their existing
checks. Constraint edits can require review even for compatible widenings or
parameter renames; do not regenerate a baseline to silence a finding.
The new bounded workflow checks the guard on Mac/Linux/Windows; that qualifies
the developer tool, not those platforms' native framework runtime.

## Native facade checkpoint

An external native-facade consumer checks the existing whitelist's public field
reads/writes, function/method signatures, app.Driver implementation and retained
toolbar alias. Stale package/toolbar comments are corrected. Focused offscreen
toolbar, appearance and settlement race tests plus native vet pass. This is
compile evidence alongside existing regressions, not a full native semantic
baseline or new physical/human qualification. No runtime implementation changed.

## CI and integration handoff

All three workflows passed for the final maintenance source 20659c9:
[Framework](https://github.com/dshills/vellum/actions/runs/40258851542),
[Stable root API](https://github.com/dshills/vellum/actions/runs/40258851606), and
[Platform adapters](https://github.com/dshills/vellum/actions/runs/40258851622).
The exact source and observed workflow timestamps/outcomes are in [ci.json](ci.json).
The earlier e2530b1 and d7e0652 API matrices also passed; the latter separates
test/vet steps to prevent Windows shell error masking. Significant checkpoints
have [Prism reviews and focused check receipts](../../reviews/macos-v1-maintenance/README.md).

The branch is ready for fast-forward integration. This handoff commit changes
only reports; the recorded green run identities remain 20659c9. Observe any main
integration CI without relabeling old results. Fix concrete failures if they
occur; do not rerun accepted native/desktop/performance campaigns or rebuild the
unchanged build9 packages.

Once integration is healthy, preserve the private v1 delivery for the owner's
use. There is no newly identified runtime or packaging defect in this checkpoint.
Human IME/VoiceOver, other hosts/displays and public distribution remain outside
the accepted personal-use scope. Do not invent feature or qualification work to
fill the offline period, and do not change notebook data or global preferences.

A 30-minute chat continuation checks the latest owner messages before more work.
When the owner returns, disable it, finish a safe current checkpoint and report
commits, verified behavior, limitations and the next concrete action.
