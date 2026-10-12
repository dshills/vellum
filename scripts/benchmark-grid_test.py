"""CPU-only statistical/record/owned-process fixture; never creates a window."""
import copy
import importlib.util
import io
import json
import os
from pathlib import Path
import tempfile
import threading
import types
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("benchmark_grid", Path(__file__).with_name("benchmark-grid.py"))
benchmark = importlib.util.module_from_spec(spec)
spec.loader.exec_module(benchmark)


def records(samples=120):
    identity = {"rows": 10000, "variable": False, "novel": False}
    result = [{"event": "grid.perf.setup", **identity, "release": True, "seed": 7, "stored_value_bytes": 10000*160},
              {"event": "grid.perf.ready", **identity}]
    for index in range(samples):
        result.append({"event": "grid.perf.sample", "index": index, "tag": index+1, "generation": index+1,
                       "frame": index+1, "callbacks": 1,
                       "native_sample": {"Tag":index+1,"Generation":index+1,"Window":"perf","Completed":True,"Failure":""},
                       "geometry": {"cells": 400, "rows": 50, "columns": 8, "scale": 1,"materialized_cells":432,"retained_cells":440,"retained_headers":9, "layout_sha256": "a"*64},
                       "metrics": {key: index+1 for key in benchmark.METRICS}})
    for sample in result[2:]:
        sample["metrics"]["ui_work_ns"] *= len(benchmark.WORK_PHASES)
        sample["native_sample"].update({native: sample["metrics"][name] for name, native in benchmark.NATIVE_PHASES.items()})
    result.append({"event": "grid.perf.end", **identity, "samples": samples, "callbacks": samples, "pending": 0, "dropped": 0})
    return result


def source_record(**overrides):
    value = {"Composition": 1, "Source": 0, "Kind": 1, "DrawNS": 90, "RasterNS": 40, "CopyNS": 30,
             "RasterAvailable": True, "CopyAvailable": True, "ImageWidth": 10, "ImageHeight": 20,
             "ImageChanged": True, "ImageUploadTimingAvailable": True, "SetFromGoImageNS": 20,
             "SetFromGoImageAttempts": 1, "ImageCacheHits": 0, "LogicalUploadRGBABytes": 800}
    value.update(overrides)
    return value


def detailed_records(samples=120, warmup=128):
    data = records(samples)
    for record in data: record["source_details"] = True
    data[1]["warmup"] = warmup
    for sample in data[2:-1]:
        sample["metrics"]["raster_upload_ns"] = 100
        sample["metrics"]["ui_work_ns"] = sum(sample["metrics"][field] for field in benchmark.WORK_PHASES)
        sample["native_sample"].update({native: sample["metrics"][name] for name, native in benchmark.NATIVE_PHASES.items()})
        sample["native_sample"]["SourceDetails"] = {"Compositions": 1, "UploadFailuresAvailable": False,
                                                   "Records": [source_record()]}
    return data


