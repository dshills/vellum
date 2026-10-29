# Quit and multiwindow lifecycle — 2026-10-29

The installed build20 Documents application passed native keyboard Command-Q
with a dirty A notebook and clean B notebook. The save decision appeared;
Cancel retained A's seven-note dirty model. B had already closed through its
ordinary clean-close path. Saving A and invoking Command-Q retired the process.
Relaunch reopened saved A with seven notes and independent B with six; closing
A retained B, and Command-Q retired the last window. Fixtures were disposable.
See 02-lifecycle.json. These are app-restricted AX/keyboard observations, not
human or physical pointer qualification.

## Internal review and checks

Reviewed driver CloseRequest gating before widget cleanup, owner queue dispatch,
native QuitClean's snapshot/close walk, active-reference retirement, repository
locks and Notes worker Stop. Focused race tests passed for native Quit/concurrent
construction, close veto/forced close, independent active-document routing,
Cancel/discard, redo-to-saved and worker cancellation. No code defect was
established by these checks; no runtime or framework change was needed.
The initial helper failed on a changed native modal title; using the observed
modal title completed the same fixture without another mutation.

## Existing behavior

Quit requests close each window independently. A canceled Quit can leave a
clean window already closed while preserving windows that refused closure.
This is not an atomic all-window transaction. Last-window close ends this
application's Run; there is no background app/Dock-reopen document manager.
A new launch creates or loads the app's ordinary startup windows. No new
lifecycle feature or platform qualification is introduced here.
