import copy, unittest
from check import graph, validate, CAP

def event(kind, identity=0, size=0, submit=0, fence=0, call=0, count=0, cached=0, owner=1):
    return dict(kind=kind,owner=owner,id=identity,bytes=size,submit=submit,fence=fence,call=call,cached_count=count,cached_bytes=cached,complete=True)
def doc(rows, mode='on'):
    rows=copy.deepcopy(rows)
    for i,row in enumerate(rows):row.update(sequence=i,clock=1+i*.001)
    return dict(mode=1 if mode=='off' else 2,count=len(rows),pending=0,overflow=False,invalid=False,unstable=False,active_devices=0,records=rows)
def cycle(mode='on'):
    rows=[event(1),event(2,1,24,call=1),event(4,1,24,submit=1),event(5,1,24,submit=1,fence=1)]
    if mode=='on':
        rows += [event(6,1,24,submit=1,fence=1,count=1,cached=24),event(3,1,24,call=2),event(4,1,24,submit=2),event(5,1,24,submit=2,fence=2),event(6,1,24,submit=2,fence=2,count=1,cached=24),event(8),event(7,1,24),event(9)]
    else:rows += [event(7,1,24,submit=1),event(8),event(9)]
    return doc(rows,mode)

def byte_fixture():
    """Synthetic off-mode protocol document; supplies no GPU verification evidence."""
    names=['zero-area','small-a','small-b','bgra','unaligned-row','padded-offset-source','large-a','large-b','above-idle-cap']+[f'queued-{i}' for i in range(1,4)]+[f'size-cycle-{i}' for i in range(1,10)]+['other-device','excluded-r8','unsubmitted-release']
    shapes=[(0,2,24,0)]+[(3,2,24,24)]*3+[(7,5,140,140),(7,5,216,140)]+[(1519,1266,7692216,7692216)]*2+[(4096,2049,33570816,33570816)]+[(3,2,24,24)]*3+[(w,3,w*12,w*12) for w in range(2,11)]+[(3,2,24,24),(3,2,6,8),(3,2,24,24)]
    writes=[];cpu=[];events=[{**event(1), 'clock':99}]
    def add(k,i,n,owner=1,**kwargs):events.append({**event(k,i,n,owner=owner,**kwargs),'clock':130+len(events)*.001})
    for number,(name,(w,h,input_size,staging)) in enumerate(zip(names,shapes),1):
        row=dict(case=name,write=number,width=w,height=h,input_bytes=input_size,staging_bytes=staging,
                 bytes_per_pixel=1 if number==23 else 4,bgra=number==4,source_stride=3 if number==23 else 40 if number==6 else w*4,source_offset=16 if number==6 else 0)
        if number==1:row['no_transfer']=True
        elif number==24:row['unsubmitted']=True
        else:row.update(full_byte_match=True,readback_bytes=w*h*(1 if number==23 else 4),actual_sha256=f'{number:064x}',expected_sha256=f'{number:064x}')
        writes.append(row)
        cpu.append(dict(ordinal=number-1,complete=True,status=1,width=w,height=h,bytes=input_size,
                        stage_pitch=0 if number==1 else 4 if number==23 else w*4,stages=0 if number==1 else 31,
                        reuse_eligible=number not in (1,23),begin=100+number-.01,end=100+number+.01,
                        **{k:0 if number==1 else 1 for k in ('create_ns','map_ns','copy_ns','flush_ns','encode_ns')}))
        if number in (1,23):continue
        owner=2 if number==22 else 1
        if owner==2:events.append({**event(1,owner=2),'clock':121.5})
        events.append({**event(2,number,staging,call=number,owner=owner),'clock':100+number})
        if number in (10,11):continue
        if number==12:
            for i in (10,11,12):add(4,i,24,submit=12)
            for i in (10,11,12):add(5,i,24,submit=12,fence=12);add(7,i,24,submit=12)
        elif number==24:add(7,number,staging)
        else:add(4,number,staging,owner,submit=number);add(5,number,staging,owner,submit=number,fence=number);add(7,number,staging,owner,submit=number)
    # Per-resource event clocks, unlike reservations across independent resources,
    # must preserve order. Keep synthetic acquisition clocks inside its CPU call.
    for i,e in enumerate(events):e.update(sequence=i,complete=True)
    events.extend([{**event(k,owner=owner),'clock':200+i*.01,'sequence':len(events)+i} for i,(owner,k) in enumerate([(1,8),(1,9),(2,8),(2,9)])])
    trace=dict(mode=1,count=len(events),pending=0,overflow=False,invalid=False,unstable=False,active_devices=0,records=events)
    return dict(kind='staging-reuse-byte-check',mode='off',runtime_returned=True,error='<nil>',release_qualified=False,
                backend=5,device='synthetic',windows_created=0,surfaces_created=0,physical_gpu_overlap_forced=False,writes=writes,
                reuse_trace=trace,native_upload_trace=dict(enabled=True,pending=0,overflow=False,invalid_stage=False,count=24,records=cpu))

