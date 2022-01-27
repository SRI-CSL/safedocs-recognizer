import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} qpdf --json {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
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
