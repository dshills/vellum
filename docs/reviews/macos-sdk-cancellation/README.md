# M1: SDK cancellation ownership

SDK commands now retire their owned POSIX child group when default main-thread
SIGTERM or Ctrl+C initiates cancellation, including a signal delivered before
`Popen` returns its child handle. Additional cancellation cannot interrupt failure
cleanup. Original handlers are restored; SIGINT is restored last so a new Ctrl+C
cannot strand SIGTERM on an inactive handler.

This closes M1 of the [Mac delivery plan](../../../specs/macos-delivery.md) in the
source checkout. The existing private SDK archive remains unchanged; the final
SDK packaging milestone must include this fix. All other Mac gates remain open.

## Checks

The patch is based on clean `d9d92a0`. [The result record](result.json) binds the
tested Python bytes, contract, raw results and review outcome. No Go/native runtime
code changed and no native window or OS preference was used by these tests.

| Suite | Passing test methods | Raw result |
| --- | --- | --- |
| Child process | 16 | [output](final-child-tests.stderr) |
| Mac build | 11 | [output](final-build-macos.stderr) |
| SDK release | 29 | [output](final-release-sdk.stderr) |
| SDK checker | 24 | [output](final-check-sdk.stderr) |
| App packaging | 16 | [output](final-package-app.stderr) |
| Distribution inventory | 11 | [output](final-distribution-inventory.stderr) |
| Total | 107 | Six headless suites |

The child suite includes twelve actual signal/stage/output-mode combinations:
SIGTERM/SIGINT during child creation, active communication and held failure
cleanup, with captured output and bounded binary logs. A real direct child and
grandchild ignore both signals; the assertions require them to be absent or
terminated zombies after group cleanup. Zombie reaping by the OS is separate.
The suite also injects an actual Ctrl+C immediately after SIGINT restoration,
checks normal/spawn-failure handler restoration, caller-owned signal policy and
worker-thread use.

The [original-source creation control](counterfactual.stderr.gz) fails all four
creation subcases with a live `Ss` child before test-fixture cleanup. The
[pre-review handler-order control](restoration-control.stderr) fails because
SIGTERM remains attached to the inactive guard. Both are expected negative
controls; their failures are preserved (the creation log is gzip-compressed
without changing its bytes). Fixture finally blocks terminate the
owned test groups even when the original implementation fails.

## Reviews

Internal code review identified the restoration-order race. The correction and
its new boundary regression were independently accepted before Prism.

One [Prism review](prism.json) used Anthropic Claude Sonnet 5.5: three files,
22,064 diff bytes, one fresh provider call/chunk and complete coverage. Its raw
CLI exit is **1**, with one medium finding (`2b57f063be7ba84e`) retained. The
finding proposes that pending cancellation replace an already propagating
failure. Independent assessment accepts the intentional existing precedence:
failure cleanup preserves its initiating error and still retires descendants.
The repeated-signal tests explicitly verify this. The release contract was
clarified; no runtime change or second provider call followed that review.

When cancellation initiates failure, SIGTERM raises `ChildCancelled` (a
`SystemExit` with status 143), and SIGINT raises `KeyboardInterrupt`. When
cancellation arrives during an already-failing cleanup, the initiating failure
remains primary. Cancellation after successful child work can turn the call into
a failure after retirement; the runner does not report success over cancellation.

## Scope

This handling applies to synchronous POSIX main-thread calls using default
SIGTERM/SIGINT policy. Custom/ignored dispositions and worker-thread process-signal
policy remain caller-owned. It cannot intercept SIGKILL, bound an OS creation call
before `Popen` returns, or catch descendants deliberately leaving the session.
Ordinary completed commands retain ordinary subprocess/background-work semantics.
Existing Windows transport fixtures pass; no Windows runtime qualification follows.

This checkpoint does not qualify GUI interaction, frame/memory/endurance budgets,
native input/accessibility, or public distribution. Live tests remain stopped.
