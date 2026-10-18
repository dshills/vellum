# Popup routing, actual menu inspection and Selector focus

This checkpoint fixes two demonstrated framework defects. An ordinary posted
application wake was starved while a modal menu owned events; the driver now
routes its exact wake and originating session packets before popup modality,
including apps without DesktopSession. Unrelated Custom events keep ordinary
popup routing. Launched and unlaunched teardown clear the hook.

Selector's own Click handler opened/consumed its menu before the base handler
could take quiet focus. The Go adapter now takes actual focus first, only when
the native control is focusable and not read-only, without handling the click or
proposing a value. Real pointer selection, actual ContainsFocus/public focus,
keyboard reopening and Escape retention pass their software FIFO regression.

Stage 83 adds bounded detached current popup geometry to optional public
DesktopSnapshot.Popups. It separates actual menu buttons from persistent option
semantics, clips scene/button bounds with one origin conversion, and uses fresh
frame/scene/button incarnation IDs. Hidden/disabled/covered/pending targets refuse.
Copied metadata cannot revive another widget identity; IDs saturate safely with
no historical widget registry. Returned labels clone admitted bytes so a short
substring cannot retain a large backing allocation. Generated IDs/fixed records
share the public semantic output budget. Custom/nested menus keep their inspectors.

The gallery's runnable `-control-native-check` reuses ordinary public constructors
and public Go FIFO routing. Its current intended record is 57 inputs/21 independently expected observations;
the four preserved attempts used the original 49-input record. The complete native check remains FAILED. The last case
reaches 24 inputs/11 source-asserted observations: the ten Selector states and one
typed numeric value 42 with exactly one numeric proposal, then refuses the next
key because numeric Enter no longer leaves the NumberField focused. This prefix
is not a successful complete independent validation. Numeric rejection/invalid/
clamping/disabled continuation and explicit original-reference close/stale-input
checks were not reached. Enter intentionally advances focus to the next control
(the outside button in this view). Author a routed refocus before the next
numeric step; keep actual-focus admission and the domain expectations strict.

Four native attempts remain preserved and sealed: ordinary posted read starvation
(90.761s), missing Selector focus (1.272s), a consumer Command-mask guard failure
(2.380s), and numeric post-Enter focus (2.575s). The third is a test bug: `key.Meta`
is a bit index, not a mask. SetFlag now constructs the same actual Command mask
used by both guard and send. No user keymap is changed. Original-predicate wake,
original Selector and original Command-mask counterfactuals fail as expected;
corrected regressions pass. The final consumer passes the complete gallery race
suite and fresh vet/build. No further native case runs in this investigation.

Each native process exits 1 with complete logs, unchanged pins and its owned
process group absent. Final tracked session state is Closed/zero pending/zero
live windows; that does not prove Cocoa deallocation or resolve the separately
unclassified hidden TUINS window. All full native/Mac gates remain open.

Other passing checks: 42 selected owned Core methods/48 cases; complete driver-root
package (`.`) race; qualified Core vet (only unchanged upstream composites and
unreachable checks disabled); independent stage83 preparation of 6267 exact files
and all 11 preparation tests; 15 synthetic full-result validator controls. The last
controls are parser tests, not native acceptance. Driver/Core results are reused
only for exact unchanged affected bytes after the consumer-only modifier fix,
including authenticated non-gallery overlay correspondence. Actual adoption
preparation/source correspondence and default-worktree native build are checked
separately. Commit requires internal review, one Anthropic Claude Sonnet 5.5
Prism checkpoint and a bounded archive retaining the raw outcomes.

The first re-preparation after label cloning refused a stale stage83 patch SHA.
The failure is retained; the checksum correction and fresh preparation pass.
The retained SDK is unchanged: 7880 files, native archive 786a, Go 1.25.6 on
Darwin/ARM64 with controlled GOMAXPROCS=2 CPU-scene routing. Default worktree native
library compilation is separate from this SDK library's runtime evidence. There
is no SDK refresh, OS setting/input-source/VoiceOver/clipboard change, physical/OS
input, presentation/performance, human accessibility or full catalogue/M2–M7
qualification. Next qualify the corrected numeric refocus record natively, then continue the
remaining named catalogue scenarios.

An initial adopted build incorrectly overlapped dependency preparation and failed
because the new inspection methods were not yet installed. That failure remains
retained; a fresh build after preparation completes, pinning prepared Go source,
passes. Nine inherited workspace overlay mappings are not gallery inputs and are
excluded explicitly from the 31 mapped gallery/driver/Core source correspondence;
no inherited workspace change is adopted. The full prepared vector still matches.

Prism’s one retained medium finding correctly identified missing numeric refocus
in the reviewed 49-input checker. Before commit, typeNumber now routes an explicit
pointer click back to the NumberField after each Enter. Four such clicks add eight
inputs: the corrected record is 57 inputs with the same 21 model/callback/geometry
observations. The real software FIFO regression verifies Enter commits 42 once
and advances focus, pointer refocus changes no model, and Up proposes 43 once.
Fresh gallery race/vet/build and 17 revised synthetic validator controls pass.
The original 49-input checker/validator and all four failed native records remain
preserved. No corrected 57-input native case runs in this investigation; that gate
remains pending. Internal delta review resolves the valid finding; the two
discarded claims are independently assessed against the pinned modifier and
atomic budget contracts. The raw CLI exit 1/one retained/two discarded remains.
