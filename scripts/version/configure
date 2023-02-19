#!/usr/bin/env sh
if [ ! -d .git ]; then
  2>&1 echo "not git controlled"
  exit 0
fi

_version() {
  curr=v$(date +%g.%m.)
  tag=$(git describe --tags --abbrev=0)
  minor=00
  if echo "$tag" | grep -q "$curr*"; then
    minor=$(echo "$tag" | cut -d '.' -f 3 | sed 's/^0//g')
    minor=$((minor+1))
    if [ $minor -lt 10 ]; then
      minor="0$minor"
    fi
  fi
  vers="$curr$minor"
  printf "%s" "$vers"
}

_version > "$1"
