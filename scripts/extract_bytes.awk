#!/usr/bin/awk -f

BEGIN {
    print "finding", pdf_object;
}

$1==pdf_object {
    print "Run:"
    print "dd of=invalid_object.bin bs=1 count="$4" skip="$3-1" if=<file>"
    print "hexdump -c -n "$4" -s "$3" -v <file>"
}
