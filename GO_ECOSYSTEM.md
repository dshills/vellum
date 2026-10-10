# Go ecosystem and reusable foundations

Research date: 9 October 2026. This is a dependency selection survey, not a claim of benchmark results or compatibility testing. Exact inspected revisions and supplementary repository metadata are in [SOURCES.json](SOURCES.json).

## Selection rule

Prefer well-separated libraries already exercised by relevant applications over copying a framework's internal packages. Assess the native implementation and the Go binding separately. A mature C/Rust engine can have a young Go binding; a CGO-free wrapper can still require native shared libraries; recent commits establish activity, not reliability.

Licenses below are repository declarations inspected in source or GitHub metadata, not an assessment of every transitive dependency. Preserve notices and record native binaries, fonts, shaders and embedded assets separately. Entries marked “survey” received metadata/API or documentation review rather than a full implementation audit. `NOASSERTION` is treated as unresolved, not as permission.

## GPU candidates

| Candidate | Architecture / API | Capabilities and integration | Maintenance and risk evidence | Declared license / judgment |
|---|---|---|---|---|
| **Gio GPU** | Renders Gio ops through internal drivers | Metal, Vulkan, OpenGL, D3D11; public `Frame`, native render-target aliases | Integrated implementation; generic device API is `internal`; not a public compute abstraction | Unlicense OR MIT. Use with Gio; do not extract the internal driver as the default strategy. [API][GIO-GPU], [drivers][GIO-DRIVER] |
| **oliverbestmann/webgpu** | CGO binding to wgpu-native; explicit `Try…` errors alongside panic wrappers; cleanup support | Buffers, textures, graphics/compute passes, surfaces, WGSL via wgpu-native; native access through WebGPU C interface | Cogent Core's current module depends on it. Inspected HEAD 6 Oct 2026, upstream v29 stated in README. Latest fix concerns native string lifetime; FFI audit remains necessary | Apache-2.0; native wgpu has its own notices. Leading wgpu prototype binding, not yet a production selection. [Binding][WG-DEVICE], [build model][WG-README], [consumer][CC-MOD] |
| **cogentcore/webgpu** | Earlier CGO binding line | Native WebGPU plus JS branch; bundled platform libraries | README explicitly directs users to oliverbestmann and says this repository will not be updated | Apache-2.0. Retain only for existing compatibility; do not start new work here. [Maintenance notice][WG-OLD] |
| **go-webgpu/webgpu** | Go FFI through goffi; dynamically loads wgpu-native | Render/compute, textures, buffer mapping, surfaces, error scopes, timestamp queries | README labels it beta; inspected HEAD 25 Aug 2026; native v29 shared binary needed; callback/ABI tests exist in source | MIT. Prototype alternative if eliminating the C compiler matters; native packaging still required. [Loader][WG-FFILOADER], [requirements][WG-FFIREADME] |
| **gogpu/wgpu** | Public Go WebGPU-like API over native Go core/HAL; optional Rust FFI backend | Advertises Vulkan/Metal/DX12/GLES/software, graphics and compute; own shader-compiler ecosystem | Active v0.34.5 snapshot, 7 Sep 2026. Changelog records recent depth-pass, synchronization, and feature-validation fixes; broad conformance not independently established | MIT. Promising alternative; require real-driver and shader tests before relying on the native Go path. [Changelog][GG-CHANGE], [architecture][GG-README] |
| **SDL3 GPU through Go binding** | Explicit render/compute API in native SDL3; Go calls through CGO or FFI | Vulkan, D3D12, Metal; textures, transfer buffers, command buffers, graphics/compute passes; natural window integration | SDL native source is substantial; Go binding inspected describes itself as early/incomplete. Native stability does not prove binding completeness | SDL zlib; purego-sdl3 Unlicense. Strong competing prototype, particularly if one platform/GPU dependency is attractive. [GPU source][SDL-GPU], [binding][SDL-BINDGPU] |
| **Ebitengine** | Go 2D image/shader API over internal GPU backends | Excellent fit for sprites, images, batched triangles and Kage shaders; not a general public WebGPU/compute API | Established project with text integration and event-driven minimum frame mode; package-global window model needs scrutiny | Apache-2.0. Reuse for narrower 2D products or as a rendering baseline. [Main loop][EB-RUN], [text][EB-TEXT] |
| **Cogent Core GPU** | Higher-level graphics/compute management over maintained WebGPU binding | Pipelines, resources, compositing helpers and 3D packages | Real consumer of selected WebGPU binding; linked to Cogent Core utility/type conventions | BSD-3-Clause. Evaluate package-level reuse; importing it may bring a wider framework dependency closure. [GPU][CC-GPU], [system][CC-GPUSYSTEM] |
| **go-gl/gl + go-gl/glfw** (survey) | Generated OpenGL API + GLFW binding | Mature OpenGL-style rendering; shaders and resources managed by caller; desktop windows | Binding metadata active in 2026; OpenGL version/platform limits remain; CGO/native prerequisites | MIT + BSD-3-Clause binding; GLFW itself zlib. Useful baseline/compatibility route, weaker sole long-term foundation. [GL][LIB-GL], [GLFW binding][LIB-GLFW] |
| **vulkan-go/vulkan** (survey) | Low-level Vulkan binding | Broad Vulkan concepts, but applications own synchronization, allocation, presentation and shaders | Repository's last pushed metadata was Aug 2023; not archived, but maintenance/modern coverage requires verification | MIT. Does not solve three-backend burden; not recommended as initial foundation. [Repository][LIB-VULKAN] |

