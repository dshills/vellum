#!/usr/bin/env python3
"""Permit optional cache upload only with five minutes of job headroom."""
import sys
import time


def may_save(start, minutes, now):
    # Two-minute upload cap plus three minutes for cancellation and job cleanup.
    elapsed = now - start
    return 0 <= elapsed < minutes * 60 - 300


def main():
    try:
        start, minutes = map(int, sys.argv[1:])
        allowed = start > 0 and minutes > 5 and may_save(start, minutes, time.time())
    except (ValueError, TypeError):
        allowed = False
    print('save=' + str(allowed).lower())


if __name__ == '__main__':
    main()
