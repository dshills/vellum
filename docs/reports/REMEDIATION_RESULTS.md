# Cogent remediation: local field fixes and idle savings; G1 remains open

9 October 2026. The bounded follow-up to [the foundation trial](FOUNDATION_RESULTS.md) fixes the demonstrated single-line grapheme failure, adds real native selection, defers optional spelling allocation and demonstrates opt-in demand painting. **Keep Cogent as the experimental candidate; do not accept G1 or start the broad workbench yet.** The next unresolved seam is native IME composition and screen-reader editing, followed by font-cache bounds. See [ADR 0003](../decisions/0003-remediation-feasibility.md).

The next [single-field native input prototype](NATIVE_INPUT_RESULTS.md) has now been implemented; it adds composition delivery while retaining the foundation hold.

## Reviewable implementation

The separate [remediation fixture](../../prototypes/remediation/README.md) preserves the original foundation source and evidence. Its runner verifies the exact pinned Cogent inputs and applies three patches to a disposable module copy. No module cache changes or private toolkit imports are required. The patches are local experiments, unaccepted upstream; production reuse depends on supported hooks or upstream acceptance.

| Patch | Change | Review size, including tests |
|---|---|---|
| [0001](../../prototypes/remediation/patches/0001-grapheme-and-selection.patch) | Logical grapheme navigation/deletion, boundary normalization, read-only editing state and validated selection setter | 5 files; 249 added / 6 removed lines |
| [0002](../../prototypes/remediation/patches/0002-lazy-spelling.patch) | Initialize spelling on its first potential use or explicit line check | 3 files; 34 added / 4 removed lines |
| [0003](../../prototypes/remediation/patches/0003-requested-paint.patch) | Opt-in coalesced paint requests with explicit caret, resize and retry deadlines | 8 files; 230 added / 12 removed lines |

