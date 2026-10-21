# Matched native/framework calibration

## Completed result — five valid pairs, release gate unchanged

The successful campaign ran 15:14:59–15:20:27 America/Chicago on 21 October 2026
(327.755 seconds, within the 480-second bound). The actual M4 Pro/macOS
26.6.2/25G83, LG 5120×1440/60 Hz, AC, scale-1 configuration is unchanged. Both
implementations match the locked 1519×1266 surface, payload rectangle
(8,8)–(1511,1258), pixel format 81, two drawables, vsync enabled, transaction
presentation disabled and next-drawable timeout disabled. Actual active/key/
visible/own-foreground-PID states hold at capture boundaries; the monitored
pre-cohort-to-end loss count stays zero, with per-request foreground checks.

All five predeclared alternating-order pairs pass strict accounting. The ten
timed processes retain 1,280 positive ordered warmups and 1,200 positive ordered
measured responses: 2,480 routed cohort states in total. Ten native setup zeroes
(two per native run) remain explicit before readiness; framework setup has no
losses. No failed case is substituted, omitted or reclassified.

| Pair / order | Native p50 / p95 / p99 / max ms | Vellum p50 / p95 / p99 / max ms | Difference of run p95 values ms |
| --- | --- | --- | --- |
| 1 N/F | 25.519 / 32.303 / 32.396 / 32.465 | 43.280 / 51.118 / 52.566 / 52.744 | 18.815 |
| 2 F/N | 25.654 / 32.416 / 32.472 / 32.478 | 45.021 / 52.063 / 52.406 / 52.710 | 19.647 |
| 3 N/F | 25.648 / 32.387 / 32.531 / 32.582 | 29.472 / 37.420 / 38.724 / 39.990 | 5.034 |
| 4 F/N | 25.499 / 32.361 / 32.439 / 32.451 | 45.132 / 52.250 / 53.106 / 53.213 | 19.889 |
| 5 N/F | 24.958 / 31.951 / 41.794 / 48.765 | 29.361 / 38.252 / 41.712 / 42.014 | 6.301 |

These are per-run statistics; samples are not pooled. The last column subtracts
two p95 values, not individual latencies, and is not a percentile of differences
or a measurement of pure framework overhead. Pair 5's native tail remains fully
included. This deliberately sparse 120 ms request fixture does not establish
60 Hz changing-state cadence, resize response or professional-workflow acceptance.

The native reference meets the existing 33.3 ms p95 number in each run of this
small fixture; Vellum exceeds it in every run. The earlier 49–65 ms busy/native
references therefore do not justify an irreducible OS-floor claim or raising the
gate. **Keep the release gate unchanged.** This result does not demonstrate that
representative framework workloads can already meet it.

### Boundary to investigate

Vellum receipt-to-acquisition p95 is 3.175–7.287 ms, versus native 0.731–1.157 ms.
Vellum acquisition-to-presentation p95 is approximately 49 ms in runs 1/2/4 and
33 ms in runs 3/5, versus native 31.544–32.133 ms. These quantiles must not be
added as if they were a decomposition of p95. The two Vellum bands differ by
roughly one refresh period. Source-bound journals have exactly 252 acquisitions
per framework run; none of the 600 measured responses has an intervening unarmed
acquisition between receipt and its attributed armed acquisition.

[phase-analysis.json](phase-analysis.json) records that accounting and measured
CPU UI-work statistics. The slower-latency runs have CPU-work p95 around
3.4–4.1 ms, while the lower-latency runs have about 6.7–7.2 ms. CPU duration alone
therefore does not explain the observed band ordering. No GPU completion,
submission or present-request host timestamps were recorded, so the next focused
boundary is actual submit/present scheduling after acquisition, before proposing
another backend repair. A fixed one-frame queue penalty or cause is not proven.

### Material limits and verification

Output/geometry and receipt-before-FIFO boundaries are matched, but input APIs
and algorithms differ: internal AppKit key-down/cached-texture blit versus Core
KeyChord/public command/normal view paint and upload. Native requires
framebufferOnly=NO for blitting. Vellum polls detached public snapshots while
native polls atomic slots; observer allocation/locking overhead is not isolated.
These differences prevent interpreting subtraction as pure production-framework
cost. No physical input-to-photon, native pixel-readback, resource allocation
balance, endurance or whole-framework release claim follows.

All eleven native process groups (pilot plus ten timed cases) and 88 probe groups
are reaped and observed absent. Both setup and timed framework scopes retire with
no pending records. Configuration probe bytes, geometry persistence, all runtime
source/binary/match pins and all 8,321 SDK members verify unchanged. The strict
classifier revalidates the actual journals; arithmetic, populations and archives
are independently checked. The previous refused attempts remain sealed.

[summary.json](summary.json) retains exact statistics and limits. The
[result evidence index](result-evidence-index.json) binds the
[result archive](result-evidence.tar.xz): actual records, commands, configuration,
ownership, match/source/binary bindings, analysis and review receipts. Executables,
full SDK and earlier archives are excluded. M3–M7 remain open. Results passed a
scoped Prism review using Anthropic Claude Sonnet 5.5. The derived JSON prose
spacing finding was corrected; archives and arithmetic were revalidated. No
findings remain unresolved.

## Attempt 2 — framework pilot valid; native startup refused

