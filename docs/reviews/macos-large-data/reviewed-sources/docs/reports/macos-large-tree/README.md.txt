# Large stored-tree foundation

This checkpoint implements worker-built `TreeData`/`TreeIndex`/`Outline`, the
managed `ui.DataTree` adapter and an independent stored-tree consumer with
separate 10k/100k domains. It extends the experimental Mac framework.

Portable tree race20, the full virtual race suite and vet pass. The managed
adapter passes repeated actual offscreen pointer/key routes, native lifecycle
and payload checks, the full native offscreen suite and both native vet modes.
The consumer passes portable race3, actual offscreen routes at owner batches
1/7/64, both vet modes and a desktop race build. Independent source review found
and corrected native payload-budget accounting, loader cancellation and failed
preparation task identity; regression evidence preserves the original failures.

One actual Cocoa race process with 100k stored nodes exits zero after all nine
stages: empty readiness, stored load, deep-chain reveal, retained hidden selection,
filtered ancestor reveal, native filter settlement, canceled worker dual-terminal
retirement, independent survivor and final pass. Its measured tree viewport is
1788×377 logical pixels at scale 1.6875. The workflow uses owner commands and
completed native geometry; offscreen primitive routing is separately tested.
No OS input, accessibility or privacy settings were changed.

The source-bound records, raw logs, immutable review inputs and producer scripts
are indexed by `archive.json`. Historical consumer v1/v2 passing test records are
preserved as superseded: v1 missed loader cancellation and v2 missed a failed
Prepare task identity edge. Consumer v3 fixes both. Native16/UI source is unchanged
from its v1 gate. A clean stage34 preparation reproduces all 6,128 prepared files
byte for byte; that proof is shared with the timing report.

This establishes the bounded implementation and local checks. Human
IME/VoiceOver, external AX, physical pointer workflows, 500-visible-row latency,
process/native memory retention and full P6 acceptance remain open. Source maps
cover enumerated project/prepared inputs and compiler identity, rather than every
module-cache or system SDK input. Binary identities are recorded; binaries are
excluded from the Git evidence archive.
