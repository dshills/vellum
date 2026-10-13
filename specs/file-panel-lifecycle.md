# macOS file-panel lifecycle acceptance

The production `Services.ChooseFile` contract remains a worker call that returns
one selected path or an error; Save chooses a destination without writing it.
Caller cancellation can return before Cocoa acknowledges dismissal. Admission
therefore remains retained until a native terminal callback or loop retirement.

An opt-in `vellum_native_checks` build provides a read-only inventory of precisely
the adapter's tracked `NSOpenPanel` / `NSSavePanel` instances. Counts for tracked,
visible, Open and Save objects are observed on the native main thread, with the
production eight-panel bound. This inventory performs no panel actions and never
selects a URL. It is absent from ordinary and offscreen builds.
It counts tracked objects and does not independently discover a visible panel
that an exceptional cleanup path has already removed from that collection.

`NativeOutstanding` is a separate Go-side sample taken after the inventory
query's own terminal acknowledgement. These two observations are not an atomic
cross-thread snapshot. During an isolated fixture, a visible single panel of the
requested class and one outstanding operation establish the pre-cancel stage.
After caller cancellation, an empty path and `context.Canceled` are necessary
but insufficient: polling with a still-live parent context must also observe
zero tracked/visible/class counts and zero retained operations before the next
panel can be admitted as an accepted lifecycle cycle.

The bounded native consumer alternates Open and Save cancellation, then closes
the application with a real visible panel still active. Its final caller uses
`context.Background`, so application lifetime must produce `context.Canceled`
without relying on caller cancellation. The caller is joined with a fresh five
second deadline, new subsequent service use must refuse with `app.ErrClosed`, and
the main cleanup observer must confirm no tracked panels or retained callbacks.
Partial completed-stage records survive failed worker or native-loop exits.

This qualifies scoped native visibility, cancellation, admission retirement and
normal loop shutdown. It does not qualify physical Cancel clicks, selected paths,
file persistence, native heap reclamation, exceptional AppKit cleanup, concurrent
multi-panel behavior or human accessibility. Actual selection and persistence
have separate [editor acceptance](../docs/reports/macos-native-choosers/README.md).
Physical Finder drag remains a separate external-input workflow.