The corrected framework pilot passes strict setup accounting and locks a
1519×1266, scale-1, pixel-format-81 surface with payload bounds
(8,8)–(1511,1258), two drawables, vsync, no transaction presentation and no
next-drawable timeout. Callback drain and shutdown pass. The following native
process refuses foreground readiness before any startup or timed request.
Its generic error did not record the exact failing field, so the prior visibility
value must not be inferred. No timed pair exists. Both processes and 16 probes
retire; the successful framework setup is preserved.

Source inspection finds that the native wait checked active/key/foreground PID
but omitted visibility required by its following guard. Both now use the same
complete predicate. A Foundation unit proves the original predicate accepts an
invisible state and the new predicate refuses visibility/key/active/PID/loss
negatives. A failed guard now retains its state dictionary. Window number is
captured on main before request production. The native build passes warning-free;
Prism using Anthropic Claude Sonnet 5.5 has complete coverage with no actionable
findings. The framework executable and its verification are reused unchanged.
This is a concrete diagnostic repair, not an unmodified repeat or SDK policy fix.

[pilot2.json](pilot2.json) and [indexed evidence](pilot2-evidence-index.json)
bind the successful setup, refused native command, exact match and review/regression
receipts in [the archive](pilot2-evidence.tar.xz). Features and release gates remain
unchanged. A fresh bounded attempt still has to establish the paired calibration.

## Pilot 1 — refused; diagnostic repair prepared

The first live pilot refused its native configuration because the visibility bit
was serialized as JSON number `1`, while the strict check requires boolean `true`.
The actual capture retained active/key/own-foreground-PID state, but is still
invalid. No match file or timed pair was admitted. The process exited with status
1 and was reaped; all eight probes retired. Three pending setup acquisitions
became explicit retirement records at closure. None is reclassified as displayed.

The diagnostic now uses explicit Foundation boolean objects. A Foundation-only
regression reproduces the original numeric expression and verifies corrected
true/false output; the strict numeric-visibility negative remains refused. It also
waits for setup callbacks before closing the pilot, with a two-second refusal
bound. Eighteen classifier controls and both source builds pass. Prism using
Anthropic Claude Sonnet 5.5 reviewed this repair; its early-return concern was
checked against the existing worker defers and runtime/worker shutdown path.
No findings remain unresolved. No shipping code, SDK, workload, cutoff or release
gate changes. A fresh bounded attempt uses the same predeclared populations.

[pilot1.json](pilot1.json) and its [indexed evidence](pilot1-evidence-index.json)
retain the rejection, raw samples, process/configuration receipts, source-review
receipt, repair regression and dispositions in [the archive](pilot1-evidence.tar.xz).

## Preparation checkpoint — 26ac90a

The owner authorized a matched calibration on 21 October 2026 and explicitly
reserved approval of release-gate changes. The [predeclared protocol](../../../specs/macos-calibration.md)
keeps features and the 33.3 ms / 60 Hz gate unchanged. No live calibration has run
at this checkpoint; M3–M7 remain open.

The [diagnostic sources](../../../tools/macos-calibration/README.md) compare the
same opaque red/green rectangle on white, with identical framebuffer, payload
bounds and native presentation configuration locked by a setup-only pilot.
Receipts precede the originating FIFO. Native posts an internal Space NSEvent;
Vellum uses the existing measured desktop session and public command registry.
Input methods and algorithms differ: native blits cached reference textures;
Vellum does ordinary owner/reconcile/paint/upload work. This is a small-operation
response comparison, not isolated CPU attribution, physical input-to-photon,
60 Hz cadence qualification or proof of an OS latency floor.

Each of five alternating-order pairs retains 128 warmups and 120 measured
changes at 120 ms request deadlines. Invalid captures stop collection; latency
above the existing gate remains an observation. No earlier failed campaign is
reclassified or rerun unchanged.

The actual software-routed fixture verifies green/red/green model transitions
and every output pixel, plus unrelated-key/no-op rejection. Seventeen classifier
controls verify attribution, populations, readiness, losses, geometry, foreground,
schedule, nonfinite data and shutdown refusal. All pass. The optimized native
reference builds warning-free; the SDK consumer builds with the existing SDK's
AppKit deprecation warnings. All 8,321 immutable SDK members match the manifest.
Earlier fixture-only setup failures are retained: required manual theme fields,
unsupported custom manual extension, outer-surface color and Space chord spelling.
The final fixture uses ordinary public Column and command APIs.

Internal review and Prism with Anthropic Claude Sonnet 5.5 are complete. Entry
fields were already atomic; the final native source uses explicit atomic loads
and stores, including output, to remove ambiguity. Foreground observers now
require an actual own active/key/visible window. Monitoring covers the timed
cohort, not every startup instant. Two source findings have recorded dispositions
and none remain unresolved; both final source builds and focused controls pass.

[preparation.json](preparation.json) binds executables/build metadata and scope.
[preparation-evidence-index.json](preparation-evidence-index.json) binds the
[preparation archive](preparation-evidence.tar.xz), containing exact build/test
logs, compiler and build metadata, reviews and dispositions. Executables and the
full SDK are excluded; original local paths remain provenance. Diagnostic source
is committed alongside this report. Actual Metal matching, foreground behavior,
presentation and process retirement still require the bounded desktop check.
