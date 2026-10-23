# macOS developer-preview: two-hour native soak

**PASS_NATIVE_WINDOW_SOAK.** Four 30-minute phases completed; process exit 0,
no outer timeout, owned process group absent. This clears the bounded two-hour
preview soak check, not full P11/native/GPU heap qualification or public release.

Run: 22 October 2026 22:11:19 UTC to 23 October 00:11:31 UTC
(17:11–19:11 Chicago, 22 October). Outer elapsed 7,211.705 seconds includes startup,
transitions and shutdown; each workload phase ran at least 1,800 seconds.

## Candidate and method

Fixture source: `627a3bc`. The [setup checkpoint](../macos-preview-soak-setup/README.md)
records focused race tests, vet, Prism review, short pilots and the sole fixture
change: cover inset 32→128 points. Strict native occlusion and endpoint checks
remain unchanged. Earlier failures are preserved, not relabeled.

The native command was built from a copy of final SDK `bb5bbba8…`, source
`10beef4`, with only that fixture file changed. All other SDK manifest file hashes
match. The linker map binds the shipped ARM64 native archive SHA
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`;
fixture binary SHA is
`7a73d546aedb6b04e7dde6b8e7edf526d16794249742e38a5fd11c4ceab74f8e`.
The delivered SDK and app were not modified. No product runtime change occurred
before or during the soak. This is the composed native framework fixture, not a
claim that the Placement app itself was interactively edited for two hours.

```sh
window-cycle-candidate -composed -cycles=1 -sample-every=1 \
  -soak=2h -soak-quiet=60s -soak-settle=10s -soak-interval=5m -timeout=2h2m
```

The four phases are visible/unfocused, genuinely occluded behind the cover,
minimized, and visible/unfocused with periodic bounded background workers.
Six eligible 60-second quiet samples completed in every phase: **24 total**.
Actual native state was accepted at every transition and each sample endpoint;
endpoint checks do not prove continuous state between observations. All six
background workers completed and committed before their corresponding quiet
samples. Final cancellation/owner retirement joins completed without error.

## Resource observations

Every quiet interval has zero primary paint wakes, empty-event post attempts and
surface-presentation attempts. These are selected framework/native counters, not
complete app-originated wakeups or actual OS-presented-frame measurements.
Quiet CPU ranges **0.0403–0.1424% of one core**. These are observations, not a new
release threshold or a responsiveness/latency result.

| Phase | Settled Go heap, bytes | RSS, bytes | Goroutines |
| --- | --- | --- | --- |
| Visible | 59,403,248–59,409,944 | 299,597,824 | 9 |
| Occluded | 71,086,992–71,090,992 | 296,894,464–296,943,616 | 9 |
| Minimized | 59,523,400–59,527,952 | 305,807,360 | 8–9 |
| Background | 59,458,104–59,466,624 | 305,872,896–306,905,088 | 8–9 |

Go heap remains stable within each arrangement. RSS changes across window-state
transitions and the allocating worker phase: first-to-last overall change is
+7,241,728 bytes; background first-to-last is +966,656 bytes, with the last samples
plateauing. Retain these deltas rather than calling RSS constant or a proven
absence of leaks. RSS is resident pages, not private footprint or native/GPU
allocation attribution. No configured memory guard was exceeded; this finite run
does not establish an indefinitely bounded native heap.

## Shutdown and limits

The cleanup snapshot reports no visible Vellum window titles, no tracked native
resources, accessibility windows, file panels or pending callbacks; callbacks
are stopped and menu/tray targets retired. Process exit is corroborated by absence
of its owned process group. One hidden, untitled `NSKVONotifying_TUINSWindow`
with `NSView` remains in the pre-exit snapshot, matching previous runs. Apple base
class origin and stable observed count are known; its retaining reference and
native heap deallocation are not. That limitation remains explicit.

This single supported-host regression passes the approved preview resource check.
It is not the 1,000-cycle/eight-hour/full-platform certification, human physical
input, VoiceOver/IME acceptance, or proof of native/GPU deallocation. Existing
latency disclosures and release policy remain unchanged. No global input,
accessibility, display or power settings were changed. Distribution-rights,
candidate CI and publication decisions remain separate and pending.

`INDEX.json` binds the records. The compressed journal retains the actual native
result; `quiet-samples.json` exposes all sample endpoints and measurements.
Home paths and spacing in the human-readable process scope disclaimer are
normalized; measured values, result flags, test timestamps, process IDs and
geometry are unchanged. The original process receipt remains local.
