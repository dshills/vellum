# Mac native autorelease ownership: headless correction

Owned patch 0047 replaces the desktop Darwin surface factory with a locally
pooled shim. It captures the content view, refuses a nil window/view before
creating a layer, installs the layer on the strong NSView.layer property, and
returns its borrowed pointer after temporary drainage. Existing normal surface
teardown still precedes view destruction. Non-Darwin delegates to prior wgpuglfw.
The source provenance and Apache-2.0 license are retained; no module-cache files
are edited. Fresh preparation verifies 6,181 files: exactly one previous core
file changes, and three factory/header leaves are added.

Ten input bridge/shared-publication paths now establish local pools. Independent
source review verifies that removing only those wrappers and two ownership
comments reproduces the parent byte-for-byte. Validation, callbacks, retained
state, geometry sort/copy/release and strdup statements remain unchanged. Stored
text copies and retained geometry buffers survive nested pools; Go callbacks
copy UTF-8 values before returning. Malloc replies survive native pool drainage.

## Exact final headless checks

The actual prepared layer header and actual input source compile in isolated,
bounded Objective-C fixtures. Faithful strong-property NSObject stand-ins and
main-thread-only sentinels verify reference ownership; original factory IMPs are
captured before replacement and restored on cleanup. The layer fixture passes
nil-window/no-view refusal, 64 creations/replacements and caller-held references,
then reports zero temporaries/live stand-ins and 64 retirements. The input fixture
completes 100 of 100 cycles with 901 observed factory calls, zero temporaries and
zero views after close. It covers committed text, geometry copying/retention,
early refusal, retained caller data, caret/settle and owned inactive replies.
Admission callbacks are stubs; composition remains unmarked. Other diagnostic
probe/query/environment pool bodies are source-reviewed but not all executed.

The final source adapter offscreen race suite, native-consumer compile and all
11 preparation tests pass. Five final receipts bind all 14 selected source files
and all 6,181 prepared identities, with no source/prepared changes during checks.
Compilation retains the existing duplicate -lobjc warning. No native window,
NSApplication/event loop, GPU drawable, IME/input source, clipboard or user
preference is created or changed. Refusal beeps are replaced only inside the
headless fixture translation unit; production behavior is preserved.

## Failures and review remain distinct

The original normal-view factory baseline fails with an undrained PoolWitness;
the original input baseline fails on attachment before its loop. Both raw exits
remain 1. A first fixed-layer fixture also exits 1: it incorrectly assumes
immediate real CAMetalLayer deallocation, despite observing zero temporaries.
CoreAnimation can defer native cleanup; the corrected ownership oracle uses
counted NSObject stand-ins. Its input fixture passes 100 cycles. That failed
surface oracle is preserved and is not an adopted-runtime failure or memory pass.

The final baseline layer fixture fails the newly checked nil-view invariant first
(two unowned stand-ins); its input fixture records zero completed of 100 requested
cycles. Neither historical failure qualifies a completed loop. Intermediate V2
headless results retain their prior source/input hashes and are not relabeled as
final checks. The archive preserves ten raw receipts/logs and version groups.

Internal review approves 14 exact source/spec/pin/license files and four decoded
patch outputs. One Prism review with Anthropic Claude Sonnet 5.5 selects all ten
implementation/test/registration/doc leaves, covers them in two chunks, returns
exit 0 with no findings, and uses 22,910 input / 3,085 output tokens. Four static
pin/provenance/copied-license leaves are verified internally. The indexed archive
retains source maps, original/v1 bridge/header fixtures, review results and raw
process records. Binaries, caches and desktop snapshots are excluded. This covers
selected source/all prepared identities, not a full compiler/cache/SDK closure.

## Live memory gate remains open

These checks prove temporary and retained-reference behavior under inert fixture
contracts. They do not prove actual CAMetalLayer/drawable/IOSurface deallocation,
preedit/IME behavior, display presentation or resolution of the measured native
physical-footprint growth. No quiet-host live process runs in this phase. Further
live/input/performance work remains deferred while the user uses the computer.
Actual physical-footprint/native-GPU acceptance, full P11/endurance, latency,
human interaction, cross-display and public distribution gates remain open.
