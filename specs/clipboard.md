# Clipboard representations

`app.ClipboardServices` is an optional extension; `app.Services` remains source
compatible. `Application.Clipboard()` returns the extension or `ErrUnavailable`.
Call read/write from workers outside application-owner callbacks, with a caller
context. The application lifetime also cancels pending service work.

The Mac adapter transports exact UTF-8 plain text, PNG bytes, HTML bytes and custom
MIME data in one native pasteboard item. It resolves custom formats using Apple's
UniformTypeIdentifiers MIME registry (macOS 11 or later). Capabilities describe
the implementation on the current OS, not qualified interoperability. PNG is not
decoded/re-encoded, HTML is not sanitized, and TIFF/image conversion is not provided.
Other platform adapters report no rich-data capabilities and return unavailable;
their existing plain-text Services methods retain their previous behavior.

MIME names must be lowercase ASCII concrete media types without parameters or
wildcards. Names are at most 127 bytes, with 1..16 distinct formats per request.
Payloads are at most 8 MiB per representation and 16 MiB in aggregate. Plain text
must be valid UTF-8 without NUL, matching the existing text contract. Binary data
can contain any byte. Go validation checks all lengths before copying payloads;
the Objective-C parser repeats the bounds and refuses MIME aliases that map to
the same native type. Such aliases and unmappable types return `ErrUnavailable`
before pasteboard access. Limits apply to accepted data and framework copies;
Cocoa may allocate/materialize a lazy external provider before its length can be
checked. A framework limit cannot bound that provider's memory or execution time.

Reads require explicit formats, preserve request order and omit absent formats.
All returned representations come from the first pasteboard item, never a mix
of several items. An empty board returns an empty result. The adapter compares
the board's change count before/after obtaining representations and returns
`ErrClipboardChanged` if another writer intervenes. This is a consistency check,
not a clipboard lock. Returned Go bytes are detached and caller owned.

Writes construct and validate one item before clearing/replacing the board. Do
not mutate input slices before the method returns. Cocoa receives its own copy
before native dispatch returns, so canceled caller storage is not read later.
The native callback checks caller/application lifetime immediately before
mutation. Cancellation can race an already executing native call; a returned
cancellation error does not prove the board was unchanged. Clear/write is not
transactional and a native failure may leave an empty board. There is no rollback.

The Mac adapter admits at most four queued/running rich clipboard operations per
process. `ErrClipboardBusy` refuses additional work immediately. Admission occurs
before payload encoding. Cancellation releases the caller, while its credit
remains held until the native block's terminal callback. Shutdown releases all
callers and permanently closes admission, including when the main queue has
stopped; queued native copies remain bounded until their blocks run or the
process ends. Plain-text Mac methods use this same bounded bridge. Requests are
not retried automatically; context errors, closed application, oversize, changed
board, full admission and unavailable transport remain distinguishable.

The bridge uses a bounded little-endian wire: count u32, followed by MIME length
u16, payload length u32, MIME bytes and payload bytes. It rejects trailing bytes,
truncation, invalid text/MIME and alias collisions. Native NSData-to-Go copies
check the full wire length before `C.GoBytes`.

Offscreen tests validate the Go contract/codec, detached results, cancellation,
stalled-queue admission, shutdown and the actual Objective-C parser/MIME resolver.
They never call NSPasteboard. Native compilation establishes no clipboard
runtime acceptance. The explicit `scripts/check-clipboard.py --live` tool adds
a separate general-pasteboard acceptance workflow when the computer is available.
Its `--self-test` uses only a new private named pasteboard, exercising exact
multi-item/empty/binary-flavor restoration and refusal on foreign or unknown
ownership. Injected foreign writes after admitted/empty journal persistence are
refused immediately before clear/write, both by the actual native guards and
the guardian's restoration path. An injected unrepresented board-level flavor
is refused by the capture path. These safety checks precede a general-board run.

The Swift guardian captures complete ordered raw items/flavors before launching
the Go fixture. It refuses unavailable data, non-item legacy state, changing
revisions or unsupported board-level types absent from the complete item-type union,
more than 32 items/32 flavors per item, type names over 1024 UTF-8
bytes/64 KiB total, representations over 8 MiB, payload over 16 MiB and a serialized
backup over 20 MiB. Lazy provider materialization occurs before accepted-size
checks. Original bytes stay local in a 0700 job directory/0600 binary plist;
they are never printed, hashed for evidence or sent for review. Restoration
preserves materialized data, item order and each item's type order, not the original live provider
or pasteboard change-count identity. Cocoa calls have Objective-C exception fences
before returning to Swift; error logs omit system/provider exception payloads.
One derived Cocoa `NSStringPboardType` alias is accepted outside the item union
only when its raw data exactly matches the first item's canonical text. Its
raw bytes are also bounded, backed up and compared after restoration. Arbitrary
unrepresented flavors and differing alias data remain refused before mutation.

Only the `vellum_native_checks` build adds write guards at the public native
transport's clear/write boundary. They require an exact original/owned revision,
a unique test marker or exact nonce-bearing legacy text, and durable admitted,
known-empty and owned journal states. Ownership is rechecked after journal I/O
immediately before clear/write. Known-empty records use the clear call's returned
revision and require matching revision and empty item/type sets. The guardian holds a global test lock and
waits for its recorded child to disappear before restoration or recovery. Unknown
admission/restoration state and foreign changes retain the protected backup and
fail instead of guessing ownership. The guardian removes the backup only after
verified byte-exact restoration. Neither fixture writes nor restoration provide
an atomic OS compare-and-swap; another writer can still intervene between calls.
The guards/API are absent from ordinary compiled builds; current SDK source
payloads can still contain diagnostic source files.

The Go fixture checks text/PNG/HTML/custom MIME bytes, requested order/omission,
detached reads, legacy text and controlled first-item consistency. Native MIME
alias refusal is checked only if the OS registry supplies a usable alias pair.
The independent Swift process compares the final test representations after Go
exits. Separate cases fail or terminate the fixture after an owned write and
require original restoration. Third-party application compatibility, broad
provider/restart behavior and adversarial concurrent-writer atomicity remain
separate from these bounded checks.

Primary APIs: [NSPasteboard types](https://developer.apple.com/documentation/appkit/nspasteboard/pasteboardtype)
and [UTType](https://developer.apple.com/documentation/uniformtypeidentifiers/uttype).
