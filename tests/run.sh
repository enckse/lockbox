#!/usr/bin/env bash
COUNT=0
DATA="bin/"
rm -rf "$DATA"
mkdir -p "$DATA"
for i in $@; do
  ./harness.sh "$i" &
  COUNT=$((COUNT+1))
done
if [ "$COUNT" -eq 0 ]; then
  echo "no tests run"
  exit 1
fi
wait
ACTUAL=$(find "$DATA" -type f -name "passed" | wc -l)
if [ "$COUNT" -ne "$ACTUAL" ]; then
  echo "tests failed"
  exit 1
fi
