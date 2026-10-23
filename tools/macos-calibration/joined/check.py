"""Same-input absolute raster/upload/schedule timeline; diagnostic, never a gate pass."""
from pathlib import Path
import importlib.util, math
CODE=Path(__file__).resolve().parent
s=importlib.util.spec_from_file_location('joined_upload_parent',CODE.parent/'upload/check.py')
upload=importlib.util.module_from_spec(s);s.loader.exec_module(upload);base=upload.base
EPS_NS=1000
WORK=('InputNS','OwnerNS','ReconcileNS','LayoutPaintNS','GatherNS','RasterUploadNS','EncodeSubmitNS')
FRAME=(*WORK,'QueuedNS','AcquireNS','PresentCallNS','TotalUIWorkNS')
SEGMENTS=('wake_to_checked','checked_to_postcheck','postcheck_to_receipt','receipt_to_raster','raster','raster_to_upload','upload','upload_to_acquire','acquire_to_shown')

def validate(doc,match):
    parent=upload.validate(doc,match,'on');need=base.need
    need(doc['joined_mode']=='absolute-raster-upload-schedule-v1','joined marker')
    trace=doc['raster_trace']
    need(trace['Configured'] is True and trace['Overflow'] is False and trace['Invalid'] is False and type(trace['Pending']) is int and trace['Pending']==0,'raster recorder terminal')
    rasters=trace['Records'];need(type(rasters) is list and 0<len(rasters)<=512,'raster capacity')
    for i,r in enumerate(rasters):
        need(type(r['Ordinal']) is int and r['Ordinal']==i and r['Begun'] is True and r['Complete'] is True and r['Returned'] is True,'raster bracket identity/completion')
        need(base.finite(r['Begin']) and base.finite(r['End']) and 0<r['Begin']<r['End'],'raster clock')
        rect=r['Target'];need(all(type(rect[e][a]) is int for e in ['Min','Max'] for a in ['X','Y']),'raster rectangle scalars')
        need(rect['Min']=={'X':0,'Y':0} and 0<=rect['Max']['X']<=4096 and 0<=rect['Max']['Y']<=4096,'raster extent')
    phases=doc['request_schedule_phases'];anchors=doc['request_schedule_anchors'];receipt_offsets=doc['receipt_lateness_ns']
    need(len(phases)==len(anchors)==len(receipt_offsets)==248,'schedule/anchor counts')
    receipts={(r['Scope'],r['Tag']):r['Received'] for r in doc['before_close']['Inputs']}
    presentations=doc['before_close']['Samples'];cpu=doc['native_upload_trace']['records'];used=set();rows=[]
    for index,m in enumerate(doc['measurements']['Samples']):
        f=m['Frame'];key=(m['Scope'],f['Tag']);receipt=receipts[key];p=phases[index];a=anchors[index]
        need(all(type(v) is int and v>0 for v in [key[0],key[1],f['Generation']]),'opaque identity scalar')
        need(all(type(p[k]) is int for k in ['Sequence','WakeLatenessNS','ForegroundCheckNS','PostCheckLatenessNS']),'schedule scalars')
        need(p['Sequence']==a['Sequence']==index+1 and type(a['Sequence']) is int and p['PostCheckLatenessNS']==doc['lateness_ns'][index] and p['ForegroundCheckNS']>=0 and p['PostCheckLatenessNS']>=p['WakeLatenessNS']+p['ForegroundCheckNS'],'schedule join/ordering')
        need(type(receipt_offsets[index]) is int and receipt_offsets[index]>=p['PostCheckLatenessNS'] and type(a['TargetFromReceiptStartNS']) is int and a['TargetFromReceiptStartNS']==-receipt_offsets[index] and type(a['ReceiptBracketNS']) is int and a['ReceiptBracketNS']>=0,'receipt Go-clock bracket/offset')
        need(all(base.finite(a[k]) for k in ['WakeHost','CheckedHost','PostCheckHost']),'schedule native clock')
        need(doc['begin']['clock']<=a['WakeHost']<=a['CheckedHost']<=a['PostCheckHost']<=receipt,'schedule native order')
        sources=f['SourceDetails'];need(sources['Compositions']==1 and len(sources['Records'])==1,'fixture requires exactly one source/composition')
        source=sources['Records'][0];need(source['Kind']==1 and source['ImageChanged'] is True and source['SetFromGoImageAttempts']==1 and source['ImageCacheHits']==0,'one changed CPU scene/image API')
        need(all(type(f[k]) is int and f[k]>=0 for k in FRAME) and f['TotalUIWorkNS']==sum(f[k] for k in WORK),'raw frame interval/total accounting')
        hits=[x for x in presentations if x['Scope']==key[0] and x['Tag']==key[1] and x['Generation']==f['Generation'] and x['Armed']]
        need(len(hits)==1,'presentation identity');feedback=hits[0];acquired=feedback['Acquired'];shown=feedback['Presented']
        need(type(feedback['DrawableID']) is int and feedback['DrawableID']>=0,'drawable identity')
        native=[u for u in cpu if receipt<=u['begin'] and u['end']<=acquired];need(len(native)==1,'one native upload')
        u=native[0];matches=[v for v in rasters if receipt<=v['Begin'] and v['End']<=u['begin']]
        need(len(matches)==1,'unique preceding raster bracket');r=matches[0]
        need(r['Ordinal'] not in used,'raster reused');used.add(r['Ordinal'])
        need(r['Target']=={'Min':{'X':0,'Y':0},'Max':{'X':match['width'],'Y':match['height']}},'raster joined geometry')
        need((r['End']-r['Begin'])*1e9<=source['RasterNS']+EPS_NS,'raster exceeds enclosing Go interval')
        clocks=[a['WakeHost'],a['CheckedHost'],a['PostCheckHost'],receipt,r['Begin'],r['End'],u['begin'],u['end'],acquired,shown]
        need(all(clocks[j]<=clocks[j+1] for j in range(len(clocks)-1)),'absolute timeline order')
        intervals={k:(clocks[i+1]-clocks[i])*1e9 for i,k in enumerate(SEGMENTS)}
        wake_total=(shown-a['WakeHost'])*1e9;receipt_total=(shown-receipt)*1e9
        need(abs(sum(intervals.values())-wake_total)<=EPS_NS and abs(sum(intervals[k] for k in SEGMENTS[3:])-receipt_total)<=EPS_NS,'disjoint partition closure')
        rows.append({'accepted_revision':m['AcceptedRevision'],'phase':'warmup' if index<128 else 'measured','scope':key[0],'opaque_tag':key[1],'generation':f['Generation'],'presentation_sequence':feedback['Sequence'],'drawable_id':feedback['DrawableID'],'raster_ordinal':r['Ordinal'],'upload_ordinal':u['ordinal'],'raw_host_clocks':dict(zip(['wake','checked','postcheck','receipt','raster_begin','raster_end','upload_begin','upload_end','acquired','shown'],clocks)),'schedule_go_ns':p,'receipt_lateness_go_ns':receipt_offsets[index],'receipt_clock_bracket_go_ns':a['ReceiptBracketNS'],'intervals_ns':intervals,'receipt_to_shown_ns':receipt_total,'wake_to_shown_ns':wake_total,'frame_work_ns':{k:f[k] for k in FRAME},'source_nested_ns':{k:source[k] for k in ['DrawNS','RasterNS','CopyNS','SetFromGoImageNS']},'native_upload_stages_ns':{k:u[k] for k in upload.PHASES}})
    first=min(receipts.values());prefix=[r for r in rasters if r['Ordinal'] not in used]
    need(len(prefix)<=128 and all(r['End']<=first for r in prefix),'unmatched raster after admission')
    need(len(used)==248,'all raster joins');measured=rows[128:]
    ranked=sorted(measured,key=lambda r:(r['receipt_to_shown_ns'],r['accepted_revision']));p95=ranked[math.ceil(len(ranked)*.95)-1]
    return {'outcome':'VALID_JOINED_RASTER_UPLOAD_SCHEDULE_DIAGNOSTIC','warmup':128,'measured':120,'retained_startup_raster_records':prefix,'joined_rows':rows,'interval_ms':{k:base.stats([r['intervals_ns'][k]/1e6 for r in measured]) for k in SEGMENTS},'receipt_to_shown_ms':base.stats([r['receipt_to_shown_ns']/1e6 for r in measured]),'p95_rank_ticket':p95,'over_33_3_ms_observations':sum(r['receipt_to_shown_ns']>33_300_000 for r in measured),'parent_validation':parent,'release_qualified':False,'gate_changed':False,'physical_input_verified':False,'interpretation':'Absolute segments share native host time and partition each input, not independent quantiles. Go nested durations, FIFO queue time and request lateness remain separate, not added to the native partition. Scheduled target has no native timestamp: no guessed Go/native clock conversion. Clock-read brackets and observer overhead retained, not calibrated. Acquisition-to-shown includes remaining encoding/submit/GPU/compositor work; no GPU-phase attribution or displayed-pixel proof.'}
