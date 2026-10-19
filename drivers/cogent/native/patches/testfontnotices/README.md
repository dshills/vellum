# Test-font supplemental declarations

Typesetting stage0005 preserves all four shipped test fonts, the test readme and
the existing APACHE/Amiri OFL bytes. It adds five text files under `font/`, outside
fixture directories so the existing SDK notice collector carries them as
supplemental asset declarations. Nested font licenses do not satisfy the module's
own-license requirement. The generic fixture exclusion remains unchanged.

Roboto's full Apache text already ships in testdata; this patch makes an exact
copy discoverable. It likewise copies the existing Amiri-specific OFL. The other
two full texts are the original Microsoft Selawik and Canonical Ubuntu-family
declarations. `NOTICE.fonts.md` binds each exact font and accompanying text and
retains the observed copyright, declared-license and source identity.

Microsoft source: https://github.com/microsoft/Selawik/blob/89362e84731d1f5777fa078fd4d5ebbd339e4378/LICENSE.txt

Canonical source: https://github.com/canonical/Ubuntu-Sans-fonts/blob/9554af00fb9d438a12c916df8451c10dcedc9b7e/LICENCE.txt

The source texts are Git-blob/size/SHA256 checked against the pinned official
commit trees. The exact source-family notices are not proof of the shipped font
files' upstream build history. In particular Selawik's modified/subset history
and reserved-name permission remain open. No fonts, runtime Go/C code, embedded
font patterns, historical SDKs or distribution approval are changed.

Input hashes require all existing font/readme/declaration bytes and absence of
the five new paths. Output hashes bind those unchanged inputs and exact added
notice bytes. Preparation refuses drift instead of replacing an unknown text.
