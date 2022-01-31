#!/usr/bin/awk -f

BEGIN {
    file_status = "none";
}

/^doc.pdf \(object [0-9]+ [0-9]+, offset [0-9]+\):/ {
    invalid_objects["PDFObject" $3 "." substr($4, 1, length($4) - 1)] = substr($6, 1, length($6) - 2);
}

/^status: rejected/ {
    file_status = "rejected";
}

/^status: valid/ {
    file_status = "valid";
}

END {
    if (file_status == "rejected")
    {
        for (obj in invalid_objects)
        {
            print(obj);
            # print(invalid_objects[obj]);
        }
    }
}
