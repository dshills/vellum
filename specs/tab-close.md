# Controlled document tab closing

Tabs options may set `Closable: true`. Such descriptors require
`Node.TabClose func(ui.TabCloseEvent)`. Closable metadata and TabClose belong only
on Tabs. Each request carries the Tabs node ID and stable PageID; labels and
indices never identify a document. Nonclosable tabs retain standard presentation;
strips containing closable options use document-tab presentation.

Closing requests an application decision on the bounded originating-window owner
queue. It neither selects nor deletes a native page before that decision. The
latest callback runs; it accepts by removing the matching option/page and choosing
a remaining selected page, or rejects by leaving accepted state unchanged. The
last close must remove/replace the Tabs node because empty Tabs are invalid.
Separate gestures may each request a close; applications should check their
current document membership when handling a request. Removed child controls can
be reparented before native page destruction, preserving uncommitted drafts.

Pointer activation of the private close button and Delete/Backspace while a tab
header is focused request closing that tab. Modified chords remain available to
application commands/window handling. Close buttons add no managed Tab stop.
Disabled ancestors/options, hidden containers, retired tabs/artwork and closed
dialog lifetimes refuse requests. Native fallback deletion is intercepted before
parent/index lookup, and the gesture is consumed even when owner admission fails.
Nil native CloseRequest preserves unmanaged backend behavior. Managed retirement
retains a guarded callback, never a nil destructive fallback.

Requests bind the originating tab, private close button and immutable capability.
Recreated close artwork, hiding/disabling and genuine dialog reopen retire old
capabilities. Native dialog dismissal closes an atomic lifetime token before
parking children, even if the application leaves Open=true. Scope tokens are
captured in an immutable owner snapshot; callbacks drained by a peer window do
not inspect or mutate originating native trees. Queue saturation reports the
ordinary control error without deleting content or adding a retry queue.

The Mac semantic close button is a named passive-focus action child of its option,
with its own exact geometry, stable qualified ID and press action when enabled.
Disabled buttons expose explicit empty action arrays. Private IDs use the reserved
`tab-close:` prefix; descriptor validation prevents manual target collisions.
Manual target incarnations change when a tab or its private artwork is replaced.

Offscreen actual pointer/key routing, native-object lifecycle, original unmanaged
behavior, rejected/accepted closes, draft reparenting, stale/private identities,
queue refusal, hidden/disabled/dialog capability retirement and serialized
semantics are required checks. These do not qualify physical input, human
VoiceOver or a live native lifecycle/performance campaign. The gallery includes
Document tabs with close rejection and a last-document empty state.

## Native acceptance sequence

The separate bounded `drivers/cogent/cmd/tab-close-smoke` command must complete
actual Cocoa action selectors, exact owner callback counts and native publication
for rejected selected/nonselected closes, preserved edited draft identity,
accepted nonselected and selected closes, last-page empty state, disabled/absent
refusal, and re-created native identities. Every stage requires a five-second
bound and owner FIFO acknowledgement after admission; accepted proposals are not
reissued while awaiting native publication. Terminal cleanup must report no
windows, accessibility registrations, native operations or pending callbacks.
It does not synthesize global keyboard/pointer input or certify human VoiceOver.

The gallery's new `-structure-native-check` authors selected-document pointer/key
rejection, accepted keyboard close, unchanged selection callback count and
new-frame stale refusal within a 50-input/42-observation scenario, reusing the
ordinary document view. Its first native run stops at a measured-frame failure
before these close observations; they remain native-pending. Actual public
Manual routing and focused regressions pass. See the
[result](../docs/reports/macos-structure-catalogue/README.md). The separate Cocoa
selector/physical/human acceptance requirements above remain unchanged.

The structural native prefix above predates the subsequent Prism tab activation
and disabled-focus fixes. Those fixes pass actual Core queue/manager routing and
complete driver/gallery race suites, but have not received a new native run.
They retain the failed scope and do not close the full structural/native gate.
