# Mac document workflow maintenance handoff

Current source and package status is summarized in [status.md](status.md).
Features, backend and rendering/pacing remain within the existing frozen scope.
The blanket desktop deferral recorded on 30 October is historical: later scoped
Documents sessions, a consumer-app check and operator focus checks are recorded
separately. They do not qualify every app or authorize unrelated desktop actions.
Use CONTRIBUTING and the current task instruction for review/publication scope;
older handoff permissions are not new authorization.

## Source and package identity

The [candidate index](../release/private-macos-candidate.json) is the checked source
of the selected private artifact identities. Its marked table and the resumption
commands are generated; use `python3 -B scripts/private-candidate.py --write-docs`
when deliberately selecting another candidate, then review the index/evidence diff.
Ordinary checkout merges do not relabel an existing package or its test results.

<!-- vellum-candidate:facts:start -->
| Record | Identity |
| --- | --- |
| Candidate | `build25-ed4f71f49bbf` |
| Packaged source | `ed4f71f49bbf623dd8e205571669539212068c62` |
| SDK SHA256 | `65c07945d2b153e7e0f7cfb64a0c3631807e7c9ce0b4e0cb26ff862a74fb718e` |
| Bundle manifest SHA256 | `4af562c72c7e5d53f572a10aa169430f100fb9fe65d97b303c5af8495d532f34` |
| Installation receipt SHA256 | `cae57baf1822fe35dd78bea79075c54ef1ec35c1021eec2b36bc53ff3cb0ad85` |
| Preflight SHA256 | `ed90a8f690ed4095c9ccb926bbd6732901459322d71f2f70af46743919c3840b` |
| Installed apps | `~/Applications/Vellum/build25-ed4f71f49bbf/apps` |
| Minimum macOS | `27.0` |
| Inspected host | `27.2` |
| Desktop checks | `deferred` |
<!-- vellum-candidate:facts:end -->

The table preserves producer-time facts. Build25 packaging/installation did not
launch app windows, hence its unchanged `deferred` label. Subsequent
[build25 Documents sessions](reviews/next-five-stability/item-5.md) and
[application-readiness work](reviews/application-readiness/README.md) have their
own exact source/app identities. Do not replace producer receipts with later
runtime claims or interpret those labels as a blanket current deferral.

Build25 requires macOS 27.0 and was inspected on 27.2; the later Documents sessions
ran on 27.2. These observations do not establish older-host compatibility or
acceptance of all five apps. Historical 26.6.2 qualification does not transfer.

Current checkout/merge and CI state belong to exact Git commits; consult
[the current status](status.md). Earlier post-build24 integration at 7505e730
passed its recorded workflows, but is not CI evidence for later tags. Historical
[packaging reliability reviews](reviews/mac-packaging-reliability/) retain their scope.

The public v1 source and six-app build13 release retain their original hashes
and scope. The private five-app set is Forge, Measure Pad, Documents, Document
Starter and Notes. It remains a separate ad-hoc, unnotarized private payload.

## Historical offscreen maintenance checkpoints

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
   checks pass. This fix is now included in build24; native testing remains deferred.
4. `2888cb31`: [dependency recheck](security/dependency-alerts-2026-10-30.md) preserves
   original notices and keeps alert 4 open with its limited disposition. A focused
   evidence guard requires reassessment on changes. The source-built component
   inventory does not certify build23's different prebuilt native archive.
5. This documentation phase updates exact-main/package/source distinctions and
   adds the [safe desktop resumption checklist](mac-desktop-resumption.md).

Each checkpoint has a [recorded internal review](reviews/mac-offscreen-maintenance24/).
Existing build23 manifests/receipts and historical native evidence are unchanged.
No external review, dependency pin, stable API, release gate or backend change.

## Build24 delivery closeout

1. `b9212489`: reviewed integration and push, followed by exact-source green CI.
2. `9b61d7a2`: fresh matched SDK/five-app producer and versioned installation;
   Go 1.25.6, relocated consumer race/vet/build and extracted signatures passed.
3. `7a5e060f`: independent durable-delivery verifier passed source/hash/receipt,
   app-metadata, installed bytes and ten strict signature checks without rebuild.
4. `1dd1b88f`: reusable static linkage checker and six focused tests. All five
   installed apps are thin ARM64 with 23 system dependencies each, no rpaths,
   system dyld, no embedded environment overrides or extra Mach-O binaries.
   Deployment metadata has the 27.0 limitation above; runtime is not executed.
5. This final documentation checkpoint binds exact identities, retained versions,
   testing boundaries and the deferred acceptance cases. The checker added after
   packaging runs from the current checkout; it is not relabeled as SDK content.

[Build24 packet summary](reports/mac-build24/README.md) and
[internal reviews](reviews/mac-build24/) distinguish packaging, independent
verification and static metadata evidence. Historical reports remain unchanged.

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
python3 -B scripts/check-private-macos-linkage_test.py
python3 -B scripts/dependency-disposition_test.py
```

For a new matched SDK/app candidate, commit first, prepare pinned native sources,
source drivers/cogent/env.sh and use the
[one-command build/verify/install recipe](private-consumer-bundle.md#build-verify-and-install-without-desktop-checks).
Use a fresh output and versioned nonsynced installation. Nothing launches.
BUNDLE_MANIFEST records completed packaging; optional installation is complete
only when DELIVERY.json binds the installed receipt. Keep that source identity
with any later native results. Hosted CI is a separate exact-commit result.

To recheck the selected installed candidate bytes without reinstalling or starting any
app, use the [existing-delivery verifier](private-macos-installation.md#verify-an-existing-delivery-without-reinstalling).
The [resumption checklist](mac-desktop-resumption.md) includes exact local paths
and the caller-pinned bundle manifest digest. Static verification does not
establish physical behavior of the included lifetime fix.

## Historical build22 deferred session — 30 October 2026

The instructions below preserve that stopped session, not the current project
status or an instruction to assume its windows still exist. In that session,
desktop work was deferred. Automated owner/worker and filesystem results cannot establish
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
