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

## Finder contract

Driver.OnOpenFiles delivers detached absolute paths on the application owner,
with at most sixteen pending paths and one owner waiter. Initial marks requests
captured before Driver.Run; warm delivery waits for native window readiness.
Opt-in bundle metadata captures cold launch events before Go driver binding.
Unbinding restores the previous native handler and drops queued references.
Apps validate their own document extension/content and retain recovery/dirty
Save/Discard/Cancel decisions. The current examples refuse multi-file batches;
Documents routes to the active window, with deterministic startup fallback.
Registration uses Alternate handler rank and never overrides the user's default.

Native Quit is intercepted at AppKit's applicationShouldTerminate: selector as
well as the existing Go Quit request hook. The original delegate implementation
is restored on retirement; opted-out applications keep upstream behavior.

## Find/Replace contract

Use the existing ui.FindBar and Unicode editor operations in Notes-family apps.
Search the selected note body only; preserve sidebar filtering. Command-F,
Option-Command-F, Command-G, Shift-Command-G and Escape use app commands and
existing native FIFO settlement. Resolve query/replacement after settlement,
refuse stale selected-note identity/composition, and record content changes in
the existing notebook journal. Navigation does not dirty a document. Replace
All is one notebook Undo/Redo step. Closing restores prose focus; query state is
session-local. No regex, cross-note replacement or new editor widget.

## CI contract

The existing full check wrapper remains the ordered union of six individually
runnable stages: tooling, portable, native, dependencies, examples, offscreen.
Preserve existing commands, test selectors, race checks, vet/build coverage,
benchmarks and timeouts. Hosted jobs restore only module downloads/Go build
cache with OS/architecture/Go/dependency/native-pin keys, then reconstruct and
verify pinned native source and run module integrity checks. Prepared native
source and published artifacts are excluded. Hosted results require a pushed
commit; local parsing and dispatch checks do not claim a green hosted run.
