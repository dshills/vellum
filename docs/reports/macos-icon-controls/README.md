# Built-in icons and button presentation

Vellum adds a retained passive `ui.Icon`, 23 portable symbol names returned by
`ui.Symbols()`, icon-bearing Button/ToggleButton and primary, secondary, quiet
and compact toolbar-action button appearances. Unknown names/raw SVG/paths are
refused. Application styles and themes override native defaults. Unlabeled icons
are decorative; labeled icons have a passive Mac image role. Icon-only actions
require a nonblank accessible name. Icons add no action or Tab stop.
Toggle symbols occupy both accepted states; omission restores the ordinary native
checked-state indicator. Presentation changes retain widget, model and callbacks.
See the [catalog](../../controls.md) and [contract](../../../specs/controls.md).

Buttons now join the bounded window-scoped owner queue. Delivery uses the current
callback and disabled ancestry; removal and window closure revoke pending work.
This deliberately replaces the preceding synchronous global Application.Do route.
The keyed reconciliation fixture now registers its model window, queues before
callback replacement and drains after reordering. A separate closure fixture keeps
another window alive and proves its owner admission while origin actions are refused.

## Verification and review

Portable race, complete offscreen adapter race, focused icon/gallery/keyed race3,
adapter vet and native gallery BUILD ONLY pass. The full adapter execution takes
90.60 seconds. Tests parse and rasterize all 23 symbols, route actual Manual pointer
and keyboard input, compare completed appearance pixels, preserve incarnation and
accepted toggle state, exercise style override/restoration, validate owner delivery,
and check named/decorative semantics. Inert Cocoa checks the production AXImage
selector without creating a window. These are distinct from human VoiceOver,
physical input or native memory/performance acceptance.

Three declared mutation overlays compile and fail real assertions: an unscoped
synchronous Do route bypasses deferred delivery; skipping retirement delivers a
removed button; removing managed FinalStyler loses explicit icon size. The same
final executable/test leaves pass race3. These are localized faulty seams in the
new adapter, not complete original-production builds.

Internal source review approves the final 18 leaves. Prism runs once with Anthropic
Claude Sonnet 5.5, complete 18-file coverage, 3 chunks, 36,592 input/2,804 output
tokens, raw exit 0 and no findings. One trailing specification blank line caught
by the precommit whitespace check was removed as a final formatting correction.
The final 18-file source differs from the reviewed and executed baseline only by
removing that one terminal LF; all executable/test
leaves are identical. Separate final internal review binds this formatting change.
Prism and runtime checks are not repeated for the unchanged behavior; no claim is
made that earlier executions saw the later documentation byte.

Initial sandbox refusal for local Unix sockets, a compile error from using a
nonexistent Switch.SetIcon member, and the legacy unregistered-window/synchronous
fixture failure are preserved as historical attempts. Toggle artwork now uses
native IconOn/IconOff; the permitted socket suite and corrected fixture pass.
18 receipts/36 logs preserve 8 raw exit-1 attempts.
Exact original, candidate and final selected payloads and mutation inputs are
archived; the historical map lists early source-payload gaps. Binaries/caches and
full compiler/SDK/native dependency closure are excluded. The 6,198 prepared identities
remain unchanged and do not by themselves certify third-party artwork/asset rights.

Separate evidence review binds the [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json).
Physical Mac input/accessibility/display workflows, native/GPU memory, sustained
performance/endurance, distribution rights/signing/release and full Mac framework
completion remain open. The gallery was built without launch.
