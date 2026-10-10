# Native input prototype: live Japanese/Chinese composition works; the foundation gate remains open

9 October 2026. Implemented the next bounded step in a separate [native-input prototype](../../prototypes/nativeinput/README.md). The field now renders marked text, receives native insertion/replacement and cancellation, groups composition for Undo, and supplies a visible current-caret rectangle to Cocoa. Real U.S. dead-key composition, native emoji selection replacement, Japanese Romaji and Chinese Pinyin conversion work. **G1 remains open** pending candidate-panel placement, full screen-reader workflows, native lifecycle support and the other recorded foundation gates.

## Implementation

The native [Cocoa client](../../prototypes/nativeinput/inputbridge/input_darwin.m) attaches to Cogent's exported GLFW view through a per-view Objective-C subclass. No GLFW source patch or global method swizzling is used. The inactive path and application shortcuts forward to the original implementation. A copied native text/selection state serves Cocoa callbacks; callbacks enqueue owned values through the existing unique-event transport. UTF-16 ranges reject surrogate interiors, invalid bounds, malformed UTF-8 and embedded NUL. Sequence acknowledgements prevent an older UI snapshot from replacing newer native callbacks.

[Patch 0004](../../prototypes/nativeinput/patches/0004-native-composition.patch), applied after the three remediation patches, supplies the field's native editing transaction and layout/render geometry notification. Provisional updates preserve the pre-composition value and undo history. Commit creates one undo transaction; cancel restores the original text and selection. Marked text is visibly underlined. Final carets and selection normalize to grapheme boundaries, while marked selections may use scalar boundaries inside a grapheme. Explicit value replacement clears composition.

