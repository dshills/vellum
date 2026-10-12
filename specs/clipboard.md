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
They never call NSPasteboard. Native compilation also establishes no clipboard
runtime acceptance. Real text/PNG/HTML/custom-MIME roundtrips, interoperability
with other apps, alias availability and multiple-item/change-count behavior
remain deferred until the user explicitly frees this computer for live tests.

Primary APIs: [NSPasteboard types](https://developer.apple.com/documentation/appkit/nspasteboard/pasteboardtype)
and [UTType](https://developer.apple.com/documentation/uniformtypeidentifiers/uttype).
