#!/usr/bin/env python3
"""Check audit evidence integrity, without turning known limitations into passes."""
import json
import pathlib
import sys


def require(condition, message):
    if not condition:
        raise ValueError(message)


def verify(root):
    audit = json.loads((root / "acceptance-audit.json").read_text())
    require(audit["schema"] in (1, 2), "unsupported audit schema")
    order = audit["ordering"]
    require(order["fifo"] and order["applied_final"] == 32, "incomplete or non-FIFO native delivery")
    require(0 <= order["applied_at_shortcut"] <= 32, "missing shortcut observation")
    require(order["ordered_with_legacy_keys"] == (order["applied_at_shortcut"] == 32), "contradictory ordering status")
    require(order["max_applied_per_wake"] <= 16, "batch limit exceeded")
    if "legacy_ordering_control" in audit:
        control = audit["legacy_ordering_control"]
        require(control["applied_at_shortcut"] == 16 and control["applied_final"] == 32 and control["fifo"]
                and control["ordered_with_legacy_keys"] is False, "legacy control no longer reproduces the reported defect")
        require(order["native_batch_events"] == 2, "fixed audit batch count differs")
    long = audit["long_text"]
    local_pass = order["ordered_with_legacy_keys"] and all(long[key] for key in (
        "exact_roundtrip", "trailing_newline_deletion", "grapheme_movement"))
    status = int((root / "acceptance-audit.exit").read_text())
    require(status == (0 if local_pass else 1), "audit exit contradicts local assertions or reports a harness error")
    require(long["observed_rune"] == long["expected_rune"], "grapheme result disagrees")
    require(audit["G1_passed"] is False and audit["unavailable_checks"], "scope of this audit cannot establish G1")
    require(audit["font_scale"]["samples"] and audit["font_scale"]["samples"][-1]["unique_locations"] > 0, "empty font sample")
    native = json.loads((root / "native-queries-lifetime.json").read_text())
    require(native["cycles"] == native["attachments"] == native["detaches"] == 1000, "incomplete lifetime run")
    require(native["accepted"] == native["applied_after_close"] == 11000, "native updates lost")
    queries = native["publication_contracts"]["native_queries"]
    expected = [([0, 2147483647], "A👩Z\nאבג", [0, 8]),
                ([2, 1], "👩", [1, 2]), ([4, 3], "\nאב", [4, 3]),
                ([20, 1], "", [-1, 0])]
    require(len(queries) == len(expected), "missing native query")
    for query, (request, text, span) in zip(queries, expected):
        value = query["observed"]
        require(query["requested"] == request and value["substring"] == text and value["substring_range"] == span,
                "native text range differs from expected UTF-16 clipping")
        require(value["rect_range"] == [1, 0] and value["point_index"] == -1,
                "geometry capability changed; update acceptance scope explicitly")
    if "native_geometry" in native["publication_contracts"]:
        geometry = native["publication_contracts"]["native_geometry"]
        require(all(geometry[k] for k in ("invalid_atomic", "refused_edit_preserves_geometry", "native_edit_invalidates_geometry", "stale_ack_rejected", "screen_point_roundtrip")), "native geometry contract failed")
        expected_rects = [[0, 5], [1, 2], [5, 3], [4, 1], [3, 0], [-1, 0]]
        require([q["observed"]["rect_range"] for q in geometry["queries"]] == expected_rects, "native geometry range mismatch")
        require([q["observed"]["point_index"] for q in geometry["queries"][:3]] == [1, 1, 6], "native geometry hit mismatch")
        require(geometry["scaled"]["screen_rect"][2:] == [10, 10] and geometry["clipped"]["screen_rect"][2:] == [7, 10], "native geometry scale/clip mismatch")
    if audit["schema"] == 2:
        bidi = audit["bidi"]
        require(bidi["visual_navigation_established"] and bidi["right_arrow_rune_positions"] == [1, 2, 3, 4, 6, 5, 7], "visual bidi handler failed")
        require("native_geometry" in native["publication_contracts"], "missing native geometry evidence for schema 2")
    print("Evidence integrity passes; foundation acceptance remains open.")


if __name__ == "__main__":
    try:
        verify(pathlib.Path(sys.argv[1]))
    except (IndexError, KeyError, ValueError, OSError) as error:
        print(f"Invalid audit evidence: {error}", file=sys.stderr)
        sys.exit(2)
