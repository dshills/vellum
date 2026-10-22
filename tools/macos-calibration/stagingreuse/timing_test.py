import importlib.util,sys,unittest
import timing_check as timing

previous=sys.modules.get('check');sys.modules['check']=timing.upload
s=importlib.util.spec_from_file_location('timing_upload_fixtures',timing.CODE.parent/'upload/check_test.py');fixtures=importlib.util.module_from_spec(s);s.loader.exec_module(fixtures)
if previous is None:sys.modules.pop('check',None)
else:sys.modules['check']=previous

def data(mode):
    d,match=fixtures.Upload().data();d['reuse_mode']=mode;events=[]
    def event(k,identity=0,size=0,submit=0,fence=0,call=0,cached_count=0,cached_bytes=0,clock=1):
        events.append(dict(sequence=len(events),complete=True,kind=k,owner=1,id=identity,bytes=size,submit=submit,fence=fence,call=call,cached_count=cached_count,cached_bytes=cached_bytes,clock=clock))
    event(1,clock=d['native_upload_trace']['records'][0]['begin']-.001)
    for i,r in enumerate(d['native_upload_trace']['records']):
        r['reuse_eligible']=True;size=r['width']*r['height']*4;identity=1 if mode=='on' else i+1
        event(3 if mode=='on' and i else 2,identity,size,call=i+1,clock=r['begin']+1e-9)
        event(4,identity,size,submit=i+1,clock=r['end']+1e-6)
        event(5,identity,size,submit=i+1,fence=i+1,clock=r['end']+2e-6)
        if mode=='on':event(6,identity,size,submit=i+1,fence=i+1,cached_count=1,cached_bytes=size,clock=r['end']+3e-6)
        else:event(7,identity,size,submit=i+1,clock=r['end']+3e-6)
    end=d['native_upload_trace']['records'][-1]['end']+.01;event(8,clock=end)
    if mode=='on':event(7,1,size,clock=end+.001)
    event(9,clock=end+.002)
    d['reuse_trace']=dict(mode=2 if mode=='on' else 1,count=len(events),pending=0,unstable=False,overflow=False,invalid=False,active_devices=0,records=events)
    return d,match

class Timing(unittest.TestCase):
    def test_both_modes_reuse_joins_and_parent_presentation_accounting(self):
        for mode,want in [('off',0),('on',120)]:
            result=timing.validate(*data(mode),mode);self.assertEqual(result['measured_reuse_hits'],want);self.assertFalse(result['ordinary_latency_qualified'])
    def test_missing_eligibility_wrong_call_and_active_device_refuse(self):
        for mutate in [lambda d:d['native_upload_trace']['records'][-1].update(reuse_eligible=False),lambda d:d['reuse_trace']['records'][1].update(call=999),lambda d:d['reuse_trace'].update(active_devices=1),lambda d:d.update(reuse_mode='off')]:
            d,m=data('on');mutate(d)
            with self.assertRaises(ValueError):timing.validate(d,m,'on')

if __name__=='__main__':unittest.main()
