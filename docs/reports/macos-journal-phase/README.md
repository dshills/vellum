# Offscreen timing-journal phase analysis

The remaining interval is still mostly after the recorded final present GPU end.
On the actual p95 input of each retained run, that interval is **81–92%** of
receipt-to-shown latency. Neither observer presence nor a reported display
timestamp establishes the cause. This analysis adopts no production change.

## Same-input attribution

Each row below identifies that run's actual p95-ranked measured input. The last
two duration columns partition that individual input exactly. They are not
independently selected phase percentiles. Times are milliseconds.

| Run / observer | Accepted revision | Receipt to shown | Before final present GPU end | Final present GPU end to shown | Post-GPU fraction |
| --- | ---: | ---: | ---: | ---: | ---: |
| Earlier / off | 243 | 51.101 | 4.270 | 46.831 | 91.6% |
| Earlier / on | 199 | 33.947 | 4.450 | 29.497 | 86.9% |
| Reversed / on | 167 | 37.882 | 7.070 | 30.812 | 81.3% |
| Reversed / off | 147 | 37.923 | 6.853 | 31.070 | 81.9% |

Every input also has a disjoint five-part partition: receipt to acquire,
acquire to first render submission, submission to first render GPU start,
render GPU span and render GPU end to shown. All 992 partitions close within
the inherited one-microsecond tolerance. Render GPU spans on these four ranked
inputs are 0.253–0.321 ms. The pre-final-GPU intervals include CPU work, waiting,
submission and GPU work; they are not a pure CPU metric. Go request lateness and
clock-read brackets are retained separately, without inventing a native-clock
scheduled target or adding them to the receipt partition.

The new receipt-to-shown quantiles exactly reproduce the complete original
validators. Earlier off/on p95 stays 51.101/33.947 ms; reversed on/off stays
37.882/37.923 ms. All four exceed the unchanged 33.3 ms gate. Phase percentile
summaries are descriptive marginal distributions and must not be added together
or used to decompose differences between runs' p95 values. Separate processes
and pairs are never pooled or joined by absolute timestamp.

## A reported clock is not a delivered callback

For each event in each on run, the analyzer independently brackets the event in
reported timestamps and callback observations. Equality belongs to the last
preceding event; duplicate observation times use the last equal observation.
No fitted lattice or assumed refresh cadence is used. All 248 inputs per case,
including warmup, remain in the derived rows; summaries use the fixed 120 measured
inputs. Off runs have no display-clock joins or inferred clocks from another run.

The following counts show when the callback associated with the **preceding
reported timestamp** had not yet been observed at the named event:

| Event | Earlier on | Reversed on |
| --- | ---: | ---: |
| receipt | 73/120 | 118/120 |
| raster_end | 74/120 | 116/120 |
| upload_end | 65/120 | 119/120 |
| first_submit | 63/120 | 117/120 |
| present_request | 64/120 | 118/120 |
| present_gpu_end | 71/120 | 116/120 |

These are recorded callback observations, sampled in the native callback. They
must not be treated as application scheduling opportunities already delivered
at the reported timestamp. The preceding reported callback may still be pending
while an older callback has already arrived. Late callback records are retained.

The nearest reported timestamp to actual shown feedback also has its own
callback observation. That observation occurs after shown by p95 **10.912 ms**
in the earlier on run and **16.340 ms** in the reversed on run. This is a recorded
clock relation, not measured physical display latency or callback processing cost.
In the earlier on run, reported timestamps in (final-present GPU end, shown]
have histogram {0:13, 1:107}, while actual callback observations in that same
interval have histogram {1:62, 2:58}. Reversed on gives reported
{0:9, 1:110, 2:1} versus delivered {1:12, 2:108}. These count different event
series; neither is a missed-vsync count, guaranteed latch count or optical proof.

## Conclusion and next bounded control

The earlier one-refresh observer advantage did not reproduce when order was
reversed. Callback delivery phase also differs across the two on runs. The data
therefore does not justify adopting a permanent display link, making a pacing
change or calling the observer neutral. The post-GPU interval remains the main
lead, but it includes unisolated presentation/scheduling work; no compositor-only,
driver-only or WebGPU-specific cause follows.

A future presentation control should match request phase and record pending
presentation/drawable state before any scheduling-policy change. It must keep
reported timestamps separate from observed callback delivery, retain complete
accounting and compare actual same-input GPU readiness with shown feedback.
This checkpoint does not prepare or authorize another native run, or change a
release criterion. M3–M7 remain open. Cadence, resize, physical input, displayed
pixels, observer overhead and ordinary-build equivalence are not qualified here.

## Verification and reproduction

Seven focused unit methods cover late versus delivered callbacks, exact timestamp
and delivery equality, duplicate observations, missing coverage, nonfinite/order
rejection, same-input closure and overlapping present request, actual rank-ticket
identity, absence of off-mode clock inference, classifier-source drift and archive
integrity failures. All four raw journals are independently reclassified with
the original eight validator/script sources pinned to the original runs. Stored
validations match exactly, with 992 joined inputs: 512 warmup and 480 measured.

All four original archives verify by fixed SHA256. Selected members verify their
indexed bytes/hash; no archive is extracted and no executable is loaded. An
independent fresh analysis produces byte-identical summary and complete row files;
existing output is refused without mutation. No native execution, rebuild, screen
use, system-setting change or SDK/app modification occurs. Prior native and SDK
checks remain the retained evidence; they were not rerun for this pure analysis.

From the repository root, with a fresh output directory:

```sh
python3 -B -m unittest discover -s tools/macos-calibration/displayclock -p phase_test.py -v
python3 -B tools/macos-calibration/displayclock/phase_analysis.py \
  --reports docs/reports --output /tmp/vellum-journal-phase
```

The [summary](summary.json), [checks](checks.json),
[analysis contract](../../../tools/macos-calibration/displayclock/PHASE.md),
indexed analysis evidence and the original [off/on](../macos-display-scheduling/README.md)
and [on/off](../macos-display-reversed/README.md) checkpoints retain the basis.
The portable full verification recipe can also be run with a fresh directory:

```sh
python3 -B docs/reports/macos-journal-phase/execution-sources/verify_checkpoint.py.txt \
  --repo . --artifacts /tmp/vellum-phase-verification
```

Internal review and Prism with Anthropic Claude Sonnet 5.5 are complete: eight
files initially and three focused follow-up files, without exclusions/truncation.
The portability and rank-maintainability findings are fixed. The fixed-archive
resource concern is dispositioned with source, pin-rejection tests and the
[expanded-size inventory](archive-inventory.json). No actionable findings remain;
see the [review receipt](review-completion.json). Features, original backend, rendering/pacing defaults,
SDK/app and the 33.3 ms release gate stay unchanged.
