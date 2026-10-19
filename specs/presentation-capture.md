# Serial native presentation capture

The stored grid and tree consumers expose the public
[native presentation API](native-presentation.md) through the opt-in
`-perf-presentation NEWFILE` flag. This is a native feedback and routed-latency
preflight. Their existing scroll loops are serial, with pauses and GC between
parts of the workload. Their intervals cannot qualify the sustained
[60 Hz release policy](macos-responsiveness.md).

## Run and validate

Build the examples with the prepared native driver as described in their READMEs.
Use an optimized, non-race Darwin ARM64 binary with Go 1.25.6 for this initial
validator. The following commands are examples for a desktop-available session;
this checkpoint has not executed them against a native window.

```sh
/tmp/vellum-loadedgrid -perf -rows=10000 -perf-presentation=/private/tmp/new-grid-presentation.json
/tmp/vellum-loadedtree -perf -nodes=100000 -perf-presentation=/private/tmp/new-tree-presentation.json
python3 -B scripts/check-presentation.py /private/tmp/new-grid-presentation.json \
  --sha256 RAW_JOURNAL_SHA256 --source CLEAN_SOURCE_COMMIT \
  --binary-sha256 EXECUTED_BINARY_SHA256
```

Use a new file for each run; creation is exclusive with mode 0600. The flag
requires `-perf` and excludes smoke/accessibility and other profiler/source/input
ledger modes. Existing data size, warmup, clipped viewport, callback and CPU-frame
assertions remain in force. GPU selection remains optional; the requested flag
does not establish the actual GPU route. Every capture-enabled performance event
is labeled `presentation_capture:true` and `cadence_qualified:false`. The ordinary
path leaves the recorder disabled and its output schema unchanged.

## Frozen cohort and lifetime

After all expected CPU work completes, and before closing the native window,
`Seal` freezes the acquisition sequence prefix and receipt count. It first requires
exact CPU/receipt tag parity and one matching final armed acquisition per completed
CPU frame within that prefix. Missing identities fail capture immediately; later
acquisitions cannot repair the frozen cohort. It allows up
to two seconds for that fixed prefix's callbacks to drain. It retains the full
snapshot, including newly acquired tail records. It does not move the cutoff,
discard losses or treat retirement/Present calls as actual presentation.

`Finish` writes both pre-close and post-close snapshots after native Run returns
and the measurement worker acknowledges terminal cleanup. A setup failure before Run returns has `RuntimeReturned:false` and no post-close
claim, even if no worker started. An unjoined worker produces a failed journal with no post-close claim. Failed/canceled runs retain
evidence. The fixed prefix must remain immutable after close; untagged tail
records remain explicit, outside the serial input cohort. Pending, retired,
skipped, unavailable or invalid records in the prefix refuse validation, except
for the precisely defined pre-input startup observations below.
Post-close pending feedback, recorder overflow or a failed recorder also refuses.

### Explicit pre-input startup observations

Apple's `MTLDrawable.presentedTime` is zero when a drawable has not been presented
or was skipped. A terminal native callback with zero time does not prove display
and is retained as `State:"skipped"`. A skipped acquisition is a startup
non-timing observation only when it belongs to the same window scope, has
`Armed:false`, `Tag:0`, `Generation:0`, `Presented:0`, and its positive finite
`Acquired` time is strictly before that scope's first native input receipt.
The complete receipt cohort must have valid, increasing clocks in that scope.
No receipt, a foreign receipt, or a skipped acquisition at/after the first receipt
provides no exemption. Pending, unavailable, invalid and retired prefix records
remain failures, including before the first input. Every setup/warmup/measured
final armed acquisition still requires genuine actual presentation.

New captures always emit `PrefixObservations` with exact `Records`, `Presented`
and `StartupSkipped` counts for the original frozen prefix. The raw samples,
sequence, cutoff and terminal states remain unchanged. Startup skips never become
presentations or contribute a timestamp. The strict validator reports
`prefix_records`, `prefix_presentations` and `startup_skipped` separately, and
calculates raw presentation intervals only from genuine presented records.
Successful older v1/v2 captures without this additive object remain supported
when they contain no startup skips. A startup exemption requires the explicit
matching counts. An existing `FAILED_DIAGNOSTIC` journal remains refused and must
never be rewritten or relabeled to apply this prospective correction.

This classification corrects diagnostic population accounting; it proves neither
first-paint reliability nor a rendering fix. It changes no measured input, latency,
schedule, timing threshold or cadence denominator. It cannot cure actual workload
backlog or a display cadence failure.

The presentation recorder's shared receipt/drawable capacity is 65536. Each journal
has a 64 MiB write/read limit. The executable is hashed once before measured work,
from a regular file of at most 512 MiB; size/mtime drift refuses capture. These are
explicit diagnostic limits, not claims about total process memory. No files are
overwritten and no existing evidence is removed on failure.

## Validation and provenance

The strict validator checks the exact schema and count bounds, duplicate JSON
keys, finite clocks, one native window incarnation, contiguous acquisition
sequence, composition order and immutable composition identity, terminal CPU
accounting, callback state counts and receipt<=acquisition<=actual presentation.
Repeated acquisition within one composition is permitted only with the same
stamp. Drawable ID zero is valid. Actual presentation times are sorted for
intervals; equal times remain zero intervals and are reported explicitly.

Every setup/warmup/measured input must have exactly one matching final armed
presentation with its completed CPU generation. Intermediate compositions remain
in the raw prefix. Grid warmup tags start at 2^32+1, followed by measured tags 1..N.
Tree setup uses tags 1,2, warmup 3..W+2, then measured W+3..W+N+2. Measured latency
statistics use only the declared measured cohort; raw latency roles and prefix
intervals remain available. Missing or extra tags, tagged tail records and
changed terminal evidence refuse validation.

The validator requires the expected journal hash, source commit and executable
hash. Build metadata must identify optimized native Darwin ARM64/CGO work. A
present VCS stamp must match a clean expected source. Builds with an external
modfile can lack a VCS stamp: acceptance then explicitly requires external
build/launch binding instead of inventing one. In either case the execution
controller must independently bind clean source, module replacements, prepared
native sources, linked native library, executable, actual launch, workload,
display/power/scale and raw files. Self-recorded metadata and valid JSON alone
are not evidence that a native run happened.

Successful output is `VALID_SERIAL_NATIVE_FEEDBACK_JOURNAL`, with raw latencies,
p50/p95/p99/max and explicit state/tail accounting. It always reports
`cadence_qualified:false` and `physical_latency_qualified:false`. It does not
close M3 or certify physical input latency. A continuously changing producer and
five-run pacing campaign remain separate work, after real callback/clock/lifecycle
proof. [Checkpoint evidence](../docs/reports/macos-presentation-capture/README.md)
covers code and headless checks only.

## Separate paced mode

The [fixed-deadline workload](paced-presentation.md) uses `-perf-60hz`, at least 600
changing states and a version 2 journal. This file's serial version 1 format and
noncadence interpretation remain valid. Both share the fixed acquisition prefix,
loss/refusal policy and worker/runtime retirement boundary.

## Settings initialization and geometry isolation

Stored grid/tree performance consumers load Core settings, when not already
initialized, before applying the process-local geometry persistence override.
This prevents the first Body's lazy settings load from restoring a saved policy
after the override. Load failure or nonfinite/nonpositive zoom refuses native
setup; successful teardown restores the loaded policy. The grid accessibility
fixture shares this guard. The guard reads settings without saving them and
does not change OS preferences. It does not independently prove native geometry
or replace the capture's external execution and display evidence.
