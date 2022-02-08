#!/bin/sh

set -e

docker build -f components/consensus/Dockerfile.base_eval3 components/consensus -t mr_consensus_base_eval3

docker build -f components/consensus/Dockerfile.process_qpdf_10.1.0 components/consensus -t mr_qpdf

docker build -f components/consensus/Dockerfile.process_mupdf_1.16.1 components/consensus -t mr_mupdf

docker build -f components/consensus/Dockerfile.process_poppler_0.86.1 components/consensus -t mr_poppler_pdftoppm
docker build -f components/consensus/Dockerfile.process_poppler_0.86.1 components/consensus -t mr_poppler_pdffonts
docker build -f components/consensus/Dockerfile.process_poppler_0.86.1 components/consensus -t mr_poppler_pdftotext

docker build -f components/consensus/Dockerfile.process_xpdf_4.02 components/consensus -t mr_xpdf_pdftoppm
docker build -f components/consensus/Dockerfile.process_xpdf_4.02 components/consensus -t mr_xpdf_pdffonts
docker build -f components/consensus/Dockerfile.process_xpdf_4.02 components/consensus -t mr_xpdf_pdftotext

# docker build -f components/consensus/Dockerfile.process_pdfbox_2.0.17 components/consensus -t mr_pdfbox_2.0.17
# docker build -f components/consensus/Dockerfile.process_pdfminer.six_20201018 components/consensus -t mr_pdfminer.six_20201018

docker build -f components/file-features/Dockerfile.process components/file-features -t mr_file-features
