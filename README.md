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

## configuration

There are two ways to configure `lb`:
- TOML configuration file(s)
- Environment variables

The TOML configuration files have higher priority over environment variables
(if both are set) where the TOML files are ultimately loaded into the
processes environment itself (once parsed). To run `lb` at least the
following variables must be set:

```
config.toml
---
# database to read
# this can also be set via LOCKBOX_STORE
store = "$HOME/.passwords/secrets.kdbx"

[credentials]
# the keying object to use to ACTUALLY unlock the passwords (e.g. using a gpg encrypted file with the password inside of it)
# this can also be set via LOCKBOX_KEY
# alternative credential settings for key files are also available
password = ["gpg", "--decrypt", "$HOME/.secrets/key.gpg"]
```

Use `lb help verbose` for additional information about options and
configuration variables

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
