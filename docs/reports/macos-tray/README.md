# Mac tray and native service lifetime checkpoint

Vellum adds optional `app.TrayServices` and `Application.Tray()` with one Mac
menu-bar status item, title/tooltip/accessibility name, bounded PNG/template icon
and registered-command menu. Existing Services implementations remain compatible.
Updates reuse the owned native object; clearing removes only that item. This
extends the existing windowed lifecycle: final-window close still ends the app.
Capability flags describe the compiled transport, not native acceptance.

PNG input is at most 1 MiB/256×256; complete worker decoding rejects corrupt
bodies before native installation. Detached straight RGBA fits 18×18 points
without distorting aspect ratio. Menus allow 128 entries; labels/enabled state
are owner snapshots, and activation rechecks command availability and current
native ownership on owner drain. Replaced/cleared/shutdown tray and application
menu actions cannot activate stale commands. Accessible names resolve explicitly
or from tooltip/title; unnamed icon-only items are refused.

Chooser, application-menu and tray dispatch share eight process-wide credits,
separate from clipboard's four. Admission precedes encoding; canceled callers
retain their credit until native acknowledgement. JSON is at most 1 MiB, C string
results 64 KiB before Go copying. Worker/consumer autorelease pools explicitly
own detached data. Callback tokens do not wrap. Requests already accepted may
change native presentation despite caller cancellation; no rollback is promised.
The [contract](../../../specs/tray.md) records the exact limits and lifecycle.

Owned backend patch45 registers up to 16 main-thread cleanup hooks. Terminal
cleanup detaches hooks under the lock, runs them in LIFO order and guarantees
remaining hooks/termination despite a callback panic, including legacy nil panic.
The Mac hook closes service admission, cancels owned panels, restores the original
application menu and removes the status item before GLFW termination. Earlier
asynchronous cleanup errors remain observable; driver/model errors are joined.
Callback execution time and broader GPU/panic recovery are outside this contract.

Fifteen new test functions and extended existing service tests cover public
compatibility and bounds, full PNG
decoding/detached pixels, owner capture, actual Objective-C JSON/bitmap/alpha/
aspect conversion and producer retention/release, canceled admission/terminal
ACK, actual bounded C-to-Go reply copying, tray/menu callback retirement and
availability, independent public consumer scenarios, and injected backend
cleanup order/removal/capacity/panic/termination and injected native dismissal exceptions. Eleven preparation regression
tests pass, including exact pin/patch order. No fixture creates NSStatusItem,
opens NSMenu/panels, reads/writes clipboard or changes OS preferences.

Root race/vet, the complete final offscreen native-module race suite, driver vet,
owned desktop cleanup race tests, public consumer race/vet and native workbench
compilation pass. Root/backend/public-consumer receipts precede V5 only where
their source bytes are unchanged. The final V5 native suite/build includes the
menu retirement and native dismissal correction. Earlier checks and one failed test expectation
(disabled-command error reporting) remain distinct; that fixture was corrected.
The executable was compiled only, with Go 1.27.2 and local macOS 26.0 minimum/
SDK 27.0 load commands. This does not qualify older Macs. Duplicate-libobjc linker
warnings are preserved, without treating them as runtime evidence.

The private SDK verifies all 6933 manifest files with trusted archive digest
`258065a19a2970613a70403c7a66a766f1fea6913d4af7c4042d7124f5172e9a`. All 31 final source hashes match. A relocated path with spaces,
in-SDK module replacements, external module checks, generated consumer race/vet/
build and full public consumer scenarios pass. Native runtime was not executed;
public distribution is not certified. All 6174 freshly prepared upstream and
typesetting files match this checkpoint.

Initial internal review preserves the legacy nil-panic, accepted-candidate
ownership/tracking, asynchronous completion ACK and preparation-order findings.
Bounded corrections add explicit panic detection, publish accepted native item
ownership before fallible setters, catch/retire panel completion failures, fence panel cancellation while bounding retained failed panels, fix
patch-order expectations and guard queued application-menu activation. This is
source-level review; no runtime leak is asserted. Final independent internal
review approves all 31 hashes. One Prism run uses Anthropic Claude Sonnet 5.5 with
complete 31-file/5-chunk coverage. Its raw exit1 and one medium
conditional duplicate-ACK finding are preserved. Independent disposition confirms
that the existing Go bridge ignores retired completions, atomically retires each
operation once and never reuses exhausted tokens. No code change or repeated
provider review is needed; no finding remains unresolved. A separate temporary
adversarial Go fixture confirms repeated/concurrent C-reply acknowledgements,
stale-token isolation and exhaustion refusal under the race detector. Its source
and terminal receipt are preserved; final source/index remain unchanged and the
fixture is excluded from the final SDK.

The user's computer-free constraint remains in force. Actual status-item
visibility/object/bar-position retention, menu clicks/keyboard operation,
VoiceOver, dark-mode template appearance, mixed-DPI behavior and cleanup through
real window/loop termination remain deferred. No native UI/input/GPU/frame
performance or sustained-load acceptance occurred. Full Mac/P10, native human
acceptance, performance and public distribution gates remain open; this is one
reviewed implementation checkpoint toward the Mac framework.

[Checkpoint](records/checkpoint.json), [raw evidence](records/evidence.tar.gz)
and [member index](records/archive-index.json) bind this offscreen scope.
