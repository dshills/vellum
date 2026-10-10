# ADR 0002: hold foundation acceptance for text, scheduling and memory

Date: 9 October 2026
Status: Provisional; local P1–P3 feasibility investigation complete, G1 not passed

## Context

[ADR 0001](0001-p0-foundation.md) selected a bounded Cogent extension trial. The new macOS fixture attaches AccessKit C to the public GLFW Cocoa view, exposes actual native controls, and routes native actions through the UI's validation. The final native run handles all 18 callback actions, including editing, checkbox state, selection, paging and shutdown. This establishes a useful extension seam without a toolkit fork.

It also reproduces grapheme deletion failures in the ordinary property field. The GLFW input path inspected at the pinned revision lacks delivery of marked text and a caret rectangle to the Go widget. Real IME completion and full screen-reader editing remain untested. The small fixture retains roughly 404 MiB of Go heap, largely font parsing and spell suggestions in the sampled profile. Idle scheduling uses an unconditional refresh-rate ticker even when no repaint occurs. See [the results](../reports/FOUNDATION_RESULTS.md) for diagnostic measurements and precise limits.

## Decision

Do not accept Architecture D for production or advance to broad P4–P8 work yet. Keep D as the current experimental candidate, A as control and B as fallback. Run a bounded remediation feasibility check before spending on a workbench API:

1. Identify supported or small upstream hooks for grapheme navigation/deletion, marked-text events, UTF-16 selection/caret bounds and AccessKit text ranges. Implement one representative field, then exercise a real IME and VoiceOver workflow.
2. Identify a maintainable demand-driven paint scheduler with explicit animation/caret deadlines. Rerun visible, occluded and minimized 60-second idle samples; instrument wakeups separately from repaint counts.
3. Test lazy spell initialization and bounded font loading/cache behavior with unchanged multilingual coverage. Report memory with and without optional spelling; do not gain a pass by dropping required text support.
4. Secure the adapter's native creation/focus/destruction hooks, bounded lossless queue and coherent field/window focus. Complete target-platform and clean-machine packaging checks before G1 acceptance.

Each change must be scoped and measured against the retained baseline. If it requires replacing most text/platform internals or maintaining an unsupported backend fork, trigger ADR 0001's B fallback and test a minimal independent shell/text/accessibility integration against the same requirements. P2 failure alone does not prove that B has lower ownership cost.

## Consequences

Native accessibility is no longer only a speculative integration route, but full assistive-technology usability is still an open gate. The next priority is text correctness and native input/accessibility integration, followed by scheduler and memory fixes. Source and evidence remain disposable. No framework API, cross-platform support promise, benchmark pass or shipping package is committed by this ADR.

## Follow-up

[ADR 0003](0003-remediation-feasibility.md) records the completed local field, scheduler and spelling remediation trial. The foundation remains on hold for the unresolved native-input, memory and platform gates.