class Ledger(unittest.TestCase):
    def test_combined_protocol_document_and_counts(self):
        result=validate(byte_fixture(),'off');self.assertEqual((result['CPU_writes'],result['GPU_verified_readbacks']),(24,22))

    def test_combined_input_readback_eligibility_and_clock_refusals(self):
        mutations=[lambda d:d['writes'][5].update(source_offset=0),lambda d:d['native_upload_trace']['records'][1].update(reuse_eligible=False),lambda d:d['native_upload_trace']['records'][22].update(reuse_eligible=True),lambda d:d['native_upload_trace']['records'][5].update(bytes=200),lambda d:d['writes'][2].update(full_byte_match=False),lambda d:d['native_upload_trace']['records'][3].update(copy_ns=1e12)]
        for mutate in mutations:
            d=byte_fixture();mutate(d)
            with self.assertRaises(ValueError):validate(d,'off')

    def test_off_and_completed_reuse_with_exact_lease_history(self):
        off=graph(cycle('off'),'off');self.assertEqual((off['fresh'],off['takes'],off['peak_idle_count']),(1,0,0))
        on=graph(cycle(),'on');self.assertEqual((on['fresh'],on['takes'],on['peak_idle_bytes']),(1,1,24));self.assertEqual([r['submit'] for r in on['leases']],[1,2])

    def test_three_pending_leases_and_unsent_cleanup(self):
        rows=[event(1)]+[event(2,i,24,call=i) for i in (1,2,3)]
        rows += [event(4,i,24,submit=1) for i in (1,2,3)]
        rows += [event(5,i,24,submit=1,fence=1) for i in (1,2,3)]
        rows += [event(7,i,24,submit=1) for i in (1,2,3)]
        rows += [event(2,4,24,call=4),event(7,4,24),event(8),event(9)]
        r=graph(doc(rows,'off'),'off');self.assertEqual(r['fresh'],4);self.assertEqual(r['leases'][-1]['submit'],0)

    def test_evict_and_drain_are_distinct_from_checkout(self):
        rows=[event(1),event(2,1,24,call=1),event(4,1,24,submit=1),event(5,1,24,submit=1,fence=1),event(6,1,24,submit=1,fence=1,count=1,cached=24),event(10,1,24),event(7,1,24),event(8),event(9)]
        r=graph(doc(rows),'on');self.assertEqual((r['evictions'],r['takes']),(1,0))

    def test_reject_early_cache_fence_duplicate_or_wrong_device(self):
        changes=[(3,'fence',0),(4,'kind',3),(5,'owner',2),(5,'id',2),(6,'submit',1),(7,'fence',1),(8,'cached_bytes',25)]
        for index,key,value in changes:
            with self.subTest(index=index,key=key):
                d=cycle();d['records'][index][key]=value
                with self.assertRaises(ValueError):graph(d,'on')
        d=cycle();d['records'][3]['kind']=6
        with self.assertRaises(ValueError):graph(d,'on')

    def test_reject_missing_destroy_or_end_and_unsafe_destruction(self):
        for index in (9,10,11):
            d=cycle();d['records'].pop(index);d['count']-=1
            for i,r in enumerate(d['records']):r['sequence']=i
            with self.assertRaises(ValueError):graph(d,'on')
        d=cycle('off');d['records'].pop(3);d['count']-=1
        for i,r in enumerate(d['records']):r['sequence']=i
        with self.assertRaises(ValueError):graph(d,'off')

    def test_reject_recorder_partial_overflow_and_scalar_faults(self):
        for key,value in [('unstable',True),('pending',1),('overflow',True),('invalid',True),('active_devices',1),('count',4097),('pending',False)]:
            with self.subTest(key=key):
                d=cycle();d[key]=value
                with self.assertRaises(ValueError):graph(d,'on')
        for key,value in [('complete',False),('sequence',99),('bytes',True),('clock',float('nan'))]:
            d=cycle();d['records'][2][key]=value
            with self.assertRaises(ValueError):graph(d,'on')

    def test_reject_capacity_breach_and_off_mode_cache(self):
        d=cycle()
        for r in d['records']:
            if r['id']==1:r['bytes']=CAP+1
            if r['kind']==6:r['cached_bytes']=CAP+1
        with self.assertRaises(ValueError):graph(d,'on')
        d=cycle();d['records'][4]['cached_count']=9
        with self.assertRaises(ValueError):graph(d,'on')
        d=cycle();d['mode']=1
        with self.assertRaises(ValueError):graph(d,'off')

if __name__=='__main__':unittest.main()
