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
    echo "test2" | "$BIN/lb" insert keys/k/one2
    echo "test" | "$BIN/lb" insert keys/k/one
    echo "test" | "$BIN/lb" insert key/a/one
    echo "test" | "$BIN/lb" insert keys/k/one
    echo -e "test3\ntest4" | "$BIN/lb" insert keys2/k/three
    "$BIN/lb" ls
    yes 2>/dev/null | "$BIN/lb" rm keys/k/one
    echo
    "$BIN/lb" ls
    "$BIN/lb" find e
    "$BIN/lb" show keys/k/one2
    "$BIN/lb" show keys2/k/three
    echo "5ae472abqdekjqykoyxk7hvc2leklq5n" | "$BIN/lb" insert test/k/totp
    "$BIN/lb" "totp" -list
    "$BIN/lb" "totp" test/k | tr '[:digit:]' 'X'
    "$BIN/lb" "hash" $LOCKBOX_STORE
    yes 2>/dev/null | "$BIN/lb" rm keys2/k/three
    echo
    yes 2>/dev/null | "$BIN/lb" rm test/k/totp
    yes 2>/dev/null | "$BIN/lb" rm test/k/one
    yes 2>/dev/null | "$BIN/lb" rm key/a/one
    "$BIN/lb" ls
}

_run 2>&1 | sed "s#$LOCKBOX_STORE##g" > $TESTS/actual.log
