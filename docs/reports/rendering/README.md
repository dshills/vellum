# Rendering hardening acceptance

The race-instrumented macOS native rendering smoke completed twice, including
normal process shutdown. Both checks reported zero paint wakeups across a one
second quiet interval, an owner update after idle, live resource rebuilding,
resize and minimize/restore. The checks do not establish a sustained CPU budget,
frame/input p99, physical GPU-reset recovery or cross-display DPI acceptance.

The live checks exposed and drove fixes for asynchronous raster composition,
canvas resource ownership and acquired surface-texture lifetime. Patch 0017 is
owned by this driver, with input/output hashes and synthetic regressions for
composition ownership, stale direct sources, bounded surface retries and shared
device-loss subscriptions. Loss notification fails the application with an
explicit ErrDeviceLost; automatic physical-device recreation remains open.

No input sources, VoiceOver or system preferences were changed.
