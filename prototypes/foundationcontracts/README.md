# Foundation contracts prototype

A separate follow-up to [foundationchecks](../foundationchecks/README.md). It layers patch 0007 on the six existing patches and adds an independent bounded mailbox. See [results](../../FOUNDATION_CONTRACTS_RESULTS.md). G1 remains open; no workbench expansion or production API commitment.

## Reproduce

Use Go 1.27.1 on macOS arm64 and the pinned sibling prototypes. From this directory:

```sh
./run.sh
. ./env.sh
build/contractcheck -nogui -messages 100000 > evidence/contracts.json
python3 fixture.py
build/contracts-fixture -input-probe -exit-after 14s -report "$PWD/evidence/native-exact.json"
python3 verify.py
```

`run.sh` creates an owned module copy from the pinned cache, verifies all patch inputs, applies nativeinput 0001–0004, foundationchecks 0005–0006 and local 0007, runs vet and focused race tests, then builds the diagnostic CLI. The GUI fixture imports the unchanged sibling native bridges and copies their pinned AccessKit library. Build that sibling first if the library is unavailable. Desktop execution requires permission for Cocoa/GLFW; it changes no OS preferences. `-nogui` is required for standalone diagnostic CLI initialization. Generated modules and upstream copies stay disposable.

## Exact buffer

Patch 0007 adds `lines.NewLinesExact()`. Its lifetime policy retains all final empty lines and joins current lines without synthetic trailing newlines. Subsequent SetText/SetTextLines, edits and Undo/Redo retain that policy. `NewLines()` and other existing constructors retain their previous behavior. This contract assumes valid UTF-8 input; the underlying rune conversion does not preserve invalid byte sequences. Text normalization/formatting commands remain explicit transformations outside this promise.

Tests cover blank/trailing lines, Unicode, CRLF, repeated replacement, Insert/Delete/Undo/Redo, actual editor backspace over newline and emoji, 800 deterministic edits compared to an independent rune/line reference, and 100 full diff/history pairs plus selected-diff tests. Exact diff application snapshots the lines, applies only supplied operations, then records a grouped full-text replacement. It fixes final-empty-line changes at the cost of incremental formatting/caret preservation and full-content allocation. Temporary file comparison buffers use the same policy; complete file-open/revert workflows remain unverified because of the recorded upstream lexer/render issues. The native fixture replaces only its multiline-buffer constructor; its marked/commit/cancel protocol probes remain the previous native client, not real Japanese/Chinese IME or VoiceOver evidence.

## Mailbox ownership

`mailbox.New(Limits)` allocates a fixed ring. `TryReserve(size)` obtains both a slot and queue-owned payload credit before external mutation; it never waits for capacity. Commit copies exactly the reserved byte count. Reservation order defines FIFO delivery, even when commits arrive out of order. A forgotten reservation blocks following messages; the owner must cancel it on failure. Canceling a middle ticket releases payload bytes but its ring slot becomes reusable only when the head advances past it.

One delivery is in flight at a time. `Ack(ID)` releases its credit after processing; simply taking a message does not free capacity. Delivery bytes are borrowed until Ack, and callers must not retain them afterward. Queue bounds cover queue-owned payload lengths and fixed slots, including reserved/in-flight messages. They do not bound caller-held payloads, ticket objects, allocator rounding, native objects, goroutines or a downstream queue. Limits are trusted prototype configuration, not a public untrusted configuration parser.

`Submit(ctx, payload)`/`Next(ctx)` may wait on dedicated worker/consumer goroutines; they must not run on the native main thread or UI event pump. Cancellation before admission returns an error; cancellation after acceptance does not retract the message. Close stops admission, cancels uncommitted tickets and retains committed/in-flight deliveries for draining. Native mutation/commit and shutdown must be serialized by their view owner.

**This mailbox is not connected to Cocoa or Cogent.** The existing native client mutates state before sending its callback, so substituting a failing TryReserve there would lose already accepted input. A production route needs preflight in the native input mutation path, defined visible refusal/pause behavior before acceptance, cleanup on native failure, serialized teardown, and a bounded/rearmed toolkit wakeup with bounded UI draining. Coalescible caret/progress state needs its own bounded policy. Never silently drop committed text/actions or block Cocoa while waiting for the UI.

## Evidence scope

`diagnostic-before-exact-diff-fix.txt` preserves the initial failing trailing-line diff probe; the corrected path and selected-diff/history matrix subsequently pass.

`contracts.json` is a single synthetic run with small eight-byte messages, not application latency. Six mailbox race tests include concurrent producers, ownership, stale IDs, cancellations, shutdown and in-flight credit. Portable Windows/Linux mailbox test binaries compile with CGO disabled; no native platform runtime is tested. The broader upstream lexer/render failures recorded in foundationchecks remain unresolved; the runner selects relevant tests only. No font-memory, accessibility, candidate-panel, GPU-lifetime or distribution gate is closed here.
