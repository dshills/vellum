#!/usr/bin/env python3
"""Private staging lifecycle and byte correctness; no presentation gate acceptance."""
from pathlib import Path
import importlib.util
CODE = Path(__file__).resolve().parent
s = importlib.util.spec_from_file_location('staging_base', CODE.parent/'check.py')
base = importlib.util.module_from_spec(s);s.loader.exec_module(base)
CAP = 32*1024*1024

def graph(trace, mode):
    need=base.need
    need(mode in ('off','on') and type(trace['mode']) is int and trace['mode']==(1 if mode=='off' else 2), 'reuse mode')
    need(type(trace['pending']) is int and trace['pending']==0 and trace['overflow'] is False and trace['invalid'] is False and trace['unstable'] is False, 'recorder terminal')
    rows=trace['records'];need(type(trace['count']) is int and trace['count']==len(rows)<=4096, 'journal bound')
    owners={};resources={};calls={};leases=[];peak_count=peak_bytes=fresh=takes=evictions=0
    for sequence,e in enumerate(rows):
        need(e['sequence']==sequence and e['complete'] is True, 'event sequence/completion')
        keys=('kind','owner','id','bytes','submit','fence','call','cached_count','cached_bytes')
        need(all(type(e[k]) is int and e[k]>=0 for k in keys) and base.finite(e['clock']) and e['clock']>0,'scalar event')
        k,o,i,n=e['kind'],e['owner'],e['id'],e['bytes']
        if k==1:
            need(o>0 and o not in owners and i==n==0, 'owner creation');owners[o]={'active':True,'idle':set(),'drained':False};continue
        need(o in owners and owners[o]['active'], 'unknown/ended owner');owner=owners[o]
        if k==8:
            need(i==n==0 and e['cached_count']==e['cached_bytes']==0 and not owner['drained'], 'drain')
            need(all(r['state'] in ('idle','destroyed') for r in resources.values() if r['owner']==o),'drain with outstanding lease')
            for identity in owner['idle']:resources[identity]['state']='evicted'
            owner['idle'].clear();owner['drained']=True;continue
        if k==9:
            need(owner['drained'] and not owner['idle'] and all(r['state']=='destroyed' for r in resources.values() if r['owner']==o),'owner end before destruction')
            owner['active']=False;continue
        need((not owner['drained'] or k==7) and i>0 and n>0, 'resource after drain/zero identity')
        if k==2:
            need(i not in resources and e['call']>0 and e['call'] not in calls, 'duplicate fresh allocation/call')
            resources[i]={'owner':o,'bytes':n,'state':'cpu','last_index':0,'clock':e['clock']};fresh+=1
        else:
            need(i in resources and resources[i]['owner']==o and resources[i]['bytes']==n,'resource/device/size mismatch')
            need(e['clock']>=resources[i]['clock'],'resource clock order');resources[i]['clock']=e['clock']
        r=resources[i]
        if k in (2,3):
            if k==3:
                need(mode=='on' and r['state']=='idle' and i in owner['idle'] and e['call']>0 and e['call'] not in calls,'checkout before completion/duplicate')
                owner['idle'].remove(i);r['state']='cpu';takes+=1
            lease={'call':e['call'],'id':i,'owner':o,'bytes':n,'acquire_sequence':sequence,'acquire_clock':e['clock'],'submit':0,'submit_sequence':None,'retire_sequence':None,'cache_sequence':None,'destroy_sequence':None}
            calls[e['call']]=lease;leases.append(lease);r['lease']=lease
        elif k==4:
            need(r['state']=='cpu' and e['submit']>r['last_index'] and e['fence']==0,'duplicate/invalid submission')
            r['state']='submitted';r['last_index']=e['submit'];r['lease'].update(submit=e['submit'],submit_sequence=sequence)
        elif k==5:
            need(r['state']=='submitted' and e['submit']==r['lease']['submit'] and e['fence']>=e['submit'],'completion without matching fence')
            r['state']='done';r['lease'].update(retire_sequence=sequence,retire_fence=e['fence'])
        elif k==6:
            need(mode=='on' and r['state']=='done' and n<=CAP and e['submit']==r['lease']['submit'] and e['fence']==r['lease']['retire_fence'],'cache without proven completion/oversize')
            owner['idle'].add(i);r['state']='idle';r['lease']['cache_sequence']=sequence
        elif k==10:
            need(r['state']=='idle' and i in owner['idle'],'evict non-idle');owner['idle'].remove(i);r['state']='evicted';evictions+=1
        elif k==7:
            need(r['state'] in ('cpu','done','evicted'),'destroy submitted/cached/duplicate resource')
            r['state']='destroyed';r['lease']['destroy_sequence']=sequence
        else:need(False,'unknown resource event')
        if k in (3,6):
            count=len(owner['idle']);size=sum(resources[i]['bytes'] for i in owner['idle'])
            need(count==e['cached_count']<=8 and size==e['cached_bytes']<=CAP,'idle caps/accounting')
            peak_count=max(peak_count,count);peak_bytes=max(peak_bytes,size)
    need(type(trace['active_devices']) is int and trace['active_devices']==0 and owners and all(not o['active'] for o in owners.values()) and all(r['state']=='destroyed' for r in resources.values()),'unclosed owner/resource')
    return {'fresh':fresh,'takes':takes,'evictions':evictions,'owners':len(owners),'peak_idle_count':peak_count,'peak_idle_bytes':peak_bytes,'leases':sorted(leases,key=lambda r:r['call'])}

