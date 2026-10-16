# Native professional workspace smoke runner

This diagnostic consumes the existing public `workspace` fixture and Vellum APIs.
It creates two real native windows and issues 18 Cocoa accessibility actions:
load stored data, edit a property and Undo/Redo, activate palette placement and
Undo/Redo the same object, sort/filter, rebuild the outline, select a measured
table cell, open/close a modal, switch document tabs and close the main window
while preserving the shared-model inspector. The final application close joins
its worker and reports tracked native service cleanup.

After preparing the pinned dependencies, build an instrumentation-only binary:

```sh
go build -tags vellum_native_checks -o build/workspace-smoke ./cmd/workspace-smoke
./build/workspace-smoke -rows 10000 -timeout 3m
./build/workspace-smoke -rows 1000000 -gpu-scene -timeout 3m
```

Run native cases serially on an otherwise available desktop. Each native action
is proposed at most once; a refusal or uncertain outcome fails conservatively.
The action helper owns a cooperative 30-second cap and preserves earlier caller
deadlines. Its native/owner hooks must honor cancellation. Completion handoff
waits until the scenario has returned from application closure.
The independent streaming data oracle checks the retained source hash, maximum
sort value and filtered row count. Table/outline inspections require completed
geometry. Native actions and owner completion have separate recorded fields.
The process disables only its own geometry persistence and restores that flag
on ordinary return; it does not change OS preferences or access the clipboard.

`-gpu-scene` records a requested optional scene renderer, which can fall back.
The process-wide counter reports framework surface-present attempts, not actual
display presentations or successful GPU paint coverage. There are no timing or
memory acceptance thresholds here. Palette activation is the accessible command
path; it does not verify physical dragging or keyboard delivery. This scenario
does not execute the separate 216-operation software record, certify all P9/P11
behavior, establish pixel identity or qualify human IME/VoiceOver/displays.

Implementation, offscreen controls and compilation are separate from actual
native execution. Do not claim a native pass until the exact built source has
completed and its external process/cleanup evidence has been reviewed.

Table selection uses a distinct cell with positive clipped measured bounds.
Retained overscan cells with empty bounds are excluded; partially visible cells
remain eligible. The table step waits for that geometry and native publication
before the single selection proposal.
