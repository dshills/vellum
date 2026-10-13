# Guarded Mac clipboard acceptance

The opt-in Mac fixture now qualifies the native rich clipboard transport against
the actual general pasteboard. An independent Swift guardian preserves bounded,
ordered raw items and flavors before the fixture starts, then restores their
materialized bytes after verifying test ownership. Its protected original is
never logged, hashed for evidence, reviewed or archived.

## Observed behavior

The race-enabled Go fixture transports exact plain text, PNG, HTML and custom
MIME bytes; checks request order, missing-format omission and detached reads;
exercises legacy text and controlled first-item consistency; and refuses a real
OS MIME alias collision without changing the board. Swift independently compares
the final test representations after Go exits.

Normal completion, intentional fixture exit 1 and forced fixture termination all
restore the original raw items/flavors, including Cocoa's derived legacy string
alias, and delete the protected backup only after verification. All three cases
passed again with the final corrected source and fresh guardian/race binaries.
These runs changed no input-source, VoiceOver or notification settings.

Private named-pasteboard tests exercise the actual native guards and Swift
restoration with foreign writes during both journal I/O gaps. They refuse unknown
admission, a wrong clear revision and an unrepresented board-level flavor, and
preserve multiple items, empty/binary flavors and the known derived text alias.
Stale dummy journal temporaries no longer block durable writes or restoration.

## Checks and review

Native and offscreen race suites pass on the pre-Prism source. The subsequent
changes affect opt-in diagnostics and the test harness; final isolation verifies
the default/offscreen/offscreen-with-opt-in package archives are unchanged and
exclude all 19 diagnostic symbols. Final opt-in vet, fresh native race build,
private safety fixtures and four Python harness regression tests pass.

Prism ran once with Anthropic Claude Sonnet 5.5: complete 12-file coverage,
45,494 input and 13,632 output tokens. Its raw exit 1 and four medium findings are
preserved. Unique mkstemp names fix stale temporary blocking in C and Swift;
cancellation is recorded then propagated; malformed/nonobject summaries produce
a failed result with source evidence. Internal review covers all 13 final files,
including the added regression tests. No second provider run was made.

Historical failures remain in the archive: early compiler/cache issues, attempted
reuse of a native item already associated with a board, strict rejection of the
known Cocoa text alias, and a sandboxed native-service refusal. These were private
checks; the successful native runs used the authorized desktop service access.
Separate frozen maps preserve each reviewed/tested source version.

## Evidence and limits

[Checkpoint](records/checkpoint.json) binds the [archive](records/evidence.tar.gz)
and its per-member [index](records/evidence-index.json). Qualification records the
final source, binary identities, terminal receipts and all 6,174 unchanged prepared
dependency files. The archiver prunes every protected private directory before
traversal and excludes all backup files. Restoration evidence comes from the
reviewed verification path and public receipts; original content stays private.

Clear/check/write is non-atomic with other writers. Restoration preserves
materialized bytes and order, not a live provider or change-count identity.
The three live cases do not qualify guardian crash or outer recovery, third-party
application interoperability, broad lazy providers or adversarial concurrency.
Human IME/VoiceOver, hardware, memory/endurance, full workload performance and
public distribution remain separate. This checkpoint does not complete the Mac
framework or a public release.
