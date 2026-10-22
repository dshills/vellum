import importlib.util,sys,unittest
from pathlib import Path
import byte_check as check
previous=sys.modules.get('check');sys.modules['check']=check.parent
s=importlib.util.spec_from_file_location('cached_byte_fixtures',Path(__file__).resolve().parent.parent/'stagingreuse/check_test.py');fixtures=importlib.util.module_from_spec(s);s.loader.exec_module(fixtures)
if previous is None:sys.modules.pop('check',None)
else:sys.modules['check']=previous

def data():
    d=fixtures.byte_fixture();d['native_upload_trace']['cache_policy']=[2,22,0,22,0,0];return d
class Cached(unittest.TestCase):
    def test_complete_parent_byte_matrix_and_exact_cache(self):
        r=check.validate(data(),'off');self.assertEqual(r['cache_policy'],[2,22,0,22,0,0]);self.assertFalse(r['release_qualified'])
    def test_refuse_scope_count_mode_type_and_parent_failure(self):
        for row in [[2,23,0,23,0,0],[1,22,22,0,0,0],[2,22,1,21,0,0],[2,22,0,22,1,0],[2,22,0,22,0,1],[2,22,0,22,0],[True,22,0,22,0,0]]:
            d=data();d['native_upload_trace']['cache_policy']=row
            with self.assertRaises(ValueError):check.validate(d,'off')
        d=data();d['writes'][1]['full_byte_match']=False
        with self.assertRaises(ValueError):check.validate(d,'off')
if __name__=='__main__':unittest.main()
