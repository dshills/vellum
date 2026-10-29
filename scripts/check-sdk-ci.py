#!/usr/bin/env python3
"""Export and verify this clean checkout's private SDK without opening windows."""
import argparse
import hashlib
import json
from pathlib import Path
import sys

sys.dont_write_bytecode = True
from child_process import check_child_output, run_child
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]


def check(output):
    output = output.resolve()
    if output == ROOT or ROOT in output.parents:
        raise ValueError('SDK CI output must be outside the source checkout')
    env = child_environment()
    revision = check_child_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, env=env,
                                  text=True, timeout=20).strip()
    if check_child_output(['git', 'status', '--porcelain', '--untracked-files=normal'],
                          cwd=ROOT, env=env, text=True, timeout=20):
        raise ValueError('SDK CI requires clean committed sources')
    output.mkdir()  # exclusive; a failed attempt retains diagnostics
    archive = output / 'Vellum SDK.tar.gz'
    with (output / 'export.stdout').open('wb') as stdout, (output / 'export.stderr').open('wb') as stderr:
        run_child([sys.executable, '-B', ROOT / 'scripts/release-sdk.py', '--output', archive],
                  cwd=ROOT, env=env, check=True, timeout=600,
                  stdout=stdout, stderr=stderr, log_limit_bytes=32 << 20)
    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    # This digest binds our own fresh producer output; it is not third-party
    # authenticity certification or distribution-audit approval.
    with (output / 'check.stdout').open('wb') as stdout, (output / 'check.stderr').open('wb') as stderr:
        run_child([sys.executable, '-B', ROOT / 'scripts/check-sdk.py', '--archive', archive,
                   '--sha256', digest], cwd=ROOT, env=env, check=True, timeout=1200,
                  stdout=stdout, stderr=stderr, log_limit_bytes=32 << 20)
    result = json.loads((output / 'check.stdout').read_text(encoding='utf-8'))
    if (result.get('event') != 'vellum.sdk.relocated'
            or result.get('archive_sha256') != digest
            or result.get('source') != {'commit': revision, 'clean': True}
            or result.get('path_contains_spaces') is not True
            or result.get('native_runtime') != 'not executed'
            or not result.get('document_replacements_within_sdk')):
        raise ValueError('SDK CI verifier result does not match the committed offscreen producer')
    # Only the successful result is exposed as the CI report. Partial logs remain
    # useful after failure, but cannot be mistaken for completed verification.
    (output / 'relocation.json').write_text(json.dumps(result, indent=2) + '\n', encoding='utf-8')
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', required=True, type=Path, help='new directory outside the checkout')
    args = parser.parse_args()
    print(json.dumps(check(args.output), indent=2))


if __name__ == '__main__':
    main()
