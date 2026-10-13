# Mac acceptance after GPU ownership fixes

The current bounded live checkpoint uses commit
`1d91d848ac973e2b264c24e61c4a3beac34f7f08`, including owned native patches 1–52.
The [report](../docs/reports/macos-post-ownership/README.md) preserves actual
receipts and independent reviews. It advances the
[window lifecycle](native-window-lifecycle.md) and
[native object lifetime](native-object-lifetime.md) checks without completing
the broader [Mac milestone](macos-completion.md).

The existing native window-cycle consumer completes 1,000 fresh windows while
retaining one survivor. Readback, canceled contexts, rejected late delivery,
retirement and final tracked inventories pass. Forty external vmmap observations
at cycles 0..975 record physical footprint within 549.4–693.1 MiB. Final native
cleanup follows cycle 1,000; the final physical-footprint sample is deliberately
omitted because the process can exit before external inspection. RSS and Go heap
remain separately reported and can grow as caches/tombstones are retained.

This is one sampled, perturbed fixture. Its 4 GiB observed-current-footprint
safety guard does not bound unsampled peaks or qualify all native/GPU allocation
lifetimes, physical GPU memory, the integrated P11 workload or overnight endurance.
Historical results from a different build do not isolate a single patch's effect.

The [stored-table protocol](grid-performance.md) is rerun for both the default
scene renderer and the optional GPU scene renderer. Each mode uses four cases,
five fresh release processes per case, 128 guarded warmups and 120 measured
events. Both populations fail the unchanged p95 ≤4 ms / p99 ≤6 ms UI-work targets.
Modes run sequentially in separate campaigns; this is not an interleaved renderer
comparison. Background indexing/load observations remain part of the evidence.
Four additional GPU source-detail processes establish the tested deferred scene
route and remain excluded from ordinary latency results.

The implementation, performance and release requirements remain active. Human
IME/VoiceOver, cross-display behavior, broader native services, physical GPU loss,
sustained targets, integrated endurance and distribution qualification need their
own evidence.
