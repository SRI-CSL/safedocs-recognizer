import subprocess
import shlex
import os
import json
from parsers.cfg_utils import python_prof, python_prof_to_callgrind, no_randomize_va


def run(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {python_prof} /kaitai-nitf/kaitai-nitf.py {filename}"), capture_output=True)
    subprocess.run(shlex.split(python_prof_to_callgrind), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    report['status'] = 'valid'
    if result.returncode != 0:
        report['status'] = 'rejected'
    
    report['stdout'] = result.stdout.decode('utf-8', errors="backslashreplace")
    report['stderr'] = result.stderr.decode('utf-8', errors="backslashreplace")

    # print(json.dumps(report, indent=2))
    return report
