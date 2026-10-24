# macOS developer-preview execution ledger

Publication update: [v0.1.0-alpha.3](https://github.com/dshills/vellum/releases/tag/v0.1.0-alpha.3)
was published as a prerelease on 23 October 2026 at
`320376ec8573a8603f75faaa3b3a7ea8a8ea0366`, after owner approval. The SDK,
Placement app ZIP and checksums are available there, with ad-hoc signing,
notarization and qualification limits disclosed. The candidate ledger below
preserves its prepublication checkpoint state, including P6's then-pending
publication decision. Approval covered that exact payload; it does not authorize
new artifacts or include the subsequent personal-use widget work.

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
| P2 | Matching alpha.3 Placement app passes the listed editing, Save/Open/history and shutdown workflows; native text/window checks pass within their recorded scopes. Earlier app24 evidence remains separately scoped. | Private candidate acceptance complete; retain final artifact and native automation limits. | R2/R3/R5 |
| P3 | Bounded idle/activity/lifecycle and two-hour soak checks now pass. Apple TUINS base class origin/stable count are known; native retaining reference and full native/GPU heap remain unqualified. | Bounded preview resource check complete; retain the disclosed native-heap limits and historical failures. | R3 |
| P4 | Owner approved the exact decision-packet payload and corrected 418-file notice bundle on 23 October; release/macos-preview-audit.json binds the decision. Future changed payloads require a new audit. | Guarded public candidates now produced and verified locally; retain historical declarations and separate publication approval. | R4 |
| P5 | Guarded public native route implemented at b0e5106; owner-approved exact payload may now pass the existing public packaging gates. | Implementation agent: retain release-audit and exact-archive notice gates; no further tooling repair required for this route. | R4 |
| P6 | Alpha.3 private artifacts, checksums and ad-hoc signing/notarization disclosure are complete; public-mode artifacts and checksums pass; final publication approval remains pending. | Implementation agent: prepare exact reviewed artifacts and release text, then owner publication decision. | R4/R5 |

This ledger tracks completed acceptance checks and remaining distribution
questions, not six newly discovered runtime defects. A new reproducible defect must name its command
or gesture, expected/actual behavior, smallest fix and focused retest here.

## Checkpoints

- R1: policy approved, branch/source selected, pinned dependencies prepared and
  bound; Prism review is clean and checkpoint commit is `578cf74`.
- R2: complete for private candidate source `d8de4fc`; [external SDK/app result](reports/macos-preview-r2/README.md).
- R3: the recorded functional and bounded resource checks pass within their
  stated scopes. Full repository check passed; [packaged app acceptance](reports/macos-preview-app-acceptance/README.md)
  covers editing/persistence/recovery, with [final delivered-byte renewal](reports/macos-preview-final-private/README.md).
  The [quiet lifecycle run](reports/macos-preview-lifecycle100/README.md) passes
  100 cycles and 99 same-name replacements; native starter and Unicode/editor/dialog
  fixtures pass. The [activity/document checkpoint](reports/macos-preview-resource-check/README.md)
  passes ten-minute activity and document safety. After fixture-only cover inset
  correction `627a3bc`, the [matched two-hour soak](reports/macos-preview-soak-two-hour/README.md)
  passes four phases, 24 eligible samples, worker retirement, tracked native cleanup
  and normal exit. Earlier 83-cycle, close-request and occlusion-setup failures
  remain unchanged in their historical reports. Apple TUINS base class origin and
  stable observed count are known; the retaining reference and native/GPU heap
  deallocation remain unqualified. No full P11 or blanket native-safety claim is made.
- R4: alpha.3 identity committed at `aaa0a3a`; guarded public native exporter repair
  at `b0e5106`; stale exported contract version corrected at `412e5b8`. All changes
  have focused tests and Prism reviews. The owner approved the exact payload and
  corrected notices on 23 October; [the audit record](../release/macos-preview-audit.json)
  binds that decision. Historical declaration-collection labels are unchanged.
- R5: [final private delivered-byte acceptance](reports/macos-preview-final-private/README.md)
  passes for the alpha.3 SDK and build-28 app. Fresh external SDK starter, strict
  extracted app signature, place/pan/zoom/history/Save/Open/normal quit pass.
  Exact-payload audit and source push/draft PR for CI are now owner-approved.
  [Guarded public candidates](reports/macos-preview-public-candidate/README.md) now
  pass extraction, external/native SDK and app checks. The owner additionally
  approved historical local paths/metadata; source publication succeeded and
  [draft PR #1](https://github.com/dshills/vellum/pull/1) is open. All four applicable
  push/PR CI runs pass at e9907d9; optional self-hosted qualification is skipped.
  The [final publication packet](reports/macos-preview-ready/README.md) records
  proposed assets, release text and identical non-documentation code tree.
  Final tag/assets publication remains owner-controlled. No public preview
  release or full Mac qualification is claimed.

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
