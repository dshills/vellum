# Mac clipboard data checkpoint

Vellum adds optional `app.ClipboardServices` and `Application.Clipboard()` without
changing the existing Services interface. The Mac adapter transports exact text,
PNG, HTML and custom MIME bytes in one item. Reads request explicit formats from
the first item, preserve order, omit missing formats and check the board's change
count. Writes prepare one validated item before replacement. Text follows the
UTF-8/no-NUL contract; PNG/HTML/binary data remain opaque, without image conversion
or sanitization. Custom MIME resolution uses the OS registry, with native alias
collisions refused before board access.

Requests have 1..16 formats, at most 127 ASCII MIME bytes, 8 MiB per representation
and 16 MiB aggregate. Four process-wide credits are admitted before encoding.
Canceled callers return promptly; queued native work retains its credit through
terminal acknowledgement. Shutdown closes admission and releases callers even
if the main queue stops. Mac plain-text methods now share this bounded transport.
Inputs are copied before delayed native consumption; results are caller owned.
Cocoa entry and consumer pools define the native payload lifetime explicitly.

Cancellation may race an executing native call and cannot guarantee unchanged
clipboard state. The native clear/write pair is not transactional and failure
may leave an empty board. Cocoa may materialize an external provider before its
length can be checked; framework data limits do not bound provider time/memory.
These implementation limits are part of the [contract](../../../specs/clipboard.md).

Two root clipboard test functions, seven driver clipboard functions (including
actual Objective-C parser/ownership fixtures), and one independent public API
consumer pass. Tests cover malformed/truncated/trailing wires, unsigned length
limits, invalid text/MIME, payload totals, ordered-subset response validation,
detached results, canceled stalled queues, shutdown/application lifetime,
optional legacy compatibility, injected distinct-MIME/native-type collisions,
and copied payload retention/release across a drained producer pool. The alias
fixture injects type resolution; it does not qualify actual OS aliases.

Root race/vet, the complete final offscreen native-module race suite, driver vet,
public-consumer checks and native workbench compilation pass. The executable was
compiled only, with Go 1.27.2; its recorded local load commands specify macOS26.0
minimum/SDK27.0. This is local compilation evidence, not an older-Mac promise.
Linker duplicate-libobjc warnings are retained. Initial root/native-module runs
failed only where the sandbox refused existing private Unix-socket binding;
authorized headless retries pass. All receipts remain distinct.

The final private SDK verifies all 6914 manifest file identities with trusted
digest `e81814c17f2b9e18cb707b01dd858395f645a588f34fec04b6536313f62b5470`. All 17 source hashes match. The relocated path contains
spaces; all four module replacements remain inside the SDK. External module
verification, generated consumer race/vet/build and full public API scenarios
pass. Native runtime is explicitly not executed and distribution remains not
certified. All 6172 prepared upstream/typesetting files match this checkpoint.

Internal V1 review identifies unsafe u32-to-int conversion on 32-bit hosts and
records Cocoa worker ownership/alias-fixture advisories. V2 checks raw unsigned
lengths before conversion and adds a shared producer pool with delayed-consumer
ownership and resolver-injected alias tests. Independent V2 source review
approves all 17 final files. One Prism run uses Anthropic Claude Sonnet 5.5,
complete 17-file/3-chunk coverage, raw exit0 and zero findings.
No repeated provider review was run.

This checkpoint never reads/writes NSPasteboard, opens native UI, changes OS
preferences or runs live input/GPU/frame performance checks. Clipboard methods
are implemented and statically compiled; actual text/PNG/HTML/custom-MIME
roundtrips, cross-app interoperation, OS aliases, multiple-item consistency and
change-count races remain deferred until the user explicitly says the computer
is free. Other platforms report the extension unavailable while retaining their
existing text services. Full Mac/P10, performance, native human acceptance and
public distribution gates remain open.

[Checkpoint](records/checkpoint.json), [raw evidence](records/evidence.tar.gz)
and [member index](records/archive-index.json) bind this offscreen scope.
