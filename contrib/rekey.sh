#!/usr/bin/env bash
if [ -z "$LOCKBOX_REKEY" ] && [ -z "$LOCKBOX_REKEYFILE" ]; then
  echo "LOCKBOX_REKEY/LOCKBOX_REKEYFILE are not set properly for rekeying"
  exit 1
fi

NEW_STORE="$(date +%Y%m%d%H%M%S).lb.kdbx"
_rekey() {
  local entry modtime tmp
  tmp=$(mktemp).kdbx
  for entry in $(lb ls); do
    modtime=$(lb stats "$entry" | grep '^modtime:' | cut -d ":" -f 2- | sed 's/^\s*//g')
    echo "migrating: $entry"
    if ! lb show "$entry" | LOCKBOX_HOOKDIR="" LOCKBOX_SET_MODTIME="$modtime" LOCKBOX_STORE="$tmp" LOCKBOX_KEY="$LOCKBOX_REKEY" LOCKBOX_KEYFILE="$LOCKBOX_REKEYFILE" lb insert "$entry" > /dev/null; then
      echo "failed"
      rm -f "$tmp"
      exit 1
    fi
  done
  mv "$tmp" "$NEW_STORE"
}

_rekey
echo "completed, '$NEW_STORE' created"
