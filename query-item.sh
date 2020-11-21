#!/bin/bash

db="db.sqlite"

sqlite3 ${db} <<EOF
SELECT ProductName, ItemName, InStock, DATETIME(Timestamp, 'localtime')
FROM stock
WHERE ItemName = "$1";
.exit
EOF
