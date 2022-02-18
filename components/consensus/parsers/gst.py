import subprocess
import shlex
from parsers.cfg_utils import callgrind, no_randomize_va


def gst(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} gst-launch-1.0 -v filesrc location = {filename} ! decodebin ! fakesink"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid'
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stderr_lines:
        if 'error' in line.lower():
            report['status'] = 'rejected'

    return report
