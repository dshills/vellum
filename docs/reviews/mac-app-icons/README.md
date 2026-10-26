# macOS app icons review

Prism / Anthropic Claude Sonnet 5.5 reviewed the staged checkpoint twice, with
complete coverage. Both raw reviews are preserved; findings were dispositioned
against code, actual native checks and Apple's format contract.

Initial review:
- Move temporary conversion outside bundle: implemented with default temporary
  directory. The real-bundle-layout regression verifies no work folder remains
  beside Resources. Successful sample package regenerated with the new source.
- Structural-only ICNS validation: framing is an early check, deliberately not a
  decoder. install unconditionally calls iconutil to decode a detached candidate
  before writing AppIcon.icns. Added a native regression with valid framing/TOC
  but no images: packaging refuses and leaves no app.

Final review:
- Combined size/error conditions / purported unvalidated ICNS path: no bypass
  exists. regular_stream and unchanged guard the detached bytes; all install
  paths reach the mandatory iconutil decode, then require PNG output. Hash-only
  verify is subsequent mutation detection, not input admission. Separate wording
  for each admission failure is optional diagnostics, not a correctness defect.
- Missing 512/1024 representations in supplied ICNS: intentionally supported.
  Apple explicitly permits incomplete sets and selects the best supplied image:
  https://developer.apple.com/library/archive/documentation/GraphicsAnimation/Conceptual/HighResolutionOSX/Optimizing/Optimizing.html
  PNG conversion generates all ten required named standard/Retina entries and
  the native round-trip test verifies each dimension. Supplied ICNS uses Apple's
  decoder rather than imposing an incompatible hardcoded type/size allowlist.
  The no-images test accepts either native nonzero exit or our explicit empty-set
  rejection, since decoder diagnostics vary with macOS; both must refuse before
  publication. It does not claim one particular decoder implementation branch.

Checks: seven icon tests, 33 existing packaging tests, 37 script-portability
checks, shell syntax, unchanged starter Go build, real private ad-hoc bundle and
exact extracted ZIP signature, ten decoded Retina images, AppKit full color and
transparent-corner checks. No broad framework/native campaign was needed for
this Python packaging change; root APIs and Go behavior did not change.

Documents-stage signing/extraction refusal is documented in the checkpoint,
with successful local temporary packaging/extraction. No global metadata repair,
user-file modification, public release or human/native desktop qualification was
performed. See docs/reports/mac-app-icons for the package evidence and producers.
