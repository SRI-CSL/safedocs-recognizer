import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str, hexdigest: str):
    os.environ["TC_SCOPE"] = "qpdf"
    os.environ["TC_FILENAME"] = "tracing.txt"
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} qpdf --json {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
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
