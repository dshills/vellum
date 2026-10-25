# Personal-use macOS v0.1 finish contract

Owner approved all five finish areas on 25 October 2026. This is the active,
feature-frozen personal delivery contract; historical production qualification
and published alpha.3 records retain their scope. Baseline: `929541a`.

## Fixed scope

The exact existing control catalogue is frozen in
[personal-macos-v01.json](../release/personal-macos-v01.json). No new widgets,
rendering/pacing experiments, new performance harnesses, or backend rewrites.
Allowed changes: system appearance/accessibility integration for existing controls,
input/focus/lifetime defect fixes, required tests, documentation and private SDK
packaging. Original-color icons remain untinted. Preserve existing public APIs;
provide a named supported Mac titlebar entry point without removing its old alias.

Target: this Apple Silicon Mac and its recorded macOS version. Other OS versions,
Intel, additional display hardware and public distribution are unqualified. This
is `v0.1.0-personal.1`, not a stable public v0.1.0 compatibility promise. No Git tag
or public release is created. The existing explicit light/dark overrides remain
available alongside opt-in following of system appearance and accessibility
preferences. Preferences are observed, never modified by the framework.

## Checkpoints

| Phase | Required outcome | Completion evidence |
| --- | --- | --- |
| 1 Scope | Frozen catalogue, supported configuration, limits and finite checklist | Reviewed contract and manifest; separate commit |
| 2 Mac behavior | System appearance/accent and accessibility preferences; preserved explicit overrides/color artwork; supported titlebar API | Focused owner/lifetime/theme tests, native compilation, Prism and commit |
| 3 Input/accessibility | Existing keyboard, Unicode/composition, focus and AX contracts; native keyboard/physical checks; authored human IME/VoiceOver checklist | Focused regressions and app checks; human-only cases explicitly pending until observed; Prism and commit |
| 4 Integration | Existing widgets with tasks/documents/multiple windows; teardown and bounded resources; lifecycle/display transitions within available hardware | One stable-candidate suite and bounded native session, repair only failures; Prism and commit |
| 5 Delivery | Matching SDK and app; fresh independent consumer from extracted SDK, notices, concise guide and actual final ZIP verification | Clean source/version/hash association, independent build, packaged launch, Prism and commit |

Each phase updates the delivery ledger. A phase's implementation and automated
acceptance may finish with named human/hardware qualification pending; that is
not reported as completed human acceptance or production readiness. Do not
silently drop or relabel a failed advertised workflow. Actual crashes, hangs,
data loss, wrong history and unbounded ownership block the personal candidate.

The old 4 ms target remains optimization work. Disclose actual responsiveness;
do not invent a new passing number or restart an open-ended profiling campaign.
Reuse verification for unchanged bytes and run focused tests after repairs.
A final app-local desktop session of at most 30 minutes is owner-approved;
no global input source or accessibility setting changes are authorized.

## Delivery

Provide an ad-hoc signed, unnotarized private Mac Widgets app and matching Go SDK,
with source identity, checksums, known limits and independent build instructions.
Use existing SDK/package tooling and guards. An old exact-payload distribution
approval never approves a changed payload. No public publication is requested.
Human VoiceOver/real IME and unsupported display/hardware checks remain explicitly
separate from native event helpers, AX automation and offscreen composition.
