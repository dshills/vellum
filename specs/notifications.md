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
create UI, play sound or change OS settings. Packaged permission/denial, real
presentation/sound/replacement/removal, concurrent ordering, foreground/delegate
interoperability and restart remain live acceptance checks until the user says
the computer is free.

Primary APIs: [UserNotifications](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter),
[authorization](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/requestauthorization(options:completionhandler:)),
[local submission](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/add(_:withcompletionhandler:))
and [targeted removal](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter/removedeliverednotifications(withidentifiers:)).
