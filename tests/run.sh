#!/bin/sh
LB_BINARY=../target/lb
DATA="testdata/$1"
TOML="$DATA/config.toml"
CLIP_WAIT=1
CLIP_TRIES=3
CLIP_COPY="$DATA/clip.copy"
CLIP_PASTE="$DATA/clip.paste"
PASS_TEST="password"
KEYF_TEST="keyfile"
BOTH_TEST="both"

_unset() {
  # shellcheck disable=SC2046
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
export LOCKBOX_CONFIG_TOML="none"
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

export LOCKBOX_CREDENTIALS_KEY_FILE=""
export LOCKBOX_CREDENTIALS_PASSWORD=""
VALID=0
if [ "$1" = "$PASS_TEST" ] || [ "$1" = "$BOTH_TEST" ]; then
  VALID=1
  export LOCKBOX_CREDENTIALS_PASSWORD="testingkey"
fi
if [ "$1" = "$KEYF_TEST" ] || [ "$1" = "$BOTH_TEST" ]; then
  VALID=1
  KEYFILE="$DATA/test.key"
  echo "thisisatest" > "$KEYFILE"
  export LOCKBOX_CREDENTIALS_KEY_FILE="$KEYFILE"
fi
if [ "$VALID" -eq 0 ]; then
  echo "invalid test"
  exit 1
fi

LOGFILE="$DATA/actual.log"
printf "%-10s ... " "$1"
{
  export LOCKBOX_HOOKS_DIRECTORY=""
  export LOCKBOX_STORE="${DATA}/passwords.kdbx"
  export LOCKBOX_TOTP_ENTRY=totp
  export LOCKBOX_INTERACTIVE=false
  export LOCKBOX_READONLY=false
  if [ "$LOCKBOX_CREDENTIALS_PASSWORD" = "" ]; then
    export LOCKBOX_CREDENTIALS_PASSWORD_MODE=none
  else
    export LOCKBOX_CREDENTIALS_PASSWORD_MODE=plaintext
  fi
  export LOCKBOX_JSON_HASH_LENGTH=0
  echo test2 |${LB_BINARY} insert keys/k/one2
  OLDMODE="$LOCKBOX_CREDENTIALS_PASSWORD_MODE"
  OLDKEY="$LOCKBOX_CREDENTIALS_PASSWORD"
  if [ "$OLDKEY" != "" ]; then
    export LOCKBOX_INTERACTIVE=true
    export LOCKBOX_CREDENTIALS_PASSWORD_MODE=ask
    export LOCKBOX_CREDENTIALS_PASSWORD=""
  else
    printf "password: "
  fi
  echo "$OLDKEY" | ${LB_BINARY} ls 2>/dev/null
  if [ "$OLDKEY" != "" ]; then
    export LOCKBOX_INTERACTIVE=false
    export LOCKBOX_CREDENTIALS_PASSWORD_MODE="$OLDMODE"
    export LOCKBOX_CREDENTIALS_PASSWORD="$OLDKEY"
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
  export LOCKBOX_HOOKS_DIRECTORY="$PWD/hooks"
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
  REKEY_ARGS=""
  NEWKEY="newkey$1"
  export LOCKBOX_HOOKS_DIRECTORY=""
  if [ -n "$LOCKBOX_CREDENTIALS_KEY_FILE" ]; then
    REKEYFILE="$DATA/newkeyfile"
    REKEY_ARGS="-keyfile $REKEYFILE"
    echo "thisisanewkey" > "$REKEYFILE"
    if [ -z "$LOCKBOX_CREDENTIALS_PASSWORD" ]; then
      REKEY_ARGS="$REKEY_ARGS -nokey"
      NEWKEY=""
    fi
  fi
  # shellcheck disable=SC2086
  echo "$NEWKEY" | ${LB_BINARY} rekey $REKEY_ARGS
  export LOCKBOX_CREDENTIALS_PASSWORD="$NEWKEY"
  export LOCKBOX_CREDENTIALS_KEY_FILE="$REKEYFILE"
  echo
  ${LB_BINARY} ls
  ${LB_BINARY} show keys/k/one2
  export LOCKBOX_JSON_MODE=plaintext
  ${LB_BINARY} json k
  export LOCKBOX_JSON_MODE=empty
  ${LB_BINARY} json k
  export LOCKBOX_JSON_MODE=hash
  export LOCKBOX_JSON_HASH_LENGTH=3
  ${LB_BINARY} json k
  # clipboard
  export LOCKBOX_CLIP_COPY_COMMAND="touch $CLIP_COPY"
  export LOCKBOX_CLIP_PASTE_COMMAND="touch $CLIP_PASTE"
  export LOCKBOX_CLIP_TIMEOUT=5
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
  OLDKEY="$LOCKBOX_CREDENTIALS_PASSWORD"
  OLDMODE="$LOCKBOX_CREDENTIALS_PASSWORD_MODE"
  OLDKEYFILE="$LOCKBOX_CREDENTIALS_KEY_FILE"
  if [ -n "$LOCKBOX_CREDENTIALS_KEY_FILE" ]; then
    export LOCKBOX_CREDENTIALS_KEY_FILE=""
    if [ -z "$LOCKBOX_CREDENTIALS_PASSWORD" ]; then
      export LOCKBOX_CREDENTIALS_PASSWORD="garbage"
    fi
  else
    KEYFILE="$DATA/invalid.key"
    echo "invalid" > "$KEYFILE"
    export LOCKBOX_CREDENTIALS_KEY_FILE="$KEYFILE"
  fi
  if [ "$OLDMODE" = "none" ]; then
    export LOCKBOX_CREDENTIALS_PASSWORD_MODE="plaintext"
  fi
  ${LB_BINARY} ls
  export LOCKBOX_CREDENTIALS_KEY_FILE="$OLDKEYFILE"
  export LOCKBOX_CREDENTIALS_PASSWORD="$OLDKEY"
  export LOCKBOX_CREDENTIALS_PASSWORD_MODE="$OLDMODE"
  # configuration
  {
    cat << EOF
store = "$LOCKBOX_STORE"
interactive = false

[clip]
copy_command = [$(echo "$LOCKBOX_CLIP_COPY_COMMAND" | sed 's/ /", "/g;s/^/"/g;s/$/"/g')]
copy_command = [$(echo "$LOCKBOX_CLIP_PASTE_COMMAND" | sed 's/ /", "/g;s/^/"/g;s/$/"/g')]
timeout = $LOCKBOX_CLIP_TIMEOUT

[json]
mode = "$LOCKBOX_JSON_MODE"
hash_length = $LOCKBOX_JSON_HASH_LENGTH

[credentials]
key_file = "$LOCKBOX_CREDENTIALS_KEY_FILE"
password_mode = "$LOCKBOX_CREDENTIALS_PASSWORD_MODE"
password = "$LOCKBOX_CREDENTIALS_PASSWORD"
EOF
  } > "$TOML"
  _unset
  export LOCKBOX_FAKE_TEST=plain
  export LOCKBOX_CONFIG_TOML="none"
  ${LB_BINARY} ls
  export LOCKBOX_CONFIG_TOML="$TOML"
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
