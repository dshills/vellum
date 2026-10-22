# Native raster root structure: captured result

The approved one-attempt capture completed in **4.782 seconds**, including probes.
It retained 11 renderer calls and 162 scalar command records: eight startup calls
and three admitted calls at 1519 × 1266. Admission uses the explicit one-shot
marker, rather than dimensions: five startup calls also have that extent but
precede the marker. All three admitted calls used an owned
scanner and completed with the same root decisions:

| Decision | Per admitted call | Total admitted |
| --- | ---: | ---: |
| Covered by next command | 4 | 12 |
| Uniform crop branch | 3 | 9 |
| RenderPath | 4 | 12 |
| Params.Render | 1 | 3 |

The single image-render call covers (8,8)–(1511,1258). Each of the four paths has
20 path values, an identity transform, neither fill nor stroke enabled, and meets
the existing crop no-paint predicate. Scalar patch colors proceed red/green/red.
Existing coverage and crop optimization is already active; the earlier eight
recorded image commands are **not eight executed full-screen fills**. Cropped
branches can write zero or several bands, so these counts do not measure bytes
written or CPU cost.

Three routed public model transitions join unique positive native presentations.
Startup calls, opaque tags and actual timestamps remain in the raw evidence; no
latency distribution is estimated from these three transitions. Geometry admission
does not attribute commands to input tags. Scalar colors are not GPU readback or
physical displayed-pixel proof. Physical pointer/keyboard testing was not done.

Independent replay verified all nine owned process groups retired, complete log
hashes, final v2 build/source/module/overlay pins, unchanged configuration and
geometry, all 8,321 frozen SDK files, and the original Metal archive. Process
retirement is not a whole-GPU allocation balance proof. Preparation passed five
focused observer methods, the original routed full-pixel/state fixture and three
classifier methods with eighteen rejection cases. These checks were reused;
no broad campaign or unchanged native retry was run.

The [protocol](../../../tools/macos-calibration/rastercommands/NATIVE.md),
[renderer diff](renderer.patch), [machine-readable result](result-summary.json),
and indexed preparation/result evidence archives retain the basis for this result.
Preparation source checkpoint: `2330a9b`; preparation archive checkpoint: `fbedfc8`.
The initial pre-admission-marker binary/test receipts are retained but were never
run natively; the final v2 build supplied this capture. The 701 unrelated deletions
in the old temporary worktree remain untouched; canonical main is used.

This narrows the investigation; **it is not a performance repair**. Next inspect
the existing no-paint RenderPath execution and raster stage boundary with source
and focused offscreen evidence before proposing another native candidate. The
metadata observer allocates and locks, so it does not establish ordinary-build
cost or observer neutrality. No renderer/cache/pacing experiment is adopted.
Features, SDK, application and shipping defaults remain frozen. The 33.3 ms
receipt-to-actual-presentation p95 gate is unchanged; M3–M7 remain open.
