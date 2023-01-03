#!/bin/bash

while read p || [ -n "$p" ]
do
sed -i -e "/${p//\//\\/}/d" ./coverage.txt
done < ./scripts/exclude-from-unit-test.txt