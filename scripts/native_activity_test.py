import copy
import json
import unittest

from native_activity import validate_activity


def fixture():
    before = {"available": True, "process_id": 123, "empty_event_post_attempts": (1 << 60) + 7, "surface_present_attempts": (1 << 60) + 17, "saturated": False}
    after = {**before, "empty_event_post_attempts": before["empty_event_post_attempts"] + 3, "surface_present_attempts": before["surface_present_attempts"] + 2}
    return {"event": "vellum.soak.sample", "phase": "visible", "sample": 1,
            "start_utc": "2026-10-16T12:00:00.000000001Z", "end_utc": "2026-10-16T12:01:00.000000001Z",
            "app_originated_wakeups_available": False, "present_count_available": False, "full_p11_qualified": False,
            "framework_native_activity": {"version": 1, "available": True, "start": before, "end": after,
                "start_read": {"start_utc": "2026-10-16T11:59:59.999999997Z", "end_utc": "2026-10-16T11:59:59.999999999Z"},
                "end_read": {"start_utc": "2026-10-16T12:01:00.000000002Z", "end_utc": "2026-10-16T12:01:00.000000003Z"},
                "empty_event_post_attempts_delta": 3, "surface_present_attempts_delta": 2}}


class ActivityValidationTests(unittest.TestCase):
    def test_exact_large_integer_detached_readonly_validation(self):
        value = fixture()
        original = copy.deepcopy(value)
        result = validate_activity(json.loads(json.dumps(value)), 123)
        self.assertEqual((result["empty_event_post_attempts_delta"], result["surface_present_attempts_delta"]), (3, 2))
        self.assertIs(result["actual_wakeups_qualified"], False)
        self.assertIs(result["actual_presentations_qualified"], False)
        self.assertIs(result["full_p11_qualified"], False)
        self.assertEqual(value, original)

    def test_protocol_pid_type_count_and_qualification_mutants(self):
        mutations = [
            lambda v: v.update(phase=[]),
            lambda v: v.pop("start_utc"),
            lambda v: v.update(app_originated_wakeups_available=True),
            lambda v: v.update(present_count_available=True),
            lambda v: v.update(full_p11_qualified=True),
            lambda v: v.pop("framework_native_activity"),
            lambda v: v["framework_native_activity"].update(version=True),
            lambda v: v["framework_native_activity"].update(version=2),
            lambda v: v["framework_native_activity"].update(extra=1),
            lambda v: v["framework_native_activity"].update(available=False),
            lambda v: v["framework_native_activity"]["start"].update(available=1),
            lambda v: v["framework_native_activity"]["end"].update(saturated=True),
            lambda v: v["framework_native_activity"]["end"].update(process_id=124),
            lambda v: v["framework_native_activity"]["start"].update(process_id=True),
            lambda v: v["framework_native_activity"]["end"].update(empty_event_post_attempts=(1 << 64) - 1),
            lambda v: v["framework_native_activity"]["start"].update(surface_present_attempts=-1),
            lambda v: v["framework_native_activity"]["end"].update(surface_present_attempts=True),
            lambda v: v["framework_native_activity"]["start"].update(empty_event_post_attempts=1.0),
            lambda v: v["framework_native_activity"].update(empty_event_post_attempts_delta=2),
            lambda v: v["framework_native_activity"].update(surface_present_attempts_delta=2.0),
            lambda v: v["framework_native_activity"]["end"].update(empty_event_post_attempts=0),
        ]
        for mutate in mutations:
            value = fixture()
            mutate(value)
            with self.subTest(value=value), self.assertRaises(ValueError):
                validate_activity(value, 123)

    def test_actual_owned_pid_required(self):
        for pid in [None, 0, -1, True, 124, "123"]:
            with self.subTest(pid=pid), self.assertRaises(ValueError):
                validate_activity(fixture(), pid)

    def test_nanosecond_brackets_and_invalid_stamps(self):
        for side, key, stamp in [
            ("start_read", "end_utc", "2026-10-16T12:00:00.000000002Z"),
            ("end_read", "start_utc", "2026-10-16T12:01:00.000000000Z"),
            ("end_read", "end_utc", "2026-10-16T12:01:00.000000001Z"),
            ("start_read", "start_utc", "0001-01-01T00:00:00Z"),
            ("start_read", "start_utc", "2026-02-30T12:00:00Z"),
            ("start_read", "start_utc", "2026-10-16T12:00:00+00:00"),
            ("start_read", "start_utc", "٢٠٢٦-10-16T11:59:59.999999997Z"),
            ("start_read", "start_utc", "2026-10-16T11:59:59.٩٩٩٩٩٩٩٩٧Z"),
            ("start_read", "start_utc", 123),
        ]:
            value = fixture()
            value["framework_native_activity"][side][key] = stamp
            with self.subTest(stamp=stamp), self.assertRaises(ValueError):
                validate_activity(value, 123)


if __name__ == "__main__":
    unittest.main()
