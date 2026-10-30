"""Reassess the recorded rand notice disposition if its pinned evidence changes.

This is an applicability-evidence guard, not a vulnerability or binary scanner.
"""
import hashlib
import json
from pathlib import Path
import re
import unittest
from unittest import mock

import native_gpu

ROOT = Path(__file__).resolve().parents[1]
INVENTORY = 'release/native-components/wgpu-darwin-arm64/INVENTORY.json'
INVENTORY_SHA256 = '443e1760808a72d4fee88330c58271e18eb87eea633d30ef180d819c6254f55e'
ARTIFACT_SHA256 = '786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883'
DECLARATION = 'NOTICES/itertools-0.13.0/Cargo.toml'
DECLARATION_SHA256 = '49abb2101a0dd9cb137df206454b6620d04929a4975921fab6682ba834435620'


def applicability(inventory, declaration):
    if (inventory.get('artifact_sha256') != ARTIFACT_SHA256
            or any(c.get('name') == 'rand' for c in inventory['components'])
            or inventory['files'].get(DECLARATION) != DECLARATION_SHA256
            or hashlib.sha256(declaration).hexdigest() != DECLARATION_SHA256):
        raise ValueError('rand applicability evidence changed; reassess alert 4')
    # Hash-bound original declaration: rand is under dev-dependencies only.
    if re.findall(rb'\[([^\]\r\n]*dependencies\.rand)\]', declaration) != [b'dev-dependencies.rand']:
        raise ValueError('rand dependency scope changed; reassess alert 4')
    return 'notice dev-dependency reference; no rand in recorded compiled component inventory'


class Disposition(unittest.TestCase):
    def evidence(self):
        raw = native_gpu.read_regular(ROOT/INVENTORY, 8 << 20)
        self.assertEqual(hashlib.sha256(raw).hexdigest(), INVENTORY_SHA256,
                         'native inventory changed: reassess the recorded rand disposition')
        return json.loads(raw), native_gpu.read_regular((ROOT/INVENTORY).parent/DECLARATION, 1 << 20)

    def test_pinned_notice_inventory_and_original_bytes(self):
        inventory, declaration = self.evidence()
        # Existing verifier checks all original notice bytes, paths and coverage.
        record = native_gpu.component_notices(ROOT, ARTIFACT_SHA256)
        self.assertIsNotNone(record)
        self.assertEqual(record['components'], 82)
        self.assertIn('notice dev-dependency', applicability(inventory, declaration))

    def test_runtime_component_or_artifact_change_requires_reassessment(self):
        for change in ('rand', 'artifact'):
            with self.subTest(change=change):
                inventory, declaration = self.evidence()
                if change == 'rand': inventory['components'].append({'name':'rand'})
                else: inventory['artifact_sha256']='0'*64
                with self.assertRaisesRegex(ValueError,'reassess'): applicability(inventory, declaration)

    def test_notice_change_requires_reassessment_without_rewriting_it(self):
        inventory, declaration = self.evidence()
        with self.assertRaisesRegex(ValueError,'reassess'): applicability(inventory, declaration+b'\n')
        # Even explicitly reblessing bytes cannot silently broaden dependency scope.
        with mock.patch.object(hashlib, 'sha256') as digest:
            digest.return_value.hexdigest.return_value=DECLARATION_SHA256
            with self.assertRaisesRegex(ValueError,'scope changed'):
                applicability(inventory, declaration.replace(b'[dev-dependencies.rand]', b'[dependencies.rand]'))


if __name__ == '__main__': unittest.main()
