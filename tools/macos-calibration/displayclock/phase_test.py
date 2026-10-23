"""Focused regressions for clock identity, boundary semantics and attribution."""
from pathlib import Path
import copy
import hashlib
import io
import json
import tarfile
import tempfile
import unittest
import phase_analysis as phase


def ticks():
    return [{'ordinal': i, 'timestamp': stamp, 'observed': stamp+.05}
            for i, stamp in enumerate([.1, .2, .3, .4])]


def row():
    return {'raw_host_clocks': {'receipt': 1., 'acquired': 1.004, 'shown': 1.04},
            'receipt_to_shown_ns': 40_000_000,
            'backend_commands': [
                {'kind': 0, 'commit_start': 1.005, 'gpu_start': 1.006, 'gpu_end': 1.007},
                {'kind': 2, 'commit_start': 1.0055, 'present_request': 1.0054,
                 'gpu_end': 1.008}],
            'receipt_lateness_go_ns': 2_000_000,
            'receipt_clock_bracket_go_ns': 1000}


class PhaseTests(unittest.TestCase):
    def test_reported_clock_is_not_delivered_callback(self):
        result = phase.ClockIndex(ticks()).join(.22)
        self.assertEqual(result['preceding_reported_ordinal'], 1)
        self.assertEqual(result['preceding_delivered_ordinal'], 0)
        self.assertFalse(result['preceding_reported_already_delivered'])
        self.assertAlmostEqual(result['preceding_reported_delivery_minus_event_ms'], 30)
        self.assertAlmostEqual(result['next_delivered_minus_event_ms'], 30)
        self.assertAlmostEqual(result['event_minus_preceding_reported_ms'], 20)
        near = phase.ClockIndex(ticks()).join(.29)
        self.assertEqual(near['nearest_reported_ordinal'], 2)
        self.assertAlmostEqual(near['event_minus_nearest_reported_ms'], -10)
        self.assertAlmostEqual(near['nearest_reported_delivery_minus_event_ms'], 60)

    def test_equality_and_duplicate_delivery_use_last_preceding(self):
        clock = phase.ClockIndex(ticks())
        self.assertEqual(clock.join(.2)['preceding_reported_ordinal'], 1)
        at_delivery = clock.join(.25)
        self.assertEqual(at_delivery['preceding_delivered_ordinal'], 1)
        self.assertTrue(at_delivery['preceding_reported_already_delivered'])
        self.assertEqual(at_delivery['event_minus_preceding_delivered_ms'], 0)
        tied = ticks(); tied[2]['observed'] = tied[1]['observed']
        self.assertEqual(phase.ClockIndex(tied).join(.25)['preceding_delivered_ordinal'], 2)

    def test_missing_coverage_nonfinite_and_order_are_rejected(self):
        for event in [.12, .42, float('nan'), float('inf'), True]:
            with self.subTest(event=event), self.assertRaises(ValueError):
                phase.ClockIndex(ticks()).join(event)
        for mutation in ['backwards_report', 'backwards_delivery', 'nonfinite']:
            values = ticks()
            if mutation == 'backwards_report': values[1]['timestamp'] = .09
            elif mutation == 'backwards_delivery': values[1]['observed'] = .14
            else: values[1]['observed'] = float('nan')
            with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                phase.ClockIndex(values)

    def test_same_input_partition_closes_and_overlap_stays_separate(self):
        clocks, parts, split = phase.timeline(row())
        expected = [4, 1, 1, 1, 33]
        for value, want in zip(parts.values(), expected): self.assertAlmostEqual(value, want)
        self.assertAlmostEqual(sum(parts.values()), 40)
        self.assertAlmostEqual(split['receipt_to_present_gpu_end'], 8)
        self.assertAlmostEqual(split['present_gpu_end_to_shown'], 32)
        # Present request is earlier than GPU end; never add that overlap to parts.
        self.assertLess(clocks['present_request'], clocks['render_gpu_end'])
        for mutation in ['reverse', 'total', 'present_gpu']:
            value = row()
            if mutation == 'reverse': value['backend_commands'][0]['commit_start'] = 1.003
            elif mutation == 'total': value['receipt_to_shown_ns'] += 1_000_000
            else: value['backend_commands'][-1]['gpu_end'] = 1.05
            with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                phase.timeline(value)

    def test_actual_rank_ticket_not_sum_of_marginal_quantiles(self):
        rows = []
        for i in range(248):
            value = row(); latency = (10+(i % 120))/1000
            value['raw_host_clocks']['shown'] = 1+latency
            value['receipt_to_shown_ns'] = latency*1e9
            value.update(accepted_revision=i+1, scope=1, opaque_tag=i+10,
                         generation=i+2, drawable_id=i+20,
                         phase='warmup' if i < 128 else 'measured')
            rows.append(value)
        expected = sorted(rows[128:], key=lambda r: r['receipt_to_shown_ns'])[113]
        checked = {'parent_validation': {'joined_rows': rows,
                   'p95_rank_ticket': expected,
                   'receipt_to_shown_ms': phase.stats([r['receipt_to_shown_ns']/1e6
                                                      for r in rows[128:]])}}
        summary, derived = phase.analyze_case(checked, {}, 'off')
        self.assertEqual(summary['p95_rank_ticket']['opaque_tag'], expected['opaque_tag'])
        self.assertIsNone(summary['display_clock_phase'])
        self.assertTrue(all('clock_phases' not in r for r in derived))
        self.assertAlmostEqual(sum(summary['p95_rank_ticket']['partition_ms'].values()),
                               summary['receipt_to_shown_ms']['p95'])
        invalid = copy.deepcopy(checked)
        invalid['parent_validation']['p95_rank_ticket'] = rows[128]
        with self.assertRaisesRegex(ValueError, 'p95 input identity'):
            phase.analyze_case(invalid, {}, 'off')

    def test_classifier_source_must_match_original_run_pins(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name); pins = {}
            for relative in phase.VALIDATORS:
                path = root/relative; path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(b'original')
                pins['/original/checkout/'+relative] = phase.sha(b'original')
            proof = phase.classifier_identity({'pins': pins}, root)
            self.assertEqual(len(proof), 8)
            (root/phase.VALIDATORS[0]).write_bytes(b'changed')
            with self.assertRaisesRegex(ValueError, 'unchanged classifier'):
                phase.classifier_identity({'pins': pins}, root)

    def test_pinned_archive_and_member_integrity(self):
        with tempfile.TemporaryDirectory() as name:
            directory = Path(name); path = directory/'result-evidence.tar.xz'
            payload = b'{"value":1}'
            with tarfile.open(path, 'w:xz') as tar:
                item = tarfile.TarInfo('journal'); item.size = len(payload)
                tar.addfile(item, io.BytesIO(payload))
            digest = hashlib.sha256(path.read_bytes()).hexdigest()
            index = {'archive_sha256': digest, 'files': {
                'journal': {'bytes': len(payload), 'sha256': phase.sha(payload)}}}
            index_path = directory/'result-evidence-index.json'
            index_path.write_text(json.dumps(index))
            data, _ = phase.read_archive(directory, 'result', digest, ['journal'])
            self.assertEqual(data['journal'], payload)
            with self.assertRaisesRegex(ValueError, 'pinned archive identity'):
                phase.read_archive(directory, 'result', '0'*64, ['journal'])
            index['files']['journal']['sha256'] = '0'*64
            index_path.write_text(json.dumps(index))
            with self.assertRaisesRegex(ValueError, 'indexed member identity'):
                phase.read_archive(directory, 'result', digest, ['journal'])


if __name__ == '__main__':
    unittest.main()
