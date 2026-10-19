"""Original geometric cursor drawings; no upstream SVG path data is used."""
from pathlib import Path
import hashlib,json

def path(d,fill='none',width=12):
    return shape('path','d="'+d+'"',fill,width)
def shape(tag,attrs,fill='#fff',width=12):
    common=f' {attrs} fill="{fill}" stroke-linecap="round" stroke-linejoin="round"'
    return f'<{tag}{common} stroke="#fff" stroke-width="{width+12}"/>' + f'<{tag}{common} stroke="#000" stroke-width="{width}"/>'
def circle(x,y,r,fill='#fff',width=10):return shape('circle',f'cx="{x}" cy="{y}" r="{r}"',fill,width)
def rect(x,y,w,h,fill='#fff',rx=0,width=10):return shape('rect',f'x="{x}" y="{y}" width="{w}" height="{h}" rx="{rx}"',fill,width)
def plain(d,color='#000',width=12):return f'<path d="{d}" fill="none" stroke="{color}" stroke-width="{width}" stroke-linecap="round" stroke-linejoin="round"/>'
arrow=path('M40 32 L40 204 L88 158 L122 222 L152 206 L116 146 L184 146 Z','#000',4)
hourglass=path('M76 40 H180 V64 L146 108 V148 L180 192 V216 H76 V192 L110 148 V108 L76 64 Z','#000',8)+path('M98 64 H158 L128 96 Z','#fff',0)+path('M98 192 L128 160 L158 192 Z','#fff',0)
cross=path('M128 40 V216 M40 128 H216',width=12)
ibeam=path('M128 52 V204 M92 52 H164 M92 204 H164',width=14)
def rotate(body,angle):return f'<g transform="rotate({angle} 128 128)">{body}</g>'
def resize(double=True,col=False):
    if col:
        return path('M48 88 V168 M208 88 V168 M72 128 H184 M72 128 L104 96 M72 128 L104 160 M184 128 L152 96 M184 128 L152 160',width=12)
    d='M44 128 H212 M212 128 L172 88 M212 128 L172 168'
    if double:d+=' M44 128 L84 88 M44 128 L84 168'
    return path(d,width=14)
badge=lambda fill='#fff':circle(184,184,42,fill,10)
assets={};hotspots={}
def add(name,body,hot=(128,128)):
    assert name not in assets
    assets[name]=('<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">\n'
                  '<!-- Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT. Original Vellum artwork. -->\n'
                  +body+'\n</svg>\n').encode()
    hotspots[name]=list(hot)
add('arrow',arrow,(40,32))
add('context-menu',arrow+rect(140,138,88,86,rx=8)+plain('M158 158 H210 M158 180 H210 M158 202 H198',width=8),(40,32))
add('help',circle(128,128,80)+plain('M104 100 C104 68 156 68 156 100 C156 116 128 120 128 142',width=16)+circle(128,170,6,'#000',0))
hand=path('M88 210 L54 140 Q46 124 59 117 Q72 112 84 133 L94 148 V54 Q94 34 110 34 Q126 34 126 54 V108 L138 102 Q152 100 158 116 Q170 108 182 120 Q194 116 206 132 L214 156 L204 204 Q202 220 184 222 H118 Q100 222 88 210 Z','#000',8)
add('pointer',hand+plain('M126 126 V156 M158 128 V160 M186 140 V166','#fff',6),(110,34))
add('progress',arrow+badge('#ff0')+plain('M184 158 V184 L202 196',width=10),(40,32))
add('wait',hourglass)
add('cell',rect(64,64,128,128)+plain('M128 64 V192 M64 128 H192',width=12))
add('crosshair',cross)
add('text',ibeam)
add('vertical-text',rotate(ibeam,90))
add('alias',arrow+badge()+plain('M160 200 V182 Q160 166 182 166 H204 M204 166 L190 154 M204 166 L190 180',width=10),(40,32))
add('copy',arrow+badge()+plain('M162 184 H206 M184 162 V206','#0f0',12),(40,32))
add('move',path('M128 40 V216 M40 128 H216 M128 40 L102 70 M128 40 L154 70 M128 216 L102 186 M128 216 L154 186 M40 128 L70 102 M40 128 L70 154 M216 128 L186 102 M216 128 L186 154',width=12))
add('not-allowed',arrow+badge('#f00')+plain('M160 160 L208 208','#fff',12),(40,32))
openhand=path('M70 170 L38 112 Q34 100 46 96 Q56 94 65 112 L82 136 L68 64 Q66 50 80 48 Q94 46 98 62 L110 112 L106 42 Q106 28 120 28 Q134 28 134 44 L138 112 L144 52 Q146 38 160 40 Q174 42 170 58 L162 122 L178 86 Q184 72 198 78 Q210 84 204 98 L188 156 L180 204 Q176 222 156 222 H108 Q84 222 76 198 Z','#000',8)
add('grab',openhand+plain('M110 152 Q140 138 166 152','#fff',7),(128,144))
fist=path('M68 188 L54 128 Q50 110 66 104 L88 102 V86 Q88 72 104 72 L124 76 Q132 62 148 72 L162 82 Q178 72 188 90 L200 114 L196 192 Q194 214 172 220 H108 Q80 216 68 188 Z','#000',8)
add('grabbing',fist+plain('M88 112 L112 148 M126 100 L142 140 M162 104 L172 140','#fff',7),(128,144))
add('resize-col',resize(col=True));add('resize-row',rotate(resize(col=True),90))
for name,angle in [('up',-90),('right',0),('down',90),('left',180),('n',-90),('e',0),('s',90),('w',180),('ne',-45),('nw',-135),('se',45),('sw',135)]:add('resize-'+name,rotate(resize(False),angle))
for name,angle in [('ew',0),('ns',90),('nesw',-45),('nwse',45)]:add('resize-'+name,rotate(resize(),angle))
for name,sign in [('zoom-in',True),('zoom-out',False)]:
    body=plain('M168 152 L218 212',width=22)+circle(118,108,58)+plain('M90 108 H146'+(' M118 80 V136'if sign else''),width=12)
    add(name,body,(118,108))
add('screenshot-selection',path('M128 128 H216 V216 H128 Z',width=10)+path('M96 128 H152 M128 96 V152',width=10))
add('screenshot-window',rect(48,56,160,144,rx=12)+plain('M48 92 H208',width=10)+rect(90,118,92,58,rx=6)+circle(136,147,14,'#000',4))
add('poof',arrow+badge()+plain('M166 166 L202 202 M166 202 L202 166','#f00',12),(40,32))
assert len(assets) == 39
if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path, help="New output directory (never overwrites)")
    destination = parser.parse_args().output
    destination.mkdir(exist_ok=False)
    for name, body in sorted(assets.items()):
        (destination / (name + ".svg")).write_bytes(body)
    manifest = {"schema": 1, "scope": "Original geometric Vellum cursor artwork", "license": "MIT", "extent": 256,
                "files": {name + ".svg": hashlib.sha256(body).hexdigest() for name, body in sorted(assets.items())},
                "hotspots_256": hotspots, "public_cursor_names": sorted(assets)}
    (destination / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
