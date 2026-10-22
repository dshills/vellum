import copy,math,unittest
from unittest.mock import patch
import check as v

def data():
 match={'width':640,'height':480,'scale':1,'pixel_format':80,'drawables':2,'vsync':True,'transactions':False,'timeout':False,'rect':[0,0,640,480]}
 state={k:value for k,value in match.items() if k!='rect'};state.update(clock=9.8,pid=10,front_pid=10,active=True,key=True,visible=True,losses=0)
 ticks=[dict(ordinal=i,timestamp=10+i/60,target=10+(i+1)/60,observed=10+i/60+.010,duration=1/60) for i in range(2400)]
 requests=[dict(sequence=n,anchor=8+(n-1)*8,phase=(n-1)%4,target=ticks[8+(n-1)*8]['timestamp']+(3+((n-1)%4+.5)/4)/60) for n in range(1,249)]
 trace=dict(configured=True,invalid=False,overflow=False,pending=0,acquiring=0,attempts=[],events=[],requests=requests)
 rows=[];events=[]
 def event(kind,attempt,tag,request,at,shown=0):return dict(kind=kind,attempt=attempt,tag=tag,request=request,at=at,observed=at+.000001,shown=shown)
 for i in range(250):
  n=i-1;receipt=9.81+i*.08 if i<2 else requests[n-1]['target']+.0005;acquired=receipt+.003;shown=0 if i==0 else receipt+.03
  row=dict(sequence=0 if i<2 else n,phase=0 if i<2 else 1 if n<=128 else 2,receipt=receipt,acquired=acquired,shown=shown,accepted=0 if i<2 else n,drawable_id=i+1,present_request=receipt+.0035,commit_start=receipt+.004,commit_end=receipt+.0041,scheduled_observed=receipt+.0043,complete_observed=receipt+.006,gpu_start=receipt+.0042,gpu_end=receipt+.005,status=4);rows.append(row)
  request=max(0,n);a=i+1;trace['attempts'].append(dict(attempt=a,scope=1,tag=i,generation=row['accepted'],drawable=a,state=3))
  if i>=2:events.append(event(1,0,i,n,receipt))
  events.extend([event(2,a,i,request,receipt+.001),event(3,a,i,request,acquired),event(5,a,i,request,receipt+.031,shown)])
 states={}
 for j,e in enumerate(sorted(events,key=lambda e:e['observed'])):
  census=lambda:[a for a in sorted(states) if states[a]==2]
  e.update(ordinal=j,pending_before=census(),acquiring_before=sum(s==1 for s in states.values()))
  if e['kind']!=1:states[e['attempt']]={2:1,3:2,5:3}[e['kind']]
  e.update(pending_after=census(),acquiring_after=sum(s==1 for s in states.values()));trace['events'].append(e)
 doc=dict(kind='native',release_qualified=False,runtime_returned=True,match=match,begin=state,end=dict(state,clock=55),records=rows,lateness_ns=[500_000]*248,pending_trace=trace,display_clock_trace=dict(configured=True,enabled=True,started=True,stopped=True,invalid=False,overflow=False,count=len(ticks),window=1,display=3,start=10,stop=54,records=ticks))
 return doc,match

class Controls(unittest.TestCase):
 def test_complete_native_and_actual_phase_populations(self):
  d,m=data();result=v.validate(d,m,'native');self.assertEqual(len(result['rows']),248);self.assertEqual([b['count'] for b in result['phase_buckets'].values()],[30]*4);self.assertFalse(result['release_qualified']);self.assertTrue(all(r['pending_at_receipt_hook']==[] for r in result['rows']));self.assertEqual(v.compare(result,dict(result,kind='framework'))['outcome'],'VALID_PHASE_MATCHED_PAIR')
 def test_framework_new_joins_preserve_existing_parent(self):
  d,m=data();d['kind']='framework';d['phase_bridges']=[dict(sequence=r['sequence'],target=r['target'],host=10,go_bracket_ns=1000) for r in d['pending_trace']['requests']];source=[]
  for i,r in enumerate(d['records'][2:]):
   source.append(dict(accepted_revision=i+1,opaque_tag=i+2,generation=i+1,drawable_id=r['drawable_id'],phase='warmup' if i<128 else 'measured',raw_host_clocks={k:r[k] for k in ['receipt','acquired','shown']},backend_commands=[dict(kind=0,gpu_end=r['gpu_end']),dict(kind=2,gpu_end=r['gpu_end'],present_request=r['present_request'])]))
  checked={'parent_validation':{'joined_rows':source}}
  with patch.object(v.parent,'validate',return_value=checked) as parent:
   self.assertEqual(v.validate(d,m,'framework')['parent_validation'],checked);parent.assert_called_once_with(d,m,'on')
  with patch.object(v.parent,'validate',side_effect=ValueError('existing strict rejection')):
   with self.assertRaisesRegex(ValueError,'existing strict rejection'):v.validate(d,m,'framework')
 def test_replay_detects_wrong_census_transition_and_terminal(self):
  for mutate in ['before','after','terminal','duplicate','missing','boolean_id']:
   d,m=data();t=d['pending_trace']
   if mutate=='before':t['events'][2]['pending_before']=[]
   elif mutate=='after':t['events'][1]['pending_after']=[]
   elif mutate=='terminal':t['pending']=1
   elif mutate=='duplicate':t['events'][2]['kind']=3
   elif mutate=='missing':t['events'].pop(0)
   else:t['events'][1]['pending_after']=[True]
   with self.subTest(mutate=mutate),self.assertRaises(ValueError):v.validate(d,m,'native')
 def test_phase_identity_receipt_join_and_timing_refusals(self):
  for mutate in ['phase','anchor','receipt','generation','drawable','callback_time','nan','gap','clock_bracket']:
   d,m=data();t=d['pending_trace']
   if mutate=='phase':t['requests'][0]['phase']=False
   elif mutate=='anchor':t['requests'][1]['anchor']+=1
   elif mutate=='receipt':d['records'][2]['receipt']+=.0021
   elif mutate=='generation':t['attempts'][2]['generation']+=1
   elif mutate=='drawable':t['attempts'][2]['drawable']=999
   elif mutate=='callback_time':next(e for e in t['events'] if e['kind']==5 and e['attempt']==3)['shown']+=.001
   elif mutate=='nan':t['requests'][0]['target']=math.nan
   elif mutate=='gap':d['display_clock_trace']['records'][3]['timestamp']+=.0002
   else:d['records'][2]['gpu_end']=d['records'][2]['shown']+.001
   with self.subTest(mutate=mutate),self.assertRaises(ValueError):v.validate(d,m,'native')
 def test_actual_phase_mismatch_refuses_pair(self):
  d,m=data();a=v.validate(d,m,'native');b=copy.deepcopy(a);b['kind']='framework';b['phase_buckets']['2']['receipt_phase_error_ms']['p95']+=1.1
  with self.assertRaisesRegex(ValueError,'per-phase receipt mismatch'):v.compare(a,b)
 def test_missing_request_and_foreground_parent_rejection(self):
  d,m=data();d['pending_trace']['requests'].pop()
  with self.assertRaises(ValueError):v.validate(d,m,'native')
  d,m=data();d['end']['losses']=1
  with self.assertRaises(ValueError):v.validate(d,m,'native')

if __name__=='__main__':unittest.main()
