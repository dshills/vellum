# Retained collapsible sections

The public `ui.Disclosure` container adds collapsible inspector/settings content.
Its heading uses ordinary Text and controlled Checked/Change expansion. A private
heading and body retain keyed native children while collapsed, preserving private
field drafts. The gallery and an external SDK consumer demonstrate the API. See
the [control catalog](../../controls.md) and [contract](../../../specs/controls.md).

## Behavior and correctness

Pointer and Enter/Space toggle the heading; Left collapses and Right expands.
Accepted state controls visible layout. Read-only headings remain focusable.
Collapsed children leave focus/modal discovery, native accessibility and managed
grid/tree input scopes. Programmatic collapse transfers child focus to its enabled
heading. Queued requests use the latest callback and are revoked after accepted
state changes, disable or retirement. Descendant actions are refused after collapse.

The separate heading target has measured geometry and a reserved `disclosure:`
identity in Manual routing. Native accessibility publishes a group and an
AXDisclosureTriangle with Expanded and checked-value metadata. Its heading object
survives expansion changes; removed child semantic objects may receive new AX
identities after reopening. Native widget identities remain retained. Boolean
value changes use the existing native value-change notification path, including
empty-content sections.

Portable, public SDK and complete offscreen adapter race suites pass, as does
adapter vet. Focused routing/lifecycle/role checks pass three times. Deterministic
software frames verify collapsed geometry, completed raster differences, pointer
hits through the inspected heading, keyboard order/focus fallback, retained drafts,
and preserved grid/tree bindings with revoked hidden input scopes.

## Hidden-grid regression

The new collapse test reproduced an existing Manual capture problem: it demanded
completed geometry from a hidden grid, repeatedly reporting LayoutPending. Grid
capture now requires semantic visibility, matching tree capture; the hidden
descriptor record remains inspectable while its materialization is omitted.
A control overlay restores only the previous unconditional grid capture and
fails the final bounded settle test again. The selected implementation passes.
The original adapter failure and intermediate focused failures are preserved.

An early native role test also expected Expanded from a fixture that did not
report it. The fixture now exposes the production Expanded selector. The first
SDK run failed because the sandbox denied its local Unix socket; the authorized
rerun and final complete SDK suite pass. These are distinct from production
control correctness and do not erase the failed receipts.

## Actual Cocoa checks

One optimized native gallery process completes 20 collapse/expand cycles through
40 actual Cocoa selector actions. All 40 native snapshots show the accepted
Expanded state, a positive heading frame, stable heading object identity and
matching heading hit tests. The field disappears from the collapsed hierarchy;
stale native field actions are refused and reopening preserves its controlled
text. The bounded worker joins and the native process exits successfully.

The fixture changes no input sources, VoiceOver, clipboard or persistent OS
preferences. Window-geometry persistence is disabled in the fixture. This is
native selector evidence, not physical keyboard/pointer interaction, human
VoiceOver, native memory reclamation or sustained latency/endurance qualification.
The previous table campaign's failed 4/6 ms targets remain unchanged.

## Review and evidence

Internal source review approves the 23 code/spec/catalog leaves. One bounded
Prism review with Anthropic Claude Sonnet 5.5 covers the final source and report
checkpoint. The [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json)
retain selected source/prepared identities, receipts, failed/corrected checks,
the counterfactual overlay, native snapshots and reviews. Prepared dependencies
remain byte-identical at 6,198 files; these records are not a full compiler,
SDK or cache closure. Full Mac performance/P11, human/physical workflows and
distribution/release gates remain open.
