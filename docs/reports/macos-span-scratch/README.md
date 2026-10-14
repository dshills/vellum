# macOS bounded span scratch

Scene packet capture previously allocated new construction arrays for every
span row of every glyph. Patch 0062 keeps a small builder-local pool after copying
each operation into detached owned storage. Normal CPU image rendering does not
use this builder; this targets the optional captured GPU scene path.

The pool is bounded to 32 row buffers and 64 KiB backing, with at most 256 spans
per retained buffer. Fixed slot metadata is additional transient storage. Pool
limits choose normal allocation; original packet admission, row-band growth and
refunds remain unchanged. Seal retires all scratch on success and failure.

Five interleaved actual software grid capture pairs, fixed 10k procedural rows,
400 visible cells, DPI96/Noto Sans15/GOMAXPROCS2, match pixels, geometry, command
count and encoding accounting. Median capture is 3.813318 to 2.988978 ms,
9,273,852 to 5,524,729 B/op, and 88,944 to 31,084 allocations/op. These are local
unprofiled observations with the native endurance run concurrent. They do not
qualify native frame latency, physical/GPU memory, presentation or P6 targets.

The diagnostic current-source CPU/capture profiles are separate fixed warm loops.
Their runtime and background scavenger samples are retained. Initial CPU
benchmark calibration dominated the first profile; the first capture fixture
failed to compile before execution. Corrected loops pass pixel and serialization
checks. No latency qualification follows from these profiles.

A pinned original-builder oracle checks 80 repeated-stream trials after every
span and flush, plus both encodings and seeded replay. A held-operation test
proves actual construction-pointer reuse without changing earlier output.
Internal review corrected a portable test expectation to derive the backing
limit from span size. Normal preparation reproduces three pinned outputs (6,214 to 6,216 leaves);
all other dependency leaves are unchanged. All 48 current-driver matrix records
match across four DPI settings and fixed/variable rows. The affected renderer
and complete offscreen adapter race/vet suites pass. All four explicit native
GPU readback roots pass without skips. Three isolated injected faults are caught
by assertions for stale row length, held output mutation and pre-admission
borrowing. The native loaded-grid candidate is built and hash-bound, unlaunched.

The initial preparation unittest retained the old 61-patch expectation and
failed; updating its expected contiguous range to 62 passes all 11 checks.
Existing upstream raster tests generate 17 PNGs beneath prepared testdata.
Their receipts retain those artifacts; regeneration restores the exact 6,216
source leaves before commit. Internal final review and one bounded Anthropic
Claude Sonnet 5.5 Prism review precede commit.

The separate 1,000-window campaign remains on e7f8e30 and cannot qualify this
new patch. Native paired capture performance is still required after that
campaign finishes; unchanged 4 ms p95 / 6 ms p99 targets remain open.