### WebGPU versus SDL3 GPU

wgpu-native is attractive for a shared UI/application GPU device, WGSL shaders, compute, explicit resources, and an existing Go consumer in Cogent Core. The cost is a sizeable native dependency and binding/version coordination. The maintained binding distributes platform static libraries as separate Go modules; a CGO build still needs a target compiler/linker and system frameworks. Its default panic wrappers conflict with a fully explicit-error public boundary, so the adapter should use `Try` methods and route asynchronous device failures explicitly. Cleanup/finalizers must be a leak safety net, not the resource-lifetime policy. [Packaging/error model][WG-README], [CGO device implementation][WG-DEVICE]

SDL3 GPU is an equally serious experiment, not an afterthought: one native platform dependency provides graphics/compute and windows, with backend-specific shader formats and an offline shader workflow. The main questions are custom shader ergonomics, Go API completeness, color/format capabilities, binding overhead, and whether its portability limits constrain target CAD/image workloads. Shadercross is a separate tool/library; do not quietly make a runtime shader compiler part of a supposedly minimal executable. [SDL GPU overview][SDL-GPUDOC], [GPU implementation][SDL-GPU]

For both paths, test texture sampling/render-target/compute format combinations, limits, timestamp support, readback, clipping, device loss, multi-window surfaces, and stable resource destruction. Explicitly test 16-bit float and sRGB/linear workflows; “WebGPU-compatible” is not a color-management strategy. Native driver availability and surface creation remain prerequisites.

## Windowing and native integration

