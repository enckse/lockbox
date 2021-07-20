#!/bin/bash
GEN=$1
VERS=$2
VERS_NAME="Version"
_version_info() {
    echo "// Version is the hash/version info for lb"
    git log -n 1 --format=%h | sed 's/^/'$VERS_NAME' = "/g;s/$/"/g'
}

_version() {
    echo "package internal

const ("
    _version_info | sed 's/^/\t/g'
    echo ")"
}

_getvers() {
    cat $1 | grep "$VERS_NAME =" | awk '{print $3}'
}

_version > $VERS
if [ -e $GEN ]; then
    OLDVERS=$(_getvers $GEN)
    NEWVERS=$(_getvers $VERS)
    if [[ "$NEWVERS" == "$OLDVERS" ]]; then
        rm -f $VERS
        exit 0
    fi
fi
mv $VERS $GEN
