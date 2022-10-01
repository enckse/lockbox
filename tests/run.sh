#!/bin/bash
BIN="$1"
TESTS="$2"

export LOCKBOX_STORE="$TESTS/lb.kdbx"
export LOCKBOX_KEYMODE="plaintext"
export LOCKBOX_KEY="plaintextkey"
export LOCKBOX_TOTP="totp"
export LOCKBOX_INTERACTIVE="no"

rm -rf $TESTS
mkdir -p $TESTS

_run() {
    echo "test" | "$BIN/lb" insert keys/one
    echo "test2" | "$BIN/lb" insert keys/one2
    echo -e "test3\ntest4" | "$BIN/lb" insert keys2/three
    "$BIN/lb" ls
    yes 2>/dev/null | "$BIN/lb" rm keys/one
    echo
    "$BIN/lb" ls
    "$BIN/lb" find e
    "$BIN/lb" show keys/one2
    "$BIN/lb" show keys2/three
    echo "5ae472abqdekjqykoyxk7hvc2leklq5n" | "$BIN/lb" insert test/totp
    "$BIN/lb" "totp" -list
    "$BIN/lb" "totp" test | tr '[:digit:]' 'X'
    "$BIN/lb" "diff" $LOCKBOX_STORE
    yes 2>/dev/null | "$BIN/lb" rm keys2/three
    echo
    yes 2>/dev/null | "$BIN/lb" rm test/totp
}

_run 2>&1 | sed "s#$LOCKBOX_STORE##g" > $TESTS/actual.log
