# Matched native/framework calibration

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
