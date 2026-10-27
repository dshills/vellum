# Current macOS delta audit and publication packet

Prepared 27 October 2026 for owner review. **Audit approval and binary publication
remain pending.** Features, stable API, backend, rendering/pacing and release
responsiveness policy are unchanged. No target app was launched, user notebook
read, global setting changed, public asset uploaded or tag moved.

## Concrete scope and identity

Code source: `4b42fb89948b8abf0f2f65970fb5980987c1af0c` (`v1.0.0`). The new
private audit SDK and six app candidates all use that committed source and the
previously approved ARM64 native archive. Notes, Mac Widgets and the two-window
Mac Shell retain their existing identities; Forma, Luma and Pulse are the three
showcase apps. Other example sources remain in the SDK; no additional binary
product scope was invented.

The reference is the **actual published alpha.3 SDK** `28a95b5c…`, not an earlier
private alpha candidate. Its full archive integrity, native archive and original
418-file/175-entry app notice bundle were checked against the owner approval in
[release/macos-preview-audit.json](../../../release/macos-preview-audit.json).
Current candidates do not inherit approval automatically.

Private candidates and exact pre-sign inputs are retained in the research
workspace's `deliveries/macos-v1-distribution-audit-20261027/`, outside Git.
They are ZIP/tar archives so this host's synchronized-folder metadata does not
alter sealed app contents. Temporary build directories are not the handoff.
[Candidate packet](candidate-packet.json) records exact SDK/app/input hashes;
[CANDIDATE_SHA256SUMS](CANDIDATE_SHA256SUMS) identifies private audit assets.
These are not public-ready download checksums.

## Delta findings

- **Go graph:** 89 effective SDK module records before and after. No added,
  removed or changed module/version/sum/declaration identities.
- **Native closure reference:** the exact `786aaad8…` source-built WGPU archive
  and its 82-component original notice tree are unchanged. Each new app link map
  names that exact SDK archive; object/member observations are retained. This is
  producer-observed build evidence, not independent compiler attestation.
- **SDK files:** 405 added, none removed, 107 changed versus published alpha.3.
  Most are first-party source/docs. Prepared Core has three new and six changed
  files from the reviewed prose-editor patches; typesetting has no file delta.
  [The complete file delta](sdk-delta.json) retains every path.
- **Prepared-source trust:** fresh checksum-verified public dependency preparation
  exactly matches all 6,033 Core and 299 typesetting files in this private SDK
  candidate. No ignored-tree difference is deferred to public production.
- **Selected embeds:** all six app targets select the same 4,441 embedded asset
  identities as their reference. Notes and Mac Widgets have like-for-like target
  comparisons. Mac Shell and the new showcases did not exist in alpha.3; their
  comparison uses approved Placement as a reference, not a claimed earlier build
  of the same app. Generated assets need the separate checks below.

### One actual notice gap corrected

The three showcase apps compile Go Regular/Bold TTF byte arrays from
`golang.org/x/image v0.44.0`. Those bytes are generated Go source, so an inventory
of `go:embed` alone does not identify the addition. A non-UI proof utility hashes
both compiled byte arrays; each matches its original pinned-module TTF file.
[Direct binary inspection](font-binary-proof.json) also finds each full TTF byte
sequence in all three actual pre-sign executables.

The module-root Go Authors notice was already present, but the original
[Bigelow & Holmes font-family notice](Go-fonts-NOTICE.txt) was absent from the old
418-file bundle. Its exact SHA256 is
`dc588a7ba9130043b67c2380cc03cc6644ed3557f6d5e8589b96acba1c64e343`.
That unmodified original text is now in Forma/Luma/Pulse's binary-bound notice
bundles, bringing them to 419 files. The first three apps retain 418. Only those
three notice-bearing app packages were repackaged and reverified; code and
pre-sign executable bytes did not change. These old private demo packages remain
historical and are not relabeled as corrected packages.

