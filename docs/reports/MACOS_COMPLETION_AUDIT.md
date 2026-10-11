# Mac framework completion audit

Audited at the `63a369f` to imaging/asynchronous-capture transition. The current
phase's exact sources, native results and review hashes belong in its acceptance
record; the earlier commit alone does not include that work. This audit compares
public implementation against [Mac scope](../../specs/macos-completion.md) and the
retained P4–P11 workloads. Historical prototype hold instructions were superseded
by subsequent implementation authorization; acceptance workloads remain relevant.
Windows/Linux runtime qualification is deferred.

## Remaining implementation

- **Supported multiline editor:** public controls currently expose a single-line
  `TextField`. Proven multiline input/geometry code remains in the foundation
  prototype. Promote a portable editor model with explicit position units,
  selection, atomic edits and revisions; a lifetime-safe native adapter; text AX;
  history/document integration; an independent editor consumer and SDK contracts.
- **Deterministic replay:** injected clock/task execution, bounded run-until-idle,
  recording and public replay consumer are still absent from P9.
- **Semantic automation:** P10 needs an opt-in local typed/versioned transport,
  IDs, stale/disabled checks, cancellation, audit results and sensitive-field
  redaction. Existing in-process/native AX actions do not supply that transport.
- **Full scalable grid:** fixed-height pooled row buttons do not implement P6's
  variable heights, horizontal virtualization, frozen columns, editable cells,
  multiple/all-row selection or jump-to-key. Sorting/filtering/viewport bounds
  already exist and should be retained.
- **Custom-content semantics and service scope:** selected canvas/domain objects
  and detailed editor text need native semantics. P10's richer clipboard MIME/image,
  outbound drag/drop, tray and notification capability reporting need explicit
  implementation or a recorded scope decision.

Document external-edit detection and crash/power-loss guarantees are expressly
excluded from the current bounded persistence contract. This audit does not add
those exclusions to the baseline requirement.

## Remaining automated native evidence

Presented clipping and MSAA/compositor behavior must be qualified separately from
offscreen snapshots. Current-framework editor/input-order/range-geometry fixtures
need an external-process AX client. P11 still requires representative integrated
frame/input percentiles; retained Go/native/font/GPU memory; 60-second idle,
10-minute activity, 1,000 rendered lifecycle cycles and overnight soak. Short
historical samples establish narrower workloads. Each new public contract needs
fresh committed SDK/package consumption evidence.

## Actual user and hardware workflows

Real VoiceOver editing/navigation/announcements; IME conversion and candidate
placement after scroll/focus/scale; physical canvas/docking gestures; Finder drops;
chooser selection, native menu activation and clipboard roundtrips; mixed-DPI
monitors, monitor removal and physical GPU-loss policy remain distinct gates.
Automated selectors are not substitutes. Routine checks do not alter OS input
sources or accessibility preferences.

## Public distribution

The selected namespace `github.com/dshills/vellum` and MIT root license resolve
owner choices. Embedded native/Rust/font redistribution still needs audit; five
native-library modules currently lack their own module license declaration.
Signing/notarization and clean-user Mac execution remain separate qualification.
Notice collection and relocated same-host consumption do not certify those gates.

Next implementation checkpoint: the supported public document editor, followed
by deterministic replay/typed semantic automation and the complete scalable grid.
