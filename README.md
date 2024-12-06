Lockbox
===

A [pass](https://www.passwordstore.org/) inspired password manager that uses a system
keyring or command for password input over using a GPG key and uses a keepass database as the backing data store.

[![build](https://github.com/seanenck/lockbox/actions/workflows/build.yml/badge.svg)](https://github.com/seanenck/lockbox/actions/workflows/build.yml)

# usage

## upfront

While `lb` uses a `.kdbx` formatted file that can be opened by a variety of tools, it is highly opinionated on how to store data in the database. Any
`.kdbx` used with `lb` should be managed by `lb` with a fallback ability to use other tools to alter the/view the file otherwise. Mainly lockbox itself
is using a common format so that it doesn't lock a user into a custom file format or dealing with gpg, age, etc. files and instead COULD be recovered
via other tooling if needed.

## environment

The following variables must be set to use `lb`, they can also be set via a
toml configuration file.

```
# the keying object to use to ACTUALLY unlock the passwords (e.g. using a gpg encrypted file with the password inside of it)
LOCKBOX_KEY="gpg --decrypt /Users/alice/.secrets/key.gpg"
# the location, on disk, of the password store
LOCKBOX_STORE=/Users/alice/.passwords/secrets.kdbx
```

Use `lb help verbose` for additional information about options and environment variables

### supported systems

`lb` should work on combinations of the following:
- linux/macOS/WSL
- zsh/bash/fish (for completions)
- amd64/arm64

built binaries are available on the [releases page](https://github.com/enckse/lockbox/releases)

## usage

### clipboard

Copy entries to clipboard
```
lb clip my/secret/password
```

### insert

Create a new entry
```
lb insert my/new/key
# or
lb multiline my/new/multi
# for multiline inserts
```

### list

List entries
```
lb ls
```

### remove

To remove an entry
```
lb rm my/old/key
```

### show

To see the text of an entry
```
lb show my/key/value
```

### totp

To get a totp token
```
lb totp show token
# 'token' must contain an entry with the name of LOCKBOX_TOTP
```

The token can be automatically copied to the clipboard too
```
lb totp clip token
```

### rekey

To rekey (change password/keyfile) use the `rekey` command
```
lb rekey -keyfile="my/new/keyfile"
```

### completions

generate shell specific completions (via auto-detect using `SHELL`)
```
lb completions
```

## git integration

To manage the `.kdbx` file in a git repository and see _actual_ text diffs add this to a `.gitconfig`
```
[diff "lb"]
    textconv = lb conv
```

Setup the `.gitattributes` for the repository to include
```
*.kdbx diff=lb
```

## build

Requires `just`

Clone this repository and:
```
just
```

_run `just check` to run tests_
