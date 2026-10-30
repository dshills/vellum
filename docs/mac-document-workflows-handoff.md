# Mac document workflow maintenance handoff

As of 30 October 2026, features and the existing backend, rendering/pacing and
release/performance policies remain frozen. Desktop checks are deferred by the
owner. Current work uses internal review; it authorizes no external review,
public release/tag/assets, new audit decision, signing identity or notarization.

## Source and package identity

- Main at `ac61d782ff226dfbff2adb0c750e68693eab88bd` is pushed and green in
  Framework, Stable root API and Platform adapters. Framework succeeded on its
  second attempt after its first attempt passed verification but timed out
  saving cache. This is prior-source CI evidence, not a result for later commits.
- The local `codex/mac-offscreen-closeout` branch adds five ordered maintenance
  checkpoints. Use `git rev-parse HEAD` and `git status --short --branch` for the
  exact candidate; this handoff does not claim the branch was pushed or merged.
- The latest runtime correction is `f5d6feb` (stale Quit/deferred Open cleanup).
  Packaging tooling is `02ee981`; documentation-only changes may follow it.
- The last durable private bundle remains build22 at source
  `4f447e85aa97dbcfa3b36d2707a679c173791777`, SDK SHA256
  `28957d16841d4b8bbecbf53f2724b68cf812cd5cad048cc093e0d83d1cc7f4fa`.
  It predates f5d6feb and does not contain those two lifecycle corrections.
  No rebuild of that bundle is claimed here.
- The installed build22 path is
  `~/Applications/Vellum/build22-4f447e85aa97/apps`. Do not infer current app
  liveness or replace an open version from this historical installation record.

The public v1 source and six-app build13 release retain their original hashes
and scope. The newer private five-app set is Forge, Measure Pad, Documents,
Document Starter and Notes. It is a separate payload.

## Completed offscreen maintenance

1. `963327a`: optional CI cache upload is separate, bounded and skipped near the
   unchanged deadline. Verification and benchmarks are retained.
2. `6059ea6`: application regressions establish recovery retention after a real
   save followed by a reported durability error, successful retry/Undo/Redo, and
   failed recovery-checkpoint retry without changing primary saved bytes.
3. `f5d6feb`: late settlement cannot reopen an aborted Quit prompt. A failed
   metadata write drops dependent Open requests and reports an explicit retry,
   preserving live data and dirty decisions. Both defects reproduced before fix.
4. `02ee981`: the private producer accepts a validated build version and optional
   verified installation in one command, with a source-bound DELIVERY receipt.
5. The current documentation checkpoint aligns restoration, prose/native chrome,
   packaging and this resumption plan with the implementation.

Review records are in [mac-offscreen-closeout](reviews/mac-offscreen-closeout/).
Focused Go 1.25.6 race tests and vet passed for Notes; focused root document fault
regressions passed. Tooling checks passed for cache budget/dispatch, private
producer/installer, app packaging and SDK verification. There is no new root
stable API, native facade or dependency-pin change.

The installation extension was also exercised against a copied, unchanged
build22 SDK and all five ZIPs in a fresh temporary installation. Archive hashes,
SDK source, extracted bytes and actual strict codesign verification passed.
No app launched. This is compatibility evidence for the existing payload, not a
new-source app build or physical acceptance.

## Repeatable developer commands

Use Go 1.25.6 as declared in go.mod. For portable Notes ownership/history/file
checks, no native window or driver initialization is needed:

```sh
(cd examples/macapps && go test -race ./notes && go vet ./notes)
go test -race ./document -run 'Test(Atomic|FailedInstallAndWrittenError|CancelAfterAdmission|WindowClose|ChooserFailure)'
python3 -B scripts/ci-cache-budget_test.py
python3 -B scripts/check-stages_test.py
python3 -B scripts/private-macos-bundle_test.py
python3 -B scripts/install-private-macos_test.py
```

For a new matched SDK/app candidate, commit first, prepare pinned native sources,
source drivers/cogent/env.sh and use the
[one-command build/verify/install recipe](private-consumer-bundle.md#build-verify-and-install-without-desktop-checks).
Use a fresh output and versioned nonsynced installation. Nothing launches.
BUNDLE_MANIFEST records completed packaging; optional installation is complete
only when DELIVERY.json binds the installed receipt. Keep that source identity
with any later native results. Hosted CI is a separate exact-commit result.

## Remaining native checks — deferred

Do not launch, activate, capture or send input to test apps while desktop work
is deferred. Automated owner/worker and filesystem results cannot establish
physical targeting, native chooser focus or Finder/LaunchServices behavior.

The bounded build22 native Quit case already passed: Save A then Cancel B kept
both seven-note Documents windows open; A's successful Save remained on disk.
At deferral A was saved and B dirty. Those are capture-time facts, not a claim
that those sessions or temporary files still exist. Other app sessions had been
retired normally, so no final post-test relaunch of all five is claimed.

When the owner authorizes a new quiet window:

1. Inspect the exact admitted app, version and disposable data paths. Determine
   whether the partial A/B fixture still exists. Preserve any unsaved state;
   never replay the fresh six-note runner blindly against seven-note windows.
   If necessary, prepare new isolated fixtures without touching user notebooks.
2. Resolve the partial B save/cancel decision safely before changing versions.
   Use a newly source-bound candidate if validating the lifecycle corrections.
3. Check external-file restart/active-window restoration and native geometry,
   then cold/warm Finder opening and dirty Save/Discard/Cancel decisions.
4. Check current-note Find/Replace keyboard routing and one-step Undo/Redo,
   generated Document Starter opening, and Forge study opening/restoration.
5. Leave ordinary app launches only if requested; record exact sources and
   separate physical, accessible-action, automated and human evidence.

The private build22 delivery contains HANDOFF.md and verification receipts. The
private native artifact area contains native-quit.json, NEXT-NATIVE.md and the
stopped runner logs; its desktop deadline is zero. Temporary fixtures/helpers may
have expired. Check them before use; the stopped run is neither an app failure
nor a pass for the unexecuted checks.

## Supported behavior and known limitations

Notes-family apps are plain-text, bounded local notebooks. Recovery takes
priority over external restoration; malformed/missing data is preserved rather
than silently overwritten. Restoration covers fixed startup windows, with
session-local Undo/Redo. Finder examples accept one notebook/study per request
and register Alternate handlers without replacing system defaults. Find is
literal Unicode in the current note, with no regex/cross-note replacement.

No cloud sync, file watching/merging, arbitrary window-set restoration or rich
text is implied. Same-window placement and experimental Forge/GPU exclusions
retain their existing contracts. Human IME/VoiceOver, other hosts/hardware,
downloaded-quarantine launch and notarization remain unqualified. Consult the
[stable API contract](../specs/stable-api.md) and [Mac support guide](macos-v1.md).
