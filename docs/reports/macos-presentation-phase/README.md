# Phase-matched presentation control — first attempt rejected

The native reference completed and passed the strict phase/pending-state
classifier. The framework diagnostic then exited with **initial natural
presentation deadline**, before any routed input. **No matched pair exists.**
The complete attempt is retained and sealed; no unchanged retry occurred.

Native retains128warmup and120measured joins, with30measured inputs in each
predeclared phase. Its receipt-to-shown p95 is **34.960ms**;
30/120responses exceed33.3ms. Per-phase p95 values for1/8,3/8,5/8,7/8 are
30.933,26.734,22.571,35.105ms. Receipt-error maxima remain below0.862ms.
These are instrumented sparse-fixture observations, not release qualification,
a universal native floor or permission to change the gate.

All251native acquisition attempts (including three startup attempts),
1001scalar events and1988display callbacks remain accounted for.
Every measured receipt hook and reconstructed receipt/present-request/final-GPU
point has zero **other** recorded pending drawables. All callbacks drain to
pending0/acquiring0. This describes acquisition-awaiting-callback state, not
Metal pool availability, reference lifetime, drawable credits or compositor queue
depth. It cannot establish the cause of presentation latency.

The framework emits four startup backend present records with zero shown feedback,
no presentation-metrics samples or input receipts, and an empty pending trace.
Its observer never starts because startup readiness fails first. This is a
failed diagnostic startup, not a framework latency observation. There are no
framework phase buckets and no subtraction from the valid native run.

An [offscreen linked-code audit](linked-source-audit.json) finds that the actual
framework `VellumPresentationLayer nextDrawable` method lacks the pending hooks
present in the supplied Objective-C overlay. The original presentation-acquire
call is linked, so the missing pending hooks do not independently explain the
whole startup failure. Source/build recipes and offscreen pixels alone were
insufficient to prove native-overlay integration. The next diagnostic repair is
to materialize native Core changes in a fresh private copy and require actual
linked-hook evidence before any new capture. The SDK/original Core remain sealed.

The attempt takes52.939seconds within100seconds; native takes
34.075seconds and framework6.001seconds. All18owned native/probe
groups retire and are observed absent. Configuration/geometry, raw log hashes,
compiled/runtime pins and all8,321SDK files verify. Framework stderr/error are
retained, without dropping the failure. The approved quiet scope is consumed;
desktop free, no native process remains. A corrected attempt needs its own
preparation, review, commit and fresh bounded desktop authority.

Preparation checkpoint`1f930d4` includes eight focused methods,
headless concurrency/overflow/JSON/drain tests, both source builds, routed full
pixels, recipe and SDK checks. Prism using Anthropic Claude Sonnet5.5 covered
21initial and7follow-up files; its overflow concern was fixed and rejected-prefix
state continuation was proven to remain visibly refused. The separate six-file actual-result and linked-code audit review completed
without findings, exclusions or truncation; see the
[result review receipt](result-review-completion.json).

Original Cogent/WebGPU/Metal design, features, SDK/app, allocation/cache/presentation
and pacing defaults, and33.3ms release gate are unchanged. M3–M7 remain open;
no permanent display link or production repair is adopted. See the
[protocol](../../../tools/macos-calibration/presentationphase/PLAN.md),
[preparation checks](preparation-checks.json), [result summary](result-summary.json),
review receipts and indexed evidence.
