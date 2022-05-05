#!/usr/bin/env python3

"""
 Copyright SRI International 2019-2022 All Rights Reserved.
 This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
"""

import json
import os
import subprocess
import shlex
import png
import psycopg2
import base64
from mmap import ACCESS_READ, mmap
from contextlib import closing

# delete gmon.out
# delete **/*.gcda and gcno files
os.system("rm -rf gmon.out")
os.system("rm -rf /consensus/gmon.out")
os.system("rm -rf /builds/src/gmon.out")
os.system("rm -rf /builds/src/bitcov.png")

with open('coverage-src-summary.json') as f:
    data = json.load(f)
    for root, dir, files in os.walk('./'):
        for f in files:
            for source_info in data:
                source_obj = source_info[source_info.rfind('/')+1:source_info.rfind('.')]
                if f.endswith('.gcda') and (source_info.endswith(f[:-5]) or source_obj.endswith(f[:-5])):
                    path = root + '/' + f
                    # print('found ' + path + ' matching source file ' + source_info)
                    # exec 'gcov {path} -l {f[:-5]}' in new temp directory
                    # proc = subprocess.run(shlex.split(
                    #     f'gcov {path} -l {f[:-5]}'), capture_output=True)
                    proc = subprocess.run(shlex.split(
                        f'gcov -l {path}'), capture_output=True)
                    break

# ls_proc = subprocess.run(shlex.split('ls -l /builds/src/'))
# construct bitmap... iterate through *##* files and read in 2nd part of filename if it exists in sourcemap json file
offset = 0
data = {}
with open('/builds/src/coverage-src-summary.json') as f:
    data = json.load(f)
    # add start offset
    for source_info in data:
        if source_info == 'total':
            continue
        entry = data[source_info]
        entry['offset'] = offset
        offset += entry['lines']

bitmap = [1] * offset

# with open('/builds/src/coverage-src-summary.json') as f:
#     data = json.load(f)
with os.scandir('/builds/src/') as entries:
    for entry in entries:
        cov_file_idx = entry.name.find('##')
        if cov_file_idx > 0 and entry.name.endswith('.gcov'):
            cov_file = entry.name[cov_file_idx+2:-5]
            for source_info in data:
                if source_info.endswith(cov_file):
                    # print('match')
                    # print(entry.name)
                    # print(cov_file)
                    # print(data[source_info]['lines'])
                    # parse entry.name file, line numbers can show up more than once (contructors?)
                    with open(entry.path) as cov:
                        with closing(mmap(cov.fileno(), 0, access=ACCESS_READ)) as c:
                            while c:
                                line = c.readline()
                                if not line:
                                    break
                                parts = line.strip().split(b':', maxsplit=2)
                                if len(parts) != 3:
                                    continue
                                if b'#' in parts[0] or b'-' in parts[0]:
                                    continue
                                hit = int(parts[1].strip())
                                # print(hit)
                                bitmap[data[source_info]['offset'] + hit] = 0
                    break

# create png file
with open('/builds/src/bitcov.png', 'wb') as png_out:
    w = png.Writer(len(bitmap), 1, greyscale=True, bitdepth=1)
    w.write(png_out, [bitmap])

# print("wrote /builds/src/bitcov.png")
# print(offset)
# subprocess.run(shlex.split('ls -l /builds/src/bitcov.png'))
# subprocess.run(shlex.split('cat /builds/src/coverage-src-summary.json'))

parser = os.environ['MR_PARSER']
url = os.environ['CURRENT_URL']
db = os.environ['MR_POSTGRES_CONN'] if 'MR_POSTGRES_CONN' in os.environ else ""
baseline = os.environ['MR_IS_BASELINE'] if 'MR_IS_BASELINE' in os.environ else 'false'

if db != "":
    connection = psycopg2.connect(db)
    cursor = connection.cursor()
    bitcov_update = "UPDATE consensus SET bitcov = %s WHERE parser = %s AND doc = %s AND baseline = %s"
    with open('/builds/src/bitcov.png', 'rb') as f:
        data = f.read()
        binary_png = psycopg2.Binary(data)

    cursor.execute(bitcov_update, (binary_png, parser, url, baseline))
    connection.commit()
    connection.close()
else:
    loc = 0
    for pixel in bitmap:
        if pixel == 0:
            loc = loc + 1
    print(f"{loc} out of {len(bitmap)} lines visited")
