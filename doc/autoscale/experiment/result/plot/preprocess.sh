#!/bin/bash

hashtable=$(mktemp -d)

for var in "$@"
do
    path="$var.csv"
    cat "$var" | tr '|' ',' | awk -F, '{if ($1<=550) {print}}'  > "$path"
done
