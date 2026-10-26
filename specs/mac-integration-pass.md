# macOS content integration and polish

Authorized 26 October 2026 after merging e261cc6: implement these five checkpoints,
review each with Prism / Anthropic Claude Sonnet 5.5, and commit separately.

1. Bounded native interaction checks of the existing content widgets. Generate
   disposable PDF/audio/path fixtures; distinguish native mouse/key events,
   accessibility actions, descriptor/offscreen checks and human qualification.
   Record reproduced failures for checkpoint 2; never relabel a failed check.
2. Correct native appearance/interaction defects: narrow layouts, focus,
   loading/error feedback, scoped Light/Dark and untinted full-color artwork.
3. Adopt the public checklist/date controls in the existing fictional Reminders
   example, preserving its command/history and input settlement semantics.
4. Add practical, compilable public API examples of owner state/history,
   asynchronous path choice and native content cleanup.
5. Export a private SDK from clean committed source and package matching Mac
   Widgets and Reminders apps. Verify origins, signatures, extraction and hashes.
   Report exact built source; a later report commit does not relabel its bytes.

No new widget families, backend/pacing experiments, performance gate changes,
public tags/assets, signing identity/notarization, audit changes, global input or
accessibility settings, or user notebook edits. This scoped authorization extends
only the above work beyond the earlier private-v1 freeze. Preserve accepted v1
archives and API baseline. New private packages have distinct build identities.

Native input checks use the existing app-restricted mouse helper where computer
use coordinates miss controls. Foreground changes refuse further input. Clipboard
Copy verification requires preserving the original clipboard; automatic approval
review requested explicit approval for the proposed in-memory save/restore check.
Until approved it is pending, with no clipboard manipulation performed.