[Upstream font notice](https://raw.githubusercontent.com/golang/image/v0.44.0/font/gofont/ttfs/README).
The SDK does not ship the external x/image TTF/source package; example consumers
resolve that checksum-pinned module separately. The font-specific notice is
included in the three apps that actually link those font bytes and in this packet.
Procedural landscapes, poster composition and full-color bundle icons are
first-party source-generated content with the examples' original MIT declarations.

App inventory component counts include extra explicit current-graph/owned-module
reference aliases. They are not counts of additional third-party modules/crates;
the 89-module and 82-native-component comparisons above are the relevant deltas.

## Candidate app archives

All are private, ad-hoc signed, unnotarized build10 candidates. The SDK candidate
SHA256 is `e6556ceaefb9b6d4e1fb911598f80fb33d56ebb5a4b32c89dd01b40267fa9e9a`.

| App | Original notice files | Candidate ZIP SHA256 |
| --- | ---: | --- |
| Vellum Notes | 418 | b0a450c85fc21b562755362c41e05a8f69772a2656060803bdf3750ebc1b1f68 |
| Vellum Mac Widgets | 418 | 5c6ec0f69c62490e86bd080b03a9e11fb910928c5718a6aebfbc58f5e8067672 |
| Vellum Mac Shell | 418 | 963dd446d90a88478b1d93060c37a2a5cd418b03a830238835d768508d83347f |
| Forma | 419 | 0af6d638ab300c728945d726d1c5825f5bfcb9cd5d582e4afe9dd0c2e55faef8 |
| Luma | 419 | 479d070b96ad36108be6da6e98e731c80b75d5a2a728f3aa5f586f4666f7354b |
| Pulse | 419 | 5311bc88c8cc22583ada0681641c83f565955cc1d71955222bd985dbd8e3f0fb |

Exact pre-sign binary, signed binary, notice-inventory and link/input-report
hashes are in the candidate JSON. The local `PRE_SIGN_INPUTS.tar.gz` preserves
pre-sign executables, original showcase icon PNGs and per-app inventories, plus shared original notice bodies;
reconstruct only the members named in each inventory. It is private build input,
not an asset proposed for public distribution.

## Verification and reused evidence

The recorded sequence has 67 successful graph/origin/build/package,
extraction/signature and independent-starter command steps across two phases:
54 initial-candidate and 13 font-notice-correction steps. Every check has a unique
phase-qualified ID. Twelve initial packaging checks are superseded by the corrected
three app bundles; 55 checks directly support the final candidates. Origin checks keep all nine
SDK module origins inside the extracted SDK; the driver itself is correctly
checked as the main module plus eight replacements. Each app was built with an
external linker map. ZIP extraction, strict signatures and binary-bound original
notices were verified. SDK payload SHA256 identities are rechecked after builds.
The independently generated starter passes race testing and builds. This is a
focused packaging/consumption pass, not a fresh framework qualification campaign.

[Existing CI](reused-ci.json) is green on this exact code source (Framework,
Platform adapters and Stable root API). Reuse [private v1 acceptance](../macos-v1-release/README.md)
and [showcase checks](../../reviews/showcase-apps/README.md) only within their
recorded source/binary/native scopes. No native interactions were repeated here.
The new six binaries require a bounded final launch/close check before public
handoff, especially the showcase apps now using the reviewed SDK native capsule.
Human IME/VoiceOver, other Mac hardware/versions/displays and wider platform use
remain unqualified. No performance gate has changed.

Both real public producers were attempted against clean source with valid inputs.
[They refused output](guard-proof.json) specifically because the distribution
audit is pending. No public output directory/archive was created. Existing audit
and publication metadata remain byte-identical; no gate was bypassed.

## Decisions still required

[Rights dispositions](rights-disposition.json) distinguish unchanged reviewed
inputs, first-party changes and the added Go-font material. The pinned
[objc2 upstream caveat](https://raw.githubusercontent.com/madsmtm/objc2/objc2-0.6.4/LICENSE.md)
still exists; the same native bytes have the prior owner's alpha.3 decision, not
an inferred new legal grant. Review whether that disposition covers the current
payload and obtain additional rights review if it does not.

1. Owner completes the current-payload distribution-audit decision using this
   packet; only then record approval in a new release-specific audit record and
   the existing release metadata. Preserve the historical alpha.3 record.
2. Choose the app distribution route: an explicitly disclosed unnotarized
   developer download, or a separately authorized Developer ID/notarization pass.
   No credentials, signing identities or Apple submissions were used here.
3. Produce final candidates through the unchanged `--public` SDK/app guards,
   attaching the exact matched notices. Approval/manifest/source-metadata and
   signing changes mean final public archive hashes can differ from these private
   audit candidates. Compare product/dependency contents; do not restart unchanged
   component review or mislabel private bytes as public.
4. Verify final public extraction/consumer/launch results, update the final asset
   hashes and release text, then obtain publication approval for those exact files.
   Keep the existing v1.0.0 source tag immutable; disclose any separate packaging
   revision and choose the release destination explicitly.

[Guarded production recipe](REPRODUCE.md) records reconstruction and the public
commands to use after approval. [Draft release text](PUBLICATION_DRAFT.md) is deliberately blocked and contains
no claim that these private files have already passed public production.

## Reproducibility and initial controller failures

[evidence.tar.gz](evidence.tar.gz) and [its member index](evidence-index.json)
retain SDK manifests, detailed selected build inputs/link observations, original
notice texts, per-app notice inventories and the actual local controllers/logs.
The archive contains source paths; it excludes executable/SDK/app archive bytes.
Controllers are historical audit evidence, not a supported unattended publisher.
Their exact code and assumptions are inspectable rather than silently discarded.

The first export refused a depleted/modified old temporary module cache. A fresh
isolated cache was seeded only with downloaded archives, the pinned Go1.25.6
was bootstrapped, and Go module verification then passed. No verifier was weakened
and the old cache/failure was preserved. Initial controller assumptions also
misapplied the external-consumer self-replacement rule to the driver main module
and assumed the Mac Shell target existed in alpha.3. Both were corrected; the
already completed SDK/Notes/Widgets work was reused. Final receipts distinguish
successful checks from retained earlier controller failures. The new font notice
was found during explicit generated-asset inspection and corrected before final
candidate hashes. No failed attempt is called successful.
