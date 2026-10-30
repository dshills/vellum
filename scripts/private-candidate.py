#!/usr/bin/env python3
"""Check the private candidate index, recorded evidence and generated guide blocks."""
import argparse
import importlib.util
import json
from pathlib import Path, PurePosixPath
import re
import sys
sys.dont_write_bytecode=True
import native_gpu
from macos_deployment import version
ROOT=Path(__file__).resolve().parents[1]
INDEX='release/private-macos-candidate.json'
SLOTS={'facts':'docs/mac-document-workflows-handoff.md','commands':'docs/mac-desktop-resumption.md'}


def read(path):
    return native_gpu.strict_json(native_gpu.read_regular(path,1<<20))


def relative(value):
    if (not isinstance(value,str) or not re.fullmatch(r'[A-Za-z0-9 ._/-]+',value)
            or PurePosixPath(value).is_absolute() or any(x in ('','.','..') for x in value.split('/'))):
        raise ValueError('candidate path must be a safe relative path')
    return value


def load(root):
    r=read(root/INDEX)
    if not isinstance(r,dict) or type(r.get('schema')) is not int or r['schema']!=1:raise ValueError('invalid candidate schema')
    for key,length in [('source_commit',40),('sdk_sha256',64),('manifest_sha256',64),('installation_sha256',64),('preflight_sha256',64)]:
        if not isinstance(r.get(key),str) or not re.fullmatch('[0-9a-f]{'+str(length)+'}',r[key]):raise ValueError('invalid candidate '+key)
    if (not isinstance(r.get('build_version'),str) or not re.fullmatch('[0-9]{1,6}',r['build_version'])
            or r.get('id')!='build'+r['build_version']+'-'+r['source_commit'][:12]):raise ValueError('candidate build/source identity mismatch')
    for key in ['installation','packet','packaging_report','verification_report','linkage_report']:relative(r.get(key))
    if r['installation']!='Applications/Vellum/'+r['id']:raise ValueError('candidate installation identity mismatch')
    for key in ['packaging_report','verification_report','linkage_report']:
        if not r[key].startswith('docs/reports/') or not r[key].endswith('.json'):raise ValueError('candidate report outside source evidence area')
    version(r.get('minimum_macos'));version(r.get('inspection_host'))
    if version(r['minimum_macos'])>version(r['inspection_host']):raise ValueError('candidate exceeds inspection host')
    if (r.get('signing')!='ad-hoc' or r.get('notarized') is not False or r.get('native_runtime')!='not executed' or r.get('desktop_checks')!='deferred'):
        raise ValueError('candidate qualification requires explicit reassessment')
    return r


def evidence(root,r):
    p=read(root/r['packaging_report']);v=read(root/r['verification_report']);l=read(root/r['linkage_report'])
    for key in ['source_commit','build_version','sdk_sha256','manifest_sha256','installation_sha256','preflight_sha256','signing','notarized','native_runtime','desktop_checks']:
        if p.get(key)!=r[key]:raise ValueError('packaging report differs: '+key)
    for key in ['source_commit','build_version','sdk_sha256','manifest_sha256']:
        if v.get(key)!=r[key]:raise ValueError('verification report differs: '+key)
    if (p.get('installation')!='~/'+r['installation'] or v.get('installation')!='~/'+r['installation']
            or v.get('apps_verified')!=5 or v.get('native_runtime')!='not executed'):
        raise ValueError('verification/installation scope differs')
    if (l.get('source_commit')!=r['source_commit'] or l.get('build_version')!=r['build_version']
            or l.get('installation_sha256')!=r['installation_sha256'] or l.get('native_runtime')!='not executed'
            or l.get('host_version')!=r['inspection_host']):raise ValueError('linkage report identity differs')
    apps=l.get('apps',[])
    if (len(apps)!=5 or len({a['name'] for a in apps})!=5 or set(p.get('apps',[]))!={a['name'] for a in apps}
            or max((version(a['deployment']['minimum']) for a in apps))!=version(r['minimum_macos'])):
        raise ValueError('candidate app/minimum coverage differs')


def render(r,slot):
    if slot=='facts':
        rows=[('Candidate',r['id']),('Packaged source',r['source_commit']),('SDK SHA256',r['sdk_sha256']),
              ('Bundle manifest SHA256',r['manifest_sha256']),('Installation receipt SHA256',r['installation_sha256']),
              ('Preflight SHA256',r['preflight_sha256']),('Installed apps','~/'+r['installation']+'/apps'),
              ('Minimum macOS',r['minimum_macos']),('Inspected host',r['inspection_host']),('Desktop checks',r['desktop_checks'])]
        return '| Record | Identity |\n| --- | --- |\n'+''.join('| '+k+' | `'+v+'` |\n' for k,v in rows)
    return '''```sh
python3 -B scripts/verify-private-macos.py \\
  --bundle "$HOME/'''+r['packet']+'''" \\
  --installation "$HOME/'''+r['installation']+'''" \\
  --manifest-sha256 '''+r['manifest_sha256']+'''
python3 -B scripts/check-private-macos-linkage.py \\
  --installation "$HOME/'''+r['installation']+'''" \\
  --installation-sha256 '''+r['installation_sha256']+'''
```
'''


def documents(root,r,write=False):
    pending=[]
    for slot,name in SLOTS.items():
        path=root/name;text=path.read_text();start='<!-- vellum-candidate:'+slot+':start -->';end='<!-- vellum-candidate:'+slot+':end -->'
        if text.count(start)!=1 or text.count(end)!=1 or text.index(start)>text.index(end):raise ValueError('missing/duplicate candidate guide markers: '+name)
        before,tail=text.split(start);_,after=tail.split(end)
        expected=before+start+'\n'+render(r,slot)+end+after
        if not write and expected!=text:raise ValueError('stale candidate guide block: '+name)
        pending.append((path,expected))
    if write:
        for path,text in pending:path.write_text(text)


def local(r):
    spec=importlib.util.spec_from_file_location('candidate_verifier',ROOT/'scripts/verify-private-macos.py')
    v=importlib.util.module_from_spec(spec);spec.loader.exec_module(v)
    installation=Path.home()/r['installation'];packet=Path.home()/r['packet']
    v.no_links(installation)
    v.installer.regular(installation/'INSTALLATION.json',1<<20)
    if v.installer.digest(installation/'INSTALLATION.json')!=r['installation_sha256']:raise ValueError('local installation receipt differs')
    result=v.verify(packet,installation,r['manifest_sha256'])
    if result['source_commit']!=r['source_commit'] or result['sdk_sha256']!=r['sdk_sha256']:raise ValueError('local candidate differs')
    return result


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--write-docs',action='store_true',help='regenerate only the two marked candidate blocks')
    parser.add_argument('--verify-local',action='store_true',help='also verify recorded archives/installation without launch')
    args=parser.parse_args()
    try:
        r=load(ROOT);evidence(ROOT,r);documents(ROOT,r,args.write_docs)
        actual=local(r) if args.verify_local else None
    except (ValueError,OSError,KeyError,TypeError) as error:parser.exit(1,str(error)+'\n')
    print(json.dumps(dict(candidate=r['id'],source_commit=r['source_commit'],checks='index/recorded evidence/guide consistency',local_verification=actual,native_runtime='not executed'),indent=2))

if __name__=='__main__':main()