| Library | Scope and evidence | Maturity / deployment / licensing assessment |
|---|---|---|
| **Gio app** | Windows/AppKit/X11/Wayland event/window layer, clipboard/editor state, native handles and CustomRenderer | Existing Go-native implementation. Inherits missing/unvalidated native services described in RESEARCH. CGO on macOS/Linux; Unlicense OR MIT. [Window][GIO-WINDOW] |
| **SDL3 + purego-sdl3** | SDL includes multiple windows, text editing events/input area, file dialogs, clipboard and tray subsystems; binding exposes representative calls and examples | Native layer reusable; Go binding early/incomplete. Dynamic SDL3 library required by inspected binding. Minimum described platforms are amd64/arm64 Windows/macOS/Linux/FreeBSD. Own a narrow adapter and fill only demonstrated gaps. [Binding status][SDL-BINDREADME], [text binding][SDL-BINDKEY], [dialog binding][SDL-BINDDIALOG] |
| **GLFW + go-gl/glfw** | Windows, contexts/surfaces, input, clipboard and native handles | Well-defined main-thread rules and X11/Wayland selection. Broad desktop text composition, native menus, accessibility and dialogs need additional work. Zlib native / BSD-3-Clause Go binding. [Thread rules][GLFW-THREAD], [binding][LIB-GLFW] |
| **ebitengine/purego** (survey) | FFI without CGO for supported platforms | Active 2026; Apache-2.0. Removes some compiler requirements, not foreign ABI/lifetime/thread constraints or library distribution. [Repository][LIB-PUREGO] |
| **golang.org/x/sys** (survey) | OS calls, especially useful Windows native interfaces | Go-maintained BSD-3-Clause package. Infrastructure, not a desktop abstraction. [Repository][LIB-SYS] |
| **godbus/dbus/v5** (survey) | D-Bus for portals, notifications and Linux services | Active 2026; BSD-2-Clause. Reuse transport rather than writing D-Bus; service protocols still need implementation. [Repository][LIB-DBUS] |
| **golang.design/x/clipboard** (survey) | Cross-platform clipboard helper | MIT, active 2026 metadata. Verify MIME coverage, image formats, session lifetime, Wayland behavior and external tools before adoption. Prefer the selected platform layer's clipboard to competing owners. [Repository][LIB-CLIPBOARD] |
| **sqweek/dialog** (survey) | Native desktop dialog helper | ISC; metadata active Jan 2026. Check platform/toolkit dependencies and ownership before combining with another event loop. [Repository][LIB-DIALOG] |
| **ncruces/zenity** (survey) | Dialog helper with platform implementations | MIT; active Aug 2026. Verify the exact release's Linux helper/runtime requirements; not a guarantee of dependency-free native dialogs. [Repository][LIB-ZENITY] |
| **fyne-io/systray** (survey) | Tray/menu integration | Apache-2.0; active Sep 2026. OS main-loop coordination and Linux desktop variability remain. Evaluate against SDL3's existing tray subsystem first. [Repository][LIB-TRAY], [SDL tray][SDL-TRAY] |

Do not choose five independent helper libraries that each believe they own the UI thread. The platform adapter should own lifecycle and coordinate native menus, dialogs, clipboard, notifications and windows through one event loop. Use capability reporting for optional facilities rather than silently failing.

## Text, fonts, Unicode, and images

| Library | What to reuse | Limits and maturity | Declared license |
|---|---|---|---|
| **go-text/typesetting** | OpenType parsing, HarfBuzz port, script/bidi/font segmentation, font scanning/fallback, shaping, line wrapping | Source-inspected; used by Gio, Fyne and Ebitengine according to upstream. v0 APIs evolve. Does not supply complete OS IME or application text-editor behavior. Cache shaped output above its reusable shaper. [Shaping][GT-SHAPING], [fonts][GT-FONTS] | Unlicense OR BSD-3-Clause; retain component notices |
| **go-text/render** (survey) | Candidate rendering helper for go-text glyphs | Smaller adjacent package, April 2026 activity. Compare supported glyph/color formats, CPU/GPU integration and cache control | GitHub classification was unresolved; inspect LICENSE and file notices before selection. [Repository][LIB-GTRENDER] |
| **golang.org/x/image** (survey) | Images, sfnt/font interfaces, fixed-point helpers, CPU drawing | Useful building blocks; basic font interfaces are not complex-script shaping or an editor | BSD-3-Clause. [Repository][LIB-IMAGE] |
| **golang.org/x/text** (survey) | Unicode/language/bidi-related utilities where needed | Coordinate versions with go-text; avoid composing incompatible Unicode boundary assumptions | BSD-3-Clause. [Repository][LIB-TEXT] |
| **rivo/uniseg** (survey) | Grapheme/word/line boundary candidate | MIT; last pushed metadata May 2024. Verify Unicode version against required fixtures; go-text already includes segmentation | MIT. [Repository][LIB-UNISEG] |
| **Go standard image packages** | Baseline decoding/encoding and image data | Appropriate for common formats; use decode limits and explicit color/alpha handling. Not a professional color-management engine | Go distribution BSD-style terms |
| **disintegration/imaging** (survey) | CPU resizing/filtering utilities | Last pushed metadata Sep 2023; stable utility candidate, not HDR/color-managed document engine | MIT. [Repository][LIB-IMAGING] |
| **srwiley/oksvg** (survey) | Small SVG icon rasterization candidate | Last pushed metadata May 2023; subset/feature coverage must be tested. Do not promise full SVG fidelity | BSD-3-Clause. [Repository][LIB-OKSVG] |
| **tdewolff/canvas** (survey) | Vector/path/text drawing and output formats | Active Sep 2026; evaluate backend, text and caching choices rather than treating it as a GPU widget toolkit | MIT. [Repository][LIB-CANVAS] |

