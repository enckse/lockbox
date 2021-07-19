Lockbox
===

A [pass](https://www.passwordstore.org/) inspired password manager that uses a system
keyring for password input over using a GPG key.

# usage

Currently `lb` is only working/developed on macOS (though it could be extended to other operating systems)

## environment

The fllowing variables must be set to use `lb`

For _macos_ set
```
# the keychange key name
LOCKBOX_KEY=com.domain.keyname
# the location, on disk, of the password store
LOCKBOX_STORE=/Users/alice/.passwords
# the keychain is macos keychain
LOCKBOX_KEYMODE=macos
# to utilize totp token generation set the offset (within the repository) where totp tokens are saved
LOCKBOX_TOTP=keys/totp/
```

## usage

### clipboard

Copy entries to clipboard
```
lb -c my/secret/password
# or lb clip
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
# or lb list
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
# must be within the subdir of LOCKBOX_TOTP
```

### stats

View password stats/information about changes
```
lb stats some/key/to/stat
```

### pwgen

Generate passwords
```
lb pwgen
```

This _requires_ these additional environment variables
```
# list of directories which provide a list of word/inputs to pwgen
PWGEN_SOURCE=/directories/:/colondelimited
# the characters allowed via password generation from the SOURCE entries
PWGEN_ALLOWED=abcABC123
# special characters that will be inserted randomly into passwords
PWGEN_SPECIAL=.,[]{};:^
# a 'sed' command that will be run against the generated password
PWGEN_SED=s/a/z/g
```

## git integration

To manage the `.lb` files in a git repository and see _actual_ text diffs and this to a `.gitconfig`
```
[diff "lb"]
    textconv = lb diff
```

## build

Clone this repository and:
```
make
```
