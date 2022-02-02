#!/usr/bin/env bash

DOCUMENT=$1

docker run --rm -i mr_file-features stdin < $DOCUMENT | awk -v pdf_object=$(docker run --rm -i mr_qpdf stdin < $DOCUMENT | awk -f invalid_object.awk) -f extract_bytes.awk