Bundled fonts have independent redistribution licenses. System discovery should run with an embedded fallback available immediately; cold font scanning must not stall first paint. Atlas pages need budgets and eviction; glyph-cache keys must include face identity, variation, size and raster mode. Raster/color glyph support and fractional scaling are prototype questions. Rich text and code editing remain larger subsystems even with a good shaper.

## Accessibility

**AccessKit is the leading adapter candidate.** Its Rust implementation provides Windows UI Automation, macOS accessibility, and Unix AT-SPI infrastructure. The C binding provides an FFI route and examples; no ready, mature Go integration was established in this audit. Build a narrow Go/C adapter if selected, own callback thread handoff and tree lifetimes, and keep the semantic schema independent of AccessKit types. The inspected SDL C example uses SDL2 APIs and needs adaptation for SDL3 native handles. [AccessKit architecture][AK-README], [C bindings][AK-CREADME], [C example][AK-SDL]

AccessKit/AccessKit-C declare MIT OR Apache-2.0 with additional Chromium-derived notices. The binding release/build distribution must be pinned separately from SDL and wgpu. It is an additional native build artifact, not free functionality acquired by putting labels in a map.

Gio's semantics are reusable as a lesson but insufficient evidence for desktop accessibility. Fyne's inspected Windows/macOS accessibility paths are build-tagged and its other-target implementation is empty. GoGPU UI has semantic tree structures, but this audit did not establish corresponding native bridge completeness. Native screen-reader validation is a hard gate for every option. [Gio roles][GIO-SEMANTIC], [Fyne bridge][FY-ACCESS], [Fyne fallback][FY-ACCESSFALLBACK], [GoGPU tree][GG-A11Y]

## Layout and framework reuse

| Candidate | Reuse assessment |
|---|---|
| **Gio layout/widget** | Best immediate Go baseline. Layout contexts and widgets couple to operations/input; reuse as a stack or deliberately adapt a small primitive, not assume each package is independent. [Flex][GIO-FLEX], [Editor][GIO-EDITOR] |
| **Fyne** | Mature retained approach worth measuring for workbench/table requirements. Cell recycling, test facilities and UI dispatch are useful precedents. BSD-3-Clause. [Table][FY-TABLE], [dispatch][FY-THREAD] |
| **Cogent Core** | Existing retained Go/WebGPU stack with splits, tables/trees, text, 3D and native system packages. BSD-3-Clause. Test extending it before writing an equivalent renderer; inspect dependencies if reusing only GPU/system packages. [Widget][CC-WIDGET], [splits][CC-SPLITS], [module][CC-MOD] |
| **GoGPU UI** | Emerging retained/reactive Go toolkit with custom GPU view and semantic/test abstractions. MIT. Evaluate implementation and real applications rather than README comparative claims. [Widget][GG-WIDGET], [GPU view][GG-GPUVIEW] |
| **AllenDang/cimgui-go** (survey) | Active Oct 2026 binding route to ImGui and useful professional-tools baseline; MIT repository. Inherits native C++ build/version and ImGui feature limitations. [Repository][LIB-CIMGUI] |
| **inkyblackness/imgui-go** (survey) | Archived, last pushed Dec 2022 in inspected metadata. Avoid for new long-term foundation; license classification unresolved in metadata. [Repository][LIB-IMGUIOLD] |
| **Wails** (survey) | Active Go application shell with webview UI; MIT. Useful deployment/productivity comparison, but does not meet the requested native-rendered UI direction. [Repository][LIB-WAILS] |

A new Vellum layout engine should begin with constrained row/column/stack/grid and explicit custom measurement. There is no requirement to implement CSS layout. Generic Yoga/Taffy bindings would add another native boundary and semantic model; consider them only if a real layout workload defeats simpler primitives. A retained docking graph is separate from ordinary widget layout.

