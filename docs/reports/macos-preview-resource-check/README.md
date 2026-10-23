# macOS preview: activity and bounded resource checkpoint

22 October 2026. Features and approved preview policy are unchanged. This is a
partial resource result, **not a passing two-hour soak or release approval**.
Fixtures were built from SDK source `d8de4fc`; their Go/native dependency inputs
match the alpha.3 product runtime. No runtime changes were made during these runs.

## Results

- Ten-minute composed workbench: `PASS_COMPOSED_ACTIVITY`, 600.144 seconds,
  2,400 steps, 600 native text edits, 200 resizes and 1,800 workers. Of those
  workers, 600 committed and 1,200 stale completions were discarded; zero workers
  were reported by the activity manager at its last observation. This is
  `PASS_COMPOSED_ACTIVITY` only, not shutdown/quiescence certification. The observation occurs inside
  an owner callback (`Dispatching=1`, `Closed=false`); it is not a closed-state
  snapshot. The process subsequently exited normally (outer exit 0); that exit alone
  does not establish native deallocation. Document safety below is an independent
  fixture result, not inferred from this observation.
- Two-hour soak attempt: failed after 1,815.94 seconds at the visible-to-occluded
  transition. Six eligible 60-second visible samples completed during a 30-minute
  visible phase. CPU ranged 0.0367–0.0908% of one core, settled Go heap
  59,384,608–59,388,096 bytes and RSS 294,633,472–294,666,240 bytes. Every quiet
  sample had zero primary paint wakes, empty-event post attempts and surface
  presentation attempts. These are selected counters, not complete wakeup or
  private/native/GPU allocation measurements.
- An unchanged 80-second short pilot reproduced the transition refusal in
  36.27 seconds. Its five-second samples cannot qualify a 60-second idle budget.
- Existing document-safety fixture: native unsaved-close refusal and modal Cancel,
  exact million-component save/load, failed Save retaining the unsaved window,
  and successful Save acknowledgment before close all pass; outer exit 0.

## Exact remaining blocker

Reproduce with the existing `window-cycle-smoke` command:

```sh
window-cycle -composed -cycles=1 -sample-every=1 -soak=80s \
  -soak-quiet=5s -soak-settle=1s -soak-interval=10s -timeout=2m
```

The target reports Cocoa frame `[2803,0,1944,1126]`; the opaque, alpha-1 cover
reports `[2771,-32,2008,1190]`, same window level, and a lower ordered index
(in front). Full frame containment is true, but the target's native
`OcclusionVisible` remains true. The fixture correctly refuses to call that an
occluded sample. Containment and ordering are not substituted for that predicate.
Cause remains unresolved: these observations do not establish a renderer failure,
a leak, or an operating-system defect. The minimized/background phases were not
reached. No complete two-hour soak pass is claimed.

Smallest next action: establish a genuine native occluded transition in a short
control on this desktop before committing another two hours. Keep the strict
state predicate and original failed records. No performance campaign or release
policy change follows automatically.

## Ownership scope

Both failed soak runs retire all tracked Vellum windows, panels and callbacks:
`NativeOutstanding=0`, `AccessibilityWindows=0`, `PendingCallbacks=0`, callbacks
stopped. One hidden, untitled `NSKVONotifying_TUINSWindow` with `NSView` remains,
as in the earlier failed and passing 100-cycle records. The metadata-only class
origin probe identifies Apple's TextInputUIMacHelper as the base `TUINSWindow`
implementation. Stable count and Apple class origin are established; the retaining
reference and native/GPU heap deallocation are not. Native snapshot code copies
scalar/string DTOs and serializes them; this is not an ownership graph.

The passing [100-cycle checkpoint](../macos-preview-lifecycle100/README.md) remains
valid within its recorded scope. Full endurance, native heap and GPU heap
qualification remain open. No global input, accessibility, display or power
settings were changed. Automated native events/fixtures are not physical pointer
or human VoiceOver/IME qualification.

`INDEX.json` binds the normalized records. Compressed journals preserve failed
results. Home paths are replaced with `<USER_HOME>`; test timestamps, process IDs
and window geometry remain. The bulky Go execution trace stays local.
