# Prose presentation for exact multiline editors

Owner-approved presentation correction for Notes, based on the existing exact
multiline buffer and renderer. Code-editor shaping is unchanged by default.
The explicit prose mode hides the gutter, honors proportional widget typography,
and wraps source paragraphs using the same shaping engine that supplies painted
glyphs and native caret geometry. Visual wrap offsets preserve source bytes and
graphemes. View callbacks capture only immutable font/width inputs and the
thread-safe shaper; removing a buffer view releases the callback.

Plain prose is not a rich-text/markup editor. It requires wrapping. Code syntax,
fixed-column tab layout and line-number presentation are not supplied in prose
mode. Existing unsupported bidi-isolate interaction limits remain.

Offscreen regressions cover proportional advances, pixel bounds, exact source
fragments, pointer/caret mapping, visual navigation, resize reflow and retained
code/prose transitions at 160 and 200 logical DPI. Driver regressions preserve
native composition, CRLF bytes, binding identity and Undo/Redo. These checks do
not establish physical pointer or human IME/VoiceOver acceptance.
