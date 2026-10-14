# Native Breadcrumb selector acceptance

The new `drivers/cogent/cmd/breadcrumb-smoke` consumer checks the existing public
controlled Breadcrumb against actual Cocoa objects in one owned window with requested logical size 720×480.
It is enabled by `vellum_native_checks`; normal/offscreen builds provide a refusal
stub. The application and its worker have bounded cancellation and native cleanup.

The positive run passes 12 stages: initial, reject folder, disable/enable path,
stale path, restore path, accept folder, accept home, reset path, disable/enable
ancestor, and read-only. Three callbacks comprise one rejected and two accepted
navigation proposals. The current location exposes AXStaticText and literal
`<Current>` naming. Ancestors expose AXButton, measured frames and matching hit
tests; decorative separators are absent. Rejection keeps ancestor native identity,
accepted paths replace current roles, absent actions refuse, and a removed
ancestor gets a new native identity after reset. Disabled path/item and read-only
or current-location press attempts refuse.

The stale-stage native selector is dispatched while an owner callback is claimed;
ordered domain IDs change before queued work settles. The accepted callback count
remains unchanged. Removing the component's ordered-path loop alone still passes
this live fixture. That non-detecting fault is retained; the native check does not
isolate the component guard or establish which earlier guard refused work. The
prior [component checkpoint](../macos-breadcrumb-controls/README.md) separately
contains an inert queued-path fault that detects that guard's removal.

A second isolated UI fault turns the current item into a button. The actual Cocoa
snapshot publishes AXButton, name `<Current>`, nonzero object identity and a
124×48 frame. The initial expected AXStaticText stage fails after its bounded
five-second wait. This is a meaningful role fault; it is not a compile-only failure.
Both the positive and fault apps report main-thread cleanup, stopped callbacks,
zero native outstanding/accessibility windows/pending callbacks and empty titles.
The fault is never applied to repository production sources.

Portable model race, native build and native vet are the selected code gates.
Runtime source is `87e7de1`/6,216 leaves plus the four new command files; documentation
is added after runtime, with all consumed code unchanged. Internal review corrected a failed stale-stage report that always recorded one
admission. The final consumer derives the count from the actual admission result.
Its rebuilt positive again passes all 12 stages with identical observation bytes.
An actual missing-ID selector refuses admission; the corrected failed stale stage
records zero and the app completes native cleanup. The first report-fault overlay
failed compilation from an unused target and never launched a native app; its
corrected version is the observed refusal. Earlier positive/path-only/current-role
receipts retain their initial source version and are not relabeled as corrected.

The owned wrapper holds
the shared live-test flock, checks return-stop files, enforces a 75-second watchdog
and 1 MiB output bound, and kills/reaps the app group on every exit. It changes no
clipboard contents, input source, global preferences or VoiceOver setting.

This is actual in-process Cocoa selector acceptance. The separate endurance app
runs concurrently; elapsed times do not qualify latency. Physical pointer/key
routing, human VoiceOver, broader display configurations, native/GPU memory,
performance and full Mac/P6/P11 remain separate. Internal review and one bounded
Anthropic Claude Sonnet 5.5 Prism checkpoint review precede commit.
