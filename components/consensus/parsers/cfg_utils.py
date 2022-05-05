"""
 Copyright SRI International 2019-2022 All Rights Reserved.
 This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
"""

import base64
import subprocess
import shlex
import os


# timeout 120 is what TA3 is using in eval3
# TODO rename callgrind
callgrind = "timeout 180"
# callgrind = "timeout 900 valgrind --tool=callgrind --callgrind-out-file=callgrind.out --log-file=valgrind.log"

python_prof = "python3.7 -m cProfile -o profile_data.pyprof"
python_prof_to_callgrind = "pyprof2calltree -i profile_data.pyprof -o callgrind.out"
# no_randomize_va = "setarch x86_64 --addr-no-randomize"
# requires --privileged
no_randomize_va = ""


def create_cfg_output(report):
    # if not os.path.exists("callgrind.out"):
    #     return
    if 'gprof' not in report:
        return
    with open('gprof_data.txt', 'w') as gprof_data:
        gprof_data.writelines(report['gprof'])
    
    # try:
    gprof2dot = subprocess.run(shlex.split(
        "gprof2dot -n0 -e0 gprof_data.txt -o gmon.dot"), capture_output=True)
    with open('gmon.dot') as f:
        report['cfg'] = f.read()
    with open('gprof_data.txt') as f:
        report['callgrind'] = f.read()
    # dot = subprocess.run(shlex.split("dot -o gmon.png -T png gmon.dot"), capture_output=True)
    # with open('gmon.png', 'rb') as f:
    #     data = f.read()
    #     encoded = base64.b64encode(data)
    #     encoded = encoded.decode("utf-8")
    #     report['cfg_image'] = f'{encoded}'

    # gprof2dot = subprocess.run(shlex.split(
    #     "gprof2dot -f callgrind callgrind.out -o callgrind.dot"), capture_output=True)
    # dot = subprocess.run(shlex.split("dot -o callgrind.png -T png callgrind.dot"), capture_output=True)
    # save callgrind.out, callgrind.dot, callgrind.png in report
    # with open('callgrind.out') as f:
    #     report['callgrind'] = f.read()
    # with open('callgrind.dot') as f:
    #     report['cfg'] = f.read()
    # with open('callgrind.png', 'rb') as f:
    #     data = f.read()
    #     encoded = base64.b64encode(data)
    #     encoded = encoded.decode("utf-8")
    #     report['cfg_image'] = f'{encoded}'
    # except:
    #     print('error generating cfg output')
    #     pass
