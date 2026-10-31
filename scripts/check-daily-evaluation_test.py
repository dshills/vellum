import importlib.util
from datetime import date
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('evaluation', Path(__file__).with_name('check-daily-evaluation.py'))
evaluation = importlib.util.module_from_spec(spec)
spec.loader.exec_module(evaluation)


class EvaluationTests(unittest.TestCase):
    def session(self, day, kind='operator', checks=None, blockers=None):
        return dict(date=day, kind=kind, checks=list(evaluation.CHECKS) if checks is None else checks,
                    open_blockers=[] if blockers is None else blockers)

    def report(self, sessions):
        return evaluation.evaluate(dict(candidate='test', sessions=sessions), 'test', date(2026, 10, 31))

    def test_repeated_same_day_or_automated_runs_cannot_qualify(self):
        self.assertEqual(self.report([self.session('2026-10-31') for _ in range(3)])['status'], 'pending')
        self.assertEqual(self.report([self.session(d, 'automated') for d in ['2026-10-29','2026-10-30','2026-10-31']])['qualifying_operator_days'], [])

    def test_three_full_operator_days_and_no_blockers_are_required(self):
        sessions = [self.session(d) for d in ['2026-10-29','2026-10-30','2026-10-31']]
        report = self.report(sessions)
        self.assertEqual(report['status'], 'recorded')
        self.assertEqual(report['qualifying_operator_days'], ['2026-10-29','2026-10-30','2026-10-31'])
        sessions[-1]['checks'] = ['edit']
        self.assertEqual(self.report(sessions)['status'], 'pending')
        sessions[-1]['open_blockers'] = ['DATA-LOSS-1']
        self.assertEqual(self.report(sessions)['status'], 'blocked')

    def test_future_date_and_candidate_drift_refuse_qualification(self):
        with self.assertRaises(ValueError):
            self.report([self.session('2026-11-01')])
        with self.assertRaises(ValueError):
            evaluation.evaluate(dict(candidate='old', sessions=[]), 'new', date(2026,10,31))


if __name__ == '__main__':
    unittest.main()
