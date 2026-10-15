# Noto Sans notice restoration

Patch 0063 adds the verbatim `OFL.txt` from the official Noto Sans v2.015 source
revision `c4a321e123e4d4ff315f57f4e0adf294fe3a95be`. Source:
https://github.com/notofonts/latin-greek-cyrillic/blob/c4a321e123e4d4ff315f57f4e0adf294fe3a95be/OFL.txt

The 4,396-byte notice has SHA256
`cee9892f9f0cc8fe882c9e9537ee6a89621d86ee7ceaf70b02e2b2b1c25c061a`
and Git blob `6843f31878c9945e1e71e1baa275546b4eefead8`. The source repository
root tree and downloaded notice bytes were compared. The pinned font's embedded
copyright identifies the same Noto project, and its version identifies 2.015.

The existing font README describes subsetting before the fonts entered the
pinned Cogent source. This patch preserves every font byte, README and embed
declaration. It restores the accompanying source notice; it does not reconstruct
or certify the original subsetting operation, its inputs or redistribution
approval. The output hash manifest binds all original Noto files and the new
notice. The input manifest requires that the notice does not already exist.

The SDK collector recognizes `OFL` declarations as supplemental notices. They
cannot satisfy a Go module's own-license gate. The overall third-party/native
distribution audit remains pending.
