# macOS preview: native soak setup correction

22 October 2026. Test-fixture geometry only; product runtime, rendering, pacing,
features and release criteria are unchanged. Prior failed runs remain unchanged
in [the resource checkpoint](../macos-preview-resource-check/README.md).

The prior cover requested only 32 points of margin around the composed target.
Its outer frame contained the target, but native occlusion remained visible.
The existing fixture now requests 128 points on each side, moving the target
farther into the cover body. Translucent title-bar/rounded-edge exposure was the
working hypothesis; this control establishes effective geometry, not the precise
compositor mechanism. Native `OcclusionVisible=false`, front ordering, opacity,
level, full-frame containment and endpoint checks are all still required.

Focused `go test -race -tags vellum_native_checks ./cmd/window-cycle-smoke -run
Soak -count=1`, command vet and native build pass with Go 1.25.6. The 80-second
pilot passes all four transitions and completes one cycle; outer exit 0 in
92.212 seconds. In the occluded phase the actual target reports
`OcclusionVisible=false`, cover frame `[2675,-128,2200,1382]`, target frame
`[2803,0,1944,1126]`, full containment and correct ordering. Minimized and
background-worker transitions also pass. Native tracked windows/panels/callbacks
retire. The same one hidden Apple TUINS window remains; no native heap ownership
or GPU deallocation proof is claimed.

Five-second quiet samples are fixture preflight only, not eligible 60-second idle
samples or two-hour qualification. The full two-hour soak is the next check on
this exact fixture with unchanged frozen product inputs. No broad repository,
performance, or lifecycle campaign is repeated for this fixture-only geometry.
Home paths are normalized; test timestamps, process IDs and geometry remain.

## Final SDK archive binding

The first pilot above uses the checkout's upstream module-cache native archive
`7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248`;
it establishes fixture geometry only, not candidate resource acceptance.
For the long run, a fresh copy of final SDK `bb5bbba8…` was used. All 8,468 SDK
manifest file hashes match except the one fixture file `soak.go`, replaced by
this reviewed inset correction. The delivered SDK extraction stays untouched.
The external linker map confirms the selected ARM64 archive SHA
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
Candidate fixture binary SHA is
`7a73d546aedb6b04e7dde6b8e7edf526d16794249742e38a5fd11c4ceab74f8e`.

A second 80-second pilot using this exact candidate binary also passes all four
phases, one completed cycle, native tracked cleanup and normal process exit
(92.423 seconds). Its five-second samples still do not qualify the two-hour check.
This is the binary selected for the full run; no product input was substituted.
