// Package encrypt handles encryption/decryption.
package encrypt

import (
	"crypto/sha512"
	"errors"
	random "math/rand"
	"os"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLength                       = 32
	secretBoxAlgorithmVersion uint8 = 1
	isSecretBox                     = "secretbox"
	aesGCMAlgorithmVersion    uint8 = 2
)

type (
	// Lockbox represents a method to encrypt/decrypt locked files.
	Lockbox struct {
		secret [keyLength]byte
		file   string
		algo   string
	}

	// LockboxOptions represent options to create a lockbox from.
	LockboxOptions struct {
		Key       string
		KeyMode   string
		File      string
		Algorithm string
	}
	algorithm interface {
		encrypt(k, d []byte) ([]byte, error)
		decrypt(k, d []byte) ([]byte, error)
		version() []byte
	}
)

func init() {
	random.Seed(time.Now().UnixNano())
}

func newAlgorithmFromVersion(vers uint8) algorithm {
	switch vers {
	case secretBoxAlgorithmVersion:
		return secretBoxAlgorithm{}
	case aesGCMAlgorithmVersion:
		return aesGCMAlgorithm{}
	}
	return nil
}

func newAlgorithm(mode string) algorithm {
	useMode := mode
	if mode == "" {
		useMode = inputs.EnvOrDefault(inputs.EncryptModeEnv, isSecretBox)
	}
	switch useMode {
	case isSecretBox:
		return secretBoxAlgorithm{}
	case "aesgcm":
		return aesGCMAlgorithm{}
	}
	return nil
}

func algoVersion(v uint8) []byte {
	return []byte{0, v}
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
	return newLockbox(options.Key, options.KeyMode, options.File, options.Algorithm)
}

func newLockbox(key, keyMode, file, algo string) (Lockbox, error) {
	b, err := inputs.GetKey(key, keyMode)
	if err != nil {
		return Lockbox{}, err
	}
	var secretKey [keyLength]byte
	copy(secretKey[:], b)
	return Lockbox{secret: secretKey, file: file, algo: algo}, nil
}

// Encrypt will encrypt contents to file.
func (l Lockbox) Encrypt(datum []byte) error {
	data := datum
	if data == nil {
		b, err := inputs.RawStdin()
		if err != nil {
			return err
		}
		data = b
	}
	box := newAlgorithm(l.algo)
	if box == nil {
		return errors.New("unknown algorithm detected")
	}
	b, err := box.encrypt(l.secret[:], data)
	if err != nil {
		return err
	}
	var persist []byte
	persist = append(persist, box.version()...)
	persist = append(persist, b...)
	return os.WriteFile(l.file, persist, 0600)
}

// Decrypt will decrypt an object from file.
func (l Lockbox) Decrypt() ([]byte, error) {
	encrypted, err := os.ReadFile(l.file)
	if err != nil {
		return nil, err
	}
	version := len(algoVersion(0))
	if len(encrypted) <= version {
		return nil, errors.New("invalid decryption data")
	}
	data := encrypted[version:]
	box := newAlgorithmFromVersion(encrypted[1])
	if box == nil {
		return nil, errors.New("unable to detect algorithm")
	}
	return box.decrypt(l.secret[:], data)
}
