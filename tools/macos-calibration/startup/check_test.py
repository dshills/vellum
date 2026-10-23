import copy,unittest
from unittest.mock import patch
import check as v

def data():
 m={'width':640,'height':480,'scale':1,'pixel_format':80,'drawables':2,'vsync':True,'transactions':False,'timeout':False,'rect':[0,0,640,480]}
 state={k:x for k,x in m.items() if k!='rect'};state.update(clock=1,pid=10,front_pid=10,active=True,key=True,visible=True,losses=0,layer_class='VellumPresentationLayer')
 samples=[];attempts=[];events=[]
 for i,(acquired,shown) in enumerate([(1.02,0),(1.04,1.05)],1):
  drawable=100+i;samples.append(dict(Sequence=i,Scope=1,Tag=0,Generation=0,Armed=False,DrawableID=drawable,State='presented' if shown else 'skipped',Acquired=acquired,Presented=shown))
  attempts.append(dict(attempt=i,scope=1,tag=0,generation=0,drawable=drawable,state=3))
  for kind,at in [(2,acquired-.001),(3,acquired),(5,max(acquired+.001,shown)+.001)]:
   pending_before=[i] if kind==5 else [];pending_after=[i] if kind==3 else []
   events.append(dict(ordinal=len(events),kind=kind,attempt=i,tag=0,request=0,at=at,observed=at+.000001,shown=shown if kind==5 else 0,pending_before=pending_before,pending_after=pending_after,acquiring_before=int(kind==3),acquiring_after=int(kind==2)))
 before=dict(Scopes=[dict(ID=1,Closed=False)],Samples=samples,Inputs=[],Failed=False,Overflow=False,Pending=0);after=copy.deepcopy(before);after['Scopes'][0]['Closed']=True
 d=dict(kind='framework',release_qualified=False,runtime_returned=True,error='<nil>',match=m,begin=state,end=dict(state,clock=2),before_close=before,after_close=after,measurements=dict(Pending=False,Samples=[]),pending_trace=dict(configured=True,invalid=False,overflow=False,pending=0,acquiring=0,requests=[],attempts=attempts,events=events),display_clock_trace=dict(configured=True,enabled=True,started=True,stopped=True,invalid=False,overflow=False,count=0,records=[],start=1.06,stop=1.07))
 return d,m

class Startup(unittest.TestCase):
 def test_positive_feedback_join_and_retained_startup_skip(self):
  d,m=data();r=v.validate(d,m);self.assertEqual(r['positive_presentations'],1);self.assertEqual(r['parent_setup']['setup_losses'],1);self.assertEqual(r['routed_inputs'],0);self.assertEqual(r['display_callbacks'],0);self.assertFalse(r['release_qualified']);self.assertFalse(r['phase_pair_qualified'])
 def test_missing_original_metrics_and_missing_hooks_refused(self):
  for what in ['metrics','hooks']:
   d,m=data()
   if what=='metrics':d['before_close']['Samples']=None
   else:d['pending_trace'].update(attempts=[],events=[])
   with self.subTest(what=what),self.assertRaises(ValueError):v.validate(d,m)
 def test_identity_lifecycle_and_scope_refusals(self):
  for what in ['no_positive','acquired','drawable','callback','duplicate','pending','requests','overflow','layer','observer','focus','scope']:
   d,m=data()
   if what=='no_positive':
    for key in ['before_close','after_close']:
     for r in d[key]['Samples']:r.update(State='skipped',Presented=0)
   elif what=='acquired':d['after_close']['Samples'][0]['Acquired']+=.0001
   elif what=='drawable':d['after_close']['Samples'][0]['DrawableID']=999
   elif what=='callback':d['pending_trace']['events'][-1]['shown']+=.001
   elif what=='duplicate':d['after_close']['Samples'].append(copy.deepcopy(d['after_close']['Samples'][-1]))
   elif what=='pending':d['pending_trace']['pending']=1
   elif what=='requests':d['pending_trace']['requests']=[{}]
   elif what=='overflow':d['pending_trace']['overflow']=True
   elif what=='layer':d['begin']['layer_class']='CAMetalLayer'
   elif what=='observer':d['display_clock_trace']['stopped']=False
   elif what=='focus':d['end']['front_pid']=99
   else:d['after_close']['Scopes'][0]['Closed']=False
   with self.subTest(what=what),self.assertRaises(ValueError):v.validate(d,m)
 def test_original_setup_validator_not_bypassed(self):
  d,m=data()
  with patch.object(v.base,'validate',side_effect=ValueError('original rejection')) as original:
   with self.assertRaisesRegex(ValueError,'original rejection'):v.validate(d,m)
   original.assert_called_once_with(d,m,pilot=True)
if __name__=='__main__':unittest.main()
