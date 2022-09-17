#!/bin/bash
BIN="$1"
TESTS="$2"

export LOCKBOX_STORE="$TESTS/lb"
export LOCKBOX_KEYMODE="plaintext"
export LOCKBOX_KEY="plaintextkey"
export LOCKBOX_TOTP="totp"
export LOCKBOX_INTERACTIVE="no"
export LOCKBOX_HOOKDIR="$TESTS/hooks"
export LOCKBOX_GIT="no"
export LOCKBOX_ALGORITHM="$3"

rm -rf $TESTS
mkdir -p $LOCKBOX_STORE
mkdir -p $LOCKBOX_STORE/$LOCKBOX_TOTP
git -C $LOCKBOX_STORE init
echo "TEST" > $LOCKBOX_STORE/init
git -C $LOCKBOX_STORE add .
git -C $LOCKBOX_STORE config user.email "you@example.com"
git -C $LOCKBOX_STORE config user.name "Your Name"
git -C $LOCKBOX_STORE commit -am "init"
HOOK=$LOCKBOX_HOOKDIR/hook
mkdir -p $LOCKBOX_HOOKDIR

_hook() {
    echo "#!/bin/sh"
    echo "echo HOOK RAN \$@"
}

_run() {
    echo "y" | "$BIN/lb" dump -yes "*"
    echo "test" | "$BIN/lb" insert keys/one
    echo "test2" | "$BIN/lb" insert keys/one2
    "$BIN/lb" show keys/*
    "$BIN/lb" dump -yes '***'
    echo -e "test3\ntest4" | "$BIN/lb" insert keys2/three
    "$BIN/lb" ls
    "$BIN/lb" "rekey"
    yes 2>/dev/null | "$BIN/lb" rm keys/one
    echo
    "$BIN/lb" list
    "$BIN/lb" find e
    "$BIN/lb" show keys/one2
    "$BIN/lb" show keys2/three
    echo "y" | "$BIN/lb" dump keys2/three
    echo "5ae472abqdekjqykoyxk7hvc2leklq5n" | "$BIN/lb" insert test/totp
    "$BIN/lb" "totp" -list
    "$BIN/lb" "totp" test | tr '[:digit:]' 'X'
    "$BIN/lb" "gitdiff" bin/lb/keys/one.lb bin/lb/keys/one2.lb
    yes 2>/dev/null | "$BIN/lb" rm keys2/three
    echo
    yes 2>/dev/null | "$BIN/lb" rm test/totp
    echo
    "$BIN/lb" kdbx -file bin/file.kdbx
    LOCKBOX_KEY="invalid" "$BIN/lb" show keys/one2
    "$BIN/lb" "rekey" -outkey "test" -outmode "plaintext"
    "$BIN/lb" rw -file bin/lb/keys/one2.lb -key "test" -keymode "plaintext" -mode "decrypt"
}

_hook > $HOOK
chmod 755 $HOOK
LOG=$TESTS/actual.log
_run 2>&1 | sed "s#$LOCKBOX_STORE##g" > $LOG
if [[ "$3" != "secretbox" ]]; then
    sed -i 's/cipher: message authentication failed/decrypt not ok/g' $LOG
fi
