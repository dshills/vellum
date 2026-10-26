#!/usr/bin/env python3
"""Create exclusively-owned, deterministic local files for Mac content checks."""
import argparse
import hashlib
import json
from pathlib import Path
import wave


def pdf_bytes():
    objects = [
        "<< /Type /Catalog /Pages 2 0 R >>",
        "<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>",
    ]
    for content in (6, 7):
        objects.append("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 400 500] "
                       "/Resources << /Font << /F1 5 0 R >> >> /Contents " + str(content) + " 0 R >>")
    objects.append("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
    for text in ("Vellum integration page one", "Vellum integration page two"):
        body = "BT /F1 20 Tf 30 400 Td (" + text + ") Tj ET\n"
        objects.append("<< /Length " + str(len(body)) + " >>\nstream\n" + body + "endstream")
    out = bytearray(b"%PDF-1.4\n")
    offsets = [0]
    for i, body in enumerate(objects):
        offsets.append(len(out))
        out.extend(f"{i + 1} 0 obj\n{body}\nendobj\n".encode("ascii"))
    xref = len(out)
    out.extend(f"xref\n0 {len(offsets)}\n0000000000 65535 f \n".encode("ascii"))
    out.extend("".join(f"{offset:010d} 00000 n \n" for offset in offsets[1:]).encode("ascii"))
    out.extend(f"trailer\n<< /Root 1 0 R /Size {len(offsets)} >>\nstartxref\n{xref}\n%%EOF\n".encode("ascii"))
    return bytes(out)


def create(destination):
    # Exclusive directory ownership: reruns never replace a user's file.
    destination.mkdir(parents=False, exist_ok=False)
    (destination / "two-pages.pdf").write_bytes(pdf_bytes())
    with wave.open(str(destination / "silent-60s.wav"), "wb") as audio:
        audio.setparams((1, 2, 8000, 0, "NONE", "not compressed"))
        audio.writeframes(bytes(8000 * 2 * 60))
    (destination / "reference.txt").write_text("Vellum disposable file-picker fixture.\n")
    files = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(destination.iterdir())}
    manifest = {"schema": 1, "purpose": "disposable local interaction fixtures", "files": files}
    (destination / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    return manifest


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path, help="new directory under an existing parent")
    args = parser.parse_args()
    print(json.dumps(create(args.output), indent=2))
