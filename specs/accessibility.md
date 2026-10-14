# Native accessibility and keyboard navigation

Vellum publishes managed controls to macOS AppKit `NSAccessibilityElement` objects attached to the actual GLFW Cocoa content view and window. This driver-owned bridge uses no prototype runtime, AccessKit binary, web view, or operating-system preference changes.

## Ownership and lifetime

Native accessibility queries read immutable AppKit snapshots on the main thread. Worker query/action requests cancel when their attachment retires and have a five-second maximum wait, including calls with a Background context, so a stopped native queue cannot leave a caller waiting indefinitely. Go builds semantic snapshots on the application owner during the originating window refresh. Native actions only admit bounded `PostWindow` work; they never enter `Do`, wait for the owner, hold a bridge lock across callbacks, or mutate controls from a foreign native thread. Delivery rechecks the native attachment token, semantic generation and window lifetime, then resolves the current owner-held action callback. Identical semantic refreshes replace callbacks without advancing the semantic generation; a changed tree retires the previous generation. Retired objects, stale generations, disabled targets and a full queue refuse actions. A native admission means queued work, not that the controlled model accepted a proposed value.

Native objects retain identity across updates by identifier. Managed IDs, option IDs, page actions and virtual row IDs use separate encoded namespaces, including both extension identity and domain row identity. Application IDs containing separators cannot collide with synthetic IDs. Native payloads validate completely before retained elements change, including identifier uniqueness, parent links and scalar/action types. Native teardown unregisters callbacks and retires elements on the main thread before GLFW destruction; it restores the content-view class only while the exact bridge-installed subclass still owns that class slot. The native NSWindow class is preserved. Failed pre-show runs release the native attachment configuration so the existing retry contract remains valid.

## Supported semantics and actions

Buttons, labels, checkboxes, selectors, sliders, number fields, text fields, tabs, scroll areas and dialogs expose roles, labels, help, enabled state, accepted values, relationships, selection and focus. Labels are independent of presentation text. Inherited disabled state removes actions. Numeric proposals reject nonfinite/malformed values and clamp finite values to the declared range. Selectors and tabs accept stable option IDs or a unique enabled option label; disabled and ambiguous labels are refused. Text fields publish their accepted editing value and directional selection from an owner-held cache. That cache refreshes after native publication, completed outgoing focus loss, and controlled text/selection reconciliation. An unchanged descriptor preserves visible preedit; a coalesced publication or focus handoff cannot leave an unfocused field exposing an older preview. These updates neither reactivate the outgoing client nor emit an extra controlled Change. Text fields publish their accepted full value; detailed native attributed text, text selection ranges and multiline editor accessibility remain separate acceptance gates.

Workbench dock headers publish panel choices and each active panel’s managed controls. Panel-local IDs are qualified by the extension and panel, so duplicated local IDs remain distinct. Header activation uses the owner-held session/workspace API; native focus is deferred to the originating window refresh. Active dock content participates in managed Tab order, and an open panel dialog isolates the whole window.

Tabs publish enabled header choices and the selected page. Closed or dismissed dialogs publish no modal content. An active enabled modal isolates the accessibility tree and managed keyboard order from its underlying window. Escape calls its controlled dismissal. A modal seeds an explicit first enabled own-scope native start-focus target before Run, so its first queued keyboard event does not depend on deferred Shown; nested modal Scenes establish their own focus, and an empty managed scope uses its body as the Escape handler anchor. Ordinary Tab/Shift-Tab follows descriptor order and wraps in the active scope; a tabs descriptor contributes one header target followed by its selected page. Custom extensions own their internal keyboard order.

When the visible accessibility scope changes, the originating refresh requests
one render of the main accessibility host. This restores current background
frames after a modal retires its native elements, including a label-only modal
closed through Escape or a controlled `Open` change. Unchanged scopes retain
demand painting. Removed elements and their cached frames remain retired; a
restored identifier receives a new native object and a completed geometry update.

Virtual tables and outlines publish a bounded materialized viewport, total/index metadata, stable domain IDs, selection, outline depth/expansion and previous/next page actions. Both table and outline pagers have real native button targets and completed geometry. Accessible page and row actions operate on the owner. Table paging clamps to the last valid viewport and refuses stale viewport counts or model errors. Outline press selects only; independent native expand/collapse actions preserve selection, reread current materialized child/expansion state and refuse leaves, offscreen rows or unadvertised actions. Cocoa uses `setAccessibilityExpanded:` and publishes `isAccessibilityExpanded`. A stale row action never selects a different domain row when a native slot is recycled. Offscreen rows are reached through paging; this implementation does not claim arbitrary offscreen row loading through the system accessibility table API.

## Managed grid publication

