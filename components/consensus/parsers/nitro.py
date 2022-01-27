import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} nitf_tester {filename}"), capture_output=True)
    report = {}
    report['status'] = 'valid'
    if result.returncode != 0:
        report['status'] = 'rejected'
    
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')

    # print(json.dumps(report, indent=2))
    return report
