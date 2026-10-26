# Five macOS widget additions after personal v1

The owner authorized these five additions on 26 October 2026. This explicitly
opens only this widget scope after the completed private v1 freeze. Preserve the
stable v1 source contract and the accepted build9 archives. Backend, performance
policy, experimental rendering/pacing and public distribution remain unchanged.
Each phase receives focused tests, Prism review with Anthropic Claude Sonnet 5.5,
and its own commit. Additive APIs do not regenerate the v1 compatibility baseline.

1. Plain-text tooltips: delayed hover and F1 for a focused host, no focus stop,
   Mac light/dark surfaces, Escape/focus/pointer/lifecycle dismissal. At most
   4096 UTF-8 bytes; no application timers or goroutines. Existing accessible
   help is preserved. Automated routing is separate from physical pointer proof.
2. Link control: HTTP(S) URLs, keyboard activation and Open/Copy Link contextual
   actions. Application-owned actions; native URL opening through a worker-safe
   service. Preserve label and accessible URL; never interpret label markup.
3. Native Share button: AppKit picker for explicit text, HTTP(S) URLs or local
   files, system-provided service icons, exact parent lifetime, bounded items,
   cancellation and shutdown cleanup. Choosing a service is user consent to
   that service; completion of the picker is not proof of delivery. Unsupported
   hosts return ErrUnavailable. Do not send/share during automated checks.
4. File drop zone: bounded all-or-nothing file reference admission using existing
   FileDrop routing, extension validation, clear acceptance/rejection status and
   Choose File fallback. Application owns I/O. Finder pre-drop acceptance remains
   limited by the existing backend's eager OS acknowledgment.
5. Command palette: reusable controlled same-window search dialog with stable
   action IDs, groups, shortcut labels, original-color artwork, bounded results,
   Up/Down/Return/Escape and focus restoration. Availability must be rechecked at
   execution. Application command registry owns actual actions/history.

Live native appearance and physical-pointer evidence is recorded separately;
portable/offscreen tests never claim human, Finder or sharing delivery acceptance.
