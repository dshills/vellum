# Native accessibility teardown retention

The full composed lifecycle campaign on the previous checkpoint stopped at cycle
600: sampled RSS reached 1,119,485,952 bytes and exceeded the unchanged 1 GiB
safety guard, while post-GC Go heap was 44,625,472 bytes. All 600 completed cycles
passed their interaction, geometry and logical cleanup checks. Native snapshots
showed MALLOC_SMALL resident growth from 218.5 MiB near cycle 202 to 505.9 MiB near
cycle 504. This was a failed campaign; its raw result and profile are preserved.

The native accessibility bridge's attach/detach entries lacked local autorelease
pools. Go-owned main callbacks execute outside GLFW's event-wait call. Detach's
Foundation allValues temporary could therefore retain each retiring element and
its node payload after the state registry and view association were removed.
A temporary native diagnostic counted 66,631 created objects across 25 windows.
Before survivor teardown, the original path had freed 33,350; adding a local
teardown pool freed 66,625, leaving the six live survivor elements. State objects
were already being freed. This isolates accessibility object lifetime separately
from Go retention and native window inventory.

Production attach and detach now own local pools. The live registry and view
association retain attached state after attach returns. Detach drains its array
only after retiring descendants, clearing geometry and removing association and
registry ownership; main-thread ordering and subclass restoration are preserved.

An inert ARC regression exercises the actual production functions for 100 cycles
with 1000 nodes each. Weak element/state references must expire before its outer
pool drains, registry membership must disappear, and the original view class must
be restored. The original code fails at cycle one; the corrected source passes.
The previous layer/input ownership fixtures and native accessibility race tests
also pass. The fixture creates no NSWindow and changes no macOS settings.

The corrected full campaign passes all 1000 cycles and final main-loop cleanup,
with no visible test windows, native service operations, accessibility windows or
pending callbacks. Every cycle reports all 1000 label frames contained, accepted
native text, editor history replay, completed grid/tree geometry and modal closure.
Across 41 post-GC samples, warm live Go heap is 45.3–46.6 MiB and RSS is
350.4–369.1 MiB; final RSS is 369.1 MiB. Native snapshots at cycles 164 and 636 show
MALLOC_SMALL resident 26.8→29.4 MiB, total malloc allocation 31.5→32.8 MiB and
physical footprint 653.9→662.4 MiB. IOSurface resident remains 72.1 MiB; owned
unmapped graphics remains about 326–327 MiB. These are observed process regions,
not physical GPU deallocation or a newly approved graphics budget. The complete
campaigns are in [review summary](records/review-summary.json) and the raw archive.

The corrected process exits zero, but the raw wrapper exits one: its git helper
strips the first porcelain line's leading space, and adding documentation changed
which line sorted first. Only that cosmetic status string differs. A bounded
endpoint check confirms the commit and all 739 selected code and 6204 prepared
hashes are identical. The raw wrapper failure and explicit resolution remain;
there is no substituted rerun or overwritten success receipt.
Geometry containment means measured native AX frames lie within the root/scroll
viewport; it does not prove physical screen occlusion, requested geometry or a
primary display scale. The campaign is a lifecycle/retention observation, not a
latency measurement. Native graphics regions and physical GPU attribution, full
P6/P11, overnight soak, human accessibility and distribution remain separate gates.

The temporary diagnostic compiler wrapper substitutes only the selected native
source during compilation. Its first Go .m overlay build did not include the
instrumentation; that run is retained but not used for object-count attribution.
The corrected build explicitly observed substitution. Neither diagnostic source
nor wrapper is part of the production adapter.
