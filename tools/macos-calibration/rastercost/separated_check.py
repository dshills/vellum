"""Validate bounded offscreen replay conditions, without release qualification."""
import statistics
CONDITIONS=('immediate','separated_120ms','immediate_pressure32m','separated_120ms_pressure32m')
def validate(doc):
    def need(ok,msg):
        if not ok:raise ValueError(msg)
    need(doc['kind']=='offscreen-separated-raster' and doc['cpu']==2 and doc['target']==[1519,1266] and doc['pressure_bytes']==32<<20,'configuration')
    need(doc['cache_flush_proven'] is False and doc['release_qualified'] is False and doc['gate_changed'] is False,'claims')
    need(type(doc['elapsed_ns']) is int and 0<doc['elapsed_ns']<10_000_000_000,'deadline')
    samples=doc['samples'];need(len(samples)==32,'complete samples');grouped={name:[] for name in CONDITIONS}
    for i,row in enumerate(samples):
        round=i//4;name=CONDITIONS[(round+i%4)%4]
        need(row['Round']==round and row['Condition']==name,'balanced order')
        need(all(type(row[k]) is int for k in ['ReplayNS','GapNS','PressureNS']) and 0<row['ReplayNS']<doc['elapsed_ns'] and row['GapNS']>=row['PressureNS']>=0,'duration')
        need(not name.startswith('separated') or row['GapNS']>=120_000_000,'separation')
        need(row['PressureNS']>0 if 'pressure' in name else row['PressureNS']==0,'pressure scope')
        color={'R':255 if round%2==0 else 0,'G':0 if round%2==0 else 255,'B':0,'A':255};need(row['Color']==color,'alternating color')
        grouped[name].append(row)
    return {name:{'count':len(rows),'min_replay_ns':min(r['ReplayNS'] for r in rows),'median_replay_ns':statistics.median(r['ReplayNS'] for r in rows),'max_replay_ns':max(r['ReplayNS'] for r in rows),'min_gap_ns':min(r['GapNS'] for r in rows),'max_gap_ns':max(r['GapNS'] for r in rows),'pressure_ns':[r['PressureNS'] for r in rows]} for name,rows in grouped.items()}
