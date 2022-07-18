package encrypt

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	random "math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/google/shlex"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLength   = 32
	nonceLength = 24
	padLength   = 256
	saltLength  = 16
	// PlainKeyMode is plaintext based key resolution.
	PlainKeyMode = "plaintext"
	// CommandKeyMode will run an external command to get the key (from stdout).
	CommandKeyMode = "command"
)

var (
	cryptoVers       = []byte{0, 1}
	cryptoVersLength = len(cryptoVers)
)

type (
	// Lockbox represents a method to encrypt/decrypt locked files.
	Lockbox struct {
		secret [keyLength]byte
		file   string
	}

	// LockboxOptions represent options to create a lockbox from.
	LockboxOptions struct {
		Key      string
		KeyMode  string
		File     string
		callback func(string) string
	}
)

// FromFile decrypts a file-system based encrypted file.
func FromFile(file string) ([]byte, error) {
	l, err := NewLockbox(LockboxOptions{File: file})
	if err != nil {
		return nil, err
	}
	return l.Decrypt()
}

// ToFile encrypts data to a file-system based file.
func ToFile(file string, data []byte) error {
	l, err := NewLockbox(LockboxOptions{File: file})
	if err != nil {
		return err
	}
	return l.Encrypt(data)
}

// NewLockbox creates a new usable lockbox instance.
func NewLockbox(options LockboxOptions) (Lockbox, error) {
	return newLockbox(options.Key, options.KeyMode, options.File)
}

func newLockbox(key, keyMode, file string) (Lockbox, error) {
	useKeyMode := keyMode
	if useKeyMode == "" {
		useKeyMode = os.Getenv(inputs.KeyModeEnv)
	}
	if useKeyMode == "" {
		useKeyMode = CommandKeyMode
	}
	useKey := key
	if useKey == "" {
		useKey = os.Getenv(inputs.KeyEnv)
	}
	if useKey == "" {
		return Lockbox{}, errors.New("no key given")
	}
	b, err := getKey(useKeyMode, useKey)
	if err != nil {
		return Lockbox{}, err
	}

	if len(b) == 0 {
		return Lockbox{}, errors.New("key is empty")
	}

	var secretKey [keyLength]byte
	copy(secretKey[:], b)
	return Lockbox{secret: secretKey, file: file}, nil
}

func pad(salt, key []byte) ([keyLength]byte, error) {
	d := pbkdf2.Key(key, salt, 4096, keyLength, sha512.New)
	if len(d) != keyLength {
		return [keyLength]byte{}, errors.New("invalid key result from pad")
	}
	var obj [keyLength]byte
	copy(obj[:], d[:keyLength])
	return obj, nil
}

func getKey(keyMode, name string) ([]byte, error) {
	var data []byte
	switch keyMode {
	case CommandKeyMode:
		parts, err := shlex.Split(name)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		b, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		data = b
	case PlainKeyMode:
		data = []byte(name)
	default:
		return nil, errors.New("unknown keymode")
	}
	return []byte(strings.TrimSpace(string(data))), nil
}

func init() {
	random.Seed(time.Now().UnixNano())
}

// Encrypt will encrypt contents to file.
func (l Lockbox) Encrypt(datum []byte) error {
	var nonce [nonceLength]byte
	padTo := random.Intn(padLength)
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	data := datum
	if data == nil {
		b, err := inputs.RawStdin()
		if err != nil {
			return err
		}
		data = b
	}
	var padding [padLength]byte
	if _, err := io.ReadFull(rand.Reader, padding[:]); err != nil {
		return err
	}
	var salt [saltLength]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	var write []byte
	write = append(write, byte(padTo))
	write = append(write, padding[0:padTo]...)
	write = append(write, data...)
	key, err := pad(salt[:], l.secret[:])
	if err != nil {
		return err
	}
	encrypted := secretbox.Seal(nonce[:], write, &nonce, &key)
	var persist []byte
	persist = append(persist, cryptoVers...)
	persist = append(persist, salt[:]...)
	persist = append(persist, encrypted...)
	return os.WriteFile(l.file, persist, 0600)
}

// Decrypt will decrypt an object from file.
func (l Lockbox) Decrypt() ([]byte, error) {
	var nonce [nonceLength]byte
	var salt [saltLength]byte
	encrypted, err := os.ReadFile(l.file)
	if err != nil {
		return nil, err
	}
	copy(salt[:], encrypted[cryptoVersLength:saltLength+cryptoVersLength])
	copy(nonce[:], encrypted[cryptoVersLength+saltLength:cryptoVersLength+saltLength+nonceLength])
	key, err := pad(salt[:], l.secret[:])
	if err != nil {
		return nil, err
	}
	decrypted, ok := secretbox.Open(nil, encrypted[cryptoVersLength+saltLength+nonceLength:], &nonce, &key)
	if !ok {
		return nil, errors.New("decrypt not ok")
	}

	padding := int(decrypted[0])
	return decrypted[1+padding:], nil
}
