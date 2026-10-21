# Offscreen Metal copy control

The private CPU upload probe places most time inside the CPU copy into staging,
not the explicit create/map calls. That can include deferred backing, page faults,
cache behavior and scheduling. The locked Metal HAL uses shared storage with
write-combined CPU caching for map-write buffers and four-byte pitch alignment.
No pool or cache-policy change follows from that timing alone.

Build this standalone Objective-C program with ARC, optimization and warnings as
errors. It opens no window, uses no framework or SDK code and changes no shipping
policy. Require a unified-memory Metal device. Compare shared/write-combined and
shared/default-cache buffers, preserving default hazard tracking in both modes.
One executable, three ordered pairs: WC/default, default/WC, WC/default. Each
process has 128 warmup and 120 measured iterations. Predeclare 20 seconds per
native process and 120 seconds globally including all configuration probes. One
attempt per case; stop and seal on the first invalid case, drift or deadline.

Each iteration creates a fresh 7,692,216-byte source buffer for a 1519×1266 RGBA8
sRGB texture, row pitch 6076. Record create, map/descriptor-validation, first-copy
and subsequent-copy CPU intervals separately. Copy a prebuilt spatial pattern A,
blit buffer→persistent private texture→persistent shared/default-cache readback,
wait for successful command completion and compare every byte. Only then copy a
different prebuilt pattern B into the same source buffer and repeat full readback.
Thus the second copy uses already touched backing, after its prior GPU use has
completed. No CPU overwrite while in flight, cross-iteration pool or asynchronous
ownership policy is introduced. Texture/readback/pattern allocation is outside
the iteration. Copy intervals include the Objective-C source-data pointer getter.
The two modes use the same patterns and dimensions; each fresh iteration starts
with A and ends with B, preventing a stale prior readback from satisfying A.

A comparator control flips the last expected byte, requires a mismatch, restores
it and requires a match. Each GPU transfer must independently complete and verify
all bytes. Record raw rows, including warmups; validate exact mode, storage,
sequence, phases, counts, finite ordered clocks, nonnegative disjoint intervals,
complete byte counts and terminal state. A program failure retains its partial
JSON and fails the case. The owned executor bounds output/process lifetime and
retires its process group. This is not proof of native allocation balance.

Use the existing exclusive lease, source/binary/config/log pins, before/after
configuration and persistent geometry checks, immutable retained SDK inventory
and owned execution. No foreground requirement applies: this is offscreen Metal,
not native UI presentation. No user document, settings, IME or accessibility state
is changed. Tests exercise positive populations and refusal cases before internal
and Prism Anthropic Claude Sonnet 5.5 review and source commit; execute only after
that checkpoint. Independently validate, review and commit results afterward.

Report each run's measured create/map/fresh/subsequent CPU distributions and
within-iteration subsequent-minus-first differences. Do not add quantiles or call
differences between separate run percentiles paired-sample overhead. Serial GPU
readback changes workload, synchronization, working set and cache state; results
cannot establish equivalence to WebGPU's asynchronous uploads, justify unsafe
buffer reuse, select a shipping cache policy alone, or close any UI latency gate.
Features remain frozen, experimental rendering/pacing off, the 33.3 ms
receipt-to-actual-presentation gate unchanged and M3–M7 open.
