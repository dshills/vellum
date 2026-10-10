"""Package the owned executable and its native accessibility library."""
import plistlib
from pathlib import Path
import shutil

root = Path(__file__).resolve().parent
bundle = root / "build/Vellum-NativeInput.app/Contents"
executable_dir = bundle / "MacOS"
(executable_dir / "native").mkdir(parents=True, exist_ok=True)
executable = executable_dir / "vellum-nativeinput"
# Replace the inode to avoid stale macOS code-signature caching on rebuilds.
executable.unlink(missing_ok=True)
shutil.copy2(root / "build/vellum-nativeinput", executable)
shutil.copy2(root / "build/native/libaccesskit.dylib", executable_dir / "native")
with (bundle / "Info.plist").open("wb") as file:
    plistlib.dump({
        "CFBundleIdentifier": "org.vellum.nativeinput",
        "CFBundleName": "Vellum Native Input",
        "CFBundleExecutable": "vellum-nativeinput",
        "CFBundlePackageType": "APPL",
        "NSPrincipalClass": "NSApplication",
        "NSHighResolutionCapable": True,
    }, file)
