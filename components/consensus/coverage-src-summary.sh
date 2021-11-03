#!/bin/sh

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
