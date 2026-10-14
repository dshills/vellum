# Mac native archive/linker observations

The distribution collector now optionally records a supplied private Mach-O and
ld-map pair. It matches absolute archive references to verified module-owned
candidates, then reports exact selected member ordinals, sizes and payload hashes.
The new bounded reader supports BSD extended/short ar names without extracting
or executing members. Relative references remain explicitly unresolved; the
collector never substitutes its own cwd for an unrecorded linker cwd.

Artifact pairing remains caller-supplied-unverified in the public tooling. This
checkpoint's fresh owned build separately records absent output paths before
dispatch and hashes both emitted artifacts immediately after compiler completion.
Neither mechanism certifies upstream source provenance or redistribution rights.
Object-table inclusion also does not imply every payload byte survives dead
stripping. Runtime execution and public distribution gates remain unchanged.

## Actual local evidence

The private workbench was built with external linking and an explicit ld-map,
using current committed Go/native implementation e319cca and local Python tooling
changes. Binary SHA256 is `934f3bfda1a4e02483e90629baa69c1e1be905ff6ba0dd63eb1284d8405c6de3`;
map SHA256 is `2d6a2be6f59b645a6bb5ae5d9930a83f49b9458fb1a35cc8a6a567d02100a931`.
The binary was built only. The arm64 map has 293 object-table entries, including
entry zero's linker-synthesized object. It selects 173 members from the
571-member Darwin arm64 libwgpu_native.a. Archive SHA256 remains
`7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248`.
The Go module graph is checksum-verified; observations contain no unresolved
archive reference for this specific pair.

The implementation-stage collector capture records 7,041 SDK files, 233 selected
nonstandard packages, 4,438 embedded build inputs including 12 fonts, and 89
module notice records. These are build input identities; they do not establish
that every listed embed asset is retained in the final binary. The new script,
tests and final spec may change the final SDK file inventory. All 6,207 prepared
dependency leaves remain exact. Fonts/native component rights still require their
separate review. Actual binary dynamic imports are system paths; this is not a
complete native source/component attribution or runtime validation.

The pinned binding's update script names upstream wgpu-native v29.0.1.1, while
its README names an older version. The
[upstream release](https://github.com/gfx-rs/wgpu-native/releases/tag/v29.0.1.1)
and [asset digest page](https://github.com/gfx-rs/wgpu-native/releases/expanded_assets/v29.0.1.1)
provide a candidate release identity. Direct metadata and asset downloads both
time out before output creation, so this checkpoint proves no upstream archive
byte match. Rust-looking object names guide further investigation; they do not
identify complete component versions, notices or source-build provenance.

## Review and checks

Parser and collector fixtures cover exact independent payload hashes, selected
subsets, malformed/unsupported ar tables, byte/name/member/object bounds,
nonregular/FIFO/symlink admission, changed streams, missing/ambiguous selections,
relative/unknown archive references, architecture/candidate drift and pending
scope. Existing collector drift/exclusive-output checks and SDK/checker/app-package
regressions must pass. Actual source-bound repeat collection and private unsigned
packaging are separately recorded; no executable runs.

Internal review found and fixed GNU-table suffix normalization, blocking FIFO
admission and implicit relative archive cwd substitution. The initial relative
path regression had a test-mocking error; the corrected test uses a real changed
collector cwd. A first real collection used an incomplete temporary module cache
and failed before producing an inventory; the existing complete cache is used
without weakening module checksum verification. The SDK relocation fixture also omitted Canvas after the portable ColorPicker
added a UI-to-Canvas import; its synthetic package set now includes Canvas and
the relocated consumer explicitly exercises ColorPicker and menu metadata.
Original failures and source versions are retained. The significant checkpoint requires internal source/evidence
review and Prism using Anthropic Claude Sonnet 5.5 before commit.

Indexed evidence will preserve code/producers, all failures/checks, selected
identities, raw maps, exact inventories and reviews. Native binaries/static
libraries are excluded, with their identities retained. The full Mac framework,
P6/P11 performance/lifecycle, physical/human input/accessibility, signing/notary
and native/font redistribution acceptance remain open. Live tests stay paused
while the human uses the computer.
