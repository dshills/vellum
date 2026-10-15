# Font notices and upstream native artifact identity

The SDK now retains conventional OFL font notices as supplemental declarations.
Owned native patch 0063 restores the verbatim Noto Sans v2.015 OFL text beside
the pinned subset fonts. The notice collector previously skipped that filename.
OFL entries cannot satisfy a Go module's own-license requirement; public release
and the overall third-party/native distribution audit remain gated.

The notice comes from Noto source revision
`c4a321e123e4d4ff315f57f4e0adf294fe3a95be`:
[upstream declaration](https://github.com/notofonts/latin-greek-cyrillic/blob/c4a321e123e4d4ff315f57f4e0adf294fe3a95be/OFL.txt).
Its 4,396 bytes hash to
`cee9892f9f0cc8fe882c9e9537ee6a89621d86ee7ceaf70b02e2b2b1c25c061a`
and match the source tree's Git blob. All six pinned Noto faces identify version
2.015 and the same copyright in their Windows Unicode name records. The existing
README describes subsetting; this checkpoint does not reconstruct or certify
that historical operation or match the subset bytes to an original font build.
Roboto Mono's six existing font files and Apache notice remain unchanged.

Fresh dependency preparation adds exactly one OFL file: all 6,216 prior prepared
leaves remain byte-identical, with 6,217 afterward. Font bytes and the
`//go:embed *.ttf` declaration are unchanged. Source review also caught the
preparation test's old patch-order limit; the final test includes patch 63.

The final nine-source pre-report endpoint passes 11 preparation, 29 SDK producer,
11 distribution inventory and 20 SDK checker tests. The new nested-OFL test fails
against the unchanged baseline collector, confirming the original omission.
The final private SDK has 7,212 manifest files and SHA256
`1462bc20d96ed1206ddafa5200e0ba07de42b5f30ab77edf53652dc282a9a86f`.
Its prepared-source and THIRD_PARTY_NOTICES copies both contain the exact OFL
bytes. Relocated standalone consumer race/vet/build and SDK contract race/vet
checks pass; producer and checker wrapper times are 5.613 and 46.537 seconds.
Earlier eight-source artifacts remain historical, preceding the patch-order test
correction. This report is written after the final private SDK checks.

The prior [native linker checkpoint](../macos-link-closure/README.md) could not
download upstream release artifacts. New successful HTTPS downloads establish
that the pinned Darwin arm64 `libwgpu_native.a` is byte-identical to
`lib/libwgpu_native.a` in the official
[wgpu-native v29.0.1.1 release](https://github.com/gfx-rs/wgpu-native/releases/tag/v29.0.1.1).
The ZIP SHA256 matches GitHub's published asset digest
`a5797a37b1adf720bcd5dcffb291edbbd5b7b14be0a3874c28e6393a655a7a3e`;
the library retains SHA256
`7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248`.
The release's digest-matched commit declaration names
`6aed50955d934ac36049ba8d002034841633ae02`. Cargo.lock, Cargo.toml and both
root license files match Git blob identities at that revision. The lock contains
140 package records; wgpu-core/hal/types and naga lock to 29.0.3. These records
are not proof of the compiled feature/dependency closure or an independently
reproduced build. No signature or release attestation is verified here.

Required checkpoint paths: [qualification](records/qualification.json),
[raw evidence](records/evidence.tar.gz) and [archive index](records/archive-index.json).
All three are required before commit. Downloaded native code is never executed;
the large upstream ZIP, private SDK archives and native library are omitted from
the evidence bundle, with their identities retained. No live window, clipboard,
input source, VoiceOver, OS preference, signing or publication is involved.
Native/font redistribution approval, complete Mac/P6/P11 acceptance and the
failed native endurance/overnight checks remain open.
