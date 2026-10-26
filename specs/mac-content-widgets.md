# Five macOS content widgets

Authorized 26 October2026: reusable checklist, file/folder picker field, date-range
picker, embedded PDF viewer and media player. Each receives focused tests,
Prism review with Anthropic Claude Sonnet5.5 and a separate commit. This opens
only these additions; preserve the completed private v1 delivery, stable API
baseline, backend and performance policy. No public assets/tags or OS settings.

1. Checklist: controlled stable item identities, completion, inline names,
   caller-formatted due labels, priority and original-color category icons.
   Up to128 items. Tab/Space/native field editing; app owns history/storage.
2. File/folder field: controlled absolute local path, full-color artwork, native
   Browse and Reveal in Finder, Clear and file-drop admission. Worker-safe exact
   parent chooser; cancellations preserve accepted state. No content reads.
3. Date range: inclusive civil endpoints in one location, ordered/bounded dates,
   paired input/calendar controls, range highlighting and presets. Caller-owned
   Today/clock and history. DST/civil validation reuses existing date controls.
4. PDF viewer: local explicit file, embedded PDFKit view, thumbnails, search,
   zoom, page controls and native selection/copy. Native resources follow exact
   managed-node/window lifetime and clipping. Bounded native hosts/loading;
   worker-side file reads; metadata/decoder limits documented separately.
5. Media player: embedded AVKit native controls for explicit local audio/video,
   initial paused state, seeking/volume, optional original-color poster. No
   autoplay, network URLs, capture, sharing or external/fullscreen presentation.
   Pause/release on hide/removal/shutdown. OS codec availability applies.

Portable/offscreen/recipe checks are distinct from live native presentation,
physical input and human accessibility/IME qualification. No new broad campaign.

Implementation checkpoints live on `codex/mac-content-widgets`, separately
reviewed/committed. Native content uses canonical descriptor publication, exact
visible semantic membership and independent raw/clip geometry (including dock
namespaces). Eight active hosts share at most two file/metadata jobs; admission
and OS decoder limits are separate. PDF: 64 MiB/1024 pages. Media: 1 GiB regular
file admission, five-second metadata wait, initial paused transport and bounded
original-color PNG poster. See docs/controls.md for cancellation and overlay limits.
The private v1 archives and release/API baselines are unchanged.
