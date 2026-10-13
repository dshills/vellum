# Mac temporary-object lifetime

Go-owned main-queue callbacks do not inherit the temporary pools inside separate
GLFW window/event calls. Native entry points that create autoreleased objects
must establish a local pool; returned or retained state must survive that pool by
explicit ownership or by value copying. Desktop main-thread ownership remains
required, and pooling does not introduce a cross-thread native operation.

The Darwin surface factory captures the content view and refuses a missing view
before allocating a Metal layer. It installs the layer on the strong NSView.layer
property before draining the temporary creation pool. The surface descriptor
borrows that pointer while the native view owns it. Existing normal teardown
releases surface/device resources before destroying the native view/window.
Non-Darwin surface construction retains its prior implementation.

The native input bridge pools attachment, shared text publication, geometry
publication, close, caret/settle and diagnostic replies. Stored text is copied;
geometry buffers are retained before nested pools drain. Go callback values are
copied before the call returns. Returned diagnostic strings are malloc-owned and
freed by their Go callers. Rejected publications still preserve accepted state.
Existing selectors, validation, cancellation and composition semantics remain
subject to their own protocol checks.

Headless ownership fixtures compile the actual bridge/header with inert view and
layer stand-ins and counted temporary objects. They check strong retention,
nil-view refusal, replacement, caller ownership, early returns and final cleanup.
Factory instrumentation is restricted to the process main thread; original
implementations are restored. No window, event loop, drawable, IME setting or
user preference is created or changed. These checks do not prove native graphics
resource deallocation, display behavior or resolution of the previously measured
IOSurface growth. A quiet-host live physical-footprint campaign remains required.
