# Stage 0103: completion-triggered busy dirty rescan

A paint admission that finds a composition active records one atomic, payload-free pending-rescan bit and retains the existing 16.7ms timer fallback. Both ordinary and measured composition completion first decrement the active counter, then consume the bit and request an immediate rescan only for a live window. Idle rescan admission and window closure clear the bit. Completion with no busy request remains idle.

The rescan checks actual dirty state; it does not force a frame. Existing resize/cache fallback flags keep their semantics. No input/paint queue priority, callback route, native ownership, frame accounting, GPU policy102 or lifetime ownership changes. The timer covers completion winning before pending publication; a redundant fallback wake can check unchanged state but must not add a composition.

Seven real software/manual-compositor regression methods cover duplicate busy coalescing, ordinary/measured immediate readiness at an unchanged clock, no-busy/no-wake, queued dirty pixels exactly once, completion during fallback publication, 16 concurrent admission/completion orderings, resize forced-frame preservation and actual backend closure. Existing manual/ownership and stage100 tests also pass with race detection and vet. Old source102 with the final tests fails the immediate-readiness assertion for both ordinary and measured paths, not compilation.

The controlled tests prove removal of the busy timer wait in the covered ordering. They do not establish this path as the cause of observed native latency. Source102 remains an unqualified experiment; its FAILED journal is untouched. Fresh source-bound SDK/native comparison and unchanged M3/lifecycle gates are required after integration.
