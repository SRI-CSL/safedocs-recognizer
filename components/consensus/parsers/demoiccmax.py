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


def iccdumpprofile(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} /builds/src/dist/Tools/IccDumpProfile/iccDumpProfile -v {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = doc_url
    report['MR_PARSER'] = parser
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

def iccapplyprofiles(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} /builds/src/dist/Tools/IccApplyProfiles/iccApplyProfiles cat_no_alpha.tif cat_with_icc.tif 0 0 0 0 0 {filename} 0"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'rejected'
    # if result.returncode == 0:
    #     report['status'] = 'valid'

    if os.path.exists('cat_with_icc.tif'):
        report['status'] = 'valid'
        os.remove('cat_with_icc.tif')

    
    gprof = subprocess.run(shlex.split("gprof /builds/src/dist/Tools/IccApplyProfiles/iccApplyProfiles gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')

    # print(json.dumps(report, indent=2))
    return report
