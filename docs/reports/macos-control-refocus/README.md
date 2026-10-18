# Retained disabled styling and actual multi-click focus

The complete ordinary Selector/NumberField native record now passes: **57 Go FIFO
inputs and 21 independently expected observations**. The source includes two Go
adapter fixes over `4d0b5f7`; the consumer, validator, expected models and input
sequence are unchanged. This qualifies that named native routing record in the
controlled configuration below. Full catalogue/M2, physical/human M5, performance,
resource/endurance, release and full Mac completion remain open.

The first fix refreshes a retained widget after its effective Disabled state
changes, after control synchronization and before synchronizing descendants.
An ancestor update could previously style children using their old Disabled
flags; clearing the flags later left read-only dimensions and Focusable ability
stale when the child's own descriptor had not changed. Passive widgets are
included. Re-enabling a parent preserves explicitly disabled children and widget
identities; it does not request focus or invoke Change.

The second fix restores actual scene focus before ordinary double/triple-click
handlers on managed text and numeric fields. A disabled field can retain a stale
visual Focused flag after an ordinary key moves actual focus. Native multi-click
handlers use that flag and consume the event without focusing the field. The Go
adapter's first-phase handler checks actual ContainsFocus, enabled/read-only state
and Focusable, then calls SetFocus only when needed. It does not handle the event,
change values or replace ordinary word/all selection.

## Evidence

- The original driver fails the two inherited NumberField/TextField focusability
  regressions; direct-disable controls pass. The styling fix passes all four
  through three cycles, plus passive Label styling and own-disabled preservation.
- A gallery-shaped software diagnostic at 729x658 and DPI162 reproduces the
  remaining actual-focus failure for rapid clicks, while advancing the manual
  clock beyond the double-click interval passes. A generic already-focused
  fixture did not reproduce it and is retained as diagnostic evidence only.
- The original multi-click handlers fail all four rapid double/triple field
  cases; all four separated controls pass. The final source passes eight cases,
  actual word/all selection, no model proposal during refocus and exactly one
  subsequent numeric/text edit. Together with the five disabled-state cases,
  thirteen focused subcases pass under race.
- The final complete driver-root package race suite, driver/gallery vet and
  native gallery build pass. No Core/dependency patch changes; stage83's full
  6267 prepared-file vectors remain identical.
- The final native process passes all 57 inputs/21 observations, including
  numeric rejection, invalid/corrected/clamped values, disabled target refusal,
  re-enabled input, outside-release consumption, exactly-once callbacks, closing
  with an actual menu open and stale original-reference input refusal. The
  independent validator verifies the complete record, not a successful prefix.

Two earlier fresh 57-input attempts in this checkpoint remain **failed and
sealed**. Both reach 54 input receipts/19 source-asserted observations and fail
`reenabled-step target not actually focused`. The styling-only native case restores
correct widths but does not restore actual focus. They are not passes. The four
older 49-input failures remain in the preceding
[checkpoint](../macos-native-controls/README.md). No sealed case is reset or retried;
each changed-source case has a distinct plan/admission/lease/stop scope.

The first software prefix fixtures refused a missing title and the manual
adapter's explicit font/color requirements before input. Those failures remain.
A first post-focus software assertion expected an append after a word/all
selection; the result correctly replaced the selected word/all text. That failed
assertion and original bytes remain; the final test explicitly expects `alpha x`
for word replacement and `x` for whole-field replacement. No native oracle changes.

## Configuration and limits

macOS26.6.2 build25G83 ARM64, actual 729x658 scene pixels at logical DPI162,
Go1.25.6 Darwin/ARM64, CPU-scene native routing, controlled GOMAXPROCS=2 and the
retained SDK's pinned native archive `786aaad8...` are explicit. This is current
Go source/stage83 over that library, not a refreshed SDK or qualification of the
source checkout's other default native archive. The retained SDK remains exactly
7880 files with whole-vector SHA256
`4ce384e5873e00f148b72458ead131f42d4b3d7a464980423cb2f01881e94f26`.

Every native process has complete logs, stable pins and confirmed owned group
absence. Final public session state is Closed, zero pending and zero live windows;
this does not prove Cocoa deallocation or classify the separate hidden TUINS
window. No OS input operator, physical input, IME/VoiceOver, preferences, clipboard,
display transition, presentation/performance, signing or public release is used.
No full Mac gate closes.

One complete Anthropic Claude Sonnet5.5 Prism call reviewed 13 files/42,269 bytes.
Its raw CLI exit 1 and two retained medium findings remain preserved. Independent
internal review refutes both against the actual definitions: IsReadOnly includes
Disabled, and controlDisabled walks every ancestor property. No runtime change or
second provider call is required. Final commit requires the bounded archive
integrity check and documentation-delta review. [Structured results](records/results.json) and the bounded
[raw evidence archive](records/checkpoint.tar.gz) retain successes, failures and
counterfactuals. Continue the remaining named native catalogue scenarios, then the
unchanged M3–M7 delivery gates.
