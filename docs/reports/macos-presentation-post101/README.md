# Source101: 60 Hz cadence restored, latency gate still open

Committed source `8ba9ca617d5c01d40fd9785a2896b8153bdc4772` passes all four native journal checks. Both fixed paced cohorts now pass the cadence threshold. Their input-to-presentation latency still fails, so full M3 and Mac release remain open.

| Workload | Input-to-presentation p95 | Intervals within 18.333 ms |
|---|---:|---:|
| Serial grid, CPU scene |34.801 ms| Diagnostic, not paced acceptance |
| Serial tree, GPU requested |52.648 ms| Diagnostic, not paced acceptance |
| Paced grid, GPU requested |52.576 ms|596/599 (99.499%) |
| Paced tree, CPU scene |50.431 ms|598/599 (99.833%) |

The unchanged policy requires at least 99% within 1.1 display periods and p95 input latency at most 33.3 ms. Both paced interval p95 values are 16.66679 ms. Grid's maximum 83.334 ms interval and 85.653 ms CPU schedule overrun remain in the original records; they are not discarded. Serial intervals do not qualify cadence. Automated routed input and native feedback do not establish physical/human input acceptance.

Paired source100→101 paced input-stage p95 falls 14.451→0.012625 ms for grid and14.044→0.010042 ms for tree; total UI work falls 18.483→3.597 ms and 20.165→4.891 ms. Native acquired-to-present p95 remains 36.263/35.911 ms. This identifies a remaining phase delay; it does not prove that a particular GPU queue policy causes it. A following Mac/Metal latency request is an experiment and requires its own native comparison.

The clean source101 private SDK archive 88dfe6812c4485df0f28c39cfb5373f4dce88d0b0bc042478d82c2073356dee4 has 8,281 verified files and manifest b71100f4e7ea6580f5a82ee44e64868c35cd92f80f5746b0e5e1f5b322b27550. Its relocation checks and all seven affected consumer prerequisites pass. Actual native runs occur after background compilation stops, under the exclusive live lease and the same 60 Hz host/configuration as source100. All 36 owned native/probe groups terminate and are observed absent; source/SDK, geometry and raw OS/display/power guards remain unchanged.

The full independent review and original journals are locally retained and SHA256-bound by evidence.json. Source99's refused journal and source100's valid but 30 Hz records remain unchanged. No input-source, VoiceOver, OS preference, personal application data, rendering backend, threshold or measured-population changes occur. Full performance matrix, broader workflows, native resource/endurance, physical/human and final packaging gates remain separate.
