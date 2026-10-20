"""Synthetic join/schema controls. No native execution or presentation evidence."""
from pathlib import Path
import copy,hashlib,importlib.util,json,tempfile,unittest,contextlib,io,os,sys
from unittest.mock import patch
p=Path(__file__).with_name('check-presentation.py');s=importlib.util.spec_from_file_location('check_presentation',p);m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
SOURCE='7'*40
BINARY='a'*64

def fixture(tree=False,sample_count=120):
 config={'Method':'routed-tree-scroll'if tree else'routed-grid-scroll','Samples':sample_count,'Warmup':64,'DataSize':10000,'Variable':False,'Deep':False,'Novel':False,'GPURequested':False}
 tags=list(range(1,64+sample_count+3))if tree else list(range((1<<32)+1,(1<<32)+65))+list(range(1,sample_count+1))
 frames=[];inputs=[];samples=[]
 for index,tag in enumerate(tags):
  at=10+index*.016
  frame={key:0 for key in m.FRAME_KEYS if key.endswith('NS')};frame.update(Tag=tag,Window='perf',Generation=index+9,Completed=True,Failure='')
  frames.append(frame);inputs.append({'Scope':1,'Tag':tag,'Received':at})
  if index==0:samples.append({'Sequence':1,'Scope':1,'Tag':tag,'Generation':0,'Composition':1,'Armed':False,'DrawableID':0,'Acquired':at+.0001,'Presented':at+.005,'State':'presented'})
  samples.append({'Sequence':len(samples)+1,'Scope':1,'Tag':tag,'Generation':index+9,'Composition':len(samples)+1,'Armed':True,'DrawableID':0,'Acquired':at+.001,'Presented':at+.01,'State':'presented'})
 snapshot={'Scopes':[{'ID':1,'Window':'perf','Closed':False}],'Inputs':inputs,'Samples':samples,'Pending':0,'Skipped':0,'Unavailable':0,'Invalid':0,'Retired':0,'Overflow':False,'Failed':False,'Failure':''}
 before={'Frames':{'Samples':frames,'Dropped':0,'Pending':0,'Failure':''},'Presentation':snapshot}
 after=copy.deepcopy(before);after['Presentation']['Scopes'][0]['Closed']=True
 tail={'Sequence':len(samples)+1,'Scope':1,'Tag':0,'Generation':0,'Composition':len(samples)+1,'Armed':False,'DrawableID':0,'Acquired':samples[-1]['Acquired']+1,'Presented':0,'State':'retired'}
 after['Presentation']['Samples'].append(tail);after['Presentation']['Retired']=1
 return {'Schema':'vellum-presentation-capture-v1','Result':'CAPTURED_DIAGNOSTIC','Window':'perf','RunError':'','BinarySHA256':BINARY,'CutoffMethod':'acquisition-sequence prefix at completed CPU cohort','Config':config,'Build':{'vcs':'git','vcs.revision':SOURCE,'vcs.modified':'false','go_version':'go1.25.6','GOOS':'darwin','GOARCH':'arm64','CGO_ENABLED':'1'},'ContinuousWork':False,'WorkerRetired':True,'RuntimeReturned':True,'AcquisitionCutoff':len(samples),'ReceiptCutoff':len(inputs),'BeforeClose':before,'AfterClose':after}

def mutate_path(j,path,value):
 node=j
 for item in path[:-1]:node=node[item]
 node[path[-1]]=value

def replace_after_prefix(j):
 native=j['BeforeClose']['Presentation']['Samples']
 for index,sample in enumerate(native,1):sample['Sequence']=index
 j['AcquisitionCutoff']=len(native)
 tail=j['AfterClose']['Presentation']['Samples'][-1]
 tail['Sequence']=len(native)+1
 j['AfterClose']['Presentation']['Samples']=copy.deepcopy(native)+[tail]

def composition_cases():
 for kind in ['reused-final','regressed','changed-generation','changed-armed','changed-tag']:
  j=fixture();native=j['BeforeClose']['Presentation']['Samples']
  if kind=='reused-final':native[2]['Composition']=native[1]['Composition']
  if kind=='regressed':native[2]['Composition']=1
  if kind=='changed-generation':native[1]['Composition']=native[0]['Composition']
  if kind=='changed-armed':native[0]['Generation']=native[1]['Generation'];native[1]['Composition']=native[0]['Composition']
  if kind=='changed-tag':
   clone=copy.deepcopy(native[0]);clone.update(Tag=0,Acquired=10.00005,Presented=10.004);native.insert(0,clone)
  replace_after_prefix(j)
  yield kind,j

