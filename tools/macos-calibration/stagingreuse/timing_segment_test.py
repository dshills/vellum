import copy,unittest
import timing_segment_check as check
import timing_test

def data(mode='on'):
    d,m=timing_test.data(mode)
    d['request_schedule_phases']=[dict(Sequence=i,WakeLatenessNS=n-20,ForegroundCheckNS=15,PostCheckLatenessNS=n) for i,n in enumerate(d['lateness_ns'],1)]
    return d,m

class Segments(unittest.TestCase):
    def test_full_parent_and_signed_segments(self):
        for mode in ['off','on']:
            d,m=data(mode);r=check.validate(d,m,mode)
            self.assertEqual(r['request_schedule_segments']['requests'],248)
        d,m=data();d['request_schedule_phases'][0].update(WakeLatenessNS=-1)
        check.validate(d,m,'on')
    def test_refuses_bad_count_sequence_type_join_and_order(self):
        for mutate in [lambda d:d['request_schedule_phases'].pop(),lambda d:d['request_schedule_phases'][0].update(Sequence=2),lambda d:d['request_schedule_phases'][0].update(ForegroundCheckNS=True),lambda d:d['request_schedule_phases'][0].update(PostCheckLatenessNS=99),lambda d:d['request_schedule_phases'][0].update(ForegroundCheckNS=-1),lambda d:d['request_schedule_phases'][0].update(WakeLatenessNS=10**9)]:
            d,m=data();mutate(d)
            with self.assertRaises(ValueError):check.validate(d,m,'on')
    def test_partial_diagnostic_does_not_accept_failed_schedule(self):
        d,m=data();d['request_schedule_phases']=d['request_schedule_phases'][:5];d['lateness_ns']=d['lateness_ns'][:5]
        d['lateness_ns'][-1]=42_563_959;d['request_schedule_phases'][-1].update(WakeLatenessNS=40_000_000,ForegroundCheckNS=2_563_954,PostCheckLatenessNS=42_563_959)
        self.assertEqual(check.phases(d,False)['requests'],5)
        with self.assertRaises(ValueError):check.validate(d,m,'on')
if __name__=='__main__':unittest.main()