class SourceDetailTests(unittest.TestCase):
    def validate(self, data):
        return benchmark.validate_records(data, 10000, False, 120, False, allow_source_details=True)

    def test_historical_ordinary_records_and_explicit_false_remain_ordinary(self):
        data = records()
        self.assertEqual(len(benchmark.validate_records(data, 10000, False, 120, False)), 120)
        for r in data: r["source_details"] = False
        for r in data[2:-1]: r["native_sample"]["SourceDetails"] = None
        self.assertEqual(len(benchmark.validate_records(data, 10000, False, 120, False)), 120)

    def test_explicit_diagnostic_is_separate_from_profile_and_ui_total(self):
        data = detailed_records()
        for options in ({}, {"allow_profile": True}):
            with self.assertRaisesRegex(ValueError, "diagnostic"):
                benchmark.validate_records(data, 10000, False, 120, False, **options)
        selected = self.validate(data)
        first = selected[0]
        self.assertEqual(first["metrics"]["ui_work_ns"], 106)
        self.assertEqual(first["native_sample"]["SourceDetails"]["Records"][0]["LogicalUploadRGBABytes"], 800)
        # Nested90/40/30/20 durations remain raw diagnostics; none are added to
        # the original100 raster/upload interval or106 total work sample.
        self.assertEqual(benchmark.summarize([selected])["ui_work_ns"]["pooled"]["min"], 106)

    def test_known_skipped_raster_cache_hit_and_unknown_multiple_image_calls(self):
        data = detailed_records()
        detail = data[2]["native_sample"]["SourceDetails"]
        detail["Records"] = [source_record(DrawNS=40, RasterNS=0, CopyNS=20, ImageChanged=False,
                                          SetFromGoImageNS=0, SetFromGoImageAttempts=0, ImageCacheHits=1,
                                          LogicalUploadRGBABytes=0),
                             source_record(Source=1, Kind=0, DrawNS=50, RasterNS=0, CopyNS=0,
                                          RasterAvailable=False, CopyAvailable=False, ImageWidth=0, ImageHeight=0,
                                          SetFromGoImageNS=20, SetFromGoImageAttempts=2, ImageCacheHits=3,
                                          LogicalUploadRGBABytes=1600)]
        self.assertEqual(len(self.validate(data)), 120)

    def test_prefix_compositions_share_exact32_ticket_cap(self):
        rows = [source_record(Composition=i//16+1, Source=i%16, DrawNS=1, RasterNS=0, CopyNS=0,
                              ImageUploadTimingAvailable=False, SetFromGoImageNS=0,
                              SetFromGoImageAttempts=0, ImageCacheHits=0, LogicalUploadRGBABytes=0)
                for i in range(32)]
        detail = {"Compositions": 2, "UploadFailuresAvailable": False, "Records": rows}
        benchmark.validate_source_details(detail, 32)
        rows.append(source_record(Composition=3))
        detail["Compositions"] = 3
        with self.assertRaisesRegex(ValueError, "capacity"):
            benchmark.validate_source_details(detail, 1000)

    def test_mode_metadata_and_presence_corruption_refuse(self):
        changes = [lambda d:d[0].update(source_details="true"), lambda d:d[1].update(source_details=False),
                   lambda d:d[-1].update(source_details=False), lambda d:d[3].update(source_details=False),
                   lambda d:d[2]["native_sample"].pop("SourceDetails"),
                   lambda d:d[1].update(warmup=2049), lambda d:d[1].update(warmup=True)]
        for change in changes:
            data = detailed_records(); change(data)
            with self.assertRaises(ValueError): self.validate(data)
        ordinary = records()
        ordinary[2]["native_sample"]["SourceDetails"] = {"Compositions": 1, "UploadFailuresAvailable": False, "Records": []}
        with self.assertRaisesRegex(ValueError, "diagnostic"):
            benchmark.validate_records(ordinary, 10000, False, 120, False, allow_source_details=True)

    def test_detailed_capacity_counts_both_warmup_and_sample_tickets(self):
        data = detailed_records(samples=2048, warmup=2048)
        with self.assertRaisesRegex(ValueError, "capacity"):
            benchmark.validate_records(data, 10000, False, 2048, False, allow_source_details=True)

    def test_missing_repeated_skipped_and_invalid_ordinals_refuse(self):
        changes = [lambda d:d.update(Compositions=0), lambda d:d.update(Compositions=2),
                   lambda d:d["Records"][0].update(Composition=0), lambda d:d["Records"][0].update(Composition=2),
                   lambda d:d["Records"][0].update(Source=1), lambda d:d["Records"][0].update(Source=32),
                   lambda d:d["Records"].append(copy.deepcopy(d["Records"][0]))]
        for change in changes:
            data = detailed_records(); detail = data[2]["native_sample"]["SourceDetails"]; change(detail)
            with self.assertRaises(ValueError): self.validate(data)

    def test_scalar_schema_availability_and_nested_interval_corruption_refuse(self):
        changes = [{"Kind": 5}, {"DrawNS": True}, {"DrawNS": -1}, {"DrawNS": 1<<63}, {"ImageChanged": 1},
                   {"DrawNS": 60}, {"CopyNS": 10}, {"RasterAvailable": False}, {"CopyAvailable": False},
                   {"ImageUploadTimingAvailable": False}, {"SetFromGoImageAttempts": 0},
                   {"ImageWidth": -1}, {"ImageWidth": 1<<63}, {"ImageWidth": 1<<62},
                   {"LogicalUploadRGBABytes": 804}, {"LogicalUploadRGBABytes": 801},
                   {"ImageCacheHits": 1}, {"SetFromGoImageAttempts": 1<<64}, {"user_text": "forbidden"}]
        changes.append({"SetFromGoImageAttempts": 0, "SetFromGoImageNS": 0,
                        "LogicalUploadRGBABytes": 0, "ImageCacheHits": 1})
        for change in changes:
            data = detailed_records(); data[2]["native_sample"]["SourceDetails"]["Records"][0].update(change)
            with self.subTest(change=change), self.assertRaises(ValueError): self.validate(data)
        for change in (lambda d:d.update(UploadFailuresAvailable=True), lambda d:d.update(Records=None),
                       lambda d:d.update(Records=[]), lambda d:d.update(Image=None)):
            data = detailed_records(); change(data[2]["native_sample"]["SourceDetails"])
            with self.assertRaises(ValueError): self.validate(data)

    def test_native_ticket_scalar_corruption_cannot_hide_behind_numeric_equality(self):
        for field, value in (("Tag", True), ("Generation", 1.0), ("InputNS", True), ("InputNS", 1.0)):
            data = detailed_records(); data[2]["native_sample"][field] = value
            with self.subTest(field=field, value=value), self.assertRaises(ValueError): self.validate(data)
        for field, value in (("tag", True), ("index", False), ("callbacks", True)):
            data = detailed_records(); data[2][field] = value
            with self.subTest(field=field), self.assertRaises(ValueError): self.validate(data)

    def test_ticket_draw_and_counter_overflow_refuse(self):
        detail = {"Compositions": 1, "UploadFailuresAvailable": False, "Records": [source_record()]}
        with self.assertRaisesRegex(ValueError, "draw sum"):
            benchmark.validate_source_details(detail, 89)
        for field in ("SetFromGoImageAttempts", "ImageCacheHits", "LogicalUploadRGBABytes"):
            rows = [source_record(Source=i, Kind=0, DrawNS=1, RasterNS=0, CopyNS=0, RasterAvailable=False,
                                  CopyAvailable=False, ImageWidth=0, ImageHeight=0, SetFromGoImageNS=0,
                                  SetFromGoImageAttempts=1, ImageCacheHits=0, LogicalUploadRGBABytes=0) for i in range(2)]
            for row in rows: row[field] = benchmark.UINT64_MAX - (3 if field=="LogicalUploadRGBABytes" else 0)
            detail["Records"] = rows
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, "sum overflow"):
                benchmark.validate_source_details(detail, 2)

    def test_nested_mode_cannot_claim_normal_warm_targets(self):
        cases = {"fixed-10000": {"timings_ns": {"ui_work_ns": {"per_run": [{"p95": 1, "p99": 1}]}}}}
        self.assertTrue(benchmark.warm_target_met(cases))
        self.assertFalse(benchmark.warm_target_met(cases, source_details=True))
        self.assertFalse(benchmark.warm_target_met(cases, novel=True))
        self.assertFalse(benchmark.warm_target_met({}))