class PresentationJournalTests(unittest.TestCase):
 def test_both_methods_all_tags_and_zero_drawable_id(self):
  for tree in [False,True]:
   j=fixture(tree);out=m.validate(j,SOURCE,BINARY)
   self.assertEqual(out['measured_input_latency_ms']['count'],120)
   self.assertAlmostEqual(out['measured_input_latency_ms']['p95'],10,places=8)
   self.assertEqual(len(out['latencies']),len(j['BeforeClose']['Frames']['Samples']))
   self.assertEqual(out['tail_records'],1)
   self.assertFalse(out['cadence_qualified']);self.assertFalse(out['physical_latency_qualified'])
 def test_generation_intermediate_and_presentation_order(self):
  j=fixture();j['BeforeClose']['Presentation']['Samples'][0]['Presented']=10.015
  j['AfterClose']['Presentation']['Samples'][0]['Presented']=10.015
  out=m.validate(j,SOURCE,BINARY);self.assertTrue(all(i>=0 for i in out['prefix_intervals_ms']))
  j['BeforeClose']['Presentation']['Samples'][0]['Presented']=10.01
  j['AfterClose']['Presentation']['Samples'][0]['Presented']=10.01
  self.assertEqual(m.validate(j,SOURCE,BINARY)['equal_presentation_times'],1)
  # Wrong final generation must fail even when another prefix carries the tag.
  j['BeforeClose']['Presentation']['Samples'][1]['Generation']+=1
  with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_external_modfile_without_vcs_requires_binary_binding(self):
  j=fixture()
  for key in ['vcs','vcs.revision','vcs.modified']:j['Build'].pop(key)
  out=m.validate(j,SOURCE,BINARY)
  self.assertIn('VCS stamp absent',out['source_binding'])
  with self.assertRaises(ValueError):m.validate(j,SOURCE,'b'*64)
 def test_schema_clock_identity_loss_and_scope_controls(self):
  controls=[
   (['Schema'],'unknown'),(['Result'],'FAILED_DIAGNOSTIC'),(['RunError'],'canceled'),(['WorkerRetired'],False),(['RuntimeReturned'],False),(['ContinuousWork'],True),(['CutoffMethod'],'successful samples only'),
   (['AcquisitionCutoff'],0),(['AcquisitionCutoff'],9999),(['ReceiptCutoff'],0),(['Config','Samples'],True),(['Config','Warmup'],65),(['Config','DataSize'],3),(['Config','Deep'],True),(['Config','Method'],'physical-pointer'),(['Config','GPURequested'],1),
   (['Build','vcs.modified'],'true'),(['Build','vcs.revision'],'8'*40),(['Build','go_version'],'go1.27.2'),(['BinarySHA256'],'b'*64),(['Build','GOOS'],'linux'),(['Build','GOARCH'],'amd64'),(['Build','CGO_ENABLED'],'0'),(['Build','-tags'],'offscreen'),(['Build','-race'],'true'),(['Build','-gcflags'],'-N -l'),
   (['BeforeClose','Frames','Dropped'],1),(['BeforeClose','Frames','Pending'],1),(['BeforeClose','Frames','Failure'],'refused'),(['BeforeClose','Frames','Samples',0,'Completed'],False),(['BeforeClose','Frames','Samples',0,'Tag'],True),(['BeforeClose','Frames','Samples',0,'Generation'],0),(['BeforeClose','Frames','Samples',0,'Window'],'other'),(['BeforeClose','Frames','Samples',0,'Unavailable'],True),(['BeforeClose','Frames','Samples',0,'SourceDetails'],{}),(['BeforeClose','Frames','Samples',0,'TotalUIWorkNS'],1),
   (['BeforeClose','Presentation','Scopes',0,'Closed'],True),(['BeforeClose','Presentation','Scopes',0,'ID'],2),(['BeforeClose','Presentation','Failed'],True),(['BeforeClose','Presentation','Overflow'],True),(['BeforeClose','Presentation','Failure'],'refused'),(['BeforeClose','Presentation','Pending'],1),
   (['BeforeClose','Presentation','Inputs',0,'Received'],float('nan')),(['BeforeClose','Presentation','Inputs',0,'Tag'],True),(['BeforeClose','Presentation','Inputs',0,'Scope'],2),(['BeforeClose','Presentation','Inputs',1,'Received'],1),
   (['BeforeClose','Presentation','Samples',0,'Sequence'],2),(['BeforeClose','Presentation','Samples',0,'Scope'],2),(['BeforeClose','Presentation','Samples',0,'Composition'],0),(['BeforeClose','Presentation','Samples',0,'Acquired'],0),(['BeforeClose','Presentation','Samples',0,'Presented'],9),(['BeforeClose','Presentation','Samples',1,'State'],'pending'),(['BeforeClose','Presentation','Samples',1,'State'],'skipped'),(['BeforeClose','Presentation','Samples',1,'State'],'retired'),(['BeforeClose','Presentation','Samples',1,'State'],'unavailable'),(['BeforeClose','Presentation','Samples',1,'State'],'invalid'),(['BeforeClose','Presentation','Samples',1,'Armed'],False),
   (['AfterClose','Presentation','Scopes',0,'Closed'],False),(['AfterClose','Presentation','Samples',0,'DrawableID'],9),(['AfterClose','Presentation','Samples',-1,'Tag'],1),(['AfterClose','Presentation','Pending'],1),(['AfterClose','Frames','Samples',0,'Generation'],999)
  ]
  for path,value in controls:
   with self.subTest(path=path,value=value):
    j=fixture();mutate_path(j,path,value)
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_missing_duplicate_unknown_and_capacity_controls(self):
  for kind in ['missing-cpu','duplicate-cpu','missing-receipt','duplicate-receipt','extra-armed','unknown-field','unknown-nested','native-record-limit','tagged-tail','cutoff-filters-final']:
   with self.subTest(kind=kind):
    j=fixture();p=j['BeforeClose']
    if kind=='missing-cpu':p['Frames']['Samples'].pop()
    if kind=='duplicate-cpu':p['Frames']['Samples'].append(copy.deepcopy(p['Frames']['Samples'][0]))
    if kind=='missing-receipt':p['Presentation']['Inputs'].pop()
    if kind=='duplicate-receipt':p['Presentation']['Inputs'].append(copy.deepcopy(p['Presentation']['Inputs'][0]))
    if kind=='extra-armed':p['Presentation']['Samples'][0]['Armed']=True;p['Presentation']['Samples'][0]['Generation']=9
    if kind=='unknown-field':j['passes60Hz']=True
    if kind=='unknown-nested':p['Presentation']['Samples'][0]['extra']=True
    if kind=='native-record-limit':p['Presentation']['Samples']=[p['Presentation']['Samples'][0]]*65537
    if kind=='tagged-tail':j['AfterClose']['Presentation']['Samples'][-1].update(Tag=1,Generation=1)
    if kind=='cutoff-filters-final':j['AcquisitionCutoff']-=1
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_composition_identity_cannot_be_reused_or_changed(self):
  for kind,j in composition_cases():
   with self.subTest(kind=kind):
    with self.assertRaisesRegex(ValueError,'composition'):m.validate(j,SOURCE,BINARY)
 def test_same_composition_may_repeat_with_identical_stamp(self):
  j=fixture();native=j['BeforeClose']['Presentation']['Samples']
  clone=copy.deepcopy(native[0]);clone['Acquired']+=.00001;clone['Presented']+=.00001
  native.insert(1,clone);replace_after_prefix(j)
  self.assertEqual(m.validate(j,SOURCE,BINARY)['prefix_presentations'],len(native))
 def test_cli_validates_binary_hash_and_remains_nonqualifying(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'journal.json';raw=json.dumps(fixture()).encode();path.write_bytes(raw)
   args=['check-presentation.py',str(path),'--sha256',hashlib.sha256(raw).hexdigest(),'--source',SOURCE,'--binary-sha256',BINARY]
   for bad in [None,'hash','source','binary']:
    argv=list(args)
    if bad=='hash':argv[argv.index('--sha256')+1]='0'*64
    if bad=='source':argv[argv.index('--source')+1]='0'*40
    if bad=='binary':argv[argv.index('--binary-sha256')+1]='0'*64
    stdout=io.StringIO()
    with patch.object(sys,'argv',argv),contextlib.redirect_stdout(stdout):code=m.main()
    result=json.loads(stdout.getvalue())
    if bad is None:self.assertEqual(code,0);self.assertFalse(result['cadence_qualified'])
    else:self.assertEqual(code,1);self.assertEqual(result['result'],'REFUSED_JOURNAL')
 def test_nonregular_and_symlink_input_refused_without_blocking(self):
  with tempfile.TemporaryDirectory()as directory:
   root=Path(directory);target=root/'target';target.write_text('{}')
   link=root/'link';link.symlink_to(target)
   fifo=root/'fifo';os.mkfifo(fifo)
   for path in [root,link,fifo]:
    with self.assertRaises((ValueError,OSError)):m.load(path,'0'*64)
 def test_bounded_strict_file_loading(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'journal.json'
   for raw in [b'{"x":1,"x":2}',b'{"x":NaN}',b'{"x":Infinity}',b'\xff',b'{"x":']:
    path.write_bytes(raw)
    with self.assertRaises((ValueError,UnicodeError)):m.load(path,hashlib.sha256(raw).hexdigest())
   raw=json.dumps(fixture()).encode();path.write_bytes(raw)
   self.assertEqual(m.load(path,hashlib.sha256(raw).hexdigest())['Schema'],'vellum-presentation-capture-v1')
   with self.assertRaises(ValueError):m.load(path,'0'*64)
   # The byte guard refuses before parsing, using a small limit for this control.
   old=m.MAX_BYTES
   try:
    m.MAX_BYTES=8
    with self.assertRaises(ValueError):m.load(path,hashlib.sha256(raw).hexdigest())
   finally:m.MAX_BYTES=old
def paced_fixture(tree=False):
 j=fixture(tree,600);j['Schema']='vellum-presentation-capture-v2';j['ContinuousWork']=True;j['Config']['Paced']=True
 config=j['Config'];tags=[f['Tag']for f in j['BeforeClose']['Frames']['Samples']][-600:];cpu={f['Tag']:f for f in j['BeforeClose']['Frames']['Samples']};tag_index={tag:i for i,tag in enumerate(tags)}
 for receipt in j['BeforeClose']['Presentation']['Inputs']:
  if receipt['Tag']in tag_index:receipt['Received']=20+(tag_index[receipt['Tag']]*1_000_000_000//60+1_500_000)/1e9
 receipts={r['Tag']:r['Received']for r in j['BeforeClose']['Presentation']['Inputs']}
 for native in j['BeforeClose']['Presentation']['Samples']:
  if native['Tag']in tag_index:native['Acquired']=receipts[native['Tag']]+.001;native['Presented']=receipts[native['Tag']]+.010
 initial_tag=config['Warmup']+2 if tree else(1<<32)+config['Warmup']
 view={'ScrollX':0,'ScrollY':0,'Width':1299,'Height':1060,'Bounds':[0,0,1299,1060],'Scale':1,'Visible':64 if tree else 512,'Rows':64,'Columns':1 if tree else 8,'Retained':70 if tree else 700,'Materialized':66 if tree else 700,'LayoutSHA256':'c'*64}
 initial={'Tag':initial_tag,'Generation':cpu[initial_tag]['Generation'],'Frame':1000,'View':copy.deepcopy(view)};slots=[]
 for i,tag in enumerate(tags):
  phase=i%64;dx=0 if tree else 96 if phase in[7,15,23,31]else -96 if phase in[39,47,55,63]else 0
  dy=(12 if i%32<16 else -12)if tree else(12 if phase<32 else -12)
  view['ScrollX']+=dx;view['ScrollY']+=dy;scheduled=i*1_000_000_000//60
  slots.append({'Index':i,'Tag':tag,'ScheduledNS':scheduled,'State':'completed','StartNS':scheduled+100000,'DispatchNS':scheduled+1000000,'EnqueuedNS':scheduled+2000000,'AcknowledgedNS':scheduled+4000000,'InspectedNS':scheduled+5000000,'EndNS':scheduled+6000000,'Observation':{'Tag':tag,'Generation':cpu[tag]['Generation'],'Frame':1001+i,'View':copy.deepcopy(view)}})
 j['Pacing']={'Hz':60,'Clock':'Go monotonic relative nanoseconds; separate from native host clock','Method':'absolute rational deadlines; serial CPU acknowledgement; retained backlog','PlannedEndNS':10_000_000_000,'Started':True,'Completed':True,'HoldStartNS':slots[-1]['EndNS'],'HoldEndNS':10_000_000_000,'Failure':'','Initial':initial,'Slots':slots}
 sync_paced(j);return j

def sync_paced(j):
 replace_after_prefix(j);before=j['BeforeClose'];j['AfterClose']['Frames']=copy.deepcopy(before['Frames']);j['AfterClose']['Presentation']['Inputs']=copy.deepcopy(before['Presentation']['Inputs'])
 last=j['BeforeClose']['Presentation']['Samples'][-1]
 tail=j['AfterClose']['Presentation']['Samples'][-1];tail['Acquired']=last['Acquired']+1;tail['Composition']=last['Composition']+1

class PacedJournalTests(unittest.TestCase):
 def test_declared_grid_offsets_and_initial_horizontal_phase(self):
  # Independent checkpoints of the Go consumer's published64-event path.
  expected={7:(96,96),15:(192,192),23:(288,288),31:(384,384),39:(288,288),47:(192,192),55:(96,96),63:(0,0)}
  x=y=0
  for i in range(64):
   dx,dy=m.pace_delta(i,False);x+=dx;y+=dy
   if i<7:self.assertEqual(x,0)
   if i in expected:self.assertEqual((x,y),expected[i])
 def test_clock_bounds_include_endpoint_rounding(self):
  p=m.validate(paced_fixture(),SOURCE,BINARY)['paced_metrics'];origin_width=p['clock_origin_uncertainty_ns']/1e6
  for name in['first_scheduled_to_presented_ms_bounds','final_presentation_after_planned_end_ms_bounds']:
   left,right=p[name];self.assertGreater(right-left,origin_width+.000001)
  self.assertIn('common nanosecond rate',p['clock_correlation_model'])

 def test_grid_tree_paced_final_state_population(self):
  for tree in[False,True]:
   out=m.validate(paced_fixture(tree),SOURCE,BINARY);p=out['paced_metrics']
   self.assertEqual(out['result'],'VALID_PACED_NATIVE_FEEDBACK_JOURNAL');self.assertEqual(p['interval_count'],599);self.assertEqual(len(p['intervals_ms']),599);self.assertTrue(p['cadence_threshold_met']);self.assertTrue(p['input_latency_threshold_met']);self.assertFalse(out['cadence_qualified']);self.assertFalse(p['release_qualified']);self.assertGreaterEqual(p['clock_origin_uncertainty_ns'],0)
 def test_paced_cannot_relabel_legacy_or_omit_slots(self):
  for kind in['relabel-only','extra-field-legacy','remove-slot','duplicate-tag','move-end','early-slot','unstarted','failed','wrong-clock','wrong-hz','initial-generation','native-bracket']:
   with self.subTest(kind=kind):
    j=paced_fixture();p=j['Pacing']
    if kind=='relabel-only':j=fixture();j['Schema']='vellum-presentation-capture-v2'
    if kind=='extra-field-legacy':j['Schema']='vellum-presentation-capture-v1';j['ContinuousWork']=False
    if kind=='remove-slot':p['Slots'].pop()
    if kind=='duplicate-tag':p['Slots'][1]['Tag']=p['Slots'][0]['Tag']
    if kind=='move-end':p['PlannedEndNS']+=1
    if kind=='early-slot':p['Slots'][1]['StartNS']=p['Slots'][1]['ScheduledNS']-1
    if kind=='unstarted':p['Slots'][1]['State']='planned'
    if kind=='failed':p['Failure']='canceled'
    if kind=='wrong-clock':p['Clock']='native timestamp minus Go timer'
    if kind=='wrong-hz':p['Hz']=True
    if kind=='initial-generation':p['Initial']['Generation']+=1
    if kind=='native-bracket':p['Slots'][5]['DispatchNS']+=10000000;p['Slots'][5]['EnqueuedNS']+=10000000;p['Slots'][5]['AcknowledgedNS']+=10000000;p['Slots'][5]['InspectedNS']+=10000000;p['Slots'][5]['EndNS']+=10000000
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_noop_clamped_or_changed_geometry_refused(self):
  for key,value in [('ScrollY',0),('Width',20),('Bounds',[0,0,10,10]),('Scale',2),('Visible',1),('LayoutSHA256','bad')]:
   j=paced_fixture();j['Pacing']['Slots'][0]['Observation']['View'][key]=value
   with self.subTest(key=key),self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_equal_or_reordered_display_states_never_pass(self):
  for reordered in[False,True]:
   j=paced_fixture();finals=[s for s in j['BeforeClose']['Presentation']['Samples']if s['Armed']and s['Tag']<1<<32];finals[10]['Presented']=finals[11]['Presented']+(.0001 if reordered else 0);sync_paced(j)
   p=m.validate(j,SOURCE,BINARY)['paced_metrics'];self.assertFalse(p['cadence_threshold_met']);self.assertFalse(p['ordered_distinct_final_times']);self.assertTrue(p['collapsed_or_reordered_states'])
 def test_intermediate_draws_cannot_dilute_final_state_gaps(self):
  j=paced_fixture();native=j['BeforeClose']['Presentation']['Samples'];finals=[s for s in native if s['Armed']and s['Tag']<1<<32]
  for s in finals[300:]:s['Presented']+=.1
  expanded=[]
  for original in native:
   expanded.append(original)
   if original['Tag']<1<<32:
    for i in range(2):
     clone=copy.deepcopy(original);clone.update(Tag=0,Generation=0,Armed=False,Acquired=original['Acquired']+(i+1)*.000001,Presented=original['Presented']+(i+1)*.000001);expanded.append(clone)
  for i,s in enumerate(expanded,1):s['Sequence']=s['Composition']=i
  j['BeforeClose']['Presentation']['Samples']=expanded;sync_paced(j)
  out=m.validate(j,SOURCE,BINARY);self.assertGreater(len(out['prefix_intervals_ms']),1000);p=out['paced_metrics'];self.assertEqual(p['interval_count'],599);self.assertEqual(len(p['over_two_periods']),1);self.assertFalse(p['input_latency_threshold_met'])
 def test_backlog_and_final_overrun_are_reported(self):
  j=paced_fixture();p=j['Pacing']
  # Fixed deadlines stay unchanged while the serial work queue runs at25ms.
  for i,slot in enumerate(p['Slots']):
   start=i*25000000+100000
   for key,delta in [('StartNS',0),('DispatchNS',1000000),('EnqueuedNS',2000000),('AcknowledgedNS',20000000),('InspectedNS',23000000),('EndNS',24900000)]:slot[key]=start+delta
  tags={slot['Tag']:slot for slot in p['Slots']}
  for r in j['BeforeClose']['Presentation']['Inputs']:
   if r['Tag']in tags:r['Received']=20+(tags[r['Tag']]['DispatchNS']+500000)/1e9
  receipts={r['Tag']:r['Received']for r in j['BeforeClose']['Presentation']['Inputs']}
  for s in j['BeforeClose']['Presentation']['Samples']:
   if s['Tag']in tags:s['Acquired']=receipts[s['Tag']]+.001;s['Presented']=receipts[s['Tag']]+.01
  p['HoldStartNS']=p['HoldEndNS']=p['Slots'][-1]['EndNS'];sync_paced(j)
  out=m.validate(j,SOURCE,BINARY)['paced_metrics'];self.assertFalse(out['cadence_threshold_met']);self.assertGreater(out['final_work_overrun_ns'],0);self.assertGreater(out['worst_schedule_debt_ns'],0);self.assertTrue(out['iteration_deadline_misses'])


def startup_fixture(tree=False,paced=False):
 j=paced_fixture(tree)if paced else fixture(tree)
 native=j['BeforeClose']['Presentation']['Samples']
 for sample in native:sample['Composition']+=1
 j['AfterClose']['Presentation']['Samples'][-1]['Composition']+=1
 first=j['BeforeClose']['Presentation']['Inputs'][0]['Received']
 native.insert(0,{'Sequence':1,'Scope':1,'Tag':0,'Generation':0,'Composition':1,'Armed':False,'DrawableID':0,'Acquired':first-.01,'Presented':0,'State':'skipped'})
 j['BeforeClose']['Presentation']['Skipped']=j['AfterClose']['Presentation']['Skipped']=1
 replace_after_prefix(j)
 j['PrefixObservations']={'Records':len(native),'Presented':len(native)-1,'StartupSkipped':1}
 return j

class StartupObservationTests(unittest.TestCase):
 def test_both_consumers_serial_and_paced_preserve_populations(self):
  for tree in [False,True]:
   for paced in [False,True]:
    baseline=m.validate(paced_fixture(tree)if paced else fixture(tree),SOURCE,BINARY)
    j=startup_fixture(tree,paced);before=copy.deepcopy(j);out=m.validate(j,SOURCE,BINARY)
    self.assertEqual(j,before)
    self.assertEqual(out['startup_skipped'],1);self.assertEqual(out['prefix_records'],out['prefix_presentations']+1)
    self.assertEqual(out['before_state_counts']['skipped'],1)
    # The inserted raw startup record advances sequence identities by exactly
    # one; measured tags/generations/roles/times and populations stay identical.
    expected_latencies=[{**row,'sequence':row['sequence']+1}for row in baseline['latencies']]
    self.assertEqual(out['latencies'],expected_latencies)
    for key in ['measured_input_latency_ms','prefix_intervals_ms','prefix_presentations','equal_presentation_times']:self.assertEqual(out[key],baseline[key])
    if paced:
     self.assertEqual(out['paced_metrics'],baseline['paced_metrics']);self.assertEqual(out['paced_metrics']['interval_count'],599)
    self.assertFalse(out['cadence_qualified']);self.assertFalse(out['physical_latency_qualified'])
 def test_precise_startup_boundary_and_all_other_loss_states(self):
  for kind in ['at-first-receipt','after-first-receipt','armed','tagged','generation','foreign-scope','nonzero-presented','missing-receipt','foreign-receipt','zero-acquired','nan-acquired','infinite-acquired','pending','unavailable','invalid','retired','missing-final','changed-postclose','truncated-prefix','missing-accounting','wrong-records','wrong-presented','wrong-skipped','extra-accounting','bool-accounting','failed-capture']:
   with self.subTest(kind=kind):
    j=startup_fixture();s=j['BeforeClose']['Presentation']['Samples'][0];first=j['BeforeClose']['Presentation']['Inputs'][0]['Received']
    if kind=='at-first-receipt':s['Acquired']=first
    if kind=='after-first-receipt':s['Acquired']=first+.00001
    if kind=='armed':s.update(Armed=True,Tag=j['BeforeClose']['Presentation']['Inputs'][0]['Tag'],Generation=9)
    if kind=='tagged':s['Tag']=j['BeforeClose']['Presentation']['Inputs'][0]['Tag']
    if kind=='generation':s['Generation']=9
    if kind=='foreign-scope':s['Scope']=2
    if kind=='nonzero-presented':s['Presented']=first-.005
    if kind=='missing-receipt':j['BeforeClose']['Presentation']['Inputs']=[]
    if kind=='foreign-receipt':j['BeforeClose']['Presentation']['Inputs'][0]['Scope']=2
    if kind=='zero-acquired':s['Acquired']=0
    if kind=='nan-acquired':s['Acquired']=float('nan')
    if kind=='infinite-acquired':s['Acquired']=float('inf')
    if kind in ['pending','unavailable','invalid','retired']:
     s['State']=kind;j['BeforeClose']['Presentation']['Skipped']=0;j['BeforeClose']['Presentation'][kind.title()]=1
    if kind=='missing-final':j['BeforeClose']['Presentation']['Samples'][-1]['Armed']=False
    if kind=='missing-accounting':j.pop('PrefixObservations')
    if kind=='wrong-records':j['PrefixObservations']['Records']-=1
    if kind=='wrong-presented':j['PrefixObservations']['Presented']+=1
    if kind=='wrong-skipped':j['PrefixObservations']['StartupSkipped']=0
    if kind=='extra-accounting':j['PrefixObservations']['hidden']=True
    if kind=='bool-accounting':j['PrefixObservations']['StartupSkipped']=True
    if kind=='failed-capture':j['Result']='FAILED_DIAGNOSTIC';j['RunError']='presentation capture: lost prefix feedback'
    replace_after_prefix(j)
    if kind in ['pending','unavailable','invalid','retired']:
     j['AfterClose']['Presentation']['Skipped']=0;j['AfterClose']['Presentation'][kind.title()]=1
    if kind=='changed-postclose':j['AfterClose']['Presentation']['Samples'][0]['Acquired']-=.001
    if kind=='truncated-prefix':j['AcquisitionCutoff']-=1
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_failed_capture_is_not_retroactively_repaired(self):
  j=startup_fixture(paced=True);j['Result']='FAILED_DIAGNOSTIC';j['RunError']='presentation capture: lost prefix feedback'
  raw=json.dumps(j).encode()
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'failed.json';path.write_bytes(raw)
   with self.assertRaisesRegex(ValueError,'failed or unsupported capture'):m.validate(m.load(path,hashlib.sha256(raw).hexdigest()),SOURCE,BINARY)
   self.assertEqual(path.read_bytes(),raw)


# Literal source-protocol checkpoints, independent of the validator expression.
DEEP_INITIAL_Y={10000:68076,100000:675576}
def shift_paced_origin(j,y):
 delta=y-j['Pacing']['Initial']['View']['ScrollY']
 j['Pacing']['Initial']['View']['ScrollY']+=delta
 for slot in j['Pacing']['Slots']:slot['Observation']['View']['ScrollY']+=delta
 return j

def deep_paced_fixture(size,gpu=False):
 j=paced_fixture(True);j['Config'].update(DataSize=size,Deep=True,GPURequested=gpu)
 return shift_paced_origin(j,DEEP_INITIAL_Y[size])

class PacedInitialOriginTests(unittest.TestCase):
 def test_deep_source_origins_are_accepted(self):
  for size in [10000,100000]:
   for gpu in [False,True]:
    with self.subTest(size=size,gpu=gpu):
     j=deep_paced_fixture(size,gpu);out=m.validate(j,SOURCE,BINARY)
     self.assertEqual(j['Pacing']['Initial']['View']['ScrollY'],DEEP_INITIAL_Y[size])
     self.assertEqual(out['result'],'VALID_PACED_NATIVE_FEEDBACK_JOURNAL')
     self.assertEqual(out['measured_input_latency_ms']['count'],600)
     self.assertEqual(out['paced_metrics']['interval_count'],599)
     self.assertTrue(out['paced_metrics']['cadence_threshold_met'])
     self.assertFalse(out['paced_metrics']['release_qualified'])
 def test_deep_arbitrary_or_wrong_size_origins_are_refused(self):
  for size,other in [(10000,100000),(100000,10000)]:
   expected=DEEP_INITIAL_Y[size]
   for y in [0,expected-1,expected+1,DEEP_INITIAL_Y[other]]:
    with self.subTest(size=size,y=y):
     j=shift_paced_origin(deep_paced_fixture(size),y)
     with self.assertRaisesRegex(ValueError,'paced initial/warmup baseline differs'):m.validate(j,SOURCE,BINARY)
 def test_wide_tree_and_all_grid_modes_keep_zero_origin(self):
  for tree,sizes in [(False,[10000,1000000]),(True,[10000,100000])]:
   for size in sizes:
    for variable in ([False]if tree else[False,True]):
     for gpu in [False,True]:
      with self.subTest(tree=tree,size=size,variable=variable,gpu=gpu):
       j=paced_fixture(tree);j['Config'].update(DataSize=size,Variable=variable,GPURequested=gpu)
       self.assertEqual(m.validate(j,SOURCE,BINARY)['paced_metrics']['interval_count'],599)
       shift_paced_origin(j,1)
       with self.assertRaisesRegex(ValueError,'paced initial/warmup baseline differs'):m.validate(j,SOURCE,BINARY)
 def test_deep_origin_cannot_be_transferred_to_other_configuration(self):
  for kind in ['wide','wrong-size','invalid-size','tree-variable','grid-method','grid-deep']:
   with self.subTest(kind=kind):
    j=deep_paced_fixture(10000)
    if kind=='wide':j['Config']['Deep']=False
    if kind=='wrong-size':j['Config']['DataSize']=100000
    if kind=='invalid-size':j['Config']['DataSize']=1000000
    if kind=='tree-variable':j['Config']['Variable']=True
    if kind=='grid-method':j['Config']['Method']='routed-grid-scroll'
    if kind=='grid-deep':j=paced_fixture();j['Config']['Deep']=True
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)


class DisplayBackpressureDeclarationTests(unittest.TestCase):
 def test_optional_bounded_declaration_changes_no_existing_result_or_gate(self):
  for tree in [False,True]:
   for paced in [False,True]:
    factory=paced_fixture if paced else fixture
    baseline=factory(tree);expected=m.validate(baseline,SOURCE,BINARY)
    self.assertNotIn('DisplayBackpressureNS',expected['config'])
    for timeout in [100_000_000,1_000_000_000,5_000_000_000]:
     j=copy.deepcopy(baseline);j['Config']['DisplayBackpressureNS']=timeout
     actual=m.validate(j,SOURCE,BINARY)
     unchanged=copy.deepcopy(expected);unchanged['config']['DisplayBackpressureNS']=timeout
     self.assertEqual(actual,unchanged)
 def test_malformed_declaration_failed_journal_and_late_prefix_stay_refused(self):
  for value in [None,True,False,0,-1,99_999_999,5_000_000_001,1e9,'1000000000']:
   j=fixture();j['Config']['DisplayBackpressureNS']=value
   with self.subTest(value=value),self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
  j=fixture();j['Config']['DisplayBackpressureNS']=1_000_000_000;j['Result']='FAILED_DIAGNOSTIC'
  with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
  j=startup_fixture();j['Config']['DisplayBackpressureNS']=1_000_000_000
  skipped=j['BeforeClose']['Presentation']['Samples'][0]
  skipped['Acquired']=j['BeforeClose']['Presentation']['Inputs'][0]['Received']+.00001
  replace_after_prefix(j)
  with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)

if __name__=='__main__':unittest.main()
