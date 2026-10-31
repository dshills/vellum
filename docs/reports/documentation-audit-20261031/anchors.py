import json,re,urllib.parse,collections
from pathlib import Path
x=json.load(open('/private/tmp/vellum-doc-audit-scan.json'));root=Path.cwd();bad=[];missing_scripts=[]
for d in x['documents']:
 if d['category']!='maintained':continue
 p=root/d['path'];s=p.read_text();clean=re.sub(r'```.*?```','',s,flags=re.S)
 for target in re.findall(r'\[[^\]]*\]\((<[^>]+>|[^\s)]+)(?:\s+[^)]*)?\)',clean):
  target=target.strip('<>');u=urllib.parse.urlsplit(target)
  if u.scheme or not u.fragment:continue
  dest=p.parent/urllib.parse.unquote(u.path) if u.path else p
  if not dest.is_file() or dest.suffix!='.md':continue
  t=dest.read_text();counts=collections.Counter();anchors=set(re.findall(r'(?:id|name)=[\"\x27]([^\"\x27]+)',t));t=re.sub(r'```.*?```','',t,flags=re.S)
  for heading in re.findall(r'^#{1,6}\s+(.+?)\s*#*$',t,re.M):
   heading=re.sub(r'<[^>]+>','',heading);heading=re.sub(r'\[([^\]]*)\]\([^)]*\)',r'\1',heading);slug=re.sub(r'[^\w\- ]','',heading.lower()).replace(' ','-');n=counts[slug];counts[slug]+=1;anchors.add(slug+('-'+str(n) if n else ''))
  if urllib.parse.unquote(u.fragment) not in anchors:bad.append({'file':d['path'],'target':target})
 for ref in set(re.findall(r'(?<![\w/-])(scripts/[A-Za-z0-9_./-]+\.(?:py|sh))',s)):
  if not (root/ref).exists():missing_scripts.append({'file':d['path'],'path':ref})
print(json.dumps({'unresolved_anchor_candidates':bad,'missing_repository_script_references':missing_scripts},indent=2))
Path('/private/tmp/vellum-doc-audit-anchor-results.json').write_text(json.dumps({'unresolved_anchor_candidates':bad,'missing_repository_script_references':missing_scripts},indent=2)+'\n')
