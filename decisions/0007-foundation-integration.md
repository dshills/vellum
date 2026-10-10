# ADR 0007: Integrate bounded native delivery and resource ownership; retain G1 hold

Date: 9 October 2026. Status: accepted for experimental evaluation; production foundation unaccepted.

The earlier mailbox prototype bounded delivery but did not reserve capacity before Cocoa mutation. Native adapters attached after Show, and repeated actual GPU/window lifetimes retained owned queue/device references. Eager system-font bitmap/static-outline resources also exceeded the isolated shaping target.

Integrate a separate owned follow-up using the same pinned Cogent revision, a proposed main-thread NativeAttach lifecycle hook, prospective native snapshot admission, a bounded pump and explicit GPU queue/device release. Add opt-in stable-resource go-text patches that defer bitmap/static-outline payloads and account retained resources. Package the fixture with a strict, relocatable macOS bundle workflow. Preserve original prototypes and failure evidence as controls.

The measured isolated seven-script heap is about 27 MiB, with matching geometry and decoded emoji pixels. A corrected 5,000 hidden-window/GPU lifetime run drains 55,000 accepted updates with substantially flatter RSS than the failing baseline. Full fixture heap remains about 56 MiB. Snapshot disk/descriptor costs, eager compatibility fallbacks and the supported-outline policy are explicit tradeoffs, not global memory guarantees.

Keep Cogent as the experimental foundation. Do not expand the broad workbench until G1 passes. Live candidate panels/full screen-reader workflows, visual bidi/arbitrary text ranges, legacy key ordering/fairness, global memory/render budgets, native platform runtimes, trusted clean-machine distribution and upstream acceptance remain open. See the [results and evidence](../FOUNDATION_INTEGRATION_RESULTS.md).
