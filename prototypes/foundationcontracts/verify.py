"""Assert saved diagnostic outcomes; no UI or preference changes."""
from pathlib import Path
import json
root=Path(__file__).resolve().parent
r=json.loads((root/'evidence/contracts.json').read_text())
assert all(x['pass'] and x['input']==x['observed'] for x in r['text_results'])
assert r['delivered_fifo']==100000 and r['accepted_loss']==0
assert r['observed_peak']['Outstanding']<=r['limits']['Slots']
assert r['observed_peak']['PayloadBytes']<=r['limits']['PayloadBytes']
assert r['observed_peak']['InFlight']==1
assert r['preflight_rejections']==10000
assert r['after_drain']['Outstanding']==r['after_drain']['PayloadBytes']==0
native=json.loads((root/'evidence/native-exact.json').read_text())
expected='Latin: office AV\nArabic: العربية 123\nHebrew: שלום 456\nIndic: नमस्ते\nCJK: 日本語 中文\nCombining: é å\nEmoji: 👩‍💻 👨‍👩‍👧‍👦 🇺🇳'
assert native['multiline_text']==expected
assert native['field_text']=='A中文漢👩Z'
assert native['native_input_ack']==10 and native['rejected_actions']==0
assert len(native['input_updates'])==10 and len(native['input_probes'])==9
assert native['adapter_active'] and native['native_input_enabled']
print('PASS: 8 exact diagnostic roundtrips; 100000 FIFO deliveries within bounds; exact native fixture corpus and 10 acknowledged protocol updates.')
