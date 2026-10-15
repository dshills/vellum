# One-pass glyph fingerprint checkpoint

A CPU profile confined to 1,000 repeated ScenePacket captures from the actual
retained 400-visible-cell/477-pool grid attributes 13.36% of sampled CPU to
`glyphCoverageGeometry` and 7.76% to its exact match. This identifies source work;
it is not native frame timing. The [contract](../../../specs/rendering.md#one-pass-glyph-fingerprint)
defines the resulting change.

Owned patch 0068 combines grammar/coordinate/bounds validation and internal FNV
candidate hashing in one traversal. The hash origin becomes the first validated
point's integer pixel origin. Integer translations keep the key and fractional
phases remain represented, including signed negative relative words. Returned
minimum origin, guarded target/clip hull and full minimum-normalized path equality
are unchanged. Hashes are private candidate indices; collisions still require
full equality. No new cache, retained state, allocation policy, work exemption or
packet representation is introduced.

The original geometry-admission/origin and exact comparison oracles remain.
Only the obsolete internal minimum-origin hash assertion is replaced by an
independent callback-walker fingerprint oracle. Tests cover 4,096 fractional
phases, integer translations, leading Close/non-Move grammar, malformed/refused
input, individual word corruption and forced hash collisions.

Five interleaved actual-grid software pairs preserve exact original/GPU encodings,
accounting and pixels. They retain 10,937 allocations per capture; median local
elapsed times are about 2.407 ms versus 2.242 ms. These shared-computer timings
are diagnostic and do not qualify a native or whole-frame gain. Checkpoint
acceptance additionally requires the full 24-observation DPI/row/scroll matrix
per arm, meaningful fault controls, renderer/adapter race/vet, pinned regeneration,
native compilation and private SDK relocation. Exact source/command identities,
internal reviews and Prism using Anthropic Claude Sonnet 5.5 must be retained in
`records/qualification.json` and the indexed evidence archive before commit.

No native application or physical GPU test is launched while the human is online.
The historical 4 ms p95 / 6 ms p99 native targets remain failed. P4/P6/P11,
native/GPU memory, endurance, physical input/displays, human accessibility,
distribution rights and full Mac completion remain open.
