# Mac glyph-mask cache: owned bounded retention

Owned patch 0046 bounds the shared raster cache to 32 strong font-face keys and
8 MiB of mask pixel data. Whole face groups use FIFO admission order; hits keep
their original position. Eviction clears vacated backing-slice entries and removes
the group's map, then relooks up the requesting face before inserting. Returned
masks remain valid through caller references. Key quantization, placement,
cacheable dimensions and render code retain the pinned parent behavior. The
existing mutex covers lookup, eviction, accounting and shared rasterization.
Limits exclude parsed-font size, container/allocator overhead, other caches,
temporary allocations and caller-held masks.

The original-compatible regression fails the parent cache at 33 face keys and
8,392,704 mask bytes, preserving raw exit 1. New tests exercise exact 8 MiB
admission, current-face self eviction, FIFO warm hits, backing-prefix retirement,
caller-held pixel preservation, refused glyphs, 960 concurrent requests across
40 faces, and 64 embedded real-font cold/warm/regenerated pixel and placement
comparisons against separately instantiated original cache methods. The oracle
preserves the parent methods and license notice with type renaming and duplicate
declarations removed; its original hash/transformation metadata are pinned.

Fresh prepare verifies input/patch/output hashes and produces 6,178 files: exactly
one previous implementation changes and four test/oracle leaves are added. The
final owned full raster suite passes with race instrumentation three times,
offscreen vet passes, the adapter offscreen race suite passes, and the opted-in
native window-cycle consumer compiles. These final receipts bind all prepared
inputs and eight implementation/pin/spec/doc files. Internal review catches and
corrects the preparation test's old 1..45 order expectation to include 46; the
whole Python preparation suite passes against the final nine selected files.
The full raster suite generates 17 known
PNG outputs; their identities are recorded and they are moved outside the
prepared tree. Native compilation retains an existing duplicate -lobjc linker
warning. Compilation is not native execution or memory acceptance.

An initial overlay suite fails only its fixture's incorrect expected case count
(64 actual versus 72 expected); its sources and raw failure remain historical.
The later preliminary overlay suite passes, but its receipt did not bind the
external overlay source leaves, so it is not used to certify exact adopted code.
The final prepared-source runs supply that identity evidence. The preparation
receipt predates a README-only heading-spacing clarification. Final Go gates and
Prism bind the unchanged eight-file implementation/pin/doc scope; the later Python
gate and internal source certificate bind all nine files and the same prepared
outputs. The one-line preparation expectation update needs no second provider call.

Internal source review approves the exact nine-file owned patch/pins/specification,
preparation regression expectation and
five decoded output leaves. One Prism review uses Anthropic Claude Sonnet 5.5,
covers all eight files in two chunks, returns exit 0 with no findings, and uses
16,818 input / 532 output tokens. The indexed archive retains raw receipts/logs,
original/current cache/test leaves, stage/source identities and review results.
Binaries, generated PNG pixels, caches and desktop snapshots are excluded.
Selected-source/all-prepared evidence is not a full compiler/cache/SDK closure.

## Native memory acceptance remains open

The parent window-retention report preserves production 1,000-cycle failures at
375 cycles and a different, earlier cache overlay's diagnostic completion of
1,000 cycles. That intervention reduces Go heap while sampled RSS still grows;
a separate 800-cycle own-process vmmap diagnostic reports physical footprint
7.9 GB then 14.3 GB and IOSurface 7.2 GB then 13.5 GB at 250/750 cycles. Neither
diagnostic qualifies the final adopted source or resolves native graphics
retention. This phase launches no native app, changes no settings and performs
no latency acceptance. Further live/input/performance checks are deferred after
the quiet-host authorization expires. Native layer/input temporary-object
ownership remains under investigation; full P11/endurance, native/GPU memory,
latency, human interaction, cross-display and public distribution remain open.
