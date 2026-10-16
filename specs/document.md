# Document persistence and close requests

`document.Session[T]` is a portable, UI-owner API for asynchronous Save, Save As
and Open. It binds one application window, one history journal, a typed codec,
a bounded worker-safe store and an optional completion callback. `New` requires
an already registered live window; session methods, history and model access are
owner-only (`Do`, `Drain`, native callbacks). Driver startup evaluates views only
after the application has finished registering its session and commands.

## Ownership, completion and cancellation

Capture runs on the owner and must return a detached immutable snapshot. Encode
and Decode run on a worker; Install consumes the validated decoded model on the
owner. Install must be atomic on error; a panic invalidates the journal and follows
the application's owner-panic shutdown policy. Codec callbacks must not reenter
Do/session/history. The application is responsible for bounding snapshot/codec
allocation and runtime. The session checks payload bytes after encoding and before
decoding; this is not a hard heap or CPU limit on arbitrary codecs.

One operation stays busy until its completion executes on the owner. IO and
chooser calls never block the owner. Completion reserves a separate bounded application ticket before submitting its
worker, so ordinary queue saturation cannot block final result admission.
`Config.Executor` injects real or manually stepped work. Submission refusal
restores busy/action/operation state and refunds the ticket. State exposes
WorkerPending and OwnerPending separately; admission remains distinct from
owner execution. Completion uses the window lifetime for delivery, so operation
cancellation still clears busy on a live window even under mailbox pressure.
Window closure cancels work and discards late installation/history/user callbacks;
it does not cancel other windows. User codecs and stores must cooperate with
cancellation; OS filesystem calls and encoding already in progress may complete.

OnComplete runs after state/history update and busy release. It may start another
operation or force `CloseWindow(ref)` using its retained opening receipt; it must
not block or nest Do. Worker panic/Goexit
reports failure and releases busy. A successful Save acknowledges only its captured
journal revision. Edits during encoding/writing make that acknowledgment stale;
written bytes may be older, dirty status remains and the current path is unchanged.
Open refuses dirty models without explicit DiscardChanges, and replaces history/model
only while captured revision and invalid state still match. Failed chooser/read/
decode/install/cancellation preserves the current model, history and path. Successful
Open uses `Journal.Replace(true, install)`, releases captures and creates a new
revision identity. Applications also advance their domain/projection epochs so
old workers cannot become current again after loading identical contents.

## File store

Default FileStore accepts absolute valid UTF-8 paths without NUL, up to 1 MiB,
and limits of 1..256 MiB (64 MiB default). Reads are capped at limit+1 and require
regular files. Unix nonblocking open prevents a FIFO from blocking before that
check. Read symlinks may resolve to regular files; save symlinks/special targets
are refused. Parent directories are trusted application/user-selected paths.

Writes use a sibling temporary file, existing regular-file permission bits
(or 0600 for a new file), file sync/close, atomic rename and parent-directory sync.
Pre-rename failure leaves original bytes and removes the temporary file.
`Result.Written` reports replacement separately from success; directory sync,
cancellation or stale revision can fail after replacement. Only Err == nil permits
save acknowledgment or a Save-and-close continuation. This is no filesystem CAS,
external-edit detection, backup, metadata/ACL preservation or power-loss guarantee.
Concurrent external writers require an application conflict policy. Windows/Linux
runtime filesystem behavior is not qualified by macOS checks.

## Window close policy

`Window.CloseRequest` runs synchronously on the UI owner, outside app.mu. Return
false to keep registration, window context and native widgets while displaying an
asynchronous prompt. `CheckCloseWindow` checks permission without teardown for
native drivers. `RequestCloseWindow` checks and then closes for application
commands. Recursive policy requests return ErrClosePending. Recheck cancellation
before acknowledging policy; application Close/CloseWindow force teardown without
prompting. These force APIs are suitable only for explicit confirmation, failure,
lifetime cancellation or test timeout. A clean accepted native close proceeds to
existing body cleanup; Scene.OnFirst rejection precedes that cleanup. Arbitrary
later native extension Close veto/rollback is outside this policy contract.

## Workbench consumer

Commands file.open (Meta+O), file.save (Meta+S), file.save-as (Meta+Shift+S) use the
same session. An unsaved close or Open shows Save, Discard, Cancel, and Cancel
operation while busy. Save closes/continues opening only after successful current
revision acknowledgment; failure leaves the prompt and unsaved native window.
Cancel retains the model; Discard/Open retains it until a successful replacement.
Repeated close requests share one prompt. Closing after an ordinary pending Save
still requires an explicit decision. CLI smoke timeouts force close deliberately.

Version 1 .vellum files encode 100,000..1,000,000 component records, exact first-N
ordered component rectangles (canvas limit 10,000), and ordered file references.
Rotation/package, canonical component IDs, geometry, reference paths, UTF-8,
version, unknown fields and trailing data are checked before installation. Tree
projection construction runs with decode on the worker. UI selection, pan/zoom,
docking layout, task status and source revision are transient. Table projection
reset and source epoch advancement invalidate results from the preceding model.

Race tests cover revision changes, mailbox pressure, cancellation, close lifetime,
installer errors/panics, save failures and owner completion chaining. The exported
SDK includes document and is tested after archive relocation/source deletion.
`go run -race ./cmd/workbench -rows=100000 -tree-rows=1000 -document-smoke` exercises
real native CloseReq, Cocoa accessibility dialog actions and temporary file IO.
Its deterministic chooser does not certify actual file-panel selection or human
accessibility workflows. See [phase evidence](../docs/reports/macos-document/README.md).

## Real macOS panel acceptance

The ordinary independent editor has externally verified Open, Save As, and both
native Cancel routes on uniquely identified owned temporary files. Successful
Save As creates exact expected bytes at the selected absolute destination; native
cancel creates none and preserves document text/path label. Controls recover,
the untouched second window survives the first native close, and both windows
close normally with process exit 0. See [receipt and scope](../docs/reports/macos-native-choosers/README.md).
This is real Cocoa selection plus FileStore persistence, separate from injected
chooser/model smokes. It does not expose private session revision/path telemetry
or qualify menu, clipboard, Finder, human input or performance acceptance.

The qualifying `live-v20` receipt compares both files to immutable original
UTF-8 bytes, binds the unchanged build inputs, and records 11 completed stages.
The external client preserves critical window-read errors and every attempted
command, bounds retries before action admission, and checks foreground focus
before file-panel shortcuts. Those checks and global event posts are not atomic.
Eight fault probes cover harness failure records, retry bounds and byte mutation;
earlier harness generations do not replace the current qualification.
