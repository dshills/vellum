# Settled native static-idle observations

Five independent ordinary Mac hello processes each complete 100 seconds of
static idle after verified GPU readback. Every process exits normally. Each run
samples process CPU/RSS twice across a settled interval of at least 60 seconds;
raw samples and all five results are retained. Current source/prepared endpoints
remain clean and identical at commit df4bcb1 throughout the campaign.

| Run | Settled seconds | CPU, % of one core | RSS before / after (MiB) |
| --- | --- | --- | --- |
| 1 | 60.019 | 0.367 | 236.984 / 236.938 |
| 2 | 60.022 | 0.333 | 227.297 / 227.375 |
| 3 | 60.020 | 0.233 | 236.109 / 236.141 |
| 4 | 60.013 | 0.300 | 227.562 / 227.594 |
| 5 | 60.020 | 0.233 | 228.125 / 228.094 |

CPU comes from macOS ps cumulative process time with a 0.01-second quantum.
RSS comes from ps resident pages and includes native/shared/driver mappings; it
is not private footprint, native allocation attribution or a Go live-heap check.
RSS changes by at most 0.079 MiB within each sampled minute. Runs do not establish
long-term leak freedom or native resource teardown counts. The raw results also
contain startup/readback/heap observations; startup is verified GPU readback,
not first interactive presented frame or controlled cold-cache startup.

This fixture contains a small counter UI plus the rectangle canvas. Its reported
canvas readback is 3054×722 physical pixels in every run; the configured 640×480
window description is not effective geometry qualification. No primary
1920×1080/1,000-control layout, alternate DPI/refresh, occluded/minimized case,
wakeup or repeated-present counter was measured. The observed CPU numbers are
below the provisional 0.5% reference for this smaller fixture only. Process RSS
is above the provisional 150 MiB reference and does not qualify the full workbench
memory gate. Complete P6/P11, overnight idle/background soak, presentation and
Mac release acceptance remain open. No OS input, VoiceOver, clipboard, display
or power setting is changed for this campaign.

The checkpoint requires an internal evidence/source review and bounded Prism
review using Anthropic Claude Sonnet 5.5. Producers, raw observations, source
identities and review outputs will be archived with a verified SHA256 index.
