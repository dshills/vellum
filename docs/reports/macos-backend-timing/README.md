# Actual backend timing diagnostic

The first observer-off/on scope is **sealed and rejected**, with no valid pair.
The process, callback drain and foreground/configuration checks pass, but the
original whole-population equality check refuses the first observer-on result:
its 1,548 command records contain 258 present buffers, while the public recorder
contains 252 presentations. All 252 public records have trace timestamp matches.
Six extra leading startup presents (two zero, four positive feedback) have request
times before the first public acquisition. Their actual drawable IDs were not
recorded; no IDs are retrofitted and this result is not relabeled valid. All two
native and sixteen probe process groups retire. See [refusal analysis](refusal-analysis.json).

The backend observer starts before the runtime, while the native public recorder
attaches to its layer scope later. The retained SDK passes actual Metal drawableID
in its presented callback (`system/driver/desktop/presentation_darwin.m:97`).
The corrected private observer records that same native ID, requiring complete
one-to-one admitted ID and timestamp feedback. Only an explicitly retained leading
startup prefix of at most64 lower IDs may precede admission; every request must
precede the first public acquisition, and its positive feedback must precede the
first cohort receipt. All prefix rows and zero/positive counts remain visible.
Any unmatched admitted/cohort record still refuses. The old scope stays rejected.

A second checker defect was found before another native run: it selected measured
frames by numeric tag threshold. Public tags are opaque and can be sparse. The
checker now uses explicit scope/tag/generation membership of the last120 measured
requests. A sparse-tag regression discriminates the original failure. Four new
methods cover valid retained prefix, unmatched/late prefix refusal and sparse tags;
14 backend plus18 unchanged parent methods pass. The fixture view/routing/pixel
regressions are reused by exact source identity and unchanged SDK. A fifth new
control refuses an unset recording flag while preserving valid native drawable
ID zero. The flag publishes the ID with Release/Acquire ordering.

A fresh private source/archive/binary under `vellum-backend-timing-v2` preserves
the rejected attempt and its source snapshots. Its [protocol](../../../tools/macos-calibration/backend/PLAN.md)
retains three off/on pairs, 240-second global/40-second per-process bounds and all
original stop rules. [Original build identity](build-identity.json) and
[revised build identity](build-identity-v2.json) disclose source, library, binary,
package versions, inherited warnings and build failures. Explicit archive-derived
Go link identity prevents binary cache reuse after native archive changes.

This is a private diagnostic derivative, never installed in the retained SDK,
shipping framework or packaged app. Observer-off/on runs the same diagnostic
binary/archive and does not prove zero observer impact or binary equivalence to
the retained original. No live acceptance follows from preparation. The 33.3 ms
release gate, feature freeze and failed earlier scopes remain unchanged. Actual
submit/GPU/presentation timing conclusions require a valid corrected campaign.

## Second checker refusal and source-only correction

The c18f74f attempt is also sealed/rejected after valid off and rejected on, with
no accepted pair. Exact drawable identities/prefix/feedback checks passed, then
timing arithmetic raised KeyError `(1,1)`: it treated an armed setup frame as a
measured request. Startup `Frame` calls can be armed without a measured receipt.
All two native and sixteen probe groups retire, callbacks drain and no overflow
occurs. See [second refusal analysis](refusal-analysis-2.json).

Both warmup and measured timing arithmetic now require explicit measurement
scope/tag/generation membership. All startup feedback remains joined/retained.
An armed-startup regression and the exact retained raw journal exercise this
correction before more desktop use. Posthoc checker success is only a regression
control; it does not accept or relabel the earlier scope. Fifteen backend plus18
parent methods pass. A fresh reviewed campaign reuses the byte-identical c18f74f
binary and private archive; no native/Go rebuild or SDK change is justified.
