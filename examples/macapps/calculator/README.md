# Calculator

A bounded basic calculator made from public Vellum controls. New returns its
UI-owner Model; View, Theme and Commands plug into a normal native driver.
Use `go run ./cmd/calculator` from the parent module.

The circular keypad follows the observed basic Calculator on macOS Tahoe 26.6.2.
Buttons and keyboard commands share Press. It supports decimal entry, arithmetic
precedence, sign, contextual percent, deletion, repeated equals and divide-by-zero
recovery. Intermediate arithmetic uses rational numbers; displayed results use
11 significant digits. This is not a scientific calculator. Input tokens,
expression length and the 32-entry history are bounded. Scientific exponents are limited to ±300 before parsing; rational numerators/denominators are checked against a 4096-bit budget after every operation. History returns a
detached snapshot. Command-Y opens a result-recall dialog; Escape clears entry.

The unified header uses genuine OS window controls. The layout reserves the
measured 52Dp native header for the tested macOS configuration; it is not adaptive
safe-area support. Apple's native history/mode toolbar items, scientific/programmer
modes, unit conversion and persistent history are omitted. No Apple artwork or
user calculator history is embedded. Exact visual acceptance remains separately
tracked; the presence of a runnable example is not a pixel-identity claim.

Tests exercise precedence, decimal arithmetic, repeated equals, percent, sign,
operator replacement, deletion, error recovery, malformed exponent prevention,
keyboard/control paths, history recall/copying and input/history bounds.

The native runner requests a fixed 240 × 420 Dp window with implicit root
scrolling disabled. The full keypad fits this viewport; it does not rely on
scrolling to reach operators or Equals. Window resize handles and the native
zoom/fullscreen control are disabled through the creation policy. History stays
in a separate bounded dialog and does not expand the calculator window.
