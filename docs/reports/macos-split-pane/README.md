# Controlled split panes

Vellum provides `ui.SplitPane`: two retained child panes, side-by-side by default
or top/bottom with Vertical Orientation. Number is the first pane's accepted
fraction; omitted Numeric means 0.05..0.95 with a 0.01 step. Explicit finite bounds
remain strictly inside0..1. Nested nodes form larger adjustable layouts.

Native pointer moves and main-axis arrows/Home/End request owner Change values.
Accepted state remains authoritative without optimistic native collapse. Current
bounds/callbacks apply on delivery; removal, disabled ancestry, pane-order or
orientation changes refuse stale requests. Read-only dividers remain focusable.
Tab visits first-pane targets, divider, then second-pane targets. Private pane
hosts retain child editing/scroll state and isolate native styling from children.
The gallery has a Split panes page; the [catalog](../../controls.md) and
[contract](../../../specs/controls.md) describe the API.

Mac semantics expose AXSplitGroup and a qualified AXSplitter with value, bounds,
actions, focus, measured geometry and divider orientation. Generated split: targets
cannot collide with descriptor IDs. Native Parts keeps its handle-0 Maker key;
renaming that key recreates the widget and loses its retained input/focus binding.
The managed renderer removes native implicit collapse below1%. Layout-only
one-pixel rounding preserves tiny positive panes when divider/gap plus two pixels
fit, then restores accepted TileSplits. Explicit minimums can still overflow.
Normal-phase digit suppression preserves registered scene command priority while
refusing native collapse/equalization outside owner control.

## Verification and corrections

Portable race, complete offscreen adapter race (including existing Manual/replay
consumers), focused split/gallery race3, adapter vet and native gallery compilation
pass against all twenty final-v3 leaves and unchanged6,198 prepared files.
The gallery was compiled only. The initial broad run's local Unix-socket bind was
refused by the sandbox; its raw exit1 is preserved. A permitted local-socket rerun
passes the same full suite without source changes. A supplemental transactional
regression and internal review identified a distinct accepted-ID rollback defect:
preparing a failed child mount changed the retained queue binding. Separate planned
IDs now route preparation; accepted IDs publish only after a complete successful
plan. The strengthened test requires accepted metadata and previously queued valid
resize delivery to survive refusal. The exact old two-production-leaf overlay fails
this same test. All five final-v3 checks run after that three-leaf correction;
seventeen other leaves match the provider baseline. No repeated provider run was
needed for this localized internal correction.

Actual Manual tests cover both orientations, pointer resizing, keyboard bounds,
measured pane ratio/divider containment, completed pixels, retained draft text and
selection, visual Tab order, identity through orientation changes, read-only and
disabled-ancestor behavior, registered digit shortcuts from a pane and divider,
and tiny positive/subpixel fractions. Direct retained fixtures test latest callback
and bounds, refusal, pane reorder, queued disabled/retired delivery, qualified
focus targets and numeric semantic actions. Inert Cocoa objects check roles and
orientation without a window. A start-focus request in an inert fixture schedules
future focus; actual shown focus is proved by Manual routing, not that inert call.
Caret geometry/visual layout can change when panes resize; logical draft/selection
comparison deliberately excludes those coordinates.

Internal review found native Maker-key retirement, implicit sub1% collapse and
digit command interception. All are corrected. Later boundary checks found
subpixel zero-width rounding and Manual reapplying physical window size as a
body-content minimum. Removing that software-only minimum lets the body grow into
the actual padded/bordered scene allocation. Existing Manual/replay consumers
pass the complete race suite after this explicit geometry correction. Live Driver
sizing was not changed or qualified.

Five genuine negative overlays restore each bad behavior independently against
the final-v1 fixtures: renamed handle loses focus; native renderer hides all four
small-fraction cases; First-phase guard swallows registered commands; old Manual
minimum clips the far pane; missing pixel rounding gives zero-area subpixel panes.
Each returns exit1 with a real assertion; final focused race3 passes unchanged
corresponding fixtures. Exact overlay payloads/inputs and source families remain
in the archive. Initial color/unused-import compiler errors, lifecycle/geometry
failures and weak fixture attempts are historical evidence, not qualification.
Multiple Manual instances originally shared one cleanup lifetime; subtests now
own each independent instance. Exact original/v2/final and selected intermediate
source payloads are retained; historical maps explicitly list remaining gaps.
30 receipts/60 logs preserve 16 raw exit1 attempts.
Binaries, caches and complete compiler/SDK closure are excluded.

## Review scope and open Mac gates

Internal source review approves the exact final twenty leaves. Prism runs once
using Anthropic Claude Sonnet5.5 with complete twenty-file baseline coverage,
4 chunks, 45120 input/5401 output tokens:
raw exit1, one medium confidence.35 concern about unchecked pane index/nil widget
at finishControl. The complete transaction validates exactly two children before
allocation, refuses nil mounts, and finishes allocation before attaching/order.
Internal review rejects the missing-invariant concern as unreachable through the
supported validated API. A supplemental regression proves three-child and nil-mount
refusal preserve structure; its review found the distinct accepted-ID rollback edge
above. The corrected production publication and strengthened queued-delivery test
are internally reviewed and pass full final qualification. The provider report
remains raw1; Prism was not repeated for that localized three-leaf correction.
Baseline provider coverage, final internal delta review and execution remain distinct.

Separate evidence review binds the [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json).
No desktop windows/input/settings or general clipboard writes occurred. Physical
Mac input/VoiceOver, native/GPU memory, sustained performance/endurance,
distribution rights/signing/release and full Mac framework qualification remain
open.
