# safedocs-recognizer
DARPA SafeDocs TA1 software suite to bundle and orchestrate various format-aware tracing tools.

## How to run

The first step is copying (or create a symlink) documents to the localdocs directory and creating the document index.

```
build_index.sh
```

The database should then be started to store processing results.

```
docker-compose up
```

Build the CLI tool

```
go build
```

### Examples

#### mupdf example

Baseline and non-baseline processing (for performance reasons and prevent multiple passes over 1mil files, the consensus component combines bitcov and cfg tools)

```
./recognizer process --component consensus --baseline --tag mr_mupdf_1.16.1 --parser mupdf --subset evalThree --universe univA
./recognizer process --component consensus --tag mr_mupdf_1.16.1 --parser mupdf --subset evalThree10kTest --universe univA
./recognizer process --component file-features --baseline --tag mr_file-features --subset evalThree
```

Integrated components
Derive model
```
./recognizer bitcov --parser mupdf --universe univA
./recognizer bitcov --parser mupdf --universe univB
```

Metrics comparing 10k non-baseline files with models A and B
```
./recognizer bitcov-diff --model mupdf_univA_model.png --parser mupdf
./recognizer bitcov-diff --model mupdf_univB_model.png --parser mupdf
```

Derive model
```
./recognizer flat-cfg --parser mupdf --universe univA
./recognizer flat-cfg --parser mupdf --universe univB
```

Metrics comparing 10k non-baseline files with models A and B
```
./recognizer flat-cfg-diff --parser mupdf --model mupdf_univA_flat_cfg_model.txt
./recognizer flat-cfg-diff --parser mupdf --model mupdf_univB_flat_cfg_model.txt
```