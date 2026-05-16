#!/bin/bash

if [ $# -lt 3 ]; then
  echo "Usage: $0 <endpoint> <fileLocation> <nrOfRequest>"
  exit 1
fi

ENDPOINT=$1
NR_OF_REQUEST=$3
FILE_LOCATION=$2

if [ ! -f "$FILE_LOCATION" ]; then
  echo "Cannot locate file: $FILE_LOCATION"
  exit 1
fi

for _ in $(seq 1 "$NR_OF_REQUEST"); do
  curl -X POST "$ENDPOINT" -F "file=@$FILE_LOCATION" -H "Content-Type: multipart/form-data" &
done
wait

