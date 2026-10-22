"""Three-transition structural evidence only; no latency or release acceptance."""
from pathlib import Path
import importlib.util,collections
s=importlib.util.spec_from_file_location('native_commands_base',Path(__file__).resolve().parent.parent/'check.py');base=importlib.util.module_from_spec(s);s.loader.exec_module(base)

def validate(doc,match):
    need=base.need
    need(doc['kind']=='framework' and doc['diagnostic_mode']=='three-transition-raster-commands' and doc['release_qualified'] is False and doc['runtime_returned'] is True and doc['error']=='<nil>','terminal/scope')
    need(doc['match']==match,'locked match')
    need(match['scale']==1 and match['drawables']==2 and match['vsync'] is True and match['transactions'] is False and match['timeout'] is False,'original native policy')
    begin,end=doc['begin'],doc['end']
    for state in [begin,end]:
        need(state['active'] is True and state['key'] is True and state['visible'] is True and state['losses']==0 and type(state['pid']) is int and state['pid']>0 and state['pid']==state['front_pid'],'foreground')
        need(all(state[k]==match[k] for k in base.FIELDS),'endpoint configuration')
    need(begin['pid']==end['pid'] and base.finite(begin['clock']) and base.finite(end['clock']) and 0<begin['clock']<end['clock'],'endpoint clock/identity')
    before,after=doc['before_close'],doc['after_close']
    for snap in [before,after]:need(snap['Failed'] is False and snap['Overflow'] is False and snap['Pending']==0,'presentation recorder failure')
    need(len(before['Scopes'])==len(after['Scopes'])==1 and after['Scopes'][0]['Closed'] is True and before['Scopes'][0]['ID']==after['Scopes'][0]['ID'],'scope closure')
    samples=before['Samples'];need(after['Samples'][:len(samples)]==samples,'prefix rewritten')
    need(doc['measurements']['Pending'] is False,'pending public measurement')
    measurements=doc['measurements']['Samples'];receipts=before['Inputs'];need(len(measurements)==len(receipts)==3,'three accepted requests')
    keyed={(r['Scope'],r['Tag']):r for r in receipts};need(len(keyed)==3,'duplicate receipt identity')
    selected=[];responses=[]
    for index,m in enumerate(measurements,1):
        f=m['Frame'];key=(m['Scope'],f['Tag'])
        need(f['Completed'] is True and not f['Failure'] and not f.get('Unavailable',False) and m['AcceptedRevision']==index and f['Generation']>m['ViewRevisionBefore'],'routed state/frame')
        need(key in keyed and key[0]==before['Scopes'][0]['ID'],'receipt/scope join');received=keyed[key]['Received']
        hits=[r for r in samples if r['Scope']==key[0] and r['Tag']==key[1] and r['Generation']==f['Generation'] and r['Armed']]
        need(len(hits)==1 and hits[0]['State']=='presented','unique positive armed feedback');r=hits[0]
        need(all(base.finite(v) for v in [received,r['Acquired'],r['Presented']]) and begin['clock']<=received<=r['Acquired']<r['Presented']<=end['clock'] and r['Presented']-received<.120,'response deadline/clock')
        if responses:need(responses[-1]['presented']<received,'overlapping requests')
        selected.append(r['Sequence']);responses.append({'accepted':index,'scope':key[0],'opaque_tag':key[1],'generation':f['Generation'],'received':received,'presented':r['Presented']})
    need(len(set(selected))==3,'reused presentation')
    first=responses[0]['received'];setup=[r for r in samples if r['Acquired']<first and (r['Scope'],r['Tag']) not in keyed]
    need(any(r['State']=='presented' and r['Presented']>0 for r in setup),'positive startup readiness')
    for r in samples:
        need(r['State']=='presented' and r['Presented']>0 or r in setup and r['Tag']==0 and r['Generation']==0 and r['Armed'] is False and r['State']=='skipped' and r['Presented']==0,'non-startup feedback loss')
        if r['Armed'] and r['Acquired']>=first:need(r['Sequence'] in selected,'unmatched armed acquisition')
    late=doc['lateness_ns'];phases=doc['request_schedule_phases'];need(len(late)==len(phases)==len(doc['receipt_lateness_ns'])==3,'schedule records')
    for i,(n,p) in enumerate(zip(late,phases),1):
        need(type(n) is int and 0<=n<=20_000_000 and all(type(p[k]) is int for k in ['Sequence','WakeLatenessNS','ForegroundCheckNS','PostCheckLatenessNS']),'schedule scalars')
        need(p['Sequence']==i and p['PostCheckLatenessNS']==n and p['ForegroundCheckNS']>=0 and n>=p['WakeLatenessNS']+p['ForegroundCheckNS'],'schedule join/ordering')
    capture=doc['raster_commands'];need(capture['Overflow'] is False and capture['Invalid'] is False and capture['CohortStarted'] is True,'command recorder failure')
    frames=capture['Frames'];need(0<len(frames)<=32 and sum(len(f['Items']) for f in frames)<=1024,'command bounds')
    decisions=collections.Counter();colors=[];admitted=0;seen=False;startup_extents=[];rect=match['rect']
    for f in frames:
        need(f['Complete'] is True and type(f['Cohort']) is bool and type(f['OwnedScanner']) is bool,'incomplete frame')
        target=f['Target'];w=target['Max']['X']-target['Min']['X'];h=target['Max']['Y']-target['Min']['Y']
        if f['Cohort']:
            seen=True;admitted+=1;need(target=={'Min':{'X':0,'Y':0},'Max':{'X':match['width'],'Y':match['height']}},'admitted extent drift')
        else:need(not seen and 0<=w<=4096 and 0<=h<=4096,'startup boundary/order');startup_extents.append(target)
        patch=[]
        for i,r in enumerate(f['Items']):
            need(type(r['Index']) is int and r['Index']==i and r['Type'] in ['*render.ContextPush','*render.ContextPop','*render.Path','*pimage.Params'],'command index/type')
            need(type(r['Scroll']) is bool and r['Scroll'] is False,'nested scroll outside scope')
            if r['Type']=='*pimage.Params':
                need(r['Decision'] in ['covered-next','cropped-uniform','render-image'],'missing image decision')
                if f['Cohort']:decisions[r['Decision']]+=1
                rr=r['Rect']
                if r['HasUniformRGBA'] and [rr['Min']['X'],rr['Min']['Y'],rr['Max']['X'],rr['Max']['Y']]==rect and r['Command']==0 and r['Mask'] is False:
                    patch.append(r['UniformRGBA'])
            elif r['Type']=='*render.Path':
                need(r['Decision']=='render-path' and type(r['HasFill']) is bool and type(r['HasStroke']) is bool and type(r['CropNoPaintCompatible']) is bool,'missing path decision')
                if f['Cohort']:decisions['render-path']+=1
            else:need(r['Decision']=='','context mislabeled')
        if f['Cohort'] and patch and (not colors or colors[-1]!=patch[-1]):colors.append(patch[-1])
    red=[65535,0,0,65535];green=[0,65535,0,65535]
    need(admitted>=3 and colors in [[red,green,red],[green,red,green,red]],'admitted patch paint sequence')
    return {'outcome':'VALID_THREE_TRANSITION_RASTER_STRUCTURE','responses':responses,'renderer_calls':len(frames),'admitted_renderer_calls':admitted,'recorded_items':sum(len(f['Items']) for f in frames),'admitted_root_decisions':dict(decisions),'admitted_patch_colors':colors,'retained_startup_extents':startup_extents,'latency_qualified':False,'physical_input_verified':False,'interpretation':'Three routed model/presentation transitions and bounded root replay metadata. Admission marks locked geometry, not input tags; scalar color order is not GPU/readback/displayed-byte proof. Cropped branch may write zero or several bands. No timing distribution, per-source/tag attribution, observer neutrality or release acceptance.'}
