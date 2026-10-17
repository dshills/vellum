# Exact small-glyph placement coverage and current native sanity

Source `2be8e790f8d5cdce5e0adac31b2adf5d73a76656` adds owned preparation stage
0076 for bounded renderer-owned small-glyph scene coverage reuse. Scoped
offscreen correctness and current-source native sanity pass. All four native
scroll cases still fail the unchanged 4 ms p95 / 6 ms p99 UI-work targets.
No native performance improvement or broad no-regression is established.

## Source and automated evidence

The shortcut uses an outline-cache epoch, nonwrapping entry ID and exact final
float32 scale/position bits. Live outline ownership, current coverage identity,
clip/scanner/spanner guards and original work charges still apply. Independent
FIFO placement metadata is capped at 8,192 entries and retains no outline/span
arrays; its additional map/metadata storage is separate from the existing
coverage backing budget. Saturated low-hit probes temporarily use the original
path. Underfull clipped misses are not covered by that cooldown. Unsupported
fonts, transforms and other refused cases retain ordinary rendering.
See [the renderer contract](../../../specs/rendering.md) and
[owned stage](../../../drivers/cogent/native/patches/glyphplacement/README.md).

Four A/B/B/A offscreen processes compare the exact prior four-file compiler
overlay with the candidate on otherwise identical source. This baseline is not
a clean historical executable. Each process produces 24 retained-grid records
across fixed/variable rows and DPI96/160/192; all byte/pixel/geometry identities
match. At DPI96, pooled ten-observation capture medians fall from 1.235 to
0.932 ms for fixed rows and 1.963 to 1.768 ms for variable rows (24.6% and 9.9%).
Bitmap routes at DPI160/192 are mixed. Five-observation synthetic cold-coordinate
and underfull clipped-miss medians rise about 2.0% and 2.3%. These are isolated
ScenePacket capture measurements, not native frames, physical GPU work or
memory qualification. CPU labels report 14 despite a GOMAXPROCS2 request;
exclusive-core or complete host isolation is not claimed.

Preparation checks, owned raster/sceneplan race checks, the complete offscreen
driver-root race suite and vet pass. The first full driver attempt times out
at five minutes and remains failed. A test-only revision removes duplicate
full per-pixel GPU integer execution from the new companion test; the original
integer oracle remains unchanged. All 24 companion identities remain equal,
and the final complete driver suite passes in 233.99 seconds. Six intentional
guard-removal source overlays each fail their intended regression. Earlier
failed test/control variants and preparation refusal records remain archived.

Internal reviews precede adoption and commit. One complete Prism call uses
Anthropic `claude-sonnet-5-5` on ten files / 53,368 diff bytes. It retains zero
findings and discards one tentative medium about missing path-key inputs.
That raw concern is independently assessed and refuted against the complete
path-input and supported-font rules. No source/test changes follow the provider
review. The native evidence below has separate internal admission/result review;
it does not trigger another provider call.

## Current native execution

A fresh release binary identifies clean `2be8e79`, with 4,678 tracked and 6,249
prepared files. SHA256 is
`73496fe0ea3410242013839565c55a5546391d015125167d246aea95c9892098`.
Both source checkouts remain clean and byte-identical through completion.
This build is separate from the earlier precommit compile-only binary.

One Cocoa million-row accessibility case passes the 400-visible-cell minimum,
2,268 overlap reuse observations, 459 replacement retirements, native cell
press delivery and closed-attachment refusal. Observations are not counts of
distinct reused objects. Four further processes each complete 128 guarded
warmup events and 120 measured callbacks exactly once, without missing or
pending records. Native endpoints match at 1519×1266 pixels, DPI162/DPR1.

| Million-row case | p95 UI work (ms) | p99 UI work (ms) |
|---|---:|---:|
| cpu-fixed | 6.698 | 7.008 |
| cpu-variable | 6.991 | 7.259 |
| gpu-fixed | 6.030 | 6.439 |
| gpu-variable | 6.216 | 6.442 |

These are four current-source sanity processes, with no matched baseline and
no complete P6 performance campaign. GPU requests do not establish actual GPU
execution or presentation. Default-DPI native placement-cache hits are not
proven. Endpoint equality does not prove continuous display stability.

The shared live-test lease, universal/local stop guards, 180-second watchdogs,
8 MiB per-stream limits and independent owned-group cleanup cover all five
processes. Each exits zero, reaches both EOFs and has its process group observed
absent before the next dispatch; the campaign is sealed afterward. The controller
passes three finalization and 17 refusal controls. A reused control-output
directory is refused before cases and preserved; a fresh version passes. An
offline analyzer initially addresses the wrong timing nesting and fails before
writing a result. Its corrected version revalidates the same raw logs; no native
process is rerun or relabeled. No other owned builds/profiles overlap ordinary
timings. Complete global host isolation is not proven.

No desktop pointer/key events or input-source, VoiceOver, clipboard or OS
preference changes are made. Fixture geometry persistence is disabled.
Physical workflows, human accessibility, native/GPU memory, sustained frame
targets, P4/P6/P9/P11 and full Mac/release qualification remain open. No public
release, signing, notarization or publishing occurs.

See [structured results](result.json), [raw-file manifest](evidence.manifest.json)
and [raw evidence](evidence.tar.gz). The bundle includes failures, source
identity manifests, controllers, reviews and raw outputs, with executables
excluded and identified by digest. It is not a portable source/toolchain closure.
