# Window construction after manual retirement

Closing the last manual window left it as the global construction context.
A subsequent ordinary window inherited that retired manual mode and silently
refused to open. The committed baseline reproduces this as a 15-second Show wait
timeout when the manual-construction test precedes rendered geometry.

Owned patch 0070 removes all three registry memberships and withdraws or replaces
the current context in one mutex transaction. Native focus checks membership,
orders focus and publishes the current context under the same mutex. Raise
refuses unregistered windows and publishes before the native call, so reentrant
close cannot be overwritten afterward. Native calls run outside the mutex.

The exact failing two-test sequence now passes without resetting global state.
Five new regressions fail against unchanged production and pass with the fix:
real manual close followed by ordinary Show; survivor/last/repeated retirement;
reentrant native Raise that closes itself; admitted Raise concurrent with close;
and concurrent focus/close events.
The broader selected `^TestVellum` core race run passes all 115 top-level tests
with no skips, including live mixed-mode refusal checks. Full adapter-module
race and vet checks and all 11 preparation Python tests pass. Fresh preparation
reproduces all 6,230 dependency files, with exactly the declared three changed
or added leaves relative to the 6,229-file baseline.

`records/evidence.tar.xz` and `records/archive-index.json` preserve the producers,
exact baseline/candidate maps, terminal logs, the initial four intended negative
assertions,
draft source and internal reviews. An earlier regression fixture incorrectly
expected focus restoration to preserve the previous current window; its original
failed run is preserved separately and supplies no acceptance evidence.
`records/qualification.json` binds the final source and review associations.
The archives exclude private executables; scripts retain original absolute paths
and are evidence records rather than a portable replay package.

The single Anthropic Claude Sonnet 5.5 Prism review returned one low concern
about native Raise after a concurrent close. Existing Mac code checks closed
state and rechecks the native handle inside its main-queue callback, while
destruction runs on that same owner. Source assessment is retained in
`records/provider-source-assessment.json`; no new production change was needed.
A deterministic concurrent-close regression was added afterward. It verifies
core context withdrawal, not Cocoa handle lifetime. The final 115-test core
run and five intended negative failures are in `records/followup-evidence.tar.xz`
and `records/followup-index.json`. The two production files are byte-identical
to the reviewed and fully adapter-tested version; adapter race/vet was not
repeated after this test-only prepared-source delta. The final test, pin and
documentation delta receives internal review separately from the provider
reviewed vector.

These are offscreen construction/routing and race observations while the user
is online. No live UI or native GPU test was admitted. Native handle lifetimes,
physical desktop focus, native object deallocation, ordinary frame budgets,
overnight acceptance and the remaining Mac gates need separate evidence.
The prior GPU-host diagnostics draft is outside this patch and must be rebased
onto the new lifecycle source before its remaining native checks.
