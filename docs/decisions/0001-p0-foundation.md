# ADR 0001: prioritize a bounded Cogent extension trial

Date: 9 October 2026
Status: Provisional; P0 local comparison complete, production foundation undecided

Follow-up: [ADR 0002](0002-foundation-gaps.md) records the local P1–P3 results and holds foundation acceptance pending remediation.

## Context

Vellum seeks ordinary Go application code, professional engineering widgets, custom shared-device GPU work, explicit commands/tasks, multilingual editing, native accessibility and maintainable desktop distribution. Research proposed an independent core (B), with full Gio (A) and Cogent Core extension (D) as mandatory reuse controls.

P0 implemented separate Gio/Cogent workspaces over the same domain model and deterministic 100,000-row documents. Both ran on macOS without toolkit patches or private imports. Cogent's public xyzcore GPU scene also rendered; pinned source shows reuse of the window surface's GPU/device. The five shared model tests passed with the race detector; both native builds and all module vet checks passed. See [P0 results](../reports/P0_RESULTS.md) for exact scope, evidence and automation limitations.

Neither workspace exposed its controls in the queried native AX tree. Real IME, Windows/Linux behavior, packaging, demanding custom GPU passes and performance budgets remain untested. Cogent's built-in reflected table also does not establish Vellum's desired typed consumer API. Application line counts were close and do not establish a general effort advantage.

## Decision

Prioritize **Architecture D for a bounded P1–P3 foundation trial**, with the P3 accessibility feasibility spike started early. Retain Gio as a control and B as the fallback. Defer implementation of an independent UI core until a concrete integration gap justifies its ownership cost.

This is a choice of the next experiment, not production acceptance and not a commitment to expose Cogent's API as Vellum's public API. The prototypes remain disposable.

## Reasons

D already supplies retained workbench widgets and a demonstrated public shared-device scene path. A also supports the common workspace, but P0 exercised only ordinary Gio GPU operations and did not demonstrate the needed arbitrary GPU boundary. Both have unresolved native accessibility. There is insufficient evidence that a new core is necessary, and insufficient evidence that either reuse candidate meets every hard requirement.

## Conditions for continuing D

1. P1 native surface/lifecycle and distribution checks succeed across the declared desktop matrix without unsupported backend forks.
2. P2 demonstrates real composition, multilingual editing, caret/selection and fallback behavior.
3. P3 demonstrates real native assistive-technology interaction, editing actions and stable virtual identities through public hooks or small maintainable changes.
4. A typed workbench API and custom GPU contract can be built with modest adapter code, retaining most toolkit functionality.

## Reversal criteria

Move to a minimal B shell/GPU/text/accessibility spike if a required capability needs a long-lived private fork, substantially replacing D internals, or unacceptable application binding/code-generation requirements. Record the specific unavailable boundary and estimated adapter/core ownership before reversing. B must satisfy the same gates; it receives no automatic pass because it offers API freedom.

If neither route can meet the gates within a bounded trial, narrow support or defer the framework rather than claim the requirements are implemented.

## Consequences

The immediate roadmap is risk retirement, not broad widget development. Cogent's larger dependency closure and reflected data controls need active scrutiny. Existing command/model/task code is reusable experimental evidence, but does not become the production API by default. The original research recommendation remains historical context; this ADR records the more specific post-P0 priority.

## Reproducibility

Gio: `gioui.org v0.10.4-0.20261006204131-9393ca1c7322`.
Cogent: `cogentcore.org/core v0.3.46-0.20261009095036-5cc519b24044`.
WebGPU: `github.com/oliverbestmann/webgpu v1.36.1-0.20261006093257-de8bdd5d5976`.
Recorded environment: macOS 26.6.2, Go 1.27.1, darwin/arm64 with Xcode. Run [the prototype instructions](../../prototypes/p0/README.md); inspect its evidence manifest. No cross-platform or clean deployment result is asserted.