def validate(doc,mode):
    need=base.need
    need(doc['kind']=='staging-reuse-byte-check' and doc['mode']==mode and doc['runtime_returned'] is True and doc['error']=='<nil>' and doc['release_qualified'] is False,'terminal/scope')
    need(doc['backend']==5 and doc['device'] and doc['windows_created']==doc['surfaces_created']==0 and doc['physical_gpu_overlap_forced'] is False,'Metal/headless scope')
    names=['zero-area','small-a','small-b','bgra','unaligned-row','padded-offset-source','large-a','large-b','above-idle-cap']+[f'queued-{i}' for i in range(1,4)]+[f'size-cycle-{i}' for i in range(1,10)]+['other-device','excluded-r8','unsubmitted-release']
    writes=doc['writes'];need(len(writes)==24 and [r['case'] for r in writes]==names,'exact case matrix')
    cpu=doc['native_upload_trace'];need(cpu['enabled'] is True and cpu['pending']==0 and cpu['overflow'] is False and cpu['invalid_stage'] is False and type(cpu['count']) is int and cpu['count']==len(cpu['records'])==24,'CPU trace terminal/count')
    lifecycle=graph(doc['reuse_trace'],mode);need(lifecycle['owners']==2,'two device scopes')
    bycall={r['call']:r for r in lifecycle['leases']};need(set(bycall)==set(range(2,23))|{24},'complete nonzero-write acquisitions')
    shapes=[(0,2,24,0)]+[(3,2,24,24)]*3+[(7,5,140,140),(7,5,216,140)]+[(1519,1266,7692216,7692216)]*2+[(4096,2049,33570816,33570816)]+[(3,2,24,24)]*3+[(w,3,w*12,w*12) for w in range(2,11)]+[(3,2,24,24),(3,2,6,8),(3,2,24,24)]
    def readback(w,n):
        need(w['full_byte_match'] is True and type(w['readback_bytes']) is int and w['readback_bytes']==n and type(w['actual_sha256']) is str and len(w['actual_sha256'])==64 and all(c in '0123456789abcdef' for c in w['actual_sha256']) and w['actual_sha256']==w['expected_sha256'],'full GPU readback')
    total=0
    for ordinal,(w,r) in enumerate(zip(writes,cpu['records']),1):
        need(all(type(w[k]) is int for k in ('width','height','input_bytes','staging_bytes','write')) and (w['width'],w['height'],w['input_bytes'],w['staging_bytes'])==shapes[ordinal-1], 'exact case dimensions/bytes')
        need(all(type(r[k]) is int for k in ('ordinal','status','width','height','bytes','stage_pitch','stages')), 'CPU scalar types')
        need(w['write']==ordinal and r['ordinal']==ordinal-1 and r['complete'] is True and r['status']==1,'write/CPU identity')
        need((r['width'],r['height'],r['bytes'])==(w['width'],w['height'],w['input_bytes']),'CPU input attribution')
        phases=('create_ns','map_ns','copy_ns','flush_ns','encode_ns')
        need(all(base.finite(r[k]) and r[k]>=0 for k in phases) and base.finite(r['begin']) and base.finite(r['end']) and 0<r['begin']<r['end'] and sum(r[k] for k in phases)<=(r['end']-r['begin'])*1e9+1,'CPU intervals')
        if ordinal==1:need(r['reuse_eligible'] is False and w['no_transfer'] is True and r['stages']==0 and r['stage_pitch']==0 and sum(r[k] for k in phases)==0,'zero area staging');continue
        if ordinal==23:
            need(r['reuse_eligible'] is False and r['stages']==31 and r['stage_pitch']==4 and w['bytes_per_pixel']==1,'non-allowlisted format entered reuse');readback(w,6);total+=6;continue
        need(w['bytes_per_pixel']==4 and w['bgra'] is (ordinal==4),'RGBA/BGRA case format')
        need((w['source_stride'],w['source_offset'])==((40,16) if ordinal==6 else (w['width']*4,0)),'source layout matrix')
        lease=bycall[ordinal];need(r['reuse_eligible'] is True and r['stages']==31 and r['stage_pitch']==w['width']*4 and lease['bytes']==w['staging_bytes']==w['width']*w['height']*4 and r['begin']<=lease['acquire_clock']<=r['end'],'stage/lease attribution')
        if ordinal==24:need(w['unsubmitted'] is True and lease['submit']==0 and lease['destroy_sequence'] is not None and lease['cache_sequence'] is None,'unsent work cached');continue
        readback(w,w['staging_bytes'])
        need(lease['submit']>0 and lease['retire_sequence'] is not None,'readback lease retirement');total+=w['readback_bytes']
    queued=[bycall[i] for i in (10,11,12)]
    need(len({r['id'] for r in queued})==3 and len({r['submit'] for r in queued})==1 and max(r['acquire_sequence'] for r in queued)<min(r['submit_sequence'] for r in queued),'three simultaneous pending leases')
    need(len({writes[i]['actual_sha256'] for i in (9,10,11)})==3 and writes[6]['actual_sha256']!=writes[7]['actual_sha256'],'different queued/large payloads')
    need(bycall[22]['owner']!=bycall[2]['owner'] and bycall[24]['owner']==bycall[2]['owner'],'device isolation')
    need(bycall[9]['bytes']>CAP and bycall[9]['cache_sequence'] is None,'oversize retained')
    if mode=='on':need(lifecycle['takes']>0 and lifecycle['evictions']>0 and lifecycle['peak_idle_count']==8,'reuse/entry cap actually exercised')
    else:need(lifecycle['takes']==lifecycle['evictions']==lifecycle['peak_idle_count']==0,'off mode cached')
    return {'outcome':'VALID_PRIVATE_STAGING_BYTE_CHECK','mode':mode,'CPU_writes':24,'GPU_verified_readbacks':22,'GPU_verified_bytes':total,'lifecycle':lifecycle,'release_qualified':False,'physical_GPU_overlap_verified':False,'interpretation':'Full-byte readbacks and scalar source-bound lifecycle accounting for these private texture staging resources. No physical allocation balance, device-fault injection, UI latency or shipping qualification.'}
