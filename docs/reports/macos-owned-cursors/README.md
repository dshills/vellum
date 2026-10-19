# Original built-in cursor artwork

Vellum's prepared Mac driver now uses 39 original geometric SVG drawings under
MIT. The previous cursor README identified an external macOS cursor source and
Apple agreement, without establishing redistribution provenance for this SDK.
The replacement artwork was authored independently before baseline bodies were
read for absence checks; none of those old SVG bodies appear in the source patch.
The original Cogent Core Go/BSD notices and module version/checksum remain.
This closes the scoped cursor-artwork provenance gap, not the broader rights audit.

All 40 existing enum identities (None plus 39 images), names, custom enum fallback
and theme color substitution remain. Hotspots now match the new drawings on a
256-unit square. The reproducible source generator, manifest and MIT declaration
ship under `drivers/cogent/native/assets/cursors`. Digest-pinned, bounded overlay
preparation admits sources/targets/patches before mutation and retains the existing
owned staging/install/rollback boundary. Asset stages accept canonical unified
patches only; the legacy patch stages retain their prior format acceptance.
An internal review found and corrected an alternate-prefix overlap-preflight gap.

The full initial SDK scan found the old zoom-in SVG body duplicated in an SVG
rendering fixture. An adjacent asset-only stage replaces that fixture too, and a
Go regression requires exact agreement with the new embedded zoom-in. The final
whole-payload scan finds none of the 39 old built-in bodies or that duplicate
fixture body. It rejects new recovery files while preserving the 79 byte-identical
Cargo.toml.orig source metadata files in the existing native notice capsule.

Validation passes:

- 54 affected preparation/SDK unit methods, including 14 overlay methods for
  checksum failures, admission limits, path/alias/symlink controls, complete patch
  preflight, source snapshots, target drift, partial-write failure and cleanup.
- Fresh pinned preparation of both dependencies: 5,998 upstream files and 297
  typesetting files. Exactly five upstream metadata/test/notice files are added,
  39 built-in SVGs plus the duplicate fixture and two text files change, and all
  typesetting bytes remain unchanged.
- Three Go regression tests, including 390 raster subcases (39 images, five sizes,
  light/dark). Hotspots land on unshadowed primary geometry, with bounds, ink,
  semantic color and cache identity checks. Actual contact sheets were visually
  inspected; no numeric display contrast or physical-pointer result is claimed.
- Generator reproduction of all 39 SVGs and the manifest, plus refusal of an
  existing destination without changing those outputs.
- An 8,119-file private SDK with complete archive/extraction integrity, old-artwork
  absence and inventory checks. Compared with the prior clean capsule, 56 payloads
  are added and 47 change; only the core module gains two supplemental cursor
  notice records. Original module metadata and the native archive remain intact.
- Relocated consumer race/vet/build and public API scenario race/vet checks;
  fresh builds of all 26 default SDK programs and one tagged helper. All selected
  local replacements resolve inside the extracted SDK. Headless modules correctly
  omit the native driver. Inner Go commands are supervised within their one
  outer owned process group; they are not individual controller receipts.
- A fresh optimized ARM64 native workbench build with an actual linker map and
  manifest-bound receipt: 141 selected native archive members and all 39 selected
  embedded cursor hashes match. Each cursor lists its supplemental MIT/NOTICE
  candidates. The unsigned app carries the same binary and a verified notice
  payload with 176 component records and 398 files. The SDK stays unchanged.

Several external validation attempts failed on sandbox/log-scope assumptions,
intermediate authoring/pin checks, the already-updated candidate baseline, a
headless-module graph assumption, and printing an integer notice count as a list.
Available raw endpoints and source versions remain retained. One result-name
collision overwrote the first failed app controller's aggregate receipt with the
successful app artifact summary. That original aggregate receipt is lost; its raw
command logs remain. The original generator check's complete lifecycle record
was part of that lost receipt, so generation is repeated under distinct controller
and artifact result names with a complete fresh owned receipt. The app-v2
controller and actual app artifact checks remain intact. Corrected checks pass;
the duplicate-artwork discovery/source correction and this evidence loss are
explicitly recorded. No failed command is relabeled as a pass. The SDK source is a disclosed dirty
precommit snapshot of e17d359; clean release provenance is recorded separately.
The [evidence record](evidence.json) binds actual sources, logs, artifacts and
limitations. Internal and requested Anthropic Claude Sonnet 5.5 Prism checkpoint
reviews are retained separately with their dispositions.

No app was launched or installed, no OS preferences or cursor were changed, and
no signing, notarization or public release occurred. Physical pointer/display
scaling, native cursor ownership/cache cleanup and broader resource/endurance,
platform/accessibility, font/native/other-asset applicability and final SDK/app
acceptance remain open. The renderer cache is still unbounded; this asset change
makes no lifecycle or cache-limit claim. Measured 60 Hz presentation/input remains
the release gate, with 4 ms a reported optimization target. Full Mac completion
remains open under the [delivery plan](../../../specs/macos-delivery.md).