Grapheme boundaries use the existing pinned go-text segmenter, whose API provides a grapheme iterator ([official documentation](https://pkg.go.dev/github.com/go-text/typesetting@v0.3.5/segmenter)). Raw insertion and completion retain rune-offset semantics. This is logical single-line editing: multiline behavior, visual bidi movement and nearest visual pointer caret placement remain unestablished. Segmentation currently scans and allocates per operation; no long-line edit performance pass is claimed.

An integration bug was also corrected in the adapter: Cogent's `Text()` getter commits an edit and clears selection. Inspection, semantic publication and sample recording now use the new read-only `EditingState().Text`; intentional command commits can still use `Text()`. Snapshot tests verify that inspection preserves selection and edit state. Early diagnostic captures made before this correction are retained and excluded from passing evidence.

## P2/P3: actual field and native selection checks

The AccessKit bridge now emits a TextRun with UTF-8 byte lengths per selectable grapheme and copies native selection payloads into owned Go values before freeing the request. Character indexes map to validated rune boundaries; native Cocoa ranges use UTF-16. Invalid UTF-8, embedded NUL and clusters larger than the C API's 255-byte length representation are rejected explicitly.

The [verified native report](../../prototypes/remediation/evidence/native-verified.json) records 13 editing snapshots and **32 native callbacks / 32 handled actions**. Actual native range actions and keyboard events produced:

| Probe | Observed result |
|---|---|
| Backspace at unselected end of `é` | Empty, deleting both runes; Undo restores the sequence |
| Backspace at unselected end of `👩‍💻` | Empty, deleting the whole three-rune sequence |
| Native selection of emoji in `A👩‍💻éZ` | Rune anchor/focus `1/4`; Backspace leaves `AéZ` |
| Successive Left keys from end of mixed string | Cursor `6`, then `4` |
| Shift-Left from cursor `4` | Whole emoji selected, reversed anchor/focus `4/1` |
| Forward Delete at unselected cursor `1` | `AéZ`, deleting the whole emoji |
| ASCII `AB` control, Backspace at end | `A` |

The subsequent native command burst also retained `22k`, rotated once, rejected rotation while disabled, and replaced visible table rows with R9–R16. See the [AX capture](../../prototypes/remediation/evidence/native-verified-ax.txt) and [screenshot](../../prototypes/remediation/evidence/native-verified.png).

These results close the reproduced failure for the representative field. They do not close P2/P3. Text-run geometry and multiline ranges are absent; the scene-pixel caret snapshot is not wired to Cocoa's IME client. Real Japanese/Chinese composition and cancellation, candidate positioning, clipboard, controlled-font shaping, visual bidi editing and complete VoiceOver/NVDA/Orca workflows remain open. Native AX inspection is not screen-reader completion.

## P1: matched scheduling and spelling diagnostics

Same local Apple M4 Pro / 64 GiB / macOS 26.6.2 host and pinned dependencies as the foundation trial; [environment](../../prototypes/remediation/evidence/environment.txt). Both scheduling conditions use the same patched fixture, unchanged multilingual corpus, system fallback fonts and active native adapter. Each static interval begins after three seconds settling with non-text focus. The periodic condition disables only the opt-in scheduler.

| Condition | Interval | Process CPU seconds | % of one CPU | Window paint dispatches during interval | Canvas callbacks | Forced-GC Go heap | Current resident |
|---|---:|---:|---:|---:|---:|---:|---:|
| Demand painting | 60.001 s | 0.171 | **0.284%** | **0** | 0 | 276.8 MiB | 420.0 MiB |
| Periodic control | 60.001 s | 1.082 | **1.804%** | **3,600** | 0 | 276.8 MiB | 421.6 MiB |
| Demand, explicitly initialized spelling | 6-second run; no idle interval | — | — | — | — | 403.1 MiB | 586.7 MiB |

Raw [demand](../../prototypes/remediation/evidence/idle-demand.log) and [periodic](../../prototypes/remediation/evidence/idle-periodic.log) logs contain interval counters; [measurements.json](../../prototypes/remediation/evidence/measurements.json) collects them. Counters in the final reports include startup and shutdown and must not be mistaken for interval counts. `PaintSchedulerWakes` counts window-loop paint dispatches, not all OS or Go runtime wakeups; canvas callbacks are not compositor presents.

The single visible demand interval falls below the provisional CPU target. This is a diagnostic finding, not formal acceptance: one interval per condition, small fixture, uncontrolled warm font cache, no five-run 1,000-control workload, no measured occluded/minimized 60-second intervals, no general animation or frame-pacing validation. The toolkit's original periodic scheduling remains the default; this fixture explicitly opts in.

Lazy spelling saves about **126.4 MiB** of initial retained Go heap. The [explicit spelling control](../../prototypes/remediation/evidence/spelling-warm.json) restores that allocation while retaining the same multilingual text. This defers optional cost, rather than reducing it when used. The remaining heap and resident footprint exceed the provisional 40 MiB / 150 MiB budgets. The [sampled heap profile summary](../../prototypes/remediation/evidence/heap-top.txt) attributes about 213 MiB directly to OpenType table buffers and substantial additional allocation to parsed glyph/font data. Sampled totals differ from exact runtime HeapAlloc; font discovery/cache ownership remains the next memory investigation.

Demand painting also completed [six public resize requests](../../prototypes/remediation/evidence/lifecycle.json), with all six expected physical content sizes, eight canvas callbacks and normal shutdown with an active adapter. A separate [transition smoke](../../prototypes/remediation/evidence/restore.json) invoked Minimize and Raise, then verified rotation and normal Finish. Its capture does not independently establish minimized dwell or an idle interval, so minimize/restore acceptance remains open. Arbitrary animation, sprite/popup paths, global styles, device loss and platform coverage also remain open.

## Reproduction and decision

The fresh-copy runner passed **17 top-level race-enabled tests** (14 targeted checks plus three existing word-position tests), Go vet, locked native build and the resize/transition smokes. See [test output](../../prototypes/remediation/evidence/reproduction-tests.txt), [build output](../../prototypes/remediation/evidence/reproduction-build.txt) and [manifest](../../prototypes/remediation/evidence/manifest.json). Interactive, idle and spelling measurements preceded only native-string rejection and reproducibility/patch-format cleanup; the final rebuilt artifact was used for resize/transition verification. No scheduler, grapheme or spelling behavior changed between those runs.

The patch scope is small enough to justify continuing the upstream-hook investigation. It does not yet justify a permanent unsupported fork, nor prove that an independent shell has lower ownership cost. The next bounded step should provide native marked-text, selection and screen-coordinate caret delivery for one field, then exercise a real IME and VoiceOver. Resolve before-show attachment and bounded lossless action transport alongside it. If those require replacing substantial toolkit/platform internals, trigger the independent-shell comparison before expanding the workbench. Windows/Linux and clean-machine distribution remain required for G1.
