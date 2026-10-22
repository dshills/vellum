# First macOS release: delivery proposal

Date: 22 October 2026. Status: **proposal awaiting owner approval**.

The recommendation is to deliver a **macOS developer preview**, proposed version
`v0.1.0-alpha.3`, in three working sessions plus one contingency session. This is
an execution budget, not a promise that unresolved defects or distribution rights
will disappear within it. The deliverable is a consumable Go SDK, a packaged
Placement application, a minimal external starter, and accurate release notes.
It is more than another internal app build. It does not claim the production
framework, full platform matrix, or existing M3–M7 requirements are complete.

The owner's request for a plan does **not** approve a changed release gate.
[The current policy](macos-responsiveness.md) remains in force until the owner
approves the specific preview policy below. No new renderer investigation is
needed to prepare that decision. No performance experiments ran while preparing
this proposal.

## Why delivery stalled

[The prototype plan](prototype-plan.md#measurement-contract) calls its numbers
provisional engineering acceptance targets. It specifies 4 ms UI CPU work,
99% of presentation intervals within 1.1 display periods, and 33.3 ms p95
input-receipt-to-presentation latency at 60 Hz. The 19 October owner decision
made measured 60 Hz responsiveness the release gate and kept 4 ms as an
optimization target. The implementation of that policy retained the prototype's
33.3 ms latency cutoff. These values were engineering choices; the plan does
not cite a user-acceptance study establishing 33.3 ms as necessary for this preview.

Frame cadence and interaction latency measure different things. One existing
paced grid run presented all 599 measured intervals within 18.333 ms while its
input-to-presentation p95 was 62.224 ms. Serial grid/tree p95 was about 48 ms.
That is evidence of a real latency gap against the specified target, and also
shows why a latency failure alone does not establish failure to sustain 60 Hz.
It does not qualify all workloads or repeated runs. See the
[default-path assessment](../docs/reports/macos-default-responsiveness/README.md).

The latest single render-priority candidate also failed, at **48.286 ms p95**,
with complete accounting and normal shutdown. It was rejected, not integrated;
[checkpoint 88fc75d](../docs/reports/macos-render-priority-result/README.md)
records the result. The small native comparison reached roughly 32 ms, so neither
impossibility nor an unavoidable OS latency floor is established. A substantially
larger presentation-ownership change is unproven and has no defensible release
estimate. It should not be the automatic next item in a first-preview delivery plan.

The process failure was allowing diagnostic work to stay on the critical path
without a product release decision. More instrumentation does not itself make
the SDK consumable, the application reliable, or the distribution ready.

## Proposed preview policy — explicit approval required

Approve these changes as one preview-only scope decision. Historical results and
the full qualification ledger stay intact; a failed result is never relabeled.

| Area | Required for this preview | Deferred claim/work |
| --- | --- | --- |
| Platform | Apple Silicon, default backend, exactly recorded macOS/display/toolchain configuration | Intel, other OS versions, multiple-display qualification, Windows/Linux, high-refresh guarantees |
| Performance | Disclose measured cadence and latency with workload/source scope; core interactions must work and complete without a stuck render/input queue | 33.3 ms p95 and repeated 99% cadence qualification become optimization/production-qualification work for this preview; do not replace them with a larger invented passing number |
| Correctness | No unresolved reproducible crash, hang, data corruption/loss, wrong Undo/Redo, stale callback mutation, or broken required workflow | Cosmetic defects and optional unqualified paths may be listed explicitly |
| Resource safety | Bounded ownership/queues/caches, passing affected race/lifecycle tests, clean shutdown and a bounded native resource check | Full 1,000-cycle and eight-hour endurance certification; unknown native ownership cannot be described as proven safe |
| Input/accessibility | Required pointer/keyboard/text workflows and existing native semantic contract tests pass | Unverified human VoiceOver/IME/cross-display cases remain clearly marked; no blanket accessibility claim |
| SDK/distribution | Clean external consumer, matching app/SDK versions, complete applicable notices and resolved rights for shipped contents | Public distribution cannot be enabled merely by changing an audit-status field |

Keyboard access, correct text persistence, data safety and object lifetime are
not optional because the release is a preview. If an advertised workflow is
broken, fix it or explicitly remove it from preview support before approval;
marking it experimental alone is not a passing test.

Keep all existing features frozen and experimental rendering/pacing disabled.
Keep Cogent/WebGPU/Metal. No Metal rewrite, new controls, new application features,
new performance harness or general architecture cleanup belongs in these sessions.

## Five checkpoints and stop conditions

The implementation agent owns these checkpoints. Prism with Anthropic Claude
Sonnet 5.5 reviews significant changed code/specification checkpoints, followed
by a separate commit. The owner decides release policy and publication; an
unresolved upstream rights question may require a maintainer or rights holder.

### R1 — Freeze the release contract and source (first half-session)

- Record approval or rejection of the preview policy before changing gate or
  release-status metadata. Adopt this document as the active preview checklist
  only if approved; retain the full Mac requirements as a separate future target.
- Select one clean release commit and create `codex/macos-preview` from it.
  Use the existing working product code. The latest Placement product source is
  `7047740`; current production-path differences from it are the stage0106
  opt-in diagnostic-storage fix and its preparation metadata/tests. Review that
  delta and bind its prepared dependency output. Do not import the rejected QoS
  candidate or private calibration copies.
- Do not silently combine the older `bd3cbc9` performance SDK with a newer app
  and label them the same release. Reuse its evidence only with stated scope.
- Create one short blocker list with reproduction, affected workflow, owner,
  required fix and retest. Start the shipped-component rights triage immediately
  so it cannot first appear during packaging.

Exit: approved contract, frozen source/dependency identity, fixed supported scope,
and a finite blocker list. If policy is not approved, report that decision point;
do not start another performance campaign or call the existing full gate passed.

### R2 — Prove someone else can build an app (second half-session)

- Reuse `drivers/cogent/prepare.sh`, `scripts/release-sdk.py`,
  `scripts/check-sdk.py`, `scripts/new-app.py` and existing build/package tooling.
  Produce a private candidate first; do not bypass public-distribution guards.
- Export and extract the SDK to a fresh location. Generate/build a new external
  Go module through public APIs, launch its native window and close it. Ensure
  its replacements resolve only to the extracted SDK, not the original checkout,
  private diagnostic trees, or research workspace.
- Build the Placement app from the same frozen source and dependency versions.
  Check that shipped instructions reproduce the build and installation steps.
  Use existing manifest/integrity tooling; no new evidence framework.

Exit: usable SDK and independent starter, plus one matching app candidate. A
missing native dependency or checkout-only path is a release blocker. Fix that
specific issue and repeat the affected check; do not rebuild unrelated fixtures.

### R3 — Accept the application and framework workflows (one session)

Use one written script against the candidate package, reusing existing automated
coverage for unchanged code. The drawing app is the principal acceptance surface;
the editor/workbench examples cover framework text, focus and multiwindow behavior.

Required application script:

1. Launch, drag palette items onto a transformed canvas, reject/cancel a drag,
   select/marquee, move/resize groups, pan/zoom and snap.
2. Copy/paste, reorder layers, group/ungroup, delete, then Undo/Redo. Verify saved
   object identities, geometry, order and group membership after Save/Open.
3. Exercise save failure/cancel, unsaved close and recovery without overwriting
   user drawings. Export one SVG and PNG; check content, not just file existence.
4. Exercise keyboard focus, shortcuts, ordinary Unicode text editing/clipboard,
   window resize, minimize/restore, background/foreground, second window and quit.
   Use the existing text fixtures for composition tests without changing the
   user's global input sources or accessibility preferences.

Record which steps use physical pointer input, routed automation or headless
checks. Retain current human VoiceOver/IME limitations unless separately verified.
Do not generalize an automated semantic check into human accessibility acceptance.

Resource check: reuse the existing harness for 100 lifecycle cycles, ten minutes
of active editing, eligible 60-second idle/minimized samples and a two-hour soak.
The soak can run unattended alongside documentation work. These are bounded
regression checks selected for this preview, not proof equivalent to the full
endurance contract. Compare settled resource samples with retained live objects
and configured bounds. Repeated unexplained growth, pending work after closure,
or an unretired application worker blocks shipment; a process exit alone does
not prove native object deallocation. Investigate the previously observed hidden
TUINS window only enough to establish owner/stable count versus accumulation.
If this cannot be resolved in-session, report the exact resource blocker.
A runtime fix after the soak begins invalidates that soak for the final candidate;
repeat the two-hour soak and affected lifecycle checks on the fixed build using
the contingency session. Documentation-only changes do not invalidate it. Do not
quietly carry an earlier binary's soak result onto changed runtime code.

Run one stable-candidate repository check using the existing `scripts/check.sh`
workflow, including native module/example checks, plus applicable CI. Run focused
tests for any fix. Do not repeat the entire suite after a documentation-only edit.
Existing performance results are disclosed with their original source scope. A
material runtime change requires an affected performance regression check; merely
missing the old 33.3 ms preview-deferred target must not restart optimization.

Exit: required workflows pass, resource observations are acceptable within the
preview's stated limits, and no unresolved release-blocking defect remains.

### R4 — Finish release contents and distribution (third session, with early triage)

- Complete the applicability review of the actual shipped Go/native/font/asset
  notices. Existing inventories are inputs, not a reason to regenerate them all.
  Resolve the recorded Apple-SDK-derived binding question and any applicable
  font/generated-asset uncertainty. Do not infer permission from an included
  license filename. If unresolved, identify the exact component and smallest
  removal/replacement/approval action; keep public publication blocked.
- Set matching SDK/API/release metadata for the approved preview version.
  Record platform support, known latency, same-window placement, drawing limits,
  API instability and unverified accessibility/input cases.
- Prepare the public GitHub prerelease text and artifacts locally. Choose and
  document the distribution route: an ordinary consumer-facing Mac download needs
  the signing/notarization path completed; a developer preview distributed without
  it must disclose that status and the expected installation procedure explicitly.
  Do not claim notarization or use a signing identity without authorization.

Exit: final contents, notices and distribution decision are reviewable. If an
external rights or signing dependency remains, the tested private candidate can
be delivered, but the public prerelease is still blocked and labeled accordingly.

### R5 — Package once, validate the delivered bytes, publish (final half-session)

- Package the final SDK and Placement app from the frozen release commit with
  checksums, build instructions, release notes and a concise known-issues list.
- Extract the final SDK and app outside the checkout; run the external starter
  and a short launch/place/save/reopen/quit test using the actual delivered bytes.
- Resolve Prism findings, confirm applicable CI on the candidate source (or an
  explicitly verified identical code tree), and commit the release checkpoint.
- Present the exact tag, release notes and assets for the owner's final public
  publication/signing authorization. After authorization, publish the GitHub
  prerelease, verify downloadable artifact hashes, and launch the delivered app.

Exit: the owner can install the app and build a separate Go application using
the matching SDK. Publication, if approved, points at those exact artifacts.
A private candidate is not reported as a completed public prerelease.

## Calendar, failure handling and definition of done

Budget: three working sessions of roughly six to eight hours, plus one explicit
contingency session for reproduced blockers. Rights/signing responses and access
to an independent test Mac are external dependencies, not included guarantees.
Use the already-tested local configuration if another Mac is unavailable and
state that support limitation; do not invent a second-machine pass.

The effort allocation is explicit; R4 starts early and uses unattended build/soak
windows where possible, but its effort is still counted:

| Checkpoint | Working-session allocation |
| --- | --- |
| R1 contract/source freeze | 0.5 |
| R2 SDK and external starter | 0.5 |
| R3 functional/resource acceptance, including the first soak | 1.0 |
| R4 contents/rights/distribution preparation | 0.5 |
| R5 final artifacts and delivery | 0.5 |
| Planned total | 3.0 |
| Contingency for reproduced fixes and affected reruns | 1.0 maximum |

Calendar order: session 1 is R1 then R2; session 2 is R3; session 3 is R4 then R5.
The early rights triage is part of R1. Any R4 work completed during unattended
build/soak time is a scheduling benefit, not a reduction required to meet the
three-session budget. Session 4 is contingency only.

Each restarted soak requires two elapsed hours plus the affected lifecycle check
and repair time from that contingency. If the remaining session cannot contain
the repair and full required rerun, the result is no-go with the exact blocker;
never shorten the soak or carry forward invalid evidence to meet the deadline.

External approvals do not consume engineering work by themselves and do not
silently extend this allocation. If their absence prevents publication, report
that precise blocker with the ready private artifacts.

After every session report: deliverable completed, exact remaining blockers,
next smallest action, and remaining session budget. Spend at most half a session
on one newly found blocker before reporting its reproduction and options. Do
not silently extend the deadline, add another diagnostic stage, or declare an
unsafe candidate ready to satisfy the schedule. At the end of the contingency
session either deliver the preview or report a finite no-go list and the exact
candidate available for testing. No automatic continuation of profiling follows.

Done means R1–R5 have their required evidence, the supported workflows work in
the packaged app, the SDK builds an independent application, shipped rights and
notices are resolved for the selected distribution route, and the release is
honestly labeled. It does not mean every old research target has turned green.

If the owner instead requires full M3–M7 qualification for this first release,
retain the existing policy. There is currently no evidence-based completion
date for its latency blocker; plan the larger backend work separately rather
than offering another short release promise. The recommended next action is
approval of this preview policy, followed by R1/R2, not another profiler run.
