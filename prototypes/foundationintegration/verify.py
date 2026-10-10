"""Verify outcomes without upgrading protocol probes into live OS evidence."""
from pathlib import Path
import json

def require(condition,message="evidence validation failed"):
 if not condition:raise RuntimeError(message)
r=Path(__file__).resolve().parent
p=json.loads((r/'evidence/saturation.json').read_text());before=p['before']
keys=['sequence','text','selected_start','selected_length','marked_start','marked_length','has_marked']
require(before['has_marked'])
require(len(p['refused_states'])==5 and p['transport']['Refused']>=7, 'refusal probes missing')
require(all(all(state[k]==before[k] for k in keys) for state in p['refused_states']))
require(p['transport']['Queue']['Outstanding']==256)
p=json.loads((r/'evidence/native-lifetime.json').read_text())
require(p['native_input_ack']==13 and p['rejected_actions']==0)
require(p['transport']['Queue']['Outstanding']==0 and p['transport']['Queue']['PayloadBytes']==0)
require(p['transport']['Queue']['Closed'])
log=(r/'evidence/native-lifetime.log').read_text()
require('native_attach before_show=true' in log and 'native_detach before_gpu_view_release=true' in log)
for name,cycles in [('lifetime-release-1000.json',1000),('lifetime-release-5000.json',5000)]:
 p=json.loads((r/'evidence'/name).read_text())
 require(p['cycles']==cycles and p['attachments']==cycles and p['detaches']==cycles)
 require(p['accepted']==p['applied_after_close']==cycles*11 and p['failure_cleanup']==1)
 require(p['samples'][-1]['goroutines']==1)
p=json.loads((r/'evidence/lifetime-final-1000.json').read_text())
require(p['publication_contracts']['atomic_invalid_publications']==6)
require(p['publication_contracts']['snapshotless_cancel_refused'])
require(p['publication_contracts']['backward_cancel_direction'])
require(p['publication_contracts']['stale_geometry_rejected'])
require(p['publication_contracts']['obsolete_baseline_invalidated'])
print('Native admission, saturation, bounded drain, attachment ordering and lifecycle counts verified.')
