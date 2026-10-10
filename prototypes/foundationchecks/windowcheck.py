"""Derive a minimized-idle fixture from the unchanged nativeinput fixture."""
from pathlib import Path
import hashlib
import subprocess

root = Path(__file__).resolve().parent
source = root.parent / 'nativeinput/main.go'
# Lock the input rather than silently adapting to another fixture revision.
expected = '82066219801384154327283fc6f6c84e7e8e0786f770636ca97606d7bee10b96'
if hashlib.sha256(source.read_bytes()).hexdigest() != expected:
    raise RuntimeError('nativeinput fixture changed; review this derivative')
s = source.read_text()
s = s.replace('flag.Parse()', 'minimized := flag.Bool("minimized", false, "minimize this owned fixture before idle sampling")\n flag.Parse()', 1)
s = s.replace('if *idle > 0 {', '''if *minimized {
 time.AfterFunc(time.Second, func() {
  f.body.Scene.RenderWindow().SystemWindow.Minimize()
 })
}
if *idle > 0 {''', 1)
s = s.replace('start := time.Now()\n\t\t\t\tcpu := cpuSeconds()', '''core.TheApp.RunOnMain(func() {
  iconified := f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).Glw.GetAttrib(glfw.Iconified)
  fmt.Printf("idle_window_state iconified=%d requested_minimized=%t\\n", iconified, *minimized)
 })
 start := time.Now()
 cpu := cpuSeconds()''',1)
owned = root / 'build/windowcheck-src/main.go'
owned.parent.mkdir(parents=True, exist_ok=True)
owned.write_text(s)
subprocess.run(['gofmt','-w',str(owned)],check=True)
subprocess.run(['go','build','-modfile=checks.mod','-o','build/windowcheck',str(owned)],cwd=root,check=True)
