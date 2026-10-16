# Render-window context retirement

Patch 0070 withdraws a closed window from all three render registries and
replaces its current construction context with a registered survivor, or nil.
Without this, closing the last manual window leaves a manual context that
silently refuses the next ordinary window.

Focus membership, ordering and current-context publication share the retirement
mutex. Raise refuses an unregistered window and publishes before its native
callout so a reentrant close cannot be overwritten afterward. Native calls run
after releasing the registry mutex; snapshots do not extend native lifetimes.

Offscreen regressions cover actual manual close followed by ordinary Show,
survivor/last/repeated retirement, registry reentry during Raise, native Raise
closing its own window, and concurrent native focus/close events. Existing live
manual-mode and mixed-mode refusal contracts remain in place. Physical desktop
focus and native object deallocation require separate live evidence.
