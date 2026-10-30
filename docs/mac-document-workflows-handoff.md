# Mac document workflow maintenance handoff

As of 30 October 2026, features and the existing backend, rendering/pacing and
release/performance policies remain frozen. Desktop checks are deferred by the
owner. Current work uses internal review; it authorizes no external review,
public release/tag/assets, new audit decision, signing identity or notarization.

## Source and package identity

- Main at `fa85ebe46da7b9cce836f865a171b2738ffaa769` is pushed and green in
  [Framework](https://github.com/dshills/vellum/actions/runs/41073526868),
  [Stable root API](https://github.com/dshills/vellum/actions/runs/41073526779)
  and [Platform adapters](https://github.com/dshills/vellum/actions/runs/41073526784).
  All three passed their original attempts. This includes the prior build23
  compatibility/fuzz/preflight/generator work; it is no longer an unmerged branch.
  These are exact-main results, not hosted results for subsequent maintenance.
- Current local work is on `codex/mac-offscreen-maintenance24`. It adds installer
  failure coverage, a read-only delivery verifier, an app-local lifetime fix,
  dependency applicability checks and this handoff refresh. Use `git rev-parse HEAD`
  for its exact documentation checkpoint. It has focused local verification;
  it has not been merged/pushed or hosted-CI-qualified by this checkpoint.
- **Final private build23 source:**
  `e0e011c605f8ee6ea84827414a1636bcca74c97e`.
  SDK SHA256:
  `a9857bff74590959d688e2827af2918d87468b1667773b6a278123f8fb4cbc75`.
  Installed apps: `~/Applications/Vellum/build23-e0e011c605f8/apps`.
  The matching SDK and all five app archives were rebuilt from this one committed
  source, copied to durable private delivery storage, and independently verified.
- BUNDLE_MANIFEST, PREFLIGHT, DELIVERY, INSTALLATION, VERIFICATION and SHA256SUMS
  identify the exact bytes, sampled Go/compiler/native inputs and installation.
  Repeated installation reverified bytes/signatures at the same target. The
  source-bound preflight digest is
  `b511219651fde2e0d358d5720877310aacc8cdd9f5c0131554246b2262a5642f`.
- Initial source-df07cd4 build23 and older build22 installations remain retained.
  Use the final e0e011c path above. A Documents-based producer attempt failed
  strict signature checks; a later source-81b3f19 attempt failed generated test
  fixture verification. Neither failed attempt produced a completion/delivery
  manifest or installed its incomplete payload.

No app window was launched during this work. Forge's headless icon-export command
returns before application/UI/data initialization. Native launch/input checks
remain deferred; no current app liveness or post-test relaunch is claimed.
The installed build23 contains f5d6feb (Quit/deferred Open). Newer source commit
`e589f36e` additionally refuses late actions after their exact window retires,
including when another window reuses the ID. That app-local correction is tested
in Notes and a generated document consumer, but is NOT in installed build23.
No newer SDK/app candidate has been packaged, installed or launched in this pass.

The public v1 source and six-app build13 release retain their original hashes
and scope. The private five-app set is Forge, Measure Pad, Documents, Document
Starter and Notes. It remains a separate ad-hoc, unnotarized private payload.

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

## Build23 closeout

1. `fd239b9`: recorded the initial verified build23 SDK/app installation.
2. `77a5f26`: historical build13/build22 Notes, recovery, metadata and build22
   Forge fixtures retain original encoder/generator hashes. Current readers and
   reverse historical-decoder checks preserve the fictional samples. No arbitrary
   downgrade compatibility or personal-document migration is implied.
3. `f6b5320`: five capped fuzz targets passed bounded race-enabled mutation runs;
   ordinary seeds, full byte-limit checks and accepted-model/history invariants
   also passed. No parser runtime fix or schema change was needed.
4. `81b3f19`: producer preflight rejects wrong Go/target/cgo, incomplete or active
   native preparation, existing/source output and known synced paths before
   reserving output. It records and rechecks sampled Go/compiler/native hashes.
5. `e0e011c`: the SDK integration check exposed omitted generated testdata.
   The reviewed correction copies fixture bytes/provenance unchanged and refuses
   links/special files; standalone and relocated generated-consumer checks pass.
   Final build23 was then rebuilt and installed from this corrected source.
6. The final documentation checkpoint records exact delivery, checks and deferred
   native work, without changing the packaged source/bytes.

[Internal reviews](reviews/mac-build23/) record scope, findings, fixes and checks.
The final producer passed SDK all-file/relocation/module-origin checks, minimal
and document consumer race/vet/build, public API scenarios, all five ZIP-extracted
and installed strict signatures, build-number/identifier checks, six archive
hashes and preflight/installation identity. These checks did not execute the
native UI. Known AppKit deprecation/linker warnings did not fail packaging.

## Subsequent offscreen maintenance

1. `42873dc8`: six installer fault boundaries and a killed synthetic extractor
   preserve the prior version and allow retry. Hidden staging debris after abrupt
   termination is documented; no power-loss guarantee or automatic data rollback.
2. `faa14efe`: `verify-private-macos.py` independently verifies an existing complete
   SDK/five-app delivery and its selected installation without build/install/launch.
   Actual final build23 passed all-file/hash/source/receipt/app-metadata checks
   plus strict signatures for all five archived and installed apps.
3. `e589f36e`: overlapping Save/Open/Quit/Close tests reproduced late Open settlement
   creating a prompt on a retired window model. Exact-window lifetime checks at
   action admission and completion fix it; four late-action cases cover Open,
   Save, Close and Delete. Full Notes race/vet and focused generated-consumer
   checks pass. This source-only fix requires a newer package for native testing.
4. `2888cb31`: [dependency recheck](security/dependency-alerts-2026-10-30.md) preserves
   original notices and keeps alert 4 open with its limited disposition. A focused
   evidence guard requires reassessment on changes. The source-built component
   inventory does not certify build23's different prebuilt native archive.
5. This documentation phase updates exact-main/package/source distinctions and
   adds the [safe desktop resumption checklist](mac-desktop-resumption.md).

Each checkpoint has a [recorded internal review](reviews/mac-offscreen-maintenance24/).
Existing build23 manifests/receipts and historical native evidence are unchanged.
No external review, dependency pin, stable API, release gate or backend change.

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
python3 -B scripts/verify-private-macos_test.py
python3 -B scripts/dependency-disposition_test.py
```

For a new matched SDK/app candidate, commit first, prepare pinned native sources,
source drivers/cogent/env.sh and use the
[one-command build/verify/install recipe](private-consumer-bundle.md#build-verify-and-install-without-desktop-checks).
Use a fresh output and versioned nonsynced installation. Nothing launches.
BUNDLE_MANIFEST records completed packaging; optional installation is complete
only when DELIVERY.json binds the installed receipt. Keep that source identity
with any later native results. Hosted CI is a separate exact-commit result.

To recheck the installed final build23 bytes without reinstalling or starting any
app, use the [existing-delivery verifier](private-macos-installation.md#verify-an-existing-delivery-without-reinstalling).
The [resumption checklist](mac-desktop-resumption.md) includes exact local paths
and the caller-pinned bundle manifest digest. Verification of those old package
bytes does not test the newer source-only lifetime fix.

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
