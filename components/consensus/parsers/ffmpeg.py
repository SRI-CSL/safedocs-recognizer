import subprocess
import shlex
from parsers.cfg_utils import callgrind, no_randomize_va


def ffmpeg(filename: str):
    report = {}
    result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} ffmpeg -v error -i {filename} -f null -"), capture_output=True)
    report['stdout'] = "-v errors\n" + result.stdout.decode('utf-8', errors='backslashreplace')
    report['stderr'] = "-v errors\n" + result.stderr.decode('utf-8', errors='backslashreplace')
    report['status'] = 'valid' if result.returncode == 0 else 'rejected'
    stdout_lines = result.stdout.decode('utf-8', errors='backslashreplace').split('\n')
    stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
    for line in stdout_lines:
        report['status'] = 'rejected'
        break
    for line in stderr_lines:
        report['status'] = 'rejected'
        break

    if report['status'] == 'valid':
        result = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} ffmpeg -v warning -i {filename} -f null -"), capture_output=True)
        report['stdout'] = "-v warnings\n" + report['stdout'] + result.stdout.decode('utf-8', errors='backslashreplace')
        report['stderr'] = "-v warnings\n" + report['stderr'] + result.stderr.decode('utf-8', errors='backslashreplace')
        stdout_lines = result.stdout.decode('utf-8', errors='backslashreplace').split('\n')
        stderr_lines = result.stderr.decode('utf-8', errors='backslashreplace').split('\n')
        for line in stdout_lines:
            report['status'] = 'valid-warnings'
            break
        for line in stderr_lines:
            report['status'] = 'valid-warnings'
            break
    
    metadata = subprocess.run(shlex.split(f"{no_randomize_va} {callgrind} ffprobe {filename}"), capture_output=True)
    report['stdout'] = report['stdout'] + "\n===ffprobe===\n"
    report['stdout'] = report['stdout'] + metadata.stdout.decode('utf-8', errors='backslashreplace')

    return report
