"""
 Copyright SRI International 2019-2022 All Rights Reserved.
 This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
"""

import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str):
    # result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} /caradoc/caradoc stats {filename}"), capture_output=True)
    result = subprocess.run(shlex.split(f"/caradoc/caradoc stats {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'PDF error' in line:
            report['status'] = 'rejected'
    stdout_lines = result.stdout.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stdout_lines:
        if "Not a PDF file" in line:
            report['status'] = 'rejected'

    # print(json.dumps(report, indent=2))
    return report