class StatisticsTests(unittest.TestCase):
    def test_nearest_rank_outliers_are_retained(self):
        result = benchmark.distribution(list(range(1, 100))+[10000])
        self.assertEqual((result["p50"], result["p95"], result["p99"], result["max"]), (50, 95, 99, 10000))
        self.assertGreater(result["coefficient_of_variation"], 1)

    def test_nonfinite_boolean_and_negative_samples_refuse(self):
        for invalid in ([], [True], [-1], [float("nan")], [float("inf")]):
            with self.assertRaises(ValueError): benchmark.distribution(invalid)

    def test_run_variation_uses_actual_per_run_p95(self):
        one = benchmark.validate_records(records(), 10000, False, 120, False)
        two = copy.deepcopy(one)
        for sample in two: sample["metrics"]["ui_work_ns"] *= 2
        result = benchmark.summarize([one, two])["ui_work_ns"]
        self.assertEqual([v["p95"] for v in result["per_run"]], [798, 1596])
        self.assertEqual(result["run_p95_variation"]["count"], 2)
        # Phase percentiles are separately reported, never substituted for the
        # actual same-ticket total work sample.
        self.assertEqual(result["pooled"]["max"], 1680)


class RecordTests(unittest.TestCase):
    def test_gpu_scene_mode_must_match_every_stage_and_sample(self):
        data = records()
        for record in data: record["gpu_scene_rendering"] = True
        with self.assertRaisesRegex(ValueError, "GPU scene"):
            benchmark.validate_records(data, 10000, False, 120, False)
        self.assertEqual(len(benchmark.validate_records(data, 10000, False, 120, False, gpu_scene_rendering=True)), 120)
        for index in (0, 1, 2, 65, -1):
            for value in (False, "true", 1, None):
                bad = copy.deepcopy(data);bad[index]["gpu_scene_rendering"] = value
                with self.subTest(index=index, value=value), self.assertRaisesRegex(ValueError, "GPU scene"):
                    benchmark.validate_records(bad, 10000, False, 120, False, gpu_scene_rendering=True)
            bad = copy.deepcopy(data);bad[index].pop("gpu_scene_rendering")
            with self.assertRaisesRegex(ValueError, "GPU scene"):
                benchmark.validate_records(bad, 10000, False, 120, False, gpu_scene_rendering=True)
        for record in data: record["gpu_scene_rendering"] = False
        self.assertEqual(len(benchmark.validate_records(data, 10000, False, 120, False)), 120)
        with self.assertRaisesRegex(ValueError, "GPU scene"):
            benchmark.validate_records(records(), 10000, False, 120, False, gpu_scene_rendering=True)

    def test_exact_order_and_geometry_validate(self):
        selected = benchmark.validate_records(records(), 10000, False, 120, False)
        self.assertEqual(len(selected), 120)
        benchmark.compare_geometry(selected, copy.deepcopy(selected))

    def test_profiled_run_is_diagnostic_and_cannot_qualify_warm_budget(self):
        profiled = records()
        for record in (profiled[0], profiled[1], profiled[-1]): record["cpu_profile"] = True
        with self.assertRaisesRegex(ValueError, "diagnostic"):
            benchmark.validate_records(profiled, 10000, False, 120, False)
        self.assertEqual(len(benchmark.validate_records(profiled, 10000, False, 120, False, allow_profile=True)), 120)
        profiled[-1]["cpu_profile"] = False
        with self.assertRaisesRegex(ValueError, "identity"):
            benchmark.validate_records(profiled, 10000, False, 120, False, allow_profile=True)

    def test_pooled_capacity_is_not_rendered_cell_workload(self):
        bad = records();bad[3]["geometry"]["cells"] = 399
        with self.assertRaisesRegex(ValueError, "rendered"): benchmark.validate_records(bad, 10000, False, 120, False)

    def test_stale_duplicate_missing_and_dropped_samples_refuse(self):
        mutations = [lambda r:r.pop(3), lambda r:r[3].update(tag=1), lambda r:r[3].update(generation=0),
                     lambda r:r[3].update(callbacks=2), lambda r:r[-1].update(pending=1), lambda r:r[-1].update(dropped=1),
                     lambda r:r[0].update(release=False), lambda r:r[0].update(rows=1000000),
                     lambda r:r[3]["metrics"].pop("ui_work_ns"),lambda r:r[3]["native_sample"].update(Generation=999),
                     lambda r:r[3]["native_sample"].update(TotalUIWorkNS=1),
                     lambda r:r[3]["metrics"].update(input_ns=True),
                     lambda r:r[3].update(frame=1), lambda r:r[3].update(generation=1)]
        for change in mutations:
            bad = records();change(bad)
            with self.assertRaises(ValueError): benchmark.validate_records(bad, 10000, False, 120, False)

    def test_rowcount_pairs_must_have_same_actual_workload(self):
        one = benchmark.validate_records(records(), 10000, False, 120, False)
        two = copy.deepcopy(one);two[8]["geometry"]["columns"] = 9
        with self.assertRaisesRegex(ValueError, "differs"): benchmark.compare_geometry(one, two)


