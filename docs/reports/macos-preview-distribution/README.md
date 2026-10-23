# Developer-preview distribution decision inputs

This is a concrete release-owner decision packet, not an approved public release.
Framework MIT licensing and the github.com/dshills/vellum namespace were selected
by the owner. Both the source-built ARM64 WGPU archive and the matching Placement
source binary are bound in decision-inputs.json. Corrected original notices now
have 418 files and 175 conservative component entries, bound to that exact source
binary by inventory SHA-256. No existing tested archive was silently overwritten.

Declaration collection covers all 89 selected SDK Go graph records, 82 native
compiled-component records, Go standard-library candidate declarations and the
selected Placement embedded-asset ancestors. Actual Placement discovery found
4,441 nonstandard embeds in 35 modules. The corrected bundle adds Go's BoringSSL
candidate declaration and the additional Google Material Symbols attribution;
its Apache license text was already present. This is conservative notice coverage,
not a claim every inventoried component is linked or every applicability question
has been independently resolved. Asset source licenses are not replaced by MIT.

The concrete unresolved question is the Apple-SDK-derived objc2 binding family.
The pinned [upstream license note](https://raw.githubusercontent.com/madsmtm/objc2/objc2-0.6.4/LICENSE.md)
states MIT for objc2/block2/objc2-foundation/objc2-encode and selectable Zlib,
Apache-2.0 or MIT for other crates. It also expressly raises uncertainty about
distribution of Apple-SDK-derived bindings. Rechecking that note on 22 October
2026 did not remove its caveat. Apple's [Xcode and SDK agreement](https://www.apple.com/legal/sla/docs/xcode.pdf)
is the referenced primary agreement; simply reading it does not establish a
component-specific disposition or mean new terms were accepted here.

The existing release/version.json third_party_distribution_audit field remains
pending. The approved preview plan requires resolved distribution rights/notices
before public release; the SDK and newly guarded app producer both refuse current
pending approval, independently of valid matched notices. A local app refusal is
retained here. Presence of license files, this packet, Prism code review, framework
MIT selection and desktop-test authorization do not approve that audit.

Smallest owner decision: retain private delivery while obtaining the needed
upstream/rights review, or explicitly record the owner's completed distribution
audit decision for these exact contents and corrected notices. This is not a
request for blanket certification of unrelated binaries or future versions.
Without that decision, finish and deliver the tested private candidate and report
the public prerelease as blocked by the recorded audit. Do not change it silently.

If the distribution audit is approved, commit that decision, produce the final
public candidate through the guards, verify the actual extracted SDK/app and
applicable CI, and present their exact hashes/release text for separate final
publication authorization. No tag, upload, push, signing identity or notarization
has been authorized by this packet. The proposed unnotarized developer-preview
route must disclose ad-hoc signing and installation status explicitly.

Current remote CI is green at public baseline 0aa31bf. It does not qualify the new
local release commits. CI for the reviewed branch can run after source publication
is authorized; source reports include normalized home paths plus test timestamps,
process IDs and native window coordinates. Release assets remain a separate action.
