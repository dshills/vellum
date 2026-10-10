#!/usr/bin/env python3
"""Build and verify a relocatable foundation bundle; Developer ID is opt-in."""
from pathlib import Path
import argparse,ctypes,hashlib,json,os,plistlib,shutil,subprocess,tempfile

def run(argv,**kw):
 return subprocess.run(argv,check=True,text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,**kw).stdout

def publish_exclusive(source,destination):
 # macOS sys/stdio.h: RENAME_EXCL = 0x00000004. Atomically preserve any
 # destination created by another publisher after the initial validation.
 libc=ctypes.CDLL('/usr/lib/libSystem.B.dylib',use_errno=True)
 rename=libc.renamex_np;rename.argtypes=[ctypes.c_char_p,ctypes.c_char_p,ctypes.c_uint];rename.restype=ctypes.c_int
 if rename(os.fsencode(source),os.fsencode(destination),4)!=0:
  error=ctypes.get_errno();raise OSError(error,os.strerror(error),str(destination))

def package(args):
 binary=args.binary.resolve();library=args.library.resolve();output=args.output.parent.resolve()/args.output.name
 if not binary.is_file() or not library.is_file():raise ValueError('binary and native library are required')
 if output.suffix!='.app' or output.exists() or output.is_symlink():raise ValueError('output must be a new .app path')
 if args.manifest.resolve().is_relative_to(output):raise ValueError('manifest must be outside the sealed bundle')
 if args.manifest.resolve() in (binary,library):raise ValueError('manifest must not overwrite an input artifact')
 if args.notarize and args.identity=='-':raise ValueError('notarization requires a Developer ID identity')
 output.parent.mkdir(parents=True,exist_ok=True)
 with tempfile.TemporaryDirectory(prefix='.vellum-package-',dir=output.parent) as td:
  bundle=Path(td)/output.name;macos=bundle/'Contents/MacOS';native=macos/'native';native.mkdir(parents=True)
  executable=macos/'VellumFixture';lib=native/'libaccesskit.dylib'
  shutil.copy2(binary,executable);shutil.copy2(library,lib)
  executable.chmod(0o755)
  plist={'CFBundleExecutable':'VellumFixture','CFBundleIdentifier':'org.vellum.foundation','CFBundleName':'Vellum Foundation','CFBundlePackageType':'APPL','CFBundleShortVersionString':'0.1.0','CFBundleVersion':'1','NSHighResolutionCapable':True}
  (bundle/'Contents/Info.plist').write_bytes(plistlib.dumps(plist))
  dependencies={}
  executable_archs=run(['lipo','-archs',str(executable)]).split()
  library_archs=run(['lipo','-archs',str(lib)]).split()
  if not executable_archs or not set(executable_archs).issubset(library_archs):raise ValueError('native library lacks an executable architecture')
  for path,architectures in [(executable,executable_archs),(lib,library_archs)]:
   dependencies[path.name]={}
   for arch in architectures:
    listing=run(['otool','-arch',arch,'-L',str(path)])
    dependencies[path.name][arch]=listing
    for line in listing.splitlines()[1:]:
     dependency=line.strip().split(' (',1)[0]
     if dependency.startswith(('/System/Library/','/usr/lib/')) and '..' not in Path(dependency).parts:continue
     if dependency in ['@rpath/libaccesskit.dylib','@executable_path/native/libaccesskit.dylib'] and path==executable:continue
     if dependency=='@rpath/libaccesskit.dylib' and path==lib:continue
     raise ValueError(f'Unbundled runtime dependency ({arch}): {dependency}')
    if path==executable:
     paths=run(['otool','-arch',arch,'-l',str(executable)])
     rpaths=[]
     for block in paths.split('Load command '):
      if 'cmd LC_RPATH\n' in block:
       for line in block.splitlines():
        if line.strip().startswith('path '):rpaths.append(line.strip().split(' (',1)[0][5:])
     if rpaths!=['@executable_path/native']:raise ValueError(f'Only the bundled-library rpath is permitted ({arch}): {rpaths}')
  sign=['codesign','--force','--sign',args.identity]
  if args.identity!='-':sign+=['--options','runtime','--timestamp']
  for path in [lib,executable,bundle]:run(sign+[str(path)])
  strict=run(['codesign','--verify','--deep','--strict','--verbose=2',str(bundle)])
  notarized=False
  if args.notarize:
   archive=Path(td)/'submission.zip';run(['ditto','-c','-k','--keepParent',str(bundle),str(archive)])
   result=json.loads(run(['xcrun','notarytool','submit',str(archive),'--keychain-profile',args.notarize,'--wait','--output-format','json']))
   if result.get('status')!='Accepted':raise ValueError('notarization was not accepted')
   run(['xcrun','stapler','staple',str(bundle)])
   run(['xcrun','stapler','validate',str(bundle)])
   run(['spctl','--assess','--type','execute','--verbose=2',str(bundle)])
   run(['codesign','--verify','--deep','--strict',str(bundle)])
   notarized=True
  manifest={'schema':1,'identity':args.identity,'developer_id_notarized':notarized,'strict_bundle_verified':True,'dependencies':dependencies,'scope':'packaging/signature verification; clean-machine/runtime validation is a separate check','files':{str(p.relative_to(bundle)):hashlib.sha256(p.read_bytes()).hexdigest() for p in bundle.rglob('*') if p.is_file()}}
  # Rename within the same filesystem only after every requested check succeeds.
  manifest_path=args.manifest.resolve();manifest_path.parent.mkdir(parents=True,exist_ok=True)
  pending_manifest=None
  try:
   with tempfile.NamedTemporaryFile(mode='w',prefix='.vellum-manifest-',dir=manifest_path.parent,delete=False) as mf:
    pending_manifest=Path(mf.name);json.dump(manifest,mf,indent=2);mf.write('\n');mf.flush()
   publish_exclusive(bundle,output)
   try:pending_manifest.replace(manifest_path)
   except BaseException:shutil.rmtree(output);raise
  finally:
   if pending_manifest is not None:pending_manifest.unlink(missing_ok=True)

 return manifest

if __name__=='__main__':
 parser=argparse.ArgumentParser(description=__doc__)
 parser.add_argument('--binary',type=Path,required=True);parser.add_argument('--library',type=Path,required=True)
 parser.add_argument('--output',type=Path,required=True);parser.add_argument('--identity',default='-')
 parser.add_argument('--notarize',metavar='KEYCHAIN_PROFILE',help='submit with an already provisioned notarytool keychain profile')
 parser.add_argument('--manifest',type=Path,required=True)
 args=parser.parse_args()
 result=package(args)
 print(f'Created verified bundle: {args.output}')
