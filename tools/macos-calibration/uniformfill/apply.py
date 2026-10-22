#!/usr/bin/env python3
"""Private offscreen candidate; frozen SDK remains read-only."""
from pathlib import Path
import hashlib,json,shutil
root=Path('/private/tmp/vellum-uniform-fill-core');sdk=Path('/private/tmp/vellum-doclinks-sdk-bd3cbc9/retained SDK/vellum-v0.1.0-alpha.2/drivers/cogent/build/upstream');code=Path(__file__).resolve().parent
p=root/'paint/pimage/pimage.go';original=(sdk/'paint/pimage/pimage.go').read_bytes();assert p.read_bytes()==original
s=original.decode();old='\t\t\tdraw.Draw(dest, pr.Rect, imagex.Unwrap(pr.Source), pr.SourcePos, pr.Op)';assert s.count(old)==1
s=s.replace(old,'\t\t\tsrc := imagex.Unwrap(pr.Source)\n\t\t\tif !fillOpaqueUniform(dest, pr.Rect, src, pr.SourcePos, pr.Op) {\n\t\t\t\tdraw.Draw(dest, pr.Rect, src, pr.SourcePos, pr.Op)\n\t\t\t}')
p.write_text(s)
for n in ['uniform_fill.go','uniform_fill_test.go']:shutil.copyfile(code/'testdata'/n,root/'paint/pimage'/n)
print('Private opaque uniform candidate only; no shipping/SDK/source/default or native change.')
