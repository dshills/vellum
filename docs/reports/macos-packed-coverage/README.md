# Packed glyph coverage checkpoint

Cached glyph coverage can now enter the scene builder as one owned span
allocation. The previous route grew individual rows and then copied them into
the packet. The new path keeps the same packet representation, logical charges,
operation boundaries and existing bounded coverage cache. CPU image rendering
continues to use its existing span route.

The fast path admits at most 256 sorted samples into an empty active stream.
It checks the original geometric construction-header and span charges before
allocating. Clipping, unsupported inputs or insufficient remaining budget use
the original sequential path, preserving partial refusal. Flush transfers the
owned spans and detaches row headers. Packed subrows never enter the separate
row scratch cache; ordinary appends validate and reserve before detaching them.
Success and refusal retire all construction state.

The actual retained 400-cell software capture measures 14,488 versus 10,937
allocations per repeated ScenePacket. Five interleaved pairs record
about 3.97 MB versus 3.91 MB allocated per capture. All ten retain identical
encoded bytes, accounting and software pixels. These are software allocation
observations on a shared computer. Local elapsed times are diagnostic and do
not establish ordinary native frame latency, GPU time or retained memory.

Checkpoint acceptance additionally requires the final source-bound five-pair
comparison, 24 observations per original/candidate matrix arm across four DPIs,
fixed/variable rows and three scroll steps, full renderer and adapter race/vet,
meaningful failure controls, pinned preparation, Mac compilation and private
SDK relocation. Independent pre-reuse builder tests check exact state, peak
budget and operation refusal, original/GPU encodings, later ordinary spans,
input isolation and retirement. Test-authoring failures remain excluded records.
Known upstream raster tests produce 17 non-source PNGs; exact hashes and cleanup
are recorded separately, and they must be absent from SDK inputs.

Internal review and one Prism review with Anthropic Claude Sonnet 5.5 are
required before commit; their raw results, resolutions and exact source/command
identities belong in the checkpoint evidence. No native executable runs while
the human is online. This does not close P4/P6/P11, native/GPU memory, physical
input/accessibility, distribution or the full Mac framework. The historical
4 ms p95 / 6 ms p99 native targets remain failed; a quiet native comparison
is still required.

Explicit vet caught two unkeyed literals when the sample type moved between
packages. Keyed fields correct those declarations without changing their values;
the failed vet and superseded SDK evidence remain separate from final checks.
