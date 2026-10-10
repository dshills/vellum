# Workbench interaction validation

The portable version 2 session and native controller are implemented. The version 1 workspace API remains compatible. See the [contract](../../../specs/workbench-interaction.md).

## Checks

Portable `go test -race ./workbench` and targeted native `go test -race -run 'TestWorkbench|TestDock' .` pass. Regression coverage includes:

- Complete window/split/tab/active/focus/bounds JSON and file roundtrips, private snapshots, global group identity checks, cycle/depth rejection and failed whole-session immutability.
- Pointer tab reorder/edge split through real widget listeners and the owner mailbox, cancellation over internal splitters/gaps, external-drop floating, retired-generation delivery and mailbox refusal. These are synthesized event tests, not a physical pointer acceptance claim.
- Fresh native panel resources across window contexts, cached per-window native view factories, and distinct native lifetime IDs across separate controllers.
- Failed native Open and Close, failed autosave after source retirement, source-window replacement compensation, and application shutdown if compensation cannot restore a coherent native topology.
- Native close hook focus/panel transfer, idempotence of retired adapter close notifications, and synchronous controller close callbacks that must not reenter a transaction. Failed synchronous closes restore the adapter mapping.
- Persisted focus rebasing, stale native geometry rejection after explicit restore, native geometry burst retention, and preservation of live user move/resize while owner capture is pending. Initial geometry, changed bounds, and explicit Load of unchanged bounds still apply.
- Bounded latest-capture coalescing, deterministic rejection of older timer callbacks delivered after newer ones, transient/minimized geometry filtering, harmless retired timer delivery, and accepted transactions that survive a refused post-commit wake.
- Smoke readiness filters the exact native lifetime resolved on the owner, rejecting retired or unrelated notifications. The smoke runner returns errors and unwinds its cleanup before main exits. Record-tree close traversal visits every extension and joins failures.
- Queued errors with no callback preserve the first failure across later errors and mailbox refusal until Sync reaches the central driver failure path without destroying retained native resources; a group handler assigned after Open still receives errors. Test setup and encoded snapshots fail immediately if construction or serialization fails.

A race-built `cmd/docking-smoke` ran successfully on macOS against the completed rendering backend in an isolated phase 1 source snapshot. It opened real floating native windows, moved a shared document between windows, verified focus transfer, requested the real native close callback, and loaded the complete autosaved session. The native smoke used owner-posted commands; it did not send physical pointer gestures. Logs: [portable race](portable-race.txt), [native adapter race](native-race.txt), [smoke readiness race](readiness-race.txt), and [live native smoke](native-smoke.txt). Native vet also passed. The final source was tested against privately prepared committed phase 1 dependency inputs, without the later accessibility patch.

Physical pointer and tab-menu acceptance remains incomplete. The interactive fixture reached READY, but the native automation app-binding call hung and was aborted before any pointer or menu action was performed; [the attempt log](physical-attempt.txt) records READY only. The fixture ended without changing OS input/accessibility preferences. Phase-specific Prism review is performed by the coordinating phase before commit. The isolated smoke excludes the later accessibility bridge; the integrated accessibility-enabled docking run is a separate check.

## Scope

Cross-window redocking is available from a tab's explicit destination window/group menu. Pointer docking and tab reordering operate within the current native window; dropping outside the source scene creates a floating window, while a drop over an internal splitter or gap cancels. OS native cross-window drag payload exchange is not implemented. Documents/undo history remain domain objects, while per-window native adapters must be created through Panel.NewView when a panel contains extensions.

Persisted session geometry uses OS units and constrains restored frames to keep title bars on current screens. Monitor identity, maximization/minimization/fullscreen and document contents are not serialized. Atomic replacement does not guarantee directory-fsync durability under power loss. Desktop acceptance on Windows/Linux is tracked in platform support.

## Native close review context

The native smoke sends `CloseReq` from its worker deliberately. In the pinned prepared Cogent sources, `AppMulti.WindowByName` locks the native window list; this fixture's READY channel handoff follows native name and close callback initialization. `renderWindow.NewRenderWindow` installs a synchronous close callback that locks the render context and calls `Scene.Close`. The Vellum driver's close event then enters the non-reentrant application owner through `a.Do`. Sending `CloseReq` from inside the owner would deadlock on that callback. Cogent's desktop `Window.OnCloseReq` also dispatches `go w.CloseReq()`, and desktop `Window.Close` marshals native GLFW destruction through `RunOnMain`. The smoke uses the same background request boundary; every document/layout mutation remains owner-posted.
