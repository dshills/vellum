# Desktop pointer modes source checkpoint

Window-bound Hide/Show, relative/raw-mode requests and SetMousePos now use the
native owner and existing main dispatcher. Status reports observed GLFW logical
mode, with raw capability fallback and dependent-write readback checks. Disabled
movement follows GLFW virtual coordinates without a second per-event warp.
See [the contract](../../../specs/desktop-pointer-modes.md).

The offscreen checks passed fresh preparation of both pinned dependencies,
23 new top-level pointer regression methods, existing cursor/queue/event controls,
the 390-case cursor artwork matrix, colors under race and Go vet. Automated
routing/lifecycle evidence uses fake native backends, never GLFW initialization
or a native window. Tests include callback retirement, deferred final restoration,
raw ordering/refusal, exact-window isolation, queue supersession, checked negative
and overflowing coordinates/deltas, DPR/focus publication during readback, first
drag/release/drop baselines, nested-request refusal and complete panic cleanup.
Exact source/check receipts are in [evidence.json](evidence.json); SDK build and
unsigned packaging receipts are separate external checkpoint artifacts.

Internal review found and corrected retirement/reentry, lost DPI baseline,
no-op baseline, button/drop/scroll routing and nested-refusal status gaps before
the final passing run. Earlier attempts remain unchanged in the external evidence
directory: v1 exposed a fixture cleanup closure referring to retired cursor.window;
v2 passed but review found additional interleavings; v3 was explicitly stopped
for a test type correction; v4 passed focused controls and was explicitly stopped
for the restoration ordering correction. v5 passed, then Prism disposition
identified a failed-readback baseline gap; v6 repaired it and passed the full
86-method selected race run including the raster matrix. V7 strengthened only
the unknown-mode test to inject failure after a real setter, then freshly
prepared both dependencies and passed all 23 pointer controls under race plus
vet. Production source is byte-identical between v6 and v7; both receipts are
retained and no earlier test is relabeled.

No physical pointer, Cocoa visibility/capture, actual Retina/focus behavior,
allocation/endurance or measured 60 Hz result is claimed. Native presentation,
M2/M4/M5 checks, final distribution/SDK approval and final Mac audit remain open.
4 ms remains an optimization target; this checkpoint does not complete Mac.
