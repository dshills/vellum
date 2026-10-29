# Scoped Mac document workflows

This work authorizes exactly five phases while preserving the existing backend,
performance/release policy and all other frozen framework features:

1. Opt-in native coordinated Quit and app-owned save decisions before any closure.
2. App-owned external-document restoration using bounded metadata and existing
   document workers; recovery and invalid/missing-file safety take precedence.
   Reuse native window geometry persistence rather than another layout engine.
3. Native Finder document opening for registered notebook/study types, delivered
   through public callbacks into the existing validation and dirty-decision path.
4. App-owned Find/Replace through existing editor/history/input settlement APIs.
5. Reusable CI stages and verified dependency caches, preserving test coverage.

Use focused regressions and internal review per phase; preserve historical
records. No new rendering mode, CAD kernel, public distribution approval,
notarization, signing identity or global appearance/input/security setting change.
Final private SDK/apps must share committed source, pass extraction/signature
checks and install into a new versioned user Applications directory. Native
observations remain separate from synthetic routing and human qualification.

## Quit contract

Applications opt in to OnQuitRequested. Native callbacks never wait for the UI
owner; at most one goroutine awaits its serialized delivery. An app QuitGroup
accepts a bounded fixed set of live file windows, settles each input, asks each
dirty decision and validates approved revisions. No participating window closes
until all approvals and recovery clears complete. Cancel and any failure unlock
survivors. A successful save before Cancel remains saved. Applications that do
not opt in retain existing close semantics. Ordinary window close stays separate.

## Restoration contract

Notes-family windows opt into a separate bounded 16 KiB WindowSession record
containing version, last successful notebook path and last activation timestamp.
Reads happen at startup; writes and document reopens use existing document
workers. Recovery is resolved first and keeps its recorded destination. Unknown,
malformed or missing files preserve the live default; invalid records are kept
until an explicit successful file action replaces them. Metadata joins Quit and
worker retirement. Native geometry persistence is reused with stable titles;
A/B startup order restores the active document. Forge extends its existing
workspace with an optional validated LastStudy path and keeps recovery precedence.
Existing notebook/recovery schemas and Undo history semantics are unchanged.
