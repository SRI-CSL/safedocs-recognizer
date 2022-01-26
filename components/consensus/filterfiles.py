#!/usr/bin/env python3
import glob
import json
import os

filetypes = [
    './**/*.c',
    './**/*.cc',
    './**/*.cpp',
    './**/*.h',
]

files = []
for fty in filetypes:
    files += glob.glob(fty,recursive=True)

parser = os.environ['MR_PARSER']
with open(f"./{parser}.json") as exclude:
    ffilter = json.load(exclude)
    for d in ffilter['directories']:
        files = filter(lambda x : not(x.startswith(d)),files)
    for f in ffilter['files']:
        files = filter(lambda x : not(x.endswith(f)),files)

results = { f : {'lines': sum(1 for _ in open(f))} for f in files}

with open('coverage-src-summary.json','w') as output:
    json.dump(results,output, sort_keys=True,indent=2)