# Scoped native temporary ownership

Patch 0047 routes desktop surface creation through a driver-owned Darwin shim.
Its local autorelease pool covers layer creation and assignment. A captured
content view must exist before creation; NSView's strong layer property owns the
borrowed descriptor pointer after the pool drains. Normal surface teardown still
precedes native view destruction. Non-Darwin calls delegate to unchanged wgpuglfw.
The module cache is not edited. The source shim's webgpu provenance and Apache-2.0
license are retained beside the patch, and input/output/patch hashes are pinned.

The adapter input bridge separately adds local pools to ten Go-callable paths or
their shared publication helper. Existing callback, validation, state retain/copy,
release, geometry-sort and malloc-return statements are preserved. Geometry
buffers are retained before nested publication and outer pools drain. Text copies
into retained state; Go callbacks copy UTF-8 values before returning. Diagnostic
char replies use strdup before pool drainage.

Run `python3 native/patches/autorelease/check.py` from drivers/cogent after prepare.
It compiles the actual prepared layer header, input source and accessibility source. Headless
fixtures use faithful strong-property NSObject stand-ins and scoped factory
swizzling in an owned process. Sentinels measure main-thread temporary retirement;
original IMPs are captured before replacement and restored on cleanup. The layer
fixture covers nil window/view, 64 creations/replacements, caller-held references
and final retirement. The input fixture covers 100 text/geometry publications,
early refusals, copied/retained buffers, caret/settle and an owned inactive reply.
An ARC accessibility fixture uses actual attach/detach on an inert NSView and an
NSObject window stand-in. It verifies weak element/state expiry, registry removal
and restored view class before an enclosing pool drains across 100 × 1000 nodes.
The original unpooled teardown fails the first-cycle ownership assertion.
It reports requested and completed counts separately. Refusal beeps are silenced
only in the fixture translation unit. There are no windows, NSApplication/event
loops, GPU drawables, input-source changes or user-preference writes.

Callbacks in the input fixture are admission stubs; composition is unmarked.
This is temporary-object ownership evidence, not complete preedit/IME/geometry
presentation, actual CAMetalLayer/IOSurface retirement or physical-memory proof.
CoreAnimation may defer real layer deallocation. Physical GPU allocation attribution
and broader native memory acceptance remain separate from these fixtures.
