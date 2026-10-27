# Final publication packet — build13, unpublished

[PUBLICATION_SET.json](PUBLICATION_SET.json) and [SHA256SUMS](SHA256SUMS) name the
exact **seven new payload archives**, all built from clean green source
`64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a`. Each app is version1.0.0/build13 and compiled from the new SDK,
including stage0108. Use only durable `deliveries/macos-v1-build13-20261027/` files.

## Proposed destination and text

Create a separate GitHub release in dshills/vellum with packaging tag
**macos-v1.0.0-build13** pointing to **64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a**. Use the exact
[proposed release body](RELEASE_BODY.md), seven archives and SHA256SUMS.
The proposed tag has no existing remote match at preparation time; recheck before
creation. This is a proposal only: no tag, release or upload has been created.

The existing v1.0.0 source tag/release stays at4b42fb8 and receives no assets or
text changes. The new macos-prefixed tag matches this SDK/app source and avoids
claiming that older tagged source contains the repair. It is a packaging label,
not a new Go module version. App1.0.0/build13 and SDKv1.0.0 metadata remain intact.
No native-module version tags are proposed.

The baseline approved audit stays byte-identical. The [report](README.md) and
[delta](sdk-delta.json) disclose prepared-Go changes and new binary/archive
identities. Original dependencies, embeds, native archive, artwork and notices
are unchanged; new executable notice bindings verify. No old exact-archive
approval is inferred for these bytes. Both public producer guards are retained.

Apps use the **ad-hoc signed, unnotarized developer-download** route.
Downloaded-quarantine launch, human IME/VoiceOver and other hosts/displays remain
unqualified. No Developer ID, notarization or security changes are included.

## Remaining final decision

The owner must authorize these exact payload hashes, checksum file, new packaging
tag at64bf9cf, destination, release text and route. This checkpoint only prepares,
reviews and commits the packet. After final authorization:

1. Recheck every named local hash and proposed tag/release availability. Create the packaging tag only at64bf9cf. Stop on unexpected conflicts; preserve existing tags/releases/assets.
2. Upload only seven listed archives and SHA256SUMS to that separate release. Exclude manifests, evidence, private pre-sign inputs and old candidates.
3. Download published bytes, verify hashes and update current links/status. Preserve the original v1.0.0 source tag/release and historical audit/report bytes.

[Delivery instructions and limits](DELIVERY_README.md).
publication_authorized=false and uploaded=false remain accurate.
