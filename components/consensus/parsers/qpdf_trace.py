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
    os.environ["TC_SCOPE"] = "qpdf"
    os.environ["TC_FILENAME"] = "tracing.txt"
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} qpdf --json {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stdout'] = report['stdout'] + "\n=====TRACING====="
    with open ("tracing.txt", "r") as tracing_file:
        report['stdout'] = report['stdout'] + tracing_file.read()
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'rejected'
    stdout_lines = result.stdout.decode('utf-8', errors='backslashreplace').split('\n')
    if len(stdout_lines) > 2:
        report['status'] = 'valid'
    
    gprof = subprocess.run(shlex.split("gprof /usr/local/bin/qpdf gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report
