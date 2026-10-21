import copy
import unittest
import check

def fixture(mode='writecombined'):
    rows = []
    for i in range(248):
        rows.append({'sequence': i+1, 'phase': 1 if i < 128 else 2,
                     'storage_mode': 0, 'cpu_cache_mode': 1 if mode == 'writecombined' else 0,
                     'begin': 1+i, 'end': 1.5+i,
                     **{k: 1 for k in ('create_ns', 'map_ns', 'first_copy_ns', 'subsequent_copy_ns')},
                     **{k: {'completed': True, 'verified_bytes': 7692216, 'submit_wait_ns': 1, 'verify_ns': 1}
                        for k in ('first_gpu', 'subsequent_gpu')}})
    return {'kind': 'offscreen-metal-copy-control', 'release_qualified': False, 'runtime_returned': True,
            'mode': mode, 'comparator_negative_control': True, 'unified_memory': True, 'device': 'test',
            'width': 1519, 'height': 1266, 'pitch': 6076, 'bytes': 7692216, 'pixel_format': 71,
            'warmup': 128, 'measured': 120, 'begin': 1, 'end': 249, 'records': rows}

class Controls(unittest.TestCase):
    def test_both_modes_and_measured_population(self):
        for mode in ('writecombined', 'default'):
            doc = fixture(mode)
            doc['records'][0]['first_copy_ns'] = 1000000
            result = check.validate(doc, mode)
            self.assertEqual(result['CPU_ms']['first_copy_ns']['max'], 0.000001)
            self.assertEqual(result['GPU_verified_transfers'], 496)
            self.assertEqual(result['paired_subsequent_minus_first_ms']['p95'], 0)

    def test_bad_terminal_or_geometry(self):
        for key, value in [('release_qualified', True), ('runtime_returned', False), ('error', 'failed'),
                           ('comparator_negative_control', False), ('unified_memory', False), ('width', 1520),
                           ('pitch', 6144), ('pixel_format', 81), ('warmup', 0), ('end', float('nan'))]:
            with self.subTest(key=key):
                doc = fixture(); doc[key] = value
                with self.assertRaises(ValueError): check.validate(doc, 'writecombined')

    def test_no_missing_duplicate_reordered_or_relabelled_rows(self):
        for mutation in (lambda r: r.pop(), lambda r: r.__setitem__(1, copy.deepcopy(r[0])),
                         lambda r: r.reverse(), lambda r: r[128].update(phase=1)):
            doc = fixture(); mutation(doc['records'])
            with self.assertRaises(ValueError): check.validate(doc, 'writecombined')

    def test_gpu_failure_partial_verification_and_mode_drift(self):
        for key, value in [('completed', False), ('verified_bytes', 1), ('verified_bytes', True),
                           ('submit_wait_ns', -1), ('verify_ns', float('inf'))]:
            for which in ('first_gpu', 'subsequent_gpu'):
                with self.subTest(key=key, which=which):
                    doc = fixture(); doc['records'][150][which][key] = value
                    with self.assertRaises(ValueError): check.validate(doc, 'writecombined')
        for key, value in [('cpu_cache_mode', 0), ('storage_mode', 1), ('cpu_cache_mode', True)]:
            doc = fixture(); doc['records'][150][key] = value
            with self.assertRaises(ValueError): check.validate(doc, 'writecombined')

    def test_clock_faults_overlap_and_impossible_intervals(self):
        for key, value in [('begin', 0), ('end', 1), ('first_copy_ns', -1),
                           ('subsequent_copy_ns', float('nan')), ('map_ns', 1e12)]:
            doc = fixture(); doc['records'][150][key] = value
            with self.assertRaises(ValueError): check.validate(doc, 'writecombined')

if __name__ == '__main__': unittest.main()
