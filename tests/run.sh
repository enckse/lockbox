#!/bin/sh
LB_BINARY=../bin/lb
DATA="bin/$1"
ENV="$DATA/env"
CLIP_WAIT=1
CLIP_TRIES=3
CLIP_COPY="$DATA/clip.copy"
CLIP_PASTE="$DATA/clip.paste"
PASS_TEST="password"
KEYF_TEST="keyfile"
BOTH_TEST="both"

_unset() {
  unset $(env | grep '^LOCKBOX' | cut -d "=" -f 1)
}

if [ -z "$1" ]; then
  CODE=0
  for TEST_RUN in "$PASS_TEST" "$KEYF_TEST" "$BOTH_TEST"; do
    if ! "$0" "$TEST_RUN"; then
      CODE=1
    fi
  done
  exit $CODE
fi

_unset
export LOCKBOX_ENV="none"
if [ ! -x "${LB_BINARY}" ]; then
  echo "binary missing?"
  exit 1
fi
if ! ${LB_BINARY} help >/dev/null; then
  echo "help unavailable by default...fatal"
  exit 1
fi
mkdir -p "$DATA"
find "$DATA" -type f -delete

export LOCKBOX_KEYFILE=""
export LOCKBOX_KEY=""
VALID=0
if [ "$1" = "$PASS_TEST" ] || [ "$1" = "$BOTH_TEST" ]; then
  VALID=1
  export LOCKBOX_KEY="testingkey"
fi
if [ "$1" = "$KEYF_TEST" ] || [ "$1" = "$BOTH_TEST" ]; then
  VALID=1
  KEYFILE="$DATA/test.key"
  echo "thisisatest" > "$KEYFILE"
  export LOCKBOX_KEYFILE="$KEYFILE"
fi
if [ "$VALID" -eq 0 ]; then
  echo "invalid test"
  exit 1
fi

LOGFILE="$DATA/actual.log"
printf "%-10s ... " "$1"
{
  export LOCKBOX_HOOKDIR=""
  export LOCKBOX_STORE="${DATA}/passwords.kdbx"
  export LOCKBOX_TOTP=totp
  export LOCKBOX_INTERACTIVE=no
  export LOCKBOX_READONLY=no
  if [ "$LOCKBOX_KEY" = "" ]; then
    export LOCKBOX_KEYMODE=none
  else
    export LOCKBOX_KEYMODE=plaintext
  fi
  export LOCKBOX_JSON_DATA_HASH_LENGTH=0
  echo test2 |${LB_BINARY} insert keys/k/one2
  OLDMODE="$LOCKBOX_KEYMODE"
  OLDKEY="$LOCKBOX_KEY"
  if [ "$OLDKEY" != "" ]; then
    export LOCKBOX_INTERACTIVE=yes
    export LOCKBOX_KEYMODE=ask
    export LOCKBOX_KEY=""
  else
    printf "password: "
  fi
  echo "$OLDKEY" | ${LB_BINARY} ls 2>/dev/null
  if [ "$OLDKEY" != "" ]; then
    export LOCKBOX_INTERACTIVE=no
    export LOCKBOX_KEYMODE="$OLDMODE"
    export LOCKBOX_KEY="$OLDKEY"
  fi
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
  # rekeying
  REKEY="$LOCKBOX_STORE.rekey.kdbx"
  REKEYFILE=""
  export LOCKBOX_HOOKDIR=""
  if [ -n "$LOCKBOX_KEYFILE" ]; then
    REKEYFILE="$DATA/newkeyfile"
    echo "thisisanewkey" > "$REKEYFILE"
  fi
  echo y |${LB_BINARY} rekey -store="$REKEY" -key="newkey$1" -keymode="plaintext" -keyfile="$REKEYFILE"
  echo
  ${LB_BINARY} ls
  ${LB_BINARY} show keys/k/one2
  export LOCKBOX_JSON_DATA=plaintext
  ${LB_BINARY} json k
  export LOCKBOX_JSON_DATA=empty
  ${LB_BINARY} json k
  export LOCKBOX_JSON_DATA=hash
  export LOCKBOX_JSON_DATA_HASH_LENGTH=3
  ${LB_BINARY} json k
  # clipboard
  export LOCKBOX_CLIP_COPY="touch $CLIP_COPY"
  export LOCKBOX_CLIP_PASTE="touch $CLIP_PASTE"
  export LOCKBOX_CLIP_MAX=5
  ${LB_BINARY} clip keys/k/one2
  CLIP_PASSED=0
  while [ "$CLIP_TRIES" -gt 0 ] ; do
    if [ -e "$CLIP_COPY" ] && [ -e "$CLIP_PASTE" ]; then
      CLIP_PASSED=1
      break
    fi
    sleep "$CLIP_WAIT"
    CLIP_TRIES=$((CLIP_TRIES-1))
  done
  if [ $CLIP_PASSED -eq 0 ]; then
    echo "clipboard test failed"
  fi
  # invalid settings
  OLDKEY="$LOCKBOX_KEY"
  OLDMODE="$LOCKBOX_KEYMODE"
  OLDKEYFILE="$LOCKBOX_KEYFILE"
  if [ -n "$LOCKBOX_KEYFILE" ]; then
    export LOCKBOX_KEYFILE=""
    if [ -z "$LOCKBOX_KEY" ]; then
      export LOCKBOX_KEY="garbage"
    fi
  else
    KEYFILE="$DATA/invalid.key"
    echo "invalid" > "$KEYFILE"
    export LOCKBOX_KEYFILE="$KEYFILE"
  fi
  if [ "$OLDMODE" = "none" ]; then
    export LOCKBOX_KEYMODE="plaintext"
  fi
  ${LB_BINARY} ls
  export LOCKBOX_KEYFILE="$OLDKEYFILE"
  export LOCKBOX_KEY="$OLDKEY"
  export LOCKBOX_KEYMODE="$OLDMODE"
  # configuration
  {
    echo "PLAINTEXT=text"
    env | grep '^LOCKBOX' | sed 's/plaintext/$LOCKBOX_FAKE_TEST$PLAINTEXT/g'
  } > "$ENV"
  _unset
  export LOCKBOX_FAKE_TEST=plain
  export LOCKBOX_ENV="none"
  ${LB_BINARY} ls
  export LOCKBOX_ENV="$ENV"
  ${LB_BINARY} ls
} 2>&1 | \
  sed 's/"modtime": "[0-9].*$/"modtime": "XXXX-XX-XX",/g' | \
  sed 's/^[0-9][0-9][0-9][0-9][0-9][0-9]$/XXXXXX/g' > "$LOGFILE"
STATE=0
RESULT="passed"
if ! diff -u "$LOGFILE" "expected.log"; then
  RESULT="failed"
  STATE=1
fi
printf "[%s]\n" "$RESULT"
exit "$STATE"
