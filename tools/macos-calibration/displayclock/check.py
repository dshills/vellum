"""Retain display-link observations without changing existing backend acceptance."""
from pathlib import Path
import bisect,importlib.util
CODE=Path(__file__).resolve().parent
s=importlib.util.spec_from_file_location('display_clock_backend',CODE.parent/'joinedbackend/check.py');parent=importlib.util.module_from_spec(s);s.loader.exec_module(parent);base=parent.base

def validate(doc,match,mode):
 result=parent.validate(doc,match);need=base.need
 need(mode in ('off','on') and doc['display_clock_mode']==mode,'display clock mode')
 t=doc['display_clock_trace'];need(t['configured'] is True and t['enabled'] is (mode=='on') and t['started'] is True and t['stopped'] is True and t['invalid'] is False and t['overflow'] is False,'display clock terminal')
 need(type(t['count']) is int and type(t['records']) is list and t['count']==len(t['records'])<=8192,'display clock capacity/count')
 need(type(t['window']) is int and t['window']>0 and type(t['display']) is int and t['display']>0,'display/window identity')
 need(base.finite(t['start']) and base.finite(t['stop']) and doc['begin']['clock']<=t['start']<t['stop']<=doc['end']['clock'],'display clock lifetime')
 ticks=t['records']
 if mode=='off':
  need(not ticks,'disabled display observer emitted callbacks');return {'outcome':'VALID_DISPLAY_CLOCK_OFF_CONTROL','parent_validation':result,'release_qualified':False,'gate_changed':False}
 need(len(ticks)>=2,'display callback coverage')
 for i,r in enumerate(ticks):
  need(type(r['ordinal']) is int and r['ordinal']==i,'callback ordinal')
  need(all(base.finite(r[k]) for k in ['observed','timestamp','target','duration']) and 0<r['timestamp']<r['target'] and r['duration']>0 and t['start']<=r['observed']<=t['stop'],'callback clocks')
  if i:need(ticks[i-1]['observed']<=r['observed'] and ticks[i-1]['timestamp']<r['timestamp'] and ticks[i-1]['target']<r['target'],'callback order')
 stamps=[r['timestamp'] for r in ticks];rows=[]
 for source in result['joined_rows']:
  present=source['backend_commands'][-1];gpu=present['gpu_end'];shown=source['raw_host_clocks']['shown']
  need(present['kind']==2 and present['gpu_start']>0 and present['gpu_start']<=gpu<=shown,'final present GPU timing required for clock join')
  # Equality belongs to the preceding tick; counted timestamps are (GPU, shown].
  lo=bisect.bisect_right(stamps,gpu);hi=bisect.bisect_right(stamps,shown)
  need(0<lo<len(ticks) and 0<hi<len(ticks),'display timestamp coverage around GPU/shown')
  nearest=min([hi-1,hi],key=lambda i:abs(stamps[i]-shown))
  # Retrospective reported timestamp/target observations, not a proof that a
  # callback had been delivered or a drawable could latch at that instant.
  rows.append({'accepted_revision':source['accepted_revision'],'scope':source['scope'],'opaque_tag':source['opaque_tag'],'generation':source['generation'],'drawable_id':source['drawable_id'],'phase':source['phase'],'present_gpu_end':gpu,'shown':shown,'gpu_preceding_tick':ticks[lo-1],'gpu_following_tick':ticks[lo],'shown_preceding_tick':ticks[hi-1],'shown_following_tick':ticks[hi],'nearest_reported_tick_to_shown':ticks[nearest],'reported_timestamps_between_gpu_and_shown':hi-lo,'gpu_to_next_reported_timestamp_ms':(stamps[lo]-gpu)*1000,'shown_minus_nearest_reported_timestamp_ms':(shown-stamps[nearest])*1000,'present_gpu_end_to_shown_ms':(shown-gpu)*1000})
 measured=rows[128:];ticket=result['p95_rank_ticket'];ranked=next(r for r in rows if r['scope']==ticket['scope'] and r['opaque_tag']==ticket['opaque_tag'] and r['generation']==ticket['generation'])
 return {'outcome':'VALID_DISPLAY_CLOCK_JOINED_DIAGNOSTIC','parent_validation':result,'joined_rows':rows,'p95_rank_ticket_clock_join':ranked,'callback_count':len(ticks),'callback_timestamp_intervals_ms':base.stats([(b['timestamp']-a['timestamp'])*1000 for a,b in zip(ticks,ticks[1:])]),'callback_observation_minus_timestamp_ms':base.stats([(r['observed']-r['timestamp'])*1000 for r in ticks]),'target_minus_timestamp_ms':base.stats([(r['target']-r['timestamp'])*1000 for r in ticks]),'measured_clock_ms':{k:base.stats([r[k] for r in measured]) for k in ['gpu_to_next_reported_timestamp_ms','shown_minus_nearest_reported_timestamp_ms','present_gpu_end_to_shown_ms']},'reported_timestamp_counts':base.stats([r['reported_timestamps_between_gpu_and_shown'] for r in measured]),'release_qualified':False,'gate_changed':False,'interpretation':'View-linked Core Animation timestamps/targets are reported clocks, not independent optical scanout, guaranteed latch opportunities or missed-vsync proof. Callback gaps and lateness retained; retrospective joins do not mean the callback was already delivered. One off/on observer-presence control, no calibrated overhead, pure-cost causal comparison or ordinary qualification.'}
