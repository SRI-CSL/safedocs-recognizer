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

### Examples

#### Run the recognizer's consensus component over the entire document index using poppler's pdftoppm and marking the results as a baseline and set the universe label as 'univA'

```
./recognizer process --component consensus --baseline --tag mr_poppler_0.86.1 --parser poppler_pdftoppm --universe univA
```
