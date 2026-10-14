# Composed native lifecycle foundation

The all-component workspace exposed two ownership bugs. Docked panels keep
managed records outside descriptor Children; normal teardown did not retire those
editor/grid/tree bindings, retaining closed widget graphs. Disposal now walks
nested private panel records. Failed rebuilds retire unadopted parked records
before deleting their widgets. A retired dock revokes queued interactions and
refuses later admission. An explicit attached-host retry advances through fresh
native state without reviving old queued requests.

Application.CloseWindow also left a removed Window.View/CloseRequest closure in
the backing slice. slices.Delete clears that slot under the existing mutex,
preserving surviving order, ID tombstones, lifetime cancellation and cleanup.
A weak-reference regression verifies a captured Go object can be collected while
the survivor/application remain live. Original code overlays fail the view-capture
and failed dock rebuild regressions; final focused race checks pass. The new test
has its own filename and preserves the preexisting 1000-context retention tests.

The native window-cycle command adds -composed. Every fresh window owns the full
UI kind catalog, an editor/journal, managed grid and indexed outline, docked shared-
device GPU canvas, modal dialog and 1000 stable dense labels. It checks GPU readback,
owner editor history replay, completed grid/tree geometry and an actual Cocoa
controlled text request observed in the domain model. Modal target publication
and retirement are checked separately from window/resource cleanup. Queued owner
probes are revoked on timeout, and executing probes join before resource closure.
The mode permits a bounded45minute deadline; the original fixture remains20minutes.

A four-cycle diagnostic before the fixes shows post-GC heap28.7→183.4MiB,
approximately39MiB retained per cycle. After nested disposal, an intermediate
20-cycle run reports72.4MiB atcycle5 and77.3MiB atcycles15/20. After clearing the
window backing slot, the selected final five-cycle preflight reports28.7MiB
initially and40.9MiB aftercycle5, with175.8→320.5MiB RSS. These builds/layouts differ;
the observations support ownership corrections, not isolated numeric attribution
or a long-run memory qualification. The complete raw samples and limits are in
[review summary](records/review-summary.json).

All five final cycles report1000 dense label frames fully contained in both the
root and scroll viewport. This is actual native AX geometry, not pixel occlusion,
physical display scale or requested1920×1080 geometry certification. Earlier
fixture revisions had a zero-width scroll viewport; Grow plus explicit height
bounds correct it. The fixture waits for positive geometry after modal closure,
because scope publication can precede paint. The software geometry regression
uses actual standard controls/labels but substitutes native Extension allocation
placeholders, since Manual deliberately refuses custom extensions. It qualifies
only that layout portion; native composition is checked separately.

The full offscreen adapter race suite and vet pass. Portable race packages pass
with the automation Unix-socket package rerun under required local permissions;
the raw sandbox-denied run remains. The restored original retention tests plus
the new regression pass together. Internal review found and corrected failed-
rebuild and stale-interaction gaps. This checkpoint requires one bounded Prism
review with Anthropic Claude Sonnet5.5, with exact source/receipt evidence archived.

This advances composed fixture implementation and short native acceptance.
The1000-cycle workload, sustained/pacing/input/frame and native/GPU budgets,
overnight soak, required physical display variants, human accessibility and
Mac distribution gates remain open. Original4/6ms frame targets still fail.