Each visible managed grid publishes a scope identified by the native attachment,
namespaced table ID, binding incarnation and table parent. Its ready transaction
contains the complete bounded semantic subtree and copied pixel frames from one
matching completed measurement. Every scoped node carries that same incarnation,
completed generation and target generation. Native validation rejects malformed
scope ancestry, mixed metadata and invalid frames before changing retained state.
Ready generation equals target generation; a layout-only transition may be pending
at the same portable generation. On the JSON wire, absent `grid_pending` means
false and absent `grid_generation` means zero, matching Go's zero-value omission;
present values must have their declared Boolean/unsigned integer types.
`InspectGrid` continues to refuse pending layout.

Owner publication of a new accepted projection replaces its action handlers and
advances the bridge semantic generation. Delivery guards also refuse model or
measurement drift before that publication. While its completed frame is pending,
the bridge retains only the last completed subtree for the same scope, incarnation
and parent. Those records remain readable with their previous values, selection,
counts and frames, but are disabled, unfocused and advertise no actions or managed
text client. Native selectors and request admission also refuse pending records.
Inspection explicitly reports pending, target/completed generation and incarnation;
last-completed records never count as current geometry. Initial pending layout
publishes only a disabled table host. An unchanged pending transaction does not
advance the global generation; a newer target supersedes it.

Other controls and other grids publish independently during a pending transition.
Actual scope removal, modal/tab isolation, replacement, namespace or parent change,
and window closure discard the old completed cache. Native objects are reused only
for a still-live identifier in the same scope/incarnation/parent; real retirement
cannot resurrect an object under the same textual ID. The owner cache contains
copied bounded records, no model pointers or action closures. Completed publication
restores current handlers and atomically updates native semantics and frames on
the main queue, reusing overlapping visible objects and retiring departing IDs.

Legacy asynchronous rectangle updates never change managed-grid frames. A changed
completed measurement, including scale, cell/editor bounds or same-generation
layout changes, invalidates the owner for a new atomic transaction. Frame ordinals
alone do not cause another publication. Detailed managed text-range geometry
retains its independent exact-client/text acceptance contract.

`AccessibilityElement.ObjectIdentity` is an opaque monotonic native identity for
acceptance inspection, not a pointer or action capability. The separate loaded-grid
`-accessibility-smoke` workflow uses stored input and at least 400 actually clipped
cells to check routed scroll reuse, pixel-to-screen frames, domain values/indexes,
binding replacement and a current native cell action. Its queries and diagnostics
are excluded from ordinary performance qualification. Fixture/race checks cover
the deterministic pending interval, stale queued actions, mixed scopes, malformed
transaction refusal and lifecycle retirement. Human screen-reader certification
and the existing P6 latency/memory targets remain separate gates.

## Geometry and notifications

A managed host publishes bounds after its children render with completed layout; the modal retained host publishes from its own Scene. Geometry publication copies widget state under the originating native render lock and never reenters the application owner or reads domain models. Completed geometry is cached independently of semantic publication so demand-rendered startup can publish bounds before matching native elements exist. A modal Scene shares the originating window render context, which excludes concurrent native reconciliation throughout its render pass. Cocoa converts pixel bounds through the current content view/backing scale/window screen transform, including clipped scrolled bounds. Native hit testing walks retained descendants before their parent and later siblings first, including descendants of zero-frame synthetic containers; equal-frame ties are deterministic. An active modal snapshot already excludes underlying content. Moving a native window keeps screen coordinates current. Value, tree-layout and focused-element notifications are posted after native snapshot publication.

## Verification and remaining gates

`go test -race` covers queue refusal, late-generation rejection, retirement, disabled ancestors, tab/dialog visibility, numeric validation, stable ID namespaces and bounded virtual rows. `cmd/accessibility-smoke` creates a real native window and queries its actual AppKit accessibility children, roles, labels, values, focus and geometry. It invokes native action/value/focus methods, verifies accepted controlled state, disabled refusal, keyboard traversal, modal trapping/dismissal, selected-page visibility and virtual paging. Tests never toggle VoiceOver, input sources or preferences. The live multiwindow checks reproduced races in backend screen polling and mutable screen-cache access; driver patch 0018 adds per-application poll admission, copied snapshots, synchronized accessors and locked logical-DPI updates. Fresh `-count=1` portable/native race tests, backend screen and current-window publication regressions, vet, two consecutive final AX smoke runs and AX-enabled integrated docking pass against the independent phase-two dependency through patch 0018. Logs record exact commands, source hashes, baseline and exit status. See the [native acceptance report](../docs/reports/accessibility/README.md) for commands, selector evidence and the integrated docking check.

These in-process native assertions are not a full screen-reader certification. Human VoiceOver navigation, announcements, macOS external AX client behavior, Windows UI Automation, Linux AT-SPI, detailed text ranges and all application-specific custom extensions remain explicit release gates. Clipboard, menus and file-panel selection are qualified by the separate platform-services phase.

