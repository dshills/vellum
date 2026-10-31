import subprocess,re,json,collections,urllib.parse
from pathlib import Path
root=Path.cwd();paths=[p for p in subprocess.check_output(['git','ls-files','--cached','--others','--exclude-standard'],text=True).splitlines() if p.lower().endswith(('.md','.mdx','.rst'))]
issues=[];flags=[];categories=collections.Counter();stats=[]
for name in paths:
 p=root/name;txt=p.read_text(errors='replace')
 category=('third-party' if name.startswith('release/native-components/') or '/material-notices/' in name or name.endswith('/NOTICE.md') else 'pinned-native' if name.startswith('drivers/cogent/native/') else 'archive' if (name.startswith(('docs/reports/','docs/reviews/','docs/research/','docs/decisions/','prototypes/','release/notes/','docs/security/')) and name != 'docs/security/README.md') or (name.startswith('tools/macos-') and not name.endswith('README.md')) else 'maintained')
 categories[category]+=1
 clean=re.sub(r'```.*?```','',txt,flags=re.S)
 links=re.findall(r'!?\[[^\]]*\]\((<[^>]+>|[^\s)]+)(?:\s+[^)]*)?\)',clean)
 links += re.findall(r'^\s*\[[^\]]+\]:\s*(\S+)',clean,re.M)
 for target in links:
  target=target.strip('<>');u=urllib.parse.urlsplit(target)
  if u.scheme or target.startswith('//'):continue
  pathname=urllib.parse.unquote(u.path)
  if not pathname:continue
  dest=(p.parent/pathname)
  if not dest.exists():issues.append(dict(file=name,target=target,category=category))
 if category=='maintained':
  lines=[]
  for i,line in enumerate(txt.splitlines(),1):
   if re.search(r'current|latest|next|defer|blocked|alpha|v1\.1\.[012]|build2[34]|pending|not yet|TODO|FIXME',line,re.I):lines.append(f'{i}: {line}')
  flags.append(dict(file=name,lines=lines))
 stats.append(dict(path=name,category=category,lines=len(txt.splitlines())))
Path('/private/tmp/vellum-doc-audit-scan.json').write_text(json.dumps(dict(base='0b11d3b092455b9967bad891fa42ae4ba5b6f8ae',counts=dict(categories),documents=stats,missing_targets=issues),indent=2)+'\n')
Path('/private/tmp/vellum-doc-audit-flags.txt').write_text('\n\n'.join(f['file']+'\n'+'\n'.join(f['lines']) for f in flags))
print('COUNTS',dict(categories));print('BROKEN',len(issues),dict(collections.Counter(i['category'] for i in issues)))
# Detailed results are written to JSON; avoid dumping archival captures.