Apple's client API supplies selection within inserted marked text ([setMarkedText documentation](https://developer.apple.com/documentation/appkit/nstextinputclient/setmarkedtext%28_%3Aselectedrange%3Areplacementrange%3A%29?changes=__9)). The bridge converts the rendered physical caret through view points to Cocoa screen coordinates and advertises its actual zero-length range; it does not fabricate arbitrary range geometry ([firstRect documentation](https://developer.apple.com/documentation/appkit/nstextinputclient/firstrect%28forcharacterrange%3Aactualrange%3A%29?changes=_2&language=objc)). Geometry notifications invalidate the input context's cached coordinates; cancellation clears the marked range and discards the conversion session ([input context documentation](https://developer.apple.com/documentation/appkit/nstextinputcontext/discardmarkedtext%28%29?changes=_2_9_1&language=objc)).

Text/selection update optimistically on Cocoa's thread; geometry is from the last completed Go layout/render. The subsequent publication invalidates coordinates so Cocoa can re-query. Immediate queries during a new marked callback may still reflect the prior rendered caret. Live CJK conversion now works; the asynchronous geometry still needs visual verification against actual candidate panels. Visible bounds constrain the insertion anchor; invisible geometry is reported unavailable. Arbitrary character rectangles, visual bidi navigation and point hit testing remain absent.

Native input is freed before AccessKit and view destruction. The view class is restored, observers/associations removed and Go handle deleted. This remains backend interposition with explicit subclass-order assumptions, attached after first show/focus. It is not yet a toolkit-supported native-input interface.

## Local evidence

The [race test run](../../prototypes/nativeinput/evidence/tests.txt) passes **23 top-level tests**, including UTF-16 scalar/interior validation, composition commit/Undo/Redo, selection-preserving cancel, focus commit, authoritative replacement and post-commit grapheme normalization. The retained graph, queue, grapheme, spelling and scheduler checks also pass. [Vet and the locked native build](../../prototypes/nativeinput/evidence/build.txt) succeed on the recorded macOS arm64 host.

Actual keyboard events using the configured U.S. input source produced these six [recorded samples](../../prototypes/nativeinput/evidence/live-keyboard.json):

| Action | Native observed field |
|---|---|
| Option-E then E after `A` | `Aé` |
| Command-Z | `A` |
| Command-Shift-Z | `Aé` |
| Option-E then Escape | `Aé`, retaining earlier edits |
| Select emoji in `A👩‍💻Z`, type `b` | `AbZ` |
| Command-Z | `A👩‍💻Z` |

Five native input callbacks were applied without rejection. See the [AX capture](../../prototypes/nativeinput/evidence/live-verified-ax.txt) and [screenshot](../../prototypes/nativeinput/evidence/live-verified.png). An early live run exposed Escape falling through after Cocoa committed a dead-key accent, causing Cogent to revert all field edits. The bridge now intercepts Escape while composition is active; the corrected run verifies preservation. Early `diagnostic-*` evidence is excluded from passing results. Switching applications between dead-key events also cancels the conversion session, so the uninterrupted commit test sends both keys before changing focus.

A [final native termination check](../../prototypes/nativeinput/evidence/termination.json) verifies that application commit saves `A´`, subsequent E produces `A´e`, and replacing an active preedit with `R` causes subsequent E to produce `Re`. Window reactivation retains logical field focus. A discovered identical-preedit/final-text bug is covered by a regression test: insertion/unmark now mark the committed value changed even when the visible string did not change in that last callback.

The [deterministic native protocol run](../../prototypes/nativeinput/evidence/protocol.json) applies ten callbacks: marked `に` → selected `日本` → commit; marked `取消` → cancellation; explicit replacement with `中文` → unmark commit; surrogate-interior replacement rejected without advancing sequence; then three callbacks in one burst producing `A中文漢👩Z`. Marked/selected ranges and returned substrings match the native state. Six public resize requests complete with the expected physical sizes and normal teardown. These calls invoke the actual client methods but do **not** constitute live Japanese/Chinese composition.

[Final current-caret diagnostics](../../prototypes/nativeinput/evidence/geometry-final.json) record native screen-coordinate queries and moving from rune offset 0 to 122 in a long field. The view anchor changes from X=16 to X=291, matching the visible scrolled caret in the [capture](../../prototypes/nativeinput/evidence/geometry-final.png). Real dead-key commit and cancellation also pass after scrolling. Attempts at vertical scrolling in this fixture did not establish a viewport change, so vertical scroll/candidate positioning remains unverified. Current-caret queries and six resizes are useful integration evidence, not a real IME candidate-window pass. The [manifest](../../prototypes/nativeinput/evidence/manifest.json) identifies source/artifact hashes and which build stage produced each capture.

## Required verification and decision

With explicit user approval, Japanese Romaji and Chinese Pinyin were temporarily enabled for a [live CJK run](../../prototypes/nativeinput/evidence/live-cjk.json). Actual key events produced these results:

| Source/action | Result |
|---|---|
| Pinyin `zhongwen`, Space | `A中文` |
| Undo / Redo | `A` / `A中文`, one composition transaction |
| Pinyin `ni`, Escape | Retains `A中文` |
| Romaji `nihongo`, conversion and Return | Appends `日本語` |
| Undo / Redo | Removes / restores only `日本語` |
| Romaji `ka`, Escape | Retains prior text |
| After horizontal scrolling: `kanji`, Space, Space, Down | Candidates change `漢字` → `感じ` → `幹事` |
| Escape after candidate cycling | Restores the original 122-rune field |

All **43 native callbacks** were acknowledged without rejection. The [Japanese conversion capture](../../prototypes/nativeinput/evidence/japanese-conversion.png) and [scrolled candidate capture](../../prototypes/nativeinput/evidence/japanese-scrolled-candidate.png) show the field's marked selection. The app-window capture excludes macOS's separate candidate panel, so its visual position is **unverified**. Live CJK resize/move and vertical-scroll checks remain open. The initial `nihongo` trial was actually Pinyin and produced `霓虹哦`; the `ni` prefix before the Japanese trial was entered in U.S. mode while switching sources. Both are retained in raw evidence. `input_environment` and `input_source` describe first show; each `input_probes` entry records the actual source at that sample.

VoiceOver was enabled and its first-run welcome completed. The process and Settings toggle confirmed it was active. The [bounded check](../../prototypes/nativeinput/evidence/live-voiceover-final.json) verifies native value/selection actions and keyboard replacement of selected `check` with `b`, producing `VoiceOver b` through marked/insert callbacks without rejection. However, automated Control-Option navigation/invoke keys produced no verifiable control change. Speech output was not observed. **This is not a full VoiceOver workflow pass**; the cause could lie in event delivery or the integration and remains unresolved. The earlier `live-voiceover` run expired while the welcome-prompt UI tool was blocked and is excluded from passing screen-reader evidence.

Original settings were restored and verified: only U.S. enabled/current, VoiceOver off, Globe key back to Show Emoji & Symbols, and Dictation back to English (United States). macOS had automatically added Japanese/Chinese Dictation languages when their input sources were enabled; those additions were also removed. See the [restored environment](../../prototypes/nativeinput/evidence/settings-restored-environment.json), [keyboard settings](../../prototypes/nativeinput/evidence/settings-restored-keyboard.txt), [VoiceOver setting](../../prototypes/nativeinput/evidence/settings-restored-voiceover.txt), and [live validation record](../../prototypes/nativeinput/evidence/live-validation.json). The runner itself does not change settings. Native test windows were closed normally.

Keep D as the experimental candidate; continue the bounded input/accessibility validation before the workbench. The new bridge demonstrates delivery without replacing GLFW's backend, but uses an unsupported interposition seam and local field patches. Seek supported creation/focus/destruction hooks and upstream acceptance; keep B as fallback if that ownership grows substantially. Multiline and bidi editing, arbitrary native text-range geometry, bounded lossless transport, clean/platform distribution, font-memory bounds and formal performance acceptance remain open. See [ADR 0004](../decisions/0004-native-input-prototype.md).
