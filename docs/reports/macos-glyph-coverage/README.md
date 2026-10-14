# Exact small glyph coverage reuse

Owned Core patch 0056 reuses ordered 16-bit scanner coverage for small static
TrueType fill-only glyphs after the existing Alpha-mask cache declines them.
Exact normalized fixed-path equality includes the original float32 rounding.
The hash only indexes candidates. Existing image and scene spanners still blend
pixels and emit packet operations. Original outline, segment, coordinate and
extent charges run before lookup; a warm cache cannot revive a refused batch.

The per-renderer FIFO retains at most 512 entries and 1 MiB of exact path/span
array backing. Entries contain no font, image, builder or packet references.
Metadata and transient recording are outside this bound. Resize and Release drop
storage. Identity transforms, static bounded outlines, value uniform fills,
clean owned scanners and a guarded hull wholly inside target/clip are required.
Other cases preserve ordinary rendering. Existing owner serialization applies.

## Evidence and limits

The current-source diagnostic captures an actual procedural 10,000-row grid with
400 clipped cells, 2,515 paint commands and 409 text commands. CPU replay matches
all composition pixels. Original profiles identify outline/scanning as substantial
paint work. Candidate replay preserves the same full-image SHA256, 4,392 packet
operations, 2,811,600 packet bytes and 1,685,984 encoded bytes. Diagnostic profiling
ran while the human used the computer and is not a paired benchmark or evidence
of native frame-rate improvement. It does not qualify stored-data workloads.

A separate 48-frame actual-driver software matrix compares all composition pixels,
geometry and cell density across DPI 96/120/144/192, fixed/variable rows and six
mixed-axis scroll steps. Every record matches. Focused tests check cold/warm and
recolored pixels, fractional/negative/large positions, clipped/unsupported
fallback, exact packet encoding, held packet immutability, entry/byte eviction,
detached backing, hash collision, work refusal and scanner ownership.
Four intentionally broken overlays fail their exact assertions. A first aggregate
refusal test incorrectly required zero cache hits before a late budget refusal;
its failure is preserved, and the corrected test compares both routes' refusal,
CPU/scanner fallback and held packet state.

Formal preparation checks all 56 patch numbers, reproduces six reviewed files
exactly and preserves every other baseline dependency leaf (6,204 → 6,207 files).
Final renderer race/vet, adapter race/vet, portable race, preparation tests and
native harness tests/build must pass. The native binary is built but not launched.
The final source/spec/report requires an internal review and a bounded Prism
review using Anthropic Claude Sonnet 5.5 before commit. Raw producers, overlays,
failures, profiles, comparisons, receipts and reviews will be retained in the indexed
records archive before commit; binaries are excluded with their hash retained.

The first matrix output accidentally shared a filename with its run receipt;
the receipt replaced pixel records. The corrected V2 matrix uses separate
exclusive filenames. The first adoption script assumed stage count equals patch
count and stopped before manifest mutation; its corrected version checks exact
patch numbers. A native test command used a nonexistent package and was replaced
with the existing harness package. Unkeyed vector test literals were corrected
after vet. Sandbox Unix-socket bind failures are preserved beside authorized
local reruns. None of these original failures qualifies a passing checkpoint.

GPU scene rendering remains opt-in. Default/native 4 ms p95 / 6 ms p99 targets,
full P6/P11, physical displays/input, human accessibility, native/GPU memory and
Mac distribution remain open. The prior interacting lifecycle run failed after
343 completed windows; its initiating close event remains unknown. Live tests
remain paused while the human uses the machine.
