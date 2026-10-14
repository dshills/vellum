# Packed owned span rows

Stage 60 copies all rows in a span operation into one detached backing buffer.
Each row's capacity is capped at its length. Admission/refund order, retained
accounting, span order, original/GPU encoded bytes and CPU reconstruction stay
unchanged. Construction slices remain separate and retire on sealing.

Five interleaved unprofiled local 400-cell driver captures measured median
3.832 ms before and 3.738 ms after. Allocation count fell from 106,337 to 88,944
(17,393 fewer); median allocated bytes were 9,283,434 and 9,273,835. These are
procedural fixed-row DPI96/NotoSans15 software captures with two benchmark threads
while the separate native endurance process ran. They do not qualify native
whole-frame timing or physical/retained memory.

The sceneplan race suite and a 600-stream independent original-row oracle pass.
Borrowed construction overwrite, append isolation, exact capacity, serialized
original/GPU bytes and seeded partial-alpha pixels are checked. Two deliberately
faulty overlays fail missing-copy and row-capacity assertions. Internal candidate
source review found no issues. Normal preparation reproduces two pinned outputs
(6,210 to 6,211 leaves); every other dependency leaf is unchanged. All 48 current
driver records match the stage-59 reference across four DPI settings and
fixed/variable rows, including pixels, payload shape and accounting. Sceneplan
race/vet, full adapter race/vet, preparation tests and native build pass. The
preparation unit tests use mocked installations and stable stage-59 dependency
snapshots, while testing the new manifest/pins; regeneration and runtime code
gates separately bind the exact stage-60 dependencies. The native binary is
built and hash-bound. Internal final review and one bounded Anthropic Claude
Sonnet 5.5 Prism review passed before the code commit. The subsequent native
follow-up completed 40 interleaved opt-in GPU processes and 4,800 samples across
four stored-table cases, comparing `9686e38`/6,210 prepared leaves against
`68c3750`/6,211. Geometry remained exact; four gated native GPU roots passed without
skips on the candidate. Native differences were mixed, with observed tail regressions; both arms
missed the unchanged warm targets. See the [source-bound follow-up report](../macos-native-followups/README.md).
The original 4 ms p95 / 6 ms p99 native targets remain unqualified by this change;
full P6/P11, human input/accessibility and Mac distribution remain open.

The follow-up report names the observed p99 increases explicitly. Their cause
is unassigned; allocation or local capture gains do not dismiss those tails.
