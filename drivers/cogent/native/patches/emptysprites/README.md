# Empty sprite layers

Stage 35 follows optional native frame-work tracing. Preparation verifies exact
input, patch and output bytes. `SpritesSource` checks all three sprite tiers
under the existing lock before allocating a painter/renderer or clearing an
image. With no nonnil active drawable sprite, it settles the modification and
returns no source. Fresh composition replaces the previous source/operation
list, so removing that layer restores underlying content without uploading a
transparent scene-sized image.

Active callbacks still run on every composition in First/Normal/Final order,
even when the modified bit is false or a blink callback emits no paint. Active
event-only sprites retain their existing deadline and input semantics. The
related `HasActive` scan skips nil entries, matching the source's existing nil
handling; its Active-only deadline criterion is unchanged.

Previously allocated sprite resources remain available for reuse. Empty resize
does not allocate replacement sprite resources; reactivation creates size-correct
resources through the existing path. This does not change the early cached
resize replay policy, introduce partial uploads, release inactive GPU textures
immediately or alter native event/paint scheduling.

Differential software regressions compare the exact frozen original
transparent-clear path with removal, movement, reactivation and resize. They
check allocation refusal, composition copy requests, tier order, cached operation
replay and an explicitly advanced real TextCursor blink callback. Managed
integration and actual release measurements are separate qualifications; a
software copy count is not a measured GPU upload count or display latency.
