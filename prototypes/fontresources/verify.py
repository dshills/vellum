from pathlib import Path
import json

def require(condition,message="font evidence validation failed"):
 if not condition:raise RuntimeError(message)
r=Path(__file__).resolve().parent
lazy=json.loads((r/'evidence/font-lazy-outline.json').read_text())
eager=json.loads((r/'evidence/font-eager-outline.json').read_text())['samples'][0]
require(len(lazy['samples'])==3)
for s in lazy['samples']:
 require(s['coverage']==eager['coverage'])
 require(s['mixed_rune_bounds']==eager['mixed_rune_bounds'])
 require(s['live']['HeapAlloc'] < eager['live']['HeapAlloc']*0.5)
 require(len(s['coverage'][-1]['decoded_bitmaps'])>0)
 require(s['after_release']['HeapAlloc'] < 10*1024*1024)
 require(s['snapshots']['eager_fallbacks']==0)
 require(s['snapshots']['count']<=32 and s['snapshots']['bytes']<=2*1024*1024*1024)

for sample in lazy["samples"]:
 if sample["live"]["HeapAlloc"]>=40*1024*1024:raise RuntimeError("isolated corpus exceeds 40 MiB")
print('Eager/lazy corpus, mixed geometry and decoded emoji pixels match; targeted heap reduction verified.')
