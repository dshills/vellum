# Outstanding live acceptance checks

Run against the pinned fixture and record build/artifact hashes, platform, input source, scale and timestamps. This checklist is not passing evidence. Temporary Romaji/Pinyin/VoiceOver changes were already authorized by the user with restoration. Snapshot the current settings at the start of a future run; restore that snapshot and verify it even after failure. No setting changed during the foundationchecks run.

## macOS candidate panel

- Compose and convert Japanese Romaji and Chinese Pinyin in the field. Capture the entire desktop, including the separate OS candidate window.
- Verify panel anchor follows the visible caret after long horizontal scrolling, field/window movement, focus loss/regain and available scale/monitor changes. Record field/caret and panel bounds in comparable screen coordinates.
- Select a candidate by keyboard and pointer, cancel composition, replace a selected emoji/grapheme, then Undo/Redo. Verify final exact text, selection, marked range and no duplicate callbacks.
- Repeat multiline/mixed-direction native ranges after those paths are implemented; current single-field current-caret evidence is insufficient.

## VoiceOver

- With VoiceOver active, record its navigation cursor/focus and spoken role/name/value/state for field, multiline editor, button, checkbox, virtual rows and expanded/collapsed tree.
- Navigate forward/backward, invoke button/checkbox, edit/read selected text and confirm the domain state/action result. Speech must be observed or captured; AX value changes alone do not pass.
- Scroll virtual content and expand/collapse the tree. Verify stable focus/identity, disabled behavior and correct availability as objects leave/reenter the visible semantic tree.
- Close/reopen an actual Cogent/GPU window; verify no stale accessible object/focus, duplicate adapter or native callback after teardown. Check the before-show tree once a supported seam exists.

## Platform/distribution matrix

- Windows: declared minimum OS/GPU, D3D backend, native Japanese/Chinese input, Narrator, DPI/monitor movement, clipboard and clean packaged launch.
- Linux: declared minimum distribution/GPU, Wayland and X11, Vulkan, real IBus/Fcitx composition, Orca/AT-SPI, scale/focus, portal packaging and clean launch.
- macOS: Metal surface resize/minimize/restore/occlusion and repeated GPU destruction, cold font cache, trusted signed/notarized package on a clean user machine.
- Across hosts: 60-second static/minimized/occluded idle, tracked Go/native/GPU memory, cold/warm startup boundaries and lossless bounded queue behavior under a stalled UI consumer.

Record each failed or unavailable check independently; do not mark the platform complete based on a compile or shape-only result.
