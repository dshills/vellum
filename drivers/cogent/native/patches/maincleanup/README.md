# Main-loop native cleanup hooks

Patch 0045 adds a bounded desktop cleanup registry. Registration/removal is
worker-safe; terminal ownership detaches callbacks and runs them once in LIFO
order on the main thread, after base queue closure and before GLFW termination.
Callbacks execute without the lifecycle/registration lock. A hook panic does not
prevent the remaining hooks or termination, and the first LIFO panic propagates.
At most 16 live registrations are retained; removal clears retired references.

The Mac service adapter registers a cleanup that retires Go admission and runs
Cocoa panel/menu/tray removal directly on this terminal native thread. This
prevents cleanup from depending on an asynchronous main-queue request that might
miss loop termination. Other adapter targets register no hook in this phase.

Tests use injected callbacks and termination, not GLFW/Cocoa runtime operations.
Actual Mac status-item and native panel teardown remains live acceptance work.
Input/output hashes and the dependency manifest bind all three patched files.
