# Empty sprite layer rendering

The native renderer previously cleared, rasterized and submitted a full-window
transparent sprite image on every fresh composition, including windows with no
drawable sprite. Stage 35 now omits that source before allocation or clear when
all three tiers lack a nonnil active sprite with a Draw callback. Fresh
composition replaces its source/operation list, restoring underlying content
after sprite retirement.

Active callbacks keep First/Normal/Final ordering and run on every composition,
even with an unchanged modification bit or a cursor's blink-off state. Active
event-only sprites keep their existing input/deadline semantics. HasActive skips
nil entries while retaining its Active-only criterion. Existing resources remain
retained for reuse; empty resize allocates no replacement, and reactivation uses
the existing size-correct allocation path. No paint scheduling, early cached
resize replay, partial upload or immediate GPU-resource reclamation policy changes.

## Correctness

The exact frozen original renderer fails the new allocation/source assertions;
the original liveness scan panics at a nil entry. Final core regressions and
existing cursor/manual/event/text checks pass three times under race. A
byte-identical original function body supplies the differential software pixel
oracle. Checks cover empty/nil/inactive/event-only tiers, repeated active callback
order, removal/movement/reactivation, 96/120/144/192 DPI resize, retained resources,
cached operation replay and an explicitly advanced actual TextCursor callback.

Managed software regressions wrap and forward the actual ComposerDrawer. They
check strict copy counts and detached pixels through 96/144 DPI retirement,
software resize/reactivation, actual field pointer focus and explicit-clock
blink/focus loss, controlled dialog/scrim, direct-scene popup and primitive
pointer drag/drop. The original renderer fails the managed empty-layer copy
assertions; the final four groups pass three times under race.

The full native adapter/commands pass offscreen race and default/offscreen vet.
Replay, grid, loaded-grid and loaded-tree consumers pass offscreen race; replay
also passes offscreen vet. Pinned preparation, preparation unit tests and SDK
producer/verifier unit tests pass. Eight detached grid frames at
96/120/144/192 DPI match the prior checkpoint's geometry, stride and RGBA digest.
The final prepared dependency inventory contains 6,129 exact files: only
core/rendersource.go and core/sprite.go change, and one core test is added.

Earlier managed fixture compile errors and zero-size popup fixture failures are
preserved with their source versions and raw logs. The final popup uses Core's
direct Scene construction and requires a nonempty rendered overlay. Historical
attempts do not substitute for the final frozen-source qualification.

These observations measure real software composition requests and pixels.
Managed frames deliberately invalidate their scene so a zero-composition pass
cannot succeed. They do not prove zero-paint scheduling, actual GPU upload
counts, window-manager resize, physical blink timing or display latency.

## Optimized native comparison

Three fresh optimized fixed-10k processes per build run interleaved A/B/B/A/A/B,
with 128 warm-up events and 120 measured events each. The baseline is the previous
unchanged-grid-refresh executable at `292ae67`. The candidate has the stage 35 patch
and the same framework Go production bytes and 400 clipped-cell geometry.
Build/source/prepared/binary identities remain stable across all measurements.
No race instrumentation, CPU profiler or external process observer runs during
the measured frames; the runner collects process observations only after end.

The table gives nearest-rank pooled elapsed UI-work percentiles from 360 samples
per build, in milliseconds. Every raw sample and outlier remains included.

| Build | p50 | p95 | p99 | Maximum | Mean | Sample SD |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Baseline `292ae67` | 13.407 | 14.178 | 14.891 | 18.912 | 13.512 | 0.508 |
| Candidate | 12.386 | 13.090 | 13.294 | 13.408 | 12.420 | 0.415 |

Pooled p95 decreases 7.67%. Per-process p95 in A/B/B/A/A/B order is
14.119/13.030/13.185/14.287/14.157/13.081 ms. The baseline maximum 18.912 ms
is retained, with cause unclassified. Coefficients of variation are 0.0376 and
0.0334 for baseline and candidate.

| Recorded stage | Baseline pooled p95 | Candidate pooled p95 |
| --- | ---: | ---: |
| Reconciliation | 3.007 | 3.692 |
| Layout/paint | 2.779 | 3.037 |
| Raster/upload | 8.907 | 7.406 |

Independent stage percentiles are not additive. Raster/upload cost decreases
in this scoped comparison; reconciliation and layout/paint observations increase.
Total elapsed UI work includes GC and matching intermediate post-input paints,
and excludes queue/surface waits and backend acquire/present calls. It is not
OS-thread CPU time or physical display latency.

Both unchanged targets still fail: 4 ms p95 and 6 ms p99. This exploratory campaign
has fewer than five fresh processes per case and omits variable heights,
million-row cases, 500-visible-row tree timing, cold caches, native/GPU memory
attribution and endurance. Full P6 and broader Mac acceptance remain open.
Next work is source-level raster/upload attribution and a separately specified
bounded damage/repair policy if the evidence supports it.

## Evidence and review

[records-index.json](records-index.json) lists exact raw members and SHA256
digests in [records.tar.gz](records.tar.gz). Source/producer snapshots are inert
text, including failed attempts. Extract to a new directory to inspect the
recipes. Original absolute paths identify where records were produced; the
index supplies archive member paths. Final evidence includes:

- qualification/core-qualification/qualification.json and original overlays,
  assertion failures, nil panic and final race logs;
- qualification/integration-qualification/qualification.json, fresh preparation,
  frozen managed original/candidate runs, full adapter and consumer race/vet,
  preparation and SDK unit checks;
- qualification/pixel-comparison/qualification.json and current/previous frame logs;
- qualification/comparison.json, baseline/build identities, host metadata and
  all six native run logs/results;
- internal core/managed reviews, independent raw measurement verification,
  frozen-source handoff, historical fixture attempts and inert producer recipes.

[Prism review](review/prism.json), [input scope](review/inputs.json) and
[process record](review/process.json) retain one significant-batch review using
Anthropic Claude Sonnet 5.5. Its explicit staged implementation/test/spec/report
payload excludes raw archives, historical source copies and producer records.
The final [internal review](final-review.json) and [audit](audit.json) bind reviewed
files, exact prepared bytes and archived checks. These checkpoint records do not
qualify public distribution, signing/notarization or the full Mac milestone.
