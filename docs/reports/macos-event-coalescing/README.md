# Event queue coalescing source checkpoint

FIFO queue compression now checks both scroll axes before modifying the incoming
event, preserves unsafe samples in order, and keeps raw/local previous positions
consistent with the newest local offset. Both Front and Back preserve Unique and
Deferred boundaries. Front keeps its existing latest-notification policy without
new pointer aggregation. See [the contract](../../../specs/event-coalescing.md).

Fresh pinned dependency preparation and race checks passed: 14 new focused queue
methods, 24 full events methods, 72 selected desktop/base methods, and five
offscreen Core routing/layout methods. The counts overlap and are not added into
a distinct-test total. The original-code overlay failed exactly three controls
for finite scroll overflow, local-coordinate consistency and Unique preservation.
Default vet passed the changed events and desktop/base packages. Broader default
Core vet failed with twelve diagnostics; a separate run on the committed baseline
reported identical diagnostics in five byte-identical files. Full Core vet remains
failed, and final source quality cleanup remains open. No analyzer was disabled.
Exact counts, limitations and receipt bindings are in [evidence.json](evidence.json).

Earlier attempts remain unchanged in the external artifact directory. V1 exposed
a missing test-constructor argument; v2 exposed a test expectation that incorrectly
required a MouseMove to carry a drag anchor. Both were fixture corrections. V3
passed all selected race controls and then failed broader Core vet as described
above. The scoped proof preserves that controller failure and separately records
the passing changed-package vet and baseline comparison. No failed run is relabeled.

Private SDK native compilation, relocation and unsigned/unlaunched app packaging
have separate external checkpoint receipts. These are source, build and offscreen
evidence. No physical input, Cocoa capture/visibility, OS services/accessibility,
presentation, allocation/idle/endurance or measured responsiveness is qualified.
The Mac release gate remains measured 60 Hz; 4 ms is an optimization target.
Remaining M2–M7 acceptance and distribution gates stay open.
