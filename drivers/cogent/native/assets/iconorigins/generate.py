"""Original neutral Vellum pictograms, authored from geometric primitives.
No upstream SVG bodies, fonts, wordmarks, or external resources are read.
"""
from pathlib import Path
import hashlib, json

def path(d, rule=None):
    return '<path d="'+d+'"'+(' fill-rule="'+rule+'"' if rule else '')+'/>'
def rect(x,y,w,h):
    return f'<rect x="{x}" y="{y}" width="{w}" height="{h}"/>'
def dot(x,y,r=2):
    return f'<circle cx="{x}" cy="{y}" r="{r}"/>'
def doc(body):
    return path('M5 2 H20 L27 9 V30 H5 Z M7 4 V28 H25 V11 H18 V4 Z M20 5 V9 H24 Z','evenodd')+body

def assets():
    math=doc(rect(10,14,12,2)+rect(15,9,2,12)+rect(10,24,12,2)+dot(11,20,1)+dot(21,20,1))
    build=doc(rect(10,13,4,4)+rect(19,22,4,4)+rect(10,22,4,4)+rect(11,17,2,6)+rect(13,19,8,2)+rect(20,19,2,4))
    bodies={
      'file-exe.svg':doc(path('M12 12 L22 19 L12 26 Z')),
      'file-markdown.svg':doc(rect(10,12,12,2)+rect(10,17,8,2)+path('M15 21 H18 V24 H21 L16.5 28 L12 24 H15 Z')),
      'git-hub.svg':path('M3 8 L16 2 L29 8 V27 L16 31 L3 27 Z M5 10 V25 L15 28 V14 Z M17 14 V28 L27 25 V10 Z M7 8 L16 12 L25 8 L16 4 Z','evenodd'),
      'git.svg':dot(9,5,3)+dot(23,9,3)+dot(9,27,3)+rect(8,7,2,18)+path('M10 21 V18 C19 18 22 15 22 10 H24 C24 17 19 20 10 21 Z'),
      'go.svg':doc(path('M9 13 H18 V10 L23 14 L18 18 V15 H9 Z M23 23 H14 V26 L9 22 L14 18 V21 H23 Z')),
      'json.svg':doc(path('M13 11 H11 V17 L9 19 L11 21 V26 H13 V20 L12 19 L13 18 Z M19 11 H21 V17 L23 19 L21 21 V26 H19 V20 L20 19 L19 18 Z')+dot(16,16,1)+dot(16,22,1)),
      'latex.svg':math, 'tex.svg':math, 'make.svg':build, 'makefile.svg':build,
      'radio_button_partial.svg':path('M16 3 A13 13 0 1 0 16 29 A13 13 0 1 0 16 3 Z M16 5 A11 11 0 1 1 16 27 A11 11 0 1 1 16 5 Z','evenodd')+path('M16 9 A7 7 0 0 0 16 23 Z'),
      'toggle_mid.svg':path('M11 7 H21 A9 9 0 0 1 21 25 H11 A9 9 0 0 1 11 7 Z M11 9 A7 7 0 0 0 11 23 H21 A7 7 0 0 0 21 9 Z','evenodd')+dot(16,16,4),
      'toml.svg':doc(dot(11,13,1)+rect(15,12,7,2)+dot(11,19,1)+rect(15,18,7,2)+dot(11,25,1)+rect(15,24,7,2)),
    }
    return {name: ('<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">\n'
        '<!-- Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT. Original neutral Vellum artwork. -->\n'
        +body+'\n</svg>\n').encode() for name,body in sorted(bodies.items())}

def manifest():
    return {'schema':1,'license':'MIT','extent':32,'generator_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
      'files':{n:hashlib.sha256(b).hexdigest() for n,b in assets().items()},
      'aliases':{'latex.svg':'tex.svg','make.svg':'makefile.svg'},
      'semantics':{'radio_button_partial.svg':'left-half inner disc; partial selection','toggle_mid.svg':'centered thumb; intermediate toggle'},
      'visual_policy':'Neutral replacements retain public file/export identities; no brand identity is represented.'}

if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--output',type=Path,required=True)
    out=p.parse_args().output;out.mkdir(exist_ok=False)
    for n,b in assets().items(): (out/n).write_bytes(b)
    (out/'manifest.json').write_text(json.dumps(manifest(),indent=2,sort_keys=True)+'\n')
