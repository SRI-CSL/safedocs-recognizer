#!/usr/bin/env python3

"""
 Copyright SRI International 2019-2022 All Rights Reserved.
 This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
"""

import glob
import json
import sys

def genCoverageSummary(parserName):
    filetypes = [
        './**/*.c',
        './**/*.cc',
        './**/*.cpp'
        './**/*.h',
        './**/*.hh'
    ]

    files = []
    for fty in filetypes:
        files += glob.glob(fty,recursive=True)

    if parserName:
        with open(f"./{parserName}.json") as exclude:
            ffilter = json.load(exclude)
            for d in ffilter['directories']:
                if d == "": continue
                files = list(filter(lambda x : not(x.startswith(d)),files))
            for f in ffilter['files']:
                if f == "": continue
                files = list(filter(lambda x : not(x.endswith(f)),files))

    results = { f : {'lines': sum(1 for _ in open(f,'rb'))} for f in files}

    with open('coverage-src-summary.json','w') as output:
        json.dump(results, output, sort_keys=True, indent=2)

if __name__ == "__main__":
    genCoverageSummary(sys.argv[1] if len(sys.argv) == 2 else "")