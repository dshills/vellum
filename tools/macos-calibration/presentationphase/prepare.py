"""Prepare fresh private consumers; shipping sources and native archive stay sealed."""
from pathlib import Path
import argparse,hashlib,json,shutil

def replace(text,old,new):
    if text.count(old)!=1:raise ValueError('recipe source changed: '+old[:80])
    return text.replace(old,new)

def prepare(repo,prior,out):
    repo,prior,out=[p.resolve() for p in [repo,prior,out]]
    if out.exists():raise ValueError('fresh output required')
    identity=json.loads((prior/'build-identity.json').read_text())
    for name,digest in identity['pins'].items():
        if hashlib.sha256(Path(name).read_bytes()).hexdigest()!=digest:raise ValueError('compiled source drift: '+name)
    code=repo/'tools/macos-calibration/presentationphase';out.mkdir();consumer=out/'consumer';shutil.copytree(prior/'consumer',consumer)
    for name in ['framework.mod','framework.sum','overlay.json','match.json']:shutil.copyfile(prior/name,out/name)
    for name in ['control.h','control.m','control_darwin.go']:shutil.copyfile(code/name,consumer/name)
    display=(consumer/'display_clock_impl.m').read_text()
    display=replace(display,'#include "display_clock.h"','#include "display_clock.h"\n#include "control.h"')
    display=replace(display,'dc.rows[dc.count++]=(struct DCTick){now,timestamp,target,duration};','dc.rows[dc.count++]=(struct DCTick){now,timestamp,target,duration};\n pc_tick(now,timestamp,target,duration);')
    (consumer/'display_clock_impl.m').write_text(display)
    main=(consumer/'main_darwin.go').read_text()
    main=replace(main,'\tif os.Args[4] != "off" && os.Args[4] != "on" { return errors.New("display clock off|on") }','\tif os.Args[4] != "on" { return errors.New("phase control requires observer on") }\n\tif err := phaseConfigure(); err != nil { return err }')
    main=replace(main,'\t\tbase := time.Now().Add(120 * time.Millisecond)','\t\tvar phaseBridges []map[string]any')
    main=replace(main,'\t\t\ttarget := base.Add(time.Duration(i-1) * 120 * time.Millisecond)','''\t\t\tnativeTarget, err := phaseTarget(i)
            if err != nil { return err }
            bridgeBefore := time.Now()
            bridgeHost := desktop.PresentationHostTime()
            bridgeAfter := time.Now()
            target := bridgeBefore.Add(time.Duration((nativeTarget-bridgeHost)*1e9))
            phaseBridges = append(phaseBridges,map[string]any{"sequence":i,"target":nativeTarget,"host":bridgeHost,"go_bracket_ns":bridgeAfter.Sub(bridgeBefore).Nanoseconds()})
            out["phase_bridges"] = phaseBridges''')
    main=replace(main,'\t\t\twake := time.Now()','\t\t\tif err := phaseWait(nativeTarget); err != nil { return err }\n\t\t\twake := time.Now()')
    main=replace(main,'\tout["after_close"] = metrics.Snapshot()','\tout["after_close"] = metrics.Snapshot()\n\tcontrol, controlErr := phaseSnapshot()\n\tout["pending_trace"] = control\n\te = errors.Join(e, controlErr)')
    main=replace(main,'time.Sleep(50 * time.Millisecond)','time.Sleep(80 * time.Millisecond)')
    (consumer/'main_darwin.go').write_text(main)
    core=Path('/private/tmp/vellum-joined-backend-core')
    native=(core/'system/driver/desktop/presentation_darwin.m').read_text()
    native=replace(native,'#include "presentation_darwin.h"','#include "presentation_darwin.h"\n#include '+json.dumps(str(consumer/'control.h')))
    native=replace(native,' id<CAMetalDrawable> drawable=[super nextDrawable];',' uint64_t pcAttempt=pc_acquire_begin(before.token,before.tag,before.generation,CACurrentMediaTime());\n id<CAMetalDrawable> drawable=[super nextDrawable];\n double pcAcquired=CACurrentMediaTime();\n if(drawable)pc_acquired(pcAttempt,drawable.drawableID,pcAcquired);else pc_unavailable(pcAttempt,pcAcquired);\n if(drawable)[drawable addPresentedHandler:^(id<MTLDrawable> x){pc_presented(pcAttempt,x.presentedTime);}];')
    native=replace(native,' double acquired=CACurrentMediaTime();',' double acquired=pcAcquired;')
    (out/'presentation_darwin.m').write_text(native)
    go=(core/'system/driver/desktop/presentation_darwin.go').read_text()
    go=replace(go,'#include "presentation_darwin.h"','#include "presentation_darwin.h"\n#include '+json.dumps(str(consumer/'control.h')))
    go+='\n// Private diagnostic compiler overlay only; never part of the SDK.\nfunc VellumDiagnosticReceipt(tag uint64, at float64) { C.pc_receipt(C.uint64_t(tag), C.double(at)) }\n'
    (out/'presentation_darwin.go').write_text(go)
    overlay=json.loads((out/'overlay.json').read_text());measurement_key=next(k for k in overlay['Replace'] if k.endswith('/desktop_measurement.go'));measurement=Path(overlay['Replace'][measurement_key]).read_text();measurement=replace(measurement,'\tr.measured.receiptFinishedAt = time.Now()','\tr.measured.receiptFinishedAt = time.Now()\n\tdesktop.VellumDiagnosticReceipt(r.tag, at)');(out/'desktop_measurement.go').write_text(measurement);overlay['Replace'][measurement_key]=str(out/'desktop_measurement.go');overlay['Replace'].update({str(core/'system/driver/desktop'/n):str(out/n) for n in ['presentation_darwin.m','presentation_darwin.go']});(out/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
    native=(repo/'tools/macos-calibration/native.m').read_text()
    native=replace(native,'#include "observe.h"','#include "observe.h"\n#include "control.h"\n#include "display_clock.h"')
    native=replace(native,' id<CAMetalDrawable>d=[_metal nextDrawable];',' uint64_t attempt=pc_acquire_begin(1,i,_sequence,cal_clock());\n id<CAMetalDrawable>d=[_metal nextDrawable];double pcAcquired=cal_clock();if(d){pc_acquired(attempt,d.drawableID,pcAcquired);dispatch_group_enter(_callbacks);[d addPresentedHandler:^(id<MTLDrawable>x){pc_presented(attempt,x.presentedTime);dispatch_group_leave(self.callbacks);}];}else pc_unavailable(attempt,cal_clock());')
    native=replace(native,' NSApplication *app=',' if(!pc_configure()||!dc_configure(1))return 2;\n NSApplication *app=')
    native=replace(native,'  double base=cal_clock()+.12;NSMutableArray *lateness=[NSMutableArray new];','  if(!dc_start())refuse("display clock start");NSMutableArray *lateness=[NSMutableArray new];')
    native=replace(native,'   double deadline=base+(n-1)*.12;while(cal_clock()<deadline){double left=deadline-cal_clock();if(left>0)usleep((useconds_t)fmin(left*1e6,1000));}','   double deadline=pc_target(n);if(!deadline||!pc_wait(deadline))refuse("phase anchor/wake");')
    native=replace(native,'atomic_store(&records[i].receipt,cal_clock());[NSApp postEvent:event atStart:NO];','double receipt=cal_clock();atomic_store(&records[i].receipt,receipt);pc_receipt(i,receipt);[NSApp postEvent:event atStart:NO];')
    native=replace(native,'  raw=cal_observe(2);','  usleep(80000);if(!dc_stop())refuse("display clock stop");char *clockRaw=dc_snapshot();char *pendingRaw=pc_snapshot();if(!clockRaw||!pendingRaw)refuse("trace snapshot");NSDictionary *clockDoc=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:clockRaw length:strlen(clockRaw)] options:0 error:nil];NSDictionary *pendingDoc=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:pendingRaw length:strlen(pendingRaw)] options:0 error:nil];free(clockRaw);free(pendingRaw);\n  raw=cal_observe(2);')
    native=replace(native,'@"kind":@"native",@"match":match,','@"kind":@"native",@"display_clock_trace":clockDoc,@"pending_trace":pendingDoc,@"match":match,')
    native=replace(native,'atomic_store(&records[i].acquired,cal_clock());','atomic_store(&records[i].acquired,pcAcquired);')
    native=replace(native,'static Entry records[512];','static _Atomic(double) scheduled[512],completed[512],gpu_start[512],gpu_end[512],present_request[512],commit_start[512],commit_end[512];static _Atomic(int) status[512];static _Atomic(uint64_t) drawable_id[512];\nstatic Entry records[512];')
    native=replace(native,'id<MTLCommandBuffer>cmd=[_queue commandBuffer];','id<MTLCommandBuffer>cmd=[_queue commandBuffer];atomic_store(&drawable_id[i],d.drawableID);dispatch_group_enter(_callbacks);dispatch_group_enter(_callbacks);[cmd addScheduledHandler:^(id<MTLCommandBuffer>x){(void)x;atomic_store(&scheduled[i],cal_clock());dispatch_group_leave(self.callbacks);}];[cmd addCompletedHandler:^(id<MTLCommandBuffer>x){atomic_store(&gpu_start[i],x.GPUStartTime);atomic_store(&gpu_end[i],x.GPUEndTime);atomic_store(&status[i],(int)x.status);atomic_store(&completed[i],cal_clock());dispatch_group_leave(self.callbacks);}];')
    native=replace(native,'[cmd presentDrawable:d];[cmd commit];','atomic_store(&present_request[i],cal_clock());[cmd presentDrawable:d];atomic_store(&commit_start[i],cal_clock());[cmd commit];atomic_store(&commit_end[i],cal_clock());')
    native=replace(native,'@"accepted":@(accepted)}','@"accepted":@(accepted),@"drawable_id":@(atomic_load(&drawable_id[i])),@"present_request":@(atomic_load(&present_request[i])),@"commit_start":@(atomic_load(&commit_start[i])),@"commit_end":@(atomic_load(&commit_end[i])),@"scheduled_observed":@(atomic_load(&scheduled[i])),@"complete_observed":@(atomic_load(&completed[i])),@"gpu_start":@(atomic_load(&gpu_start[i])),@"gpu_end":@(atomic_load(&gpu_end[i])),@"status":@(atomic_load(&status[i]))}')
    (out/'native.m').write_text(native)
    (out/'reused-build-identity.json').write_text(json.dumps(identity,indent=2)+'\n')
    return out

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--prior',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);a=p.parse_args();print(prepare(Path(__file__).resolve().parents[3],a.prior,a.artifacts))