class FakeProcess:
    def __init__(self, payload):
        reader, writer = os.pipe()
        self.stdout = os.fdopen(reader, "rb", buffering=0)
        self.pid, self.returncode = 123, None
        def emit():
            try:
                for first in range(0, len(payload), 127): os.write(writer, payload[first:first+127])
            finally: os.close(writer)
        self.thread = threading.Thread(target=emit)
        self.thread.start()
    def poll(self): return self.returncode
    def wait(self, timeout=None):
        self.thread.join(timeout)
        self.returncode = 0
        return 0


class ProcessTests(unittest.TestCase):
    def run_fixture(self, data, *, source_details=False, alloc_profile=None, gpu_scene_rendering=False):
        payload = b"native ordinary log\n" + b"\n".join(json.dumps(v).encode() for v in data)+b"\n"
        process = FakeProcess(payload)
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)/"run"
            with mock.patch.object(benchmark.subprocess, "Popen", return_value=process) as launch, mock.patch.object(benchmark, "process_sample", return_value={"rss_bytes": 1234}) as observe, mock.patch.object(benchmark, "terminate_owned_group") as kill:
                try: result = benchmark.run_case(Path("fake"), directory, 10000, False, 120, 128, False, 5, source_details=source_details, alloc_profile=alloc_profile, gpu_scene_rendering=gpu_scene_rendering)
                except Exception as error: result = error
            raw = (directory/"native.log").read_bytes()
            outcome = json.loads((directory/"result.json").read_text())
            self.assertEqual(raw, payload)
            self.assertEqual(launch.call_args.kwargs["bufsize"], 0)
            self.assertTrue(launch.call_args.kwargs["start_new_session"])
            self.assertEqual("-perf-source-details" in launch.call_args.args[0], source_details)
            self.assertEqual(any(v.startswith("-perf-alloc-profile=") for v in launch.call_args.args[0]), alloc_profile is not None)
            self.assertEqual("-gpu-scene" in launch.call_args.args[0], gpu_scene_rendering)
            observe.assert_called_once_with(process.pid)
            return result, outcome

    def test_chunked_pipe_exact_raw_transcript_and_terminal_samples(self):
        result, outcome = self.run_fixture(records())
        self.assertIsInstance(result, tuple)
        self.assertEqual(len(result[0]), 120)
        self.assertEqual(outcome["exit_code"], 0)
        self.assertEqual(len(outcome["process_observations"]), 1)
        self.assertEqual(outcome["result"], "PASS")

    def test_gpu_scene_launch_flag_and_mismatched_child_result(self):
        data = records()
        for record in data: record["gpu_scene_rendering"] = True
        result, outcome = self.run_fixture(data, gpu_scene_rendering=True)
        self.assertIsInstance(result, tuple)
        self.assertEqual(outcome["records"], data)
        self.assertEqual(outcome["result"], "PASS")
        result, outcome = self.run_fixture(records(), gpu_scene_rendering=True)
        self.assertIsInstance(result, ValueError)
        self.assertEqual(outcome["result"], "FAILED")
        self.assertIn("GPU scene", outcome["error"])

    def test_allocation_launch_identity_and_diagnostic_refusal(self):
        data = records()
        for record in data: record["alloc_profile"] = True
        data[1]["warmup"] = 128
        result, outcome = self.run_fixture(data, alloc_profile=Path("fixture"))
        self.assertIsInstance(result, tuple)
        self.assertEqual(outcome["records"], data)
        result, outcome = self.run_fixture(records(), alloc_profile=Path("fixture"))
        self.assertIsInstance(result, ValueError)
        self.assertIn("requested allocation", outcome["error"])
        for options in ({}, {"allow_profile": True}, {"allow_source_details": True}, {"allow_input_details": True}):
            with self.assertRaisesRegex(ValueError, "diagnostic allocation"):
                benchmark.validate_records(data, 10000, False, 120, False, **options)
        for index in (0, 1, 4, -1):
            corrupted = copy.deepcopy(data)
            corrupted[index]["alloc_profile"] = False
            with self.assertRaisesRegex(ValueError, "identity mismatch"):
                benchmark.validate_records(corrupted, 10000, False, 120, False, allow_alloc_profile=True)
        for field in ("cpu_profile", "source_details", "input_details"):
            corrupted = copy.deepcopy(data)
            for record in corrupted: record[field] = True
            with self.assertRaisesRegex(ValueError, "diagnostic mode"):
                benchmark.validate_records(corrupted, 10000, False, 120, False,
                                           allow_alloc_profile=True, allow_input_details=True)
        cases = {"fixture": {"timings_ns": {"ui_work_ns": {"per_run": [{"p95": 1, "p99": 1}]}}}}
        self.assertFalse(benchmark.warm_target_met(cases, alloc_profile=True))

    def test_allocation_invalid_modes_refuse_before_launch(self):
        for options in ({"alloc_profile": ""}, {"alloc_profile": "bad\x00prefix"},
                        {"alloc_profile": "p"*4096}, {"alloc_profile": "alloc", "cpu_profile": "cpu"},
                        {"alloc_profile": "alloc", "source_details": True}):
            with tempfile.TemporaryDirectory() as temp, mock.patch.object(benchmark.subprocess, "Popen") as launch:
                with self.assertRaisesRegex(ValueError, "allocation diagnostic"):
                    benchmark.run_case(Path("fake"), Path(temp)/"run", 10000, False, 120, 128, False, 5, **options)
                launch.assert_not_called()
                self.assertFalse((Path(temp)/"run").exists())

    def test_invalid_gpu_scene_mode_refuses_before_any_launch(self):
        with tempfile.TemporaryDirectory() as temp, mock.patch.object(benchmark.subprocess, "Popen") as launch:
            with self.assertRaisesRegex(ValueError, "GPU scene"):
                benchmark.run_case(Path("fake"), Path(temp)/"run", 10000, False, 120, 128, False, 5, gpu_scene_rendering="true")
            launch.assert_not_called()
            self.assertFalse((Path(temp)/"run").exists())

    def test_setup_and_ready_validate_without_external_process_observers(self):
        result, outcome = self.run_fixture(records())
        self.assertIsInstance(result, tuple)
        self.assertEqual([r["event"] for r in outcome["records"][:2]],
                         ["grid.perf.setup", "grid.perf.ready"])
        self.assertEqual([r["event"] for r in outcome["process_observations"]], ["grid.perf.end"])
        self.assertEqual(outcome["process_observation_scope"]["warm_rss"], "unavailable")
        self.assertEqual(outcome["process_observation_scope"]["baseline_rss"], "unavailable")
        self.assertEqual(outcome["process_observation_scope"]["cpu_delta"], "unavailable")
        self.assertEqual(len(result[0]), 120)

    def test_timeout_retires_group_even_after_direct_child_exit(self):
        reader, writer = os.pipe()
        process = types.SimpleNamespace(stdout=os.fdopen(reader, "rb", buffering=0), pid=123,
                                        returncode=0, poll=lambda: 0, wait=lambda timeout: 0)
        # Simulate a descendant holding the inherited pipe: direct child is
        # terminal, but no EOF is available. Existing owned-group helper has
        # separate real POSIX grandchild cleanup coverage.
        try:
            with tempfile.TemporaryDirectory() as temp:
                directory = Path(temp)/"run"
                with mock.patch.object(benchmark.subprocess, "Popen", return_value=process), mock.patch.object(benchmark, "terminate_owned_group") as kill:
                    with self.assertRaises(benchmark.subprocess.TimeoutExpired):
                        benchmark.run_case(Path("fake"), directory, 10000, False, 120, 128, False, .01)
                kill.assert_called_once_with(process)
                outcome = json.loads((directory/"result.json").read_text())
                self.assertEqual(outcome["result"], "FAILED")
                self.assertEqual(outcome["exit_code"], 0)
                self.assertIn("TimeoutExpired", outcome["error"])
        finally:
            os.close(writer)
            process.stdout.close()

    def test_failure_retains_exact_raw_output_and_error(self):
        bad = records();bad[-1]["callbacks"] = 119
        result, outcome = self.run_fixture(bad)
        self.assertIsInstance(result, ValueError)
        self.assertEqual(outcome["result"], "FAILED")
        self.assertIn("completion mismatch", outcome["error"])
        self.assertEqual(len(outcome["log_sha256"]), 64)

    def test_detailed_flag_and_exact_nested_raw_records_are_preserved(self):
        data = detailed_records()
        result, outcome = self.run_fixture(data, source_details=True)
        self.assertIsInstance(result, tuple)
        self.assertEqual(outcome["records"], data)
        self.assertEqual(outcome["result"], "PASS")

    def test_requested_detail_mode_refuses_ordinary_result_even_when_other_fields_match(self):
        result, outcome = self.run_fixture(records(), source_details=True)
        self.assertIsInstance(result, ValueError)
        self.assertIn("requested source detail", outcome["error"])

    def test_requested_detailed_warmup_must_match_ready_metadata(self):
        result, outcome = self.run_fixture(detailed_records(warmup=64), source_details=True)
        self.assertIsInstance(result, ValueError)
        self.assertIn("warmup count mismatch", outcome["error"])

    def test_corrupt_detailed_result_retains_raw_failure_without_warm_fallback(self):
        data = detailed_records()
        data[2]["native_sample"]["SourceDetails"]["Records"][0]["LogicalUploadRGBABytes"] = 0
        result, outcome = self.run_fixture(data, source_details=True)
        self.assertIsInstance(result, ValueError)
        self.assertEqual(outcome["result"], "FAILED")
        self.assertEqual(outcome["records"], data)

    def test_detail_capacity_refusal_happens_before_any_launch(self):
        with tempfile.TemporaryDirectory() as temp, mock.patch.object(benchmark.subprocess, "Popen") as launch:
            with self.assertRaisesRegex(ValueError, "capacity"):
                benchmark.run_case(Path("fake"), Path(temp)/"run", 10000, False, 2048, 2048, False, 5, source_details=True)
            launch.assert_not_called()
            self.assertFalse((Path(temp)/"run").exists())


