#!/usr/bin/python3

import os
import hashlib
import subprocess
from subprocess import PIPE


for root, subdirs, files in os.walk("./", followlinks=True):
    if "./" == root:
        continue
    for file in files:
        sha256_hash = hashlib.sha256()
        hexdigest = ""
        try:
            fpath = root[2:] + '/' + file
            # get sha256 hash
            #with open(fpath, 'rb') as f:
            #    for byte_block in iter(lambda: f.read(4096), b""):
            #        sha256_hash.update(byte_block)
            #    hexdigest = sha256_hash.hexdigest()
            
            # get magic
            #proc = subprocess.run(['file', fpath], stdout=PIPE, stderr=PIPE)
            #magic = proc.stdout.split(b":")[1].strip().decode()

            print(fpath)#, hexdigest, os.stat(fpath).st_size, '"' + magic + '"', sep=',')
        except:
            pass
