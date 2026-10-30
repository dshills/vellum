# Mac document workflow maintenance handoff

As of 30 October 2026, features and the existing backend, rendering/pacing and
release/performance policies remain frozen. Desktop checks are deferred by the
owner. Current work uses internal review; it authorizes no external review,
public release/tag/assets, new audit decision, signing identity or notarization.

## Source and package identity

- Main at `b9212489afc921171fc5abbd5275d8919ae930dd` is pushed and green in
  [Framework](https://github.com/dshills/vellum/actions/runs/41091357397),
  [Stable root API](https://github.com/dshills/vellum/actions/runs/41091357457)
  and [Platform adapters](https://github.com/dshills/vellum/actions/runs/41091357554).
  All three passed their first attempts. The previous five maintenance commits
  are merged, including the retired-window action fix at e589f36e.
- **Current private build24 source:** the same `b9212489` green commit.
  SDK SHA256:
  `5a825eb07c480f2df248435b7c0a55e0b0cb5104ccd2de3b8f146c55c2798382`.
  Installed apps: `~/Applications/Vellum/build24-b9212489afc9/apps`.
  The SDK and all five app archives were freshly built from that one source,
  copied to durable private delivery storage, independently verified and installed.
- Bundle manifest SHA256:
  `516456cc39bc0b3c2b8faf14871cc820afda813a0a5f269e23da15fd97c77727`.
  Installation receipt SHA256:
  `1c8a8b8425759434214631ff73b8ec9293af9c89dd82847f111cdea544a39fcd`.
  Preflight SHA256:
  `75c0bcc3e5ecad4d43bd3b0f7e5b7989c147a480992e685a3c4565d4b94009aa`.
  BUNDLE_MANIFEST, PREFLIGHT, DELIVERY, INSTALLATION, PACKAGING, VERIFICATION,
  LINKAGE, CI and SHA256SUMS retain their distinct scopes in the private packet.
- Subsequent report/checker work is on `codex/mac-build24-closeout`. Use
  `git rev-parse HEAD` for its current documentation/tooling checkpoint.
  These newer commits have focused local checks and are not claimed merged,
  pushed or hosted-CI-qualified. They do not relabel the built source or SDK.
- Build23 at source `e0e011c605f8ee6ea84827414a1636bcca74c97e`, SDK
  `a9857bff74590959d688e2827af2918d87468b1667773b6a278123f8fb4cbc75`, and older
  installations remain retained. Do not overwrite, migrate or delete them.
  The original failed producer attempts retain their historical diagnostic scope.

No app window was launched. The Forge icon-only exporter exits before UI/data
initialization. Native launch/input checks remain deferred; no current app
liveness or post-test relaunch is claimed. The e589f36e lifecycle correction is
now included in installed build24 and tested in the SDK's generated consumer;
its physical/native behavior remains unverified.

**Deployment limitation:** all five build24 executables declare minimum macOS
27.0, SDK 27.0, and were statically inspected on the current 27.2 host. They
cannot run on the historically qualified 26.6.2 host. This is a private current-Mac
candidate, not an extension of the existing framework support matrix or a new
qualified release. No target was clamped or support promise changed. Launch on
27.2 still needs the deferred native checks.

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

To recheck the installed final build24 bytes without reinstalling or starting any
app, use the [existing-delivery verifier](private-macos-installation.md#verify-an-existing-delivery-without-reinstalling).
The [resumption checklist](mac-desktop-resumption.md) includes exact local paths
and the caller-pinned bundle manifest digest. Static verification does not
establish physical behavior of the included lifetime fix.

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
