# macOS alpha.3: guarded public-mode candidates, still unpublished

23 October 2026. The owner approved the exact-payload distribution audit and
source/draft-PR CI. That decision is recorded at `ed35948` in
[the audit record](../../../release/macos-preview-audit.json). Existing guards
remain mandatory. Eight focused approval/notice/source/drift tests pass.

Both local producers now succeed in `--public` mode. They use the approved
native archive `786aaad8…`, original app source binary `0d753959…`, and corrected
notice inventory `0efc1c58…` (418 files, 175 entries). Collection inventories retain
historical pending/review-required declarations; the owner supplies release-level
audit approval. The verifier does not independently certify distribution rights:
its `public_distribution_audit: not certified` field describes that verifier scope,
while the manifest correctly records the separately approved owner decision.

## Candidate bytes and checks

| Candidate | SHA-256 |
| --- | --- |
| `vellum-v0.1.0-alpha.3-darwin-arm64-sdk.tar.gz` | `28a95b5c2354673d97161df30e828008a5aa18465ce18545a42e8caf0a82f269` |
| `vellum-v0.1.0-alpha.3-macos-arm64.zip` | `bd58d2eb3d94486e07683a289a77299dde03fd450bd59bd65240f83d6d2f0e50` |
| Build-29 signed app executable | `32e00cd8f249c9d175a30a42ab5d0b3d57a9f86d393831aa7b9c26366230d342` |

SDK source is clean `ed359484a5be720948e92430f7496b36ec5cd352`.
Fresh relocation verifies 8,451 file hashes and contained replacements, external
module cache verification, standalone consumer race/vet/build, public API
race/vet, and native window/readback/normal exit using Go 1.25.6. The readbacks
are functional checks, not new latency/cadence measurements.

Product Go/native/dependency inputs match the prior private SDK except for the
already-reviewed soak command's cover inset. Public dependency preparation also
omits 17 renderer testdata PNGs present in the private snapshot; the changed-file
list is retained. Those image fixtures belong to test assertions, not Placement
runtime embeds. No application/runtime feature or rendering policy changed.

The build-29 app is **ad-hoc signed and not notarized**. Fresh ZIP extraction and
strict signature verification pass. Native app-only CGEvent input plus CUA
accessibility/keyboard verified launch, one Block placement, pan/zoom, Undo/Redo
identity, native Save As, New/Open with exact saved geometry/camera/snap, and
ordinary Quit. The exact owned PID retired and strict signature verification still
passes. These are native automation checks, not human physical pointer acceptance.

The final ZIP changes only adjacent README wording from the first tested ZIP:
all 423 files in its freshly extracted app bundle have identical hashes to that
tested bundle. The final bundle file manifest and its digest are retained. The
original test ZIP is preserved locally; the final hash above names the proposed
release asset. This reuses applicable unchanged-code verification rather than
claiming a second set of native interactions occurred at the final extraction path.

Prior [app acceptance](../macos-preview-app-acceptance/README.md),
[100-cycle checks](../macos-preview-lifecycle100/README.md),
[activity/document safety](../macos-preview-resource-check/README.md), and
[matched two-hour soak](../macos-preview-soak-two-hour/README.md) retain their scopes.
Full P11/native/GPU heap, human VoiceOver/IME and physical input, other machines,
cross-display/high-refresh and production qualification remain open. Historical
latency disclosures and the preview policy are unchanged.

## Publication boundary

No release tag, GitHub prerelease or downloadable asset was published. Release
notes and SHA256SUMS are prepared locally. The final tag target and publication
request remain dependent on applicable candidate CI and separate owner approval.

Source push/draft PR is also currently held by automatic approval review: older
branch history contains unredacted local home paths beyond the normalization
stated in the original question. A read-only fetch confirms 80 changed files have
157 added path-bearing text lines relative to public `origin/main` at `0aa31bf`;
this count is not a complete archive privacy inventory. The owner has been asked
whether to publish those historical paths or prepare a sanitized publication
branch. No alternate push/API bypass was attempted. Historical evidence hashes
and branch history are preserved locally. This is a source-publication decision,
not a new runtime defect or audit reversal.

`INDEX.json` binds the records. New report absolute directory/tool prefixes are mapped to descriptive
placeholders; original local receipts remain unchanged. Artifact and file hashes,
measured results, test-only process IDs, timestamps and geometry are unchanged.
This presentation normalization does not resolve the separate historical branch
privacy hold.
