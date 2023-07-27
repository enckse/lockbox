#!/usr/bin/env bash
LB_BINARY=../bin/lb
DATA=bin
CLIP_WAIT=1
CLIP_TRIES=3
CLIP_COPY="$DATA/clip.copy"
CLIP_PASTE="$DATA/clip.paste"

_execute() {
  export LOCKBOX_HOOKDIR=""
  export LOCKBOX_STORE="${DATA}/passwords.kdbx"
  export LOCKBOX_TOTP=totp
  export LOCKBOX_INTERACTIVE=no
  export LOCKBOX_READONLY=no
  export LOCKBOX_KEYMODE=plaintext
  export LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH=0
  echo test2 |${LB_BINARY} insert keys/k/one2
  echo test |${LB_BINARY} insert keys/k/one
  echo test |${LB_BINARY} insert key/a/one
  echo test |${LB_BINARY} insert keys/k/one
  echo test |${LB_BINARY} insert keys/k/one/
  echo test |${LB_BINARY} insert /keys/k/one
  echo test |${LB_BINARY} insert keys/aa/b//s///e
  printf "test3\ntest4\n" |${LB_BINARY} insert keys2/k/three
  printf "test3\ntest4\n" |${LB_BINARY} multiline keys2/k/three
  ${LB_BINARY} ls
  echo y |${LB_BINARY} rm keys/k/one
  echo
  ${LB_BINARY} ls
  ${LB_BINARY} ls | grep e
  ${LB_BINARY} json
  echo
  ${LB_BINARY} show keys/k/one2
  ${LB_BINARY} show keys2/k/three
  ${LB_BINARY} json keys2/k/three
  echo
  echo 5ae472abqdekjqykoyxk7hvc2leklq5n |${LB_BINARY} totp insert test/k
  echo 5ae472abqdekjqykoyxk7hvc2leklq5n |${LB_BINARY} totp insert test/k/totp
  ${LB_BINARY} totp ls
  ${LB_BINARY} totp show test/k
  ${LB_BINARY} totp once test/k
  ${LB_BINARY} totp minimal test/k
  ${LB_BINARY} conv "$LOCKBOX_STORE"
  echo y |${LB_BINARY} rm keys2/k/three
  echo
  echo y |${LB_BINARY} rm test/k/totp
  echo
  echo y |${LB_BINARY} rm test/k/one
  echo
  echo
  echo test2 |${LB_BINARY} insert move/m/ka/abc
  echo test |${LB_BINARY} insert move/m/ka/xyz
  echo test2 |${LB_BINARY} insert move/ma/ka/yyy
  echo test |${LB_BINARY} insert move/ma/ka/zzz
  echo test |${LB_BINARY} insert move/ma/ka2/zzz
  echo test |${LB_BINARY} insert move/ma/ka3/yyy
  echo test |${LB_BINARY} insert move/ma/ka3/zzz
  ${LB_BINARY} mv move/m/* move/mac/
  ${LB_BINARY} mv move/ma/ka/* move/mac/
  ${LB_BINARY} mv move/ma/ka2/* move/mac/
  ${LB_BINARY} mv move/ma/ka3/* move/mac/
  ${LB_BINARY} mv key/a/one keyx/d/e
  ${LB_BINARY} ls
  echo y |${LB_BINARY} rm move/*
  echo y |${LB_BINARY} rm keyx/d/e
  echo
  ${LB_BINARY} ls
  echo test2 |${LB_BINARY} insert keys/k2/one2
  echo test |${LB_BINARY} insert keys/k2/one
  echo test2 |${LB_BINARY} insert keys/k2/t1/one2
  echo test |${LB_BINARY} insert keys/k2/t1/one
  echo test2 |${LB_BINARY} insert keys/k2/t2/one2
  export LOCKBOX_HOOKDIR="$PWD/hooks"
  echo test |${LB_BINARY} insert keys/k2/t2/one
  echo
  ${LB_BINARY} ls
  echo y |${LB_BINARY} rm keys/k2/t1/*
  echo
  ${LB_BINARY} ls
  echo y |${LB_BINARY} rm keys/k2/*
  echo
  ${LB_BINARY} ls
  echo
  _rekey
  _clipboard
  _invalid
}

_invalid() {
  local keyfile
  if [ -n "$LOCKBOX_KEYFILE" ]; then
    export LOCKBOX_KEYFILE=""
    if [ -z "$LOCKBOX_KEY" ]; then
      export LOCKBOX_KEY="garbage"
    fi
  else
    keyfile="$DATA/invalid.key"
    echo "invalid" > "$keyfile"
    export LOCKBOX_KEYFILE="$keyfile"
  fi
  ${LB_BINARY} ls
}

_rekey() {
  local rekey rekeyFile 
  rekey="$LOCKBOX_STORE.rekey.kdbx"
  rekeyFile=""
  export LOCKBOX_HOOKDIR=""
  if [ -n "$LOCKBOX_KEYFILE" ]; then
    rekeyFile="$DATA/newkeyfile"
    echo "thisisanewkey" > "$rekeyFile"
  fi
  echo y |${LB_BINARY} rekey -store="$rekey" -key="newkey" -keymode="plaintext" -keyfile="$rekeyFile"
  echo
  ${LB_BINARY} ls
  ${LB_BINARY} show keys/k/one2
  export LOCKBOX_JSON_DATA_OUTPUT=plaintext
  ${LB_BINARY} json k
  export LOCKBOX_JSON_DATA_OUTPUT=empty
  ${LB_BINARY} json k
  export LOCKBOX_JSON_DATA_OUTPUT=hash
  export LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH=3
  ${LB_BINARY} json k
}

_clipboard() {
  local clipTries
  export LOCKBOX_CLIP_COPY="touch $CLIP_COPY"
  export LOCKBOX_CLIP_PASTE="touch $CLIP_PASTE"
  export LOCKBOX_CLIP_MAX=5
  ${LB_BINARY} clip keys/k/one2
  clipTries="$CLIP_TRIES"
  while [ "$clipTries" -gt 0 ] ; do
    if [ -e "$CLIP_COPY" ] && [ -e "$CLIP_PASTE" ]; then
      return
    fi
    sleep "$CLIP_WAIT"
    clipTries=$((clipTries-1))
  done
  echo "clipboard test failed"
}

_logtest() {
  _execute 2>&1 | \
    sed 's/"modtime": "[0-9].*$/"modtime": "XXXX-XX-XX",/g' | \
    sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$/XXXXXX/g'
}

_evaluate() {
  local logfile
  logfile="$DATA/actual.log"
  _logtest > "$logfile"
  if ! diff -u "$logfile" "expected.log"; then
    echo "failed"
    exit 1
  fi
  echo "passed"
}

if [ -z "$1" ]; then
  echo "no test given"
  exit 1
fi
echo "$1"
echo "============"

mkdir -p "$DATA"
find "$DATA" -type f -delete

export LOCKBOX_KEYFILE=""
export LOCKBOX_KEY=""
VALID=0
if [ "$1" == "password" ] || [ "$1" == "both" ]; then
  VALID=1
  export LOCKBOX_KEY="testingkey"
fi
if [ "$1" == "keyfile" ] || [ "$1" == "both" ]; then
  VALID=1
  KEYFILE="$DATA/test.key"
  echo "thisisatest" > "$KEYFILE"
  export LOCKBOX_KEYFILE="$KEYFILE"
fi
if [ "$VALID" -eq 0 ]; then
  echo "invalid test"
  exit 1
fi
_evaluate
