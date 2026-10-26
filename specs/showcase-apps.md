# macOS showcase applications

Build three independent runnable example apps using existing public Go APIs:
Forma (canvas design), Luma (original image library), Pulse (sample telemetry).
The purpose is polished, representative README screenshots with real interactions,
not new framework capability or release qualification. No new framework features/API/backend,
performance gate, global settings or user-data change. Keep sample data explicit.

Each demo has real controls and application-owned state, stable identities,
keyboard-accessible actions and bounded data. Forma retains Undo/Redo; Luma has
search/favorites/looks; Pulse has a virtualized source, filtering and chart ranges.
Data is fictional, in memory, and resets on launch. Original artwork is generated
from source; no Apple assets or personal photos/files are loaded.

Verify models and routing with focused tests; inspect the actual native windows
and capture those windows for README images. Do not replace app pixels with
mockups or compose fictional window chrome. Keep reproduction commands and
limits beside the examples. Review checkpoints with the established Prism /
Anthropic Claude Sonnet 5.5 process and commit. No public release assets/tags.

User feedback during native inspection identified cramped/misaligned sidebar
icons and incorrect dark table selection. Correct the app sidebar metrics and
the existing grid's scoped palette/spacing behavior in both render routes. Keep
read-only selected rows continuous, with editable/unselected active cursors
retained. This targeted appearance repair is authorized; no rendering policy,
backend or performance-gate change. Verify light/dark and DPI pixel parity.
