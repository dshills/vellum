# Foreground failure context for the timed diagnostic

The first corrected timed cohort was rejected during warmup in a675ccf. It did
not retain its first invalid per-request observation. This diagnostic-only change
records that state and sequence, plus a separate follow-up native observation of
the foreground application bundle/name and the diagnostic's own window count.
The follow-up observation can differ from the first state and does not establish
why focus changed. Original foreground rejection and strict timing validators
remain in force. No extra native observation runs for admitted requests.

The fresh private consumer reuses the same physical Core, SDK, native archive,
match and 128-warmup/120-measured four-phase schedule. Only the Go failure branch,
its helper and failure-only Objective-C context differ. All other copied consumer
bytes verify against the previous preparation. The new binary is cf8fd277…;
rendering, pacing and the 60 Hz/33.3 ms release gate remain unchanged. This
instrumented fixture cannot by itself qualify the representative release workload.

Two focused Go tests pass with the race detector: all 248 valid inputs collect no
extra context, and later teardown cannot replace the first failed state. Three
Python routing/patch methods pass, covering thirteen refusal subcases. A focused
check initially found that the Objective-C patch could be reapplied; the recipe
now rejects missing, repeated and already-patched anchors. Both initial and final
build/check receipts remain in the archive. The final build passes; every owned
command retires with complete logs and stable input hashes. Source snapshots,
plan and indexed checks retain exact local provenance; archived helpers are not
advertised as portable standalone commands.

Internal review and Prism with Anthropic Claude Sonnet 5.5 precede commit. The
fresh runtime scope uses the existing unchanged bounded runner: 40 seconds per
native process, 60 seconds overall, fail first, retain every partial journal and
seal on completion. The previous rejected scope remains sealed. Current user
authority permits screen use for remaining release tests. No UI has run in this
preparation. Review and commit its actual result separately. No feature, public
API, default policy, SDK, release criterion or public-distribution change follows.
