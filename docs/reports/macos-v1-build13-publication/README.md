# macOS v1 build13 — published

The owner approved the final packet and requested merge/push/publication.
The reviewed checkpoint5bc55e4 was fast-forwarded to main and pushed. The separate
[macos-v1.0.0-build13 release](https://github.com/dshills/vellum/releases/tag/macos-v1.0.0-build13)
is public with exactly one SDK, six apps and SHA256SUMS. Its annotated packaging
tag resolves to the common green build source64bf9cf. The original v1.0.0 source
tag/release and alpha.3 release remain unchanged.

[Exact publication/authorization record](../../../release/macos-v1-build13-publication.json)
· [Published checksums](SHA256SUMS) · [Actual public download checks](download-checksums.txt)
· [Byte/metadata result](public-download-verification.json).

All eight files were downloaded from the published release. The seven SDK/app
payload hashes match the approved packet exactly; the native signatures and
bounded launch/exposure tests from the [build13 checkpoint](../macos-v1-build13/README.md)
are reused for those identical bytes. The release body matches the reviewed
SHA256 exactly. Existing releases' titles, bodies, source targets and asset
identities are unchanged. [Remote tag/main refs](remote-refs.txt) were checked
before this documentation-only receipt commit; that file does not claim to
identify the later main HEAD. No additional app/performance campaign was run.

## GitHub filename adaptation

GitHub replaces spaces with dots, including after an asset-name update. Thus
Vellum Notes, Mac Widgets and Mac Shell download filenames contain dots. Only
those three filename entries in SHA256SUMS were adjusted to the actual names;
all seven archive hash values and the reviewed release text are unchanged.
The published checksum file has SHA256
`be4bfaa1c05baac1414ae3a295df312fe5effadc0591ec717ddbb02864f76a06`.
The exact original staged checksum remains historical atc9ba736e…; no historical
packet, local original archive or checksum was overwritten. Use this release's
checksum asset with its downloaded filenames. All seven rows report OK.

The eight assets were verified while in the owned draft and again after public
publication. GitHub's draft-by-tag lookup returned404 and later used its internal
untagged draft name; numeric release/asset IDs provided readback and downloads.
The final publication explicitly sets the approved tag and exact64bf9cf target.
No duplicate release, private-input upload or extra payload was introduced.

Apps remain **ad-hoc signed and unnotarized developer downloads**. Downloaded
quarantine, human IME/VoiceOver and other hosts/displays remain unqualified.
Features, stable API, backend, responsiveness policy and audit decisions are
unchanged. Historical release/version metadata and preauthorization reports keep
their capture-time false/private flags; they do not grant approval to future
payloads. This exact-payload publication record supplies the later owner decision.
[Publication documentation review](../../reviews/macos-v1-build13-publication/README.md).
