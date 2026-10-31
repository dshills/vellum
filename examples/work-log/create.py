#!/usr/bin/env python3
"""Create a Work Log consumer from a verified exported SDK, never the checkout."""
import argparse
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'scripts'))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--sdk', type=Path, required=True)
    parser.add_argument('--sha256', required=True)
    parser.add_argument('--output', type=Path, required=True, help='new directory containing SDK and application source')
    args = parser.parse_args()
    spec = importlib.util.spec_from_file_location('work_log_sdk_checker', ROOT/'scripts/check-sdk.py')
    checker = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(checker)
    verified = checker.verify_archive(args.sdk, args.sha256)
    output = args.output.expanduser().resolve()
    if output.is_relative_to(ROOT):
        raise ValueError('output must be outside the checkout')
    output.mkdir(mode=0o700)  # exclusive; never replace existing source or notebooks
    sdk = checker.extract_verified(args.sdk, output/'sdk', verified)
    app = output/'Work Log'
    subprocess.run([sys.executable, '-B', str(sdk/'scripts/new-app.py'), str(app),
                    '--module=example.com/vellum-work-log', '--template=document'], check=True)
    main_go = app/'main.go'
    text = main_go.read_text(encoding="utf-8")
    old = 'Title: "Vellum Document"'
    if text.count(old) != 1:
        raise ValueError('SDK document template changed; review the Work Log customization')
    main_go.write_text(text.replace(old, 'Title: "Vellum Work Log"'), encoding='utf-8')
    # Storage identity derives from the application module and remains stable.
    (output/'CONSUMER.json').write_text(json.dumps(dict(
        sdk_sha256=args.sha256, sdk_release=verified['manifest']['release']['version'],
        source=verified['manifest']['source'], application_module='example.com/vellum-work-log',
        customization='window title only; generated application-owned document recipe',
        status='generated; not yet built or runtime qualified'), indent=2)+'\n')
    print(app)


if __name__ == '__main__':
    main()
