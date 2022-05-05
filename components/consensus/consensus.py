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
import subprocess
import psycopg2
import parsers.caradoc
import parsers.mupdf
import parsers.qpdf
import parsers.poppler
import parsers.xpdf
import parsers.pdfbox
import parsers.gdal
import parsers.nitro
import parsers.kaitai_nitf
import parsers.qpdf_trace
import parsers.pdfminer_six
import parsers.demoiccmax
import parsers.gst
import parsers.ffmpeg
from parsers.cfg_utils import create_cfg_output


def process(parsers):
    parser = os.environ['MR_PARSER']
    db = os.environ['MR_POSTGRES_CONN'] if 'MR_POSTGRES_CONN' in os.environ else ""
    baseline = os.environ['MR_IS_BASELINE'] if 'MR_IS_BASELINE' in os.environ else "false"
    universe = os.environ['MR_UNIVERSE'] if 'MR_UNIVERSE' in os.environ else "n/a"
    urls = os.environ['MR_DOC_URL'] if 'MR_DOC_URL' in os.environ else "out.doc"
    
    is_baseline = False
    if baseline == "true":
        is_baseline = True
    filename = 'doc.pdf'
    url_list = urls.split()
    if db != "":
        connection = psycopg2.connect(db)
        cursor = connection.cursor()

    for url in url_list:
        os.system("find /builds/src/ -type f -name '*.gcda' -delete")
        if 'http' in url:
            with urllib.request.urlopen(url) as response, open(filename, 'wb') as output:
                shutil.copyfileobj(response, output)

        sha256_hash = hashlib.sha256()
        hexdigest = ""
        with open(filename, 'rb') as f:
            for byte_block in iter(lambda: f.read(4096),b""):
                sha256_hash.update(byte_block)
            hexdigest = sha256_hash.hexdigest()

        report = parsers.get(parser)(filename)
        report['digest'] = hexdigest
        report['MR_DOC_URL'] = url
        report['MR_PARSER'] = parser
        # failed baseline should not be part of the baseline CFG
        report['callgrind'] = ""
        report['cfg'] = ""
        # why not gather cfg and prof data for rejected files?
        #if not (report['status'] == 'rejected' and is_baseline):
        create_cfg_output(report)

        if db != "":
            # insert_query = "INSERT INTO " + table_name + " (parser, doc, digest, status, stdout, stderr, callgrind, cfg, cfg_image) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)"
            # cfg_image is about 3MB per doc, doesn't scale
            # TODO create rest endpoint to create the png
            insert_query = "INSERT INTO consensus (parser, doc, baseline, digest, status, stdout, stderr, callgrind, cfg, tag) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"
            # binary_png = psycopg2.Binary(base64.b64decode(report['cfg_image']))
            cursor.execute(insert_query, (parser, url, is_baseline, hexdigest, report['status'], report['stdout'], report['stderr'], report['callgrind'], report['cfg'], universe))
            connection.commit()
            # connection.close()

        # check for bitcov tool mode
        os.environ["CURRENT_URL"] = url
        if os.path.isdir("/builds/src"):
            proc = subprocess.run(['python3', '/consensus/coverage.py'], cwd='/builds/src', capture_output=True)
            if db == "":
                print("status: " + report['status'])
                print("")
                print("stderr")
                print(report["stderr"])
                print("")
                print("bitcov")
                print(proc.stdout.decode('utf-8', errors='backslashreplace').strip())

    if db != "":
        connection.close()

if __name__ == "__main__":
    parsers = {
        "caradoc": parsers.caradoc.run,
        "mupdf": parsers.mupdf.run,
        "qpdf": parsers.qpdf.run,
        "qpdf_trace": parsers.qpdf_trace.run,
        "poppler_pdftotext": parsers.poppler.pdftotext,
        "poppler_pdftoppm": parsers.poppler.pdftoppm,
        "poppler_pdffonts": parsers.poppler.pdffonts,
        "xpdf_pdftotext": parsers.xpdf.pdftotext,
        "xpdf_pdftoppm": parsers.xpdf.pdftoppm,
        "xpdf_pdffonts": parsers.xpdf.pdffonts,
        "pdfbox": parsers.pdfbox.run,
        "nitro": parsers.nitro.run,
        "gdal": parsers.gdal.run,
        "kaitai_nitf": parsers.kaitai_nitf.run,
        "pdfminer_six": parsers.pdfminer_six.run,
        "iccdumpprofile": parsers.demoiccmax.iccdumpprofile,
        "iccapplyprofiles": parsers.demoiccmax.iccapplyprofiles,
        "gst": parsers.gst.gst,
        "ffmpeg": parsers.ffmpeg.ffmpeg
    }

    if len(sys.argv) == 2 and sys.argv[1] == "stdin":
        with open("doc.pdf", "wb") as outfile:
            for line in sys.stdin.buffer:
                outfile.write(line)
    
    process(parsers)
