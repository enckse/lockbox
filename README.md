Lockbox
===

A [pass](https://www.passwordstore.org/) inspired password manager that uses a system
keyring or command for password input over using a GPG key and uses a keepass database as the backing data store.

![build](https://github.com/enckse/lockbox/actions/workflows/main.yml/badge.svg)

# usage

## environment

The following variables must be set to use `lb`

For example set:
```
# the keying object to use to ACTUALLY unlock the passwords (e.g. using a gpg encrypted file with the password inside of it)
LOCKBOX_KEY="gpg --decrypt /Users/alice/.secrets/key.gpg"
# the location, on disk, of the password store
LOCKBOX_STORE=/Users/alice/.passwords/secrets.kdbx
```

In cases where `lb` outputs colored terminal output, this coloring behavior can be disabled:
```
LOCKBOX_NOCOLOR="yes"
```

To disable clipboard _completions_ for bash
```
LOCKBOX_NOCLIP="yes"
```

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
# use -m for a multiline entry
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
lb totp token
# 'token' must contain an entry with the name of LOCKBOX_TOTP
```

The token can be automatically copied to the clipboard too
```
lb totp -clip token
```

## git integration

To manage the `.lb` files in a git repository and see _actual_ text diffs add this to a `.gitconfig`
```
[diff "lb"]
    textconv = lb hash
```

Setup the `.gitattributes` for the repository to include
```
*.lb diff=lb
```

## build

Requires `make`

Clone this repository and:
```
make
```
