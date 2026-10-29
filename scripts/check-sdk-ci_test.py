"""CI admission/result checks; real relocation is exercised separately."""
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location('sdk_ci', Path(__file__).with_name('check-sdk-ci.py'))
ci = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ci)


class Checks(unittest.TestCase):
    def test_inside_checkout_refused_before_execution(self):
        with mock.patch.object(ci, 'check_child_output') as child:
            with self.assertRaisesRegex(ValueError, 'outside'):
                ci.check(ci.ROOT / 'build/ci')
            child.assert_not_called()

    def test_dirty_checkout_does_not_reserve_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            output = Path(tmp) / 'result'
            with mock.patch.object(ci, 'check_child_output', side_effect=['a' * 40, ' M source.go']):
                with self.assertRaisesRegex(ValueError, 'clean committed'):
                    ci.check(output)
            self.assertFalse(output.exists())

    def exercise(self, output, mutate=lambda result: None, fail=False):
        calls = []
        digest = hashlib.sha256(b'own archive bytes').hexdigest()
        result = dict(event='vellum.sdk.relocated', archive_sha256=digest,
                      source=dict(commit='a' * 40, clean=True), path_contains_spaces=True,
                      native_runtime='not executed', document_replacements_within_sdk={'module': '.'})
        mutate(result)
        def run(argv, **kwargs):
            calls.append(argv)
            self.assertTrue(kwargs['check'])
            self.assertGreater(kwargs['timeout'], 0)
            self.assertEqual(kwargs['log_limit_bytes'], 32 << 20)
            if len(calls) == 1:
                (output / 'Vellum SDK.tar.gz').write_bytes(b'own archive bytes')
            elif fail:
                raise OSError('verifier failed')
            else:
                kwargs['stdout'].write(json.dumps(result).encode())
        with mock.patch.object(ci, 'check_child_output', side_effect=['a' * 40, '']), mock.patch.object(ci, 'run_child', side_effect=run):
            ci.check(output)
        self.assertNotIn('--public', calls[0])
        self.assertEqual(calls[1][-2:], ['--sha256', digest])
        self.assertNotIn('--runtime', calls[1])
        self.assertEqual(json.loads((output / 'relocation.json').read_text()), result)

    def test_success_binds_private_export_digest_without_runtime(self):
        with tempfile.TemporaryDirectory() as tmp:
            self.exercise(Path(tmp) / 'output')

    def test_failure_cannot_leave_success_report(self):
        for fail in (False, True):
            with self.subTest(child_failure=fail), tempfile.TemporaryDirectory() as tmp:
                output = Path(tmp) / 'output'
                with self.assertRaises((ValueError, OSError)):
                    self.exercise(output, lambda r: r.update(archive_sha256='0' * 64), fail)
                self.assertFalse((output / 'relocation.json').exists())

    def test_mismatched_source_live_runtime_and_missing_document_evidence_refused(self):
        for field, value in [('source', {'commit': 'b' * 40, 'clean': True}),
                             ('path_contains_spaces', False), ('native_runtime', {}),
                             ('document_replacements_within_sdk', {})]:
            with self.subTest(field=field), tempfile.TemporaryDirectory() as tmp:
                with self.assertRaisesRegex(ValueError, 'does not match'):
                    self.exercise(Path(tmp) / 'output', lambda r: r.update({field: value}))


if __name__ == '__main__':
    unittest.main()
