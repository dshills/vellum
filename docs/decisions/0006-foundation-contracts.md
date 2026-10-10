# ADR 0006: exact application text and reserved transport credits

Date: 9 October 2026
Status: Provisional; local contracts demonstrated; G1 open

## Context

[ADR 0005](0005-foundation-checks.md) confirmed altered trailing newlines and unbounded native event admission. The follow-up implements an isolated exact-text buffer and independent bounded mailbox; [results](../reports/FOUNDATION_CONTRACTS_RESULTS.md) record scope and evidence.

## Decision

Add an explicit lifetime policy through `NewLinesExact`, retaining final empty lines and exporting current lines without synthetic newlines. Preserve the existing buffer API's legacy behavior. Exact diff application uses a selected-operation line snapshot and grouped full replacement to represent final empty-line changes; this trades incremental formatting/caret preservation for correctness and remains a prototype. Temporary comparison buffers retain the same policy. Keep the patch separate for upstream review; do not treat invalid UTF-8 or formatter changes as exact-byte preservation.

Use reservation-before-mutation as the transport ownership hypothesis. Account both slots and payload bytes through application acknowledgement, maintain FIFO admission order, and retain accepted deliveries during close. Nonblocking native preflight must visibly refuse/pause before acceptance; worker-only paths may wait with context cancellation.

Do not swap this mailbox into the current native callback: that callback follows native state mutation. Integrating only an enqueue limit there could silently lose accepted text. Native preflight, owner-serialized teardown and bounded/rearmed event-pump wakeups remain prerequisites. Coalescible progress/caret state needs a separate bounded contract.

## Consequences

Representative exact-text/edit/history tests and the native fixture corpus now pass. A standalone mailbox delivers 100,000 accepted messages under fixed queue bounds and concurrent race tests pass. Native transport is still unbounded in the prior fixture and font-memory/lifetime/accessibility/platform/distribution gates remain open. D stays experimental, A control, B fallback. No production API or workbench expansion is accepted.
