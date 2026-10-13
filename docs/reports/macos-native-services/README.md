# Mac native menu and tray acceptance

An opt-in `vellum_native_checks` build now checks the adapter's installed AppKit
application menu and NSStatusItem, native selector routing and terminal cleanup.
The ordinary SDK excludes its Go diagnostics, Objective-C probes and symbols.
The development guide documents the separate desktop fixture.

## Observed behavior

The fixture installs a two-command application menu, invokes its installed action,
changes owner availability before refreshing native presentation, and verifies
that owner dispatch refuses the newly unavailable command. Refresh presents the
disabled state. Replacement rejects an actual retained old target/item callback;
the new installed action reaches the current command.

The native tray presents the supplied title, name, tooltip and template icon. A
4×2 straight-alpha bitmap fits 18×9 Cocoa points and its menu contains the expected
commands and separator. Updating reuses the installed native item. Replacement
and clear reject retained old callbacks; clear retires production ownership.
After the first actual Cocoa window disappears, the second remains visible and
the application-wide tray action still reaches the owner.

Terminal main-loop inspection runs after production service retirement and before
GLFW termination. It observes the original application menu restored, no owned
tray/targets and stopped callback admission with zero pending operation/menu/tray
records. Test-retained stale targets/items are released before that inspection.

## Checks and review

One ordinary selector run and two race-enabled 100-cycle runs pass. The last race
run tests the final source, including the defensive JSON correction. These are
functional create/clear cycles, not a native heap or endurance qualification.
Root, native adapter and offscreen race suites pass. Vet passes for root, default,
offscreen and opt-in builds; affected opt-in vet was repeated after correction.
Default/offscreen package metadata and symbols exclude all seven diagnostics;
their package archive hashes are unchanged by the opt-in-only correction.

Internal review approves all eight final source/contract files. Prism ran once
with Anthropic Claude Sonnet 5.5, complete eight-file coverage in two chunks,
18,106 input and 3,590 output tokens. Its raw exit 1 and one medium finding
are preserved: a possible nil UTF8String before strdup in the diagnostic encoder.
The encoder now returns NULL through the existing failure path; internal review,
a fresh race build, live stress check, opt-in vet and isolation check cover that
change. No second provider run was made.

The first live diagnostic failed because a boxed C comparison encoded a JSON
number where Go expected a Boolean. Explicit BOOL casts corrected the diagnostic
producer. That initial run also overlapped a diagnostic-only source edit and is
historical evidence. Original and final frozen source maps are kept separately.
The first isolation producer reused its wrapper receipt filename, so its detailed
metadata was overwritten; the corrected producer and final receipt keep distinct
names. Original stdout still records the completed isolation assertions.

## Evidence and limits

[Checkpoint](records/checkpoint.json) binds the [archive](records/evidence.tar.gz)
and per-member [index](records/evidence-index.json). Qualification identifies the
eight final files, original/final receipts, race binary and all 6,174 unchanged
prepared dependency files. Raw failed/intermediate diagnostics remain archived.

Checks ran within the authorized fifteen-hour Mac testing window. They changed
no preferences, clipboard contents, notification permissions or user files.
Installed NSMenu actions and retained old target callbacks were invoked in-process.
Physical pointer/keyboard menu clicks, bar-position pixels, appearance tint,
VoiceOver, mixed-DPI screens, heap deallocation and endurance remain separate.
Rich clipboard and notification runtime checks also remain open. This bounded
checkpoint does not complete the Mac UI framework or qualify a public release.
