#!/bin/bash
BIN=../bin/
TESTS="$PWD/bin"

export LOCKBOX_STORE="$TESTS/lb"
export LOCKBOX_KEYMODE="plaintext"
export LOCKBOX_TOTP="totp"
export PWGEN_SOURCE="$PWD"
export PWGEN_SPECIAL="u"
export PWGEN_SED="s/[[:alnum:]]/u/g;s/\./u/g"

rm -rf $TESTS
mkdir -p $LOCKBOX_STORE
mkdir -p $LOCKBOX_STORE/$LOCKBOX_TOTP

_run() {
    echo "test" | $BIN/lb insert keys/one
    echo "test2" | $BIN/lb insert keys/one2
    echo -e "test3\ntest4" | $BIN/lb insert keys2/three
    $BIN/lb ls
    $BIN/lb pwgen -special -length 10
    $BIN/lb-rekey
    yes | $BIN/lb rm keys/one
    echo
    $BIN/lb list
    $BIN/lb show keys/one2
    $BIN/lb show keys2/three
    echo "5ae472abqdekjqykoyxk7hvc2leklq5n" | $BIN/lb insert totp/test
    $BIN/lb-totp ls
    $BIN/lb-totp test | head -3 | tail -n 1
    $BIN/lb-stats keys/one
    $BIN/lb-diff bin/lb/keys/one.lb bin/lb/keys/one2.lb
    yes | $BIN/lb rm keys2/three
    echo
    yes | $BIN/lb rm totp/test
    echo
    $BIN/lb-rekey -outkey "test" -outmode "plaintext"
    $BIN/lb-rw -file bin/lb/keys/one2.lb -key "test" -keymode "plaintext" -mode "decrypt"
}

LOG=$TESTS/lb.log
_run | sed "s#$PWD/##g" > $LOG
diff -u $LOG expected.log
if [ $? -ne 0 ]; then
    exit 1
fi
