#!/usr/bin/env python3
"""Report recorded Documents evaluation coverage; never invent sessions."""
import argparse
from datetime import date
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CHECKS = {'edit', 'find-replace', 'undo-redo', 'save', 'quit-restart'}


def evaluate(record, candidate, today):
    if record['candidate'] != candidate:
        raise ValueError('session ledger belongs to a different candidate')
    days, blockers = set(), set()
    for session in record['sessions']:
        observed = date.fromisoformat(session['date'])
        if observed > today:
            raise ValueError('future sessions cannot count as observed use')
        if session['kind'] not in {'operator', 'automated'}:
            raise ValueError('session kind must be operator or automated')
        checks = set(session['checks'])
        if not checks <= CHECKS:
            raise ValueError('unknown workflow check')
        issues = session['open_blockers']
        if not isinstance(issues, list) or any(not isinstance(i, str) or not i.strip() for i in issues):
            raise ValueError('open blockers must be nonempty issue IDs')
        blockers.update(issues)
        if session['kind'] == 'operator' and checks == CHECKS and not issues:
            days.add(observed.isoformat())
    status = 'blocked' if blockers else ('recorded' if len(days) >= 3 else 'pending')
    return dict(candidate=candidate, status=status, qualifying_operator_days=sorted(days),
                required_days=3, open_blockers=sorted(blockers),
                evidence='operator-reported ledger, not independently observed by this checker')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--require-recorded', action='store_true', help='fail unless three qualifying days are recorded without open blockers')
    args = parser.parse_args()
    candidate = json.loads((ROOT/'release/private-macos-candidate.json').read_text())['id']
    record = json.loads((ROOT/'docs/daily-use/sessions.json').read_text())
    result = evaluate(record, candidate, date.today())
    print(json.dumps(result, indent=2))
    if args.require_recorded and result['status'] != 'recorded':
        parser.exit(1, 'Multi-day operator evaluation is not yet recorded.\n')


if __name__ == '__main__':
    main()
