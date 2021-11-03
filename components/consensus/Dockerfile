FROM ubuntu:18.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
                software-properties-common libpq-dev postgresql-client git build-essential \
                qpdf mupdf-tools poppler-utils zlib1g-dev libgmp-dev pkg-config m4 valgrind \
                graphviz libncurses-dev ssh libgirepository1.0-dev libnss3-dev ca-certificates \
                git cmake libfreetype6-dev libfontconfig-dev \
                libjpeg-dev libopenjp2-7-dev qtbase5-dev \
                gobject-introspection libglib2.0-dev libgtk-3-dev \
                scons libglib2.0 apt-transport-https ca-certificates gnupg wget curl

RUN add-apt-repository ppa:deadsnakes/ppa
RUN apt-get update && apt-get install -y python3.7 python3.7-dev curl
RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
RUN python3.7 get-pip.py

RUN add-apt-repository ppa:ubuntugis/ppa
RUN apt-get update && apt-get install -y gdal-bin

RUN wget -O - https://apt.kitware.com/keys/kitware-archive-latest.asc 2>/dev/null | gpg --dearmor - | tee /etc/apt/trusted.gpg.d/kitware.gpg >/dev/null
RUN apt-add-repository 'deb https://apt.kitware.com/ubuntu/ bionic main'
RUN apt-get update && apt-get install -y cmake

RUN add-apt-repository ppa:ubuntu-toolchain-r/test
RUN apt-get update && apt-get install -y gcc-9 g++-9

RUN add-apt-repository ppa:avsm/ppa
RUN apt-get update && apt-get install -y opam ocamlbuild

# build and install nitro
WORKDIR /
RUN curl -L https://github.com/mdaus/nitro/archive/NITRO-2.9.0.tar.gz -o NITRO-2.9.0.tar.gz
RUN tar zxvf NITRO-2.9.0.tar.gz
WORKDIR /nitro-NITRO-2.9.0
RUN mkdir build
WORKDIR /nitro-NITRO-2.9.0/build
RUN cmake ..
RUN cmake --build . --config RelWithDebInfo -j
RUN cmake --build . --config RelWithDebInfo --target install
RUN ctest
WORKDIR /nitro-NITRO-2.9.0/modules/c/nitf/tests
RUN gcc -g -I../include -I/nitro-NITRO-2.9.0/build/modules/c/nrt/include -I/nitro-NITRO-2.9.0/installLinux-RelWithDebInfo/include \
        -I/nitro-NITRO-2.9.0/build/modules/c/nitf/include -L/nitro-NITRO-2.9.0/installLinux-RelWithDebInfo/lib test_image_loading.c \
        -lnitf-c -lnrt-c -lcgm-c -lj2k-c -ljpeg-c -llibjpeg -lopenjpeg -lxerces-c -ldl -o nitf_tester
ENV PATH="/nitro-NITRO-2.9.0/modules/c/nitf/tests:${PATH}"

WORKDIR /NITF_PARSER
COPY ngc-nitf/cxxopts.hpp /usr/include
COPY ngc-nitf/NITF_PARSER/* /NITF_PARSER/
RUN git clone https://gitlab.special-circumstanc.es/hammer/hammer.git
WORKDIR /NITF_PARSER/hammer
RUN scons install --variant=debug prefix=/usr
WORKDIR /NITF_PARSER
RUN g++-9 -std=c++17 -g *.cpp -o nitf-parser $(pkg-config --cflags --libs libhammer)
RUN ./nitf-parser
ENV PATH="/NITF_PARSER:${PATH}"

# kaitai
WORKDIR /kaitai-nitf
RUN pip3 install kaitaistruct
COPY kaitai-nitf/kaitai-nitf.py .
RUN wget https://formats.kaitai.io/nitf/src/python/nitf.py

# build and install caradoc
RUN opam init --disable-sandboxing
RUN opam switch create 4.04.0
RUN opam install -y ocamlfind
RUN opam install -y cryptokit ounit menhir curses
ENV PATH="/root/.opam/4.04.0/bin:${PATH}"
WORKDIR /
RUN git clone https://github.com/caradoc-org/caradoc.git
WORKDIR /caradoc
RUN make

RUN pip3 install psycopg2
RUN pip3 install gprof2dot
RUN pip3 install pyprof2calltree

WORKDIR /consensus
COPY consensus.py /consensus/
COPY parsers/cfg_utils.py /consensus/parsers/

COPY parsers/caradoc.py /consensus/parsers/
COPY parsers/mupdf.py /consensus/parsers/
COPY parsers/poppler.py /consensus/parsers/
COPY parsers/qpdf.py /consensus/parsers/
COPY parsers/qpdf_trace.py /consensus/parsers/
COPY parsers/nitro.py /consensus/parsers/
COPY parsers/gdal.py /consensus/parsers/
COPY parsers/ngcnitf.py /consensus/parsers/
COPY parsers/kaitai_nitf.py /consensus/parsers/
COPY parsers/adobe_reader.py /consensus/parsers/

CMD [ "/consensus/consensus.py" ]
