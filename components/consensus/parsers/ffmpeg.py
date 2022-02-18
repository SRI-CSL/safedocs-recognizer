import subprocess
import shlex
from parsers.cfg_utils import callgrind, no_randomize_va


def ffmpeg(filename: str):
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} ffmpeg -v error -i {filename} -f null -"), capture_output=True)
    report = {}
    report['stdout'] = result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid' if result.returncode == 0 else 'rejected'

    return report
