"""Generate the original full-color starter icon with Python's standard library."""
from pathlib import Path
import math
import struct
import zlib

size = 1024
rows = bytearray()
for y in range(size):
    rows.append(0)
    for x in range(size):
        # Inset rounded blue tile, an ivory page and a warm folded corner.
        edge = math.hypot(max(abs(x - 511.5) - 286, 0), max(abs(y - 511.5) - 286, 0))
        alpha = round(255 * max(0, min(1, 176 - edge)))
        color = (30, round(155 - 70 * y / size), round(240 - 30 * y / size), alpha)
        if 294 <= x < 730 and 218 <= y < 806:
            color = (250, 249, 244, 255)
            if y < 358 and x > 590 and x - 590 > y - 218:
                color = (255, 190, 60, 255)
            if 366 <= x < 646 and any(a <= y < a + 24 for a in (416, 502, 588, 674)):
                color = (77, 123, 182, 255)
        rows.extend(color)

def chunk(kind, body):
    return struct.pack(">I", len(body)) + kind + body + struct.pack(">I", zlib.crc32(kind + body))
image = (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 6, 0, 0, 0)) +
         chunk(b"IDAT", zlib.compress(rows, 9)) + chunk(b"IEND", b""))
Path(__file__).with_name("app-icon.png").write_bytes(image)
