# Current M2 functional qualification

All eleven existing native methods pass on source `93b4730` (19 October 2026).
Fresh executables use that source's prepared Core stage 97 and typesetting; only
the independently pinned source-built GPU linkage capsules come from the older
private SDK. This is current-source qualification, not SDK source equivalence.

| Method | Executed evidence |
| --- | --- |
| Dedicated Cocoa selectors | Tab close: 8 stages; Breadcrumb: 12 stages; ColorPicker: 8 transitions |
| Controls | 57 routed inputs / 21 observations |
| Choices | 30 / 23 |
| Navigation | 47 / 41 |
| Media | 71 / 42 |
| Structure | 50 / 42 |
| Text/editor | 68 / 33 |
| Professional workspace 10k rows | Original 216 operations / 27 captures |
| Professional workspace 1m rows | Original 216 operations / 27 captures |

The unchanged full software workspace record also passes at both row counts and
owner batches 1/7/64: 1,296 operations and 162 exact original model/image captures.
The [compatibility checkpoint](../../reviews/macos-style-compat/README.md) fixes
the two regressions that this gate exposed, with internal and Anthropic Claude
Sonnet 5.5 Prism review. No records, oracles, goldens or acceptance assertions were
re-authored to pass.

Eight current-source prerequisites pass (effective catalogue/model race, two
vet checks, five optimized builds). The 11 native processes run serially under one
exclusive lease with fixed source, input and binary hashes, original limits and
strict validators. All logs are complete with empty stderr; every owned group is
terminal and observed absent, per-case temporary directories are empty, source/
prepared/linker/settings boundaries match before/after, and all scopes are sealed.
The external `macos-m2-renewal-after-compat` directory retains the raw records and
producer/admission receipts. The [stored summary](records.json) binds their hashes.

M2's controlled functional workflow/catalogue gate is satisfied together with the
existing [dedicated-contract audit](../macos-current-catalogue/README.md). This
closes the three dedicated Cocoa methods and current shared-runtime renewals
previously listed as M2 gaps. Earlier failed attempts remain failed and sealed.

These methods have distinct evidence limits. Catalogue/text/workspace input uses
the actual Go adapter FIFO, not hardware input. Cocoa actions use native selectors;
some between-snapshot checks are source assertions, and ColorPicker does not emit
detached native snapshots. Breadcrumb stale-path guard isolation remains covered
by software, not independently isolated by its native record. Geometry/semantics
and CPU-scene captures do not qualify all native pixels, progress painting or
Slider/palette physical dragging. One invisible, empty-title TUINS Cocoa window
remains unclassified; process retirement is not a native-deallocation proof.
Those physical/human checks, real presentation performance and resource/endurance
qualification remain in M3–M5. M6/M7 and full Mac completion remain open.
