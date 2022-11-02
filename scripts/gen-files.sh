#!/bin/env bash
#
# Generate random files.
# Mostly useful for making small files,
# as larger ones will take a lot of time
#

if [ "$#" -ne 3 ]
then
    echo "missing required arguments"
    echo "usage: $0 <dest> <count> <size>"
    exit 1
fi

dest=$1
count=$2
size=$3

echo "generating..."

for (( i=1 ; i<=$count ; i++ ));
do
    echo "on ${i}"
    uuid=$(uuidgen -r)
    filename="${dest}/${uuid}.bin"
    dd status=none if=/dev/urandom of=${filename} bs=1 count=${size}
done

echo "done"
