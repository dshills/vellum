# Mac window visibility and owner redraw

The Mac adapter now observes its original NSWindow's occlusion-visible bit and
requests an owner redraw on a visible transition. A clean scene is explicitly
marked for rendering; merely waking the paint loop could skip composition after
a previous failed acquisition consumed scene dirt. The event bypasses popup
modality without draining unrelated application work or changing selection.

The observer subscribes before the initial main-thread sample. Early samples are
retained until OnShow binds the source after rendering configuration. Duplicate
notifications and bursts retain one exact source/state/wake capability. Every Go
bridge retirement removes pending nodes and clears the source; Cocoa detach
removes the observer/handler before window destruction. Dequeued stale events
cannot consume a newer wake. A retained local callback snapshot avoids a nil
second property read if a foreign manual notification races invalidation. An
already captured foreign callback is not joined by invalidation; the production
main-thread gate and retired Go state keep it inert.

Public `DesktopSession.Visibility(ctx, originalRef)` returns copied
Known/Visible/Sequence values through the original FIFO without requiring or
granting a completed frame. Unsupported/unobserved backends return Known=false.
Sequence is attachment-local. This is a sampled Cocoa bit, never drawable
readiness or presentation; Frame preserves first-composition failure and no
automatic acquisition or retry. No input eligibility rule changes.

## Software evidence

- focused-v9: PASS, 14.560 seconds; owned group absent.
- driver-post-provider: PASS, 290.469 seconds; owned group absent.
- gallery-post-provider: PASS, 34.916 seconds; owned group absent.
- vet-post-provider: PASS, 12.407 seconds; owned group absent.
- build-post-provider: PASS, 12.282 seconds; owned group absent.

Actual Manual/Core routing checks a clean scene and an actual open Chooser popup:
one visible transition recomposes once; duplicates, foreign state and retirement
cannot redraw. State tests cover early binding, bounded burst coalescing,
source/wake identity, stale receipts, overflow and concurrent observation/close.
Public visibility inspection, cancellation and existing terminal unavailable
Frame behavior pass. A Go-only counterfactual removes Scene.NeedsRender: both
clean/popup tests fail with the expected missing-composition assertion, twice.

The native Objective-C unit is compiled from actual candidate files. Foundation
fixtures create no NSApplication or NSWindow: they prove registration/filtering,
initial sampling, handler revocation, weak release after pool drain and retained
callback lifetime under deterministic getter-triggered invalidation. They do not
prove real Cocoa notification timing, Metal availability or displayed pixels.

Earlier software failures remain recorded: promoted CustomEvent.Data cannot be
set in its outer struct literal; Go overlays left the old SDK Objective-C unit in
the native compiler, causing the new fixture link symbol to be missing; and the
second pointer in the fixture's combined weak declaration was actually strong.
A compiler diagnostic established that ownership-fixture error. Both probes are
now separately weak, with installed/cleared/released positive controls. No
production observer leak or stack-block lifetime defect is inferred from those
fixture failures. The later local-callback fix received fresh checks above.

Checks use Go1.25.6/GOMAXPROCS2, actual candidate sources and an external dependency
manifest pinning prepared Core/typesetting and the retained source-built786a
native library/module. Checkout go.mod, prepared outputs and SDK bytes stay
unchanged. These software checks launch no native application. Historical raw receipts remain in
the local checkpoint workspace; the bounded archive retains their logs/ownership
records and condensed receipts with original receipt hashes, plus complete
current final receipts and exact sources. This is not a refreshed SDK package.

An internal source review is required before the single Anthropic Claude Sonnet
5.5 Prism checkpoint review; the commit requires its raw output and an
independently checked disposition to be retained with the records. The final commit review checks that endpoint.

## Still open

The native lifecycle record below qualifies this one clean/popup hide/show
scenario. No old sealed catalogue case is replayed. It does not diagnose earlier
acquisition failures or promise recovery when a backend yields without a Cocoa
state change. Complete structural and text/modal catalogue acceptance remains
required under a fresh scope.

The failed text/modal35-input19-state prefix and structural failures remain
failed. Full native catalogue, physical/OS input, IME, human VoiceOver, native
pixels/presentation, M2/M5, performance/resource/rights/distribution and full Mac
completion remain open. No system settings or SDK refresh occurs.

## Native visibility lifecycle evidence

Clean runtime source `2aca9b6` passes a separately admitted one-window native
consumer in 2.960 seconds. Four actual same-window Cocoa notifications and five
copied visibility observations match two hide/show cycles, first with a clean
scene and then with an actual Selector popup. Both hidden explicit Frame requests
return typed terminal unavailable from one occluded acquisition attempt and grant
no Snapshot capability. Ordinary composition calls begin after each observed
visible notification and return before a new explicit measured request. The two
new requests complete; the popup and its original item incarnation remain intact.

There are seven explicit Frames, five successful snapshots, two Go FIFO pointer
records and 21 total session requests. Ten ordinary End calls are observed in
total: this is not one native frame per transition or proof of GPU/presentation
success. The software counterfactual above separately identifies the observer's
scene-dirt requirement; native expose handling may also cause redraw. The private
entry-ordinal witness is deliberately conservative about notification ordering.

Original reference refusal, private observer retirement and tracked session/AX/
callback cleanup pass. Owned group1195 is absent with complete stable streams and
whole clean source5498/prepared6276/SDK7880 post-boundaries intact. One inactive,
untitled, hidden TUINS window remains sampled; native deallocation/zero-Cocoa-
window, memory, physical/OS input, native pixels and presentation are unqualified.
Controlled Go1.25.6/GOMAXPROCS2/SDK786a CPU-scene routing is the tested scope.

The preceding private consumer fails before its first public session operation:
Issued0/Measured0, CLI1 after37.789 seconds, owned group97947 absent. Its original
case remains failed and sealed. The helper boxed the C logical expression as JSON
integers0/1 while Go required a boolean, discarding replies and timing out. Actual
original/fixed-helper Foundation fixtures with nil/fake objects prove numeric
versus boolean wire types without constructing NSApplication or NSWindow. The
original malformed native bytes were discarded, so their exact contents are not
retrospectively observed. The separately corrected consumer uses explicit JSON
booleans and bounded raw decode-fault evidence. No product defect or Cocoa queue
delivery diagnosis is inferred from this private failure.

Actual-source build/vet, 26 independent negative validator controls, source/plan
admission and independent result reviews cover the corrected private consumer.
No runtime source or SDK/prepared byte changes follow the existing Prism-reviewed
implementation. Both native scopes, source/controllers, exact plans, receipts and
raw streams are in [the bounded native archive](records/native-checkpoint.tar.gz)
with [member hashes](records/native-manifest.json). Binaries/SDK and earlier unused
preflight compile versions remain local; they are not copied recursively into
this archive. No second implementation/provider review or full Mac gate closure.

## Provider disposition

One complete uncached Anthropic Claude Sonnet5.5 Prism call covers the original
15-file41488-byte payload. Raw CLI1/one retained and one discarded finding remain
preserved. The retained sequence-limit concern is addressed by explicit
visibility-redraw retirement documentation and a queued-wake exhaustion test;
ordinary application-driven painting is unaffected. The discarded publication
concern is addressed structurally: the native bridge now initializes visibility
before registry publication. No pre-fix concrete race is claimed as reproduced.
Current final checks cover the corrected constructor and overflow regression;
independent review reconciles all post-provider changes. No second provider call.
