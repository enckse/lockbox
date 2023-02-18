#!/bin/sh
rm -rf ".git"
if ! make build; then
  echo "build failed"
  exit 1
fi
if ! make check; then
  echo "check failed"
  exit 1
fi
