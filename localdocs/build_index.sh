#!/bin/sh

rm -f sd_index.gz

if [ ! -f sd_index.gz ]; then
    ./create_index.py > sd_index
    gzip sd_index
    echo "index built"
fi
