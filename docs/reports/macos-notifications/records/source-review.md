# Notification source review V1

APPROVED_SOURCE. No actionable findings in the 15 frozen files. Frozen and current hashes match.

Only RequestNotificationPermission requests alert/sound authorization. Settings is nonprompting, and Notify checks settings without authorizing. The native bridge requires application bundle identity before using the center; capabilities describe transport rather than current availability. Delivery acknowledgment is distinct from actual presentation.

Delivery and targeted withdrawal use the same validated `vellum.notification.` namespace. Withdrawal never removes all notifications. The spec accurately records asynchronous removal and concurrent delivery limits, cancellation without rollback, and delivered notifications surviving application shutdown. No notification-center delegate is installed.

The bounded detached JSON producer, shared eight-credit registry, nonreused tokens, active lifetime checks, pooled consumers, exception-to-terminal callbacks and before-copy reply bound remain intact. The sole shared services_darwin.go change relative to the approved tray V5 is status 4 mapping to ErrNotificationPermission.

Fixtures cover real UN content/request data construction, malformed wire, injected settings normalization and actual C-to-Go permission-error mapping. They do not call the notification center, display notifications, prompt, play sound, or modify preferences. Existing headless receipt hashes and raw log hashes were checked; recorded root/driver/public SDK race, driver vet and native compilation exits are zero. This review does not attest full compiler or SDK closure or real bundled live behavior.

The JSON certificate records the exact 15 source hashes, receipt identities and explicit limits. The planned Prism review is separate and has not been certified here.
