#!/usr/bin/awk -f

BEGIN {
    print "extracting", pdf_object;
}

$1==pdf_object {
    print "dd if=<file> of=invalid_object.bin bs=1 count="$4" skip="$3-1
    print "hexdump -v <file> -c -n "$4" -s "$3
}
