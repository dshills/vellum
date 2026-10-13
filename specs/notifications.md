# Mac local notifications

`app.NotificationServices` is optional; legacy Services remain compatible.
`Application.Notifications()` returns the extension when provided. Mac/cgo
implements permission snapshots/requests, immediate local delivery and targeted
withdrawal. Capabilities describe implemented transports independently of bundle
availability, permission and actual presentation. Other adapters report false
capabilities and unavailable operations.

Call from workers after native services become available. The Mac bridge requires
an app bundle with a nonempty identifier and APPL package type before accessing
UserNotifications. Ordinary Go command-line executables return Unavailable.
Existing bundle tooling supplies that metadata. Compilation does not establish
OS recognition, signing or packaged runtime behavior.

`NotificationSettings` queries current settings without prompting. Permission is
Unknown, NotDetermined, Denied, Authorized or Provisional; future native states
normalize to Unknown. AlertEnabled/SoundEnabled reflect enabled settings, not
guaranteed banners or sound. Only `RequestNotificationPermission` requests alert/
sound authorization and may prompt. Denial returns a Denied snapshot; native
errors return Unavailable. Settings may change later; this API cannot revoke them.

`Notify` never requests authorization. NotDetermined/Denied return
`ErrNotificationPermission`; unknown native authorization returns Unavailable.
Authorized/provisional requests use immediate local content. Success means OS
acceptance independently of notification retention, Focus/Do Not Disturb, sound
preferences, foreground behavior or actual display. The OS/application controls
the notification-center delegate and presentation policy. This API defines no
custom categories, activation commands, attachments, badges, remote push or delays.

IDs are 1..128 ASCII bytes from letters, digits, '.', '_' and '-'. The reserved
native namespace is `vellum.notification.` plus the public ID within the app's
center. Other notification code in that app must not use this prefix. Reusing an
ID asks the OS to replace that request/delivered item. Title/subtitle each allow
256 UTF-8 bytes, body 4096, with no NUL or invalid UTF-8. Title or body must be
nonblank. Sound is opt-in and defaults off.

`WithdrawNotifications` accepts 1..32 distinct valid IDs and targets only their
reserved identifiers through pending and delivered removal APIs. It acknowledges
dispatch; those methods have no completion callback, so return does not establish
actual removal. No remove-all operation is used. Concurrent Notify/Withdraw have
no transaction/last-call-wins guarantee; wait for accepted delivery before
withdrawal when sequencing matters. An earlier delivery's asynchronous permission
read may otherwise initiate submission later.

Operations share the eight generic native service credits. Admission precedes
JSON encoding/copy; canceled callers retain credit until native terminal ACK.
Requests are capped at 32 KiB before C copying, separately from the shared 1 MiB
maximum. Required-field/typed settings decoding is capped at 256 bytes; the shared
C-to-Go bound applies before copying. Worker and queued/asynchronous callback
pools own detached data through completion. Tokens/retirement follow the shared
service contract. Late callbacks after shutdown cannot reopen admission.

Keep input slices immutable until return. Cancellation can race accepted
permission/delivery/removal and cannot undo OS state, dismiss an authorization
prompt or force a stalled callback. Eight retained requests bound framework
admission/payload ownership; OS history and callback/dialog duration are separate.
Delivered items may outlive the process and are not automatically withdrawn at
shutdown.

Offscreen tests cover options, compatibility, public consumers, context preflight,
native parsing, actual UNNotificationContent/Request data objects, detached
producers, injected settings normalization and permission-error callbacks. They
never access UNUserNotificationCenter, request permission, send/withdraw messages,
create UI, play sound or change OS settings.

## Separate native acceptance

The opt-in `vellum_native_checks` consumer `cmd/notifications-smoke` checks a bare
executable separately from an isolated private `.app`. Bare operations, including
the permission method, must refuse before accessing the center; bundle metadata
is checked before that method is invoked. Packaged checks read actual settings
and use unique per-run nonce IDs. They never request permission, install a
notification delegate, request sound or alter OS settings.

`InspectNativeNotifications` is available only in that build. It accepts 1..32
valid public IDs and requires a bundle identifier beginning with
`com.vellum.native-checks.notifications.`. Only the requested reserved identifiers
are returned. Pending and delivered lists are separate asynchronous observations,
not an atomic union. Native input is limited to 8192 bytes, each OS collection to
4096 entries, and aggregate reply JSON to 64 KiB. A conservative escaped-text
budget reserves each selected row before collection; oversized content, duplicate
results or a complete reply that exceeds the bound are refused, never truncated.
Thus 32 maximal bodies need not fit the diagnostic reply even though the public
service accepts a 32-ID withdrawal batch. Ordinary/offscreen builds exclude the
probe, consumer and their symbols; the SDK source payload may contain them.

On NotDetermined/Denied, the consumer requires a permission error, unchanged
settings and no retained owned item. On Authorized/Provisional, it requires actual
delivered content, replacement and targeted removal preserving a second item.
This conditional branch is not qualified by a permission-refusal run. Restart
checks the previous run's exact IDs in the same bundle; an initially stale item
fails the check even if later cleanup removes it.

The worker's 45-second deadline precedes the native loop's 90-second deadline.
Cleanup uses a fresh 15-second context while services remain alive, waits for
retained native operations to ACK, dispatches withdrawal only for the fixture's
exact IDs and polls for observed empty lists. A timeout preserves failure rather
than claiming cleanup from a canceled caller or removal ACK. Outstanding counts
and inventories are sampled observations; they do not establish concurrent
writer ordering or future OS state. Permission prompts/denial changes, visible
presentation/sound, authorized replacement/removal, foreground/delegate
interoperability and broader restart behavior remain acceptance work until their
corresponding live evidence exists.

An unexpected native-loop exit cancels the worker and waits 16 seconds, longer
than its fresh cleanup deadline, for a terminal outcome. A late worker report is
preserved; no reply produces a scoped failed report. Both paths mark cleanup
unverified and refuse a phase pass because the loop has already stopped. This
preserves evidence without claiming that an ended loop can complete withdrawal.

Primary APIs: [UserNotifications](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter),
[authorization](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/requestauthorization(options:completionhandler:)),
[local submission](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/add(_:withcompletionhandler:))
and [targeted removal](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/removedeliverednotifications(withidentifiers:)).
