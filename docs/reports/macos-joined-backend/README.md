# Joined CPU/native backend timing

The original Cogent/WebGPU design continues. This checkpoint is a private
instrumentation build, not a direct Metal replacement, performance repair or
release qualification. No native capture has run yet.

The exact reviewed CPU-upload and Metal-backend observers are combined in fresh
private source copies, with the existing scalar raster journal and three reviewed
source-detail overlays. Allocation, cache, rendering and pacing policy are
unchanged. Dependency names/versions remain identical; only the two exact
wgpu-core/hal registry entries become private paths. The archive is never installed
into the frozen SDK or Placement app.

The same 248 routed requests join raster/upload/scheduling to actual native
render commits, GPU start/end, present request and drawable feedback. The new
classifier reuses both strict parents, adds post-acquisition interval closure,
keeps the actual p95-ranked ticket, retains every command identity and rejects
unassigned trailing submissions. CPU commits/presentation requests can overlap
GPU execution. GPU span includes gaps between buffers; its end-to-shown remainder
is not an isolated compositor measurement.

Six focused classifier methods/19 negative mutations pass. The unchanged
raster recorder's five prior race methods are reused by exact source comparison.
The original routed complete-pixel/state fixture passes with the raster clock;
fresh native and Go builds, full source/pins/log/owned-process checks, exact recipe
reproduction and all 8,321 frozen SDK files verify. An initial native build chose
an incomplete offline Cargo cache and stopped before compilation; its logs are
retained. The corrected build uses the existing populated cache. Inherited cfg,
LLVM stripping, AppKit deprecation and duplicate-library warnings remain visible.

The [protocol](../../../tools/macos-calibration/joinedbackend/PLAN.md) permits one
newly authorized attempt, 40 seconds native /60 seconds overall, after review and
commit. Failed captures stay sealed; no unchanged retry. Observer overhead is not
calibrated and no shipping-build equivalence, physical input, displayed-pixel or
whole-GPU-allocation claim follows. Features/defaults/SDK/app and the 33.3 ms gate
remain unchanged. M3–M7 remain open.

Prism Anthropic Claude Sonnet5.5 found incomplete command ownership beyond the
cohort. The fix explicitly assigns every native command to either a joined
cohort drawable or a bounded leading startup present before first receipt. The
regression constructs an unrelated intercohort render/present accepted by both
parent classifiers and refused by the combined classifier. No binary or rendering
policy change was needed.

The follow-up review's conditional module-registration concern does not apply to
the executed import graph: all five parent validators import by explicit paths,
with no runtime self-import, name lookup or registration-dependent definitions.
Test-only alias changes load historical fixtures. The source audit and both
complete Prism reviews/dispositions are retained in the preparation archive.
