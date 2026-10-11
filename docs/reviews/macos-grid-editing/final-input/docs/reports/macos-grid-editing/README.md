# Editable grid cell checkpoint

This checkpoint covers bounded portable drafts, per-cell conflict detection,
application validation and atomic immutable-source completion. The managed Mac
adapter reuses one native text field clipped inside the grid cell and settles
originating text input before commit. The independent consumer exercises sparse
versioned updates, conflicts, correction/cancellation and separate window state.

Qualification records bind commands, exit status, exact raw logs and selected
source hashes before/after. Core race/vet evidence and native/consumer checks are
separate. Actual Cocoa smoke measures completed geometry and settlement; routed
offscreen tests exercise retained pointer/key controls. Neither is human IME or
external accessibility acceptance, dependency build-source closure, sustained
latency/memory measurement or completion of P6/the macOS framework milestone.

Internal review and Prism records are in
[the review directory](../../reviews/macos-grid-editing/README.md). Historical
failed exploratory runs retain captured source/log bindings and remain
distinct from final qualification. Development logs without before/after input
maps are explicitly unqualified. No OS input, VoiceOver or privacy settings
are changed. Windows/Linux acceptance remains deferred.

The final portable record is [core-v4](core-v4/qualification.json): repeated
race checks for `virtual`/`ui`, all core-module race tests and core vet. The
[native-v1 record](native-v1/qualification.json) covers repeated inline-field and
existing field/editor compatibility tests, both offscreen/default vet and native
compilation. Its [handoff](native-v1/handoff.json) preserves selected source maps,
inert snapshots, prepared-tree fingerprints and raw logs. These records completed
against unchanged selected inputs. Consumer and Cocoa records are qualified
separately; the aggregate ledger identifies the final accepted records.

The [qualifier deletion check](qualifier-deletion-check/regression.json) uses an
isolated real tracked-file deletion. It demonstrates the original evidence
capture crash and the corrected wrapper's explicit Go failure record. The wrapper
now retains absent/nonregular source entries; the complete core checks were
rerun after this correction.

The original portable review tests demonstrate clipped-cell admission and cell
version-history failures before correction and pass unchanged afterward. Native
development failures are retained as unqualified history. Consumer development
records preserve the Cancel retirement failure, later corrected by the adapter,
and observed source drift during qualification. A withdrawn construction concern
is recorded separately; no runtime fix was needed after verifying the window
lifetime invariant. Historical pass logs do not override known source drift.
