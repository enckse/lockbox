lockbox
===

A [pass](https://www.passwordstore.org/) inspired password manager that uses a system
keyring, command, or plaintext solution for password input (over using a gpg key only) and uses a kdbx database as the backing data store.

# install

Build locally or install via go
```
go install github.com/enckse/lockbox/cmd/lb@latest
```

# usage

## upfront

While `lb` uses a `.kdbx` formatted file that can be opened by a variety of tools, it is highly opinionated on how it will store data in the database. Any
`.kdbx` used with `lb` should be managed by `lb` with a fallback ability to use other tools to alter/view the file otherwise. Mainly, lockbox itself
uses a common format so that it does not lock a user into a custom file format nor rely entirely on gpg.
`lb` does try to play nice with standard fields used within kdbx files, but it may disagree with other tools on how to manage/store/update them.

## configuration

`lb` uses TOML configuration file(s)

```
config.toml
---
# database to read
store = "$HOME/.passwords/secrets.kdbx"

[credentials]
# the keying object to use to ACTUALLY unlock the passwords (e.g. using a gpg encrypted file with the password inside of it)
# alternative credential settings for key files are also available
password = ["gpg", "--decrypt", "$HOME/.secrets/key.gpg"]
```

Use `lb help verbose` for additional information about functionality and
`lb help config` for details on configuration variables

### supported systems

`lb` should work on combinations of the following:
- linux/macOS/WSL
- zsh/bash (for completions)
- amd64/arm64

## usage

### clipboard

Copy entries to clipboard
```
lb clip my/secret/password
```

### insert

Create a new entry
```
lb insert my/new/key/password
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
lb show my/key/notes
```

### totp

To get a totp token
```
lb totp show token/path/otp
```

The token can be automatically copied to the clipboard too
```
lb totp clip token/path/otp
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

Clone this repository and:
```
make
```

_run `make check` to run tests_
