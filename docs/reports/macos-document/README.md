# macOS document persistence acceptance

This phase supplies a reusable public document session, atomic file store, history
replacement transaction and owner close-request policy. The workbench consumes
those APIs for Save/Save As/Open and an asynchronous Save/Discard/Cancel dialog.
Component properties, exact canvas geometry and ordered file references persist;
UI viewport/selection/layout remain independent. Failures and edits during IO
preserve dirty state and refuse automatic close.

The source/check proof records exact final source identities, commands, exits and
log hashes. Full fresh repository checks include portable/native race tests, vet,
owned backend regressions and independently consuming modules. Focused race tests
repeat root concurrency/IO cases twenty times and native/model/workflow cases five
times. The relocated archive regression exports document/app/ui/history and runs
file persistence after deleting its original fixture source. Public SDK consumers
also exercise the generic session and confirmed close policy.

The native fixture runs in an actual Metal/GLFW/AppKit workbench. It requests
native close, activates dialog controls through Cocoa accessibility, verifies
Cancel retains the window, Discard/Open restores full saved domain data, an
injected Save failure leaves dirty changes open, and successful Save closes only
after acknowledgment. The final race-enabled fixture passes with 1,000,000
component records, 96 rectangles and a 1,000-component outline. Existing 100,000-row
GPU readback/history smoke also passes with the added document toolbar/dialog.
It uses a temporary document and deterministic chooser.
Actual file-panel selection, external AX/human screen-reader workflows, external
file conflict detection and power-loss durability remain separate gates. System
input sources/VoiceOver were unchanged, and the fixture windows closed.

Internal implementation reviews precede one significant Prism batch using the
user-selected Anthropic Claude Sonnet 5.5. See [review records](../../reviews/macos-document/README.md).
The experimental Mac framework remains in progress; this phase does not certify
production readiness, public distribution or Windows/Linux runtime behavior.
