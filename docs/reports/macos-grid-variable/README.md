# Indexed variable-row grid checkpoint

The portable Grid accepts optional indexed row geometry through
`virtual.GridRowMetrics`. Fixed-height callers keep their existing configuration;
variable rows use actual natural cell heights, pixel-distance paging and stable
key anchors with clamped in-row offsets. `RowHeights` supplies cancellable,
immutable prefix preparation for loaded data. Its O(N) preparation and retained
index bytes are explicit source costs; owner transitions inspect only the
bounded admitted viewport/overscan range.

The managed adapter uses those portable rectangles for its existing native
panes, hit testing and visible-cell accessibility. Actual materialized rows,
including overscan, remain subject to native local dimension limits. An unused
fixed-height fallback does not constrain an indexed source's actual small rows.
The independent Grid application demonstrates 10k and 1M procedural rows with
stable-key heights, reverse/filter and per-window height revision toggles. Run it
with `-variable-rows`; its default fixed mode and original tests remain supported.

Internal review reproduced three invalid-index defects: inconsistent direct
bounds versus pixel inverse, backward paging from contradictory spans, and
accepting a resize with an invalid newly exposed interior row. The correction
checks predecessor ordering and inverse endpoints, then validates every admitted
span before committing. The original failing test bodies and their corrected
green execution remain in the review evidence; formatted copies are permanent
portable regressions. This local validation does not certify hidden spans or
permit an unrevisioned source geometry mutation.

Portable race checks cover brute geometry, helper ownership/endpoints,
cancellation, overflow, atomic source failure/revision drift, anchor shrink/grow/
reorder and tall-row paging. A common viewport over 10k and 1M alternating
11/37-pixel rows ×20 columns makes identical reads: 24 RowKey, 24 IndexOf,
240 Cell, 154 boundary and 102 pixel callbacks. It materializes 240 cells and
10 headers in either dataset. This is a bounded-work proof, without a latency,
allocation or stored-million-row memory claim.

Native offscreen tests exercise actual primitive pointer, key and wheel routing
and clipped odd-height geometry at 96/120/144/192 scale settings. They cover
pixel paging, stable anchors, stale press retirement and native capacity/dimension
refusals. Consumer tests also route the height-toggle control and verify window
independence. Qualification records identify exact source inputs, prepared
dependency fingerprints, command exits and raw logs. Exploratory test-assumption
failures remain historical, separate from the final source-stable checks.

The desktop smoke runs fixed and variable modes sequentially using actual Cocoa
windows and completed measured frames. Its explicit owner commands exercise
far jumps, frozen geometry, indexed replacement, variable pixel Page, closed
binding refusal and window lifetimes. It does not synthesize physical desktop
input or certify the compositor's teardown. The smoke result and native runtime
process exit are separate evidence.

[qualification.json](qualification.json) binds the final tested inputs and logs.
The [review record](../../reviews/macos-grid-variable/README.md) preserves
internal findings, Prism using Anthropic Claude Sonnet 5.5 and their resolution.

This advances P6 but leaves range/all-row selection, editable cell conflicts,
loaded-data projection/index preparation integration, large-tree acceptance and
sustained latency/frame/memory measurements open. External AX and human IME/
VoiceOver checks remain separate. No input sources or macOS preferences were
changed; no public release, signing or notarization ran.
