import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str, hexdigest: str):
    # result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} /caradoc/caradoc stats {filename}"), capture_output=True)
    result = subprocess.run(shlex.split(f"/caradoc/caradoc stats {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
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
