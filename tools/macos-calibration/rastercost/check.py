"""Admit a complete bounded offscreen microbenchmark, never a release result."""
import re,math,statistics
CASES=('full_replay','image_commands','no_paint_paths','last_solid_fill','empty_replay')
def parse(text):
    if '\nPASS\n' not in text or 'FAIL' in text:raise ValueError('missing successful terminal')
    rows={name:[] for name in CASES}
    for line in text.splitlines():
        if not line.startswith('BenchmarkRasterReplayBreakdown/'):continue
        m=re.fullmatch(r'BenchmarkRasterReplayBreakdown/(\w+)-(\d+)\s+(\d+)\s+([0-9.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op',line)
        if not m:raise ValueError('malformed benchmark row')
        name,cpu,n,ns,b,a=m.groups();ns=float(ns)
        if name not in rows or cpu!='2' or int(n)<=0 or not math.isfinite(ns) or ns<=0 or int(b)!=0 or int(a)!=0:raise ValueError('invalid case/cpu/iterations/cost/allocation')
        rows[name].append(ns)
    if any(len(v)!=3 for v in rows.values()):raise ValueError('three observations per case required')
    return {k:{'ns_per_op':v,'min_ns':min(v),'median_ns':statistics.median(v),'max_ns':max(v),'bytes_per_op':0,'allocs_per_op':0} for k,v in rows.items()}
