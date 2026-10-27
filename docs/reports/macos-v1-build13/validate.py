"""Read-only validation of this named local publication set; executes no app."""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path, PurePosixPath
import plistlib
import re
import sys
import tarfile
import zipfile

if not __debug__:
    raise SystemExit('Packet validation requires assertions; rerun without -O.')

sys.dont_write_bytecode = True
T = Path(__file__).resolve().parent
R = T.parents[2]
sys.path.insert(0, str(R / 'scripts'))
spec = importlib.util.spec_from_file_location('sdk_check', R / 'scripts/check-sdk.py')
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)
def digest_stream(stream):
    digest = hashlib.sha256()
    for chunk in iter(lambda: stream.read(1 << 20), b''):
        digest.update(chunk)
    return digest.hexdigest()


def sha(path):
    with path.open('rb') as stream:
        return digest_stream(stream)

j = lambda p: json.loads(p.read_text())


def validate(delivery):
    packet = j(T / 'PUBLICATION_SET.json')
    final = j(T / 'final-package.json')
    native = j(T / 'native-result.json')
    head = '64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a'
    assert packet['source_commit'] == final['source_commit'] == native['source_commit'] == head
    assert packet['publication_authorized'] is False and packet['uploaded'] is False
    assert final['publication_authorized'] is False and final['uploaded'] is False
    assert packet['signing'] == 'ad-hoc (-)' and packet['notarized'] is False
    assert packet['destination']['tag_commit'] == head
    assert packet['destination']['tag'] == 'macos-v1.0.0-build13'
    assert packet['destination']['new_tag_proposed'] is True
    assert packet['destination']['modify_existing_release'] is False
    assert j(delivery / 'PUBLICATION_SET.json') == packet
    assert j(delivery / 'manifest.json') == final
    assert sha(T / packet['destination']['release_body']) == packet['destination']['release_body_sha256']
    assert (delivery / 'README.md').read_bytes() == (T / 'DELIVERY_README.md').read_bytes()
    assert len(packet['assets']) == 7 and len(final['apps']) == 6
    assert len({a['file'] for a in packet['assets']}) == 7
    expected = {a['file']: a['sha256'] for a in packet['assets']}
    assert (delivery / 'SHA256SUMS').read_bytes() == (T / 'SHA256SUMS').read_bytes()
    assert sha(T / 'SHA256SUMS') == packet['checksum_file']['sha256']
    rows = [s.split('  ', 1) for s in (T / 'SHA256SUMS').read_text().splitlines()]
    assert len(rows) == 7 and {name: digest for digest, name in rows} == expected
    for a in packet['assets']:
        assert Path(a['file']).name == a['file'] and a['source_commit'] == head
        assert sha(delivery / a['file']) == a['sha256']
    sdk = next(a for a in packet['assets'] if a['kind'] == 'sdk')
    assert sdk['file'] == final['sdk_archive'] and sdk['sha256'] == final['sdk_sha256']
    verified = checker.verify_archive(delivery / sdk['file'], sdk['sha256'])
    manifest = verified['manifest']
    assert manifest['source'] == {'commit': head, 'clean': True}
    assert manifest['private_build'] is False and len(manifest['files']) == final['verified_files'] == 8861
    assert manifest['files']['drivers/cogent/build/upstream/core/renderwindow.go'] == '3960374de3a8b104f116865632f645f31ac999320151b1da22a4ba41a7b1412b'
    by_id = {a['id']: a for a in packet['assets'] if a['kind'] == 'app'}
    assert set(by_id) == {a['id'] for a in final['apps']} == {'notes', 'widgets', 'shell', 'forma', 'luma', 'pulse'}
    for a in final['apps']:
        assert by_id[a['id']]['file'] == a['archive']
        assert by_id[a['id']]['sha256'] == a['archive_sha256']
        assert a['source_commit'] == head and a['build'] == '13'
        assert a['runtime_result'].startswith('pass:') and native['apps'][a['id']]['result'] == 'pass'
        with zipfile.ZipFile(delivery / a['archive']) as z:
            names = z.namelist()
            assert len(names) == len(set(names)) and len(names) < 1000
            assert sum(info.file_size for info in z.infolist()) <= 256 << 20
            for info in z.infolist():
                path = PurePosixPath(info.filename)
                assert not path.is_absolute() and '..' not in path.parts and '\\' not in info.filename
                assert str(path) == info.filename.rstrip('/')
                assert info.file_size <= 128 << 20
            root = a['name'] + '.app/Contents/'
            raw = z.read(root + 'Resources/APP_MANIFEST.json')
            assert hashlib.sha256(raw).hexdigest() == a['manifest_sha256']
            am = json.loads(raw)
            assert am['release_authorization']['source'] == {'commit': head, 'clean': True}
            assert am['private_build'] is False and am['native_distribution_audit'] == 'approved'
            assert am['version'] == '1.0.0' and am['build_version'] == '13'
            assert am['signed_identity'] == '-' and am['notarized'] is False
            assert am['target_runtime'] == 'unverified'
            assert am['source_binary_sha256'] == a['source_binary_sha256']
            assert hashlib.sha256(z.read(root + 'MacOS/' + a['name'])).hexdigest() == a['packaged_binary_sha256']
            info = plistlib.loads(z.read(root + 'Info.plist'))
            assert info['CFBundleVersion'] == '13' and info['CFBundleShortVersionString'] == '1.0.0'
            assert info['CFBundleIdentifier'] == a['identifier']
            prefix = root + 'Resources/THIRD_PARTY_NOTICES/'
            raw = z.read(prefix + 'INVENTORY.json')
            assert hashlib.sha256(raw).hexdigest() == a['notice_inventory_sha256']
            ni = json.loads(raw)
            assert ni['artifact_sha256'] == a['source_binary_sha256']
            assert len(ni['files']) == a['notice_files'] == (419 if a['id'] in ('forma', 'luma', 'pulse') else 418)
            assert {n[len(prefix):] for n in names if n.startswith(prefix) and not n.endswith('/')} == {'INVENTORY.json', *ni['files']}
            for name, identity in ni['files'].items():
                import app_notices
                app_notices.notice_name(name)
                raw = z.read(prefix + name)
                assert len(raw) == identity['bytes'] and hashlib.sha256(raw).hexdigest() == identity['sha256']
            if a['id'] in ('forma', 'luma', 'pulse'):
                assert any(f['sha256'] == 'dc588a7ba9130043b67c2380cc03cc6644ed3557f6d5e8589b96acba1c64e343' for f in ni['files'].values())
                assert root + 'Resources/' + info['CFBundleIconFile'] in names
    assert all(c['exit_code'] == 0 for c in final['checks'] + final['post_launch_checks'])
    assert native['owned_exact_process_paths_absent'] and not native['exit_codes_observed']
    assert not native['user_notebook_accessed'] and not native['global_preferences_changed']
    ci = j(T / 'reused-ci.json')
    assert len(ci) == 3 and all(c['head_sha'] == head and c['status'] == 'completed' and c['conclusion'] == 'success' for c in ci)
    assert sha(R / 'release/macos-v1-distribution-audit.json') == final['audit_record_sha256'] == packet['historical_audit_sha256']
    assert sha(R / 'release/macos-preview-audit.json') == 'c86448a24e5ece322e648cb1bb9011c5b2c24e74152fa553b232a3ab53a114e1'
    audit = j(T / 'private-inputs.json')
    assert sha(delivery / audit['file']) == audit['sha256']
    evidence = j(T / 'evidence-index.json')
    assert sha(T / evidence['archive']) == evidence['sha256']
    identities = {f['file']: f for f in evidence['files']}
    assert len(identities) == len(evidence['files'])
    seen = set()
    admitted_bytes = 0
    with tarfile.open(T / evidence['archive'], 'r|gz') as tar:
        for member in tar:
            assert member.isfile() and member.name in identities and member.name not in seen
            path = PurePosixPath(member.name)
            assert not path.is_absolute() and '..' not in path.parts and '\\' not in member.name
            assert str(path) == member.name
            admitted_bytes += member.size
            assert member.size <= 128 << 20 and admitted_bytes <= 512 << 20
            identity = identities[member.name]
            assert member.size == identity['bytes']
            with tar.extractfile(member) as stream:
                assert digest_stream(stream) == identity['sha256']
            seen.add(member.name)
    assert seen == set(identities)
    assert 'session/native/luma-control-close-after.ax.txt' in identities
    for file in list(T.glob('*.md')) + [R / 'README.md', R / 'docs/macos-v1.md', R / 'docs/README.md', R / 'specs/macos-v1-release.md']:
        for link in re.findall(r'!?\[[^\]]*\]\(([^)]+)\)', file.read_text()):
            if ':' in link or link.startswith('#'):
                continue
            assert (file.parent / link.split('#', 1)[0]).exists(), (file, link)
    return {'pass': True, 'source_commit': head, 'payload_archives': 7, 'apps': 6, 'sdk_files': 8861, 'evidence_files': len(seen), 'archive_metadata_notice_and_checksum_bindings': 'pass', 'historical_audits_unchanged': True, 'ci_exact_source': 'three successful workflows', 'publication_authorized': False, 'uploaded': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--delivery', required=True, type=Path)
    print(json.dumps(validate(parser.parse_args().delivery), indent=2))
