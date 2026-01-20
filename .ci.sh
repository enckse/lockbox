#!/bin/sh -e
ENGINE=podman
[ -n "$CONTAINER_ENGINE" ] && ENGINE="$CONTAINER_ENGINE"

_ci() {
  "$ENGINE" run --rm -v "$PWD:/app" -w /app golang:latest sh -c "make $1"
}

_ci ""
_ci "check"
_ci "releases"
