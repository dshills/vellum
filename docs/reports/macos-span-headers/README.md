# Builder-local span headers

Owned patch 0066 reduces temporary allocations in the existing scene capture
path. It stores the active span operation inline and reuses exact power-of-two
construction row-header bands within one builder. Nine slots retain at most
511 headers (12,264 backing bytes on this 64-bit machine), with no span references
in idle bands. Larger and clamped non-power-of-two bands bypass reuse.

Reservation still precedes borrowing. Growth transfers references before clearing
the old header container. Flush first copies exact packet rows/span storage,
then clears and retains construction headers. Seal retires all scratch and inline
state on success or refusal. Logical byte/operation charges, ordering, winding,
refusal and original/GPU payloads remain unchanged; no global pool is introduced.

A CPU profile confined to repeated scene capture from an actual retained
400-cell software grid identified Builder.Span and allocation as substantial
source work. Five unprofiled interleaved original/candidate pairs preserve exact
logical accounting, original/GPU encoded parts and software pixels. Median allocated
bytes per capture fall from 5,524,730 to 3,967,368; allocations fall from 31,084
to 14,488. The fixture is procedural, fixed-row, DPI 96, Noto Sans 15 with two
allowed Go processors (`GOMAXPROCS=2`). Wall times are retained as diagnostics only on this shared
computer. These are allocated bytes/objects, not retained/native/GPU memory or
whole-frame latency.

The actual grid matrix also matches 24 observations per arm across DPI
96/120/160/240, fixed/variable rows and three vertical/horizontal scroll steps,
including every encoded part, accounting and software pixels.

Qualification requires the independent original-builder state/encoding/pixel
oracles, focused bounds/alias/clear/retirement tests, meaningful fault controls,
pin-verified regeneration, renderer/driver race checks, vet, native compilation
and private SDK relocation. Internal source/evidence review and one significant
Prism review using Anthropic Claude Sonnet 5.5 precede commit. Checkpoint acceptance requires exact receipts and
source hashes in the qualification record and evidence archive.

No native app, physical pointer/GPU test or performance campaign is launched.
The existing 4 ms p95 / 6 ms p99 native targets remain failed at their historical
endpoints. A matched quiet-machine native comparison is still required; this
checkpoint does not qualify a frame-rate gain, complete P4/P6/P11, native memory,
human accessibility or full Mac/distribution acceptance.
