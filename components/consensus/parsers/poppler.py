import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def run(filename: str, hexdigest: str):
    pdftops(filename, hexdigest)


def pdftops(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdftops {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdftops gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report


def pdftoppm(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdftoppm {filename} -"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    # don't store binary
    report['stdout'] = '' #result.stdout.decode('utf-8')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdftoppm gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')

    # print(json.dumps(report, indent=2))
    return report


def pdftotext(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdftotext {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdftotext gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report


def pdffonts(filename: str, hexdigest: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdffonts {filename}"), capture_output=True)
    report = {}
    report['MR_DOC_URL'] = os.environ['MR_DOC_URL']
    report['MR_PARSER'] = os.environ['MR_PARSER']
    report['digest'] = hexdigest
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdffonts gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report