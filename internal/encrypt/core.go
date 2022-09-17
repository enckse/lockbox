// Package encrypt handles encryption/decryption.
package encrypt

import (
	"errors"
	"os"

	"github.com/enckse/lockbox/internal/inputs"
)

const (
	keyLength = 32
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
