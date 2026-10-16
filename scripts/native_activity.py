"""Validate one current soak API-attempt observation, not whole-soak acceptance.

The caller must obtain owned_pid from its actual owned process receipt. PID
equality alone establishes neither process lifetime nor artifact authenticity.
"""
import calendar
import datetime
import re

_COUNTERS = {"available", "process_id", "empty_event_post_attempts", "surface_present_attempts", "saturated"}
_ACTIVITY = {"version", "available", "start", "end", "start_read", "end_read", "empty_event_post_attempts_delta", "surface_present_attempts_delta"}
_STAMP = re.compile(r"([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})(?:\.([0-9]{1,9}))?Z\Z")
_MAX = (1 << 64) - 1


def _require(condition, reason):
    if not condition:
        raise ValueError(reason)


def _stamp(value):
    _require(type(value) is str and 20 <= len(value) <= 30, "timestamp must be a bounded UTC RFC3339 string")
    match = _STAMP.fullmatch(value)
    _require(match is not None, "timestamp must preserve UTC nanoseconds")
    year, month, day, hour, minute, second = map(int, match.groups()[:6])
    date = datetime.datetime(year, month, day, hour, minute, second, tzinfo=datetime.timezone.utc)
    nanos = int((match.group(7) or "").ljust(9, "0"))
    _require(not (year == month == day == 1 and hour == minute == second == nanos == 0), "zero timestamp")
    return calendar.timegm(date.utctimetuple()) * 1_000_000_000 + nanos


def _counter(value, owned_pid):
    _require(type(value) is dict and set(value) == _COUNTERS, "counter schema mismatch")
    _require(value["available"] is True and value["saturated"] is False, "counter unavailable/saturated")
    _require(type(value["process_id"]) is int and value["process_id"] == owned_pid, "counter PID differs from owned process receipt")
    for name in ("empty_event_post_attempts", "surface_present_attempts"):
        _require(type(value[name]) is int and 0 <= value[name] < _MAX, "counter value invalid or saturated")


def validate_activity(sample, owned_pid):
    """Require exact counters and read brackets from one owned live process.

    This is a component for a source-bound native campaign validator. Other
    validators must still check process/source/terminal provenance, CPU interval,
    native window state, complete trace cardinality and all remaining budgets.
    """
    _require(type(owned_pid) is int and owned_pid > 0, "missing actual owned process PID")
    _require(type(sample) is dict and sample.get("event") == "vellum.soak.sample", "not a soak sample")
    _require(type(sample.get("phase")) is str and sample["phase"] in {"visible", "occluded", "minimized", "background"}, "unknown phase")
    _require(type(sample.get("sample")) is int and sample["sample"] > 0, "invalid sample ordinal")
    for key in ("app_originated_wakeups_available", "present_count_available", "full_p11_qualified"):
        _require(key in sample and sample[key] is False, "API attempts cannot qualify OS wakeups, presentations or P11")
    value = sample.get("framework_native_activity")
    _require(type(value) is dict and set(value) == _ACTIVITY, "activity schema mismatch")
    _require(type(value["version"]) is int and value["version"] == 1 and value["available"] is True, "activity version/unavailability")
    before, after = value["start"], value["end"]
    _counter(before, owned_pid)
    _counter(after, owned_pid)
    for name in ("empty_event_post_attempts", "surface_present_attempts"):
        delta = value[name + "_delta"]
        _require(type(delta) is int and 0 <= delta < _MAX and after[name] >= before[name] and delta == after[name] - before[name], "counter delta mismatch")
    times = []
    for name in ("start_read", "end_read"):
        window = value[name]
        _require(type(window) is dict and set(window) == {"start_utc", "end_utc"}, "read bracket schema mismatch")
        times.append((_stamp(window["start_utc"]), _stamp(window["end_utc"])))
    quiet_start, quiet_end = _stamp(sample.get("start_utc")), _stamp(sample.get("end_utc"))
    _require(times[0][0] <= times[0][1] <= quiet_start < quiet_end <= times[1][0] <= times[1][1], "counter reads do not bracket quiet interval")
    return {"version": 1, "owned_process_id": owned_pid,
            "empty_event_post_attempts_delta": value["empty_event_post_attempts_delta"],
            "surface_present_attempts_delta": value["surface_present_attempts_delta"],
            "actual_wakeups_qualified": False, "actual_presentations_qualified": False,
            "full_p11_qualified": False}
