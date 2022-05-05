#!/usr/bin/env python3

"""
 Copyright SRI International 2019-2022 All Rights Reserved.
 This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
"""

import os
import sys
import urllib.request
from urllib.parse import urlparse
import shutil
import hashlib
import copy
import psycopg2
import subprocess
import json


def flatten(root):
    current = root
    stack = []
    output = []

    while current is not None:
        stack.extend(current['subEls'])
        del current['subEls']
        output.append(current)
        if len(stack) > 0:
            current = stack[0]
            stack.pop(0)
        else:
            current = None

    return output

def remove_values(root):
    current = root
    stack = []

    while current is not None:
        stack.extend(current['subEls'])
        current.pop('value', None)
        current.pop('decoded', None)
        current.pop('b64contents', None)
        if len(stack) > 0:
            current = stack[0]
            stack.pop(0)
        else:
            current = None


def process():
    # MR_PARSER not used in this case
    # parser = os.environ['MR_PARSER']
    db = os.environ['MR_POSTGRES_CONN'] if 'MR_POSTGRES_CONN' in os.environ else ""
    baseline = os.environ['MR_IS_BASELINE'] if 'MR_IS_BASELINE' in os.environ else "false"
    urls = os.environ['MR_DOC_URL'] if 'MR_DOC_URL' in os.environ else "out.doc"

    is_baseline = False
    if baseline == "true":
        is_baseline = True
    filename = 'doc.pdf'
    url_list = urls.split()
    for url in url_list:
        if db != "":
            p = urlparse(url)
            filepath = p.path.rsplit("/", 1)[-1]
            connection = psycopg2.connect(db)
            cursor = connection.cursor()
            exists_query = "SELECT doc FROM file_features WHERE substring(doc from '(?:.+/)(.+)') = '%s'"
            cursor.execute(exists_query % filepath)
            if cursor.fetchone() != None:
                print(f'{filepath} has already been processed, skipping')
                continue
        
        if 'http' in url:
            with urllib.request.urlopen(url) as response, open(filename, 'wb') as output:
                shutil.copyfileobj(response, output)

        sha256_hash = hashlib.sha256()
        hexdigest = ""
        with open(filename, 'rb') as f:
            for byte_block in iter(lambda: f.read(4096),b""):
                sha256_hash.update(byte_block)
            hexdigest = sha256_hash.hexdigest()

        polyfile = subprocess.run(["timeout", "5m", "polyfile", filename], capture_output=True)
        features = polyfile.stdout.decode("utf-8")

        proc = subprocess.run(['file', filename], capture_output=True)
        magic = proc.stdout.split(b":")[1].strip().decode()
        
        features_list = []
        # make an empty response valid json
        if features == '':
            features = "{}"
            features_list = '[]'
        else:
            # don't store the file in base64 form in the db
            features = json.loads(features)
            del features['b64contents']

            remove_values(features['struc'][0])
            
            struc_copy = copy.deepcopy(features['struc'][0])
            features_list = flatten(struc_copy)
            features_list.sort(key=lambda f: f['offset'])
            features_list_s = json.dumps(features_list)
            features_s = json.dumps(features)

        if db != "":
            insert_query = "INSERT INTO file_features (doc, baseline, magic, digest, features, features_list) VALUES (%s, %s, %s, %s, %s, %s)"
            cursor.execute(insert_query, (url, is_baseline, magic, hexdigest, features_s, features_list_s))
            connection.commit()
        else:
            for feature in features_list:
                print(feature['name'], feature['type'], feature['offset'], feature['size'])


def test_flatten1():
    root = {'name': 'root', 'subEls': []}
    a = {'name': 'a', 'subEls': []}
    b = {'name': 'b', 'subEls': []}
    c = {'name': 'c', 'subEls': []}
    d = {'name': 'd', 'subEls': []}
    root['subEls'].append(a)
    a['subEls'].append(b)
    a['subEls'].append(c)
    root['subEls'].append(d)
    flat = flatten(root)
    print(flat)


def test_flatten2():
    root = {'name': 'root', 'subEls': []}
    a = {'name': 'a', 'subEls': []}
    b = {'name': 'b', 'subEls': []}
    c = {'name': 'c', 'subEls': []}
    d = {'name': 'd', 'subEls': []}
    root['subEls'].append(a)
    a['subEls'].append(b)
    a['subEls'].append(c)
    b['subEls'].append(d)
    flat = flatten(root)
    print(flat)


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "stdin":
        with open("doc.pdf", "wb") as outfile:
            for line in sys.stdin.buffer:
                outfile.write(line)

    process()
