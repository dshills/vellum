# macOS alpha.3: ready for final publication decision

23 October 2026. The frozen preview has completed scoped R1–R5 preparation,
application/SDK acceptance, bounded resources, exact-payload audit and source CI.
Final publication remains an owner decision. No tag, GitHub prerelease or
downloadable asset was published by this checkpoint.

The owner explicitly approved publishing historical local paths and checkpoint
metadata after the earlier automatic approval rejection. The ordinary push
succeeded and [PR #1](https://github.com/dshills/vellum/pull/1) is open as a draft.
The historical-path source-publication hold is resolved; historical reports keep
their then-current state.

## Source CI and identical code tree

All four push/PR workflow runs pass at
e9907d922faa8816ea79882f472bac6313b91072: macOS Framework (including scripts/check.sh
and existing benchmarks), Linux X11 software-GPU smoke and Windows build/unit
tests. Optional self-hosted qualification was skipped as configured.
[ci.json](ci.json) records run links and exact head.

This final report, release notes and ledger update change only docs/. The
non-documentation Git tree has 2,841 entries and SHA-256
f3d758f27685572638283ed7aa1f0b6d8fd73b510973024388b8bc2339934b10.
Compute SHA-256 over git ls-tree -r --full-tree HEAD output lines outside docs/,
in Git output order with newlines. The commit step verifies equality to the
qualified head. This covers source, tests, workflows, release metadata and original
notice files. No local broad suite or native campaign is repeated for docs.

Applicable CI at this identical code tree satisfies the approved R5 contract.
Automatically triggered later CI retains its own head/status and is not silently
reported as completed here.

## Exact proposed publication

Proposed tag: v0.1.0-alpha.3, targeting the reviewed commit containing this packet.
The final owner request pins its full SHA after commit. Keep PR #1 draft and
unmerged unless separately authorized; no main-branch merge is needed to publish
this preview tag from the reviewed branch.

Three assets: already-tested SDK archive, app ZIP and [SHA256SUMS](SHA256SUMS).
[RELEASE-NOTES.md](RELEASE-NOTES.md) is the exact proposed GitHub prerelease body.
The [public-candidate report](../macos-preview-public-candidate/README.md) binds
archive hashes, actual build inputs, final extraction/signature, native automation
and unchanged bundle-byte reuse. The app is ad-hoc signed and unnotarized; no
signing identity, notarization request or system-setting change is proposed.

The SDK was produced from clean ed35948; the app retains its source binary and
build-29 provenance. Later report commits do not re-label build sources. The tag
adds reviewed evidence and current publication text. Full platform, physical-input,
human accessibility, native/GPU heap and latency limits remain unchanged.

## Concrete alert triage

GitHub surfaced four existing repository alerts during push.
[alert-triage.json](alert-triage.json) records advisory links, manifests, reported
ranges and dispositions. Three concern prototypes/p0/gio/go.mod, excluded from
the SDK; selected x/image v0.44.0 and x/net v0.57.0 are outside those reported
vulnerable ranges. The remaining low-severity rand alert comes from an original
copied itertools Cargo.toml development declaration; rand is absent from the 82
selected compiled native component records. Original notices and dependencies
are unchanged. Alerts remain open; this is scoped triage, not a security audit.

Reproduce the archive exclusion and selected inventory checks with Python 3.11+:

    python3 verify-sdk-alert-scope.py /path/to/vellum-v0.1.0-alpha.3-darwin-arm64-sdk.tar.gz

The script verifies the pinned archive hash, inspects every tar member for
prototype paths, and checks the shipped Go/native inventory versions. The
[recorded output](sdk-alert-scope.json) binds that check to the exact payload.
It does not establish that building the separate historical prototype is safe.
Updating its dependencies remains outside this frozen release checkpoint.

## Final action and budget

After owner approval: recheck exact clean source/tag/asset hashes, create the
prerelease at the pinned commit, upload those three assets, download and verify
published bytes, and launch the verified app. Finish within the existing session
budget; no new runtime or performance work is implied. Session 2 began
2026-10-23 08:30:07 UTC, hard end 16:30:07 UTC. The plan allows three sessions plus
one contingency; session 1 is complete. Approval does not silently extend it.
