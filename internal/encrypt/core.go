// Package encrypt handles encryption/decryption.
package encrypt

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	random "math/rand"
	"os"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLength            = 32
	algorithmBaseVersion = 0
	padLength            = 256
)

const (
	noopBoxAlgorithVersion algorithmVersions = iota
	secretBoxAlgorithmVersion
	aesGCMAlgorithmVersion
)

var (
	defaultAlgorithm = secretBoxAlgorithm{}
	algorithms       = []algorithm{defaultAlgorithm, aesGCMAlgorithm{}}
)

type (
	algorithmVersions uint8
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
		version() algorithmVersions
		name() string
		dataSize() int
	}
)

func init() {
	random.Seed(time.Now().UnixNano())
}

func newAlgorithmFromVersion(vers algorithmVersions) algorithm {
	for _, a := range algorithms {
		if a.version() == vers {
			return a
		}
	}
	return nil
}

func newAlgorithm(mode string) algorithm {
	useMode := mode
	if mode == "" {
		useMode = inputs.EnvOrDefault(inputs.EncryptModeEnv, defaultAlgorithm.name())
	}
	for _, a := range algorithms {
		if useMode == a.name() {
			return a
		}
	}
	return nil
}

func algoVersion(v uint8) []byte {
	return []byte{algorithmBaseVersion, v}
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
	if len(data) == 0 {
		return errors.New("no data given")
	}
	padTo := random.Intn(padLength)
	var padding [padLength]byte
	if _, err := io.ReadFull(rand.Reader, padding[:]); err != nil {
		return err
	}
	box := newAlgorithm(l.algo)
	if box == nil {
		return errors.New("unknown algorithm detected")
	}
	var write []byte
	write = append(write, byte(padTo))
	write = append(write, padding[0:padTo]...)
	write = append(write, data...)
	b, err := box.encrypt(l.secret[:], write)
	if err != nil {
		return err
	}
	var persist []byte
	algo := algoVersion(uint8(box.version()))
	persist = append(persist, algo...)
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
	if encrypted[0] != algorithmBaseVersion {
		return nil, errors.New("unknown input data header")
	}
	box := newAlgorithmFromVersion(algorithmVersions(encrypted[1]))
	if box == nil {
		return nil, errors.New("unable to detect algorithm")
	}
	if len(data) <= box.dataSize() {
		return nil, errors.New("data is invalid for decryption")
	}
	decrypted, err := box.decrypt(l.secret[:], data)
	if err != nil {
		return nil, err
	}
	padding := int(decrypted[0])
	return decrypted[1+padding:], nil
}
