# ADR 0012: own and maintain the pinned native integration

Date: 28 October 2026
Status: Adopted maintenance policy for the current backend

## Context

ADR 0001 selected a bounded Cogent extension trial and identified a long-lived
private fork as a reason to reassess ownership. The delivered integration now
contains 108 Cogent patches across 98 stages, touching 376 distinct paths, plus
six typesetting patches touching 23 paths. Patch text includes context and tests;
these counts are not net production line counts. This is substantial maintained
integration code. Source hashes and current counts are recorded in
release/native-patch-inventory.json and checked by scripts/check.sh.

## Decision and accountability

Retain the working Cogent/WebGPU backend and the portable Go/native-driver module
boundary. Treat the prepared sources as Vellum's maintained integration fork,
reproduced from pinned upstream versions and checked patches. The Vellum project
maintainer owns patch admission, pin updates, notices and qualification decisions.
No responsibility is assigned to upstream maintainers or unconsenting third parties.

Patch changes must identify their subsystem: native window/input, text/shaping,
rendering/GPU, managed widgets, accessibility, or preparation/distribution. Record
the invariant repaired, focused regression, upstream base and impacted services
in the checkpoint review. Update the descriptive patch inventory after review.
The CI guard checks manifest identity, patch hashes and the resulting counts; it
does not certify runtime behavior, complete diff semantics or licensing.

## Upgrade and upstreaming strategy

Prepare an upstream update in isolation. Reapply patches in dependency order;
classify each as upstream-equivalent, still required, conflicted or obsolete.
Retire a patch only with equivalent focused regression coverage. Run the portable
API guard, affected native adapter checks, independent SDK consumer and the
specific native checks warranted by changed input/window/GPU paths. Reuse unchanged
qualification. Preserve the last working pin until the candidate passes.

Keep independently useful fixes separable for possible upstream submission.
Track their disposition in their checkpoint; sending contributions requires its
own explicit task. Avoid new dependence on private upstream internals when an
existing supported boundary suffices. A managed fork repository or rebasing system
can replace staged patches later if measured upgrade work warrants that change.

## Architecture boundaries and reconsideration

Keep the serialized application owner, exact window references, bounded task
completion, domain-owned history and asynchronous document revision checks.
Applications consume public Vellum APIs; low-level extensions remain explicitly
experimental. Keep minimal starter templates separate from diagnostic consumers.
The broad Node/driver surface is a maintenance cost; prefer private internal
helpers and focused adapters while preserving stable public import paths.

Reconsider the integration strategy when an upstream update cannot preserve
these invariants within bounded focused work, when patches require repeated
replacement of major upstream subsystems, or when reproducible developer builds
cannot be maintained. Document the specific conflict and costs before selecting
a different backend. This decision introduces no rendering/pacing experiment,
release gate, platform claim or new distribution approval.
