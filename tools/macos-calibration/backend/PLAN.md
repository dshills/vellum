# Actual framework backend timing observer

Feature freeze and release gate remain unchanged. The native combined/split
hypothesis is closed as insufficient. Observe actual render submits, GPU execution
and present requests in a private clone of the exact audited wgpu-hal 29.0.3 and
wgpu-native source. Never install the diagnostic archive in the retained SDK,
shipping framework or packaged app. Its ABI is private and diagnostic-only.

Patch only commit call sites and present submission with fixed 4096 atomic records,
no eviction, and per-record command-buffer scheduled/completed and drawable
presented callbacks. Record queue identity, ordinal, kind (render/extra signal/
present), CPU commit/request/observed clocks, actual GPU start/end/status, and
actual drawable presentation. No callbacks retain buffers/drawables in storage;
Metal copies owned callback blocks. Overflow, incomplete callbacks, missing clocks,
status failure and ambiguous attribution refuse the run. Configure once before
creating the runtime; collect after shutdown with a two-second drain and 2 MiB
JSON bound. Record source/build/library identities and preserve failed builds.

Use the unchanged fixture view, commands and receipt-before-queue boundary,
128 warmups +120 measured requests at120ms, existing complete parent classifier,
locked geometry/configuration and foreground checks. Both on/off cases run the
same diagnostic binary/library; off records no callbacks/clocks and still branches
through the private wrappers. This is an observer-impact control, not proof of
zero overhead or binary equivalence with the retained original archive.

Run three pairs off/on, on/off, off/on, one attempt each, 240 seconds globally
and 40 seconds per process, exclusive lease/owned executor/config probes as before.
Stop at the first invalid process, focus/config/source drift, deadline, missing or
ambiguous timing. No unchanged retry or fallback. Preserve setup losses and tails.
No new setup pilot is needed: use the sealed valid match and refuse drift.

The first f9462e8 scope is sealed/rejected: its observer recorded six startup
presents (two zero/four positive) before public recorder admission. All252 public
records matched; equality of whole-lifetime versus admitted populations was an
incorrect assumption. Actual extra drawable IDs were not recorded then; do not
retrofit or accept that scope. Use a fresh private build/campaign adding IDs.

On-mode requires one queue, strictly ordered commit reservations, successful
completion for every recorded buffer and complete callback drain. Native drawable
IDs and admitted public IDs must each be unique/increasing. A release-published
recorded flag distinguishes a valid native ID zero from an unset slot; zero is
allowed by the Metal API. Nonpresent rows must not claim drawable recording. Match every admitted
ID exactly once, with actual shown timestamps equal within one microsecond
(serialization precision only), including admitted zero/skipped feedback. Allow
only an explicitly retained leading startup prefix of at most64 lower drawable
IDs whose requests precede first public acquisition, and whose positive feedback
precedes first cohort receipt. Record every prefix ID and zero/positive count;
exclude none from the raw trace. Any unmatched admitted/cohort ID refuses. For
each cohort frame attribute all render commands after the preceding present
ordinal and before the matched present ordinal. Require at least one, positive
GPU times, execution after commit and completion before shown. Select the final
120 measured frames by explicit measurement scope/tag/generation membership;
frame tags are opaque and need not be consecutive. Empty present/signal buffers
may have zero GPU times. Callback observations are CPU clocks and do not impose
invented GPU/callback ordering.

Report per-run distributions for receipt-to-submit, submit-to-GPU-start, render
GPU span, last-render-GPU-end-to-present-request (signed) and GPU-end-to-shown.
Do not add quantiles, pool populations, call differences pure framework cost or
adopt a repair/gate waiver. Observer-enabled timing cannot alone establish the
shipping path's latency. Focused synthetic refusal controls, original view/routing
verification reuse, build, internal review and Prism Anthropic Claude Sonnet5.5
precede native execution. Validate/review/commit actual results separately.
