"""Rename only the pinned invalid-GVAR test fixture; never repair its font data.

Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
The input/output font data is separately licensed under OFL-1.1; see OFL.txt.
"""
from pathlib import Path
import argparse
import hashlib
import json
import os
import stat
import struct

SOURCE_SHA256 = "720afadd389c79e39a02502e952fecb6f5d78eeca5f095c14b9963ead45a8f1c"
LICENSE_SHA256 = "77b7c2506d4efb22e09c8ccf10159f4956eab3ef7c007fef95de136bcf45300c"
FAMILY = "Vellum Invalid Variation Fixture"
PRIMARY_NAMES = {1: FAMILY, 3: FAMILY + "; revision 1", 4: FAMILY,
                 6: "VellumInvalidVariationFixture-Regular"}
OLD_NAMES = {1: "Selawik Variations test",
             3: "Version 1.020;MS  ;Selawik Variations test Regular",
             4: "Selawik Variations test", 6: "SelawikVariationstest-Regular"}
ASSET = "VellumInvalidVariationFixture.ttf"


def read_regular(path, bound, exact_size=None):
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0) | getattr(os, "O_NONBLOCK", 0)
    identity = lambda s: (s.st_dev, s.st_ino, s.st_size, s.st_mtime_ns, s.st_ctime_ns)
    descriptor = os.open(path, flags)
    try:
        # Keep one owner even when constructing the stream itself fails.
        with os.fdopen(descriptor, "rb", closefd=False) as stream:
            before = os.fstat(stream.fileno())
            if (not stat.S_ISREG(before.st_mode) or before.st_size > bound
                    or exact_size is not None and before.st_size != exact_size):
                raise ValueError("bounded regular file of the expected size required")
            data = stream.read(before.st_size)
            if len(data) != before.st_size or stream.read(1):
                raise ValueError("input size changed during read")
            if (identity(os.fstat(stream.fileno())) != identity(before)
                    or identity(os.stat(path, follow_symlinks=False)) != identity(before)):
                raise ValueError("input identity changed during read")
    finally:
        os.close(descriptor)
    return data


def digest(data):
    return hashlib.sha256(data).hexdigest()


def checksum(data):
    padded = data + bytes((-len(data)) % 4)
    return sum(struct.unpack(">" + "I" * (len(padded) // 4), padded)) & 0xffffffff


def rename_fixture(source):
    if len(source) != 454405 or digest(source) != SOURCE_SHA256:
        raise ValueError("exact pinned original fixture required")
    if source[:4] != b"\x00\x01\x00\x00" or struct.unpack_from(">H", source, 4)[0] != 28:
        raise ValueError("unexpected pinned SFNT header")
    tables = {}
    for index in range(28):
        tag, _, offset, length = struct.unpack_from(">4sIII", source, 12 + index * 16)
        if tag in tables or offset < 460 or offset + length > len(source):
            raise ValueError("invalid pinned SFNT directory")
        tables[tag] = source[offset:offset + length]

    old_name = tables[b"name"]
    version, count, strings_at = struct.unpack_from(">HHH", old_name)
    if version != 0 or count != 26 or strings_at < 6 + count * 12:
        raise ValueError("unexpected pinned name table")
    records, strings, changed = bytearray(), bytearray(), set()
    for index in range(count):
        platform, encoding, language, name_id, length, offset = struct.unpack_from(">6H", old_name, 6 + index * 12)
        if strings_at + offset + length > len(old_name):
            raise ValueError("invalid pinned name record")
        raw = old_name[strings_at + offset:strings_at + offset + length]
        if name_id in PRIMARY_NAMES:
            if ((platform, encoding, language) != (3, 1, 1033)
                    or name_id in changed or raw.decode("utf-16-be") != OLD_NAMES[name_id]):
                raise ValueError("unexpected primary identifier")
            raw = PRIMARY_NAMES[name_id].encode("utf-16-be")
            changed.add(name_id)
        records += struct.pack(">6H", platform, encoding, language, name_id, len(raw), len(strings))
        strings += raw
    if changed != PRIMARY_NAMES.keys():
        raise ValueError("missing primary identifiers")
    unchanged_name_records = count - len(changed)
    tables[b"name"] = struct.pack(">HHH", 0, count, 6 + len(records)) + records + strings

    # Keep every other table byte-for-byte, including the intentionally invalid
    # empty glyf/loca data and its incompatible gvar. Only the SFNT checksum
    # adjustment in head is rewritten for the newly named container.
    head = bytearray(tables[b"head"])
    head[8:12] = bytes(4)
    tables[b"head"] = bytes(head)
    count = len(tables)
    selector = count.bit_length() - 1
    search_range = 16 << selector
    result = bytearray(struct.pack(">4s4H", source[:4], count, search_range,
                                   selector, count * 16 - search_range))
    result += bytes(count * 16)
    for index, tag in enumerate(sorted(tables)):
        data = tables[tag]
        offset = len(result)
        struct.pack_into(">4sIII", result, 12 + index * 16, tag, checksum(data), offset, len(data))
        if tag == b"head":
            head_offset = offset
        result += data
        result += bytes((-len(result)) % 4)
    struct.pack_into(">I", result, head_offset + 8, (0xb1b0afba - checksum(result)) & 0xffffffff)
    result = bytes(result)
    manifest = {"schema": 1, "license": "OFL-1.1", "scope": "Renamed invalid-GVAR regression fixture; not a runtime font",
                "generator_sha256": digest(Path(__file__).read_bytes()),
                "source_sha256": SOURCE_SHA256, "output_sha256": digest(result),
                "output_bytes": len(result), "asset": ASSET,
                "primary_names": PRIMARY_NAMES, "changed_name_ids": sorted(changed),
                "unchanged_other_name_records": unchanged_name_records,
                "non_name_table_rule": "Exact payload identity except head bytes8:12 checksum adjustment",
                "table_payloads_sha256": {tag.decode("ascii"): digest(data) for tag, data in sorted(tables.items())}}
    # The manifest's head identity deliberately normalizes checkSumAdjustment.
    return result, manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path, help="New directory; never overwrites")
    args = parser.parse_args()
    source = read_regular(args.source, 454405, exact_size=454405)
    data, manifest = rename_fixture(source)
    license_data = read_regular(Path(__file__).with_name("OFL.txt"), 65536)
    if digest(license_data) != LICENSE_SHA256:
        raise ValueError("original OFL declaration changed")
    args.output.mkdir(exist_ok=False)
    (args.output / ASSET).write_bytes(data)
    (args.output / "OFL.txt").write_bytes(license_data)
    (args.output / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
