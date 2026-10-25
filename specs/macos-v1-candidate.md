# macOS v1 candidate finish contract

The owner approved all five areas on 25 October 2026. This contract supersedes
the personal-v0.1 API freeze only for the following work. Keep the existing
Cogent/WebGPU backend and experimental rendering/pacing disabled. No new widget
campaign, CAD features, performance threshold or broad qualification campaign.
Baseline: `d5e6b05`, including Notes build6.

The deliverable is a private macOS v1 release candidate and matching Notes/Mac
Widgets packages. It is not authorization to publish a stable release, tags or
assets. Target Apple Silicon on the recorded host macOS; other OS/hardware stay
unqualified. Human IME/VoiceOver remain explicitly unqualified until assessed;
automated semantics cannot substitute for those results.

| Phase | Required outcome | Checks and finish condition |
| --- | --- | --- |
| 1 Shared Mac presentation | Shared typography with emoji fallback, frameless prose styling and compact fields; Notes consumes those public defaults | Light/dark actual geometry, overrides and retained text checks; Prism and commit |
| 2 App shell | Independent native titlebar toolbars on multiple window lifetimes, original-color icons, command availability and state refresh; preserve old entry points | Window isolation, stale/disabled/modal/canceled action refusal, lifecycle and native compilation; Prism and commit |
| 3 Native interactions | Physical helper checks for pointer targeting, selection, keyboard and dialog focus; human limits documented | Bounded app-local native session and focused regressions for any actual defect; Prism and commit |
| 4 Reliability | Resolve intermittent cancellation-metrics assertion without losing cancellation or admitted input; assess ordinary foreground first paint | Deterministic cancellation/owner tests, bounded repeated checks and a native foreground observation; Prism and commit |
| 5 Candidate delivery | Public API compatibility/support matrix, accurate candidate metadata, independent SDK and matching app packages | Portable checks once after substantive changes; affected-module tests; fresh independent consumer/signature checks; Prism, commit, package and relaunch Notes |

The owner authorized one desktop window of up to20 minutes when ready, including
the existing native mouse-event helper for Notes and Mac Widgets. Preserve global
input sources/accessibility preferences and the clipboard. Prepare/review/build
offscreen before starting that window. Reuse verification of unchanged paths.
Record unsupported or unqualified cases rather than implying v1 acceptance.
Actual advertised crashes, hangs, data loss, incorrect history, stale callbacks
or teardown failures require fixes before candidate delivery. Cosmetic follow-up
and public distribution work belong in an explicit backlog.

## App-shell API boundary

Phase2 adds the per-window `Driver.TitlebarToolbars` factory and optional
`Driver.CommandMenus`, preserving both singular legacy fields. Factories bind
exact WindowRefs and may return independent command registries. Native toolbars
keep an immutable item catalogue and dynamic owner-checked command availability.
Menus remain process-wide and use the shared registry; active-document routing
for independent registries is caller-owned. Concurrent toolbar attachments are
bounded at32 and item catalogues at16; customization may show at most32 items.
One coalesced availability update per toolbar and one menu refresh worker bound
pending work. Closing/reopening a window cannot revive its previous callbacks.
