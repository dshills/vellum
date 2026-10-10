# Platform services and qualification

Vellum implements Windows and Linux service adapters while keeping target runtime
qualification separate from source compilation. Services keep the application
worker-only contract in `app.Services`; caller or application cancellation stops
pending operations. OS input-source and accessibility preferences are never
modified by these checks.

## Implemented services

| Platform | Clipboard | Single-file chooser | Command menus | Target runtime evidence in this checkout |
| --- | --- | --- | --- | --- |
| macOS | Existing Cocoa text transport | Existing Cocoa panels | Existing process native menu | Existing macOS foundation evidence |
| Windows | GLFW Win32 clipboard on native main queue | Windows PowerShell 5.1 STA process using WinForms OpenFileDialog/SaveFileDialog | Cogent menu bar per native window | Not executed on this macOS host |
| Linux X11 | GLFW X11 clipboard on native main queue | GTK chooser through Zenity | Cogent menu bar per native window | Not executed on this macOS host |
| Linux Wayland session with XWayland | Same compiled X11 backend; must verify XWayland explicitly | GTK chooser availability depends on desktop | Same per-window bar | Not qualified |
| Native Wayland | No Vellum qualification | No native Wayland guarantee | No native Wayland guarantee | Blocked |

Windows requires `powershell.exe` with WinForms in an interactive desktop session,
a cgo C toolchain and a functioning WebGPU backend. Linux requires X11/XWayland,
the GLFW build libraries, a functioning Vulkan/WebGPU adapter and `zenity` on PATH.
Missing helpers return `app.ErrUnavailable`; failed transports preserve errors.
These adapters provide native chooser widgets and framework command menus inside
OS windows. They do not claim Win32 HMENU or Linux global-menu integration.

## Transport and ownership

Clipboard work uses the pinned desktop main queue with context/shutdown selects.
Each callback checks cancellation again before executing, and results travel over
a private buffered channel: no canceled caller storage is later mutated. A native
clipboard call already executing can finish while cancellation is reported.
The platform text flavor can be absent; `FormatUnavailable` on a read returns
empty text. GLFW’s Go wrappers can log and suppress platform errors, so each
operation clears stale C error state first, then reads the thread-local native
error slot inside that same main-queue callback. Current-operation platform
failures remain typed errors; a write never accepts a missing-format error.
Clipboard restoration in the optional smoke preserves plain text only; use it
on a disposable desktop if other clipboard formats matter.

Chooser operations use `exec.CommandContext`, bounded stdout/stderr and direct
argv or stdin JSON. Pipe draining has a one-second upper cleanup delay if helper descendants retain stdout/stderr after the direct helper exits or is canceled. User titles/paths never become shell or PowerShell source.
The fixed PowerShell implementation is UTF-16LE Base64; data is UTF-8 JSON.
Extension filters accept ASCII alphanumeric, hyphen and underscore extensions to
avoid glob/filter metacharacter reinterpretation. Linux removes only the single
output delimiter newline, preserving embedded filename whitespace. A GTK chooser with cancellation exit status and empty diagnostics or narrowly recognized module/theme warnings returns `app.ErrCanceled`; unknown diagnostics and display/startup failures retain their errors. Diagnostic output retains an 8 KiB prefix and drains the remainder. Stdout enforces its 1 MiB limit through the actual process pipe; the writer has no promoted methods that can bypass that limit. Context cancellation/deadlines preserve their
context errors; a stopped native main queue returns `app.ErrClosed`. Save mode selects a path and does not write a document.

Menu groups/items are copied into an immutable process snapshot. Each window
mounts and refreshes its own bar under that window's render lock and the UI owner;
no foreign window tree is mutated while servicing another window. Menu callback
activation posts to the application owner and rechecks command availability.
Disabled state and labels are captured by `SetMenu`, matching the macOS contract.
Closing windows releases mount records; shutdown releases all menu snapshots.

## Qualification checks and CI

Run `drivers/cogent/prepare.sh`, then:

- `scripts/check-platform.sh` runs root/native tests, vet and builds the smoke.
  Its success reports build/unit evidence only.
- `scripts/check-platform.sh --runtime` opens two temporary windows, verifies GPU
  rectangle pixels and six native resize/readback cycles across both windows, installs menus, reads the clipboard, cancels a chooser request after one second (panel display readiness and selected paths require acceptance),
  closes the first window and verifies surviving-window pixels/lifetime isolation,
  then verifies final-window shutdown. Missing services or a missing GPU fail.
- `--clipboard-write` additionally checks Unicode text roundtrip and restoration.
  This is optional and is enabled only on the disposable Linux CI desktop.
- `--qualification` performs the runtime checks, emits JSON qualification gaps,
  then returns exit 2 while required target adapters/acceptance remain missing.
- `scripts/check-platform-fixture.py` compiles both service implementations against
  the current host's pinned Cogent APIs in temporary modules and exercises menu
  ownership. This is explicitly not a cross-build or target-runtime result.

`.github/workflows/platforms.yml` defines Linux X11/Xvfb/Mesa runtime fixtures,
Windows hosted cgo build/unit checks, and an explicitly dispatched interactive
self-hosted release gate on a manually selected Windows/Linux runner with an interactive desktop. The workflow grants its token read access to repository contents only. Merely adding the workflow does not mean it ran.
Environment session metadata is not proof of compositor or display-server support.
All native executable checks need an actual target and display; portable helper
cross-compilation cannot replace that evidence.

## Remaining release gates

Windows still requires a native TSF/IME composition adapter and UI Automation
bridge. Linux requires IBus/Fcitx composition and AT-SPI bridges; native Wayland
needs a backend and qualification. These interfaces must expose Vellum's bounded
ordered text edit/composition stream and semantic/focus actions, not synthesize
characters from key codes. Existing GLFW character/key input provides basic text
input only. No compile check establishes composition or screen-reader support.

Both targets require real GPU hardware, mixed-DPI resize/minimize/restore,
clipboard interoperation with another app, selected open/save paths, keyboard menu
activation, multiwindow churn and screen-reader acceptance. Full release gating
fails until those adapters and checks exist; this phase supplies concrete services
and repeatable runtime checks without recording unsupported claims.

Primary references: [PowerShell executable parameters](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_powershell_exe?view=powershell-5.1),
[Zenity option implementation](https://github.com/GNOME/zenity/blob/master/src/option.c),
and the pinned GLFW/Cogent source used by the driver.

The smoke records progress behind a mutex and returns detached evidence slices, including on timeout or driver failure. Reporting does not depend on the worker having completed. When a worker result is already available after the native run ends, diagnostics preserve both worker and native failures. A failed native run never waits on unfinished worker work; a successful native run waits for the worker result or caller cancellation. Chooser filter validation is shared across Windows and Linux before a helper can start.

Shared GPU qualification includes owned patch19 cache replay admission and cross-window Snapshot ownership. Borrowed canvas texture batches use fresh composition; Snapshot reads the viewport already captured on the owner after layout and never reads another window's concurrently changing widget bounds.

Resize qualification requires an observed native pixel-size transition to each
requested size, plus a completed canvas viewport change on every resized axis.
Native geometry is read under the originating window mutex outside the application
owner; canvas allocation is read on the owner. Each resize has a five-second bound
within the overall smoke timeout. Ignored/clamped requests or an already-matching
native size fail qualification explicitly; they never count as exercised transitions.

Before owner-posted GPU readback, both initial and later windows must publish the
actual native canvas Show event. Service availability alone does not establish
that the first window's native scene topology has completed construction.
