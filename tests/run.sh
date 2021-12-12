#!/bin/bash
BIN="$1"
TESTS="bin"
if [ ! -x "$BIN/lb" ]; then
    echo "binaries missing?"
    exit 1
fi

export LOCKBOX_STORE="$TESTS/lb"
export LOCKBOX_KEYMODE="plaintext"
export LOCKBOX_KEY="plaintextkey"
export LOCKBOX_TOTP="totp"
export LOCKBOX_INTERACTIVE="no"
export PWGEN_SOURCE="$PWD"
export PWGEN_SPECIAL="u"
export PWGEN_SED="s/[[:alnum:]]/u/g;s/\./u/g"

rm -rf $TESTS
mkdir -p $LOCKBOX_STORE
mkdir -p $LOCKBOX_STORE/$LOCKBOX_TOTP
git -C $LOCKBOX_STORE init
echo "TEST" > $LOCKBOX_STORE/init
git -C $LOCKBOX_STORE add .
git -C $LOCKBOX_STORE config user.email "you@example.com"
git -C $LOCKBOX_STORE config user.name "Your Name"
git -C $LOCKBOX_STORE commit -am "init"

_run() {
    echo "test" | "$BIN/lb" insert keys/one
    echo "test2" | "$BIN/lb" insert keys/one2
    "$BIN/lb" show keys/*
    echo -e "test3\ntest4" | "$BIN/lb" insert keys2/three
    "$BIN/lb" ls
    "$BIN/lb-pwgen" -special -length 10
    "$BIN/lb-rekey"
    yes 2>/dev/null | "$BIN/lb" rm keys/one
    echo
    "$BIN/lb" list
    "$BIN/lb" find e
    "$BIN/lb" show keys/one2
    "$BIN/lb" show keys2/three
    echo "5ae472abqdekjqykoyxk7hvc2leklq5n" | "$BIN/lb" insert totp/test
    "$BIN/lb-totp" -ls
    "$BIN/lb-totp" test | tr '[:digit:]' 'X'
    "$BIN/lb-diff" bin/lb/keys/one.lb bin/lb/keys/one2.lb
    yes 2>/dev/null | "$BIN/lb" rm keys2/three
    echo
    yes 2>/dev/null | "$BIN/lb" rm totp/test
    echo
    LOCKBOX_KEY="invalid" "$BIN/lb" show keys/one2
    "$BIN/lb-rekey" -outkey "test" -outmode "plaintext"
    "$BIN/lb-rw" -file bin/lb/keys/one2.lb -key "test" -keymode "plaintext" -mode "decrypt"
}

LOG=$TESTS/lb.log
_run 2>&1 | sed "s#$LOCKBOX_STORE##g" > $LOG
if ! diff -u $LOG expected.log; then
    exit 1
fi
