# Mac local notifications checkpoint

Optional `app.NotificationServices` and `Application.Notifications()` add current
permission/alert/sound snapshots, explicit alert/sound authorization, immediate
local delivery and targeted withdrawal in Mac app bundles. Notify never requests
permission. Denied/undetermined authorization returns ErrNotificationPermission.
Only explicit RequestNotificationPermission can prompt. Sound defaults off.

Reserved native IDs, bounded UTF-8 content, distinct withdrawal batches, shared
eight-credit dispatch, copied/pooled payloads and context/shutdown checks follow
the [contract](../../../specs/notifications.md). Success means accepted delivery
or dispatched removal; actual presentation/removal, ordering and rollback are
separate. The center delegate is unchanged. Delivered items may outlive the app.

Seven new tests cover root bounds/IDs/optional compatibility, independent public
consumer permission/replacement/withdrawal, context preflight, required typed
settings, actual native parsing/content/request objects, detached producers,
injected settings normalization and the actual C-to-Go permission error.
No test accesses UNUserNotificationCenter, prompts, delivers/withdraws, opens UI,
plays sound or changes OS settings. Full root/native offscreen race, root/driver/
consumer vet, public consumer race and final native compilation pass.

Initial ARC switch/block-lifetime compile failures are preserved and corrected
with per-case scopes. An initial full public consumer run failed on sandboxed
private Unix-socket binding; its authorized retry passes. Duplicate-libobjc
warnings remain. Go 1.27.2/macOS 26.0 minimum/SDK 27.0 describe local compilation,
without runtime or older-Mac acceptance.

Private SDK verification covers all 6943 files, all 15 source hashes and digest
`7a716c08d54b97ad487b0a532253d63312f350137dc63b6c78819da4eb461449`. Relocation with spaces, in-SDK replacements, module integrity,
generated consumer race/vet/build and public scenarios pass; runtime is not
executed and distribution is not certified. All 6174 prepared files remain
unchanged from the preceding committed tray checkpoint; dependency pins/patches
are unchanged.

Internal source review approves all 15 frozen files. One Anthropic Claude
Sonnet 5.5 Prism review has complete 15-file/3-chunk coverage,
raw exit0 and zero findings. Promotion preserves the previously reviewed
clean/base/sole-child/ff-only/source+prepared guards with mechanical phase/base/
result/prose substitutions; their provenance is retained.

The computer-free constraint persists. Packaged authorization/denial, real
presentation/sound/replacement/removal, concurrent ordering, foreground/delegate
interoperation and restart remain live checks. Full Mac/P10, input/accessibility,
sustained performance and public distribution remain open.

[Checkpoint](records/checkpoint.json), [raw evidence](records/evidence.tar.gz)
and [member index](records/archive-index.json) bind this offscreen scope.
