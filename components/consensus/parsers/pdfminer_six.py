import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str):
    # result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} dumppdf.py -a {filename}"), capture_output=True)
    result = subprocess.run(shlex.split(f"dumppdf.py -a {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'rejected'
    if result.returncode == 0:
        report['status'] = 'valid'

    # print(json.dumps(report, indent=2))
    return report
