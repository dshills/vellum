# Current private SDK: native drawing, 3D and imaging

The private SDK from clean `5eef841` passes relocated generated-consumer
race/vet/build and public API race/vet checks, plus three separate race-enabled
native demo runs. This brings the corrected drawing, viewport startup, imaging worker cleanup
and stage-0077 key-publication ownership into one independently consumed SDK.
Framework code is unchanged by this evidence checkpoint.

The archive contains 7,441 hashed payload files and one manifest. Its exact member
set adds four stage-0077 patch leaves and one prepared regression leaf to the
previous SDK. All 7,336 included source/prepared counterparts match the committed
checkout, with eleven explicitly hashed Go module overlays. This denominator
includes examples and documentation; it differs from the earlier 7,027
framework-prefix-only baseline comparison. At verification, both checkouts had 4,738 tracked
source files and the exact tested 6,250-file prepared dependency vector.
The Go notice inventory contains 89 modules and remains a limited inventory,
with native/Rust/standard-library/font rights and distribution audit open.

The SDK archive is 29,404,606 bytes, SHA-256
`e1acb49f0941fca3f691aec35878e4f1c19ae15fcb37325718f0942b30bf5b69`.
The private ARM64 static GPU archive is 37,401,760 bytes, SHA-256
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
It remains a caller-supplied candidate, with no independent authenticity or public
distribution certification. No installed library is replaced.

The retained extraction has spaces in its path. Generated applications and all
three demos resolve exactly the SDK root, driver, prepared core/typesetting and
private GPU capsule. The three demo executables are Go 1.27.2 race-enabled thin
ARM64 Mach-O files, with minimum macOS 13.0 metadata and no WGPU dylib dependency.
Each map selects 141 exact packaged static archive members. Minimum deployment
metadata does not establish execution on older macOS or other hardware.

| Native SDK case | Actual result | Owned process group | Native streams |
| --- | --- | --- | --- |
| Drawing | Exit 0 in 7.805 s | 8099, retired and recorded absent | 1,460-byte stdout; one deliberate budget diagnostic |
| 3D viewport | Exit 0 in 5.540 s | 8219, retired and recorded absent | 12,493-byte stdout; three deliberate fault diagnostics |
| Imaging | Exit 0 in 6.721 s | 8348, retired and recorded absent | 4,044-byte stdout; empty stderr |

These durations describe bounded test executions, not frame or input latency.

Drawing checks retained images/meshes, gradient/linear-alpha and transparent
readbacks, camera/pan retention, budget refusal, error callback behavior,
independent windows, rebuild and repeated resize. Its schematic has 100,000 source
objects, with 1,195 tested and 1,000 visible in the sampled view. The waveform has
one million source samples. A 4,096-square image initially uploads 1,024 tiles;
one changed tile uploads 65,536 bytes. Cumulative upload counters are not native
memory bounds.

The viewport checks camera changes, depth ordering, independent buffer namespaces,
controlled preparation/draw failures, resource rebuild, resize and closure. Its
joint initial targets are 792×498 and 630×468, both at target generation 2, and
remain compatible with the chrome check. These assertions exercise the corrected
startup baseline instead of accepting provisional window geometry.

Imaging checks compute into RGBA16Float and tone-mapped straight-sRGB readbacks,
partial alpha/transparent pixels, admission refusal, sync/async capture, recursive
capture refusal, resize metadata, cancellation, rebuild, independent windows,
minimize/restore and closure. Reference exposures are RGBA `(213,156,0,128)` and
`(231,188,0,128)`. Delivery after resize was actually true; either correctly
preserved scheduling order is valid. Canceled/closed callback count is zero and
all six normal-close predicates pass. Private decoder joining and native heap
reclamation are not measured by this public consumer.

The first SDK build remains `FAILED_BUILD`. Seventeen direct commands exited
zero, then the harness rejected Go build metadata's legitimate relative GPU
replacement spelling. The matching Go module graph already reported the correct
absolute SDK origin. A fresh continuation resolves that spelling from the
consuming module and still requires the exact selected capsule. Eight focused
controls accept actual relative/absolute equivalents and reject missing,
duplicate, empty, wrong-relative and different-SDK replacements. The continuation
finishes retained drawing header/map validation and executes only the remaining
sixteen viewport/imaging commands. The archive, checker and drawing builds are
not rerun. The old failed overall status and raw streams remain preserved.

The fixed SDK Python CLI graph uses private test hosting in the outer controller's
owned process group. Three actual headless controls cover communication, the
child-creation-before-try gap and held cleanup, each with a TERM-ignoring child.
All groups become absent; held cleanup requires forced exit -9 after the TERM
grace. These are three cases with six actual processes. The intentionally stopped
controls retain `logs_complete: false`; they are not ordinary successful command
runs. This does not qualify standalone SDK helper timeout semantics or arbitrary
programs that escape their process group. The earlier 27 pure guard controls are
reused support, with unchanged guarded function bodies, not new test executions.
The initial signal-only wrapper drafts were not executed or used for SDK checks.

Packaging and direct Go/tool commands deny network access. Only the hosted
relocation checker permits local Unix IPC while denying all IP traffic. Native
cases use that same IP-deny/Unix-allow profile. Nested SDK scripts retain their
production child environment, including proxy/SumDB configuration; execution is
network-denied and cache-backed. Source/member identities are checked at command
launch and terminal observation, not continuously polled.

All direct build and native command streams, owners, raw failures, source vectors,
module/linkage metadata, harness controls and internal reviews are indexed in the
[evidence manifest](evidence.manifest.json) and [bounded archive](evidence.tar.xz).
The [result record](result.json) contains the native scalar summaries and individual
statuses. Principal executables, the SDK and static archive are separately
fingerprinted and retained in the durable local checkpoint after commit. The SDK
continues to identify tested source `5eef841`; later documentation commits do not
relabel it as rebuilt from those commits.

Internal reviews cover the source-unchanged packaging and native evidence. No new
Prism call is made: the significant implementation checkpoint already received
one Anthropic Claude Sonnet 5.5 review before `5eef841`.

OS input sources, VoiceOver, clipboard and preferences are unchanged. Fixture
geometry isolation is process-local. Physical pointer/keyboard and human
accessibility workflows, full-image/compositor presentation, cross-display
behavior, ordinary/sustained latency, native/GPU memory, endurance, native-service
and component/rights closure, signing/notarization, public distribution and full
Mac acceptance remain open.