class SetupFailureTests(unittest.TestCase):
    def test_launch_failure_records_terminal_failure(self):
        with tempfile.TemporaryDirectory() as temp:
            directory=Path(temp)/'run'
            with mock.patch.object(benchmark.subprocess,'Popen',side_effect=OSError('fixture launch unavailable')):
                with self.assertRaises(OSError):
                    benchmark.run_case(Path('fake'),directory,10000,False,120,128,False,5)
            result=directory/'result.json'
            self.assertTrue(result.is_file(),'launch failure lost per-run result metadata')
            record=json.loads(result.read_text())
            self.assertEqual(record['result'],'FAILED')
            self.assertIsNone(record['exit_code'])
            self.assertIn('fixture launch unavailable',record['error'])

    def test_selector_setup_failure_retires_child_and_records_failure(self):
        process=mock.Mock()
        process.stdout=io.BytesIO()
        process.pid=123
        process.returncode=-9
        process.poll.return_value=None
        process.wait.return_value=-9
        selector=mock.Mock()
        selector.register.side_effect=OSError('fixture selector unavailable')
        try:
            with tempfile.TemporaryDirectory() as temp:
                directory=Path(temp)/'run'
                with mock.patch.object(benchmark.subprocess,'Popen',return_value=process), mock.patch.object(benchmark.selectors,'DefaultSelector',return_value=selector), mock.patch.object(benchmark,'terminate_owned_group') as retire:
                    with self.assertRaises(OSError):
                        benchmark.run_case(Path('fake'),directory,10000,False,120,128,False,5)
                    retire.assert_called_once_with(process)
                selector.close.assert_called_once_with()
                result=json.loads((directory/'result.json').read_text())
                self.assertEqual(result['result'],'FAILED')
                self.assertIn('fixture selector unavailable',result['error'])
        finally:
            process.stdout.close()


