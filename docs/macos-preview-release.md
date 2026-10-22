# macOS developer-preview execution ledger

The owner approved [the delivery plan](../specs/macos-first-release-plan.md) on
22 October 2026. Features are frozen. Default Cogent/WebGPU/Metal only; experimental
rendering, pacing and rejected private candidate patches remain disabled.

## Candidate and budget

- Branch: `codex/macos-preview`; source baseline `ee87ac3`.
- Placement product source: `7047740`; subsequent product-path changes are the
  stage0106 diagnostic storage correction and preparation metadata/tests, with no
  default rendering/event/API changes. New dependency preparation passes, including
  stage0106 output hashes. No private QoS candidate is present.
- Target: Apple Silicon; measured local macOS/display configuration only. Full
  platform, physical-input and human accessibility qualification is not claimed.
- Proposed final release identity: `v0.1.0-alpha.3`. Intermediate private SDKs may
  retain alpha.2 metadata until the coordinated R4 version checkpoint; these are
  candidates, not the final alpha.3 release.
- Session 1 began 2026-10-22 18:17:48 UTC; hard end 2026-10-23 02:17:48 UTC.
  It covers R1 and R2. Session 2 covers R3; session 3 covers R4/R5. One additional
  contingency session is the maximum. Unattended waits do not create new sessions.
- Session allocation and stop/no-go rules remain those of the approved plan.
  No automatic performance investigation follows a preview-deferred numeric miss.

The candidate binding is recorded in [preview-candidate.json](preview-candidate.json).
Full prepared-file hashes and preparation logs are local execution records;
reproduction uses the tracked preparation contract, not those machine paths.

## Finite blocker list

| ID | Required outcome / reproduction | Owner and smallest next action | Checkpoint |
| --- | --- | --- | --- |
| P1 | A new external Go module must build/run using the extracted candidate SDK without checkout/private-tree dependencies; R2 passes for SDK source d8de4fc; final alpha.3 delivered-byte renewal also passes at R5. | Implementation agent: export private SDK, verify relocation, generate/build/run starter. | R2 |
| P2 | Matching packaged Placement app must pass listed editing, Save/Open/history, text/window and shutdown workflows; earlier app24 evidence is scoped, not final artifact acceptance. | Implementation agent: build from frozen candidate, reuse applicable tests, run the written acceptance script. | R2/R3/R5 |
| P3 | Native idle/activity/lifecycle/soak must establish bounded retained work/resources; historical hidden TUINS window ownership and accumulation remain unclassified. | Implementation agent: use existing lifecycle harness and prescribed samples; classify stable owner/count or report exact growth/cleanup defect. | R3 |
| P4 | Public SDK/app rights and notices remain pending; presence inventories alone do not approve redistribution. Upstream objc2 notice explicitly raises Apple-SDK-derived binding terms. | Implementation agent: review actual shipped component applicability; seek a concrete rights-holder/owner decision only for the unresolved component. | R4 |
| P5 | Guarded public native route implemented at b0e5106; actual public packaging remains blocked by pending rights approval. | Implementation agent: retain release-audit and exact-archive notice gates; no further tooling repair required for this route. | R4 |
| P6 | Alpha.3 private artifacts, checksums and ad-hoc signing/notarization disclosure are complete; public artifact/publication approval remains pending. | Implementation agent: prepare exact reviewed artifacts and release text, then owner publication decision. | R4/R5 |

These are uncompleted acceptance tasks and known distribution questions, not six
newly discovered runtime defects. A new reproducible defect must name its command
or gesture, expected/actual behavior, smallest fix and focused retest here.

## Checkpoints

- R1: policy approved, branch/source selected, pinned dependencies prepared and
  bound; Prism review is clean and checkpoint commit is `578cf74`.
- R2: complete for private candidate source `d8de4fc`; [external SDK/app result](reports/macos-preview-r2/README.md).
- R3: acceptance in progress; full repository check passed and initial packaged
  drawing workflows passed. The first 100-cycle run stopped at 83 cycles; the owner
  subsequently confirmed other desktop activity. Its raw failure remains unchanged
  and is inconclusive as a controlled run. The [fresh quiet run](reports/macos-preview-lifecycle100/README.md) passes
  100 cycles and 99 same-name replacement checks. The alpha.3 independent native
  starter and Unicode text/editor/dialog fixture also pass. Hidden-window ownership,
  packaged-app acceptance, active editing and the two-hour soak remain open. The first
  soak stopped after native close requests at about 152 seconds; no two-hour soak
  pass is claimed. See the [first checkpoint](reports/macos-preview-r3-first/README.md).
  [Packaged alpha.3 editing/persistence/recovery acceptance](reports/macos-preview-app-acceptance/README.md)
  now passes for its listed interactions, including an exact-content controlled
  recovery round trip and normal process retirement. The [resource checkpoint](reports/macos-preview-resource-check/README.md) passes ten-minute
  active editing and document safety. A new two-hour attempt stopped at 30 minutes
  because native occlusion never became false under a containing cover; a short
  pilot reproduces the setup refusal. The soak remains blocked. Apple TUINS base
  class origin and stable count are known; its retaining reference remains unknown.
  Final corrected-notice packaging is prepared; delivered-byte renewal is separate.
- R4: alpha.3 identity committed at `aaa0a3a`; guarded public native exporter repair
  at `b0e5106`; stale exported contract version corrected at `412e5b8`. All changes
  have focused tests and Prism reviews. Applicable notice collection and real
  distribution-rights approval remain separate and pending.
- R5: [final private delivered-byte acceptance](reports/macos-preview-final-private/README.md)
  passes for the alpha.3 SDK and build-28 app. Fresh external SDK starter, strict
  extracted app signature, place/pan/zoom/history/Save/Open/normal quit pass.
  Public distribution rights, candidate CI and final publication decision remain
  pending; the two-hour soak remains blocked. No public preview release or full
  Mac qualification is claimed.

## Disclosed performance

The retained older bd3cbc9 SDK's serial grid/tree p95 is about 48 ms; its paced
grid p95 is 62.224 ms while 599/599 observed intervals meet 18.333 ms. One grid
run does not qualify sustained cadence generally. The rejected private priority
candidate measured 48.286 ms p95. Preserve these scopes and failures; none is a
measurement of a newly packaged alpha.3 app or physical input-to-photon latency.
See the [assessment](reports/macos-default-responsiveness/README.md) and
[rejected candidate](reports/macos-render-priority-result/README.md).

## R2 packaging finding for R5

Documents-location packaging once failed strict signing verification due to
Finder/resource-fork metadata; the tool cleaned the output. Temporary staging
passes. R5 must verify signature after extraction of the actual final archive;
do not call this a fixed general packaging defect without that check.
