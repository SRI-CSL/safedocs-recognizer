#!/bin/sh

# Copyright SRI International 2019-2022 All Rights Reserved.
# This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.

# find . -type f \( -name "*.cpp" -o -name "*.cc" -o -name "*.h" \) -print0 | sort -z | xargs -r0 wc -l | gawk '
#     BEGIN { ORS = ""; print " [ "}
#     { printf "%s{\"filename\": \"%s\", \"lines\": %s}",
#           separator, $2, $1
#       separator = ", "
#     }
#     END { print " ] " }
# ' | jq . > coverage-src-summary.json

find . -type f \( -name "*.cpp" -o -name "*.c" -o -name "*.cc" -o -name "*.h" \) -print0 | sort -z | xargs -r0 wc -l | gawk '
    BEGIN { ORS = ""; print " { "}
    { printf "%s \"%s\": {\"lines\": %s}",
          separator, $2, $1
      separator = ", "
    }
    END { print " } " }
' | jq . > coverage-src-summary.json
