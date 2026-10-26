# Offline v1 maintenance

The owner asked for autonomous highest-value work while offline on25 October2026.
Keep the existing feature/backend/performance freeze and private-release limits.
Other projects and their desktop activity are outside this maintenance scope.

## Completed integration

The reviewed v1 closeout was fast-forwarded to main and pushed at9f05a37. Runtime
and SDK bytes remain the qualified382e0e4/build9 delivery; report-only integration
does not relabel them. No tags or public app assets were published. The existing
main workflows also passed after integration: [Framework40254929774](https://github.com/dshills/vellum/actions/runs/40254929774) and [Platform adapters40254929791](https://github.com/dshills/vellum/actions/runs/40254929791).
The release report retains the original exact-payload green runs.

## Root API guard checkpoint

The separately versioned tools/apicompat module adds no dependency to the root or
native driver module. It checks the13 stable root packages with Go's pinned
semantic diff and an additional conservative generic-constraint check. The pilot
caught a constraint-narrowing omission in the pinned upstream tool; the regression
remains asserted. Constant/type/function/field/interface/method/package cases,
read-only baseline checks, hash/membership corruption, policy shrink, module
origin, child environment, build-excluded cgo and dirty/hidden source refusal are covered.

The reviewed export baseline was captured from clean stable source at9f05a37
(same public API as the382e0e4 release), hashes to
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

## Next bounded work

1. The guard passed [its first three-host CI matrix](https://github.com/dshills/vellum/actions/runs/40258277100)
   at e2530b1. The subsequent test/vet step split is separately reviewed to prevent
   Windows shell error masking; [its follow-up matrix also passed](https://github.com/dshills/vellum/actions/runs/40258446890)
   at d7e0652. Remaining existing Framework/Platform workflows must be observed
   before integration.
   Fix only reproduced guard defects, preserving its assertions and scope.
2. Improve repeatable private packaging/developer handoff where a concrete gap
   exists; keep the accepted build9 archives unchanged.
3. An external native-facade consumer now checks the existing whitelist's field
   reads/writes, function/method signatures, app.Driver implementation and retained
   toolbar alias. Stale package/toolbar comments are corrected. This is compile
   evidence alongside focused offscreen regressions, not a full native semantic
   baseline or new physical/human qualification. No runtime implementation changed.
4. Reuse unchanged verification, do not start new performance/qualification
   campaigns, and preserve notebook data and global system preferences.

A30-minute chat continuation checks the latest owner messages before more work.
When the owner returns, disable it, finish a safe current checkpoint and report
commits, verified behavior, limitations and the next concrete action.
