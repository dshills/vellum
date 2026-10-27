# Vellum macOS v1.0.0 — build13 developer downloads

One prepared SDK and six macOS apps, all built from clean green source
**64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a**. This release's packaging tag **macos-v1.0.0-build13** identifies that same commit.
The pre-existing root Go source/API tag v1.0.0 remains at4b42fb8. The new macos-
prefixed tag identifies these downloads; it does not introduce a new Go module
version. Apps retain version1.0.0/build13 and the SDK retains v1.0.0 metadata.

## Apps and SDK

- **Vellum Notes:** local plain-text notebook, Mac prose layout, folders/search,
  native toolbar, persistence/recovery and Undo/Redo.
- **Vellum Mac Widgets:** Mac-themed control catalogue, native titlebar toolbars,
  app-scoped managed themes and full-color artwork.
- **Vellum Mac Shell:** independent A/B windows and native toolbars; the retained
  scene restores correctly when an edited sibling closes.
- **Forma:** original poster/layer studio with pan/zoom/fit and Undo/Redo.
- **Luma:** original generated landscapes, filmstrip, looks and RGB histogram.
- **Pulse:** offline fictional service metrics, GPU chart and virtualized table.
- **Darwin ARM64 SDK:** prepared dependencies, pinned native capsule and an
  independent app generator. All six executables were built from this SDK.

The SDK and every app include stage0108's exposure repair. Features, stable APIs,
Cogent/WebGPU backend, experimental rendering/pacing defaults and release
responsiveness policy remain unchanged. The showcase data/artwork are original
offline fixtures and reset on launch; broader GPU APIs remain experimental.

## Verify and try

Verify all seven payloads with `shasum -a 256 -c SHA256SUMS` before extraction.
Use Archive Utility to extract app ZIPs. Apps are **ad-hoc signed and unnotarized
developer downloads**. Downloaded-quarantine launch has not been qualified.
No Developer ID identity, notarization or security-policy change is used.

SDK consumers need Go1.25.6, Python3 and Xcode Command Line Tools on Apple
Silicon. Keep the extracted vellum-v1.0.0 folder intact. From that SDK, run
`python3 scripts/new-app.py /absolute/new/app --module example.com/my-app`,
then build/run from the generated app. All replacements must resolve into the
SDK. It is already prepared: do not run prepare.sh or overlay another library.
The source archive associated with this macos-prefixed tag identifies the same
commit as the binaries; building that checkout requires its ordinary preparation.
The older root v1.0.0 automatic source archive retains its older source identity.

## Verification and support boundary

All required workflows passed on the exact build source64bf9cf:

- [Framework](https://github.com/dshills/vellum/actions/runs/40504082685)
- [Platform adapters](https://github.com/dshills/vellum/actions/runs/40504082565)
- [Stable root API](https://github.com/dshills/vellum/actions/runs/40504082914)

Fresh independent SDK consumer checks, race-enabled extracted-SDK exposure
regressions, ZIP extraction, exact notice/executable bindings, strict signatures
and final checksums passed. All six apps rendered and closed on Apple Silicon/
macOS26.6.2; edited Shell B2 closed to untouched A0 before any action in A.
The Notes check used only a disposable notebook; it did not access personal data.
The first background Notes capture needed native Raise. One Luma close action
needed a fresh native target; a separate single-close control passed. Those
observations remain disclosed; no sole product cause was established.

These native AX/capture checks are separate from physical-pointer qualification.
Prior unchanged input/history/soak evidence retains its original source/scope.
Human IME/VoiceOver, other OS/hardware/displays, cross-window canvas drag and
broader professional GPU workloads remain unqualified. Wayland is unsupported.
No new performance or full-platform qualification claim is made.

## Distribution

Vellum's code is MIT; original third-party/native/font/artwork notices remain
included. The baseline payload audit is owner-approved. This release's final
payload decision covers the disclosed prepared-Go repair and new exact archive
hashes separately; historical audit records and the old source tag are unchanged.
