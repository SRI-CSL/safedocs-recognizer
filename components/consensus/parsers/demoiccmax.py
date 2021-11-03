import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va

import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def iccdumpprofile(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} /builds/src/dist/Tools/IccDumpProfile/iccDumpProfile -v {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'rejected'
    if result.returncode == 0:
        report['status'] = 'valid'
    
    gprof = subprocess.run(shlex.split("gprof /builds/src/dist/Tools/IccDumpProfile/iccDumpProfile gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')

    # print(json.dumps(report, indent=2))
    return report
