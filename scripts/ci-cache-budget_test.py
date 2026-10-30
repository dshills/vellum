import importlib.util
from pathlib import Path
import re
import subprocess
import sys
import unittest

spec = importlib.util.spec_from_file_location('budget', Path(__file__).with_name('ci-cache-budget.py'))
budget = importlib.util.module_from_spec(spec)
spec.loader.exec_module(budget)


class Budget(unittest.TestCase):
    def test_cold_job_near_deadline_skips_optional_work(self):
        for minutes in (35, 60):
            with self.subTest(minutes=minutes):
                self.assertTrue(budget.may_save(1000, minutes, 1000))
                self.assertTrue(budget.may_save(1000, minutes, 1000 + minutes * 60 - 301))
                self.assertFalse(budget.may_save(1000, minutes, 1000 + minutes * 60 - 300))
                self.assertFalse(budget.may_save(1000, minutes, 1000 + minutes * 60))
                self.assertFalse(budget.may_save(1000, minutes, 999))

    def test_cache_upload_consumes_exactly_the_restored_paths(self):
        workflow = (Path(__file__).resolve().parents[1] / '.github/workflows/framework.yml').read_text()
        jobs = re.split(r'^  [a-z-]+:\n', workflow, flags=re.M)[1:]
        self.assertEqual(len(jobs), 2)
        for job in jobs:
            blocks = re.findall(r'uses: actions/cache/(restore|save)@v4\n        with:\n          path: \|\n((?:            .+\n)+)', job)
            self.assertEqual([b[0] for b in blocks], ['restore', 'save'])
            self.assertEqual(blocks[0][1], blocks[1][1])
            self.assertIn('key: ${{ steps.go-cache.outputs.cache-primary-key }}', job)
            self.assertIn('continue-on-error: true\n        timeout-minutes: 2', job)
        self.assertNotIn('uses: actions/cache@', workflow)

    def test_bad_clock_input_fails_closed_without_failing_checks(self):
        for args in ([], ['bad', '60'], ['0', '60'], ['1000', '5'], ['1', '60', 'extra']):
            result = subprocess.run([sys.executable, '-B', str(Path(budget.__file__)), *args],
                                    text=True, capture_output=True, timeout=10)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout, 'save=false\n')


if __name__ == '__main__':
    unittest.main()
