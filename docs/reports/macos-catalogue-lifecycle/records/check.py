from pathlib import Path
import importlib.util,json,os,signal,sys
sys.dont_write_bytecode=True
Q=Path(__file__).resolve().parent;A=Q.parent;T=Path('/private/tmp/vellum-macos-native-results-97aee13');P=A/'macos-acquisition-diagnostics-4ce6ba4'
s=importlib.util.spec_from_file_location('frame_guard',A/'macos-sdk-stage0079-abe67c6-cache-v2/run.py');h=importlib.util.module_from_spec(s);s.loader.exec_module(h)
sdk=Path(json.loads((P/'plan.json').read_bytes())['sdk']);name,kind=sys.argv[1:]
assert not(Q/name).exists() and kind in ['gallery','focused','vet','build']
overlay=Q/'overlay.json'
if not overlay.exists():
 d=json.loads((A/'macos-media-catalogue-db98a6b/overlay-fixed.json').read_bytes())
 for source in list(d['Replace']):
  rel=Path(source).relative_to(sdk);candidate=T/rel
  if candidate.is_file():d['Replace'][source]=str(candidate)
  elif str(candidate).endswith('.go'):d['Replace'][source]=''
 for candidate in (T/'drivers/cogent/cmd/gallery').glob('*.go'):
  source=sdk/candidate.relative_to(T)
  if not source.is_file()or h.sha(source)!=h.sha(candidate):d['Replace'][str(source)]=str(candidate)
 overlay.write_text(json.dumps(d,indent=2)+'\n')
paths=[Path(p)for p in json.loads(overlay.read_bytes())['Replace'].values()if p]+[overlay,Path(__file__),Path(h.__file__),h.GUARD,h.GO,T/'scripts/child_process.py'];pins={str(p):h.sha(p)for p in paths}
base=['-p=1','-overlay='+str(overlay),'-tags=offscreen']
if kind=='focused':args=['test','-v','-race','-count=1',*base,'-timeout=3m','-run','^TestCatalogue','./cmd/gallery']
elif kind=='gallery':args=['test','-v','-race','-count=1',*base,'-timeout=3m','./cmd/gallery']
elif kind=='vet':args=['vet',*base,'./cmd/gallery']
elif kind=='build':args=['build','-p=1','-overlay='+str(overlay),'-o',str(Q/'native-gallery'),'./cmd/gallery']
env={'PATH':str(h.GO.parent)+':/usr/bin:/bin:/usr/sbin:/sbin','HOME':os.environ['HOME'],'GOPATH':'/Users/dshills/Development/Go','GOMODCACHE':'/Users/dshills/Development/Go/pkg/mod','GOWORK':'off','GOENV':'off','GOTOOLCHAIN':'local','GOFLAGS':'-mod=readonly','GOMAXPROCS':'2','GOPROXY':'off','GOSUMDB':'off','CGO_ENABLED':'1','CGO_LDFLAGS_ALLOW':'^-Wl,-rpath,@executable_path/native$','PYTHONDONTWRITEBYTECODE':'1'}
profile=h.PROFILE+'(deny file-write* (subpath '+json.dumps(str(sdk))+'))'
def interrupted(signum,frame):raise InterruptedError('frame check signal '+str(signum))
previous=signal.signal(signal.SIGTERM,interrupted)
try:r=h.g.execute(['/usr/bin/sandbox-exec','-p',profile,'/usr/bin/env','-i',*[k+'='+v for k,v in env.items()],str(h.GO),*args],sdk/'drivers/cogent',Q/name,lambda:False,lambda:all(h.sha(p)==v for p,v in pins.items()),timeout=540 if kind=='driver'else 300,limit=4<<20)
finally:signal.signal(signal.SIGTERM,previous)
r['source_pins']=pins;r['native_executed']=False
if kind=='build'and r['exit_code']==0:r['binary']={'path':str(Q/'native-gallery'),'sha256':h.sha(Q/'native-gallery')}
(Q/(name+'.json')).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps({k:r.get(k)for k in ['exit_code','error','seconds','logs_complete','pins_stable','owned']}));sys.exit(0 if r['exit_code']==0 and r['error']is None and r['logs_complete']and r['pins_stable']and r['owned']['group_absent']and not r['owned']['errors']else 1)
