"""Separate request wake/check intervals without relaxing existing acceptance."""
from pathlib import Path
import importlib.util
s=importlib.util.spec_from_file_location('segmented_timing_parent',Path(__file__).with_name('timing_check.py'))
parent=importlib.util.module_from_spec(s);s.loader.exec_module(parent)
base=parent.base

def phases(doc,complete=True):
    rows=doc['request_schedule_phases'];late=doc['lateness_ns']
    base.need(isinstance(rows,list) and len(rows)==len(late) and (len(rows)==248 if complete else 0<len(rows)<=248),'phase count')
    for i,r in enumerate(rows,1):
        base.need(all(type(r.get(k)) is int for k in ['Sequence','WakeLatenessNS','ForegroundCheckNS','PostCheckLatenessNS']),'phase integer fields')
        base.need(r['Sequence']==i and r['PostCheckLatenessNS']==late[i-1],'phase sequence/legacy lateness join')
        base.need(r['ForegroundCheckNS']>=0 and r['PostCheckLatenessNS']>=r['WakeLatenessNS']+r['ForegroundCheckNS'],'phase ordering')
    return {'requests':len(rows),'maximum_wake_lateness_ns':max(r['WakeLatenessNS'] for r in rows),
            'maximum_foreground_check_ns':max(r['ForegroundCheckNS'] for r in rows),
            'raw_segments':rows,'causality':'Observation only. Different maxima need not be the same request; no overhead calibration or timer repair inferred.'}

def validate(doc,match,mode):
    result=parent.validate(doc,match,mode)
    return {**result,'request_schedule_segments':phases(doc)}
