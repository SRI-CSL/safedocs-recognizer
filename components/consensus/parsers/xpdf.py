import subprocess
import shlex
import os
import json
from parsers.cfg_utils import callgrind, no_randomize_va


def pdftoppm(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdftoppm {filename} -"), capture_output=True)
    report = {}
    # don't store binary
    report['stdout'] = '' #result.stdout.decode('utf-8')
    report['stderr'] = result.stderr.decode('utf-8')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8').split('\n')
    for line in stderr_lines:
        if 'Syntax Error: Couldn\'t find a font for' in line:
            continue
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdftoppm gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report


def pdftotext(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdftotext {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8')
    report['stderr'] = result.stderr.decode('utf-8')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8').split('\n')
    for line in stderr_lines:
        if 'Syntax Error: Couldn\'t find a font for' in line:
            continue
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdftotext gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report


def pdffonts(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} pdffonts {filename}"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'Syntax Error: Couldn\'t find a font for' in line:
            continue
        if 'Syntax Error' in line:
            report['status'] = 'rejected'

    gprof = subprocess.run(shlex.split("gprof /usr/bin/pdffonts gmon.out"), capture_output=True)
    if gprof.returncode == 0:
        report['gprof'] = gprof.stdout.decode('utf-8')
    
    # print(json.dumps(report, indent=2))
    return report