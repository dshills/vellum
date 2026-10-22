"""Startup-only proof; does not qualify latency or relax the phase protocol."""
from pathlib import Path
import importlib.util,sys
p=Path(__file__).resolve().parent.parent/'presentationphase/check.py'
s=importlib.util.spec_from_file_location('startup_phase',p);phase=importlib.util.module_from_spec(s);sys.modules[s.name]=phase;s.loader.exec_module(phase)
base=phase.base;need=base.need

def validate(doc,match):
 need(doc['kind']=='framework','framework startup only')
 for key in ['before_close','after_close']:
  need(type(doc[key]['Samples']) is list and 0<len(doc[key]['Samples'])<=4096,'startup samples required')
 checked=base.validate(doc,match,pilot=True)
 trace=doc['pending_trace'];receipts,seen,_,_=phase.replay(trace)
 need(not trace['requests'] and not receipts,'startup cannot contain phase requests/receipts')
 need(seen,'startup native acquisition hooks absent')
 used=set();positive=0
 for r in doc['after_close']['Samples']:
  need(base.finite(r['Acquired']) and base.finite(r['Presented']) and r['Acquired']>0,'startup scalar clocks')
  hits=[a for a,events in seen.items() if 3 in events and events[3]['at']==r['Acquired'] and trace['attempts'][a-1]['drawable']==r['DrawableID']]
  need(len(hits)==1 and hits[0] not in used,'unique metrics/native acquisition join');a=hits[0];used.add(a);events=seen[a]
  need(5 in events and events[5]['shown']==r['Presented'],'startup callback join')
  need((r['State']=='presented' and r['Presented']>=r['Acquired']) or (r['State']=='skipped' and r['Presented']==0 and r['Tag']==r['Generation']==0 and r['Armed'] is False),'startup feedback state')
  positive+=r['Presented']>0
 need(positive>0,'positive native startup presentation')
 for endpoint in ['begin','end']:
  need(type(doc[endpoint]['layer_class']) is str and doc[endpoint]['layer_class'].endswith('VellumPresentationLayer'),'unexpected actual Metal layer class')
 need(doc['begin']['layer_class']==doc['end']['layer_class'],'native layer class changed')
 t=doc['display_clock_trace']
 need(all(t[k] is True for k in ['configured','enabled','started','stopped']) and t['invalid'] is False and t['overflow'] is False,'startup display observer terminal')
 need(type(t['count']) is int and 0<=t['count']==len(t['records'])<=8192,'startup observer bounds')
 need(base.finite(t['start']) and base.finite(t['stop']) and doc['begin']['clock']<=t['start']<=t['stop']<=doc['end']['clock'],'startup observer lifetime')
 return {'outcome':'VALID_CORRECTED_FRAMEWORK_STARTUP_ONLY','parent_setup':checked,'presentation_samples':len(doc['after_close']['Samples']),'positive_presentations':positive,'attempts':len(trace['attempts']),'events':len(trace['events']),'terminal_pending':trace['pending'],'terminal_acquiring':trace['acquiring'],'layer_class':doc['begin']['layer_class'],'display_callbacks':t['count'],'scope_closed':True,'routed_inputs':0,'phase_pair_qualified':False,'release_qualified':False,'gate_changed':False}
