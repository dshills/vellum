# Objective-C compiler-overlay check — offscreen only

A scalar-only reproduction with the diagnostic's exact **Go 1.25.6 darwin/arm64**
toolchain establishes why its supplied Objective-C acquisition hooks were absent
from the first linked binary. Go and C replacements take effect; the `.m` file
is compiled from its original physical contents. Materializing all replacements
in a physical source copy produces all three replacement values.

| Build | C marker | Objective-C marker | Go marker |
| --- | ---: | ---: | ---: |
| Original files | 11 | 31 | 41 |
| Compiler overlay | 22 | 31 | 42 |
| Physical replacement files | 22 | 32 | 42 |

The [probe source](execution-sources/probe.py.txt) generates three tiny fixtures,
compiles them with the same `-x objective-c` flag and exact existing toolchain
wrapper, then runs only scalar command-line programs. They do not import Cocoa,
Metal or a UI framework, create an application/window, or acquire a drawable.
All three builds and outputs pass; seven bounded owned commands, including the
version check, finish with complete logs and all groups retired. Network access
is denied. Original captures, the corrected binaries and source/configuration
pins remain unchanged. No broader test campaign or application build is repeated.

The installed toolchain's `cmd/go/internal/work/exec.go` omits `p.MFiles` from
its non-Go overlay lists, then compiles those files through their physical paths.
Its exact source hash is retained in the [result](result.json); build command
logs and fixture bytes are in the indexed [evidence archive](evidence.tar.gz).
Index keys name archive members. Its explicit `committed_copies` mapping proves
that archived `probe.py` equals committed `execution-sources/probe.py.txt`.
Absolute paths in result pins describe local execution provenance; the full
external dependencies and prior binaries are intentionally not published.
This is a version-specific observation, not a claim about every Go release or
all compiler-overlay uses.

The [already committed physical-copy preparation](../macos-presentation-physical/README.md)
removes this dependency on Objective-C overlays, and its actual linked hooks
and headless factory checks already pass. No new build repair is introduced here.
The missing new hooks do **not** explain why the original presentation observer
also recorded no startup samples. Successful corrected startup and a matched
native/framework latency result still require a live check; neither is claimed.

The user explicitly deferred the next live attempt: **keep working offscreen**.
The corrected capture remains unstarted. No screen use, native UI launch, settings
change or new live authority follows from this checkpoint. Features, the existing
Cogent/WebGPU/Metal design, shipping SDK/app, defaults and 33.3 ms release gate
are unchanged; M3–M7 remain open. Internal review and Prism using Anthropic
Claude Sonnet 5.5 cover this result; see the review completion receipt. Binaries,
cache, full toolchain, native archive, SDK and private Core are excluded from the
evidence bundle. No public push is requested.