## Deployment reality

| Concern | Gio-based route | Independent SDL3 + wgpu route |
|---|---|---|
| Local developer build | `go build` after OS prerequisites; Windows path avoids CGO, macOS/Linux use it | `go build` after native binding artifacts/toolchain; maintained wgpu binding uses CGO; SDL purego path still needs a runtime library |
| Static versus dynamic | Native OS/driver libraries remain even with one executable | wgpu static archives can be linked; SDL/AccessKit linkage must be chosen and tested. CGO-free dynamic wrappers require DLL/dylib/so distribution |
| Windows delivery | Executable plus assets, DPI manifest/app identity; installer optional | Same plus selected native DLLs if dynamic. Clean-machine graphics/runtime checks and signing pipeline |
| macOS delivery | App bundle for normal desktop distribution, resources/Info.plist, target architectures | Same plus signing/loading of included native libraries and correct framework linkage |
| Linux delivery | Build dependencies and runtime display/GPU libraries; choose minimum distro baseline | Same plus native artifacts. Test Wayland and X11, portals and sandbox packaging; avoid compiling against only the newest glibc |
| Cross compilation | Native dependencies and target SDK complicate pure `GOOS/GOARCH` expectations | Matching static archives, linker, SDK and native ABI required; dynamic FFI simplifies some builds but not testing |
| GPU availability | Backend/driver dependent | Vulkan/Metal/D3D12 requirements; define unsupported-device and device-loss behavior |

The product goal should be **one ordinary Go build command after documented setup**, plus a reproducible packaging command and CI-produced artifacts. “One native executable” does not mean no OS frameworks, driver libraries, signing metadata, font assets, or optional services. macOS distribution should include a tested Developer ID signing/notarization workflow; Windows should include tested signing/distribution choices; Linux should have a stated runtime baseline. None was executed in this research.

## Recommended reuse shortlist

1. **Text:** go-text/typesetting directly; selected x/image/x/text helpers; bundled licensed fallback font.
2. **GPU experiment:** maintained oliverbestmann/wgpu-native binding versus SDL3 GPU. Keep go-webgpu and gogpu/wgpu as challengers if they pass the same ABI/driver workload.
3. **Platform experiment:** SDL3 through a narrow adapter; Gio full-stack as the principal low-cost control; Gio CustomRenderer as a separately tested shell option.
4. **Accessibility:** AccessKit C bridge, with Go glue and native tests budgeted explicitly.
5. **Domain-independent workbench:** build only the missing command/task/semantic/test/docking integration after baseline comparison.

Do not build a Unicode shaper, GPU backend family, shader compiler, D-Bus implementation, or cross-platform accessibility protocol stack. Do not depend on unmaintained GPU forks or assume internal framework packages are a stable public foundation.

<!-- Source reference definitions -->

