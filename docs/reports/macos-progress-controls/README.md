# Circular progress controls

The portable vocabulary adds ProgressRing and ProgressGauge, both controlled
passive leaves with finite fractions in 0..1 and literal optional center text.
Native circular/semicircular meter types are set before initial styling. Keyed
updates retain identity; kind changes replace the widget. Explicit size/growth
styles override and restore defaults. Both publish a read-only Cocoa progress
indicator with named semantics, accepted value and zero/one range.

Portable validation/focus, retained styling/semantics, completed software frames
at DPI96/120/144/192 and managed-to-Cocoa inert role checks pass. Fractions
0/0.25/0.75/1 change actual pixels with no animation/timer work. Gallery and the
full acceptance catalog include both controls. The relocated SDK fixture now
uses the exported kinds. Ten code/build gates pass, including the full
portable and adapter race suites, vet, relocated SDK fixture, affected Gallery
V2 race tests and final native build. The initial portable race run could not
bind its test Unix socket under the sandbox; the permitted rerun passes and both
receipts are retained. Internal review repaired native readiness and preservation
of driver errors joined with cancellation, with a regression test.

The dedicated progress-smoke mode completed four ordered values in a real Cocoa
window. Both named progress indicators retain native identity and positive
geometry, publish accepted values, refuse read-only setters, and report actual
main-thread cleanup with stopped callbacks and zero pending/native/AX/window
resources. It runs only after the rendering benchmark process group is reaped.
Native geometry/value observations are separate from software pixel tests;
physical interaction and human VoiceOver were not exercised. Internal final
review and one Anthropic Claude Sonnet 5.5 Prism review precede commit.

These checks do not qualify physical input, human VoiceOver, complete current
P6/P11 or Mac distribution. No OS preference/input-source/clipboard change is
needed. The existing native latency targets remain unchanged.
