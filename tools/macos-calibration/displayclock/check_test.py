import copy,importlib.util,sys,unittest
from check import validate,parent,CODE
ours=sys.modules['check'];sys.modules['check']=parent
s=importlib.util.spec_from_file_location('clock_backend_fixtures',CODE.parent/'joinedbackend/check_test.py');f=importlib.util.module_from_spec(s);s.loader.exec_module(f);sys.modules['check']=ours
class Clock(unittest.TestCase):
 def data(self,mode='on'):
  d,m=f.JoinedBackend().data();d['display_clock_mode']=mode
  for r in d['backend_trace']['records']:
   if r['kind']==2:r.update(gpu_start=r['commit_start']+.0001,gpu_end=r['commit_start']+.0001)
  ticks=[dict(ordinal=i,timestamp=1.01+i*.02,target=1.03+i*.02,duration=.02,observed=1.011+i*.02) for i in range(1560)] if mode=='on' else []
  d['display_clock_trace']=dict(configured=True,enabled=mode=='on',started=True,stopped=True,invalid=False,overflow=False,count=len(ticks),window=42,display=7,start=1.0,stop=32.3,records=ticks);return d,m
 def test_on_and_same_ticket(self):
  r=validate(*self.data(),'on');self.assertEqual(len(r['joined_rows']),248);self.assertEqual(r['p95_rank_ticket_clock_join']['phase'],'measured');self.assertFalse(r['release_qualified'])
 def test_off(self):self.assertEqual(validate(*self.data('off'),'off')['outcome'],'VALID_DISPLAY_CLOCK_OFF_CONTROL')
 def test_callback_terminal_corruption(self):
  for k,v in [('configured',False),('enabled',False),('started',False),('stopped',False),('invalid',True),('overflow',True),('count',9000),('window',0),('display',0),('start',float('nan')),('stop',0)]:
   d,m=self.data();d['display_clock_trace'][k]=v
   with self.subTest(key=k),self.assertRaises(ValueError):validate(d,m,'on')
 def test_record_corruption(self):
  for k,v in [('ordinal',1),('observed',1000),('timestamp',float('nan')),('target',0),('duration',-1)]:
   d,m=self.data();d['display_clock_trace']['records'][-1][k]=v
   with self.subTest(key=k),self.assertRaises(ValueError):validate(d,m,'on')
 def test_missing_coverage_and_gpu(self):
  for change in [lambda d:d['display_clock_trace']['records'].__setitem__(slice(None),d['display_clock_trace']['records'][:-100]),lambda d:d['backend_trace']['records'][-1].update(gpu_start=0,gpu_end=0),lambda d:d['display_clock_trace']['records'].__setitem__(1,copy.deepcopy(d['display_clock_trace']['records'][0]))]:
   d,m=self.data();change(d);d['display_clock_trace']['count']=len(d['display_clock_trace']['records'])
   with self.assertRaises(ValueError):validate(d,m,'on')
 def test_late_callback_and_gap_are_retained(self):
  d,m=self.data();ticks=d['display_clock_trace']['records'];ticks[100]['observed']=ticks[100]['target']+.005
  del ticks[101:104]
  for i,r in enumerate(ticks):r['ordinal']=i
  d['display_clock_trace']['count']=len(ticks);v=validate(d,m,'on');self.assertGreater(v['callback_observation_minus_timestamp_ms']['max'],20);self.assertGreater(v['callback_timestamp_intervals_ms']['max'],70)
 def test_off_rejects_callbacks(self):
  d,m=self.data();d['display_clock_mode']='off';d['display_clock_trace']['enabled']=False
  with self.assertRaises(ValueError):validate(d,m,'off')
 def test_exact_timestamp_boundaries(self):
  d,m=self.data();gpu=d['backend_trace']['records'][-1]['gpu_end'];shown=d['before_close']['Samples'][-1]['Presented']
  stamps=[r['timestamp'] for r in d['display_clock_trace']['records'] if abs(r['timestamp']-gpu)>1e-8 and abs(r['timestamp']-shown)>1e-8]+[gpu,shown];stamps.sort()
  d['display_clock_trace']['records']=[dict(ordinal=i,timestamp=t,target=t+.02,duration=.02,observed=t+.001) for i,t in enumerate(stamps)];d['display_clock_trace']['count']=len(stamps)
  row=validate(d,m,'on')['joined_rows'][-1]
  self.assertEqual(row['gpu_preceding_tick']['timestamp'],gpu);self.assertEqual(row['shown_preceding_tick']['timestamp'],shown);self.assertEqual(row['shown_minus_nearest_reported_timestamp_ms'],0)
  self.assertEqual(row['reported_timestamps_between_gpu_and_shown'],sum(gpu<t<=shown for t in stamps))
if __name__=='__main__':unittest.main()
