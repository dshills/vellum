"""Build an owned exact-buffer derivative of the unchanged nativeinput fixture."""
from pathlib import Path
import hashlib,shutil,subprocess
root=Path(__file__).resolve().parent
source=root.parent/'nativeinput/main.go'
expected='82066219801384154327283fc6f6c84e7e8e0786f770636ca97606d7bee10b96'
if hashlib.sha256(source.read_bytes()).hexdigest()!=expected:raise RuntimeError('Baseline fixture changed; review derivative')
s=source.read_text()
assert s.count('lines.NewLines()')==1
s=s.replace('lines.NewLines()','lines.NewLinesExact()')
owned=root/'build/fixture-src/main.go';owned.parent.mkdir(parents=True,exist_ok=True);owned.write_text(s)
(root/'build/native').mkdir(exist_ok=True)
shutil.copy2(root.parent/'nativeinput/build/native/libaccesskit.dylib',root/'build/native/libaccesskit.dylib')
subprocess.run(['gofmt','-w',str(owned)],check=True)
subprocess.run(['go','build','-modfile=contracts.mod','-o','build/contracts-fixture',str(owned)],cwd=root,check=True)
