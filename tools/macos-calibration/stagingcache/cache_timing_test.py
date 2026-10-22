from pathlib import Path
import importlib.util,sys,unittest
import cache_timing_check as check
sys.path.insert(0,str(Path(__file__).resolve().parent.parent/'stagingreuse'))
s=importlib.util.spec_from_file_location('texture_cache_segment_fixtures',Path(__file__).resolve().parent.parent/'stagingreuse/timing_segment_test.py');fixtures=importlib.util.module_from_spec(s);s.loader.exec_module(fixtures)

def data(mode):
    d,m=fixtures.data('on');d['cache_mode']=mode;d['native_upload_trace']['cache_policy']=[1,1,1,0,0,0] if mode=='writecombined' else [2,1,0,1,0,0];return d,m
class CacheTiming(unittest.TestCase):
    def test_both_cache_modes_parent_and_measured_reuse(self):
        for mode in ['writecombined','default']:
            r=check.validate(*data(mode),mode);self.assertEqual(r['measured_reuse_hits'],120);self.assertFalse(r['ordinary_latency_qualified'])
    def test_refuses_extra_scope_wrong_cache_mode_type_and_parent_failures(self):
        for mutate in [lambda d:d['native_upload_trace'].update(cache_policy=[2,2,0,2,0,0]),lambda d:d['native_upload_trace'].update(cache_policy=[2,1,1,0,0,0]),lambda d:d['native_upload_trace'].update(cache_policy=[True,1,0,1,0,0]),lambda d:d['native_upload_trace'].update(cache_policy=[2,1,0,1,1,0]),lambda d:d.update(cache_mode='writecombined'),lambda d:d.update(reuse_mode='off'),lambda d:d['request_schedule_phases'].pop()]:
            d,m=data('default');mutate(d)
            with self.assertRaises(ValueError):check.validate(d,m,'default')
if __name__=='__main__':unittest.main()
