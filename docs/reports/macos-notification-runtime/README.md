# Mac notification bundle and permission acceptance

The optional notification service now has a separate, nonprompting real-Cocoa
acceptance consumer and bounded inventory probe. The ordinary SDK's notification
implementation is unchanged. Default/offscreen builds exclude the new probe and
consumer; diagnostic source files remain part of the repository/source payload.

## Actual host result

Two bare processes refuse settings, the permission method, delivery, withdrawal
and inventory as unavailable. Bundle metadata is checked before those calls;
the permission method refuses before accessing the notification center.

Six private ad hoc bundle lifetimes reach the actual UNUserNotificationCenter.
They observe NotDetermined permission with alerts/sound false, both valid Notify
calls return the permission error, settings remain unchanged, and queried pending/
delivered IDs are empty. Restart queries include the preceding run's exact nonce
IDs under the same bundle identifier. Native dispatch runs on the main queue;
sampled outstanding operations are zero before cleanup. Scoped withdrawal and
observed empty inventories precede process exit, terminal native callbacks retire,
and the closed public service refuses another settings operation.

The final three processes use the corrected race-enabled consumer: bare refusal,
packaged refusal and restart. The preceding five runs are separately identified
as four/six-source history. The probe and production adapter bytes are unchanged
through the lifecycle correction. Bundle metadata binds pre-sign source binary,
actual signed executable, APPL information and private ad hoc identity. No public
identity, notarization, permission prompt, delegate, sound or preference change
was used. General clipboard contents and input/VoiceOver settings were untouched.

This does not qualify Denied state, an authorization prompt, Authorized/Provisional
delivery, replacement or removal of an existing item. Withdrawal of absent owned
IDs confirms dispatch/observation only. Empty inventories and outstanding counts
are sampled observations, not concurrent ordering or future OS state. Human
visible presentation, sound and foreground/delegate policy remain open.

## Bounds, tests and review

The opt-in probe accepts at most 32 valid public IDs and only the isolated check
bundle namespace. It returns only those reserved identifiers, keeps pending and
delivered reads separate, caps input at 8192 bytes and scans each OS array at most
4096 entries. A conservative escaped-text budget precedes selected-row collection;
aggregate reply JSON is at most 64 KiB. Overflow and duplicates refuse the complete
reply, never truncate. Two data-object/canceled/shutdown test methods pass three
times under race without contacting the center. They cover malformed/NUL IDs,
counts, oversized aggregate bodies, duplicate/native reply refusal, detached
content, unrequested IDs, required fields and canceled/nil/stopped admission.

The first pure log records invalid inventory decoding and a nil-context panic.
The decode failure was diagnosed as a boxed C Sound comparison encoding a number;
explicit BOOL boxing and nil-context preflight corrected the observed failures.
That diagnosis is an inference, not a field-specific log observation. Raw failure
and source hashes remain historical; its old complete source bytes were not saved,
so it is not a replayable original-source qualification.

The complete opt-in adapter race suite and vet pass before the Prism correction.
Only command lifecycle/spec and a new command test change afterwards. The affected
command race regressions pass three times, affected vet passes, and a fresh race
build plus three actual native processes verify the final source. The nine new
diagnostic symbols and their Go/Objective-C files are excluded from default,
offscreen and offscreen-plus-opt-in compilation. Final isolation metadata is bound;
the initial detail filename was reused by that producer. Initial raw exit/stdout
remain historical, with no earlier/final archive-equality claim.

Prism runs once with Anthropic Claude Sonnet 5.5: complete six-file/two-chunk
coverage, 22,057 input and 4,177 output tokens, raw exit 1 and one medium finding.
An early native-loop exit could outlast the five-second worker wait and lose its
report. The corrected consumer waits 16 seconds, longer than fresh 15-second
cleanup, and preserves a late outcome or emits a scoped failure. Both early-loop
paths explicitly mark cleanup/delivery/withdrawal unqualified. Deterministic tests
cover normal completion, late worker detail/error and no reply. They do not
qualify an actual fatal Cocoa-loop failure or recover its notifications.
Internal correction review approves all seven final files and the finding's fix;
no second provider call was made.

## Evidence and remaining acceptance

[Checkpoint](records/checkpoint.json) binds the [archive](records/evidence.tar.gz)
and [member index](records/evidence-index.json). It includes all 22 wrapper receipts,
raw stdout/stderr, source snapshots/manifests, separate original/final reviews,
Prism disposition, scoped fixture IDs and package metadata. App bundles/binaries
are pruned before archive enumeration; metadata retains their hashes. All 6,174
prepared dependency files remain byte-identical to the preceding checkpoint.
Selected source/binary evidence is not a complete compiler/cache/SDK closure.

Checks ran inside the authorized fifteen-hour quiet window. Authorized delivery/
replacement/removal, permission changes, visible presentation, foreground/delegate
interoperability, real early-loop failure/recovery, native memory/endurance and
broader Mac hardware/performance/distribution remain open. This bounded runtime
acceptance does not complete the Mac UI framework or qualify public release.