[AK-CREADME]: https://github.com/AccessKit/accesskit-c/blob/8b6ed37c20ed4c59390e253407983333053662ba/README.md
[AK-README]: https://github.com/AccessKit/accesskit/blob/466e24e252103f4d82bd6fbab98141b551e46e0c/README.md
[AK-SDL]: https://github.com/AccessKit/accesskit-c/blob/8b6ed37c20ed4c59390e253407983333053662ba/examples/sdl/hello_world.c
[CC-GPU]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/gpu/gpu.go
[CC-GPUSYSTEM]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/gpu/system.go
[CC-MOD]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/go.mod
[CC-SPLITS]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/splits.go
[CC-WIDGET]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/widget.go
[EB-RUN]: https://github.com/hajimehoshi/ebiten/blob/dcb93b02c135c20ea835952e54e34f889e97de4d/run.go
[EB-TEXT]: https://github.com/hajimehoshi/ebiten/blob/dcb93b02c135c20ea835952e54e34f889e97de4d/text/v2/gotext.go
[FY-ACCESS]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/internal/driver/glfw/accessibility_windows.go
[FY-ACCESSFALLBACK]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/internal/driver/glfw/accessibility_notdarwin.go
[FY-TABLE]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/widget/table.go
[FY-THREAD]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/thread.go
[GG-A11Y]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/a11y/tree.go
[GG-CHANGE]: https://github.com/gogpu/wgpu/blob/052a7e526b23454334761537410a5e50406bdfe4/CHANGELOG.md
[GG-GPUVIEW]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/core/gpuview/widget.go
[GG-README]: https://github.com/gogpu/wgpu/blob/052a7e526b23454334761537410a5e50406bdfe4/README.md
[GG-WIDGET]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/widget/widget.go
[GIO-DRIVER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/internal/driver/driver.go
[GIO-EDITOR]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/widget/editor.go
[GIO-FLEX]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/layout/flex.go
[GIO-GPU]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/gpu.go
[GIO-SEMANTIC]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/semantic/semantic.go
[GIO-WINDOW]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/window.go
[GLFW-THREAD]: https://www.glfw.org/docs/latest/intro_guide.html#thread_safety
[GT-FONTS]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/fontscan/fontmap.go
[GT-SHAPING]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/shaping/shaping.go
[LIB-CANVAS]: https://github.com/tdewolff/canvas
[LIB-CIMGUI]: https://github.com/AllenDang/cimgui-go
[LIB-CLIPBOARD]: https://github.com/golang-design/clipboard
[LIB-DBUS]: https://github.com/godbus/dbus
[LIB-DIALOG]: https://github.com/sqweek/dialog
[LIB-GL]: https://github.com/go-gl/gl
[LIB-GLFW]: https://github.com/go-gl/glfw
[LIB-GTRENDER]: https://github.com/go-text/render
[LIB-IMAGE]: https://github.com/golang/image
[LIB-IMAGING]: https://github.com/disintegration/imaging
[LIB-IMGUIOLD]: https://github.com/inkyblackness/imgui-go
[LIB-OKSVG]: https://github.com/srwiley/oksvg
[LIB-PUREGO]: https://github.com/ebitengine/purego
[LIB-SYS]: https://github.com/golang/sys
[LIB-TEXT]: https://github.com/golang/text
[LIB-TRAY]: https://github.com/fyne-io/systray
[LIB-UNISEG]: https://github.com/rivo/uniseg
[LIB-VULKAN]: https://github.com/vulkan-go/vulkan
[LIB-WAILS]: https://github.com/wailsapp/wails
[LIB-ZENITY]: https://github.com/ncruces/zenity
[SDL-BINDDIALOG]: https://github.com/JupiterRider/purego-sdl3/blob/8523da70a0417adda4233281ab44f5cc5b157d9b/sdl/sdl_dialog.go
[SDL-BINDGPU]: https://github.com/JupiterRider/purego-sdl3/blob/8523da70a0417adda4233281ab44f5cc5b157d9b/sdl/sdl_gpu.go
[SDL-BINDKEY]: https://github.com/JupiterRider/purego-sdl3/blob/8523da70a0417adda4233281ab44f5cc5b157d9b/sdl/sdl_keyboard.go
[SDL-BINDREADME]: https://github.com/JupiterRider/purego-sdl3/blob/8523da70a0417adda4233281ab44f5cc5b157d9b/README.md
[SDL-GPU]: https://github.com/libsdl-org/SDL/blob/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/src/gpu/SDL_gpu.c
[SDL-GPUDOC]: https://wiki.libsdl.org/SDL3/CategoryGPU
[SDL-TRAY]: https://github.com/libsdl-org/SDL/tree/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/src/tray
[WG-DEVICE]: https://github.com/oliverbestmann/webgpu/blob/de8bdd5d5976db28883db41211e049446078f8fd/wgpu/device.go
[WG-FFILOADER]: https://github.com/go-webgpu/webgpu/blob/907b82e37b752ea9a2735e5beec4c9e9925f6315/wgpu/wgpu.go
[WG-FFIREADME]: https://github.com/go-webgpu/webgpu/blob/907b82e37b752ea9a2735e5beec4c9e9925f6315/README.md
[WG-OLD]: https://github.com/cogentcore/webgpu/blob/834052c4692498a8afff69fce01bdb8450ab4381/README.md
[WG-README]: https://github.com/oliverbestmann/webgpu/blob/de8bdd5d5976db28883db41211e049446078f8fd/README.md