class DiagnosticReportTests(unittest.TestCase):
    def test_cli_gpu_scene_mode_is_preserved_and_forwarded(self):
        data = records()
        for record in data: record["gpu_scene_rendering"] = True
        selected = benchmark.validate_records(data, 10000, False, 120, False, gpu_scene_rendering=True)
        for sample in selected:
            sample.update(go_process_alloc_bytes=1, go_process_mallocs=1, source_callbacks={"Keys": 1, "Inverse": 1, "Cells": 1})
        with tempfile.TemporaryDirectory() as temp:
            binary = Path(temp)/"binary";binary.write_bytes(b"CPU-only fixture; never launched")
            output = Path(temp)/"results"
            argv = ["benchmark-grid.py", "--run", "--binary", str(binary), "--output", str(output), "--gpu-scene"]
            with mock.patch.object(benchmark.sys, "argv", argv), mock.patch.object(benchmark.platform, "system", return_value="Darwin"), \
                 mock.patch.object(benchmark, "metadata_command", return_value={}), mock.patch.object(benchmark, "selected_inputs", return_value={}), \
                 mock.patch.object(benchmark, "run_case", return_value=(selected, {"result": "PASS"})) as run, \
                 mock.patch("sys.stdout", new_callable=io.StringIO):
                benchmark.main()
            self.assertEqual(run.call_count, 20)
            self.assertTrue(all(call.kwargs["gpu_scene_rendering"] for call in run.call_args_list))
            summary = json.loads((output/"summary.json").read_text())
            self.assertTrue(summary["gpu_scene_rendering"])
            self.assertFalse(summary["diagnostic"])
            self.assertEqual(summary["warm_targets_ns"], {"p95": 4_000_000, "p99": 6_000_000})

    def test_cli_detailed_campaign_is_labeled_diagnostic_even_below_warm_targets(self):
        selected = benchmark.validate_records(detailed_records(), 10000, False, 120, False, allow_source_details=True)
        for sample in selected:
            sample.update(go_process_alloc_bytes=1, go_process_mallocs=1, source_callbacks={"Keys": 1, "Inverse": 1, "Cells": 1})
        with tempfile.TemporaryDirectory() as temp:
            binary = Path(temp)/"binary"; binary.write_bytes(b"CPU-only test fixture, never launched")
            output = Path(temp)/"results"
            argv = ["benchmark-grid.py", "--run", "--binary", str(binary), "--output", str(output), "--source-details"]
            with mock.patch.object(benchmark.sys, "argv", argv), mock.patch.object(benchmark.platform, "system", return_value="Darwin"), \
                 mock.patch.object(benchmark, "metadata_command", return_value={}), mock.patch.object(benchmark, "selected_inputs", return_value={}), \
                 mock.patch.object(benchmark, "run_case", return_value=(selected, {"result": "PASS"})) as run, \
                 mock.patch("sys.stdout", new_callable=io.StringIO):
                benchmark.main()
            self.assertEqual(run.call_count, 20)
            self.assertTrue(all(call.kwargs["source_details"] for call in run.call_args_list))
            summary = json.loads((output/"summary.json").read_text())
            self.assertEqual(summary["result"], "DIAGNOSTIC")
            self.assertTrue(summary["diagnostic"])
            self.assertTrue(summary["source_details"])
            self.assertFalse(summary["warm_target_met"])
            self.assertEqual(summary["warm_targets_ns"], {"p95": 4_000_000, "p99": 6_000_000})

    def test_cli_detailed_capacity_refusal_precedes_output_or_process(self):
        with tempfile.TemporaryDirectory() as temp:
            output = Path(temp)/"results"
            argv = ["benchmark-grid.py", "--run", "--binary", str(Path(temp)/"missing"), "--output", str(output),
                    "--source-details", "--warmup", "2048", "--samples", "2048"]
            with mock.patch.object(benchmark.sys, "argv", argv), mock.patch.object(benchmark.platform, "system", return_value="Darwin"), \
                 mock.patch.object(benchmark, "run_case") as run, mock.patch("sys.stderr", new_callable=io.StringIO):
                with self.assertRaises(SystemExit): benchmark.main()
            run.assert_not_called()
            self.assertFalse(output.exists())



