# macOS calibration diagnostics

These fixtures implement the predeclared [matched comparison](../../specs/macos-calibration.md).
They add no framework API, renderer policy or Placement app feature. The Go
consumer uses the existing measured desktop session and native presentation APIs;
the standalone native program routes an internal Space event through AppKit.
Neither injects global keyboard input or measures physical input-to-photon time.

`check_test.py` exercises strict attribution/refusal. The framework's offscreen
`TestRoutedStateAndExactPatchPixels` checks actual routing and every scene pixel,
including an unrelated-key/no-op control. Offscreen passes cannot prove native
presentation. Run optimized native builds separately, with the immutable SDK
replacements and source-built Metal archive recorded in build metadata.

The native program accepts the setup pilot's locked match JSON. The framework
consumer accepts `pilot MATCH.json` (exclusive creation) or `timed MATCH.json`.
`run.py --artifacts ROOT --reuse EXISTING_ARTIFACTS --run-authorized` requires the
existing owned executor/lease/SDK integrity helpers and explicitly approved quiet
desktop time. Its source/binary/SDK identities and actual commands are retained.
This runner is a local investigation adapter, not a standalone public benchmark.

The payload is an opaque red/green rectangle on opaque white. Native precomputes
two textures, then blits; framework executes its normal view/paint/upload path.
Native requires a non-framebuffer-only drawable for blitting. This algorithm and
usage difference is explicit; compare response distributions for matched output,
not isolated rendering instructions or a supposed universal compositor floor.
Five pairs and all invalid captures remain visible. The release gate is unchanged.
