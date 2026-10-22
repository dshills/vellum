# Corrected framework timed cohort — rejected during warmup

The first corrected framework timed cohort stopped at request 33 after 32
completed warmup inputs. No measured input was reached. The process reported
`foreground/configuration changed during cohort` and `display clock invalid or overflow`.
The existing strict validator independently rejects the journal for foreground loss.
This closes no responsiveness or release requirement.

Preparation was reviewed and committed as 3d017bd; binary 6825d90e and the
locked match were unchanged. The run used the owner's current authorization
for remaining release screen tests. Total duration was 12.111 seconds, with
5.218 seconds in the native command, within the 60/40-second bounds. All nine
owned native/configuration groups retired and were observed absent. All 18 logs,
source/binary pins, 8321 SDK files, geometry and configuration checks pass.

All 42 acquisition attempts, 158 events, 259 display ticks, 33 scheduled
requests and 32 completed input measurements remain in the archive. Independent
replay verifies terminal pending/acquiring zero and the closed recorder scope.
All 36 scoped samples join uniquely by acquisition timestamp, drawable ID and
shown callback. Of all 42 callbacks, 40 have positive shown times and two
pre-observer callbacks have zero shown time; neither is discarded.

The app began active, key and visible, with itself foreground and zero losses.
At the end another PID was foreground and three loss notifications were counted.
The observer queries NSApp.keyWindow: a nil layer and zero dimensions after key
loss do not prove that the app window was destroyed. The other process's identity
and cause of focus change are unknown. No user interference or renderer fault
is inferred. The diagnostic currently omits its first invalid per-request state;
record that failure context before another live attempt rather than repeat this
unchanged preparation. The existing scope is sealed and will not be rerun.

See the [summary](result-summary.json), [verifier](execution-sources/verify_result.py.txt)
and [indexed raw evidence](evidence-index.json). The archive contains complete
logs, exact command and human authority, metadata, match, stop marker and bound
inputs. Explicit mappings verify the committed copies. Executables, full SDK,
Core, caches and previous sealed captures are excluded. Paths are local provenance.
Internal verification and Prism with Anthropic Claude Sonnet 5.5 cover this
result before commit. No native pair, latency percentile, cadence qualification,
physical-input verification or full macOS release readiness is claimed. Features,
rendering/pacing defaults and the 60 Hz/33.3 ms release gate remain unchanged.

Archive `capture/REL` maps to `prepared/REL` in the original local artifacts
directory named in the index. The verifier is an execution-source snapshot, not
a standalone portable command. Member hashes and committed-copy equality can be
checked from the archive alone; full input/SDK revalidation also requires the
external pinned workspace.
