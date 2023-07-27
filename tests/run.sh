#!/usr/bin/env bash
for i in $@; do
  ./harness.sh "$i" &
done
wait
