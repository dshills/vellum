import unittest,copy,math
from check import validate,native_doc
class Controls(unittest.TestCase):
 def data(self,kind='native'):
  match={'width':640,'height':480,'scale':1,'pixel_format':80,'drawables':2,'vsync':True,'transactions':False,'timeout':False,'rect':[0,0,640,480]}
  state={**match,'clock':1,'pid':10,'front_pid':10,'active':True,'key':True,'visible':True,'losses':0};state.pop('rect')
  rows=[{'sequence':0,'phase':0,'receipt':1.01,'acquired':1.02,'shown':0,'accepted':0},{'sequence':0,'phase':0,'receipt':1.03,'acquired':1.04,'shown':1.05,'accepted':0}]
  for i in range(1,249):
   t=2+(i-1)*.120;rows.append({'sequence':i,'phase':1 if i<=128 else 2,'receipt':t,'acquired':t+.01,'shown':t+.05,'accepted':i})
  d={'kind':kind,'release_qualified':False,'runtime_returned':True,'match':match,'begin':state,'end':dict(state,clock=40),'records':rows,'lateness_ns':[1000]*248}
  if kind=='framework':
   samples=[{'Sequence':i+1,'Scope':1,'Tag':0 if r['phase']==0 else r['sequence'],'Generation':0 if r['phase']==0 else r['sequence']+10,'Armed':r['phase']!=0,'State':'skipped' if r['shown']==0 else 'presented','Acquired':r['acquired'],'Presented':r['shown']} for i,r in enumerate(rows)]
   d['before_close']={'Scopes':[{'ID':1,'Closed':False}],'Samples':samples,'Inputs':[{'Scope':1,'Tag':r['sequence'],'Received':r['receipt']} for r in rows[2:]],'Failed':False,'Overflow':False,'Pending':0}
   d['after_close']=copy.deepcopy(d['before_close']);d['after_close']['Scopes'][0]['Closed']=True;d['error']='<nil>'
   d['measurements']={'Pending':False,'Samples':[{'Scope':1,'AcceptedRevision':i,'ViewRevisionBefore':i+9,'Frame':{'Tag':i,'Generation':i+10,'Completed':True,'Failure':''}} for i in range(1,249)]}
  return d,match
 def refused(self,d,m):
  with self.assertRaises(ValueError):validate(d,m)
 def test_native_valid_with_retained_startup_loss(self):
  d,m=self.data();v=validate(d,m);self.assertEqual(v['setup_losses'],1);self.assertEqual(v['measured'],120);self.assertFalse(v['release_qualified'])
 def test_framework_valid_join(self):
  d,m=self.data('framework');self.assertEqual(validate(d,m)['receipt_to_present_ms']['p95'],validate(*self.data())['receipt_to_present_ms']['p95'])
 def test_match_drift(self):
  d,m=self.data();m=copy.deepcopy(m);m['width']=800;self.refused(d,m)
 def test_transient_focus_loss(self):
  d,m=self.data();d['end']['losses']=1;self.refused(d,m)
 def test_background_pid(self):
  d,m=self.data();d['begin']['front_pid']=99;self.refused(d,m)
 def test_missing_measured_feedback(self):
  d,m=self.data();d['records'][-1]['shown']=0;self.refused(d,m)
 def test_replaced_state(self):
  d,m=self.data();d['records'][-1]['accepted']=247;self.refused(d,m)
 def test_no_readiness(self):
  d,m=self.data();d['records'][1]['shown']=3;self.refused(d,m)
 def test_nonfinite(self):
  d,m=self.data();d['records'][-1]['shown']=math.nan;self.refused(d,m)
 def test_schedule_loss(self):
  d,m=self.data();d['lateness_ns'][1]=25_000_000;self.refused(d,m)
 def test_missing_receipt(self):
  d,m=self.data('framework');d['before_close']['Inputs'].pop();self.refused(d,m)
 def test_ambiguous_armed_frame(self):
  d,m=self.data('framework');d['before_close']['Samples'].append(copy.deepcopy(d['before_close']['Samples'][-1]));d['after_close']['Samples']=copy.deepcopy(d['before_close']['Samples']);self.refused(d,m)
 def test_unaccepted_model(self):
  d,m=self.data('framework');d['measurements']['Samples'][-1]['AcceptedRevision']=247;self.refused(d,m)
 def test_missing_shutdown(self):
  d,m=self.data('framework');d['after_close']['Scopes'][0]['Closed']=False;self.refused(d,m)
 def test_native_terminal_required(self):
  with self.assertRaises(ValueError):native_doc([self.data()[0]])
 def test_nonfinite_setup(self):
  d,m=self.data();d['records'][0]['acquired']=math.nan;self.refused(d,m)
 def test_unmatched_armed_frame(self):
  d,m=self.data('framework');r=copy.deepcopy(d['before_close']['Samples'][-1]);r['Tag']=999;r['Sequence']=999;d['before_close']['Samples'].append(r);d['after_close']['Samples']=copy.deepcopy(d['before_close']['Samples']);self.refused(d,m)
 def test_numeric_visibility_is_not_a_boolean(self):
  d,m=self.data();d['begin']['visible']=1;self.refused(d,m)
if __name__=='__main__':unittest.main()
