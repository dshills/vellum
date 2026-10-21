import unittest, importlib.util, sys
from check import validate, policy, upload, CODE
ours = sys.modules['check']; sys.modules['check'] = upload
s = importlib.util.spec_from_file_location('cache_upload_fixtures', CODE.parent/'upload/check_test.py')
f = importlib.util.module_from_spec(s); s.loader.exec_module(f); sys.modules['check'] = ours

class CacheControl(unittest.TestCase):
    def data(self, mode='writecombined'):
        d, m = f.Upload().data()
        count = d['native_upload_trace']['count']+2
        d['cache_mode'] = mode
        d['cache_policy'] = [1,count,count,0,0,0] if mode == 'writecombined' else [2,count,0,count,0,0]
        return d, m

    def test_both_modes_and_parent_upload_accounting(self):
        for mode in ('writecombined', 'default'):
            result = validate(*self.data(mode), mode)
            self.assertEqual(result['joined_cohort_uploads'], 248)
            self.assertEqual(result['cache_policy'][1], 250)

    def test_unconfigured_wrong_mode_mixed_storage_and_counters(self):
        for row in ([0,250,250,0,0,0], [2,250,250,0,0,0], [1,250,249,1,0,0],
                    [1,250,250,0,1,0], [1,250,250,0,0,1], [1,250,249,0,0,0],
                    [1,1,1,0,0,0], [1,250,250,0,0], [1,True,1,0,0,0], [1,20000,20000,0,0,0]):
            with self.subTest(row=row):
                doc, _ = self.data(); doc['cache_policy'] = row
                with self.assertRaises(ValueError): policy(doc, 'writecombined')

    def test_native_failure_still_refuses(self):
        doc, match = self.data('default')
        doc['native_upload_trace']['records'][-1]['stages'] = 27
        with self.assertRaises(ValueError): validate(doc, match, 'default')

if __name__ == '__main__': unittest.main()