Source basis: [Apple custom-control accessibility guide](https://developer.apple.com/library/archive/documentation/Accessibility/Conceptual/AccessibilityMacOSX/ImplementingAccessibilityforCustomControls.html) and [native expanded-state setter](https://developer.apple.com/documentation/appkit/nsaccessibilityprotocol/setaccessibilityexpanded%28_%3A%29), with the installed AppKit SDK headers checked for actual selectors and role/notification constants.

The owned native adapter snapshots the current-window global under its publication mutex. Screen targeting and temporary sizing read one stable window snapshot; raising a stage updates its own window title. Concurrent publication and screen targeting have a race regression in the prepared dependency.

## Native publication work and payload ownership

The unversioned native geometry channel omits the complete subtree of a bound
managed DataGrid: host, rows, cells, headers and private editor. Their frames
already commit with the validated semantic scope envelope; Cocoa continues to
refuse any legacy rectangle attempting to overwrite them. Full manual inspection
still exposes those bounds. Ordinary controls, virtual tables/trees, managed
trees and active workbench/dialog namespaces retain their legacy geometry.
Detailed text-range geometry remains a separate exact-client/text channel and
includes the private grid editor.

Snapshot validation returns its checked ID-to-node map for the native commit.
Identity reuse reads previous scope ownership from the old element dictionary
before any element is mutated. Whole-snapshot validation, monotonic epochs,
allocation before mutation, hierarchy order and notifications retain their
existing rules; this does not introduce a delta or partial commit protocol.

Asynchronous JSON producers copy C bytes and enqueue their strongly capturing
blocks inside an explicit producer autorelease pool. Consumer blocks establish
separate pools for native parsing, queries and actions. String-based action and
text-query producers also copy inside their own pools. No C pointer survives a
return to Go. Owned transport payloads remain alive while queued and can be
released after consumption or discard without an ambient producer pool. This
lifetime rule does not bound the total queued work or prove process/native/GPU
memory reclamation; those remain separate acceptance measurements.

The ordinary owner refresh consumes the fresh node and callback containers
produced by its semantic traversal. A ready publication may reuse those
containers, but the last-completed grid cache still detaches records, actions and
pointer fields. Borrowed internal update/prepare entries retain separate output
and callback containers. Pending replacement keeps the original scope/incarnation,
parent, disabled-state and stale-action rules, including independent ready grids
and ordinary controls. Removal and closure still retire cache references.

Scope grouping records input indices rather than growing copied node arrays.
After successful whole-node JSON encoding (and a second encoding when pending
replacement changes nodes), the native envelope surrounds those exact encoded
bytes with fixed delimiters and a generated decimal generation. Application text
never enters the envelope unencoded. This avoids another RawMessage encoding
without changing complete wire records, ordering, semantic-generation equality,
callback replacement or native validation/commit/notification behavior. Ownership
transfer and fewer intermediate copies do not qualify total retained, native or
GPU memory, or sustained latency by themselves.

## Native integer validation

Unsigned generation/incarnation fields reject booleans, negative values,
fractions and values above UINT64_MAX. Positive fields also reject zero. Exact
integral Foundation NSNumber representations may be checked through their signed
or unsigned integer accessor; floating, decimal and unknown representations retain
the strict digit-string validation, preserving full-width boundaries. Complete
snapshot validation still precedes retained-tree mutation.

The [native validation checkpoint](../docs/reports/macos-native-validation/README.md)
records original/candidate boundary parity, native 400-cell acceptance and a
451-node validation-only benchmark. Sampled process/native stacks and this local
benchmark remain separate from recorded UI-work latency, P6/P11 and human Mac
accessibility qualification.


## Managed outline publication

Managed `ui.DataTree` uses the same bounded pending/completed transaction as
managed tables. The existing Grid-prefixed JSON and inspection field names apply
to both; scoped native hosts are restricted to table/outline roles. Each completed
row/page frame is published with its matching semantics, incarnation and target/
completed generation. Unversioned geometry callbacks cannot overwrite it. Pending
layout preserves last-readable rows/frames and compatible Cocoa object identities,
while actions, enabled/focused state and queued handlers are revoked. Initial
pending layout contains only its disabled host. Completion restores the current
snapshot atomically; removal, binding replacement, parent and scope-kind changes
retire old objects. Independent ordinary controls and view scopes remain current.

The tree adapter fences same-generation scene layout and pending-index metadata,
and checks readiness again at owner action delivery. Native selectors also reject
pending records. Retained table/tree scene references and geometry are withdrawn
before modal transfers, keyed ancestor moves and workbench parking. This protects
views moved by an ancestor rather than remounted individually. A pending index
worker retains its old accepted index's ordinary interaction once layout is ready.
The [checkpoint](../docs/reports/macos-tree-publication/README.md) separates manual
and inert ownership oracles, actual Cocoa-object transactions and native workflows
from human VoiceOver, physical input and full Mac qualification.
