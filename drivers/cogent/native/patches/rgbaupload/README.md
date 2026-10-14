# Checked RGBA upload rows

Owned patch0061 preserves RGBA parent/subimage row stride in queue writes.
Checked dimensions, row storage and final-row extent are admitted before native
texture creation/write. Empty/reversed bounds are refused before subtraction.
Single-row unused stride is normalized; final padding is not required. Pixel
bytes are borrowed only for the immediate queue write, without row packing.
Existing texture ownership/device/usage policy stays in place.

Six hundred deterministic layouts and malformed/overflow vectors accompany four
headless upload/copy/map cases, including odd padding and post-upload mutation.
The original tight-stride assumption fails three native cases. No native windows,
presentation or frame-performance qualification is implied. New tests are MIT.
Preparation checks original/output files and patch SHA256.
