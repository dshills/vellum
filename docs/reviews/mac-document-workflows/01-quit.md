# Coordinated Quit — internal review

Added opt-in owner-delivered native Quit and a bounded app-owned Notes QuitGroup,
wired into Notes, Document Starter and A/B Documents. No root API, dependency
patch or backend policy change. Existing single-window Forge/Measure Pad behavior
is retained. Public native facade compile coverage includes the new callback.

Reviewed owner/foreign-lock separation, bounded repeated requests, handler setup
before native construction, input settlement, all-window decision holding,
journal revision checks, save/recovery failures and callback retirement.
Review corrected asynchronous settlement-error notification, immediate recovery
error notification, held-state capture access, and a retired-reference retry
that could strand survivors. Holds are bypassed only for synchronous owner-side
preparation/capture; normal commands stay gated.

Focused race regressions passed for cancel preserving clean+dirty windows,
all approvals, saved-first/Cancel-last, stale revision, recovery failure, native
settlement refusal, closed refs, native callback coalescing and existing native
Quit/close policies. Later final-package native checks exercise Command-Q on the
actual integrated build; unit routing is not physical pointer qualification.
No external code-review service was used.
