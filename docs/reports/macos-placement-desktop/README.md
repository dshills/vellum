# Native placement desktop probe

The reusable placement feature and independent demo retain their existing
software-routing evidence. This checkpoint adds an opt-in native observer, tests
its failure/cleanup controls and preserves one **failed desktop attempt**. It does
not establish native drag/drop, keyboard placement or full Mac acceptance.

Source is clean commit `00f060bb6c1a504fbc74c39c2a406db54f8d611c`: 4612 tracked
files and 6243 prepared files. The default example still calls
its original implementation. Only the Mac `vellum_placement_probe` build tag
selects a separate main entry point, with consistent native system/core identities.
A Go test executable would select the upstream offscreen driver; it is explicitly
refused as native evidence. See the [example](../../../examples/placement/README.md)
and [contract](../../../specs/placement.md).

| Check | Actual result | Scope |
|---|---|---|
| Private writer, observer retirement and panic/refusal controls, race | PASS, 3.253 s | No native input/window |
| Independent placement routing suite, offscreen race | PASS, 5.551 s | Software events/model/history |
| Native probe vet and executable compile | PASS, 0.470 / 3.170 s | Compile/static only |
| Fresh clean committed native bundle | PASS, 3.896 s | Build only |
| Native desktop session | FAIL, exit 1, 601.477 s | Ten-minute context deadline |
| External desktop app binding | FAIL, 924.3729 s | Server `-10005: timeoutReached`; no inputs requested |
| Terminal cleanup and fresh process observation | PASS | Owned group/exact executable absent |

The session began at `2026-10-17T00:03:09.373228+00:00` and ended at
`2026-10-17T00:13:10.858219+00:00` UTC. Twelve startup samples contain eleven unavailable
accessibility snapshots, followed by **one** reachable native AXWindow with 24
elements. The observer skips unchanged model state; these are not ten minutes of
continuous accessibility or presentation sampling. A thirteenth, closed record
contains `context deadline exceeded`, with nil native/observer errors. The saved
JSONL is 26,559 bytes. Objects, Undo/Redo counts, activation/change/removal counts,
gesture state and reserved completions remain empty/zero throughout.

The desktop inventory found the exact running bundle before `cua.getApp` timed
out. No source click, drag, key action, placement, selection, transform, snapping,
rejection or removal case was attempted. The cause of the external binding
timeout remains unassigned; this evidence does not distinguish application,
external-accessibility and automation-service causes. The probe closed at its own
deadline while the binding request was still pending; the outer 710-second
watchdog did not have to force termination. Its controller reaped the owned group
and held the shared lease through cleanup. A later independent process observation
again found no group members or exact probe executable. No further CUA attempt
was made. Source/build/review pins and log hashes remained stable.

Internal source review fixed the original offscreen test-driver selection,
observer teardown without proven join, private test-directory setup and secondary
panic-path error retention. Failed source/check endpoints are preserved. One
bounded Prism review used Anthropic `claude-sonnet-5-5`, six files / 21,492 bytes,
one chunk/call, complete coverage and one raw medium finding. Independent review
refuted that finding against the actual owner/mutex, completion and nonblocking
wake paths; the raw provider exit 1 and finding remain in the archive. There was
no second provider call for this checkpoint.

The probe never synthesizes inputs or calls placement/history commands. Its
diagnostic controls can remove a binding/window after observing a threshold drag.
Target removal is armed at the first sampled drag; a future run must inspect the
pre-removal Accepted/TargetID state before claiming accepted-candidate cancellation.
Successful process closure alone is never workflow acceptance. Physical hardware,
continuous visible previews, held-Escape/capture loss, ancestor scrolling,
cross-display/scale transitions, cross-window capture, GPU presentation/performance,
native heap reclamation and human assistive technology remain unqualified.
Cross-window payload/capture transfer remains unsupported by this initial API.
P4/P6/P9/P11, ordinary latency, overnight, interoperability and distribution
acceptance retain their existing open gates and failures.

[Result summary](result.json), [file manifest](evidence.manifest.json) and
[preserved evidence](evidence.tar.gz) contain the original source, failures,
reviews, provider outcome, producers, full source vectors, actual receipts and
native records. Executable bytes are hash-only exclusions.

Archive SHA-256: `5f3ecdfaf80cdacd5ee746cf8c17b3da984a01141dbb75eeaa502e8fb69a434b`.
Manifest SHA-256: `3a5c04d0ea6b764cc0ef00803a0f824886f4a10b70fd1a7f713fac4d88d793cb`.
The archive contains 93 entries,
1,668,609 expanded bytes and
585,329 compressed bytes.