class WarmupAndInputDetailTests(unittest.TestCase):
    def guarded(self):
        data = records()
        for record in data: record["warmup_guard"] = 1
        data[1].update(warmup=128, warm_callbacks=128,
                       warm_start_viewport={"Width":768,"Height":624,"ScrollX":0,"ScrollY":0},
                       warm_end_viewport={"Width":768,"Height":624,"ScrollX":0,"ScrollY":0})
        return data

    def test_closed_warmup_guard_rejects_offset_callback_and_version_drift(self):
        self.assertEqual(len(benchmark.validate_records(self.guarded(),10000,False,120,False)),120)
        for change in (lambda d:d[1]["warm_end_viewport"].update(ScrollY=31),
                       lambda d:d[1].update(warm_callbacks=129),
                       lambda d:d[1]["warm_start_viewport"].update(ScrollX=True),
                       lambda d:d[0].update(warmup_guard=True),
                       lambda d:d[-1].update(warmup_guard=2),
                       lambda d:d[2].update(warmup_guard=True),
                       lambda d:d[3].update(input_details=True)):
            data=self.guarded();change(data)
            with self.assertRaises(ValueError):benchmark.validate_records(data,10000,False,120,False)

    def test_input_details_never_silently_qualify_as_ordinary(self):
        data=self.guarded()
        for record in data:record["input_details"]=True
        for options in ({},{"allow_profile":True},{"allow_source_details":True}):
            with self.assertRaisesRegex(ValueError,"diagnostic input"):
                benchmark.validate_records(data,10000,False,120,False,**options)
        self.assertEqual(len(benchmark.validate_records(data,10000,False,120,False,allow_input_details=True)),120)
        data[1]["input_details"]=False
        with self.assertRaises(ValueError):benchmark.validate_records(data,10000,False,120,False,allow_input_details=True)

if __name__ == "__main__": unittest.main()
