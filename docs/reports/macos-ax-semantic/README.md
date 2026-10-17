# Builtin accessibility semantic encoding

The owner accessibility path now writes exact builtin scalar snapshots directly
into an owned immutable string using the private fixed wire schema. A whole-call
preflight sends every unsupported/named/custom/mutable value or nonfinite scalar
back through the original standard encoder. This preserves custom callback order,
side effects and errors, including a callback that repairs a later invalid value.
Sticky immutable-value eligibility, pending/ready cache authority, generations,
action admission and native publication remain unchanged. No asynchronous work is
introduced and no public schema or API changes.

Field order, interface omitempty, nil/empty arrays, zero pointers, negative zero,
float32 thresholds and HTML/control/Unicode escaping match the standard encoder.
One fixed builtin constant determines the toolchain's invalid UTF-8 replacement
spelling; valid U+FFFD remains literal. Schema/method audit tests require future
node changes to be reconciled with the fixed writer. The allocation hint borrows
only a prior owned string's length with matching baseline count under the bridge
mutex. It carries no identity/epoch authority, releases the lock before encoding,
and clamps arithmetic/reservation to 1 MiB. Output growth remains unrestricted;
no scratch pool or new retained source/backing cache is introduced.

## Native decision

Five interleaved release pairs use fixed 10k rows, 400 nonempty clipped cells,
128 warm and 120 measured events per process. All paired geometry checks pass:
50 rows/eight columns, 477 retained cells/nine headers, 162 logical DPI (1.6875),
1519x1266 client on Apple M4 Pro/LG HDR DQHD 60 Hz, macOS 26.6.2 (25G83).
Go 1.27.2 ARM64 release builds use trimpath/buildvcs=false without race detector,
profiler or overlay. Ordinary runtime GOMAXPROCS/GOMEMLIMIT/GOFLAGS are unset.

| Run | Baseline p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| 1 | 4.875082 | 4.771915 |
| 2 | 4.913624 | 4.788377 |
| 3 | 4.868877 | 4.828208 |
| 4 | 5.018164 | 4.675586 |
| 5 | 4.774999 | 4.946290 |

Four of five paired p95s improve. Median run-p95 is **4.875→4.788 ms** (1.8% lower);
median run-median UI work is 3.793→3.649 ms and reconcile is 1.382→1.175 ms.
These separately reported percentiles/phases are not added. Median per-run median
process-wide Go allocation observations are **5,894,448→5,300,168 B** (10.1% lower),
including fixture/runtime work. One pair regresses, and process allocations vary
between runs; this is a scoped observation, not a universal guarantee. All p99s
are below 6 ms, but every p95 exceeds 4 ms. **No Mac completion gate closes.**
Adopt the measured reduction while retaining the unchanged acceptance matrix.

Separate source-details cases record 120 captures per build, all GPU accepted,
zero fallback/reuse. Median capture is 1.025→1.017 ms and encoding 0.402→0.391 ms;
logical input remains 559,988 B. These diagnostic records are not pooled into
ordinary timing and do not establish display presentation or physical input.

## Checks and identity

The complete driver offscreen race suite passes (286.931 s command). Internal
review catches a shared-backing test oracle; it is corrected to independent
json.Marshal bytes and the focused race tests pass (11.033 s). Corrected legacy
nojsonv2 tests pass (7.232 s), as do the current default implementation, vet and
release build. Oracles cover all 30 ordered fields, malformed/escaped strings,
2,500 randomized all-field snapshots, 10,000 random finite float bit patterns,
custom whole-call effects/errors, retained output and capacity-only hints.
The initial invalid-UTF8 mismatch and its fixed-constant correction are archived.
The full suite precedes the test-only oracle correction; production source does
not change after the release build. No minimum-toolchain SDK qualification is
inferred from the two encoder implementations on this installed compiler.

Baseline runtime is the previously verified 6e366fc text-batch binary, SHA
38a98dcaddb7855a2476943aa468ae6f8294f11f665e0b88946ea65e677fb9da.
02f8dbe changes documentation only. This is not a fresh historical baseline
rebuild. Its measured receipt has 6,259 prepared files; the prior final checkpoint
adds one documented test-only file. Current candidate stage0079 stays at 6,260
files, aggregate SHA
1d9861ac10a0d83f607a36fd70dc79f00ce51d427c792b30b6a1331c0c3451ba,
unchanged before/after native timing. Candidate binary SHA is
f66659d09f1547d991f19cad5579e6084751ae678ac59899f93c6453f396de1b.
Both compiled production leaves are pinned, including the new untracked writer.

The runner observes direct-child exit/EOF and attempts owned-group cleanup;
a later endpoint query finds no matching owned executable. Strict per-PID/PGID
absence is not claimed. No input source, clipboard, VoiceOver or OS preference
changes. Raw samples, commands, failures, source and reviews are indexed in
records.json/records.tar.gz without copied SDKs/binaries.

Internal review approves the corrected source and raw native-result accounting.
One fresh complete Prism Anthropic Claude Sonnet 5.5 call reviews all six staged
source/spec/report files, 31,989 B in one chunk, no truncation/cache. CLI exit 0
has **zero retained findings, one discarded medium and zero suppressed**. The
raw concern guesses the hint may self-deadlock; patch evidence has a file-path
mismatch. Independent source review also refutes the concern: the first call
precedes locking, the completed-cache branch unlocks before continuing, preparation
defers unlock before its return, and final publication locks after the second
hint. Production entry wrappers hold no bridge mutex. No source edit, extra test
or repeated provider call follows. Both current prepared trees match the unchanged
6,260-file stage0079 aggregate.

About 0.79 ms remains above the median run-p95 budget. Further work must target
material synchronous publication and full-capture costs with complete-frame
measurements. M2 native/M3/M4/M5/M6/M7, including the original full workload,
resource/endurance, human/display, rights/final SDK and audit gates, remain open.
