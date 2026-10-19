# Core source quality checkpoint

Default Go vet must pass Core, events and the desktop/base driver packages under
the pinned toolchain and offscreen build tag, without disabling analyzers.
This scoped requirement does not assert a repository-wide or native build-tag
vet pass. The final M7 source audit and other Mac acceptance gates remain open.

Imported literal fields use explicit names while preserving their values/order.
The normal BenchmarkTable and BenchmarkForm setup, AssertRender and bm.N bodies
remain identical to the preceding checkpoint. Only unreachable profiling blocks
after unconditional returns are removed; benchmarks are not replaced with Skip
or shortened workloads. Existing targeted profiling in core/render.go remains.

Use fresh pinned preparation, default vet and existing relevant offscreen tests.
Compile/list both rendering benchmarks without executing them when the owner is
online. These checks establish no timing, native presentation or performance
result. Preserve earlier failed vet records and rejected performance samples.
Mac release still requires measured 60 Hz responsiveness; 4 ms is an optimization
target. Final native input, presentation, endurance and distribution checks remain.
