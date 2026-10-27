# Final macOS v1 asset publication packet — unpublished

Local public-mode production is complete. Owner payload audit is approved;
**final asset publication is not yet authorized**. No upload, tag change, merge,
Developer ID signing or notarization is performed by this checkpoint.

## Proposed exact upload set

Use the six entries in [RECOMMENDED_SHA256SUMS](RECOMMENDED_SHA256SUMS): the SDK,
Vellum Notes, Vellum Mac Widgets, Forma, Luma and Pulse. The six filenames/digests
are the concrete proposed asset set. The complete seven-output inventory remains
in [SHA256SUMS](SHA256SUMS); **do not upload Vellum Mac Shell.zip** while its
[contrast issue](SHELL_CONTRAST_ISSUE.md) is held.

The durable files live in the research workspace's
`deliveries/macos-v1-public-candidates-20261027/`, not in Git or a temporary tree.
The SDK is frozen product source 4b42fb8 plus packaging/approval metadata 1fc594b;
apps reuse the audit's exact pre-sign binaries/notices/icons, then receive public
metadata and ad-hoc signatures. Do not claim that the unchanged v1.0.0 tag contains
the later approval/report commits or move that tag to manufacture that claim.

## Draft asset announcement

> Prepared macOS ARM64 developer downloads for the existing v1.0.0 source release:
> a relocatable SDK, Notes, the Mac Widgets catalogue, and Forma/Luma/Pulse showcase
> apps. Product code remains frozen at 4b42fb8; public packaging/approval metadata
> comes from 1fc594b. Verify the provided recommended SHA256 list before extraction.
> Apps are ad-hoc signed and unnotarized; a downloaded-quarantine launch has not
> been qualified. Notes is a local plain-text notebook; the showcases use offline
> fixture data. Human IME/VoiceOver and other hardware/macOS/display configurations
> remain unqualified. The optional Mac Shell fixture is withheld for dark contrast.

## Final owner decision and execution

Obtain approval identifying these exact six hashes, the disclosed ad-hoc/
unnotarized route, the README/release text and the destination (for example,
attaching to the existing v1.0.0 source release with a disclosed packaging revision).
Audit approval alone is not that upload decision. Only after that approval,
recheck the archive hashes, attach the selected files and recommended checksum
list, verify remote byte hashes and update current docs to the actual download
URLs. Do not upload private audit inputs, evidence archives or the held fixture.

The existing root source release/tag remains immutable. Native-module tags are
outside this packet. A separately authorized notarization route would produce
new asset hashes and require a new final packet; it is not a hidden prerequisite
for the currently disclosed developer-download option.
