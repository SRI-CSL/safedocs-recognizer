#!/usr/bin/python3

import argparse
from pprint import pprint
from nitf import Nitf

parser = argparse.ArgumentParser()
parser.add_argument("file")
args = parser.parse_args()
n = Nitf.from_file(args.file)
pprint(n.__dict__)
pprint(n.header.__dict__)
for i in n.image_segments:
    print("image_segment")
for g in n.graphics_segments:
    print("graphic_segment")
for t in n.text_segments:
    print(t)
