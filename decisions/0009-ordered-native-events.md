# ADR 0009: Reserve native work at its toolkit FIFO position

Date: 10 October 2026. Status: accepted for experimental integration; production G1 remains open.

The native mailbox preserved its own FIFO but its rearmed wake moved behind already queued legacy shortcuts. Retain bounded payload admission and batches of at most 16 while replacing the shared tail wake in the integrated fixture with deferred FIFO metadata nodes reserved atomically alongside mailbox credits.

Use a proposed opt-in Cogent deque extension that waits for readiness, merges adjacent committed native nodes only within the configured batch limit, and removes canceled/closed-owner nodes. Ordinary events separate native groups, preserving both earlier-update-before-key and key-before-later-update order. Existing priority window events retain precedence. Producers must resolve reservations; an abandoned reservation can stall ordinary FIFO events.

This adds bounded per-admission metadata and a condition-variable wait rather than increasing native work per dispatch. Payload credit remains held through application. Caret publication remains coalesced separately. Close removes this owner's queued nodes and retains committed mailbox deliveries for a bounded quiescent-owner shutdown drain.

The fixed audit observes 32 earlier updates at the shortcut; its legacy control still observes only 16. Boundary/race/native-lifecycle evidence is in [the fix report](../INPUT_ORDERING_FIX.md). This closes that demonstrated defect, subject to upstream extension acceptance, while retaining the other G1 gates and explicit limits on fairness, performance and native shadow-state synchronization.
