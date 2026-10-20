# Pre-Show quit checkpoint

The Driver correction passed the controlled pre-Show CPU-deadline regression and its real old-Driver counterfactual. **Native acceptance remains open:** ordinary close and the after-Show case failed; no new presentation or performance qualification was established.

The actual SDK104 comparison at source `344447c639d3b92fbd8100c391aa76606a0d4ce4` failed its first serial-grid case at the **330.044 s watchdog**. Stdout contained setup only; the presentation journal was empty. The remaining three cases were not run. Empty output does not prove no first frame occurred. A separate diagnostic deliberately captured SIGQUIT stacks after **8.147 s**, with expected exit **2**. Its idle main/Core loops and absent monitor/perf-worker/render stacks are one interrupted snapshot, not normal shutdown or proof of the original trigger. Both failed/diagnostic scopes remain unchanged.

Source inspection confirmed that Driver.Run discarded quit requests while `shown` was false. A terminal pre-Show failure can prevent Show and permanently lose that request. The correction retains bounded quit intent independently of Show, completes the whole topology-construction batch and unlaunched-tree cleanup, then dispatches existing asynchronous service shutdown→Core QuitReq exactly once outside both owners. Supported synchronous Mac WindowStage.Run completion establishes safe-to-quit construction without mutable Scene/Stage traversal. First-error and no-window returns remain unchanged; Core104/prepared bytes are unchanged.

Actual headless checks passed **782 top-level test results across packages: 649 Driver-root plus 133 command/internal**, with zero failures/skips. The ten new quit-protocol methods are a subset. Vet, both module graphs, Darwin executable builds and optimized metadata checks passed. The original controller remains FAILED because its population validator used the wrong count; its successful race run was retained, followed by seven previously unrun phases. The [combined audit](evidence-index.json) binds all eight actual phases.

Fresh-process native results were:

- Corrected pre-Show: **PASS**, genuine **250,042,834 ns** Prepare work with a 100 ms gate, typed presentation-deadline refusal, Run return, worker join and normal cleanup, without Show.
- Old Driver pre-Show: expected hard-watchdog failure after **10.834 s**. Its eligible entry record does not establish completed work, refusal type or normal cleanup.
- Corrected ordinary close and after-Show: **FAILED**, both returning “drawable was not presented” before slow work. Run/worker/close and terminal cleanup were observed; the intended after-Show CPU deadline was not exercised. Ordinary close was not rerun.

Both metrics APIs were nil. These are CPU-deadline/lifecycle checks, not lost-presented-callback proof. All owned groups retired; full source/SDK boundaries and geometry stayed stable. Hidden Cocoa-window records remain retained without deallocation claims. The [native continuation audit](evidence-index.json) preserves the overall FAILED outcome.

**PENDING:** Prism/disposition, integration/commit, fresh SDK and actual packaged `cmd/display-backpressure-smoke` verification. Builds used that source over the existing helper command path. Renewed presentation acceptance remains unverified. Selected receipts, raw logs, controllers and audits are preserved in [evidence.tar.gz](evidence.tar.gz) and mapped to their original paths and hashes by [evidence-index.json](evidence-index.json). The original report and its `inputs.json` are archived unchanged; full source/SDK/prepared vectors and executables are retained externally rather than duplicated. No policy, gate or classifier relaxation is proposed.
