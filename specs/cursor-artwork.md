# Built-in Mac cursor artwork

The prepared native driver replaces all 39 built-in cursor SVGs with original
Vellum geometric artwork, provided under MIT. The existing 40 enum values
(`None` plus 39 image identities), public names and generated enum code are
preserved. `None` still means no cursor preference and has no SVG. Missing SVGs
still return errors; application enum types without a hotspot retain the center
fallback. This adds no new public API or cursor capability restriction.

The source, deterministic generator, license and manifest are in
`drivers/cogent/native/assets/cursors`. Generation requires a new output
directory. The manifest binds every SVG SHA-256, its 256-unit extent and hotspot.
The replacement artwork uses the existing exact quoted `#fff`, `#000`, `#f00`,
`#0f0` and `#ff0` theme tokens. Go implementation/BSD notices and the dependency's
original module version/checksum remain intact. The artwork's MIT declaration is
supplemental; it does not relicense the module or other embedded assets.

## Preparation contract

Upstream stage0089 replaces all SVGs through digest-pinned asset overlays. Its
text patch changes hotspot coordinates and the cursor README and adds metadata,
MIT/NOTICE files and two regression files. It contains no previous SVG bodies.
An adjacent asset-only stage replaces the duplicate zoom-in SVG rendering
fixture through the same source/pin path. The SDK must not contain old cursor
artwork, patch recovery copies or rejected
hunks. Baseline artifacts remain historical evidence outside the new SDK.

Before patch execution or asset mutation, preparation admits all patch snapshots
and all overlay sources/targets, verifies exact old/new pins, rejects unsafe
paths, symlinks, nonregular files, aliases, duplicates and patch/overlay overlap,
and opens every asset writer without truncation. It caps stages at 256 assets,
64 KiB per asset and 4 MiB aggregate source bytes; patches are capped at 8 MiB
per member and 32 MiB per stage. Exact-size reads include a one-byte growth probe.
Asset stages accept canonical unified `a/` and `b/` paths, optional tab-separated
timestamps and counted hunks; other patch formats refuse before execution.
Legacy stages without assets retain their existing patch format acceptance.
Admitted bytes are used even if a source path changes afterward. Targets are
rechecked before the first truncation. Descriptors close on completion/failure.

Asset writes are not atomic as a group. Preparation's existing owned staging
and installation/rollback boundary protects previously installed dependencies;
a failed stage is never installed. Recovery needed after an installation rollback
failure is preserved under that existing contract. Successful asset overlays
create no backup copies in the prepared payload.

## Verification and limitations

Regression coverage must render all 39 images in both light/dark schemes at
16, 24, 32, 48 and 64 pixels (390 cases). Hotspot checks use unshadowed raster
geometry, integer scaling, bounds and nonempty coverage, with semantic badge
colors and repeated enum/size cache identity checks. Theme/cache state is
restored between tests. Inspect actual rendered contact sheets separately.
These checks do not quantify display contrast or physical pointer usability.

An asset/hotspot change alters executable inputs: rebuild affected examples and
the private SDK native workbench. Preserve exact metadata and notice identities,
verify the complete SDK for old artwork/recovery-file absence, and keep source,
SDK and build provenance distinct. Earlier unchanged-executable evidence from
notice-only checkpoints does not qualify these replacement inputs.

CPU raster and headless builds are separate from physical pointer verification,
Cocoa display scaling and native cursor lifetime qualification. The existing
unbounded enum/size renderer cache and native cursor replacement cleanup remain
separate resource/lifecycle work. This artwork checkpoint does not close those
checks, the broader asset/rights audit, measured 60 Hz responsiveness or full
Mac release acceptance.
